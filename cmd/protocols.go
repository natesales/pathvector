package cmd

import (
	"encoding/json"
	"os"
	"path"
	"sort"

	"github.com/creasty/defaults"
	log "github.com/sirupsen/logrus"

	"github.com/natesales/pathvector/pkg/config"
	"github.com/natesales/pathvector/pkg/templating"
)

// birdPaths returns the BIRD control socket and BIRD config directory from c,
// falling back to the config defaults when c is nil or the values are unset
func birdPaths(c *config.Config) (socket string, directory string) {
	var def config.Config
	if err := defaults.Set(&def); err != nil {
		log.Warnf("Setting config defaults: %v", err)
	}
	socket, directory = def.BIRDSocket, def.BIRDDirectory
	if c != nil {
		if c.BIRDSocket != "" {
			socket = c.BIRDSocket
		}
		if c.BIRDDirectory != "" {
			directory = c.BIRDDirectory
		}
	}
	return socket, directory
}

// readProtocolNames reads the BIRD protocol name to peer name map (protocols.json) from the BIRD directory
func readProtocolNames(birdDirectory string) (map[string]*templating.Protocol, error) {
	contents, err := os.ReadFile(path.Join(birdDirectory, "protocols.json"))
	if err != nil {
		return nil, err
	}
	var protocols map[string]*templating.Protocol
	if err := json.Unmarshal(contents, &protocols); err != nil {
		return nil, err
	}
	return protocols, nil
}

// resolveProtocols maps each name to BIRD protocol names. A name matching a peer name from
// protocols maps to all BIRD protocols of that peer; any other name is taken as a BIRD protocol name.
// The result is deduplicated and keeps the order of names.
func resolveProtocols(names []string, protocols map[string]*templating.Protocol) []string {
	var out []string
	seen := map[string]bool{}
	add := func(p string) {
		if !seen[p] {
			seen[p] = true
			out = append(out, p)
		}
	}

	for _, name := range names {
		var matches []string
		for birdName, p := range protocols {
			if p != nil && p.Name == name {
				matches = append(matches, birdName)
			}
		}
		sort.Strings(matches)
		for _, m := range matches {
			add(m)
		}
		// Real BIRD protocol name (or no matching peer name)
		if _, isBIRDName := protocols[name]; isBIRDName || len(matches) == 0 {
			add(name)
		}
	}
	return out
}
