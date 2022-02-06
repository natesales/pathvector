package cmd

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"regexp"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/natesales/pathvector/pkg/bird"
)

var socket = ""

func init() {
	cliCmd.Flags().StringVarP(&socket, "socket", "s", "/var/run/bird/bird.ctl", "BIRD socket")
	rootCmd.AddCommand(cliCmd)
}

func read(r io.Reader) {
	resp, err := bird.Read(r)
	if err != nil {
		return
	}

	reg := regexp.MustCompile(`[0-9]{4}-? ?`)
	resp = reg.ReplaceAllString(resp, "")
	resp = strings.ReplaceAll(resp, "\n ", "\n")
	resp = strings.ReplaceAll(resp, "\n\n", "\n")
	resp = strings.TrimSuffix(resp, "\n")

	fmt.Println(resp)
}

var cliCmd = &cobra.Command{
	Use:   "cli",
	Short: "Interactive CLI",
	Run: func(cmd *cobra.Command, args []string) {
		c, err := net.Dial("unix", socket)
		if err != nil {
			log.Fatal(err)
		}
		defer c.Close()

		read(c)

		if len(args) > 0 {
			if _, err := c.Write([]byte(strings.Join(args, " ") + "\r\n")); err != nil {
				log.Fatalf("Unable to write to BIRD socket: %v", err)
			}
			read(c)
			return
		}

		r := bufio.NewReader(os.Stdin)
		for {
			fmt.Print("bird> ")
			cmd, _ := r.ReadString('\n')
			cmd = strings.Replace(cmd, "\n", "", -1)
			if cmd != "" {
				if _, err := c.Write([]byte(cmd + "\r\n")); err != nil {
					log.Fatalf("Unable to write to BIRD socket: %v", err)
				}
				read(c)
			}
		}
	},
}
