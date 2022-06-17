package cmd

import (
	"fmt"
	"reflect"

	"github.com/spf13/cobra"

	"github.com/natesales/pathvector/plugins"
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

func versionBanner() {
	fmt.Printf(`Pathvector %s
Built %s on %s
Plugins: `, version, commit, date)
	if len(plugins.Get()) > 0 {
		fmt.Println("")
		for name, plugin := range plugins.Get() {
			fmt.Printf("  %s - %s [%s]\n", name, plugin.Description(), reflect.TypeOf(plugin).PkgPath())
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
