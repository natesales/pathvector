package main

import (
	log "github.com/sirupsen/logrus"

	"github.com/natesales/pathvector/cmd"
)

// Build process flags
var (
	version = "devel"
	commit  = "unknown"
	date    = "unknown"
)

//go:generate ./docs/generate.sh

func main() {
	if err := cmd.Execute(version, commit, date); err != nil {
		log.Fatal(err)
	}
}
