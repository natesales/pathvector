package cmd

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/natesales/pathvector/pkg/bird"
	"github.com/natesales/pathvector/pkg/templating"
	"github.com/natesales/pathvector/pkg/util"
)

var (
	realProtocolNames bool
	onlyBGP           bool
	showTags          bool
	tagFilter         []string
)

func init() {
	statusCmd.Flags().BoolVarP(&realProtocolNames, "real-protocol-names", "r", false, "use real protocol names")
	statusCmd.Flags().BoolVarP(&onlyBGP, "bgp", "b", false, "only show BGP protocols")
	statusCmd.Flags().BoolVar(&showTags, "tags", false, "show tags column")
	statusCmd.Flags().StringArrayVarP(&tagFilter, "filter", "f", []string{}, "tags to filter by")
	rootCmd.AddCommand(statusCmd)
}

var statusCmd = &cobra.Command{
	Use:     "status [session]",
	Aliases: []string{"s", "status"},
	Short:   "Show protocol status",
	Run: func(cmd *cobra.Command, args []string) {
		c, err := loadConfig()
		if err != nil {
			log.Warnf("Error loading config, falling back to no-config output parsing: %s", err)
		}
		// TODO: Use defaults

		if len(args) > 0 {
			query := strings.Join(args, " ")

			// Load protocol names map
			protos, err := protocols(c.BIRDDirectory)
			if err != nil {
				log.Fatal(err)
			}

			log.Debugf("Looking for protocol for %s", query)
			birdProtoName, richName := protocolByQuery(query, protos)
			if birdProtoName == "" {
				log.Fatalf("no protocol found for query: %s", query)
			}

			log.Infof("Showing status for %s (%s)", richName, birdProtoName)
			commandOutput, _, err := bird.RunCommand("show protocols all "+birdProtoName, c.BIRDSocket)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println(commandOutput)
		} else {
			commandOutput, _, err := bird.RunCommand("show protocols all", c.BIRDSocket)
			if err != nil {
				log.Fatal(err)
			}

			var protos map[string]*templating.Protocol
			if !realProtocolNames {
				protos, err = protocols(c.BIRDDirectory)
				if err != nil {
					log.Fatal(err)
				}
			}

			protocolStates, err := bird.ParseProtocols(commandOutput)
			if err != nil {
				log.Fatal(err)
			}

			header := []string{"Peer", "AS", "Neighbor", "State", "In", "Out", "Since", "Info"}
			if showTags {
				header = append(header, "Tags")
			}
			util.PrintTable(header, func() [][]string {
				var table [][]string
				for _, protocolState := range protocolStates {
					if !onlyBGP || protocolState.BGP != nil {
						neighborAddr, neighborAS := "-", "-"
						if protocolState.BGP != nil {
							neighborAS = parseTableInt(protocolState.BGP.NeighborAS)
							neighborAddr = protocolState.BGP.NeighborAddress
						}

						// Lookup peer in protocol JSON
						protocolName := protocolState.Name
						var tags []string
						if p, found := protos[protocolState.Name]; found {
							protocolName = p.Name
							tags = p.Tags
						}

						if len(tagFilter) == 0 || containsAny(tagFilter, tags) {
							row := []string{
								protocolName,
								neighborAS,
								neighborAddr,
								colorStatus(protocolState.State),
								parseTableInt(protocolState.Routes.Imported),
								parseTableInt(protocolState.Routes.Exported),
								protocolState.Since,
								colorStatus(protocolState.Info),
							}
							if showTags {
								row = append(row, strings.Join(tags, ", "))
							}
							table = append(table, row)
						}
					}
				}
				return table
			}())
		}
	},
}

// containsAny checks if two string slices contain any of the same elements
func containsAny(a []string, b []string) bool {
	for _, i := range a {
		for _, j := range b {
			if i == j {
				return true
			}
		}
	}
	return false
}

func parseTableInt(i int) string {
	if i == -1 {
		return ""
	}
	return fmt.Sprintf("%d", i)
}

func colorStatus(s string) string {
	switch {
	case s == "up" || s == "Established":
		return color.GreenString(s)
	case strings.Contains(s, "Error") || s == "down":
		return color.RedString(s)
	case strings.Contains(s, "Connect") || s == "start":
		return color.YellowString(s)
	default:
		return s
	}
}
