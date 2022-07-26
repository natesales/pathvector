package cmd

import (
	"net"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestBirdShell(t *testing.T) {
	unixSocket := "test.sock"

	// Delete socket
	_ = os.Remove(unixSocket)

	go func() {
		time.Sleep(time.Millisecond * 10) // Wait for the server to start
		rootCmd.SetArgs([]string{
			"birdsh",
			"-s", unixSocket,
		})
		if err := rootCmd.Execute(); err != nil {
			t.Error(err)
		}
	}()

	l, err := net.Listen("unix", unixSocket)
	assert.Nil(t, err)

	defer l.Close()
	conn, err := l.Accept()
	if err != nil {
		return
	}
	defer conn.Close()

	_, err = conn.Write([]byte("0001 Fake BIRD response 1"))
	assert.Nil(t, err)

	_, err = conn.Write([]byte("0001 Fake BIRD response 2"))
	assert.Nil(t, err)
}
