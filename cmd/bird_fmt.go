package cmd

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/natesales/pathvector/pkg/bird"
	"github.com/natesales/pathvector/pkg/util/log"
)

func init() {
	rootCmd.AddCommand(birdFmtCmd)
}

func isDir(path string) (bool, error) {
	fi, err := os.Stat(path)
	if err != nil {
		return false, err
	}

	return fi.IsDir(), nil
}

func formatFile(path string) error {
	if !strings.HasSuffix(path, ".conf") {
		log.Debugf("Skipping %s", path)
		return nil
	}

	log.Infof("Formatting %s", path)

	unformatted, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	formatted := bird.Reformat(string(unformatted))

	if err := os.WriteFile(path, []byte(formatted), 0644); err != nil {
		return err
	}

	return nil
}

var birdFmtCmd = &cobra.Command{
	Use:   "bird-fmt [file/directory]",
	Short: "Format BIRD config",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			log.Fatal("No file/directory specified")
		}

		target := args[0]

		// Check if directory
		dir, err := isDir(target)
		if err != nil {
			log.Fatal(err)
		}
		if dir {
			// For file in walk directory
			if err := filepath.Walk(target, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					log.Fatal(err)
				}
				return formatFile(path)
			}); err != nil {
				log.Fatal(err)
			}
		} else {
			if err := formatFile(target); err != nil {
				log.Fatal(err)
			}
		}
	},
}
