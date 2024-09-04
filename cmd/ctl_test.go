package cmd

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/natesales/pathvector/pkg/templating"
)

func TestCtlProtocolByQuery(t *testing.T) {
	protocolsJSON := `{"HURRICANE_ELECTRIC_AS6939_v4":{"Name":"Hurricane Electric","Tags":null},"HURRICANE_ELECTRIC_AS6939_v6":{"Name":"Hurricane Electric","Tags":null}}`
	var protocols map[string]*templating.Protocol
	assert.Nil(t, json.Unmarshal([]byte(protocolsJSON), &protocols))

	for _, tc := range []struct {
		expected string
		query    string
	}{
		{"HURRICANE_ELECTRIC_AS6939_v4", "Hurricane Electric v4"},
		{"HURRICANE_ELECTRIC_AS6939_v4", "hurricane v4"},
		{"HURRICANE_ELECTRIC_AS6939_v6", "he v6"},
	} {
		t.Run(tc.query, func(t *testing.T) {
			birdProto, _ := protocolByQuery(tc.query, protocols)
			assert.Equal(t, tc.expected, birdProto)
		})
	}
}
