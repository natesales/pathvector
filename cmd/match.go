package cmd

import (
	"fmt"
	"os"
	"strconv"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/natesales/pathvector/internal/match"
	"github.com/natesales/pathvector/internal/process"
)

var (
	yamlFormat    bool
	matchLocalASN uint
)

func init() {
	matchCmd.Flags().UintVarP(&matchLocalASN, "local-asn", "l", 0, "Local ASN to match")
	matchCmd.Flags().BoolVarP(&yamlFormat, "yaml", "y", false, "Should YAML configuration be generated? (else plaintext)")
	rootCmd.AddCommand(matchCmd)
}

var matchCmd = &cobra.Command{
	Use:   "match ASN",
	Short: "Find common IXPs for a given ASN",
	Run: func(cmd *cobra.Command, args []string) {

		// Load the config file from config file
		log.Debugf("Loading config from %s", configFile)
		configFile, err := os.ReadFile(configFile)
		if err != nil {
			log.Fatal("Reading config file: %s", err)
		}
		c, err := process.Load(configFile)
		if err != nil {
			log.Fatal(err)
		}
		log.Debugln("Finished loading config")

		var peeringDbTimeout uint
		peeringDbTimeout = 10
		if matchLocalASN == 0 {
			matchLocalASN = uint(c.ASN)
			peeringDbTimeout = c.PeeringDBQueryTimeout
		}

		if len(args) != 1 {
			log.Fatal("Usage: pathvector match ASN")
		}

		peerASN, err := strconv.Atoi(args[0])
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println(match.CommonIXs(uint32(matchLocalASN), uint32(peerASN), yamlFormat, peeringDbTimeout, c.PeeringDBAPIKey))
	},
}
