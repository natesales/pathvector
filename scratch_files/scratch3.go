package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"reflect"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

type Config struct {
	ASN       uint             `yaml:"asn" description:"Autonomous System Number" validate:"required"`
	Peers     map[string]*Peer `yaml:"peers" description:"BGP peer configuration"`
	Templates map[string]*Peer `yaml:"templates" description:"BGP peer configuration templates"`
}

type Peer struct {
	Template    *string   `yaml:"template" description:"Configuration template"`
	ASN         *uint     `yaml:"asn" description:"Local ASN" validate:"required"`
	NeighborIPs *[]string `yaml:"neighbors" description:"List of neighbor IPs" validate:"required,ip"`
	LocalPref   *uint     `yaml:"local-pref" description:"BGP local preference" default:"100"`
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

//func setFieldDefault(defaultString string) interface{} {
//	var defaultValue interface{}
//	switch bf.Elem().Kind() {
//	case reflect.String:
//		defaultValue = defaultString
//	case reflect.Uint:
//		defaultValueInt, err := strconv.Atoi(defaultString)
//		if err != nil {
//			log.Fatalf("cant convert '%s' to uint", defaultString)
//		}
//		defaultValue = uint(defaultValueInt)
//	case reflect.Bool:
//		var err error // explicit declaration used to avoid scope issues of defaultValue
//		defaultValue, err = strconv.ParseBool(defaultString)
//		if err != nil {
//			log.Fatalf("can't parse bool %s", defaultString)
//		}
//	case reflect.Struct, reflect.Slice:
//		// Ignore structs and slices
//	default:
//		log.Fatalf("unknown kind %+v", bf.Elem().Kind())
//	}
//	return defaultValue
//}

func main() {
	configFile, err := ioutil.ReadFile("config.yml")
	if err != nil {
		log.Fatal("reading config file: " + err.Error())
	}

	var c Config
	if err := yaml.UnmarshalStrict(configFile, &c); err != nil {
		log.Fatal(err)
	}

	displayPeers(c.Peers)

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
				templateFieldValue := peerValue.FieldByName(fieldName)
				//defaultValueString := at.Field(i).Tag.Get("default")
				//log.Printf("%s def tag %s", name, defaultValueString)
				if templateFieldValue.IsValid() {
					if fieldName != "Template" { // Ignore the template attribute
						tValue := templateValue.Field(i).Interface()
						pVal := reflect.Indirect(templateFieldValue).Interface()

						log.Printf("%s pVal==nil: %+v pValType: %T", fieldName, pVal == nil, pVal)

						//if (templateValue != pVal) && (pVal != defaultValue) && !tmplValue.Field(i).IsNil() {
						if (tValue != nil) && (pVal != nil) && (templateValue.Field(i).Elem().IsValid()) {
							log.Printf("[%s] field %s, setting %+v -> %+v", peerName, fieldName, templateFieldValue.Elem(), templateValue.Field(i).Elem())
							log.Printf("[%s] for field %s, peer has %T, template has %T", peerName, fieldName, templateFieldValue.Elem(), templateValue.Field(i).Elem())
							log.Printf("(templateValue != pVal)=%+v (templateValue != nil)=%+v", tValue != pVal, tValue != nil)
							templateFieldValue.Set(templateValue.Field(i))
						}
					}
				} else {
					log.Fatal("invalid field %s", fieldName)
				}
			}
		}
	}

	displayPeers(c.Peers)
}
