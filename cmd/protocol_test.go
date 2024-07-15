package cmd

import "testing"

func TestProtocol(t *testing.T) {
        rootCmd.SetArgs([]string{
                "protocol",
                "reload",
                "device1",
        })
        if err := rootCmd.Execute(); err != nil {
                t.Error(err)
        }
}
