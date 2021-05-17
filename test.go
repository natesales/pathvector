package main

import (
	"fmt"
	"reflect"

	log "github.com/sirupsen/logrus"
)

func main() {
	myPeer := peer{Asn: 34553}

	t := reflect.TypeOf(myPeer)
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		description := field.Tag.Get("description")

		if description == "" {
			log.Fatalf("code error: %s doesn't have a description", field.Name)
		}

		fmt.Printf("%v (%v) description: %s\n", field.Name, field.Type.Name(), description)
	}
}
