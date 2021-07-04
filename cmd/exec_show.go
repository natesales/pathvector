package cmd

import (
	"github.com/spf13/cobra"
)

func init() {
	execCmd.AddCommand(execShowCmd)
}

var execShowCmd = &cobra.Command{
	Use:     "show",
	Short:   "Show protocols",
	Aliases: []string{"s"},
	Run: func(cmd *cobra.Command, args []string) {
		execRemoteCommand("/show", map[string]string{})
	},
}
