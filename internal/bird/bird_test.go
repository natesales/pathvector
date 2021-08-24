package bird

import (
	"net"
	"os"
	"strings"
	"testing"
	"time"

	log "github.com/sirupsen/logrus"
)

func TestBirdConn(t *testing.T) {
	unixSocket := "test.sock"

	// Delete socket
	_ = os.Remove(unixSocket)

	go func() {
		time.Sleep(time.Millisecond * 10) // Wait for the server to start
		resp, err := RunCommand("bird command test\n", unixSocket)
		if err != nil {
			t.Error(err)
		}
		// Print bird output as multiple lines
		for _, line := range strings.Split(strings.Trim(resp, "\n"), "\n") {
			log.Printf("BIRD response (multiline): %s", line)
		}
	}()

	log.Println("Starting fake BIRD socket server")
	l, err := net.Listen("unix", unixSocket)
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()
	conn, err := l.Accept()
	if err != nil {
		return
	}
	defer conn.Close()

	if _, err := conn.Write([]byte("0001 Fake BIRD response 1")); err != nil {
		t.Error(err)
	}

	buf := make([]byte, 1024)
	n, err := conn.Read(buf[:])
	if err != nil {
		t.Error(err)
	}
	if string(buf[:n]) != "bird command test\n" {
		t.Errorf("expected 'bird command test' got %s", string(buf[:n]))
	}

	if _, err := conn.Write([]byte("0001 Fake BIRD response 2")); err != nil {
		t.Error(err)
	}
}
