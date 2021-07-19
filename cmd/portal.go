package cmd

import (
	"github.com/natesales/pathvector/internal/config"
	"github.com/natesales/pathvector/internal/portal"
	"io/ioutil"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	hostname   string
	portalHost string
	portalKey  string
)

func init() {
	portalUpdateCmd.Flags().StringVar(&hostname, "hostname", "", "router hostname")
	portalUpdateCmd.Flags().StringVar(&portalHost, "portal-host", "", "peering portal host")
	portalUpdateCmd.Flags().StringVar(&portalKey, "portal-key", "", "peering portal API key")
	rootCmd.AddCommand(portalUpdateCmd)
}

var portalUpdateCmd = &cobra.Command{
	Use:   "portal-update",
	Short: "Update peering-portal with local sessions",
	Run: func(cmd *cobra.Command, args []string) {
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

		if err := portal.Record(portalHost, portalKey, hostname, c.Peers, c.BIRDSocket); err != nil {
			log.Fatal(err)
		}
	},
}
