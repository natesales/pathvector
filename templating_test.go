package main

import (
	"embed"
	"testing"
)

//go:embed templates/*
var testingEmbedFS embed.FS

func TestLoadTemplates(t *testing.T) {
	if err := loadTemplates(testingEmbedFS); err != nil {
		t.Error(err)
	}
}

func TestWriteUIFile(t *testing.T) {
	writeUiFile(&config{WebUIFile: "/tmp/wireframe-ui.html"})
}

func TestWriteBlankVRRPConfig(t *testing.T) {
	writeVRRPConfig(&config{KeepalivedConfig: "/tmp/wireframe-keepalived.conf"})
}

func TestWriteVRRPConfig(t *testing.T) {
	writeVRRPConfig(&config{
		KeepalivedConfig: "/tmp/wireframe-keepalived.conf",
		VRRPInstances: []vrrpInstance{{
			State: "primary",
		}},
	})
}
