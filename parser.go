package main

import (
	"encoding/json"
	"flag"
	"github.com/pelletier/go-toml"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"text/template"
	"time"
)

type Peer struct {
	Asn            uint32   `yaml:"asn" toml:"ASN" json:"asn"`
	AsSet          string   `yaml:"as-set" toml:"AS-Set" json:"as-set"`
	MaxPfx4        int64    `yaml:"maxpfx4" yaml:"MaxPfx4" json:"maxpfx4"`
	MaxPfx6        int64    `yaml:"maxpfx6" yaml:"MaxPfx6" json:"maxpfx6"`
	PfxLimitAction string   `yaml:"pfxlimitaction" yaml:"PfxLimitAction" json:"pfxlimitaction"`
	PfxFilter4     []string `yaml:"pfxfilter4" yaml:"PfxFilter4" json:"PfxFilter4"`
	PfxFilter6     []string `yaml:"pfxfilter6" yaml:"PfxFilter6" json:"PfxFilter6"`
	ImportPolicy   string   `yaml:"import" toml:"ImportPolicy" json:"import"`
	ExportPolicy   string   `yaml:"export" toml:"ExportPolicy" json:"export"`
	NeighborIps    []string `yaml:"neighbors" toml:"Neighbors" json:"neighbors"`
	Multihop       bool     `yaml:"multihop" toml:"Multihop" json:"multihop"`
	Passive        bool     `yaml:"passive" toml:"Passive" json:"passive"`
	Disabled       bool     `yaml:"disabled" toml:"Disabled" json:"disabled"`
	AutoMaxPfx     bool     `yaml:"automaxpfx" toml:"AutoMaxPfx" json:"automaxpfx"`
	AutoPfxFilter  bool     `yaml:"autopfxfilter" toml:"AutoPfxFilter" json:"autopfxfilter"`
	QueryTime      string   `yaml:"-" toml:"-" json:"-"`
}

type Config struct {
	Asn      uint32           `yaml:"asn" toml:"ASN" json:"asn"`
	RouterId string           `yaml:"router-id" toml:"Router-ID" json:"router-id"`
	Prefixes []string         `yaml:"prefixes" toml:"Prefixes" json:"prefixes"`
	Peers    map[string]*Peer `yaml:"peers" toml:"Peers" json:"peers"`
}

type PeerTemplate struct {
	Peer             Peer
	Name             string
	PfxFilterString4 string // Contains string representation of IPv4 prefix filter
	PfxFilterString6 string // Contains string representation of IPv6 prefix filter
	Global           Config
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
	birdSocket      = flag.String("socket", "/run/bird/bird.ctl", "BIRD control socket")
)

