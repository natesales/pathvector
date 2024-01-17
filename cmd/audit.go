package cmd

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/natesales/pathvector/pkg/audit"
)

func init() {
	rootCmd.AddCommand(auditCmd)
}

var auditCmd = &cobra.Command{
	Use:   "audit",
	Short: "Audit configuration for common misconfigurations",
	Run: func(cmd *cobra.Command, args []string) {
		c, err := loadConfig()
		if err != nil {
			log.Fatal(err)
		}
		alerts := audit.Check(c)
		log.Infof("%d alerts detected", len(alerts))
	},
}
