package main

import (
	"errors"
	"io/ioutil"
	"net"

	"github.com/creasty/defaults"
	"github.com/go-playground/validator/v10"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

type peer struct {
	Description string `yaml:"description" description:"Peer description"`
	Disabled    bool   `yaml:"disabled" description:"Should the sessions be disabled?"`

	// BGP Attributes
	Asn              uint     `yaml:"asn" description:"Local ASN"`
	NeighborIPs      []string `yaml:"neighbors" description:"List of neighbor IPs"`
	Prepends         uint     `yaml:"prepends" description:"Number of times to prepend local AS on export"`
	LocalPref        uint     `yaml:"local-pref" description:"BGP local preference"`
	Multihop         bool     `yaml:"multihop" description:"Should BGP multihop be enabled? (255 max hops)"`
	Listen           string   `yaml:"listen" description:"BGP listen address"`
	LocalPort        uint16   `yaml:"local-port" description:"Local TCP port" default:"179"`
	NeighborPort     uint16   `yaml:"neighbor-port" description:"Neighbor TCP port" default:"179"`
	Passive          bool     `yaml:"passive" description:"Should we listen passively?" default:"false"`
	NextHopSelf      bool     `yaml:"next-hop-self" description:"Should BGP next-hop-self be enabled?" default:"false"`
	Bfd              bool     `yaml:"bfd" description:"Should BFD be enabled?" default:"false"`
	Communities      []string `yaml:"communities" description:"List of communities to add on export"`
	LargeCommunities []string `yaml:"large-communities" description:"List of large communities to add on export"`
	Password         string   `yaml:"password" description:"BGP MD5 password"`
	RsClient         bool     `yaml:"rs-client" description:"Should this peer be a route server client?" default:"false"`
	RrClient         bool     `yaml:"rr-client" description:"Should this peer be a route reflector client?" default:"false"`

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
	State     string   `yaml:"state" description:"VRRP instance state ('primary' or 'backup')" validate:"required"`
	Interface string   `yaml:"interface" description:"Interface to send VRRP packets on" validate:"required"`
	VRID      uint     `yaml:"vrid" description:"RFC3768 VRRP Virtual Router ID (1-255)" validate:"required"`
	Priority  uint8    `yaml:"priority" description:"RFC3768 VRRP Priority" validate:"required"`
	VIPs      []string `yaml:"vips" description:"List of virtual IPs" validate:"required,cidr"`

	VIPs4 []string `yaml:"-" description:"-"`
	VIPs6 []string `yaml:"-" description:"-"`
}

type runtimeConfig struct {
	BirdDirectory    string `yaml:"bird-directory" description:"Directory to store BIRD configs"`
	BirdSocket       string `yaml:"bird-socket" description:"UNIX control socket for BIRD"`
	KeepalivedConfig string `yaml:"keepalived-config" description:"Configuration file for keepalived"`
	WebUiFile        string `yaml:"web-ui-file" description:"File to write web UI to"`
}

type bgpConfig struct {
	Asn              uint     `yaml:"asn" description:"Autonomous System Number"`
	Prefixes         []string `yaml:"prefixes" description:"List of prefixes to announce" validate:"cidr"`
	Communities      []string `yaml:"communities" description:"List of RFC1997 BGP communities"`
	LargeCommunities []string `yaml:"large-communities" description:"List of RFC8092 large BGP communities"`

	Templates map[string]peer `yaml:"templates" description:"BGP peer template configuration"`
	Peers     map[string]peer `yaml:"peers" description:"BGP peer configuration"`

	Prefixes4 []string `yaml:"-" description:"-"`
	Prefixes6 []string `yaml:"-" description:"-"`
}

type augments struct {
	Accept4 []string          `yaml:"accept4" description:"List of BIRD protocols to import into the IPv4 table"`
	Accept6 []string          `yaml:"accept6" description:"List of BIRD protocols to import into the IPv6 table"`
	Reject4 []string          `yaml:"reject4" description:"List of BIRD protocols to not import into the IPv4 table"`
	Reject6 []string          `yaml:"reject6" description:"List of BIRD protocols to not import into the IPv6 table"`
	Statics map[string]string `yaml:"statics" description:"List of static routes to include in BIRD"`

	Statics4 map[string]string `yaml:"-" description:"-"`
	Statics6 map[string]string `yaml:"-" description:"-"`
}

type config struct {
	RouterId     string `yaml:"router-id" description:"Router ID (dotted quad notation)"`
	IrrServer    string `yaml:"irr-server" description:"Internet routing registry server" default:"rr.ntt.net"`
	RtrServer    string `yaml:"rtr-server" description:"RPKI-to-router server" default:"rtr.rpki.cloudflare.com"`
	RtrPort      int    `yaml:"rtr-port" description:"RPKI-to-router port" default:"8282"`
	KeepFiltered bool   `yaml:"keep-filtered" description:"Should filtered routes be kept in memory?"`
	MergePaths   bool   `yaml:"merge-paths" description:"Should best and equivalent non-best routes be imported for ECMP?"`
	Source4      string `yaml:"source4" description:"Source IPv4 address"`
	Source6      string `yaml:"source6" description:"Source IPv6 address"`

	Runtime runtimeConfig `yaml:"runtime" description:"Runtime configuration"`
	Bgp     bgpConfig     `yaml:"bgp" description:"BGP configuration"`

	Interfaces    map[string]iface `yaml:"interfaces" description:"Network interface configuration"`
	VRRPInstances []vrrpInstance   `yaml:"vrrp" description:"List of VRRP instances"`
	Augments      augments         `yaml:"augments" description:"Augments"`
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

// loadConfig loads a configuration file from YAML, JSON, or TOML
func loadConfig(filename string) (*config, error) {
	configFile, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, errors.New("reading config file: " + err.Error())
	}

	var config config
	if err := yaml.UnmarshalStrict(configFile, &config); err != nil {
		log.Fatalf("yaml unmarshal: %v", err)
	}

	validate := validator.New()
	if err := validate.Struct(&config); err != nil {
		log.Fatalf("validation: %+v", err)
	}

	if err := defaults.Set(&config); err != nil {
		log.Fatalf("defaults: %+v", err)
	}

	// Parse origin routes by assembling OriginIPv{4,6} lists by address family
	for _, prefix := range config.Bgp.Prefixes {
		pfx, _, err := net.ParseCIDR(prefix)
		if err != nil {
			return nil, errors.New("invalid origin prefix: " + prefix)
		}

		if pfx.To4() == nil { // If IPv6
			config.Bgp.Prefixes4 = append(config.Bgp.Prefixes4, prefix)
		} else { // If IPv4
			config.Bgp.Prefixes6 = append(config.Bgp.Prefixes6, prefix)
		}
	}

	// Initialize static maps
	config.Augments.Statics4 = map[string]string{}
	config.Augments.Statics6 = map[string]string{}

	// Parse static routes
	for prefix, nexthop := range config.Augments.Statics {
		pfx, _, err := net.ParseCIDR(prefix)
		if err != nil {
			return nil, errors.New("invalid static prefix: " + prefix)
		}
		if net.ParseIP(nexthop) == nil {
			return nil, errors.New("invalid static nexthop: " + nexthop)
		}

		if pfx.To4() == nil { // If IPv6
			config.Augments.Statics6[prefix] = nexthop
		} else { // If IPv4
			config.Augments.Statics4[prefix] = nexthop
		}
	}

	// Parse VRRP configs
	for _, vrrpInstance := range config.VRRPInstances {
		// Sort VIPs by address family
		for _, vip := range vrrpInstance.VIPs {
			ip, _, err := net.ParseCIDR(vip)
			if err != nil {
				return nil, errors.New("invalid VIP: " + vip)
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
	}

	return &config, nil // nil error
}
