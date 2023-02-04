package irr

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/natesales/pathvector/pkg/config"
)

// FirstASSet picks the first AS set if there are multiple
func FirstASSet(asSet string) string {
	output := asSet

	// If the as-set has a space in it, split and pick the first one
	if strings.Contains(output, " ") {
		log.Warnf("Original AS set %s contains a space. Selecting first element %s", asSet, output)
		output = strings.Split(output, " ")[0]
	}

	return output
}

// withSourceFilter returns the AS set or AS set with the IRR source replaced with the -S SOURCE syntax
// AS34553 -> AS34553
// RIPE::AS34553 -> -S RIPE AS34553
func withSourceFilter(asSet string) string {
	if strings.Contains(asSet, "::") {
		log.Debugf("Using IRRDB source from AS set %s", asSet)
		tokens := strings.Split(asSet, "::")
		return fmt.Sprintf("-S %s %s", tokens[0], tokens[1])
	}
	return asSet
}

// PrefixSet uses bgpq4 to generate a prefix filter and return only the filter lines
func PrefixSet(asSet string, family uint8, irrServer string, queryTimeout uint, bgpqArgs string) ([]string, error) {
	// Run bgpq4 for BIRD format with aggregation enabled
	cmdArgs := fmt.Sprintf("-h %s -Ab%d %s", irrServer, family, withSourceFilter(asSet))
	if bgpqArgs != "" {
		cmdArgs = bgpqArgs + " " + cmdArgs
	}
	log.Debugf("Running bgpq4 %s", cmdArgs)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(queryTimeout))
	defer cancel()
	//nolint:golint,gosec
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

// ASMembers uses bgpq4 to generate an AS member set
func ASMembers(asSet string, irrServer string, queryTimeout uint, bgpqArgs string) ([]uint32, error) {
	// Run bgpq4 for BIRD format with aggregation enabled
	cmdArgs := fmt.Sprintf("-h %s -tj %s", irrServer, withSourceFilter(asSet))
	if bgpqArgs != "" {
		cmdArgs = bgpqArgs + " " + cmdArgs
	}
	log.Debugf("Running bgpq4 %s", cmdArgs)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(queryTimeout))
	defer cancel()
	//nolint:golint,gosec
	cmd := exec.CommandContext(ctx, "bgpq4", strings.Split(cmdArgs, " ")...)
	stdout, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var r map[string][]uint32
	if err := json.Unmarshal(stdout, &r); err != nil {
		return nil, fmt.Errorf("bgpq4 JSON Unmarshal: %s", err)
	}

	return r["NN"], nil
}

// Update updates a peer's IRR prefix set
func Update(peerData *config.Peer, irrServer string, queryTimeout uint, bgpqArgs string) error {
	// Check for empty as-set
	if peerData.ASSet == nil || *peerData.ASSet == "" {
		return fmt.Errorf("peer has filter-irr enabled and no as-set defined")
	}

	// Does the peer have any IPv4 or IPv6 neighbors?
	var hasNeighbor4, hasNeighbor6 bool
	if peerData.NeighborIPs != nil {
		for _, n := range *peerData.NeighborIPs {
			if strings.Contains(n, ".") {
				hasNeighbor4 = true
			} else if strings.Contains(n, ":") {
				hasNeighbor6 = true
			} else {
				log.Fatalf("Invalid neighbor IP %s", n)
			}
		}
	}

	// Handle acceptChildPrefixes
	bgpqArgs4 := bgpqArgs
	bgpqArgs6 := bgpqArgs
	if peerData.IRRAcceptChildPrefixes != nil && *peerData.IRRAcceptChildPrefixes {
		if bgpqArgs4 != "" {
			bgpqArgs4 += " "
		}
		bgpqArgs4 += "-R 24"

		if bgpqArgs6 != "" {
			bgpqArgs6 += " "
		}
		bgpqArgs6 += "-R 48"
	}

	prefixesFromIRR4, err := PrefixSet(*peerData.ASSet, 4, irrServer, queryTimeout, bgpqArgs4)
	if err != nil {
		return fmt.Errorf("unable to get IPv4 IRR prefix list from %s: %s", *peerData.ASSet, err)
	}
	if peerData.PrefixSet4 == nil {
		peerData.PrefixSet4 = &[]string{}
	}
	pfx4 := append(*peerData.PrefixSet4, prefixesFromIRR4...)
	peerData.PrefixSet4 = &pfx4
	if len(pfx4) == 0 && hasNeighbor4 {
		log.Warnf("peer has IPv4 session(s) but no IPv4 prefixes")
	}

	prefixesFromIRR6, err := PrefixSet(*peerData.ASSet, 6, irrServer, queryTimeout, bgpqArgs6)
	if err != nil {
		return fmt.Errorf("unable to get IPv6 IRR prefix list from %s: %s", *peerData.ASSet, err)
	}
	if peerData.PrefixSet6 == nil {
		peerData.PrefixSet6 = &[]string{}
	}
	pfx6 := append(*peerData.PrefixSet6, prefixesFromIRR6...)
	peerData.PrefixSet6 = &pfx6
	if len(pfx6) == 0 && hasNeighbor6 {
		log.Warnf("peer has IPv6 session(s) but no IPv6 prefixes")
	}

	return nil // nil error
}
