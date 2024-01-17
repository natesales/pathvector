package cmd

import (
	"fmt"
	"slices"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/natesales/pathvector/pkg/bird"
)

func init() {
	rootCmd.AddCommand(&cobra.Command{
		Use:   "reload [in|out] [session]",
		Short: "Reload a session",
		Run: func(cmd *cobra.Command, args []string) {
			reloadRestartHandler(args, "reload")
		},
	})
	rootCmd.AddCommand(&cobra.Command{
		Use:   "restart [session]",
		Short: "Restart a session",
		Run: func(cmd *cobra.Command, args []string) {
			reloadRestartHandler(args, "restart")
		},
	})
}

func usage() {
	log.Fatal("Usage: pathvector reload [direction] [session]")
}

func parseArgs(args []string) (string, string) {
	if len(args) == 0 {
		usage()
	}

	direction := args[0]
	query := strings.Join(args[1:], " ")
	if !slices.Contains([]string{"in", "out"}, direction) {
		direction = "both"
		query = strings.Join(args, " ")
	}

	return query, direction
}

func reloadRestartHandler(args []string, verb string) {
	// Load config file
	c, err := loadConfig()
	if err != nil {
		log.Fatal(err)
	}

	query, direction := parseArgs(args)

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

	if !confirmYesNo(fmt.Sprintf("Are you sure you want to %s %s (%s)?", verb, richName, birdProtoName)) {
		log.Fatal("Cancelled")
	}

	// Reload protocol
	birdCmd := verb
	if verb == "reload" && direction != "both" {
		birdCmd += " " + direction
	}
	birdCmd += " " + birdProtoName

	out, _, err := bird.RunCommand(birdCmd, c.BIRDSocket)
	if err != nil {
		log.Fatal(err)
	}
	log.Info(out)
}
