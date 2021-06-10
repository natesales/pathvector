package main

import (
	"encoding/json"
	"fmt"
	//"github.com/creasty/defaults"
	"io/ioutil"
	"reflect"
	"strconv"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

type config struct {
	ASN       uint             `yaml:"asn" description:"Autonomous System Number" validate:"required"`
	Peers     map[string]*peer `yaml:"peers" description:"BGP peer configuration"`
	Templates map[string]*peer `yaml:"templates" description:"BGP peer configuration templates"`
}

type peer struct {
	Template    *string   `yaml:"template" description:"Configuration template"`
	ASN         *uint     `yaml:"asn" description:"Local ASN" validate:"required"`
	NeighborIPs *[]string `yaml:"neighbors" description:"List of neighbor IPs" validate:"required,ip"`
	LocalPref   *uint     `yaml:"local-pref" description:"BGP local preference" default:"100"`
	FilterIRR   *bool     `yaml:"filter-irr" description:"Should IRR filtering be applied?" default:"true"`
	FilterRPKI  *bool     `yaml:"filter-rpki" description:"Should RPKI filtering be applied?" default:"true"`
}

type template struct {
	ASN         *uint     `yaml:"asn" description:"Local ASN" validate:"required"`
	NeighborIPs *[]string `yaml:"neighbors" description:"List of neighbor IPs" validate:"required,ip"`
	LocalPref   *uint     `yaml:"local-pref" description:"BGP local preference" default:"100"`
	FilterIRR   *bool     `yaml:"filter-irr" description:"Should IRR filtering be applied?" default:"true"`
	FilterRPKI  *bool     `yaml:"filter-rpki" description:"Should RPKI filtering be applied?" default:"true"`
}

//func (c *peer) UnmarshalYAML(unmarshal func(interface{}) error) error {
//	defaults.Set(c)
//
//	type plain peer
//	if err := unmarshal((*plain)(c)); err != nil {
//		return err
//	}
//
//	return nil
//}

func displayPeers(peers map[string]*peer) {
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

	var c config
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
		if *peerData.Template != "" {
			template := c.Templates[*peerData.Template]
			if template == nil {
				log.Fatalf("template %s not found", *peerData.Template)
			}
			peerValue := reflect.ValueOf(*template)
			bv := reflect.ValueOf(c.Peers[peerName]).Elem()

			at := peerValue.Type()
			for i := 0; i < at.NumField(); i++ {
				name := at.Field(i).Name
				bf := bv.FieldByName(name)
				defaultValueString := at.Field(i).Tag.Get("default")
				//log.Printf("%s def tag %s", name, defaultValueString)
				if bf.IsValid() {
					if name != "Template" { // Ignore the template attribute
						//var defaultValue interface{}
						var defaultValue interface{}
						if defaultValueString != "" {
							switch bf.Elem().Kind() {
							case reflect.String:
								defaultValue = defaultValueString
							case reflect.Uint:
								defaultValueInt, err := strconv.Atoi(defaultValueString)
								if err != nil {
									log.Fatalf("cant convert '%s' to int", defaultValueString)
								}
								defaultValue = uint(defaultValueInt)
							case reflect.Bool:
								var err error // explicit declaration used to avoid scope issues of defaultValue
								defaultValue, err = strconv.ParseBool(defaultValueString)
								if err != nil {
									log.Fatalf("can't parse bool %s", defaultValueString)
								}
							case reflect.Struct, reflect.Slice:
								// Ignore structs and slices
							default:
								log.Fatalf("unknown kind %+v", bf.Elem().Kind())
							}
						} else {
							defaultValue = nil
						}
						defVal := reflect.ValueOf(defaultValue)
						//log.Printf("default value string %s kind %+v val %+v", defaultValueString, bf.Elem().Kind(), defaultValue)

						//log.Printf("[%s] template: %+v peer: %+v default %+v",
						//	peerName,
						//	reflect.Indirect(peerValue.Field(i)),
						//	reflect.Indirect(bf),
						//	defaultValue)
						// if (template.value != default.value) && (peer.value == default.value)
						templateValue := peerValue.Field(i).Interface()
						pVal := reflect.Indirect(bf).Interface()

						//if (templateValue != pVal) && (pVal != defaultValue) && !peerValue.Field(i).IsNil() {
						if (templateValue != pVal) && (templateValue != nil) && (pVal == defVal.Interface()) {
							log.Printf("[%s] for field %s, peer has %+v, template has %+v, and default is %+v", peerName, name, bf.Elem(), peerValue.Field(i).Elem(), defaultValue)
							log.Printf("[%s] for field %s, peer has %T, template has %T and default is %T", peerName, name, bf.Elem(), peerValue.Field(i).Elem(), defaultValue)
							log.Printf("(templateValue != pVal)=%+v (templateValue != nil)=%+v", templateValue != pVal, templateValue != nil)
							bf.Set(peerValue.Field(i))
						}
					}
				} else {
					log.Fatal("invalid field %s", name)
				}
			}
		}
	}

	displayPeers(c.Peers)
}
