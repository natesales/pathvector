package cmd

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"io/ioutil"

	"github.com/natesales/pathvector/internal/config"
	"github.com/natesales/pathvector/internal/portal"
)

func init() {
	rootCmd.AddCommand(portalCmd)
}

var portalCmd = &cobra.Command{
	Use:     "portal-update",
	Aliases: []string{"p"},
	Short:   "Update portal status",
	Run: func(cmd *cobra.Command, args []string) {
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

		if err := portal.Record(c.PortalHost, c.PortalKey, c.Hostname, c.Peers, c.BIRDSocket); err != nil {
			log.Fatal(err)
		}
	},
}
