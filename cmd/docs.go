package cmd

import (
	"github.com/spf13/cobra"

	"github.com/natesales/pathvector/internal/config"
)

func init() {
	rootCmd.AddCommand(docsCmd)
}

var docsCmd = &cobra.Command{
	Use:    "docs",
	Short:  "Generate documentation",
	Hidden: true,
	Run: func(cmd *cobra.Command, args []string) {
		config.PrintDocs()
	},
}
