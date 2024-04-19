package cmd

import (
	"fmt"
	"os"
	"reflect"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/natesales/pathvector/pkg/bird"
	"github.com/natesales/pathvector/pkg/plugin"
	"github.com/natesales/pathvector/pkg/process"
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

		log.Debugf("Loading config from %s", configFile)
		configFile, err := os.ReadFile(configFile)
		if err != nil {
			log.Fatalf("Reading config file: %s", err)
		}
		c, err := process.Load(configFile)
		if err != nil {
			log.Fatal(err)
		}
		log.Debug("Finished loading config")

		_, birdVersion, err := bird.RunCommand("", c.BIRDSocket)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("BIRD: %s\n", birdVersion)
	},
}
