package main

import (
	log "github.com/sirupsen/logrus"
	"os"
)

// writeUiFile renders and writes the web UI file
func writeUiFile(config *Config) {
	// Create the ui output file
	log.Debug("Creating global config")
	uiFileObj, err := os.Create(opts.UiFile)
	if err != nil {
		log.Fatalf("Create UI output file: %v", err)
	}
	log.Debug("Finished creating UI file")

	// Render the UI template and write to disk
	log.Debug("Writing ui file")
	err = uiTemplate.ExecuteTemplate(uiFileObj, "ui.tmpl", config)
	if err != nil {
		log.Fatalf("Execute ui template: %v", err)
	}
	log.Debug("Finished writing ui file")
}
