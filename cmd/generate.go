package cmd

import (
	"github.com/spf13/cobra"

	"github.com/natesales/pathvector/pkg/process"
)

var (
	withdraw bool
)

func init() {
	generateCmd.Flags().BoolVarP(&withdraw, "withdraw", "w", false, "Withdraw all routes")
	rootCmd.AddCommand(generateCmd)
}

var generateCmd = &cobra.Command{
	Use:     "generate",
	Short:   "Generate router configuration",
	Aliases: []string{"gen", "g"},
	Run: func(cmd *cobra.Command, args []string) {
		process.Run(configFile, lockFile, version, noConfigure, dryRun, withdraw)
	},
}
