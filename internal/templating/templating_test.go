package templating

import (
	"testing"

	"github.com/natesales/pathvector/internal/config"
)

func TestLoadTemplates(t *testing.T) {
	if err := Load(); err != nil {
		t.Error(err)
	}
}

func TestWriteUIFile(t *testing.T) {
	WriteUIFile(&config.Global{WebUIFile: "/tmp/pathvector-test/ui.html"})
}

func TestWriteBlankVRRPConfig(t *testing.T) {
	WriteVRRPConfig(&config.Global{KeepalivedConfig: "/tmp/pathvector-test/keepalived.conf"})
}

func TestWriteVRRPConfig(t *testing.T) {
	WriteVRRPConfig(&config.Global{
		KeepalivedConfig: "/tmp/pathvector-test/keepalived.conf",
		VRRPInstances: []config.VRRPInstance{{
			State: "primary",
		}},
	})
}
