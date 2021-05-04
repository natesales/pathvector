package main

import (
	"io"
	"net"
	"strings"

	log "github.com/sirupsen/logrus"
)

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
	resp, err := io.ReadAll(conn)
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

	resp, err = io.ReadAll(conn)
	if err != nil {
		return err
	}

	// Print bird output as multiple lines
	for _, line := range strings.Split(strings.Trim(string(resp), "\n"), "\n") {
		log.Printf("BIRD response (multiline): %s", line)
	}

	return nil // nil error
}
