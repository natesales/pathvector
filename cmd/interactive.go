package cmd

import (
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(interactiveCmd)
}

var interactiveCmd = &cobra.Command{
	Use:   "interactive",
	Short: "Interactive CLI",
	Run: func(cmd *cobra.Command, args []string) {
		// TODO
	},
}
