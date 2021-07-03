package cmd

import (
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"

	"github.com/natesales/pathvector/internal/bird"
	"github.com/natesales/pathvector/internal/config"
	"github.com/natesales/pathvector/internal/processor"
	"github.com/natesales/pathvector/internal/templating"
	"github.com/natesales/pathvector/internal/util"
	"github.com/natesales/pathvector/proto"
)

var listenAddr string
var updateInterval int

var global *config.Global

var globalSrv protobuf.ReloadService_FetchResponseServer

type reloadServer struct{}

type gRPCLogger struct{}

func (l gRPCLogger) Write(p []byte) (int, error) {
	fmt.Print(string(p))
	if err := globalSrv.Send(&protobuf.ReloadResponse{Message: string(p)}); err != nil {
		log.Warnf("gRPC send: %v", err)
	}

	return 0, nil
}

func (s reloadServer) FetchResponse(req *protobuf.ReloadRequest, srv protobuf.ReloadService_FetchResponseServer) error {
	globalSrv = srv
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		log.SetOutput(gRPCLogger{})
		log.Printf("Pathvector server version %s", version)

		if err := loadConfig(); err != nil {
			log.Print(err)
		}

		// Reload peer config(s)
		if err := writePeerConfigs(); err != nil {
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
			log.Infof("Web UI is not defined, NOT writing UI")
		}

		if err := replaceRunningConfig(global); err != nil {
			log.Print(err)
		}
	}()

	wg.Wait()
	return nil
}

var daemonCmd = &cobra.Command{
	Use:   "daemon",
	Short: "Start the pathvector daemon",
	Run: func(cmd *cobra.Command, args []string) {
		log.Printf("Starting gRPC listener on %s", listenAddr)
		lis, err := net.Listen("tcp", listenAddr)
		if err != nil {
			log.Fatalf("tcp listen: %v", err)
		}

		s := grpc.NewServer()
		protobuf.RegisterReloadServiceServer(s, reloadServer{})

		if err := s.Serve(lis); err != nil {
			log.Fatalf("gRPC serve: %v", err)
		}
	},
}

func init() {
	daemonCmd.Flags().StringVarP(&listenAddr, "listen", "l", ":8084", "API listen endpoint")
	daemonCmd.Flags().IntVarP(&updateInterval, "interval", "i", 12, "Regular update interval in hours")
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
	if err = bird.Run("configure", global.BirdSocket, global.BirdSocketConnectTimeout); err != nil {
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
