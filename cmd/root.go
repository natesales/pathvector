package cmd

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// Build process flags
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
	dryRun      bool
	noConfigure bool
)

// CLI Commands
var rootCmd = &cobra.Command{
	Use:   "pathvector",
	Short: "Pathvector is a declarative BGP routing platform that automates route optimization and control plane configuration with secure and repeatable routing policy.",
}

func init() {
	cobra.OnInitialize(func() {
		if verbose {
			log.SetLevel(log.DebugLevel)
		}
	})
	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "/etc/pathvector.yml", "Configuration file in YAML, TOML, or JSON format")
	rootCmd.PersistentFlags().StringVar(&lockFile, "lock", "", "Lock file (check disabled if empty)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Show verbose log messages")
	rootCmd.PersistentFlags().BoolVarP(&dryRun, "dry-run", "d", false, "Don't modify configuration")
	rootCmd.PersistentFlags().BoolVarP(&noConfigure, "no-configure", "n", false, "Don't configure BIRD")
}

func Execute(v string, c string, d string) error {
	version = v
	commit = c
	date = d
	return rootCmd.Execute()
}
