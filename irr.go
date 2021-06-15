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
		return "", errors.New("Code error: getIRRPrefixSet family must be 4 or 6")
	}
	return fmt.Sprintf("PFXSET_%s_IP%d", *sanitize(asSet), family), nil
}


// Use bgpq4 to generate a prefix filter and return only the filter lines

func getIRRPrefixSet(asSet string, family uint8, c *config) ([]string, error) {
	// Run bgpq4 for BIRD format with aggregation enabled
	 filterName, err := asSetToFilterName(asSet, family)
	 if err != nil {
		  return []string{}, err
 	}
 	cmdArgs := fmt.Sprintf("-h %s -Ab%d -l %s %s", c.IRRServer, family, filterName, asSet)
 	log.Infof("Running bgpq4 %s", cmdArgs)
 	ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(cliFlags.IRRQueryTimeout))
 	defer cancel()
 	cmd := exec.CommandContext(ctx, "bgpq4", strings.Split(cmdArgs, " ")...)
 	stdout, err := cmd.Output()
	if err != nil {
		return []string{}, fmt.Errorf("bgpq4 error: %v, %s", err.Error(), stdout)
 	}
 	var out []string
	for i, line := range strings.Split(string(stdout), "\n") {
	if i == 0 { // skip first line, as it is the definition line
 		continue
  	}
  	if strings.Contains(line, "];") { // skip last line and stop
   		break
  	}
	out = append(out, strings.TrimRight(line, ","))
 	}
	return out, nil // nil error
}
