package cmd

import (
	"fmt"
	"reflect"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/natesales/pathvector/pkg/bird"
	"github.com/natesales/pathvector/pkg/plugin"
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

func versionBanner() string {
	buf := fmt.Sprintf(`Pathvector %s
Built %s on %s
`, version, commit, date)
	if len(plugin.Get()) > 0 {
		buf += "Plugins:\n"
		for name, p := range plugin.Get() {
			buf += fmt.Sprintf("  %s - %s [%s]\n", name, p.Description(), reflect.TypeOf(p).PkgPath())
		}
	} else {
		buf += "No plugins"
	}

	return buf
}

func printVersionBanner() {
	fmt.Println(versionBanner())
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version information",
	Run: func(cmd *cobra.Command, args []string) {
		printVersionBanner()

		c, err := loadConfig()
		if err != nil {
			log.Fatal(err)
		}

		_, birdVersion, err := bird.RunCommand("", c.BIRDSocket)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("BIRD: %s\n", birdVersion)
	},
}
