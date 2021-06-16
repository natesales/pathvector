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
	log.Printf("Running bgpq4 %s", cmdArgs)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(10))
	defer cancel()
	cmd := exec.CommandContext(ctx, "bgpq4", strings.Split(cmdArgs, " ")...)
	stdout, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	// Remove whitespace and commas from output
	output := strings.ReplaceAll(string(stdout), ",\n    ", "\n")

	// Remove array prefix
	output = strings.ReplaceAll(output, "NN = [\n    ", "")

	// Remove array suffix
	output = strings.ReplaceAll(output, "];", "")

	// Remove whitespace (in this case there should only be trailing whitespace)
	output = strings.TrimSpace(output)

	// Split output by newline
	return strings.Split(output, "\n"), nil
}
