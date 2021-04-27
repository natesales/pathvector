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
	Asn                uint     `yaml:"asn" json:"asn" toml:"ASN"`
	Type               string   `yaml:"type" json:"type" toml:"Type"`
	Prepends           uint     `yaml:"prepends" json:"prepends" toml:"Prepends"`
	LocalPref          uint     `yaml:"local-pref" json:"local-pref" toml:"LocalPref"`
	Multihop           bool     `yaml:"multihop" json:"multihop" toml:"Multihop"`
	Passive            bool     `yaml:"passive" json:"passive" toml:"Passive"`
	Disabled           bool     `yaml:"disabled" json:"disabled" toml:"Disabled"`
	Password           string   `yaml:"password" json:"password" toml:"Password"`
	Port               uint16   `yaml:"port" json:"port" toml:"Port"`
	PreImport          string   `yaml:"pre-import" json:"pre-import" toml:"PreImport"`
	PreExport          string   `yaml:"pre-export" json:"pre-export" toml:"PreExport"`
	NeighborIPs        []string `yaml:"neighbors" json:"neighbors" toml:"Neighbors"`
	AsSet              string   `yaml:"as-set" json:"as-set" toml:"ASSet"`
	ImportLimit4       uint     `yaml:"import-limit4" json:"import-limit4" toml:"ImportLimit4"`
	ImportLimit6       uint     `yaml:"import-limit6" json:"import-limit6" toml:"ImportLimit6"`
	SkipFilter         bool     `yaml:"skip-filter" json:"skip-filter" toml:"SkipFilter"`
	RsClient           bool     `yaml:"rs-client" json:"rs-client" toml:"RSClient"`
	RrClient           bool     `yaml:"rr-client" json:"rr-client" toml:"RRClient"`
	Bfd                bool     `yaml:"bfd" json:"bfd" toml:"BFD"`
	EnforceFirstAs     bool     `yaml:"enforce-first-as" json:"enforce-first-as" toml:"EnforceFirstAS"`
	EnforcePeerNexthop bool     `yaml:"enforce-peer-nexthop" json:"enforce-peer-nexthop" toml:"EnforcePeerNexthop"`
	SessionGlobal      string   `yaml:"session-global" json:"session-global" toml:"SessionGlobal"`
	ExportDefault      bool     `yaml:"export-default" json:"export-default" toml:"ExportDefault"`
	NoSpecifics        bool     `yaml:"no-specifics" json:"no-specifics" toml:"NoSpecifics"`
	AllowBlackholes    bool     `yaml:"allow-blackholes" json:"allow-blackholes" toml:"AllowBlackholes"`
	Communities        []string `yaml:"communities" json:"communities" toml:"Communities"`
	LargeCommunities   []string `yaml:"large-communities" json:"large-communities" toml:"LargeCommunities"`
	Description        string   `yaml:"description" json:"description" toml:"Description"`
	Listen             string   `yaml:"listen" json:"listen" toml:"Listen"`

	QueryTime  string   `yaml:"-" json:"-" toml:"-"`
	Name       string   `yaml:"-" json:"-" toml:"-"`
	PrefixSet4 []string `yaml:"-" json:"-" toml:"-"`
	PrefixSet6 []string `yaml:"-" json:"-" toml:"-"`
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
	Asn              uint              `yaml:"asn" json:"asn" toml:"ASN"`
	RouterId         string            `yaml:"router-id" json:"router-id" toml:"Router-ID"`
	Prefixes         []string          `yaml:"prefixes" json:"prefixes" toml:"Prefixes"`
	Statics          map[string]string `yaml:"statics" json:"statics" toml:"Statics"`
	Peers            map[string]*Peer  `yaml:"peers" json:"peers" toml:"Peers"`
	VRRPInstances    []*VRRPInstance   `yaml:"vrrp" json:"vrrp" toml:"VRRP"`
	IrrDb            string            `yaml:"irrdb" json:"irrdb" toml:"IRRDB"`
	RtrServer        string            `yaml:"rtr-server" json:"rtr-server" toml:"RTR-Server"`
	RtrPort          int               `yaml:"rtr-port" json:"rtr-port" toml:"RTR-Port"`
	KeepFiltered     bool              `yaml:"keep-filtered" json:"keep-filtered" toml:"KeepFiltered"`
	MergePaths       bool              `yaml:"merge-paths" json:"merge-paths" toml:"MergePaths"`
	PrefSrc4         string            `yaml:"pref-src4" json:"pref-src4" toml:"PrefSrc4"`
	PrefSrc6         string            `yaml:"pref-src6" json:"pref-src6" toml:"PrefSrc6"`
	FilterDefault    bool              `yaml:"filter-default" json:"filter-default" toml:"FilterDefault"`
	DefaultEnabled   bool              `yaml:"enable-default" json:"enable-default" toml:"EnableDefault"`
	Communities      []string          `yaml:"communities" json:"communities" toml:"Communities"`
	LargeCommunities []string          `yaml:"large-communities" json:"large-communities" toml:"LargeCommunities"`
	KernelInject4    []string          `yaml:"kernel-inject4" json:"kernel-inject4" toml:"KernelInject4"`
	KernelInject6    []string          `yaml:"kernel-inject6" json:"kernel-inject6" toml:"KernelInject6"`

	OriginSet4 []string          `yaml:"-" json:"-" toml:"-"`
	OriginSet6 []string          `yaml:"-" json:"-" toml:"-"`
	Static4    map[string]string `yaml:"-" json:"-" toml:"-"`
	Static6    map[string]string `yaml:"-" json:"-" toml:"-"`
	Hostname   string            `yaml:"-" json:"-" toml:"-"`
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
	if config.RouterId != "" && net.ParseIP(config.RouterId).To4() == nil {
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
