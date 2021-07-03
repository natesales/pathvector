package cmd

import (
	"fmt"
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
	"github.com/natesales/pathvector/internal/processor"
	"github.com/natesales/pathvector/internal/templating"
	"github.com/natesales/pathvector/internal/util"
)

// Set by build process
var version = "dev"
var platform = "generic"
var commit = "unknown"
var date = "unknown"

var description = "Pathvector is a declarative routing control plane platform for BGP with robust filtering and route optimization."

var (
	configFile string
	addr       string
	verbose    bool
	rootCmd    = NewRootCommand()
)

var global *config.Global

var streamWriter http.ResponseWriter
var streamFlusher http.Flusher

type internalLogger struct{}

func (l internalLogger) Write(p []byte) (int, error) {
	fmt.Print(string(p))
	if streamWriter != nil {
		streamWriter.Write(p)
		streamFlusher.Flush()
	}
	return 0, nil
}

func setupStream(w http.ResponseWriter, r *http.Request) error {
	streamWriter = w

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	log.SetOutput(internalLogger{})
	verbose := r.URL.Query().Get("verbose")
	if verbose == "true" {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}

	var ok bool
	streamFlusher, ok = w.(http.Flusher)
	if !ok {
		return fmt.Errorf("connection does not support streaming")
	}
	return nil // nil error
}

func NewRootCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "pathvector",
		Short: description,
		Run: func(cmd *cobra.Command, args []string) {
			http.HandleFunc("/protocols", func(w http.ResponseWriter, r *http.Request) {
				if err := setupStream(w, r); err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}

				log.Debugln("Loading configuration")
				if err := loadConfig(); err != nil {
					log.Print(err)
				}

				log.Debugln("Running 'show protocols'")
				if err := bird.Run("show protocols", global.BirdSocket, global.BirdSocketConnectTimeout, true); err != nil {
					log.Print(err)
				}
			})

			http.HandleFunc("/load", func(w http.ResponseWriter, r *http.Request) {
				if err := setupStream(w, r); err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}

				var asn uint64
				asnString := r.URL.Query().Get("asn")
				if asnString == "" {
					asn = 0
				} else {
					var err error
					asn, err = strconv.ParseUint(asnString, 10, 32)
					if err != nil {
						log.Print(err)
					}
				}

				log.Printf("Pathvector server version %s", version)

				if err := loadConfig(); err != nil {
					log.Print(err)
				}

				// Reload peer config(s)
				if err := writePeerConfigs(uint32(asn)); err != nil {
					log.Print(err)
				}

				log.Debugln("Validating BIRD config")
				if err := bird.Validate(global); err != nil {
					log.Print(err)
				}
				log.Debugln("BIRD config validation passed")

				// Write VRRP config
				if err := templating.WriteVRRPConfig(global); err != nil {
					log.Print(err)
				}

				if global.WebUIFile != "" {
					if err := templating.WriteUIFile(global); err != nil {
						log.Print(err)
					}
				} else {
					log.Debugln("Web UI is not defined, NOT writing UI")
				}

				if err := replaceRunningConfig(global, uint32(asn)); err != nil {
					log.Print(err)
				}
			})

			log.Printf("Starting HTTP server on %s", addr)
			log.Fatal(http.ListenAndServe(addr, nil))
		},
	}
}

// Execute executes the root command
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(func() {
		if verbose {
			log.SetLevel(log.DebugLevel)
		}
	})
	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "/etc/pathvector.yml", "config file")
	rootCmd.PersistentFlags().StringVarP(&addr, "listen", "l", ":8084", "API listen address")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "enable verbose logging")
}

// loadConfig reads the YAML config file, templates, and writes out bird.conf
func loadConfig() error {
	// Clear protocol names slice
	templating.ProtocolNames = []string{}

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

// writePeerConfigs writes an ASN's configuration to disk, or all ASNs if ASN is 0
func writePeerConfigs(asn uint32) error {
	// Remove old cache
	log.Debugln("Purging cache")
	var files []string
	var err error
	if asn == 0 {
		files, err = filepath.Glob(path.Join(global.CacheDirectory, "AS*.conf"))
	} else {
		files, err = filepath.Glob(path.Join(global.CacheDirectory, fmt.Sprintf("AS%d*.conf", asn)))
	}
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
		if (asn == 0) || (asn != 0 && *peerData.ASN == int(asn)) {
			log.Debugf("Processing %s AS%d", peerName, *peerData.ASN)
			if err := processor.Run(global, peerName, peerData); err != nil {
				return err
			}
		}
	}

	return nil // nil error
}

// replaceRunningConfig removes the old BIRD configuration, copies the new one from cache, and reloads BIRD
func replaceRunningConfig(global *config.Global, asn uint32) error {
	// Remove old configs
	var files []string
	var err error
	if asn == 0 {
		files, err = filepath.Glob(path.Join(global.BirdDirectory, "AS*.conf"))
	} else {
		files, err = filepath.Glob(path.Join(global.BirdDirectory, fmt.Sprintf("AS%d*.conf", asn)))
	}
	if err != nil {
		return err
	}
	for _, f := range files {
		log.Debugf("Removing old BIRD config file %s", f)
		if err := os.Remove(f); err != nil {
			return fmt.Errorf("removing old BIRD config files: %v", err)
		}
	}

	// Copy from cache to bird config
	files, err = filepath.Glob(path.Join(global.CacheDirectory, "*.conf"))
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

	log.Debugln("Reconfiguring BIRD")
	if err = bird.Run("configure", global.BirdSocket, global.BirdSocketConnectTimeout, false); err != nil {
		return err
	}
	log.Debugln("Finished BIRD reconfigure")

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
