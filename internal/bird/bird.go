package bird

import (
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/natesales/pathvector/internal/config"
)

func birdRead(reader io.Reader) (string, error) {
	buf := make([]byte, 1024)
	n, err := reader.Read(buf[:])

	if err != nil {
		return "", fmt.Errorf("BIRD read: %v", err)
	}

	return string(buf[:n]), nil // nil error
}

// Run runs a bird command
func Run(command string, socket string, timeout uint) error {
	log.Debugf("Connecting to BIRD socket with timeout %d", timeout)
	conn, err := net.DialTimeout("unix", socket, time.Duration(timeout)*time.Second)
	if err != nil {
		return fmt.Errorf("BIRD socket connect: %v", err)
	}
	//noinspection GoUnhandledErrorResult
	defer conn.Close()

	log.Println("Connected to BIRD socket")
	resp, err := birdRead(conn)
	if err != nil {
		return err
	}
	log.Debugf("BIRD init response: %s", resp)

	log.Debugf("Sending BIRD command: %s", command)
	_, err = conn.Write([]byte(strings.Trim(command, "\n") + "\n"))
	log.Debugf("Sent BIRD command: %s", command)
	if err != nil {
		return fmt.Errorf("BIRD write error: %s\n", err)
	}

	log.Debugln("Reading from socket")
	resp, err = birdRead(conn)
	if err != nil {
		return err
	}
	log.Debugln("Done reading from socket")

	// Print bird output as multiple lines
	for _, line := range strings.Split(strings.Trim(resp, "\n"), "\n") {
		log.Printf("BIRD response (multiline): %s", line)
	}

	return nil // nil error
}

// Validate runs BIRD for config validation
func Validate(global *config.Global) error {
	cmd := exec.Command(global.BirdBinary, "-c", "bird.conf", "-p")
	cmd.Dir = global.CacheDirectory
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
