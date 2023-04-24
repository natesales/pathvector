package bird

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
	"golang.org/x/mod/semver"

	"github.com/natesales/pathvector/pkg/util"
)

// Minimum supported BIRD version
const supportedMin = "2.0.7"

// Read reads from an io.Reader
func Read(reader io.Reader) (string, error) {
	// TODO: This buffer isn't a good solution, and might not fit the full response from BIRD
	buf := make([]byte, 16384)
	n, err := reader.Read(buf[:])

	if err != nil {
		return "", fmt.Errorf("BIRD read: %v", err)
	}

	return string(buf[:n]), nil // nil error
}

// ReadClean reads from the provided reader and trims unneeded whitespace and bird 4-digit numbers
func ReadClean(r io.Reader) {
	resp, err := Read(r)
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

// RunCommand runs a BIRD command and returns the output, version, and error
func RunCommand(command string, socket string) (string, string, error) {
	log.Debugln("Connecting to BIRD socket")
	conn, err := net.Dial("unix", socket)
	if err != nil {
		return "", "", err
	}
	//noinspection GoUnhandledErrorResult
	defer conn.Close()

	log.Debug("Connected to BIRD socket")
	resp, err := Read(conn)
	if err != nil {
		return "", "", err
	}
	log.Debugf("BIRD init response: %s", resp)

	// Check BIRD version
	birdVersion := strings.Split(resp, " ")[2]
	if semver.Compare(birdVersion, supportedMin) == -1 {
		log.Warnf("BIRD version %s older than minimum supported version %s", birdVersion, supportedMin)
	}

	log.Debugf("Sending BIRD command: %s", command)
	_, err = conn.Write([]byte(strings.Trim(command, "\n") + "\n"))
	log.Debugf("Sent BIRD command: %s", command)
	if err != nil {
		return "", "", err
	}

	log.Debugln("Reading from socket")
	resp, err = Read(conn)
	if err != nil {
		return "", "", err
	}
	log.Debugln("Done reading from socket")

	return resp, birdVersion, nil // nil error
}

// Validate checks if the cached configuration is syntactically valid
func Validate(binary string, cacheDir string) {
	log.Debugf("Validating BIRD config")
	var outb, errb bytes.Buffer
	birdCmd := exec.Command(binary, "-c", "bird.conf", "-p")
	birdCmd.Dir = cacheDir
	birdCmd.Stdout = &outb
	birdCmd.Stderr = &errb
	var errbT string
	if err := birdCmd.Run(); err != nil {
		origErr := err
		errbT = strings.TrimSuffix(errb.String(), "\n")

		// Check for validation error in format:
		// bird: ./AS65530_EXAMPLE.conf:20:43 syntax error, unexpected '%'
		match, err := regexp.MatchString(`bird:.*:\d+:\d+.*`, errbT)
		if err != nil {
			log.Fatalf("BIRD error regex match: %s", err)
		}
		errorMessageToLog := errbT
		if match {
			errorMessageToLog = "BIRD validation error:\n" // Clear error message so we can write the new nicely formatted one
			respPartsSpace := strings.Split(errbT, " ")
			respPartsColon := strings.Split(respPartsSpace[1], ":")
			errorMessage := strings.Join(respPartsSpace[2:], " ")
			errorFile := respPartsColon[0]
			errorLine, err := strconv.Atoi(respPartsColon[1])
			if err != nil {
				log.Fatalf("BIRD error line int parse: %s", err)
			}
			errorChar, err := strconv.Atoi(respPartsColon[2])
			if err != nil {
				log.Fatalf("BIRD error line int parse: %s", err)
			}
			log.Debugf("Found error in %s:%d:%d message %s", errorFile, errorLine, errorChar, errorMessage)

			// Read output file
			file, err := os.Open(path.Join(cacheDir, errorFile))
			if err != nil {
				log.Fatalf("unable to read BIRD output file for error parsing: %s", err)
			}
			defer file.Close()

			scanner := bufio.NewScanner(file)
			line := 1
			for scanner.Scan() {
				if (line >= errorLine-1) && (line <= errorLine+1) { // Print one line above and below the error line
					errorMessageToLog += scanner.Text() + "\n"
				}
				if line == errorLine {
					errorMessageToLog += strings.Repeat(" ", errorChar-1) + "^ " + errorMessage + "\n"
				}
				line++
			}
			if err := scanner.Err(); err != nil {
				log.Fatalf("BIRD output file scan: %s", err)
			}
		}
		if errorMessageToLog == "" {
			errorMessageToLog = origErr.Error()
		}
		log.Fatalf("BIRD: %s\n", errorMessageToLog)
	}

	log.Infof("BIRD config validation passed")
}

// MoveCacheAndReconfigure moves cached files to the production BIRD directory and reconfigures
func MoveCacheAndReconfigure(birdDirectory string, cacheDirectory string, birdSocket string, noConfigure bool) {
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
		if err := util.MoveFile(f, newFileLoc); err != nil {
			log.Fatalf("Moving cache file to bird directory: %v", err)
		}
	}

	if !noConfigure {
		log.Info("Reconfiguring BIRD")
		resp, _, err := RunCommand("configure", birdSocket)
		if err != nil {
			log.Fatal(err)
		}
		// Print bird output as multiple lines
		for _, line := range strings.Split(strings.Trim(resp, "\n"), "\n") {
			log.Printf("BIRD response (multiline): %s", line)
		}
	}
}

// Reformat takes a BIRD config file as a string and outputs a nicely formatted version as a string
func Reformat(input string) string {
	formatted := ""
	for _, line := range strings.Split(input, "\n") {
		if strings.HasSuffix(line, "{") || strings.HasSuffix(line, "[") {
			formatted += "\n"
		}

		if !func(input string) bool {
			for _, chr := range input {
				if string(chr) != " " {
					return false
				}
			}
			return true
		}(line) {
			formatted += line + "\n"
		}
	}
	return formatted
}