func getPeeringDbData(asn uint32) PeeringDbData {
	httpClient := http.Client{Timeout: time.Second * 5}
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

func getPrefixFilter(macro string, family uint8) []string {
	// Run bgpq4 for BIRD format with aggregation enabled
	cmd := exec.Command("bgpq4", "-Ab"+strconv.Itoa(int(family)), macro)
	stdout, err := cmd.Output()
	if err != nil {
		log.Fatalln(err.Error())
	}

	// Remove whitespace and commas from output
	output := strings.ReplaceAll(string(stdout), ",\n    ", "\n")

	// Remove array prefix
	output = strings.ReplaceAll(output, "NN = [\n    ", "")

	// Remove array suffix
	output = strings.ReplaceAll(output, "];", "")

	// Remove whitespace (in this case there should only be trailing whitespace)
	output = strings.TrimSpace(output)

	// Split output by newline
	return strings.Split(output, "\n")
}

func readNoBuffer(reader io.Reader) string {
	buf := make([]byte, 1024)
	n, err := reader.Read(buf[:])

	if err != nil {
		log.Fatalf("BIRD read error: ", err)
	}

	return string(buf[:n])
}

func runBirdCommand(command string) {
	log.Println("Connecting to BIRD socket")
	conn, err := net.Dial("unix", *birdSocket)
	if err != nil {
		log.Fatalf("BIRD socket connect: %v", err)
	}
	//noinspection GoUnhandledErrorResult
	defer conn.Close()

	log.Println("Connected to BIRD socket")
	log.Printf("BIRD init response: %s", readNoBuffer(conn))

	log.Printf("Sending BIRD command: %s", command)
	_, err = conn.Write([]byte(strings.Trim(command, "\n") + "\n"))
	log.Printf("Sent BIRD command: %s", command)
	if err != nil {
		log.Fatalf("BIRD write error:", err)
	}

	log.Printf("BIRD response: %s", readNoBuffer(conn))
}

func main() {
	log.Info("Starting BCG")
	log.Info("Generating peer specific files")

	peerTemplate, err := template.New("").Funcs(template.FuncMap{
		"Contains": func(s, substr string) bool { return strings.Contains(s, substr) },
	}).ParseFiles("peer.tmpl")
	if err != nil {
		log.Fatalf("Read peer specific template: %v", err)
	}

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
		// Set the query time to a default string
		peerData.QueryTime = "[No time-specific operations performed]"

		// Set default pfxlimitaction
		if peerData.PfxLimitAction == "" {
			peerData.PfxLimitAction = "disable"
		} else if !(peerData.PfxLimitAction == "disable" || peerData.PfxLimitAction == "restart" || peerData.PfxLimitAction == "block" || peerData.PfxLimitAction == "warn") {
			log.Fatalf("Peer %s has an invalid pfxlimitaction. Acceptable values are warn, block, restart, and disable", peerName)
		}

		// If no AS-Set is defined and the import policy requires it
		if !peerData.AutoPfxFilter && peerData.ImportPolicy == "cone" {
			// Check for empty AS-Set
			if peerData.AsSet == "" {
				log.Fatalf("Peer %s has a cone filtered import policy and has no AS-Set defined. Set autopfxfilter to true to enable automatic IRRDB imports", peerName)
			} else if !strings.HasPrefix(peerData.AsSet, "AS") { // If AS-Set doesn't start with "AS" TODO: Better validation here. What is a valid AS-Set?
				log.Warnf("AS-Set for %s (as-set: %s) doesn't start with 'AS' and might be invalid", peerName, peerData.AsSet)
			}

			// Check for empty prefix filters
			if peerData.ImportPolicy != "none" && (len(peerData.PfxFilter4) == 0 || len(peerData.PfxFilter6) == 0) {
				log.Fatalf("Peer %s has a cone filtered import policy and has no prefix filters defined. Set autopfxfilter to true to enable automatic IRRDB imports", peerName)
			}
		}

		// Check for no max prefixes
		if !peerData.AutoMaxPfx && (peerData.MaxPfx4 == 0 || peerData.MaxPfx6 == 0) {
			log.Warningf("Peer %s has no max-prefix limits configured. Set automaxpfx to true to pull from PeeringDB.", peerName)
		}

		var peeringDbData PeeringDbData

		// If MaxPfx limits should be pulled from PeeringDB
		if peerData.AutoMaxPfx {
			if peeringDbData == (PeeringDbData{}) {
				log.Infof("Running PeeringDB query for AS%d", peerData.Asn)
				peeringDbData = getPeeringDbData(peerData.Asn)
			}

			peerData.MaxPfx4 = int64(peeringDbData.MaxPfx4)
			peerData.MaxPfx6 = int64(peeringDbData.MaxPfx6)

			log.Printf("AutoMaxPfx AS%d MaxPfx4: %d", peerData.Asn, peerData.MaxPfx4)
			log.Printf("AutoMaxPfx AS%d MaxPfx6: %d", peerData.Asn, peerData.MaxPfx6)

			// Update the "latest operation" timestamp
			peerData.QueryTime = time.Now().Format(time.RFC1123)
		}

		// If PfxFilter sets should be pulled from PeeringDB
		if peerData.AutoPfxFilter {
			if peeringDbData == (PeeringDbData{}) {
				log.Infof("Running PeeringDB query for AS%d", peerData.Asn)
				peeringDbData = getPeeringDbData(peerData.Asn)
			}

			log.Infof("Running IRRDB query for AS%d", peerData.Asn)
			peerData.PfxFilter4 = getPrefixFilter(peeringDbData.AsSet, 4)
			peerData.PfxFilter6 = getPrefixFilter(peeringDbData.AsSet, 6)

			log.Printf("AutoPfxFilter AS%d Aggregated Entries: %d", peerData.Asn, len(peerData.PfxFilter4))
			log.Printf("AutoPfxFilter AS%d Aggregated Entries: %d", peerData.Asn, len(peerData.PfxFilter6))

			// Update the "latest operation" timestamp
			peerData.QueryTime = time.Now().Format(time.RFC1123)
		}

		// Validate import policy
		if !(peerData.ImportPolicy == "any" || peerData.ImportPolicy == "cone" || peerData.ImportPolicy == "none") {
			log.Fatalf("Peer %s has an invalid import policy. Acceptable values are 'any', 'cone', or 'none'", peerName)
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

	// Create peer specific file
	for peerName, peerData := range config.Peers {
		// Create the peer specific file
		peerSpecificFile, err := os.Create(path.Join(*outputDirectory, "AS"+strconv.Itoa(int(peerData.Asn))+".conf"))
		if err != nil {
			log.Fatalf("Create peer specific output file: %v", err)
		}

		var pfxFilterString4, pfxFilterString6 = "", ""

		if peerData.ImportPolicy == "cone" {
			// Build prefix filter sets in BIRD format
			for i, prefix := range peerData.PfxFilter4 {
				pfxFilterString4 += "    " + prefix
				if i != len(peerData.PfxFilter4)-1 {
					pfxFilterString4 += ",\n"
				}
			}

			for i, prefix := range peerData.PfxFilter6 {
				pfxFilterString6 += "    " + prefix
				if i != len(peerData.PfxFilter6)-1 {
					pfxFilterString6 += ",\n"
				}
			}
		}

		err = peerTemplate.ExecuteTemplate(peerSpecificFile, "peer.tmpl", &PeerTemplate{*peerData, peerName, pfxFilterString4, pfxFilterString6, config})
		if err != nil {
			log.Fatalf("Execute template: %v", err)
		}

		log.Infof("Wrote peer specific config for AS%d", peerData.Asn)
	}

	runBirdCommand("configure")
}
