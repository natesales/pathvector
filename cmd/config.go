package cmd

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/natesales/logknife"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/natesales/pathvector/pkg/bird"
	"github.com/natesales/pathvector/pkg/process"
)

var sensitiveKeys = []string{
	"peeringdb-api-key",
	"password",
}

var sanitize bool

func init() {
	configCmd.Flags().BoolVarP(&sanitize, "sanitize", "s", false, "Sanitize sensitive information from config")
	rootCmd.AddCommand(configCmd)
}

// uname runs the "uname -a" command and returns the output
func uname() string {
	out, err := exec.Command("uname", "-a").Output()
	if err != nil {
		log.Warnf("uname: %s", err)
		return "unknown"
	}
	return string(out)
}

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Export configuration, optionally sanitized with logknife",
	Run: func(cmd *cobra.Command, args []string) {
		log.Debugf("Loading config from %s", configFile)
		cf, err := os.ReadFile(configFile)
		if err != nil {
			log.Fatalf("Reading config file: %s", err)
		}
		config := string(cf)

		var buf string

		birdVersion := "unknown"
		c, err := process.Load(cf)
		if err != nil {
			buf += fmt.Sprintf("# Error loading config: %s\n", err)
		} else {
			log.Debug("Finished loading config")
			_, birdVersion, err = bird.RunCommand("", c.BIRDSocket)
			if err != nil {
				birdVersion = fmt.Sprintf("error: %s", err)
			}
		}

		for _, line := range strings.Split(versionBanner(), "\n") {
			buf += fmt.Sprintf("# %s\n", line)
		}
		buf += "# BIRD " + birdVersion + "\n"
		buf += fmt.Sprintf("# System %s", uname())
		if sanitize {
			buf += "# Sanitized config"
		} else {
			buf += "# Config"
		}
		buf += fmt.Sprintf(" exported from %s on %s\n", configFile, time.Now().Format(time.RFC822Z))
		fmt.Println(buf)

		if sanitize {
			// Apply sanitized keys
			for _, key := range sensitiveKeys {
				re := regexp.MustCompile(fmt.Sprintf(`(.*)%s:.*`, key))
				config = re.ReplaceAllString(config, fmt.Sprintf("${1}%s: REDACTED", key))
			}

			logknife.Knife(bytes.NewBuffer([]byte(config)), false, true, false, "")
		} else {
			fmt.Print(config)
		}
	},
}
