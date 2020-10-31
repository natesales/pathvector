package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/kennygrant/sanitize"
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

var release = "devel" // This is set by go build

// Peer contains all information specific to a single peer network
type Peer struct {
	Asn         uint32   `yaml:"asn" toml:"ASN" json:"asn"`
	Type        string   `yaml:"type" toml:"Type" json:"type"`
	Prepends    uint32   `yaml:"prepends" toml:"Prepends" json:"prepends"`
	LocalPref   uint32   `yaml:"local-pref" toml:"LocalPref" json:"local-pref"`
	Multihop    bool     `yaml:"multihop" toml:"Multihop" json:"multihop"`
	Passive     bool     `yaml:"passive" toml:"Passive" json:"passive"`
	Disabled    bool     `yaml:"disabled" toml:"Disabled" json:"disabled"`
	PreImport   string   `yaml:"pre-import" toml:"PreImport" json:"pre-import"`
	PreExport   string   `yaml:"pre-export" toml:"PreExport" json:"pre-export"`
	NeighborIps []string `yaml:"neighbors" toml:"Neighbors" json:"neighbors"`

	AsSet      string   `yaml:"-" toml:"-" json:"-"`
	QueryTime  string   `yaml:"-" toml:"-" json:"-"`
	MaxPrefix4 uint     `yaml:"-" toml:"-" json:"-"`
	MaxPrefix6 uint     `yaml:"-" toml:"-" json:"-"`
	Name       string   `yaml:"-" toml:"-" json:"-"`
	PrefixSet4 []string `yaml:"-" toml:"-" json:"-"`
	PrefixSet6 []string `yaml:"-" toml:"-" json:"-"`
}

// Config contains global configuration about this router and BCG instance
type Config struct {
	Asn       uint32           `yaml:"asn" toml:"ASN" json:"asn"`
	RouterId  string           `yaml:"router-id" toml:"Router-ID" json:"router-id"`
	Prefixes  []string         `yaml:"prefixes" toml:"Prefixes" json:"prefixes"`
	Peers     map[string]*Peer `yaml:"peers" toml:"Peers" json:"peers"`
	IrrDb     string           `yaml:"irrdb" toml:"IRRDB" json:"irrdb"`
	RtrServer string           `yaml:"rtr-server" toml:"RTR-Server" json:"rtr-server"`

	OriginSet4 []string `yaml:"-" toml:"-" json:"-"`
	OriginSet6 []string `yaml:"-" toml:"-" json:"-"`
}

// PeeringDbResponse contains the response from a PeeringDB query
type PeeringDbResponse struct {
	Data []PeeringDbData `json:"data"`
}

// PeeringDbData contains the actual data from PeeringDB response
type PeeringDbData struct {
	Name    string `json:"name"`
	AsSet   string `json:"irr_as_set"`
	MaxPfx4 uint   `json:"info_prefixes4"`
	MaxPfx6 uint   `json:"info_prefixes6"`
}

// Config struct passed to peer template
type PeerTemplate struct {
	Peer   Peer
	Config Config
}

// Flags
var (
	configFilename     = flag.String("config", "/etc/bcg/config.yml", "Configuration file in YAML, TOML, or JSON format")
	outputDirectory    = flag.String("output", "/etc/bird/", "Directory to write output files to")
	templatesDirectory = flag.String("templates", "/etc/bcg/templates/", "Templates directory")
	birdSocket         = flag.String("socket", "/run/bird/bird.ctl", "BIRD control socket")
	dryRun             = flag.Bool("dryrun", false, "Skip modifying BIRD config. This can be used to test that your config syntax is correct.")
	debug              = flag.Bool("debug", false, "Show debugging messages")
)

// Query PeeringDB for an ASN
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

