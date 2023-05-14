package cmd

import (
	"fmt"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/natesales/pathvector/pkg/match"
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
	Use:   "match [asn]",
	Short: "Find common IXPs for a given ASN",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			log.Fatal("Usage: pathvector match ASN")
		}

		c, err := loadConfig()
		if err != nil {
			log.Fatal(err)
		}

		var peeringDbTimeout uint
		peeringDbTimeout = 10
		if matchLocalASN == 0 {
			matchLocalASN = uint(c.ASN)
			peeringDbTimeout = c.PeeringDBQueryTimeout
		}

		peerASN, err := strconv.Atoi(strings.TrimPrefix(strings.ToUpper(args[0]), "AS"))
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println(match.CommonIXs(uint32(matchLocalASN), uint32(peerASN), yamlFormat, peeringDbTimeout, c.PeeringDBAPIKey))
	},
}
