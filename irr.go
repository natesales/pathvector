package main

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

// asSetToFilterName converts an as-set into a BIRD-safe filter name
func asSetToFilterName(asSet string, family uint8) (string, error) {
	if !(family == 4 || family == 6) {
		return "", errors.New("code error: getIRRPrefixSet family must be 4 or 6")
	}
	return fmt.Sprintf("PFXSET_%s_IP%d", sanitize(asSet), family), nil
}

// Use bgpq4 to generate a prefix filter and return only the filter lines
func getIRRPrefixSet(asSet string, family uint8, c *config) (string, error) {
	// Run bgpq4 for BIRD format with aggregation enabled
	filterName, err := asSetToFilterName(asSet, family)
	if err != nil {
		return "", err
	}
	cmdArgs := fmt.Sprintf("-h %s -Ab%d %s -l %s", c.IRRServer, family, asSet, filterName)
	log.Infof("Running bgpq4 %s", cmdArgs)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(cliFlags.IRRQueryTimeout))
	defer cancel()
	cmd := exec.CommandContext(ctx, "bgpq4", strings.Split(cmdArgs, " ")...)
	stdout, err := cmd.Output()
	if err != nil {
		return "", errors.New(fmt.Sprintf("bgpq4 error: %v, %s", err.Error(), stdout))
	}
	return "define " + string(stdout), nil // nil error
}
