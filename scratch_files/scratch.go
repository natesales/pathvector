package main

import (
	"fmt"
	"io/ioutil"
	"reflect"
	"strconv"

	"github.com/creasty/defaults"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

type config struct {
	ASN       uint             `yaml:"asn" description:"Autonomous System Number" validate:"required"`
	Peers     map[string]*peer `yaml:"peers" description:"BGP peer configuration"`
	Templates map[string]peer  `yaml:"templates" description:"BGP peer templates"`
}

type peer struct {
	Template    string   `yaml:"template" description:"Configuration template"`
	ASN         uint     `yaml:"asn" description:"Local ASN" validate:"required"`
	NeighborIPs []string `yaml:"neighbors" description:"List of neighbor IPs" validate:"required,ip"`
	LocalPref   uint     `yaml:"local-pref" description:"BGP local preference" default:"100"`
	FilterIRR   bool     `yaml:"filter-irr" description:"Should IRR filtering be applied?" default:"true"`
}

func (c *peer) UnmarshalYAML(unmarshal func(interface{}) error) error {
	defaults.Set(c)

	type plain peer
	if err := unmarshal((*plain)(c)); err != nil {
		return err
	}

	return nil
}

func (c *config) UnmarshalYAML(unmarshal func(interface{}) error) error {
	defaults.Set(c)

	type plain config
	if err := unmarshal((*plain)(c)); err != nil {
		return err
	}

	return nil
}

func main() {
	configFile, err := ioutil.ReadFile("config.yml")
	if err != nil {
		log.Fatal("reading config file: " + err.Error())
	}

	var c config
	if err := yaml.UnmarshalStrict(configFile, &c); err != nil {
		log.Fatal(err)
	}

	for name, peer := range c.Peers {
		fmt.Printf("%s %+v\n", name, peer)
	}

	for peerName, peerData := range c.Peers { // For each peer
		if peerData.Template != "" {
			template := c.Templates[peerData.Template]

			av := reflect.ValueOf(template)
			bv := reflect.ValueOf(c.Peers[peerName]).Elem()

			at := av.Type()
			for i := 0; i < at.NumField(); i++ {
				name := at.Field(i).Name
				bf := bv.FieldByName(name)
				defaultValueString := at.Field(i).Tag.Get("default")
				if defaultValueString != "" {
					var defaultValue interface{}
					switch bf.Kind() {
					case reflect.String:
						defaultValue = defaultValueString
					case reflect.Uint:
						defaultValueInt, err := strconv.Atoi(defaultValueString)
						if err != nil {
							panic(err) // TODO
						}
						defaultValue = uint(defaultValueInt)
					case reflect.Bool:
						defaultValue, err = strconv.ParseBool(defaultValueString)
						if err != nil {
							panic(err) // TODO
						}
					case reflect.Struct, reflect.Slice:
						// Ignore structs and slices
					default:
						log.Fatalf("unknown kind %+v", at.Kind())
					}
					log.Printf("default value string %s kind %+v val %+v", defaultValueString, bf.Kind(), defaultValue)

					if defaultValue != nil {
						if bf.IsValid() {
							log.Printf("[%s] template: %+v peer: %+v", peerName, av.Field(i), bf)
							// if (template.value != default.value) && (peer.value == default.value)
							templateValue := av.Field(i).Interface()
							peerValue := bf.Interface()

							//           template_value != peer_value
							//if (av.Field(i).Interface() != bf.Interface()) && () {
							log.Printf("templateValue %+v defaultValue %+v, (templateValue != defaultValue): %+v and (peerValue == defaultValue): %+v",
								templateValue,
								defaultValue,
								templateValue != defaultValue,
								peerValue == defaultValue,
							)
							if (templateValue != defaultValue) && (peerValue == defaultValue) {
								log.Printf("[%s] setting %s, %+v -> %+v", peerName, name, bf, av.Field(i))
								bf.Set(av.Field(i))
							}
						} else {
							log.Fatal("invalid field %s", name)
						}
					}
				}
			}
		}
	}

	for name, peer := range c.Peers {
		fmt.Printf("%s %+v\n", name, peer)
	}
}
