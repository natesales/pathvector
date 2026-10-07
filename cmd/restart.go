package cmd

import (
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/natesales/pathvector/pkg/bird"
)

var restartAll bool

func init() {
	restartCmd.Flags().BoolVarP(&restartAll, "all", "a", false, "restart all BIRD protocols")
	rootCmd.AddCommand(restartCmd)
}

// birdErrorResponses are substrings of BIRD responses that indicate a restart failed
var birdErrorResponses = []string{"No protocols match", "syntax error"}

var restartCmd = &cobra.Command{
	Use:   "restart [name...]",
	Short: "Restart BIRD protocols",
	Long: `Restart BIRD protocols.

Each name may be a peer name as shown by "pathvector status" (restarting all of the peer's
BIRD protocols, e.g. both IPv4 and IPv6 sessions) or a BIRD protocol name as shown by
"pathvector status --real-protocol-names".`,
	Run: func(cmd *cobra.Command, args []string) {
		if !restartAll && len(args) == 0 {
			log.Fatal("Usage: pathvector restart <name> [name...] (or --all)")
		}
		if restartAll && len(args) > 0 {
			log.Fatal("Protocol names can't be combined with --all")
		}

		c, err := loadConfig()
		if err != nil {
			log.Warnf("Error loading config, using default BIRD socket and directory: %s", err)
		}
		birdSocket, birdDirectory := birdPaths(c)

		var targets []string
		if restartAll {
			targets = []string{"all"}
		} else {
			protocols, err := readProtocolNames(birdDirectory)
			if err != nil {
				log.Warnf("Reading protocol names, treating names as BIRD protocol names: %v", err)
			}
			targets = resolveProtocols(args, protocols)
		}

		failed := 0
		for _, target := range targets {
			command := "restart all"
			if !restartAll {
				// Quote the name so BIRD treats it as a protocol pattern: an unquoted unknown
				// name is a syntax error ("unexpected CF_SYM_UNDEFINED"), whereas a quoted one
				// gives a clear "No protocols match". Reject characters that would break the quoting.
				if strings.ContainsAny(target, "\" \t\r\n") {
					log.Errorf("Invalid protocol name %q", target)
					failed++
					continue
				}
				command = fmt.Sprintf("restart \"%s\"", target)
			}

			log.Debugf("Restarting %s", target)
			resp, _, err := bird.RunCommand(command, birdSocket)
			if err != nil {
				log.Fatal(err)
			}
			resp = strings.TrimSpace(resp)

			isError := false
			for _, e := range birdErrorResponses {
				if strings.Contains(resp, e) {
					isError = true
				}
			}
			if isError {
				log.Errorf("%s: %s", target, resp)
				failed++
				continue
			}
			for _, line := range strings.Split(resp, "\n") {
				fmt.Println(strings.TrimSpace(line))
			}
		}

		if failed > 0 {
			log.Fatalf("Failed to restart %d of %d protocols", failed, len(targets))
		}
	},
}
