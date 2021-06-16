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
	cliFlags.WebUIFile = "/tmp/pathvector-go-test-ui.html"
	writeUIFile(&config{})
}

func TestWriteBlankVRRPConfig(t *testing.T) {
	cliFlags.KeepalivedConfig = "/tmp/pathvector-go-test-keepalived.conf"
	writeVRRPConfig(&config{})
}

func TestWriteVRRPConfig(t *testing.T) {
	writeVRRPConfig(&config{
		VRRPInstances: []vrrpInstance{{
			State: "primary",
		}},
	})
}
