package main

import (
	"io"
	"net"
	"strings"

	"github.com/joomcode/errorx"
	log "github.com/sirupsen/logrus"
)

func birdRead(reader io.Reader) (string, error) {
	buf := make([]byte, 1024)
	n, err := reader.Read(buf[:])

	if err != nil {
		return "", errorx.Decorate(err, "BIRD read") // empty return string
	}

	return string(buf[:n]), nil // nil error
}

// runBirdCommand runs a bird command
func runBirdCommand(command string, socket string) error {
	log.Debug("Connecting to BIRD socket")
	conn, err := net.Dial("unix", socket)
	if err != nil {
		log.Fatalf("BIRD socket connect: %v", err)
	}
	//noinspection GoUnhandledErrorResult
	defer conn.Close()

	log.Println("Connected to BIRD socket")
	resp, err := birdRead(conn)
	if err != nil {
		return err
	}
	log.Printf("BIRD init response: %s", resp)

	log.Printf("Sending BIRD command: %s", command)
	_, err = conn.Write([]byte(strings.Trim(command, "\n") + "\n"))
	log.Printf("Sent BIRD command: %s", command)
	if err != nil {
		log.Fatalf("BIRD write error: %s\n", err)
	}

	resp, err = birdRead(conn)
	if err != nil {
		return err
	}

	// Print bird output as multiple lines
	for _, line := range strings.Split(strings.Trim(resp, "\n"), "\n") {
		log.Printf("BIRD response (multiline): %s", line)
	}

	return nil // nil error
}
