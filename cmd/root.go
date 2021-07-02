package cmd

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// Set by build process
var version = "dev"
var platform = "generic"
var commit = "unknown"
var date = "unknown"

var description = "Pathvector is a declarative routing control plane platform for BGP with robust filtering and route optimization."

var (
	configFile string
	verbose    bool
	rootCmd    = NewRootCommand()
)

func NewRootCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "pathvector",
		Short: description,
	}
}

// Execute executes the root command
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(func() {
		if verbose {
			log.SetLevel(log.DebugLevel)
		}
	})
	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "/etc/pathvector.yml", "config file")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "enable verbose logging")
}
