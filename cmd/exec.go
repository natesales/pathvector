package cmd

import (
	"bufio"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"io"
	"net/http"
	"net/url"
)

var (
	server string
)

func init() {
	execCmd.Flags().StringVarP(&server, "server", "s", "http://localhost:8084", "API endpoint")
	rootCmd.AddCommand(execCmd)
}

var execCmd = &cobra.Command{
	Use:     "exec",
	Short:   "Execute a remote pathvector command",
	Aliases: []string{"e"},
}

func execRemoteCommand(path string, params map[string]string) {
	u, err := url.Parse(server)
	if err != nil {
		log.Fatal(err)
	}
	u.Path = path
	q := u.Query()
	if verbose {
		q.Set("verbose", "true")
	}

	// Set query params
	for k, v := range params {
		q.Set(k, v)
	}
	u.RawQuery = q.Encode()

	log.Printf("Connecting to %s", u)
	resp, err := http.Get(u.String())
	if err != nil {
		log.Fatal(err)
	}

	reader := bufio.NewReader(resp.Body)
	for {
		line, err := reader.ReadBytes('\n')
		if err == io.EOF {
			return
		} else if err != nil {
			log.Fatal(err)
		}

		fmt.Print(string(line))
	}
}
