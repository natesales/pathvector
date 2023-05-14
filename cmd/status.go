package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/fatih/color"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/natesales/pathvector/pkg/bird"
	"github.com/natesales/pathvector/pkg/util"
)

var (
	userProtocolNames bool
	onlyBGP           bool
)

func init() {
	statusCmd.Flags().BoolVarP(&userProtocolNames, "user-protocol-names", "u", false, "use user-defined protocol names")
	statusCmd.Flags().BoolVarP(&onlyBGP, "bgp", "b", false, "only show BGP protocols")
	rootCmd.AddCommand(statusCmd)
}

var statusCmd = &cobra.Command{
	Use:     "status",
	Aliases: []string{"s", "status"},
	Short:   "Show protocol status",
	Run: func(cmd *cobra.Command, args []string) {
		c, err := loadConfig()
		if err != nil {
			log.Warnf("Error loading config, falling back to no-config output parsing: %s", err)
		}

		commandOutput, _, err := bird.RunCommand("show protocols all", c.BIRDSocket)
		if err != nil {
			log.Fatal(err)
		}

		// Read protocol names map
		var protocolNames map[string]string
		if userProtocolNames {
			contents, err := os.ReadFile(path.Join("/etc/bird/", "protocol_names.json"))
			if err != nil {
				log.Fatalf("Reading protocol names: %v", err)
			}
			if err := json.Unmarshal(contents, &protocolNames); err != nil {
				log.Fatalf("Unmarshalling protocol names: %v", err)
			}
		}

		protocolStates, err := bird.ParseProtocols(commandOutput)
		if err != nil {
			log.Fatal(err)
		}

		util.PrintTable([]string{"Peer", "AS", "Neighbor", "State", "In", "Out", "Since", "Info"}, func() [][]string {
			var table [][]string
			for _, protocolState := range protocolStates {
				if !onlyBGP || protocolState.BGP != nil {
					if protocolState.BGP == nil {
						table = append(table, []string{
							protocolName(protocolState.Name, protocolNames),
							"-",
							"-",
							colorStatus(protocolState.State),
							parseTableInt(protocolState.Routes.Imported),
							parseTableInt(protocolState.Routes.Exported),
							protocolState.Since,
							colorStatus(protocolState.Info),
						})
					} else { // BGP
						table = append(table, []string{
							protocolName(protocolState.Name, protocolNames),
							parseTableInt(protocolState.BGP.NeighborAS),
							protocolState.BGP.NeighborAddress,
							colorStatus(protocolState.State),
							parseTableInt(protocolState.Routes.Imported),
							parseTableInt(protocolState.Routes.Exported),
							protocolState.Since,
							colorStatus(protocolState.Info),
						})
					}
				}
			}
			return table
		}())
	},
}

func protocolName(n string, names map[string]string) string {
	if userSuppliedName, found := names[n]; found {
		return userSuppliedName
	} else {
		return n
	}
}

func parseTableInt(i int) string {
	if i == -1 {
		return ""
	}
	return fmt.Sprintf("%d", i)
}

func colorStatus(s string) string {
	if s == "up" || s == "Established" {
		return color.GreenString(s)
	} else if strings.Contains(s, "Error") || s == "down" {
		return color.RedString(s)
	} else if strings.Contains(s, "Connect") || s == "start" {
		return color.YellowString(s)
	}
	return s
}
