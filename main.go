package main

import (
	log "github.com/sirupsen/logrus"

	"github.com/natesales/pathvector/cmd"
)

//go:generate ./generate.sh

func main() {
	if err := cmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
