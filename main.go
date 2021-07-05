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

	"github.com/jessevdk/go-flags"
	log "github.com/sirupsen/logrus"
)

var version = "devel" // set by the build process

// Embedded filesystem

//go:embed templates/*
var embedFs embed.FS

// run allows integration testing with an arbitrary slice of arguments. During normal runtime,
// the main function will pass os.Args as the argument.
func run(args []string) {
	if len(args) == 2 && args[1] == "generate-config-docs" {
		documentConfig()
		os.Exit(1)
	}

	// Parse cli flags
	_, err := flags.ParseArgs(&cliFlags, args)
	if err != nil {
		if !strings.Contains(err.Error(), "Usage") {
			log.Fatal(err)
		}
		os.Exit(1)
	}

	// Enable debug logging in development releases
	if cliFlags.Verbose {
		log.SetLevel(log.DebugLevel)
	}

	if cliFlags.ShowVersion {
		log.Printf("Pathvector version %s (https://pathvector.io)\n", version)
		os.Exit(0)
	}

	// Validate mode flag
	if !(cliFlags.Mode == "generate" || cliFlags.Mode == "daemon") {
		log.Fatalf("Invalid mode '%s', expected 'generate' or 'daemon'", cliFlags.Mode)
	}

	// Check lockfile
	if cliFlags.LockFileDirectory != "" {
		if _, err := os.Stat(path.Join(cliFlags.LockFileDirectory, cliFlags.Mode+".lock")); err == nil {
			log.Fatal("Lockfile exists, exiting")
		} else if os.IsNotExist(err) {
			// If the lockfile doesn't exist, create it
			log.Debugln("Lockfile doesn't exist, creating one")
			if err := ioutil.WriteFile(path.Join(cliFlags.LockFileDirectory, cliFlags.Mode+".lock"), []byte(""), 0755); err != nil {
				log.Fatalf("Writing lockfile: %v", err)
			}
		} else {
			log.Fatalf("Accessing lockfile: %v", err)
		}
	}

	log.Debugf("Starting pathvector %s mode: %s", version, cliFlags.Mode)

	// Load the config file from config file
	log.Debugf("Loading config from %s", cliFlags.ConfigFile)
	configFile, err := ioutil.ReadFile(cliFlags.ConfigFile)
	if err != nil {
		log.Fatal("Reading config file: " + err.Error())
	}
	globalConfig, err := loadConfig(configFile)
	if err != nil {
		log.Fatal(err)
	}
	log.Debugln("Finished loading config")

	// Mode conditional
	if cliFlags.Mode == "generate" {
		// Load templates from embedded filesystem
		log.Debugln("Loading templates from embedded filesystem")
		err = loadTemplates(embedFs)
		if err != nil {
			log.Fatal(err)
		}
		log.Debugln("Finished loading templates")

		// Create cache directory
		log.Debugf("Making cache directory %s", cliFlags.CacheDirectory)
		if err := os.MkdirAll(cliFlags.CacheDirectory, os.FileMode(0755)); err != nil {
			log.Fatal(err)
		}

		// Create the global output file
		log.Debug("Creating global config")
		globalFile, err := os.Create(path.Join(cliFlags.CacheDirectory, "bird.conf"))
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
		files, err := filepath.Glob(path.Join(cliFlags.CacheDirectory, "AS*.conf"))
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
			peerFileName := path.Join(cliFlags.CacheDirectory, fmt.Sprintf("AS%d_%s.conf", *peerData.ASN, *sanitize(peerName)))
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
		cmd := exec.Command(cliFlags.BirdBinary, "-c", "bird.conf", "-p")
		cmd.Dir = cliFlags.CacheDirectory
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			log.Fatalf("BIRD config validation: %v", err)
		}
		log.Infof("BIRD config validation passed")

		if !cliFlags.DryRun {
			// Write VRRP config
			writeVRRPConfig(globalConfig)

			if cliFlags.WebUIFile != "" {
				writeUIFile(globalConfig)
			} else {
				log.Infof("Web UI is not defined, NOT writing UI")
			}

			// Remove old configs
			birdConfigFiles, err := filepath.Glob(path.Join(cliFlags.BirdDirectory, "AS*.conf"))
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
			files, err := filepath.Glob(path.Join(cliFlags.CacheDirectory, "*.conf"))
			if err != nil {
				log.Fatal(err)
			}
			for _, f := range files {
				fileNameParts := strings.Split(f, "/")
				fileNameTail := fileNameParts[len(fileNameParts)-1]
				newFileLoc := path.Join(cliFlags.BirdDirectory, fileNameTail)
				log.Debugf("Moving %s to %s", f, newFileLoc)
				if err := MoveFile(f, newFileLoc); err != nil {
					log.Fatalf("Moving cache file to bird directory: %v", err)
				}
			}

			if !cliFlags.NoConfigure {
				log.Infoln("Reconfiguring BIRD")
				if err = runBirdCommand("configure", cliFlags.BirdSocket); err != nil {
					log.Fatal(err)
				}
			} else {
				log.Infoln("Option --no-configure is set, NOT reconfiguring bird")
			}
		} // end dry run check
	} else if cliFlags.Mode == "daemon" {
		log.Infof("Starting optimizer")

		sourceMap := map[string][]string{} // Peer name to list of source addresses
		for peerName, peerData := range globalConfig.Peers {
			if peerData.OptimizerEnabled != nil && *peerData.OptimizerEnabled {
				if peerData.OptimizerProbeSources == nil || len(*peerData.OptimizerProbeSources) < 1 {
					log.Fatalf("[%s] has optimize enabled but no probe sources", peerName)
				}
				sourceMap[peerName] = *peerData.OptimizerProbeSources
			}
		}

		if len(sourceMap) == 0 {
			log.Fatal("No peers have optimization enabled, exiting now")
		}

		log.Fatal(startProbe(globalConfig.Optimizer, sourceMap))
	}

	// Delete lockfile
	if cliFlags.LockFileDirectory != "" {
		if err := os.Remove(path.Join(cliFlags.LockFileDirectory, cliFlags.Mode+".lock")); err != nil {
			log.Fatalf("Removing lockfile: %v", err)
		}
	}
}

func main() {
	run(os.Args)
}
