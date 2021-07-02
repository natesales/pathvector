package cmd

import (
	"fmt"
	"github.com/natesales/pathvector/internal/processor"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/natesales/pathvector/internal/bird"
	"github.com/natesales/pathvector/internal/config"
	"github.com/natesales/pathvector/internal/templating"
	"github.com/natesales/pathvector/internal/util"
)

var listenAddr string

var global *config.Global

func apiResponse(w http.ResponseWriter, ok bool, msg string) {
	if !ok {
		log.Warn(msg)
		w.WriteHeader(http.StatusInternalServerError)
	}

	if _, err := w.Write([]byte(msg)); err != nil {
		log.Warnf("HTTP write: %v", err)
	}
}

var daemonCmd = &cobra.Command{
	Use:   "daemon",
	Short: "Start the pathvector daemon",
	Run: func(cmd *cobra.Command, args []string) {
		http.HandleFunc("/load", func(w http.ResponseWriter, r *http.Request) {
			if err := loadConfig(); err != nil {
				apiResponse(w, false, err.Error())
				return
			}

			// Reload peer config(s)
			asn := r.URL.Query().Get("asn")
			var peerWriteErr error
			if asn == "" { // If the ASN query param is not set, reload all
				peerWriteErr = writePeerConfigs()
			} else { // If the ASN query param is set, only update the given ASN
				asnInt, err := strconv.ParseUint(asn, 10, 32)
				if err != nil {
					apiResponse(w, false, fmt.Sprintf("invalid ASN %s", asn))
					return
				}
				peerWriteErr = writeSinglePeerConfig(uint(asnInt))
			}
			if peerWriteErr != nil {
				apiResponse(w, false, peerWriteErr.Error())
				return
			}

			log.Debugln("Validating BIRD config")
			if err := bird.Validate(global); err != nil {
				apiResponse(w, false, fmt.Errorf("BIRD config validation: %v", err).Error())
				return
			}
			log.Debugln("BIRD config validation passed")

			// Write VRRP config
			if err := templating.WriteVRRPConfig(global); err != nil {
				apiResponse(w, false, err.Error())
				return
			}

			if global.WebUIFile != "" {
				if err := templating.WriteUIFile(global); err != nil {
					apiResponse(w, false, err.Error())
					return
				}
			} else {
				log.Infof("Web UI is not defined, NOT writing UI")
			}

			if err := replaceRunningConfig(global); err != nil {
				apiResponse(w, false, err.Error())
				return
			}

			apiResponse(w, true, "Update complete")
		})

		log.Infof("Starting Pathvector %s on %s", version, listenAddr)
		log.Fatal(http.ListenAndServe(listenAddr, nil))
	},
}

func init() {
	daemonCmd.Flags().StringVarP(&listenAddr, "listen", "l", ":8080", "API listen endpoint")
	rootCmd.AddCommand(daemonCmd)
}

// loadConfig reads the YAML config file, templates, and writes out bird.conf
func loadConfig() error {
	// Load the config file from config file
	log.Debugf("Loading config from %s", configFile)
	configFile, err := ioutil.ReadFile(configFile)
	if err != nil {
		return fmt.Errorf("reading config file: %s", err)
	}
	global, err = config.Load(configFile)
	if err != nil {
		return fmt.Errorf("loading configuration: %s", err)
	}
	log.Debugln("Finished loading config")

	// Load templates from embedded filesystem
	log.Debugln("Loading templates from embedded filesystem")
	err = templating.Load()
	if err != nil {
		return fmt.Errorf("loading templates: %s", err)
	}
	log.Debugln("Finished loading templates")

	// Create cache directory
	log.Debugf("Making cache directory %s", global.CacheDirectory)
	if err := os.MkdirAll(global.CacheDirectory, os.FileMode(0755)); err != nil {
		return fmt.Errorf("creating cache directory: %s", err)
	}

	// Create the global output file
	log.Debug("Creating global config")
	globalFile, err := os.Create(path.Join(global.CacheDirectory, "bird.conf"))
	if err != nil {
		return fmt.Errorf("creating global bird.conf: %s", err)
	}
	log.Debug("Finished creating global config file")

	// Render the global template and write to buffer
	log.Debug("Writing global config file")
	err = templating.GlobalTemplate.ExecuteTemplate(globalFile, "global.tmpl", global)
	if err != nil {
		return fmt.Errorf("executing global template: %s", err)
	}
	log.Debug("Finished writing global config file")

	// Print global config
	util.PrintStructInfo("pathvector.global", global)

	return nil // nil error
}

