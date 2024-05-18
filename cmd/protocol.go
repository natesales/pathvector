package cmd

import (
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/natesales/pathvector/pkg/bird"
)

var (
	allowCommands		[]string
	commentMessage		string
)

func init() {
	protocolCmd.Flags().StringVarP(&commentMessage, "message", "m", "", "enable/disable custom message")
	rootCmd.AddCommand(protocolCmd)
}

var protocolCmd = &cobra.Command{
	Use:     "protocol <(r)estart|re(l)oad|(e)nable|(d)isable> <protocol name>",
	Args:	 func(cmd *cobra.Command, args []string) error {
				if len(args) < 1 {
					log.Fatal("requires a command <restart|reload|enable|disable>")
				} else if allowCommand(args[0]) == false {
					log.Fatal("This command is not allowed: ", args[0])
				} else if len(args) < 2 {
					log.Fatal("requires protocol name")
				}
				return nil
			 },
	Aliases: []string{"p", "protocols"},
	Short:   "Protocol command (restart, reload, enable or disable protocol sessions)",
	Long:    "With this command you can restart, reload, enable or disable a protocol, if you want you can use \"all\" as the protocol name to for example restart all protocol sessions.",
	Run: func(cmd *cobra.Command, args []string) {
		c, err := loadConfig()
		if err != nil {
			log.Warnf("Error loading config, falling back to no-config output parsing: %s", err)
		}

		log.Infof("Starting bird protocol command")

		commandOutput, _, err := bird.RunCommand(runCMD(args, commentMessage), c.BIRDSocket)
		if err != nil {
			log.Fatal(err)
		} else if strings.Contains(commandOutput, "syntax error") || strings.Contains(commandOutput, "unexpected CF_SYM_UNDEFINED") || strings.Contains(commandOutput, "expecting END or CF_SYM_KNOWN or TEXT") {
			log.Fatal("The protocol name was not found!")
		}
		log.Debugf("Command Output: %s", commandOutput)

		fmt.Printf("Command %s succeeded for protocol: %s\n", args[0], args[1])
	},
}

// allowCommand check if this command is allowed to run
func allowCommand(cmd string) bool {
	allowCommands := []string{"restart", "reload", "enable", "disable", "r", "l", "e", "d"}
	for _, allowed := range allowCommands {
		if allowed == cmd {
			return true
		}
	}
	return false
}

// runCMD generate the run command
func runCMD(args []string, message string) string {
	switch args[0] {
		case "d":
			args[0] = "disable"
			break
		case "e":
			args[0] = "enable"
			break
		case "r":
			args[0] = "restart"
			break
		case "l":
			args[0] = "reload"
			break
	}

	if args[0] == "disable" && message != "" {
		return args[0] + " " + args[1] + " \"" + message + "\""
	} else if args[0] == "enable" && message != "" {
		return args[0] + " " + args[1] + " \"" + message + "\""
	} else if args[0] == "disable" {
	 	return args[0] + " " + args[1] + " \"Protocol manually " + args[0] + "d by pathvector\""
	}

	return args[0] + " " + args[1]
}