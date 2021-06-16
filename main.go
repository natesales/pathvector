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
	"reflect"
	"strings"

	"github.com/jessevdk/go-flags"
	log "github.com/sirupsen/logrus"
)

var version = "devel" // set by the build process

// Embedded filesystem

//go:embed templates/*
var embedFs embed.FS

// printStructInfo prints a configuration struct values
func printStructInfo(label string, instance interface{}) {
	// Fields to exclude from print output
	excludedFields := []string{""}
	s := reflect.ValueOf(instance).Elem()
	typeOf := s.Type()
	for i := 0; i < s.NumField(); i++ {
		attrName := typeOf.Field(i).Name
		if !(contains(excludedFields, attrName)) {
			v := reflect.Indirect(s.Field(i))
			if v.IsValid() {
				log.Infof("[%s] field %s = %v\n", label, attrName, v)
			}
		}
	}
}

// runPeeringDbQuery updates peer values from PeeringDB
func runPeeringDbQuery(peerName string, peerData *peer) {
	pDbData, err := getPeeringDbData(*peerData.ASN)
	if err != nil {
		log.Fatalf("[%s] unable to get PeeringDB data: %+v", peerName, err)
	}

	// Set import limits
	if *peerData.AutoImportLimits {
		*peerData.ImportLimit6 = pDbData.ImportLimit4
		*peerData.ImportLimit6 = pDbData.ImportLimit6

		if pDbData.ImportLimit4 == 0 {
			log.Warnf("[%s] has an IPv4 import limit of zero from PeeringDB", peerName)
		}
		if pDbData.ImportLimit6 == 0 {
			log.Warnf("[%s] has an IPv6 import limit of zero from PeeringDB", peerName)
		}
	}

	// Set as-set if auto-as-set is enabled and there isn't a manual AS set defined
	if *peerData.AutoASSet && peerData.ASSet == nil {
		if pDbData.ASSet == "" {
			log.Fatalf("[%s] doesn't have an as-set in PeeringDB", peerName)
			// TODO: Exit or skip this peer?
		}

		// If the as-set has a space in it, split and pick the first one
		if strings.Contains(pDbData.ASSet, " ") {
			pDbData.ASSet = strings.Split(pDbData.ASSet, " ")[0]
			log.Warnf("[%s] has a space in their PeeringDB as-set field. Selecting first element %s", peerName, pDbData.ASSet)
		}

		// Trim IRRDB prefix
		if strings.Contains(pDbData.ASSet, "::") {
			peerData.ASSet = &strings.Split(pDbData.ASSet, "::")[1]
			log.Warnf("[%s] has an IRRDB prefix in their PeeringDB as-set field. Using %s", peerName, *peerData.ASSet)
		} else {
			peerData.ASSet = &pDbData.ASSet
		}
	}
}

// run allows integration testing with an arbitrary slice of arguments. During normal runtime,
// the main function will pass os.Args as the argument.
func run(args []string) {
	if len(args) == 2 && args[1] == "generate-config-docs" {
		documentConfig()
		os.Exit(1)
	} else if len(os.Args) == 2 && args[1] == "generate-cli-docs" {
		documentCliFlags()
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
	if //noinspection GoBoolExpressions
	version == "devel" || cliFlags.Verbose {
		log.SetLevel(log.DebugLevel)
	}

	if cliFlags.ShowVersion {
		log.Printf("Wireframe version %s (https://github.com/natesales/wireframe)\n", version)
		os.Exit(0)
	}

	// Validate mode flag
	if !(cliFlags.Mode == "generate" || cliFlags.Mode == "daemon") {
		log.Fatalf("Invalid mode '%s', expected 'generate' or 'daemon'", cliFlags.Mode)
	}

	// Check lockfile
	if cliFlags.LockFileDirectory != "" {
		if _, err := os.Stat(path.Join(cliFlags.LockFileDirectory, cliFlags.Mode+".lock")); err == nil {
			log.Fatal("Wireframe lockfile exists, exiting")
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

	log.Debugf("Starting wireframe %s mode: %s", version, cliFlags.Mode)

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
		printStructInfo("wireframe.global", globalConfig)

		// Iterate over peers
		for peerName, peerData := range globalConfig.Peers {
			// Set sanitized peer name
			peerData.ProtocolName = sanitize(peerName)

			// If a PeeringDB query is required
			if *peerData.AutoImportLimits || *peerData.AutoASSet {
				log.Debugf("[%s] has auto-import-limits or auto-as-set, querying PeeringDB", peerName)

				runPeeringDbQuery(peerName, peerData)
			} // end peeringdb query enabled

			// Build IRR prefix sets
			if *peerData.FilterIRR {
				// Check for empty as-set
				if peerData.ASSet == nil || *peerData.ASSet == "" {
					log.Fatalf("[%s] has filter-irr enabled and no as-set defined", peerName)
				}

				prefixesFromIRR4, err := getIRRPrefixSet(*peerData.ASSet, 4, globalConfig.IRRServer)
				if err != nil {
					log.Fatalf("[%s] unable to get IRR prefix list from %s", peerName, *peerData.ASSet)
				}
				if peerData.PrefixSet4 == nil {
					peerData.PrefixSet4 = &[]string{}
				}
				pfx4 := append(*peerData.PrefixSet4, prefixesFromIRR4...)
				peerData.PrefixSet4 = &pfx4
				if len(pfx4) == 0 {
					log.Fatalf("[%s] has a prefix filter defined but no IPv4 prefixes", peerName)
				}

				prefixesFromIRR6, err := getIRRPrefixSet(*peerData.ASSet, 6, globalConfig.IRRServer)
				if err != nil {
					log.Fatalf("[%s] unable to get IRR prefix list from %s", peerName, *peerData.ASSet)
				}
				if peerData.PrefixSet6 == nil {
					peerData.PrefixSet6 = &[]string{}
				}
				pfx6 := append(*peerData.PrefixSet6, prefixesFromIRR6...)
				peerData.PrefixSet6 = &pfx6
				if len(pfx6) == 0 {
					log.Fatalf("[%s] has a prefix filter defined but no IPv6 prefixes", peerName)
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