// writePeerConfigs writes all the peer configs to disk
func writePeerConfigs() error {
	log.Debugf("Writing all peer configs")

	// Remove old cache
	log.Debugln("Purging cache")
	files, err := filepath.Glob(path.Join(global.CacheDirectory, "AS*.conf"))
	if err != nil {
		return err
	}
	for _, f := range files {
		if err := os.Remove(f); err != nil {
			return fmt.Errorf("removing old config files: %v", err)
		}
	}

	// Iterate over peers
	for peerName, peerData := range global.Peers {
		if err := processor.Run(global, peerName, peerData); err != nil {
			return err
		}
	}

	return nil // nil error
}

// writeSinglePeerConfig writes a single ASN's configs to disk
func writeSinglePeerConfig(asn uint) error {
	log.Debugf("Writing peer configs for AS%d", asn)

	// Remove old cache
	log.Debugf("Purging cache for AS%d", asn)
	files, err := filepath.Glob(path.Join(global.CacheDirectory, fmt.Sprintf("AS%d*.conf", asn)))
	if err != nil {
		return err
	}
	for _, f := range files {
		if err := os.Remove(f); err != nil {
			return fmt.Errorf("removing old config files: %v", err)
		}
	}

	// Iterate over peers
	for peerName, peerData := range global.Peers {
		if uint(*peerData.ASN) == asn {
			if err := processor.Run(global, peerName, peerData); err != nil {
				return err
			}
		}
	}

	return nil // nil error
}

// replaceRunningConfig removes the old BIRD configuration, copies the new one from cache, and reloads BIRD
func replaceRunningConfig(global *config.Global) error {
	// Remove old configs
	birdConfigFiles, err := filepath.Glob(path.Join(global.BirdDirectory, "AS*.conf"))
	if err != nil {
		return err
	}
	for _, f := range birdConfigFiles {
		log.Debugf("Removing old BIRD config file %s", f)
		if err := os.Remove(f); err != nil {
			return fmt.Errorf("removing old BIRD config files: %v", err)
		}
	}

	// Copy from cache to bird config
	files, err := filepath.Glob(path.Join(global.CacheDirectory, "*.conf"))
	if err != nil {
		return err
	}
	for _, f := range files {
		fileNameParts := strings.Split(f, "/")
		fileNameTail := fileNameParts[len(fileNameParts)-1]
		newFileLoc := path.Join(global.BirdDirectory, fileNameTail)
		log.Debugf("Moving %s to %s", f, newFileLoc)
		if err := util.MoveFile(f, newFileLoc); err != nil {
			return fmt.Errorf("moving cache file to bird directory: %v", err)
		}
	}

	log.Infoln("Reconfiguring BIRD")
	if err = bird.Run("configure", global.BirdSocket); err != nil {
		return err
	}
	log.Infoln("Finished BIRD reconfigure")

	return nil // nil error
}

//	log.Infof("Starting optimizer")
//
//	sourceMap := map[string][]string{} // Peer name to list of source addresses
//	for peerName, peerData := range global.Peers {
//		if peerData.OptimizerEnabled != nil && *peerData.OptimizerEnabled {
//			if peerData.OptimizerProbeSources == nil || len(*peerData.OptimizerProbeSources) < 1 {
//				return fmt.Errorf("[%s] has optimize enabled but no probe sources", peerName)
//			}
//			sourceMap[peerName] = *peerData.OptimizerProbeSources
//		}
//	}
//
//	if len(sourceMap) == 0 {
//		log.Fatal("No peers have optimization enabled, exiting now")
//	}
//
//	log.Fatal(startProbe(global.Optimizer, sourceMap))
//}
