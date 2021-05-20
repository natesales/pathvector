package main

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

type peer struct {
	Description string `yaml:"description" description:"Peer description"`
	Disabled    bool   `yaml:"disabled" description:"Should the sessions be disabled?"`

	// BGP Attributes
	Asn              uint     `yaml:"asn" description:":Local ASN"`
	NeighborIPs      []string `yaml:"neighbors" description:"List of neighbor IPs"`
	Prepends         uint     `yaml:"prepends" description:"Number of times to prepend local AS on export"`
	LocalPref        uint     `yaml:"local-pref" description:"BGP local preference"`
	Multihop         bool     `yaml:"multihop" description:"Should BGP multihop be enabled? (255 max hops)"`
	Listen           string   `yaml:"listen" description:"BGP listen port"`
	NeighborPort     uint16   `yaml:"port" description:"Neighbor TCP port (default 179)"`
	Passive          bool     `yaml:"passive" description:"Should we listen passively?"`
	NextHopSelf      bool     `yaml:"next-hop-self" description:"Should BGP next-hop-self be enabled?"`
	Bfd              bool     `yaml:"bfd" description:"Should BFD be enabled?"`
	Communities      []string `yaml:"communities" description:"List of communities to add on export"`
	LargeCommunities []string `yaml:"large-communities" description:"List of large communities to add on export"`
	Password         string   `yaml:"password" description:"BGP MD5 password"`
	RsClient         bool     `yaml:"rs-client" description:"Should this peer be a route server client?"`
	RrClient         bool     `yaml:"rr-client" description:"Should this peer be a route reflector client?"`

	// Filtering
	Template           string `yaml:"template" description:"Template to inherit configuration from"`
	AsSet              string `yaml:"as-set" description:"Peer's as-set for filtering"`
	ImportLimit4       uint   `yaml:"import-limit4" description:"Maximum number of IPv4 prefixes to import"`
	ImportLimit6       uint   `yaml:"import-limit6" description:"Maximum number of IPv6 prefixes to import"`
	EnforceFirstAs     bool   `yaml:"enforce-first-as" description:"Should we only accept routes who's first AS is equal to the configured peer address?"`
	EnforcePeerNexthop bool   `yaml:"enforce-peer-nexthop" description:"Should we only accept routes with a next hop equal to the configured neighbor address?"`
	MaxPfxAction       string `yaml:"max-prefix-action" description:"What action should be taken when the max prefix limit is tripped?"`
	AllowBlackholes    bool   `yaml:"allow-blackholes" description:"Should this peer be allowed to send routes with the blackhole community?"`

	// Export options
	ExportDefault bool `yaml:"export-default" description:"Should a default route be exported to this peer?"`
	NoSpecifics   bool `yaml:"no-specifics" description:"Should more specific routes be exported to this peer?"`

	// Custom daemon configuration
	SessionGlobal  string `yaml:"session-global" description:"Configuration to add to each session before any defined BGP protocols"`
	PreImport      string `yaml:"pre-import" description:"Configuration to add before importing routes"`
	PreExport      string `yaml:"pre-export" description:"Configuration to add before exporting routes"`
	PreImportFinal string `yaml:"pre-import-final" description:"Configuration to add immediately before the final accept/reject on import"`
	PreExportFinal string `yaml:"pre-export-final" description:"Configuration to add immediately before the final accept/reject on export"`
}

type vrrpInstance struct {
	State     string   `yaml:"state" description:"VRRP instance state ('primary' or 'backup')"`
	Interface string   `yaml:"interface" description:"Interface to send VRRP packets on"`
	VRID      uint     `yaml:"vrid" description:"RFC3768 VRRP Virtual Router ID (1-255)"`
	Priority  uint8    `yaml:"priority" description:"RFC3768 VRRP Priority"`
	VIPs      []string `yaml:"vips" description:"List of virtual IPs"`
}

