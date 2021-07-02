package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(reloadCmd)
}

var reloadCmd = &cobra.Command{
	Use:   "reload",
	Short: "Reload the current configuration",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Reloading")
	},
}
