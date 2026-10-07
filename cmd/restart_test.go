package cmd

import (
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/natesales/pathvector/pkg/config"
	"github.com/natesales/pathvector/pkg/templating"
)

func TestResolveProtocols(t *testing.T) {
	protocols := map[string]*templating.Protocol{
		"EXAMPLE_AS65510_v4": {Name: "Example"},
		"EXAMPLE_AS65510_v6": {Name: "Example"},
		"OTHER_AS65520_v6":   {Name: "Other"},
	}

	assert.Equal(t, []string{"EXAMPLE_AS65510_v4", "EXAMPLE_AS65510_v6"}, resolveProtocols([]string{"Example"}, protocols))
	assert.Equal(t, []string{"EXAMPLE_AS65510_v4"}, resolveProtocols([]string{"EXAMPLE_AS65510_v4"}, protocols))
	assert.Equal(t, []string{"device1"}, resolveProtocols([]string{"device1"}, protocols))
	assert.Equal(t,
		[]string{"EXAMPLE_AS65510_v4", "EXAMPLE_AS65510_v6", "OTHER_AS65520_v6"},
		resolveProtocols([]string{"Example", "EXAMPLE_AS65510_v6", "Other"}, protocols),
	)
	assert.Equal(t, []string{"Example"}, resolveProtocols([]string{"Example"}, nil))
}

func TestReadProtocolNames(t *testing.T) {
	dir := t.TempDir()
	//nolint:golint,gosec
	assert.Nil(t, os.WriteFile(path.Join(dir, "protocols.json"), []byte(`{"EXAMPLE_AS65510_v4":{"Name":"Example","Tags":["a"]}}`), 0644))

	protocols, err := readProtocolNames(dir)
	assert.Nil(t, err)
	assert.Equal(t, "Example", protocols["EXAMPLE_AS65510_v4"].Name)
	assert.Equal(t, []string{"a"}, protocols["EXAMPLE_AS65510_v4"].Tags)

	_, err = readProtocolNames(path.Join(dir, "missing"))
	assert.NotNil(t, err)
}

func TestBIRDPaths(t *testing.T) {
	socket, dir := birdPaths(nil)
	assert.Equal(t, "/run/bird/bird.ctl", socket)
	assert.Equal(t, "/etc/bird/", dir)

	socket, dir = birdPaths(&config.Config{BIRDSocket: "/tmp/bird.ctl", BIRDDirectory: "/tmp/bird/"})
	assert.Equal(t, "/tmp/bird.ctl", socket)
	assert.Equal(t, "/tmp/bird/", dir)
}
