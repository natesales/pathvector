package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"reflect"

	"github.com/natesales/pathvector/plugins"
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Pathvector %s commit %s date %s\n", version, commit, date)
		if len(plugins.Get()) > 0 {
			fmt.Println("Plugins:")
			for name, plugin := range plugins.Get() {
				fmt.Printf("  %s - %s [%s]\n", name, plugin.Description(), reflect.TypeOf(plugin).PkgPath())
			}
		} else {
			fmt.Println("No plugins")
		}
	},
}
