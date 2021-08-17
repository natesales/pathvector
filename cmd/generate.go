package cmd

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/natesales/pathvector/internal/bird"
	"github.com/natesales/pathvector/internal/config"
	"github.com/natesales/pathvector/internal/embed"
	"github.com/natesales/pathvector/internal/irr"
	"github.com/natesales/pathvector/internal/peeringdb"
	"github.com/natesales/pathvector/internal/portal"
	"github.com/natesales/pathvector/internal/templating"
	"github.com/natesales/pathvector/internal/util"
)

func init() {
	rootCmd.AddCommand(generateCmd)
}

var generateCmd = &cobra.Command{
	Use:     "generate",
	Short:   "generate router configuration",
	Aliases: []string{"gen", "g"},
	Run: func(cmd *cobra.Command, args []string) {
		// Check lockfile
		if lockFile != "" {
			if _, err := os.Stat(lockFile); err == nil {
				log.Fatal("Lockfile exists, exiting")
			} else if os.IsNotExist(err) {
				// If the lockfile doesn't exist, create it
				log.Debugln("Lockfile doesn't exist, creating one")
				if err := ioutil.WriteFile(lockFile, []byte(""), 0755); err != nil {
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
		c, err := config.Load(configFile)
		if err != nil {
			log.Fatal(err)
		}
		log.Debugln("Finished loading config")

		// Load templates from embedded filesystem
		log.Debugln("Loading templates from embedded filesystem")
		err = templating.Load(embed.FS)
		if err != nil {
			log.Fatal(err)
		}
		log.Debugln("Finished loading templates")

		// Create cache directory
		log.Debugf("Making cache directory %s", c.CacheDirectory)
		if err := os.MkdirAll(c.CacheDirectory, os.FileMode(0755)); err != nil {
			log.Fatal(err)
		}

		// Create the global output file
		log.Debug("Creating global config")
		globalFile, err := os.Create(path.Join(c.CacheDirectory, "bird.conf"))
		if err != nil {
			log.Fatalf("Create global BIRD output file: %v", err)
		}
		log.Debug("Finished creating global config file")

		// Render the global template and write to buffer
		log.Debug("Writing global config file")
		err = templating.GlobalTemplate.ExecuteTemplate(globalFile, "global.tmpl", c)
		if err != nil {
			log.Fatalf("Execute global template: %v", err)
		}
		log.Debug("Finished writing global config file")

		// Remove old peer-specific configs
		if err := util.RemoveFileGlob(path.Join(c.CacheDirectory, "AS*.conf")); err != nil {
			log.Fatalf("Removing old config files: %v", err)
		}

		// Print global config
		util.PrintStructInfo("pathvector.global", c)

		// Iterate over peers
		for peerName, peerData := range c.Peers {
			log.Printf("Processing AS%d %s", *peerData.ASN, peerName)

			// If a PeeringDB query is required
			if *peerData.AutoImportLimits || *peerData.AutoASSet {
				log.Debugf("[%s] has auto-import-limits or auto-as-set, querying PeeringDB", peerName)

				peeringdb.Update(peerData, c.PeeringDBQueryTimeout)
			} // end peeringdb query enabled

			// Build IRR prefix sets
			if *peerData.FilterIRR {
				if err := irr.Update(peerData, c.IRRServer, c.IRRQueryTimeout); err != nil {
					log.Fatal(err)
				}
			}

			util.PrintStructInfo(peerName, peerData)

			// Create peer file
			peerFileName := path.Join(c.CacheDirectory, fmt.Sprintf("AS%d_%s.conf", *peerData.ASN, *util.Sanitize(peerName)))
			peerSpecificFile, err := os.Create(peerFileName)
			if err != nil {
				log.Fatalf("Create peer specific output file: %v", err)
			}

			// Render the template and write to buffer
			var b bytes.Buffer
			log.Debugf("[%s] Writing config", peerName)
			err = templating.PeerTemplate.ExecuteTemplate(&b, "peer.tmpl", &templating.Wrapper{Name: peerName, Peer: *peerData, Config: *c})
			if err != nil {
				log.Fatalf("Execute template: %v", err)
			}

			// Reformat config and write template to file
			if _, err := peerSpecificFile.Write([]byte(bird.Reformat(b.String()))); err != nil {
				log.Fatalf("Write template to file: %v", err)
			}

			log.Debugf("[%s] Wrote config", peerName)

		} // end peer loop

		// Run BIRD config validation
		bird.Validate(c.BIRDBinary, c.CacheDirectory)

		if !dryRun {
			// Write VRRP config
			templating.WriteVRRPConfig(c.VRRPInstances, c.KeepalivedConfig)

			if c.WebUIFile != "" {
				templating.WriteUIFile(c)
			} else {
				log.Infof("Web UI is not defined, NOT writing UI")
			}

			bird.MoveCacheAndReconfigure(c.BIRDDirectory, c.CacheDirectory, c.BIRDSocket, noConfigure)
		} // end dry run check

		// Update portal
		if c.PortalHost != "" {
			log.Infoln("Updating peering portal")
			if err := portal.Record(c.PortalHost, c.PortalKey, c.Hostname, c.Peers, c.BIRDSocket); err != nil {
				log.Fatal(err)
			}
		}

		// Delete lockfile
		if lockFile != "" {
			if err := os.Remove(lockFile); err != nil {
				log.Fatalf("Removing lockfile: %v", err)
			}
		}
	},
}
