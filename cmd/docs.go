package cmd

import (
	"github.com/natesales/pathvector/pkg/config"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(docsCmd)
}

var docsCmd = &cobra.Command{
	Use:    "docs",
	Short:  "Generate documentation",
	Hidden: true,
	Run: func(cmd *cobra.Command, args []string) {
		config.DocumentConfig()
	},
}
