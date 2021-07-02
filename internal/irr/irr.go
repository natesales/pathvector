package irr

import (
	"context"
	"fmt"
	"github.com/natesales/pathvector/internal/config"
	"os/exec"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

// Use bgpq4 to generate a prefix filter and return only the filter lines
func getIRRPrefixSet(asSet string, family uint8, irrServer string, timeout uint) ([]string, error) {
	// Run bgpq4 for BIRD format with aggregation enabled
	cmdArgs := fmt.Sprintf("-h %s -Ab%d %s", irrServer, family, asSet)
	log.Debugf("Running bgpq4 %s", cmdArgs)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(timeout))
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

func BuildPrefixSet(peerData *config.Peer, irrServer string, timeout uint) error {
	// Check for empty as-set
	if peerData.ASSet == nil || *peerData.ASSet == "" {
		return fmt.Errorf("peer has filter-irr enabled and no as-set defined")
	}

	prefixesFromIRR4, err := getIRRPrefixSet(*peerData.ASSet, 4, irrServer, timeout)
	if err != nil {
		return fmt.Errorf("unable to get IRR prefix list from %s", *peerData.ASSet)
	}
	if peerData.PrefixSet4 == nil {
		peerData.PrefixSet4 = &[]string{}
	}
	pfx4 := append(*peerData.PrefixSet4, prefixesFromIRR4...)
	peerData.PrefixSet4 = &pfx4
	if len(pfx4) == 0 {
		return fmt.Errorf("peer has a prefix filter defined but no IPv4 prefixes")
	}

	prefixesFromIRR6, err := getIRRPrefixSet(*peerData.ASSet, 6, irrServer, timeout)
	if err != nil {
		return fmt.Errorf("unable to get IRR prefix list from %s", *peerData.ASSet)
	}
	if peerData.PrefixSet6 == nil {
		peerData.PrefixSet6 = &[]string{}
	}
	pfx6 := append(*peerData.PrefixSet6, prefixesFromIRR6...)
	peerData.PrefixSet6 = &pfx6
	if len(pfx6) == 0 {
		return fmt.Errorf("peer has a prefix filter defined but no IPv6 prefixes")
	}

	return nil // nil error
}
