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

		log.Debugf("Loading config from %s", configFile)
		configFile, err := os.ReadFile(configFile)
		if err != nil {
			log.Fatalf("Reading config file: %s", err)
		}
		c, err := process.Load(configFile)
		if err != nil {
			log.Fatal(err)
		}
		log.Debugln("Finished loading config")

		_, birdVersion, err := bird.RunCommand("", c.BIRDSocket)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("BIRD: %s\n", birdVersion)
	},
}
