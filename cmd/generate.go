package cmd

import (
	"github.com/spf13/cobra"

	"github.com/natesales/pathvector/pkg/process"
)

var (
	withdraw bool
	offline  bool
)

func init() {
	generateCmd.Flags().BoolVarP(&withdraw, "withdraw", "w", false, "Withdraw all routes")
	generateCmd.Flags().BoolVar(&offline, "offline", false, "Don't query IRR or PeeringDB, only use data cached by previous runs")
	rootCmd.AddCommand(generateCmd)
}

var generateCmd = &cobra.Command{
	Use:     "generate",
	Short:   "Generate router configuration",
	Aliases: []string{"gen", "g"},
	Run: func(cmd *cobra.Command, args []string) {
		process.Run(configFile, lockFile, version, noConfigure, dryRun, withdraw, offline)
	},
}
