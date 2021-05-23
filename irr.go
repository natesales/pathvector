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
func asSetToFilterName(asSet string, family uint8) string {
	if !(family == 4 || family == 6) {
		log.Fatal("code error: getIRRPrefixSet family must be 4 or 6")
	}

	return fmt.Sprintf("PFXSET_%s_IP%d", strings.Replace(strings.Replace(asSet, ":", "_", -1), "-", "_", -1), family)
}

// Use bgpq4 to generate a prefix filter and return only the filter lines
func getIRRPrefixSet(asSet string, family uint8, irrdb string, c *config) (string, error) {
	if !(family == 4 || family == 6) {
		log.Fatal("code error: getIRRPrefixSet family must be 4 or 6")
	}

	// Run bgpq4 for BIRD format with aggregation enabled
	cmdArgs := fmt.Sprintf("-h %s -Ab%d %s -l %s", irrdb, family, asSet, asSetToFilterName(asSet, family))
	log.Infof("Running bgpq4 %s", cmdArgs)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(c.IRRQueryTimeout))
	defer cancel()
	cmd := exec.CommandContext(ctx, "bgpq4", strings.Split(cmdArgs, " ")...)
	stdout, err := cmd.Output()
	if err != nil {
		return "", errors.New(fmt.Sprintf("bgpq4 error: %v, %s", err.Error(), stdout))
	}
	return "define " + string(stdout), nil // nil error
}
