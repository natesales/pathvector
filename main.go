package main

import (
	"fmt"
	"os"
	"strings"

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
	if //goland:noinspection GoBoolExpressions
	version == "devel" || strings.Contains(version, "SNAPSHOT") {
		fmt.Fprintln(os.Stderr, `*******************************************************************************
WARNING: This is a development build. It may not be ready for production use.
Please submit any bugs to https://github.com/natesales/pathvector/issues
*******************************************************************************`)
	}

	if err := cmd.Execute(version, commit, date); err != nil {
		log.Fatal(err)
	}
}
