package cmd

import (
	"fmt"
	"reflect"

	"github.com/spf13/cobra"

	"github.com/natesales/pathvector/pkg/plugin"
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

func versionBanner() {
	fmt.Printf(`Pathvector %s
Built %s on %s
Plugins: `, version, commit, date)
	if len(plugin.Get()) > 0 {
		fmt.Println("")
		for name, p := range plugin.Get() {
			fmt.Printf("  %s - %s [%s]\n", name, p.Description(), reflect.TypeOf(p).PkgPath())
		}
	} else {
		fmt.Println("(none)")
	}
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version information",
	Run: func(cmd *cobra.Command, args []string) {
		versionBanner()
	},
}
