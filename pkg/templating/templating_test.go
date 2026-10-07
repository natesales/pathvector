package templating

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/natesales/pathvector/pkg/config"
	"github.com/natesales/pathvector/pkg/embed"
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
	WriteVRRPConfig(map[string]*config.VRRPInstance{}, "/tmp/pathvector-go-test-keepalived.conf")
}

func TestWriteVRRPConfig(t *testing.T) {
	WriteVRRPConfig(map[string]*config.VRRPInstance{"VRRP 1": {State: "primary"}}, "/tmp/pathvector-go-test-keepalived.conf")
}

func TestWriteVRRPConfigVIPInterface(t *testing.T) {
	if err := Load(embed.FS); err != nil {
		t.Fatal(err)
	}

	instances := map[string]*config.VRRPInstance{
		"1": {
			State:        "MASTER",
			Interface:    "eth0",
			VRID:         1,
			Priority:     255,
			VIPInterface: "lo",
			VIPs4:        []string{"192.0.2.1/32"},
			VIPs6:        []string{"2001:db8::1/128"},
		},
		"2": {
			State:     "BACKUP",
			Interface: "eth1",
			VRID:      2,
			Priority:  100,
			VIPs4:     []string{"198.51.100.1/32"},
		},
	}
	file := filepath.Join(t.TempDir(), "keepalived.conf")
	WriteVRRPConfig(instances, file)

	out, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	conf := string(out)

	for _, expected := range []string{
		"interface eth0",
		"192.0.2.1/32 dev lo",
		"2001:db8::1/128 dev lo",
		"interface eth1",
	} {
		if !strings.Contains(conf, expected) {
			t.Errorf("expected %q in keepalived config:\n%s", expected, conf)
		}
	}

	// Instances without vip-interface keep the plain VIP syntax
	if strings.Contains(conf, "198.51.100.1/32 dev") {
		t.Errorf("unexpected dev on VIP without vip-interface:\n%s", conf)
	}
}
