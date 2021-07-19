package templating

import (
	"testing"

	"github.com/natesales/pathvector/internal/config"
	"github.com/natesales/pathvector/internal/embed"
)

func TestLoadTemplates(t *testing.T) {
	if err := Load(embed.FS); err != nil {
		t.Error(err)
	}
}

func TestWriteUIFile(t *testing.T) {
	WriteUIFile(&config.Config{WebUIFile: "/tmp/pathvector-go-test-ui.html"})
}

func TestWriteBlankVRRPConfig(t *testing.T) {
	WriteVRRPConfig([]config.VRRPInstance{}, "/tmp/pathvector-go-test-keepalived.conf")
}

func TestWriteVRRPConfig(t *testing.T) {
	WriteVRRPConfig([]config.VRRPInstance{{State: "primary"}}, "/tmp/pathvector-go-test-keepalived.conf")
}
