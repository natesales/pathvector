package config

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net"
	"os"
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
	NeighborIPs        []string `yaml:"neighbors" toml:"Neighbors" json:"neighbors"`
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
	Communities        []string `yaml:"communities" toml:"Communities" json:"communities"`
	LargeCommunities   []string `yaml:"large-communities" toml:"LargeCommunities" json:"large-communities"`
	Description        string   `yaml:"description" toml:"Description" json:"description"`

	QueryTime  string   `yaml:"-" toml:"-" json:"-"`
	Name       string   `yaml:"-" toml:"-" json:"-"`
	PrefixSet4 []string `yaml:"-" toml:"-" json:"-"`
	PrefixSet6 []string `yaml:"-" toml:"-" json:"-"`
}

// VRRPInstance stores a VRRP instance
type VRRPInstance struct {
	State     string   `yaml:"state" json:"state" toml:"State"`
	Interface string   `yaml:"interface" json:"interface" toml:"Interface"`
	VRRID     uint     `yaml:"vrrid" json:"vrrid" toml:"VRRID"`
	Priority  uint8    `yaml:"priority" json:"priority" toml:"Priority"`
	VIPs      []string `yaml:"vips" json:"vips" toml:"VIPs"`

	VIPs4 []string `yaml:"-" json:"-" toml:"-"`
	VIPs6 []string `yaml:"-" json:"-" toml:"-"`
}

// Config contains global configuration about this router and BCG instance
type Config struct {
	Asn            uint              `yaml:"asn" toml:"ASN" json:"asn"`
	RouterId       string            `yaml:"router-id" toml:"Router-ID" json:"router-id"`
	Prefixes       []string          `yaml:"prefixes" toml:"Prefixes" json:"prefixes"`
	Statics        map[string]string `yaml:"statics" toml:"Statics" json:"statics"`
	Peers          map[string]*Peer  `yaml:"peers" toml:"Peers" json:"peers"`
	VRRPInstances  []*VRRPInstance   `yaml:"vrrp" toml:"VRRP" json:"vrrp"`
	IrrDb          string            `yaml:"irrdb" toml:"IRRDB" json:"irrdb"`
	RtrServer      string            `yaml:"rtr-server" toml:"RTR-Server" json:"rtr-server"`
	RtrPort        int               `yaml:"rtr-port" toml:"RTR-Port" json:"rtr-port"`
	KeepFiltered   bool              `yaml:"keep-filtered" toml:"KeepFiltered" json:"keep-filtered"`
	MergePaths     bool              `yaml:"merge-paths" toml:"MergePaths" json:"merge-paths"`
	PrefSrc4       string            `yaml:"pref-src4" toml:"PrefSrc4" json:"PrefSrc4"`
	PrefSrc6       string            `yaml:"pref-src6" toml:"PrefSrc6" json:"PrefSrc6"`
	FilterDefault  bool              `yaml:"filter-default" toml:"FilterDefault" json:"filter-default"`
	DefaultEnabled bool              `yaml:"enable-default" toml:"EnableDefault" json:"enable-default"`

	OriginSet4 []string          `yaml:"-" toml:"-" json:"-"`
	OriginSet6 []string          `yaml:"-" toml:"-" json:"-"`
	Static4    map[string]string `yaml:"-" toml:"-" json:"-"`
	Static6    map[string]string `yaml:"-" toml:"-" json:"-"`
	Hostname   string            `yaml:"-" toml:"-" json:"-"`
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
			return errors.New("Address " + addr + " is not a valid IPv4 or IPv6 prefix in CIDR notation")
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
		return nil, errorx.Decorate(err, "Reading config file")
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

	// Get hostname
	config.Hostname, err = os.Hostname()
	if err != nil {
		return nil, errorx.Decorate(err, "Getting hostname")
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

	// Parse origin routes by assembling OriginIPv{4,6} lists by address family
	for _, prefix := range config.Prefixes {
		pfx, _, err := net.ParseCIDR(prefix)
		if err != nil {
			return nil, errorx.Decorate(err, "Invalid origin prefix: "+prefix)
		}

		if pfx.To4() == nil { // If IPv6
			config.OriginSet6 = append(config.OriginSet6, prefix)
		} else { // If IPv4
			config.OriginSet4 = append(config.OriginSet4, prefix)
		}
	}

	log.Info("Origin IPv4: ", config.OriginSet4)
	log.Info("Origin IPv6: ", config.OriginSet6)

	// Initialize static maps
	config.Static4 = map[string]string{}
	config.Static6 = map[string]string{}

	// Parse static routes
	for prefix, nexthop := range config.Statics {
		pfx, _, err := net.ParseCIDR(prefix)
		if err != nil {
			return nil, errorx.Decorate(err, "Invalid static prefix: "+prefix)
		}
		if net.ParseIP(nexthop) == nil {
			return nil, errorx.Decorate(err, "Invalid static next hop: "+nexthop)
		}

		if pfx.To4() == nil { // If IPv6
			config.Static6[prefix] = nexthop
		} else { // If IPv4
			config.Static4[prefix] = nexthop
		}
	}

	// Print static routes
	if len(config.Static4) > 0 {
		log.Infof("IPv4 statics: %+v", config.Static4)
	}
	if len(config.Static6) > 0 {
		log.Infof("IPv6 statics: %+v", config.Static6)
	}

	// Parse VRRP configs
	for _, vrrpInstance := range config.VRRPInstances {
		// Sort VIPs by address family
		for _, vip := range vrrpInstance.VIPs {
			ip, _, err := net.ParseCIDR(vip)
			if err != nil {
				return nil, errorx.Decorate(err, "Invalid VIP")
			}

			if ip.To4() == nil { // If IPv6
				vrrpInstance.VIPs6 = append(vrrpInstance.VIPs6, vip)
			} else { // If IPv4
				vrrpInstance.VIPs4 = append(vrrpInstance.VIPs4, vip)
			}
		}

		// Validate vrrpInstance
		if vrrpInstance.State == "primary" {
			vrrpInstance.State = "MASTER"
		} else if vrrpInstance.State == "backup" {
			vrrpInstance.State = "BACKUP"
		} else {
			return nil, errors.New("VRRP state must be 'primary' or 'backup', unexpected " + vrrpInstance.State)
		}

		if vrrpInstance.Interface == "" {
			return nil, errors.New("VRRP interface is not defined")
		}

		if len(vrrpInstance.VIPs) < 1 {
			return nil, errors.New("VRRP instance must have at least one VIP defined")
		}
	}

	return &config, nil // nil error
}
