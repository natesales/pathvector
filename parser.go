package main

import (
	"encoding/json"
	"flag"
	"github.com/pelletier/go-toml"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"text/template"
	"time"
)

type Peer struct {
	Asn          uint32   `yaml:"asn" toml:"ASN" json:"asn"`
	AsSet        string   `yaml:"as-set" toml:"AS-Set" json:"as-set"`
	MaxPfx4      int64    `yaml:"maxpfx4" yaml:"MaxPfx4" json:"maxpfx4"`
	MaxPfx6      int64    `yaml:"maxpfx6" yaml:"MaxPfx6" json:"maxpfx6"`
	ImportPolicy string   `yaml:"import" toml:"ImportPolicy" json:"import"`
	ExportPolicy string   `yaml:"export" toml:"ExportPolicy" json:"export"`
	NeighborIps  []string `yaml:"neighbors" toml:"Neighbors" json:"neighbors"`
	Multihop     bool     `yaml:"multihop" toml:"Multihop" json:"multihop"`
	Passive      bool     `yaml:"passive" toml:"Passive" json:"passive"`
	Disabled     bool     `yaml:"disabled" toml:"Disabled" json:"disabled"`
	AutoMaxPfx   bool     `yaml:"automaxpfx" toml:"AutoMaxPfx" json:"automaxpfx"`
}

type Config struct {
	Asn      uint32           `yaml:"asn" toml:"ASN" json:"asn"`
	RouterId string           `yaml:"router-id" toml:"Router-ID" json:"router-id"`
	Prefixes []string         `yaml:"prefixes" toml:"Prefixes" json:"prefixes"`
	Peers    map[string]*Peer `yaml:"peers" toml:"Peers" json:"peers"`
}

type PeerTemplate struct {
	Peer Peer
	Name string
}

type PeeringDbResponse struct {
	Data []PeeringDbData `json:"data"`
}

type PeeringDbData struct {
	Name    string `json:"name"`
	AsSet   string `json:"irr_as_set"`
	MaxPfx4 uint32 `json:"info_prefixes4"`
	MaxPfx6 uint32 `json:"info_prefixes6"`
}

var (
	configFilename  = flag.String("config", "config.yml", "Configuration file in YAML, TOML, or JSON format")
	outputDirectory = flag.String("output", "output/", "Directory to write output files to")
)

func getPeeringDbData(asn uint32) PeeringDbData {
	httpClient := http.Client{Timeout: time.Second * 2}
	req, err := http.NewRequest(http.MethodGet, "https://peeringdb.com/api/net?asn="+strconv.Itoa(int(asn)), nil)
	if err != nil {
		log.Fatalf("PeeringDB GET (This peer might not have a PeeringDB page): %v", err)
	}

	res, err := httpClient.Do(req)
	if err != nil {
		log.Fatalf("PeeringDB GET Request: %v", err)
	}

	if res.Body != nil {
		//noinspection GoUnhandledErrorResult
		defer res.Body.Close()
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatalf("PeeringDB Read: %v", err)
	}

	var peeringDbResponse PeeringDbResponse
	if err := json.Unmarshal(body, &peeringDbResponse); err != nil {
		log.Fatalf("PeeringDB JSON Unmarshal: %v", err)
	}

	return peeringDbResponse.Data[0]
}

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

		// Check for no max prefixes
		if !peerData.AutoMaxPfx && (peerData.MaxPfx4 == 0 || peerData.MaxPfx6 == 0) {
			log.Warningf("Peer %s has no max-prefix limits configured. Set automaxpfx to true to pull from PeeringDB.", peerName)
		}

		if peerData.AutoMaxPfx {
			log.Infof("Running PeeringDB query for AS%d", peerData.Asn)
			peeringDb := getPeeringDbData(peerData.Asn)
			peerData.MaxPfx4 = int64(peeringDb.MaxPfx4)
			peerData.MaxPfx6 = int64(peeringDb.MaxPfx6)

			log.Printf("AS%d MaxPfx4: %d", peerData.Asn, peerData.MaxPfx4)
			log.Printf("AS%d MaxPfx6: %d", peerData.Asn, peerData.MaxPfx6)
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

	log.Infof("Modified config: %+v", config)
	log.Info("Generating peer specific files")

	peerTemplate, err := template.ParseFiles("templates/peer_specific.tmpl")
	if err != nil {
		log.Fatalf("Read peer specific template: %v", err)
	}

	// Create peer specific file
	for peerName, peerData := range config.Peers {
		// Create the peer specific file
		peerSpecificFile, err := os.Create(path.Join(*outputDirectory, "AS"+strconv.Itoa(int(peerData.Asn))+".txt"))
		if err != nil {
			log.Fatalf("Create peer specific output file: %v", err)
		}

		err = peerTemplate.Execute(peerSpecificFile, &PeerTemplate{*peerData, peerName})
		if err != nil {
			log.Fatalf("Write peer specific output file: %v", err)
		}

		log.Infof("Wrote peer specific config for AS%d", peerData.Asn)
	}
}
