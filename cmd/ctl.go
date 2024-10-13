package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/lithammer/fuzzysearch/fuzzy"

	"github.com/natesales/pathvector/pkg/templating"
	"github.com/natesales/pathvector/pkg/util/log"
)

func protocols(birdDirectory string) (map[string]*templating.Protocol, error) {
	// Read protocol names map
	var protos = map[string]*templating.Protocol{}
	contents, err := os.ReadFile(path.Join(birdDirectory, "protocols.json"))
	if err != nil {
		return nil, fmt.Errorf("reading protocol names: %v", err)
	}
	if err := json.Unmarshal(contents, &protos); err != nil {
		return nil, fmt.Errorf("unmarshalling protocol names: %v", err)
	}

	return protos, nil
}

// normalize makes a string all lowercase and removes spaces, dashes, and underscores
func normalize(s string) string {
	return strings.ReplaceAll(
		strings.ReplaceAll(
			strings.ReplaceAll(strings.ToLower(s), " ", ""),
			"-", ""),
		"_", "")
}

// protocolByQuery returns a BIRD BGP protocol string by a given name
func protocolByQuery(query string, protocols map[string]*templating.Protocol) (string, string) {
	if query == "all" {
		return "all", "all"
	}

	// Expand AFI suffix
	if strings.HasSuffix(query, "4") && !strings.HasSuffix(query, "v4") {
		query = strings.TrimSuffix(query, "4") + " v4"
	} else if strings.HasSuffix(query, "6") && !strings.HasSuffix(query, "v6") {
		query = strings.TrimSuffix(query, "6") + " v6"
	}

	// TODO: This doesn't return the same result for an identical query
	query = normalize(query)
	for birdProto, meta := range protocols {
		if fuzzy.Match(query, normalize(birdProto)) || fuzzy.Match(query, normalize(meta.Name)) {
			return birdProto, meta.Name
		}
	}
	return "", ""
}

// confirmYesNo asks a [y/N] question and returns true if the user selects yes
func confirmYesNo(question string) bool {
	log.Printf("%s [y/N] ", question)
	var response string
	_, _ = fmt.Scanln(&response)
	return response == "y" || response == "Y"
}
