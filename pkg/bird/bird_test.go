package bird

import (
	"net"
	"os"
	"strings"
	"testing"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestBirdConn(t *testing.T) {
	unixSocket := "test.sock"

	// Delete socket
	_ = os.Remove(unixSocket)

	go func() {
		time.Sleep(time.Millisecond * 10) // Wait for the server to start
		resp, _, err := RunCommand("bird command test\n", unixSocket)
		assert.Nil(t, err)

		// Print bird output as multiple lines
		for _, line := range strings.Split(strings.Trim(resp, "\n"), "\n") {
			log.Printf("BIRD response (multiline): %s", line)
		}
	}()

	log.Println("Starting fake BIRD socket server")
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

	buf := make([]byte, 1024)
	n, err := conn.Read(buf[:])
	assert.Nil(t, err)
	assert.Equal(t, "bird command test\n", string(buf[:n]))

	_, err = conn.Write([]byte("0001 Fake BIRD response 2"))
	assert.Nil(t, err)
}
