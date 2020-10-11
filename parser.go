package main

import (
	"encoding/json"
	"flag"
	"github.com/pelletier/go-toml"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"net"
	"strings"
)

type Neighbor struct {
	Asn          uint32   `yaml:"asn" toml:"ASN" json:"asn"`
	AsSet        string   `yaml:"as-set" toml:"AS-Set" json:"as-set"`
	ImportPolicy string   `yaml:"import" toml:"ImportPolicy" json:"import"`
	ExportPolicy string   `yaml:"export" toml:"ExportPolicy" json:"export"`
	NeighborIps  []string `yaml:"neighbors" toml:"Neighbors" json:"neighbors"`
	Multihop     bool     `yaml:"multihop" toml:"Multihop" json:"multihop"`
	Passive      bool     `yaml:"passive" toml:"Passive" json:"passive"`
	Disabled     bool     `yaml:"disabled" toml:"Disabled" json:"disabled"`
}

type Config struct {
	Asn      uint32              `yaml:"asn" toml:"ASN" json:"asn"`
	RouterId string              `yaml:"router-id" toml:"Router-ID" json:"router-id"`
	Prefixes []string            `yaml:"prefixes" toml:"Prefixes" json:"prefixes"`
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

	log.Infof("Loaded config: %+v", config)

	// Validate Router ID in dotted quad format
	if net.ParseIP(config.RouterId).To4() == nil {
		log.Fatalf("Router ID %s is not in valid dotted quad notation", config.RouterId)
	}

	// Validate CIDR notation of originated prefixes
	for _, addr := range config.Prefixes {
		if _, _, err := net.ParseCIDR(addr); err != nil {
			log.Fatalf("%s is not a valid IPv4 or IPv6 prefix in CIDR notation", addr)
		}
	}

	// Validate peers
	for peerName, peerData := range config.Peers {
		// If no AS-Set is defined and the import policy requires it
		if peerData.ImportPolicy == "macro" {
			if peerData.AsSet != "" {
				log.Fatalf("Peer %s has a filtered import policy and has no AS-Set defined", peerName)
			} else if !strings.HasPrefix(peerData.AsSet, "AS") { // If AS-Set doesn't start with "AS" TODO: Better validation here. What is a valid AS-Set?
				log.Warnf("AS-Set for %s (as-set: %s) doesn't start with 'AS' and might be invalid", peerName, peerData.AsSet)
			}
		}

		// Validate import policy
		if !(peerData.ImportPolicy == "any" || peerData.ImportPolicy == "macro" || peerData.ImportPolicy == "none") {
			log.Fatalf("Peer %s has an invalid import policy. Acceptable values are 'any', 'macro', or 'none'", peerName)
		}

		// Validate export policy
		if !(peerData.ExportPolicy == "any" || peerData.ExportPolicy == "cone" || peerData.ExportPolicy == "none") {
			log.Fatalf("Peer %s has an invalid export policy. Acceptable values are 'any', 'cone', or 'none'", peerName)
		}

		// Validate neighbor IPs
		for _, addr := range peerData.NeighborIps {
			if net.ParseIP(addr) == nil {
				log.Fatalf("Neighbor address of peer %s (addr: %s) is not a valid IPv4 or IPv6 address", peerName, addr)
			}
		}
	}
}
