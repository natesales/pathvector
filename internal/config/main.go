package config

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net"
	"strconv"
	"strings"

	"github.com/joomcode/errorx"
	"github.com/pelletier/go-toml"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

// Default constants
const (
	DefaultRtrServer = "127.0.0.1"
	DefaultRtrPort   = 8282

	DefaultIRRServer = "rr.ntt.net"
)

// Peer contains all information specific to a single peer network
type Peer struct {
	Asn                uint     `yaml:"asn" toml:"ASN" json:"asn"`
	Type               string   `yaml:"type" toml:"Type" json:"type"`
	Prepends           uint     `yaml:"prepends" toml:"Prepends" json:"prepends"`
	LocalPref          uint     `yaml:"local-pref" toml:"LocalPref" json:"local-pref"`
	Multihop           bool     `yaml:"multihop" toml:"Multihop" json:"multihop"`
	Passive            bool     `yaml:"passive" toml:"Passive" json:"passive"`
	Disabled           bool     `yaml:"disabled" toml:"Disabled" json:"disabled"`
	Password           string   `yaml:"password" toml:"Password" json:"password"`
	Port               uint16   `yaml:"port" toml:"Port" json:"port"`
	PreImport          string   `yaml:"pre-import" toml:"PreImport" json:"pre-import"`
	PreExport          string   `yaml:"pre-export" toml:"PreExport" json:"pre-export"`
	NeighborIps        []string `yaml:"neighbors" toml:"Neighbors" json:"neighbors"`
	AsSet              string   `yaml:"as-set" toml:"ASSet" json:"as-set"`
	ImportLimit4       uint     `yaml:"import-limit4" toml:"ImportLimit4" json:"import-limit4"`
	ImportLimit6       uint     `yaml:"import-limit6" toml:"ImportLimit6" json:"import-limit6"`
	SkipFilter         bool     `yaml:"skip-filter" toml:"SkipFilter" json:"skip-filter"`
	RsClient           bool     `yaml:"rs-client" toml:"RSClient" json:"rs-client"`
	RrClient           bool     `yaml:"rr-client" toml:"RRClient" json:"rr-client"`
	Bfd                bool     `yaml:"bfd" toml:"BFD" json:"bfd"`
	EnforceFirstAs     bool     `yaml:"enforce-first-as" toml:"EnforceFirstAS" json:"enforce-first-as"`
	EnforcePeerNexthop bool     `yaml:"enforce-peer-nexthop" toml:"EnforcePeerNexthop" json:"enforce-peer-nexthop"`
	SessionGlobal      string   `yaml:"session-global" toml:"SessionGlobal" json:"session-global"`
	ExportDefault      bool     `yaml:"export-default" toml:"ExportDefault" json:"export-default"`
	NoSpecifics        bool     `yaml:"no-specifics" toml:"NoSpecifics" json:"no-specifics"`
	AllowBlackholes    bool     `yaml:"allow-blackholes" toml:"AllowBlackholes" json:"allow-blackholes"`
	StripPrivateASNs   bool     `yaml:"strip-private-asns" toml:"StripPrivateASNs" json:"strip-private-asns"`
	Communities        []string `yaml:"communities" toml:"Communities" json:"communities"`
	LargeCommunities   []string `yaml:"large-communities" toml:"LargeCommunities" json:"large-communities"`
	Description        string   `yaml:"description" toml:"Description" json:"description"`

	QueryTime  string   `yaml:"-" toml:"-" json:"-"`
	Name       string   `yaml:"-" toml:"-" json:"-"`
	PrefixSet4 []string `yaml:"-" toml:"-" json:"-"`
	PrefixSet6 []string `yaml:"-" toml:"-" json:"-"`
}