type runtimeConfig struct {
	BirdDirectory    string `yaml:"bird-directory" description:"Directory to store BIRD configs"`
	BirdSocket       string `yaml:"bird-socket" description:"UNIX control socket for BIRD"`
	KeepalivedConfig string `yaml:"keepalived-config" description:"Configuration file for keepalived"`
	WebUiFile        string `yaml:"web-ui-file" description:"File to write web UI to"`
}

// Config contains global configuration about this router and Wireframe instance
type Config struct {
	Runtime          *runtimeConfig    `yaml:"runtime" description:"Runtime configuration"`
	Asn              uint              `yaml:"asn" json:"asn" toml:"ASN"`
	RouterId         string            `yaml:"router-id" json:"router-id" toml:"Router-ID"`
	Prefixes         []string          `yaml:"prefixes" json:"prefixes" toml:"Prefixes"`
	Statics          map[string]string `yaml:"statics" json:"statics" toml:"Statics"`
	VRRPInstances    []*vrrpInstance   `yaml:"vrrp" json:"vrrp" toml:"VRRP"`
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
	KernelAccept4    []string          `yaml:"kernel-accept4" json:"kernel-accept4" toml:"KernelAccept4"`
	KernelAccept6    []string          `yaml:"kernel-accept6" json:"kernel-accept6" toml:"KernelAccept6"`
	KernelReject4    []string          `yaml:"kernel-reject4" json:"kernel-reject4" toml:"KernelReject4"`
	KernelReject6    []string          `yaml:"kernel-reject6" json:"kernel-reject6" toml:"KernelReject6"`

	Templates  map[string]*peer  `yaml:"templates" json:"templates" toml:"Templates"`
	Peers      map[string]*peer  `yaml:"peers" json:"peers" toml:"Peers"`
	Interfaces map[string]*iface `yaml:"interfaces" json:"interfaces" toml:"Interfaces"`

	OriginSet4 []string          `yaml:"-" json:"-" toml:"-"`
	OriginSet6 []string          `yaml:"-" json:"-" toml:"-"`
	Static4    map[string]string `yaml:"-" json:"-" toml:"-"`
	Static6    map[string]string `yaml:"-" json:"-" toml:"-"`
	Hostname   string            `yaml:"-" json:"-" toml:"-"`
}

// addr represents an IP address and netmask for easy YAML validation
type addr struct {
	Address net.IP
	Mask    uint8
}

// iface represents a network interface
type iface struct {
	Mtu       uint   `yaml:"mtu" json:"mtu" toml:"MTU"`
	XDPRTR    bool   `yaml:"xdprtr" json:"xdprtr" toml:"XDPRTR"`
	Addresses []addr `yaml:"addresses" json:"addresses" toml:"Addresses"`
	Dummy     bool   `yaml:"dummy" json:"dummy" toml:"Dummy"`
	Down      bool   `yaml:"down" json:"down" toml:"Down"`
}

// Wrapper stores a Peer and Config passed to the template
type Wrapper struct {
	Peer   peer
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

// setPeerDefaults sets the default values of a peer
func setPeerDefaults(name string, peer *peer) {
	// Set default local pref
	if peer.LocalPref == 0 {
		peer.LocalPref = 100
	}

	// Set default description
	if peer.Description == "" {
		peer.Description = "AS" + strconv.Itoa(int(peer.Asn)) + " " + name
	}

	// Set default max prefix violation action
	if peer.MaxPfxAction == "" {
		peer.MaxPfxAction = "disable"
	}
}

// loadConfig loads a configuration file from YAML, JSON, or TOML
func loadConfig(filename string) (*Config, error) {
	configFile, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, errorx.Decorate(err, "Reading config file")
	}

	var config Config

	_splitFilename := strings.Split(filename, ".")
	switch extension := _splitFilename[len(_splitFilename)-1]; extension {
	case "yaml", "yml":
		log.Info("Using YAML configuration format")
		err := yaml.UnmarshalStrict(configFile, &config)
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
