package main

import (
	"bytes"
	"embed"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// Embedded filesystem

//go:embed templates/*
var embedFs embed.FS

// Build process flags
var (
	version = "devel"
	commit  = "unknown"
	date    = "unknown"
)

// CLI Flags
var (
	configFile            string
	lockFileDirectory     string
	verbose               bool
	dryRun                bool
	noConfigure           bool
	birdDirectory         string
	birdBinary            string
	cacheDirectory        string
	birdSocket            string
	keepalivedConfig      string
	webUIFile             string
	peeringDbQueryTimeout uint
	irrQueryTimeout       uint
)

// CLI Commands
var (
	rootCmd = &cobra.Command{
		Use:   "pathvector",
		Short: "Pathvector is a declarative routing platform for BGP which automates route optimization and control plane configuration with secure and repeatable routing policies.",
		Run: func(cmd *cobra.Command, args []string) {
			// Check lockfile
			if lockFileDirectory != "" {
				if _, err := os.Stat(lockFileDirectory); err == nil {
					log.Fatal("Lockfile exists, exiting")
				} else if os.IsNotExist(err) {
					// If the lockfile doesn't exist, create it
					log.Debugln("Lockfile doesn't exist, creating one")
					if err := ioutil.WriteFile(lockFileDirectory, []byte(""), 0755); err != nil {
						log.Fatalf("Writing lockfile: %v", err)
					}
				} else {
					log.Fatalf("Accessing lockfile: %v", err)
				}
			}

			log.Debugf("Starting pathvector %s", version)

			// Load the config file from config file
			log.Debugf("Loading config from %s", configFile)
			configFile, err := ioutil.ReadFile(configFile)
			if err != nil {
				log.Fatal("Reading config file: " + err.Error())
			}
			globalConfig, err := loadConfig(configFile)
			if err != nil {
				log.Fatal(err)
			}
			log.Debugln("Finished loading config")

			// Load templates from embedded filesystem
			log.Debugln("Loading templates from embedded filesystem")
			err = loadTemplates(embedFs)
			if err != nil {
				log.Fatal(err)
			}
			log.Debugln("Finished loading templates")

			// Create cache directory
			log.Debugf("Making cache directory %s", cacheDirectory)
			if err := os.MkdirAll(cacheDirectory, os.FileMode(0755)); err != nil {
				log.Fatal(err)
			}

			// Create the global output file
			log.Debug("Creating global config")
			globalFile, err := os.Create(path.Join(cacheDirectory, "bird.conf"))
			if err != nil {
				log.Fatalf("Create global BIRD output file: %v", err)
			}
			log.Debug("Finished creating global config file")

			// Render the global template and write to buffer
			log.Debug("Writing global config file")
			err = globalTemplate.ExecuteTemplate(globalFile, "global.tmpl", globalConfig)
			if err != nil {
				log.Fatalf("Execute global template: %v", err)
			}
			log.Debug("Finished writing global config file")

			// Remove old peer-specific configs
			files, err := filepath.Glob(path.Join(cacheDirectory, "AS*.conf"))
			if err != nil {
				log.Fatal(err)
			}
			for _, f := range files {
				if err := os.Remove(f); err != nil {
					log.Fatalf("Removing old config files: %v", err)
				}
			}

			// Print global config
			printStructInfo("pathvector.global", globalConfig)

			// Iterate over peers
			for peerName, peerData := range globalConfig.Peers {
				log.Printf("Processing AS%d %s", *peerData.ASN, peerName)

				// Set sanitized peer name
				peerData.ProtocolName = sanitize(peerName)

				// If a PeeringDB query is required
				if *peerData.AutoImportLimits || *peerData.AutoASSet {
					log.Debugf("[%s] has auto-import-limits or auto-as-set, querying PeeringDB", peerName)

					if err := runPeeringDbQuery(peerData); err != nil {
						log.Debugf("[%s] %v", peerName, err)
					}
				} // end peeringdb query enabled

				// Build IRR prefix sets
				if *peerData.FilterIRR {
					if err := buildIRRPrefixSet(peerData, globalConfig.IRRServer); err != nil {
						log.Fatal(err)
					}
				}

				printStructInfo(peerName, peerData)

				// Create peer file
				peerFileName := path.Join(cacheDirectory, fmt.Sprintf("AS%d_%s.conf", *peerData.ASN, *sanitize(peerName)))
				peerSpecificFile, err := os.Create(peerFileName)
				if err != nil {
					log.Fatalf("Create peer specific output file: %v", err)
				}

				// Render the template and write to buffer
				var b bytes.Buffer
				log.Debugf("[%s] Writing config", peerName)
				err = peerTemplate.ExecuteTemplate(&b, "peer.tmpl", &wrapper{peerName, *peerData, *globalConfig})
				if err != nil {
					log.Fatalf("Execute template: %v", err)
				}

				// Reformat config and write template to file
				if _, err := peerSpecificFile.Write([]byte(reformatBirdConfig(b.String()))); err != nil {
					log.Fatalf("Write template to file: %v", err)
				}

				log.Debugf("[%s] Wrote config", peerName)

			} // end peer loop

			// Run BIRD config validation
			log.Debugln("Validating BIRD config")
			birdCmd := exec.Command(birdBinary, "-c", "bird.conf", "-p")
			birdCmd.Dir = cacheDirectory
			birdCmd.Stdout = os.Stdout
			birdCmd.Stderr = os.Stderr
			if err := birdCmd.Run(); err != nil {
				log.Fatalf("BIRD config validation: %v", err)
			}
			log.Infof("BIRD config validation passed")

			if !dryRun {
				// Write VRRP config
				writeVRRPConfig(globalConfig)

				if webUIFile != "" {
					writeUIFile(globalConfig)
				} else {
					log.Infof("Web UI is not defined, NOT writing UI")
				}

				// Remove old configs
				birdConfigFiles, err := filepath.Glob(path.Join(birdDirectory, "AS*.conf"))
				if err != nil {
					log.Fatal(err)
				}
				for _, f := range birdConfigFiles {
					log.Debugf("Removing old BIRD config file %s", f)
					if err := os.Remove(f); err != nil {
						log.Fatalf("Removing old BIRD config files: %v", err)
					}
				}

				// Copy from cache to bird config
				files, err := filepath.Glob(path.Join(cacheDirectory, "*.conf"))
				if err != nil {
					log.Fatal(err)
				}
				for _, f := range files {
					fileNameParts := strings.Split(f, "/")
					fileNameTail := fileNameParts[len(fileNameParts)-1]
					newFileLoc := path.Join(birdDirectory, fileNameTail)
					log.Debugf("Moving %s to %s", f, newFileLoc)
					if err := MoveFile(f, newFileLoc); err != nil {
						log.Fatalf("Moving cache file to bird directory: %v", err)
					}
				}

				if !noConfigure {
					log.Infoln("Reconfiguring BIRD")
					if err = runBirdCommand("configure", birdSocket); err != nil {
						log.Fatal(err)
					}
				} else {
					log.Infoln("Option --no-configure is set, NOT reconfiguring bird")
				}
			} // end dry run check

			// Delete lockfile
			if lockFileDirectory != "" {
				if err := os.Remove(lockFileDirectory); err != nil {
					log.Fatalf("Removing lockfile: %v", err)
				}
			}
		},
	}

	subCommands = []*cobra.Command{
		{
			Use:    "probe",
			Short:  "Start optimization probe",
			Hidden: true,
			Run: func(cmd *cobra.Command, args []string) {
				log.Debugf("Loading config from %s", configFile)
				configFile, err := ioutil.ReadFile(configFile)
				if err != nil {
					log.Fatal("Reading config file: " + err.Error())
				}
				globalConfig, err := loadConfig(configFile)
				if err != nil {
					log.Fatal(err)
				}
				log.Debugln("Finished loading config")

				log.Infof("Starting optimizer")
				sourceMap := map[string][]string{} // Peer name to list of source addresses
				for peerName, peerData := range globalConfig.Peers {
					if peerData.OptimizerEnabled != nil && *peerData.OptimizerEnabled {
						if peerData.OptimizerProbeSources == nil || len(*peerData.OptimizerProbeSources) < 1 {
							log.Fatalf("[%s] has optimize enabled but no probe sources", peerName)
						}
						sourceMap[fmt.Sprintf("%d%s%s", *peerData.ASN, optimizationDelimiter, peerName)] = *peerData.OptimizerProbeSources
					}
				}
				log.Debugf("Optimizer probe sources: %v", sourceMap)
				if len(sourceMap) == 0 {
					log.Fatal("No peers have optimization enabled, exiting now")
				}
				globalOptimizer = globalConfig.Optimizer
				if err := startProbe(sourceMap); err != nil {
					log.Fatal(err)
				}
			},
		}, {
			Use:    "docs",
			Short:  "Generate documentation",
			Hidden: true,
			Run: func(cmd *cobra.Command, args []string) {
				documentConfig()
			},
		}, {
			Use:   "version",
			Short: "Show version information",
			Run: func(cmd *cobra.Command, args []string) {
				log.Printf("Pathvector %s commit %s date %s\n", version, commit, date)
			},
		},
	}
)

func init() {
	cobra.OnInitialize(func() {
		if verbose {
			log.SetLevel(log.DebugLevel)
		}
	})
	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "/etc/pathvector.yml", "Configuration file in YAML, TOML, or JSON format")
	rootCmd.PersistentFlags().StringVar(&lockFileDirectory, "lock-file-directory", "", "Lock file directory (lockfile check disabled if empty")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Show verbose log messages")
	rootCmd.PersistentFlags().BoolVarP(&dryRun, "dry-run", "d", false, "Don't modify configuration")
	rootCmd.PersistentFlags().BoolVarP(&noConfigure, "no-configure", "n", false, "Don't configure BIRD")
	rootCmd.PersistentFlags().StringVar(&birdDirectory, "bird-directory", "/etc/bird/", "Directory to store BIRD configs")
	rootCmd.PersistentFlags().StringVar(&birdBinary, "bird-binary", "/usr/sbin/bird", "Path to bird binary")
	rootCmd.PersistentFlags().StringVar(&cacheDirectory, "cache-directory", "/var/run/pathvector/cache/", "Directory to store runtime configuration cache")
	rootCmd.PersistentFlags().StringVar(&birdSocket, "bird-socket", "/run/bird/bird.ctl", "UNIX control socket for BIRD")
	rootCmd.PersistentFlags().StringVar(&keepalivedConfig, "keepalived-config", "/etc/keepalived.conf", "Configuration file for keepalived")
	rootCmd.PersistentFlags().StringVar(&webUIFile, "web-ui-file", "", "File to write web UI to (disabled if empty)")
	rootCmd.PersistentFlags().UintVar(&peeringDbQueryTimeout, "peeringdb-query-timeout", 10, "PeeringDB query timeout in seconds")
	rootCmd.PersistentFlags().UintVar(&irrQueryTimeout, "irr-query-timeout", 30, "IRR query timeout in seconds")

	for _, cmd := range subCommands {
		rootCmd.AddCommand(cmd)
	}
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