// Config contains global configuration about this router and BCG instance
type Config struct {
	Asn            uint             `yaml:"asn" toml:"ASN" json:"asn"`
	RouterId       string           `yaml:"router-id" toml:"Router-ID" json:"router-id"`
	Prefixes       []string         `yaml:"prefixes" toml:"Prefixes" json:"prefixes"`
	Peers          map[string]*Peer `yaml:"peers" toml:"Peers" json:"peers"`
	IrrDb          string           `yaml:"irrdb" toml:"IRRDB" json:"irrdb"`
	RtrServer      string           `yaml:"rtr-server" toml:"RTR-Server" json:"rtr-server"`
	RtrPort        int              `yaml:"rtr-port" toml:"RTR-Port" json:"rtr-port"`
	KeepFiltered   bool             `yaml:"keep-filtered" toml:"KeepFiltered" json:"keep-filtered"`
	MergePaths     bool             `yaml:"merge-paths" toml:"MergePaths" json:"merge-paths"`
	PrefSrc4       string           `yaml:"pref-src4" toml:"PrefSrc4" json:"PrefSrc4"`
	PrefSrc6       string           `yaml:"pref-src6" toml:"PrefSrc6" json:"PrefSrc6"`
	FilterDefault  bool             `yaml:"filter-default" toml:"FilterDefault" json:"filter-default"`
	DefaultEnabled bool             `yaml:"enable-default" toml:"EnableDefault" json:"enable-default"`

	OriginSet4 []string `yaml:"-" toml:"-" json:"-"`
	OriginSet6 []string `yaml:"-" toml:"-" json:"-"`
	Hostname   string   `yaml:"-" toml:"-" json:"-"`
}

// Wrapper stores a Peer and Config passed to the template
type Wrapper struct {
	Peer   Peer
	Config Config
}

// setConfigDefaults sets the default values of a Config struct
func setConfigDefaults(config *Config) error {
	// Set default IRRDB
	if config.IrrDb == "" {
		config.IrrDb = DefaultIRRServer
	}

	// Set default RTR server
	if config.RtrServer == "" {
		config.RtrServer = DefaultRtrServer
	}

	// Set default RTR port
	if config.RtrPort == 0 {
		config.RtrPort = DefaultRtrPort
	}

	// Validate Router ID in dotted quad format
	if net.ParseIP(config.RouterId).To4() == nil {
		return errors.New("Router ID " + config.RouterId + " is not in valid dotted quad notation")
	}

	// Validate CIDR notation of originated prefixes
	for _, addr := range config.Prefixes {
		if _, _, err := net.ParseCIDR(addr); err != nil {
			return errors.New(addr + " is not a valid IPv4 or IPv6 prefix in CIDR notation")
		}
	}

	return nil // nil error
}

// setPeerDefaults sets the default values of a Peer struct
func setPeerDefaults(name string, peer *Peer) {
	// Set default local pref
	if peer.LocalPref == 0 {
		peer.LocalPref = 100
	}

	// Set default description
	if peer.Description == "" {
		peer.Description = "AS" + strconv.Itoa(int(peer.Asn)) + " " + name
	}
}

// Load loads a configuration file (YAML, JSON, or TOML)
func Load(filename string) (*Config, error) {
	configFile, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, errorx.Decorate(err, "reading config file")
	}

	var config Config

	_splitFilename := strings.Split(filename, ".")
	switch extension := _splitFilename[len(_splitFilename)-1]; extension {
	case "yaml", "yml":
		log.Info("Using YAML configuration format")
		err := yaml.Unmarshal(configFile, &config)
		if err != nil {
			return nil, errorx.Decorate(err, "YAML unmarshal")
		}
	case "toml":
		log.Info("Using TOML configuration format")
		err := toml.Unmarshal(configFile, &config)
		if err != nil {
			return nil, errorx.Decorate(err, "TOML unmarshal")
		}
	case "json":
		log.Info("Using JSON configuration format")
		err := json.Unmarshal(configFile, &config)
		if err != nil {
			return nil, errorx.Decorate(err, "JSON unmarshal")
		}
	default:
		return nil, errors.New("Files with extension " + extension + " are not supported. Acceptable values are yaml, toml, json")
	}

	// Set global config defaults
	err = setConfigDefaults(&config)
	if err != nil {
		return nil, err
	}

	// Set individual peer defaults
	for name, peer := range config.Peers {
		setPeerDefaults(name, peer)
	}

	return &config, nil // nil error
}
