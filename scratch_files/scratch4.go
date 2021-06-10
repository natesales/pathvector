package main

import (
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"reflect"
	"strconv"
)

type Config struct {
	ASN       uint             `yaml:"asn" description:"Autonomous System Number" validate:"required"`
	Peers     map[string]*Peer `yaml:"peers" description:"BGP peer configuration"`
	Templates map[string]*Peer `yaml:"templates" description:"BGP peer configuration templates"`
}

type Peer struct {
	Template    *string   `yaml:"template" description:"Configuration template" default:"-"`
	ASN         *int      `yaml:"asn" description:"Local ASN" validate:"required" default:"-"`
	NeighborIPs *[]string `yaml:"neighbors" description:"List of neighbor IPs" validate:"required,ip" default:"-"`
	LocalPref   *int      `yaml:"local-pref" description:"BGP local preference" default:"100"`
	FilterIRR   *bool     `yaml:"filter-irr" description:"Should IRR filtering be applied?" default:"true"`
	FilterRPKI  *bool     `yaml:"filter-rpki" description:"Should RPKI filtering be applied?" default:"true"`
}

func displayPeers(peers map[string]*Peer) {
	for name, peer := range peers {
		if out, err := json.Marshal(peer); err != nil {
			log.Fatal(err)
		} else {
			fmt.Printf("%s %+v\n", name, string(out))
		}
	}
}

func main() {
	configFile, err := ioutil.ReadFile("config.yml")
	if err != nil {
		log.Fatal("reading config file: " + err.Error())
	}

	var c Config
	if err := yaml.UnmarshalStrict(configFile, &c); err != nil {
		log.Fatal(err)
	}

	// Check for invalid templates
	for templateName, templateData := range c.Templates {
		if templateData.Template != nil && *templateData.Template != "" {
			log.Fatalf("templates must not have a template attribute set, but %s does", templateName)
		}
	}

	// Assign values from template
	for peerName, peerData := range c.Peers { // For each peer
		if peerData.Template != nil && *peerData.Template != "" {
			template := c.Templates[*peerData.Template]
			if template == nil {
				log.Fatalf("template %s not found", *peerData.Template)
			}
			templateValue := reflect.ValueOf(*template)
			peerValue := reflect.ValueOf(c.Peers[peerName]).Elem()

			templateValueType := templateValue.Type()
			for i := 0; i < templateValueType.NumField(); i++ {
				fieldName := templateValueType.Field(i).Name
				peerFieldValue := peerValue.FieldByName(fieldName)
				if fieldName != "Template" { // Ignore the template attribute
					pVal := reflect.Indirect(peerFieldValue)
					peerHasValueConfigured := pVal.IsValid()
					tValue := templateValue.Field(i)
					templateHasValueConfigured := !tValue.IsNil()

					if peerHasValueConfigured {
						// Dont do anything
					} else if templateHasValueConfigured && !peerHasValueConfigured {
						// Use the templates value
						peerFieldValue.Set(templateValue.Field(i))
					}

					//log.Printf("[%s] %s val: %+v type: %T hasConfigured: %v", peerName, fieldName, tValue, tValue, templateHasValueConfigured)
				}
			}
		} // end peer template processor

		// Set default values
		peerValue := reflect.ValueOf(c.Peers[peerName]).Elem()
		templateValueType := peerValue.Type()
		for i := 0; i < templateValueType.NumField(); i++ {
			fieldName := templateValueType.Field(i).Name
			fieldValue := peerValue.FieldByName(fieldName)
			defaultString := templateValueType.Field(i).Tag.Get("default")
			if defaultString == "" {
				log.Fatalf("code error: field %s has no default value", fieldName)
			}
			//log.Printf("peer %s field %s value %+v", peerName, fieldName, fieldValue)
			if fieldValue.IsNil() {
				elemToSwitch := templateValueType.Field(i).Type.Elem().Kind()
				switch elemToSwitch {
				case reflect.String:
					fieldValue.Set(reflect.ValueOf(&defaultString))
				case reflect.Int:
					defaultValueInt, err := strconv.Atoi(defaultString)
					if err != nil {
						log.Fatalf("cant convert '%s' to uint", defaultString)
					}
					fieldValue.Set(reflect.ValueOf(&defaultValueInt))
				case reflect.Bool:
					var err error // explicit declaration used to avoid scope issues of defaultValue
					defaultBool, err := strconv.ParseBool(defaultString)
					if err != nil {
						log.Fatalf("can't parse bool %s", defaultString)
					}
					fieldValue.Set(reflect.ValueOf(&defaultBool))
				case reflect.Struct, reflect.Slice:
					// Ignore structs and slices
				default:
					log.Fatalf("unknown kind %+v for field %s", elemToSwitch, fieldName)
				}
			}
		}
	}

	displayPeers(c.Peers)
}