// Use bgpq4 to generate a prefix filter and return only the filter lines
func getPrefixFilter(asSet string, family uint8, irrdb string) []string {
	// Run bgpq4 for BIRD format with aggregation enabled
	log.Infof("Running bgpq4 -h %s -Ab%d %s", irrdb, family, asSet)
	cmd := exec.Command("bgpq4", "-h", irrdb, "-Ab"+strconv.Itoa(int(family)), asSet)
	stdout, err := cmd.Output()
	if err != nil {
		log.Fatalf("bgpq4 error: %v", err.Error())
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

// Nonbuffered io Reader
func readNoBuffer(reader io.Reader) string {
	buf := make([]byte, 1024)
	n, err := reader.Read(buf[:])

	if err != nil {
		log.Fatalf("BIRD read error: ", err)
	}

	return string(buf[:n])
}

// Run a bird command
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

// Normalize a string to be filename-safe
func normalize(input string) string {
	// Remove non-alphanumeric characters
	input = sanitize.Path(input)

	// Make uppercase
	input = strings.ToUpper(input)

	// Replace spaces with underscores
	input = strings.ReplaceAll(input, " ", "_")

	// Replace slashes with dashes
	input = strings.ReplaceAll(input, "/", "-")

	return input
}

// Load a configuration file (YAML, JSON, or TOML)
func loadConfig() Config {
	configFile, err := ioutil.ReadFile(*configFilename)
	if err != nil {
		log.Fatalf("Reading %s: %v", *configFilename, err)
	}

	var config Config

	_splitFilename := strings.Split(*configFilename, ".")
	switch extension := _splitFilename[len(_splitFilename)-1]; extension {
	case "yaml", "yml":
		log.Info("Using YAML configuration format")
		err := yaml.Unmarshal(configFile, &config)
		if err != nil {
			log.Fatalf("YAML Unmarshal: %v", err)
		}
	case "toml":
		log.Info("Using TOML configuration format")
		err := toml.Unmarshal(configFile, &config)
		if err != nil {
			log.Fatalf("TOML Unmarshal: %v", err)
		}
	case "json":
		log.Info("Using JSON configuration format")
		err := json.Unmarshal(configFile, &config)
		if err != nil {
			log.Fatalf("JSON Unmarshal: %v", err)
		}
	default:
		log.Fatalf("Files with extension '%s' are not supported. (Acceptable values are yaml, toml, json", extension)
	}

	return config
}

func main() {
	// Enable debug logging in development releases
	if //noinspection GoBoolExpressions
	release == "devel" || *debug {
		log.SetLevel(log.DebugLevel)
	}

	flag.Usage = func() {
		fmt.Printf("Usage for bcg (%s) https://github.com/natesales/bcg:\n", release)
		flag.PrintDefaults()
	}
	flag.Parse()

	log.Info("Starting BCG")

	funcMap := template.FuncMap{
		"Contains": func(s, substr string) bool {
			return strings.Contains(s, substr)
		},

		"Iterate": func(count *uint) []uint {
			var i uint
			var Items []uint
			for i = 0; i < (*count); i++ {
				Items = append(Items, i)
			}
			return Items
		},

		"BirdSet": func(filter []string) string {
			// Build a formatted BIRD prefix list
			output := ""
			for i, prefix := range filter {
				output += "    " + prefix
				if i != len(filter)-1 {
					output += ",\n"
				}
			}

			return output
		},
	}

	log.Debug("Loading templates")

	// Generate peer template
	peerTemplate, err := template.New("").Funcs(funcMap).ParseFiles(path.Join(*templatesDirectory, "peer.tmpl"))
	if err != nil {
		log.Fatalf("Read peer specific template: %v", err)
	}

	// Generate global template
	globalTemplate, err := template.New("").Funcs(funcMap).ParseFiles(path.Join(*templatesDirectory, "global.tmpl"))
	if err != nil {
		log.Fatalf("Read peer specific template: %v", err)
	}

	log.Debug("Finished loading templates")

	// Load the config file from configFilename flag
	log.Debugf("Loading config from %s", *configFilename)
	config := loadConfig()
	log.Debug("Finished loading config")

	log.Debug("Linting global configuration")

	// Set default IRRDB
	if config.IrrDb == "" {
		config.IrrDb = "rr.ntt.net"
	}
	log.Infof("Using IRRDB server %s", config.IrrDb)

	// Set default RTR server
	if config.RtrServer == "" {
		config.RtrServer = "127.0.0.1"
	}
	log.Infof("Using RTR server %s", config.RtrServer)

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

	log.Debug("Finished linting global config")
	log.Debug("Writing global config")

	// Create the global output
	globalFile, err := os.Create(path.Join(*outputDirectory, "bird.conf"))
	if err != nil {
		log.Fatalf("Create global BIRD output file: %v", err)
	}

	log.Debug("Finished writing global config")
	log.Debug("Building origin sets")

	if len(config.Prefixes) == 0 {
		log.Fatal("There are no origin prefixes defined")
	}

	// Assemble originIpv{4,6} lists by address family
	var originIpv4, originIpv6 []string
	for _, prefix := range config.Prefixes {
		if strings.Contains(prefix, ":") {
			originIpv6 = append(originIpv6, prefix)
		} else {
			originIpv4 = append(originIpv4, prefix)
		}
	}

	log.Debug("Finished building origin sets")

	log.Debug("OriginIpv4: ", originIpv4)
	log.Debug("OriginIpv6: ", originIpv6)

	config.OriginSet4 = originIpv4
	config.OriginSet6 = originIpv6

	// Render the global template and write to disk
	if !*dryRun {
		log.Debug("Writing global config file")
		err = globalTemplate.ExecuteTemplate(globalFile, "global.tmpl", config)
		if err != nil {
			log.Fatalf("Execute template: %v", err)
		}
		log.Debug("Finished writing global config file")
	} else {
		log.Info("Dry run is enabled, skipped writing global config file")
	}

	// Iterate over peers
	for peerName, peerData := range config.Peers {
		// Set peerName
		peerData.Name = peerName

		// Set default query time
		peerData.QueryTime = "[No operations performed]"

		log.Infof("Checking config for %s AS%d", peerName, peerData.Asn)

		// Validate peer type
		if !(peerData.Type == "upstream" || peerData.Type == "peer" || peerData.Type == "downstream") {
			log.Fatalf("    type attribute is invalid. Must be upstream, peer, or downstream", peerName)
		}

		log.Infof("    type: %s", peerData.Type)

		// Set default local pref
		if peerData.LocalPref == 0 {
			peerData.LocalPref = 100
		}

		// Only query PeeringDB and IRRDB for peers and downstreams
		if peerData.Type != "upstream" {
			peerData.QueryTime = time.Now().Format(time.RFC1123)
			peeringDbData := getPeeringDbData(peerData.Asn)

			peerData.MaxPrefix4 = peeringDbData.MaxPfx4
			peerData.MaxPrefix6 = peeringDbData.MaxPfx6

			if strings.Contains(peeringDbData.AsSet, "::") {
				peerData.AsSet = strings.Split(peeringDbData.AsSet, "::")[1]
			} else {
				peerData.AsSet = peeringDbData.AsSet
			}

			peerData.PrefixSet4 = getPrefixFilter(peerData.AsSet, 4, config.IrrDb)
			peerData.PrefixSet6 = getPrefixFilter(peerData.AsSet, 6, config.IrrDb)

			// Update the "latest operation" timestamp
			peerData.QueryTime = time.Now().Format(time.RFC1123)
		} else { // If upstream
			peerData.MaxPrefix4 = 1000000 // 1M routes
			peerData.MaxPrefix6 = 100000  // 100k routes
		}

		log.Infof("    local pref: %d", peerData.LocalPref)
		log.Infof("    max prefixes: IPv4 %d, IPv6 %d", peerData.MaxPrefix4, peerData.MaxPrefix6)

		// Check for additional options
		if peerData.AsSet != "" {
			log.Infof("    as-set: %s", peerData.AsSet)
		}

		if peerData.Prepends > 0 {
			log.Infof("    prepends: %d", peerData.Prepends)
		}

		if peerData.Multihop {
			log.Infof("    multihop")
		}

		if peerData.Passive {
			log.Infof("    passive")
		}

		if peerData.Disabled {
			log.Infof("    disabled")
		}

		if peerData.PreImport != "" {
			log.Infof("    pre-import: %s", peerData.PreImport)
		}

		if peerData.PreExport != "" {
			log.Infof("    pre-export: %s", peerData.PreExport)
		}

		// Log neighbor IPs
		log.Infof("    neighbors:")
		for _, ip := range peerData.NeighborIps {
			log.Infof("      %s", ip)
		}

		if !*dryRun {
			// Create the peer specific file
			peerSpecificFile, err := os.Create(path.Join(*outputDirectory, "AS"+strconv.Itoa(int(peerData.Asn))+"_"+normalize(peerName)+".conf"))
			if err != nil {
				log.Fatalf("Create peer specific output file: %v", err)
			}

			//var pfxFilterString4, pfxFilterString6 = "", ""
			//
			//if peerData.ImportPolicy == "cone" {
			//	// Build prefix filter sets in BIRD format
			//	pfxFilterString4 = buildBirdSet(peerData.MaxPrefix4)
			//	pfxFilterString6 = buildBirdSet(peerData.MaxPrefix6)
			//}

			// Render the template and write to disk
			err = peerTemplate.ExecuteTemplate(peerSpecificFile, "peer.tmpl", &PeerTemplate{*peerData, config})
			if err != nil {
				log.Fatalf("Execute template: %v", err)
			}

			log.Infof("Wrote peer specific config for AS%d", peerData.Asn)
		} else {
			log.Infof("Dry run is enabled, skipped writing peer config")
		}
	}

	if !*dryRun {
		runBirdCommand("configure")
	}
}
