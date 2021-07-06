package main

import (
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"
)

func birdRead(reader io.Reader) (string, error) {
	buf := make([]byte, 1024)
	n, err := reader.Read(buf[:])

	if err != nil {
		return "", fmt.Errorf("BIRD read: %v", err)
	}

	return string(buf[:n]), nil // nil error
}

// runBirdCommand runs a bird command
func runBirdCommand(command string, socket string) error {
	log.Debugln("Connecting to BIRD socket")
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
	log.Debugf("BIRD init response: %s", resp)

	log.Debugf("Sending BIRD command: %s", command)
	_, err = conn.Write([]byte(strings.Trim(command, "\n") + "\n"))
	log.Debugf("Sent BIRD command: %s", command)
	if err != nil {
		log.Fatalf("BIRD write error: %s\n", err)
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

// birdValidate checks if the cached configuration is syntactically valid
func birdValidate() {
	log.Debugln("Validating BIRD config")
	birdCmd := exec.Command(birdBinary, "-c", "bird.conf", "-p")
	birdCmd.Dir = cacheDirectory
	birdCmd.Stdout = os.Stdout
	birdCmd.Stderr = os.Stderr
	if err := birdCmd.Run(); err != nil {
		log.Fatalf("BIRD config validation: %v", err)
	}
	log.Infof("BIRD config validation passed")
}

// moveCacheAndReconfig moves cached files to the production BIRD directory and reconfigures
func moveCacheAndReconfig() {
	// Remove old configs
	birdConfigFiles, err := filepath.Glob(path.Join(birdDirectory, "AS*.conf"))
	if err != nil {
		log.Fatal(err)
	}
	for _, f := range birdConfigFiles {
		log.Debugf("Removing old BIRD config file %s", f)
		if err := os.Remove(f); err != nil {
			log.Fatalf("Removing old BIRD config files: %v", err)
		}
	}

	// Copy from cache to bird config
	files, err := filepath.Glob(path.Join(cacheDirectory, "*.conf"))
	if err != nil {
		log.Fatal(err)
	}
	for _, f := range files {
		fileNameParts := strings.Split(f, "/")
		fileNameTail := fileNameParts[len(fileNameParts)-1]
		newFileLoc := path.Join(birdDirectory, fileNameTail)
		log.Debugf("Moving %s to %s", f, newFileLoc)
		if err := moveFile(f, newFileLoc); err != nil {
			log.Fatalf("Moving cache file to bird directory: %v", err)
		}
	}

	if !noConfigure {
		log.Infoln("Reconfiguring BIRD")
		if err = runBirdCommand("configure", birdSocket); err != nil {
			log.Fatal(err)
		}
	} else {
		log.Infoln("Option --no-configure is set, NOT reconfiguring bird")
	}
}
