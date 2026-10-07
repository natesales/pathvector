package cmd

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"path"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

// fakeBIRDServer serves a fake BIRD control socket and records the commands it receives
type fakeBIRDServer struct {
	l        net.Listener
	mu       sync.Mutex
	commands []string
}

func newFakeBIRDServer(t *testing.T, socket string) *fakeBIRDServer {
	l, err := net.Listen("unix", socket)
	if err != nil {
		t.Fatal(err)
	}
	s := &fakeBIRDServer{l: l}
	go func() {
		for {
			conn, err := l.Accept()
			if err != nil {
				return
			}
			go s.handle(conn)
		}
	}()
	t.Cleanup(func() { _ = l.Close() })
	return s
}

func (s *fakeBIRDServer) handle(conn net.Conn) {
	defer conn.Close()
	if _, err := conn.Write([]byte("0001 BIRD 2.14 ready.\n")); err != nil {
		return
	}
	r := bufio.NewReader(conn)
	for {
		line, err := r.ReadString('\n')
		if err != nil {
			return
		}
		command := strings.TrimSpace(line)
		s.mu.Lock()
		s.commands = append(s.commands, command)
		s.mu.Unlock()

		var resp string
		switch {
		case command == "show protocols all":
			resp = "2002-Name       Proto      Table      State  Since         Info\n" +
				"1002-EXAMPLE_AS65510_v4 BGP        ---        up     2022-01-01 00:00:00  Established\n" +
				"1006-  BGP state:          Established\n" +
				"    Neighbor address: 203.0.113.12\n" +
				"    Neighbor AS:      65510\n" +
				"    Local AS:         34553\n" +
				" \n0000 \n"
		case strings.HasPrefix(command, "restart "):
			resp = fmt.Sprintf("0012-%s: restarted\n0000 \n", strings.Trim(strings.TrimPrefix(command, "restart "), `"`))
		default:
			resp = "1000-BIRD 2.14\n0013 Daemon is up and running\n"
		}
		if _, err := conn.Write([]byte(resp)); err != nil {
			return
		}
	}
}

func (s *fakeBIRDServer) Commands() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]string{}, s.commands...)
}

// TestCommandsUseConfiguredSocket checks that commands talking to BIRD use the bird-socket and bird-directory config options
func TestCommandsUseConfiguredSocket(t *testing.T) {
	dir := t.TempDir()
	socket := path.Join(dir, "bird.ctl")
	server := newFakeBIRDServer(t, socket)

	//nolint:golint,gosec
	assert.Nil(t, os.WriteFile(path.Join(dir, "protocols.json"), []byte(`{"EXAMPLE_AS65510_v4":{"Name":"Example","Tags":null},"EXAMPLE_AS65510_v6":{"Name":"Example","Tags":null}}`), 0644))
	conf := path.Join(dir, "pathvector.yml")
	//nolint:golint,gosec
	assert.Nil(t, os.WriteFile(conf, []byte(fmt.Sprintf(`asn: 34553
router-id: 192.0.2.1
bird-socket: %s
bird-directory: %s
cache-directory: %s
no-announce: true
`, socket, dir, path.Join(dir, "cache"))), 0644))

	for _, args := range [][]string{
		{"version"},
		{"config"},
		{"status"},
		{"restart", "Example"},
		{"birdsh", "show", "status"},
	} {
		rootCmd.SetArgs(append(args, "-c", conf))
		assert.Nil(t, rootCmd.Execute(), args)
	}

	assert.Equal(t, []string{
		"show protocols all",
		`restart "EXAMPLE_AS65510_v4"`,
		`restart "EXAMPLE_AS65510_v6"`,
		"show status",
	}, server.Commands())
}
