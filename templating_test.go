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
	webUIFile = "/tmp/pathvector-go-test-ui.html"
	writeUIFile(&Config{})
}

func TestWriteBlankVRRPConfig(t *testing.T) {
	keepalivedConfig = "/tmp/pathvector-go-test-keepalived.conf"
	writeVRRPConfig(&Config{})
}

func TestWriteVRRPConfig(t *testing.T) {
	writeVRRPConfig(&Config{
		VRRPInstances: []VRRPInstance{{
			State: "primary",
		}},
	})
}
