package cmd

import (
	"github.com/spf13/cobra"
	"strconv"
)

var (
	asn uint32
)

func init() {
	execReloadCmd.Flags().Uint32VarP(&asn, "asn", "a", 0, "ASN to reload (0 for all ASNs)")
	execCmd.AddCommand(execReloadCmd)
}

var execReloadCmd = &cobra.Command{
	Use:     "reload",
	Short:   "Reload the current configuration",
	Aliases: []string{"r"},
	Run: func(cmd *cobra.Command, args []string) {
		execRemoteCommand("/reload", map[string]string{"asn": strconv.Itoa(int(asn))})
	},
}
