package main

import (
	"encoding/json"
	"flag"
	"github.com/pelletier/go-toml"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"strings"
)

// debian package this, progress bar (maybe?), peeribgdb+irr cache?, multiple BGPD temolates

// Add validate to this
type Neighbor struct {
	Asn uint32 `yaml:"asn" toml:"ASN" json:"asn"`
	//AsSet        string   `yaml:"asn" toml:"Asn" json:"asn"`
	ImportPolicy string   `yaml:"import" toml:"ImportPolicy" json:"import"`
	ExportPolicy string   `yaml:"export" toml:"ExportPolicy" json:"export"`
	NeighborIps  []string `yaml:"neighbors" toml:"Neighbors" json:"neighbors"` // or ip address type
}

type Config struct {
	Asn      uint32              `yaml:"asn" toml:"ASN" json:"asn"`
	RouterId string              `yaml:"router-id" toml:"Router-ID" json:"router-id"`
	Prefixes []string            `yaml:"prefixes" toml:"Prefixes" json:"prefixes"` // or an "ipnetwork" type?
	Peers    map[string]Neighbor `yaml:"peers" toml:"Peers" json:"peers"`
}

var (
	configFilename = flag.String("config", "config.yml", "Configuration file in YAML, TOML, or JSON format")
)

func main() {
	flag.Parse()

	configFile, err := ioutil.ReadFile(*configFilename)
	if err != nil {
		log.Fatalf("Reading %s: %v", *configFilename, err)
	}

	var config Config

	_splitFilename := strings.Split(*configFilename, ".")
	switch extension := _splitFilename[len(_splitFilename)-1]; extension {
	case "yaml", "yml":
		log.Info("Using YAML configuration format")
		err = yaml.Unmarshal(configFile, &config)
		if err != nil {
			log.Fatalf("YAML Unmarshal: %v", err)
		}
	case "toml":
		log.Info("Using TOML configuration format")
		err = toml.Unmarshal(configFile, &config)
		if err != nil {
			log.Fatalf("TOML Unmarshal: %v", err)
		}
	case "json":
		log.Info("Using JSON configuration format")
		err = json.Unmarshal(configFile, &config)
		if err != nil {
			log.Fatalf("JSON Unmarshal: %v", err)
		}
	default:
		log.Fatalf("Files with extension '%s' are not supported. (Acceptable values are yaml, toml, json", extension)
	}

	log.Println(config)
}
