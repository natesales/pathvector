package cmd

import (
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/natesales/pathvector/pkg/config"
	"github.com/natesales/pathvector/pkg/plugin"
	"github.com/natesales/pathvector/pkg/process"
)

// These are set indirectly by the build process. The cmd.Execute() function takes these from the main package and sets them in this (cmd) package.
var (
	version = "devel"
	commit  = "unknown"
	date    = "unknown"
)

// CLI Flags
var (
	// Global
	configFile  string
	lockFile    string
	verbose     bool
	trace       bool
	dryRun      bool
	noConfigure bool
)

// CLI Commands
var rootCmd = &cobra.Command{
	Use:   "pathvector",
	Short: "Pathvector is a declarative edge routing platform that automates route optimization and control plane configuration with secure and repeatable routing policy.",
}

func loadConfig() (*config.Config, error) {
	// Load the config file from config file
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
	return c, nil
}

func init() {
	cobra.OnInitialize(func() {
		if verbose {
			log.SetLevel(log.DebugLevel)
		}
		if trace {
			log.SetLevel(log.TraceLevel)
		}
	})
	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "/etc/pathvector.yml", "YAML configuration file")
	rootCmd.PersistentFlags().StringVar(&lockFile, "lock", "", "Lock file (check disabled if empty)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Show verbose log messages")
	rootCmd.PersistentFlags().BoolVarP(&trace, "trace", "t", false, "Show trace log messages")
	rootCmd.PersistentFlags().BoolVarP(&dryRun, "dry-run", "d", false, "Don't modify configuration")
	rootCmd.PersistentFlags().BoolVarP(&noConfigure, "no-configure", "n", false, "Don't configure BIRD")

	// RegisterCommands registers each command plugin
	for _, p := range plugin.Get() {
		pluginCommand := p.Command()
		if pluginCommand != nil {
			// A plugin may provide its own "cli" command (the interactive configuration CLI,
			// documented as `pathvector cli`, was moved to a plugin). Cobra resolves a name to the
			// first matching command, so drop birdsh's "cli" alias to let the plugin's command win.
			if pluginCommand.Name() == "cli" || pluginCommand.HasAlias("cli") {
				birdshCmd.Aliases = removeString(birdshCmd.Aliases, "cli")
			}
			rootCmd.AddCommand(pluginCommand)
		}
	}
}

// removeString returns s without any elements equal to r
func removeString(s []string, r string) []string {
	var out []string
	for _, v := range s {
		if v != r {
			out = append(out, v)
		}
	}
	return out
}

func Execute(v, c, d string) error {
	version = v
	commit = c
	date = d
	return rootCmd.Execute()
}
