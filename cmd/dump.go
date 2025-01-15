package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/natesales/pathvector/pkg/util"
	"github.com/natesales/pathvector/pkg/util/log"
)

var (
	dumpYaml bool
)

func init() {
	dumpCmd.Flags().BoolVar(&dumpYaml, "yaml", false, "use YAML output (else use formatted table output)")
	rootCmd.AddCommand(dumpCmd)
}

var dumpCmd = &cobra.Command{
	Use:   "dump",
	Short: "Dump configuration",
	Run: func(cmd *cobra.Command, args []string) {
		c, err := loadConfig()
		if err != nil {
			log.Fatal(err)
		}

		if dumpYaml {
			yamlBytes, err := yaml.Marshal(&c)
			if err != nil {
				log.Fatal(err)
			}
			log.Println(string(yamlBytes))
		} else {
			var data [][]string
			for peerName, peerData := range c.Peers {
				data = append(data, []string{
					peerName,
					fmt.Sprintf("%d", *peerData.ASN),
					fmt.Sprintf("%d", *peerData.LocalPref),
					fmt.Sprintf("%d", *peerData.Prepends),
					strings.Join(*peerData.NeighborIPs, ", "),
					util.StrDeref(peerData.Template),
					strings.Join(*peerData.BooleanOptions, ", "),
				})
			}

			util.PrintTable([]string{"Name", "ASN", "Local Pref", "Prepends", "Neighbors", "Template", "Options"}, data)
		}
	},
}
