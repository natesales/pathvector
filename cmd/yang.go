package cmd

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/natesales/pathvector/pkg/yang"
)

func init() {
	rootCmd.AddCommand(yangCmd)
}

var yangCmd = &cobra.Command{
	Use:   "yang",
	Short: "Print YANG model of the configuration",
	Run: func(cmd *cobra.Command, args []string) {
		module, err := yang.Generate()
		if err != nil {
			log.Fatalf("Generating YANG module: %s", err)
		}
		fmt.Print(module)
	},
}
