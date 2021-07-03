package cmd

import (
	"github.com/spf13/cobra"
)

func init() {
	execCmd.AddCommand(execProtocolsCmd)
}

var execProtocolsCmd = &cobra.Command{
	Use:     "protocols",
	Short:   "Show protocols",
	Aliases: []string{"p"},
	Run: func(cmd *cobra.Command, args []string) {
		execRemoteCommand("/protocols", map[string]string{})
	},
}
