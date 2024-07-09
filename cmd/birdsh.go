package cmd

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/natesales/pathvector/pkg/bird"
)

var socket = ""

func init() {
	birdshCmd.Flags().StringVarP(&socket, "socket", "s", "", "BIRD socket file, will read from bird-socket config option if empty")
	rootCmd.AddCommand(birdshCmd)
}

var birdshCmd = &cobra.Command{
	Use:   "birdsh",
	Short: "Lightweight BIRD shell",
	Run: func(cmd *cobra.Command, args []string) {
		if socket == "" {
			conf, err := loadConfig()
			if err != nil {
				log.Fatal(err)
			}
			socket = conf.BIRDSocket
		}

		c, err := net.Dial("unix", socket)
		if err != nil {
			log.Fatal(err)
		}
		defer c.Close()

		bird.ReadClean(c)

		if len(args) > 0 {
			if _, err := c.Write([]byte(strings.Join(args, " ") + "\r\n")); err != nil {
				log.Fatalf("Unable to write to BIRD socket: %v", err)
			}
			bird.ReadClean(c)
			return
		}

		r := bufio.NewReader(os.Stdin)
		for {
			fmt.Print("bird> ")
			cmd, _ := r.ReadString('\n')
			cmd = strings.ReplaceAll(cmd, "\n", "")
			if cmd != "" {
				if _, err := c.Write([]byte(cmd + "\r\n")); err != nil {
					log.Fatalf("Unable to write to BIRD socket: %v", err)
				}
				bird.ReadClean(c)
			}
		}
	},
}
