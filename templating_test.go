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
