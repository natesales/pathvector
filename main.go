package main

import (
	"bytes"
	"embed"
	"fmt"
	"io/ioutil"
	"os"
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
			log.Infof("[%s] field %s = %v\n", label, attrName, reflect.Indirect(s.Field(i)))
		}
	}
}

func main() {
	if len(os.Args) == 2 && os.Args[1] == "generate-config-docs" {
		documentConfig()
		os.Exit(1)
	} else if len(os.Args) == 2 && os.Args[1] == "generate-cli-docs" {
		documentCliFlags()
		os.Exit(1)
	}

	// Parse cli flags
	_, err := flags.ParseArgs(&cliFlags, os.Args)
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
		log.Printf("wireframe version %s (https://github.com/natesales/wireframe)\n", version)
		os.Exit(0)
	}

	log.Debugf("Starting wireframe %s", version)

	//Load templates from embedded filesystem
	log.Debugln("Loading templates from embedded filesystem")
	err = loadTemplates(embedFs)
	if err != nil {
		log.Fatal(err)
	}
	log.Debugln("Finished loading templates")

	// Load the config file from config file
	log.Debugf("Loading config from %s", cliFlags.ConfigFile)
	configFile, err := ioutil.ReadFile(cliFlags.ConfigFile)
	if err != nil {
		log.Fatal("reading config file: " + err.Error())
	}
	globalConfig, err := loadConfig(configFile)
	if err != nil {
		log.Fatal(err)
	}
	log.Debugln("Finished loading config")

	if !cliFlags.DryRun {
		// Create the global output file
		log.Debug("Creating global config")
		globalFile, err := os.Create(path.Join(cliFlags.BirdDirectory, "bird.conf"))
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
		// TODO: Store last known-good config
		files, err := filepath.Glob(path.Join(cliFlags.BirdDirectory, "AS*.conf"))
		if err != nil {
			log.Fatal(err)
		}
		for _, f := range files {
			if err := os.Remove(f); err != nil {
				log.Fatalf("Removing old config files: %v", err)
			}
		}
	} else {
		log.Info("Dry run is enabled, skipped writing global config and removing old peer configs")
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

			// Set as-set
			if *peerData.AutoASSet {
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
					*peerData.ASSet = strings.Split(pDbData.ASSet, "::")[1]
					log.Warnf("[%s] has an IRRDB prefix in their PeeringDB as-set field. Using %s", peerName, *peerData.ASSet)
				} else {
					*peerData.ASSet = pDbData.ASSet
				}
			}
		} // end peeringdb query enabled

		// Build IRR prefix sets
		if *peerData.FilterIRR {
			// Check for empty as-set
			if peerData.ASSet == nil || *peerData.ASSet == "" {
				log.Fatalf("[%s] has filter-irr enabled and no as-set defined", peerName)
			}

			prefixesFromIRR4, err := getIRRPrefixSet(*peerData.ASSet, 4, globalConfig)
			if err != nil {
				log.Warnf("[%s] has an IRRDB prefix in their PeeringDB as-set field. Using %s", peerName, *peerData.ASSet)
			}
			*peerData.PrefixSet4 = append(*peerData.PrefixSet4, prefixesFromIRR4)
			prefixesFromIRR6, err := getIRRPrefixSet(*peerData.ASSet, 6, globalConfig)
			if err != nil {
				log.Warnf("[%s] has an IRRDB prefix in their PeeringDB as-set field. Using %s", peerName, *peerData.ASSet)
			}
			*peerData.PrefixSet6 = append(*peerData.PrefixSet6, prefixesFromIRR6)
		}

		printStructInfo(peerName, peerData)

		// Write peer file
		if !cliFlags.DryRun {
			// Create the peer specific file
			peerFileName := path.Join(cliFlags.BirdDirectory, fmt.Sprintf("AS%d_%s.conf", *peerData.ASN, *sanitize(peerName)))
			peerSpecificFile, err := os.Create(peerFileName)
			if err != nil {
				log.Fatalf("Create peer specific output file: %v", err)
			}

			// Render the template and write to buffer
			var b bytes.Buffer
			log.Debugf("[%s] Writing config", peerName)
			err = peerTemplate.ExecuteTemplate(&b, "peer.tmpl", &wrapper{peerName, *peerData, *globalConfig})
			if err != nil {
				log.Fatalf("execute template: %v", err)
			}

			// Reformat config and write template to file
			if _, err := peerSpecificFile.Write([]byte(reformatBirdConfig(b.String()))); err != nil {
				log.Fatalf("write template to file: %v", err)
			}

			log.Debugf("[%s] Wrote config", peerName)
		} else {
			log.Infof("Dry run is enabled, skipped writing peer config(s)")
		}
	} // end peer loop

	if !cliFlags.DryRun {
		// Write VRRP config
		writeVRRPConfig(globalConfig)

		if cliFlags.WebUIFile != "" {
			writeUIFile(globalConfig)
		} else {
			log.Infof("web UI is not defined, NOT writing UI")
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
}
