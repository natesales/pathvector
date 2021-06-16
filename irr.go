package main

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

// Use bgpq4 to generate a prefix filter and return only the filter lines
func getIRRPrefixSet(asSet string, family uint8, irrServer string) ([]string, error) {
	// Run bgpq4 for BIRD format with aggregation enabled
	cmdArgs := fmt.Sprintf("-h %s -Ab%d %s", irrServer, family, asSet)
	log.Debugf("Running bgpq4 %s", cmdArgs)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(10))
	defer cancel()
	cmd := exec.CommandContext(ctx, "bgpq4", strings.Split(cmdArgs, " ")...)
	stdout, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var prefixes []string
	for i, line := range strings.Split(string(stdout), "\n") {
		if i == 0 { // Skip first line, as it is the definition line
			continue
		}
		if strings.Contains(line, "];") { // Skip last line and return
			break
		}
		// Trim whitespace and remove the comma, then append to the prefixes slice
		prefixes = append(prefixes, strings.TrimSpace(strings.TrimRight(line, ",")))
	}

	return prefixes, nil
}
