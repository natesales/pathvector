package config

import (
	"github.com/creasty/defaults"
	"github.com/go-ping/ping"
)

// Peer stores a single peer config
type Peer struct {
	Template *string `yaml:"template" description:"Configuration template" default:"-"`

	Description *string `yaml:"description" description:"Peer description" default:"-"`
	Disabled    *bool   `yaml:"disabled" description:"Should the sessions be disabled?" default:"false"`

	// BGP Attributes
	ASN                  *int      `yaml:"asn" description:"Local ASN" validate:"required" default:"0"`
	NeighborIPs          *[]string `yaml:"neighbors" description:"List of neighbor IPs" validate:"required,ip" default:"-"`
	Prepends             *int      `yaml:"prepends" description:"Number of times to prepend local AS on export" default:"0"`
	LocalPref            *int      `yaml:"local-pref" description:"BGP local preference" default:"100"`
	SetLocalPref         *bool     `yaml:"set-local-pref" description:"Should an explicit local pref be set?" default:"true"`
	Multihop             *bool     `yaml:"multihop" description:"Should BGP multihop be enabled? (255 max hops)" default:"false"`
	Listen4              *string   `yaml:"listen4" description:"IPv4 BGP listen address" default:"-"`
	Listen6              *string   `yaml:"listen6" description:"IPv6 BGP listen address" default:"-"`
	LocalASN             *int      `yaml:"local-asn" description:"Local ASN as defined in the global ASN field" default:"-"`
	LocalPort            *int      `yaml:"local-port" description:"Local TCP port" default:"179"`
	NeighborPort         *int      `yaml:"neighbor-port" description:"Neighbor TCP port" default:"179"`
	Passive              *bool     `yaml:"passive" description:"Should we listen passively?" default:"false"`
	Direct               *bool     `yaml:"direct" description:"Specify that the neighbor is directly connected" default:"false"`
	NextHopSelf          *bool     `yaml:"next-hop-self" description:"Should BGP next-hop-self be enabled?" default:"false"`
	BFD                  *bool     `yaml:"bfd" description:"Should BFD be enabled?" default:"false"`
	Password             *string   `yaml:"password" description:"BGP MD5 password" default:"-"`
	RSClient             *bool     `yaml:"rs-client" description:"Should this peer be a route server client?" default:"false"`
	RRClient             *bool     `yaml:"rr-client" description:"Should this peer be a route reflector client?" default:"false"`
	RemovePrivateASNs    *bool     `yaml:"remove-private-asns" description:"Should private ASNs be removed from path before exporting?" default:"true"`
	MPUnicast46          *bool     `yaml:"mp-unicast-46" description:"Should this peer be configured with multiprotocol IPv4 and IPv6 unicast?" default:"false"`
	AllowLocalAS         *bool     `yaml:"allow-local-as" description:"Should routes originated by the local ASN be accepted?" default:"false"`
	AddPathTx            *bool     `yaml:"add-path-tx" description:"Enable BGP additional paths on export?" default:"false"`
	AddPathRx            *bool     `yaml:"add-path-rx" description:"Enable BGP additional paths on import?" default:"false"`
	ImportNextHop        *string   `yaml:"import-next-hop" description:"Rewrite the BGP next hop before importing routes learned from this peer" default:"-"`
	ExportNextHop        *string   `yaml:"export-next-hop" description:"Rewrite the BGP next hop before announcing routes to this peer" default:"-"`
	Confederation        *int      `yaml:"confederation" description:"BGP confederation (RFC 5065)" default:"-"`
	ConfederationMember  *bool     `yaml:"confederation-member" description:"Should this peer be a member of the local confederation?" default:"false"`
	TTLSecurity          *bool     `yaml:"ttl-security" description:"RFC 5082 Generalized TTL Security Mechanism" default:"false"`
	InterpretCommunities *bool     `yaml:"interpret-communities" description:"Should well-known BGP communities be interpreted by their intended action?" default:"true"`
	DefaultLocalPref     *int      `yaml:"default-local-pref" description:"Default value for local preference" default:"-"`
	AdvertiseHostname    *bool     `yaml:"advertise-hostname" description:"Advertise hostname capability" default:"false"`
	DisableAfterError    *bool     `yaml:"disable-after-error" description:"Disable peer after error" default:"false"`
	PreferOlderRoutes    *bool     `yaml:"prefer-older-routes" description:"Prefer older routes instead of comparing router IDs (RFC 5004)" default:"false"`

	ImportCommunities    *[]string `yaml:"import-communities" description:"List of communities to add to all imported routes" default:"-"`
	ExportCommunities    *[]string `yaml:"export-communities" description:"List of communities to add to all exported routes" default:"-"`
	AnnounceCommunities  *[]string `yaml:"announce-communities" description:"Announce all routes matching these communities to the peer" default:"-"`
	RemoveCommunities    *[]string `yaml:"remove-communities" description:"List of communities to remove before from routes announced by this peer" default:"-"`
	RemoveAllCommunities *int      `yaml:"remove-all-communities" description:"Remove all standard and large communities beginning with this value" default:"-"`

	ASPrefs *map[uint32]uint32 `yaml:"as-prefs" description:"Map of ASN to import local pref (not included in optimizer)" default:"-"`

	// Filtering
	ASSet                   *string `yaml:"as-set" description:"Peer's as-set for filtering" default:"-"`
	ImportLimit4            *int    `yaml:"import-limit4" description:"Maximum number of IPv4 prefixes to import" default:"1000000"`
	ImportLimit6            *int    `yaml:"import-limit6" description:"Maximum number of IPv6 prefixes to import" default:"200000"`
	EnforceFirstAS          *bool   `yaml:"enforce-first-as" description:"Should we only accept routes who's first AS is equal to the configured peer address?" default:"true"`
	EnforcePeerNexthop      *bool   `yaml:"enforce-peer-nexthop" description:"Should we only accept routes with a next hop equal to the configured neighbor address?" default:"true"`
	ForcePeerNexthop        *bool   `yaml:"force-peer-nexthop" description:"Rewrite nexthop to peer address" default:"false"`
	MaxPrefixTripAction     *string `yaml:"max-prefix-action" description:"What action should be taken when the max prefix limit is tripped?" default:"disable"`
	AllowBlackholeCommunity *bool   `yaml:"allow-blackhole-community" description:"Should this peer be allowed to send routes with the blackhole community?" default:"false"`
	BlackholeIn             *bool   `yaml:"blackhole-in" description:"Should imported routes be blackholed?" default:"false"`
	BlackholeOut            *bool   `yaml:"blackhole-out" description:"Should exported routes be blackholed?" default:"false"`

	FilterIRR                  *bool `yaml:"filter-irr" description:"Should IRR filtering be applied?" default:"false"`
	FilterRPKI                 *bool `yaml:"filter-rpki" description:"Should RPKI invalids be rejected?" default:"true"`
	StrictRPKI                 *bool `yaml:"strict-rpki" description:"Should only RPKI valids be accepted?" default:"false"`
	FilterMaxPrefix            *bool `yaml:"filter-max-prefix" description:"Should max prefix filtering be applied?" default:"true"`
	FilterBogonRoutes          *bool `yaml:"filter-bogon-routes" description:"Should bogon prefixes be rejected?" default:"true"`
	FilterBogonASNs            *bool `yaml:"filter-bogon-asns" description:"Should paths containing a bogon ASN be rejected?" default:"true"`
	FilterTransitASNs          *bool `yaml:"filter-transit-asns" description:"Should paths containing transit-free ASNs be rejected? (Peerlock Lite)'" default:"false"`
	FilterPrefixLength         *bool `yaml:"filter-prefix-length" description:"Should too large/small prefixes (IPv4 8 > len > 24 and IPv6 12 > len > 48) be rejected?" default:"true"`
	FilterNeverViaRouteServers *bool `yaml:"filter-never-via-route-servers" description:"Should routes containing an ASN reported in PeeringDB to never be reachable via route servers be filtered?" default:"false"`

	AutoImportLimits *bool `yaml:"auto-import-limits" description:"Get import limits automatically from PeeringDB?" default:"false"`
	AutoASSet        *bool `yaml:"auto-as-set" description:"Get as-set automatically from PeeringDB? If no as-set exists in PeeringDB, a warning will be shown and the peer ASN used instead." default:"false"`

	HonorGracefulShutdown *bool `yaml:"honor-graceful-shutdown" description:"Should RFC8326 graceful shutdown be enabled?" default:"true"`

	Prefixes *[]string `yaml:"prefixes" description:"Prefixes to accept" default:"-"`

	// Export options
	AnnounceDefault    *bool `yaml:"announce-default" description:"Should a default route be exported to this peer?" default:"false"`
	AnnounceOriginated *bool `yaml:"announce-originated" description:"Should locally originated routes be announced to this peer?" default:"true"`
	AnnounceAll        *bool `yaml:"announce-all" description:"Should all routes be exported to this peer?" default:"false"`

	// Custom daemon configuration
	SessionGlobal *string `yaml:"session-global" description:"Configuration to add to each session before any defined BGP protocols" default:"-"`

	PreImport      *string `yaml:"pre-import" description:"Configuration to add at the beginning of the import filter" default:"-"`
	PreExport      *string `yaml:"pre-export" description:"Configuration to add at the beginning of the export filter" default:"-"`
	PreImportFinal *string `yaml:"pre-import-final" description:"Configuration to add immediately before the final accept/reject on import" default:"-"`
	PreExportFinal *string `yaml:"pre-export-final" description:"Configuration to add immediately before the final accept/reject on export" default:"-"`

	PreImportFile      *string `yaml:"pre-import-file" description:"Configuration file to append to pre-import" default:"-"`
	PreExportFile      *string `yaml:"pre-export-file" description:"Configuration file to append to pre-export" default:"-"`
	PreImportFinalFile *string `yaml:"pre-import-final-file" description:"Configuration file to append to pre-import-final" default:"-"`
	PreExportFinalFile *string `yaml:"pre-export-final-file" description:"Configuration file to append to pre-export-final" default:"-"`

	// Optimizer
	OptimizerProbeSources *[]string `yaml:"probe-sources" description:"Optimizer probe source addresses" default:"-"`
	OptimizeInbound       *bool     `yaml:"optimize-inbound" description:"Should the optimizer modify inbound policy?" default:"false"`

	ProtocolName                *string   `yaml:"-" description:"-" default:"-"`
	Protocols                   *[]string `yaml:"-" description:"-" default:"-"`
	PrefixSet4                  *[]string `yaml:"-" description:"-" default:"-"`
	PrefixSet6                  *[]string `yaml:"-" description:"-" default:"-"`
	ImportStandardCommunities   *[]string `yaml:"-" description:"-" default:"-"`
	ImportLargeCommunities      *[]string `yaml:"-" description:"-" default:"-"`
	ExportStandardCommunities   *[]string `yaml:"-" description:"-" default:"-"`
	ExportLargeCommunities      *[]string `yaml:"-" description:"-" default:"-"`
	AnnounceStandardCommunities *[]string `yaml:"-" description:"-" default:"-"`
	AnnounceLargeCommunities    *[]string `yaml:"-" description:"-" default:"-"`
	RemoveStandardCommunities   *[]string `yaml:"-" description:"-" default:"-"`
	RemoveLargeCommunities      *[]string `yaml:"-" description:"-" default:"-"`
	BooleanOptions              *[]string `yaml:"-" description:"-" default:"-"`
}

// VRRPInstance stores a single VRRP instance
type VRRPInstance struct {
	State     string   `yaml:"state" description:"VRRP instance state ('primary' or 'backup')" validate:"required"`
	Interface string   `yaml:"interface" description:"Interface to send VRRP packets on" validate:"required"`
	VRID      uint     `yaml:"vrid" description:"RFC3768 VRRP Virtual Router ID (1-255)" validate:"required"`
	Priority  uint     `yaml:"priority" description:"RFC3768 VRRP Priority" validate:"required"`
	VIPs      []string `yaml:"vips" description:"List of virtual IPs" validate:"required,cidr"`

	VIPs4 []string `yaml:"-" description:"-"`
	VIPs6 []string `yaml:"-" description:"-"`
}

// BFDInstance stores a single BFD instance
type BFDInstance struct {
	Neighbor   *string `yaml:"neighbor" description:"Neighbor IP address" default:"-"`
	Interface  *string `yaml:"interface" description:"Interface (pattern accepted)" default:"-"`
	Interval   *uint   `yaml:"interval" description:"RX and TX interval" default:"200"`
	Multiplier *uint   `yaml:"multiplier" description:"Number of missed packets for the state to be declared down" default:"10"`

	ProtocolName *string `yaml:"-" description:"-" default:"-"`
}

// MRTInstance stores a single MRT instance
type MRTInstance struct {
	File     *string `yaml:"file" description:"File to store MRT dumps (supports strftime replacements and %N as table name)" default:"/var/log/bird/%N_%F_%T.mrt"`
	Interval *uint   `yaml:"interval" description:"Number of seconds between dumps" default:"300"`
	Table    *string `yaml:"table" description:"Routing table to read from" default:"-"`
}

// Augments store BIRD specific options
type Augments struct {
	Accept4        []string          `yaml:"accept4" description:"List of BIRD protocols to import into the IPv4 table"`
	Accept6        []string          `yaml:"accept6" description:"List of BIRD protocols to import into the IPv6 table"`
	Reject4        []string          `yaml:"reject4" description:"List of BIRD protocols to not import into the IPv4 table"`
	Reject6        []string          `yaml:"reject6" description:"List of BIRD protocols to not import into the IPv6 table"`
	Statics        map[string]string `yaml:"statics" description:"List of static routes to include in BIRD"`
	SRDCommunities []string          `yaml:"srd-communities" description:"List of communities to filter routes exported to kernel (if list is not empty, all other prefixes will not be exported)"`

	SRDStandardCommunities []string          `yaml:"-" description:"-"`
	SRDLargeCommunities    []string          `yaml:"-" description:"-"`
	Statics4               map[string]string `yaml:"-" description:"-"`
	Statics6               map[string]string `yaml:"-" description:"-"`
}

// ProbeResult stores a single probe result
type ProbeResult struct {
	Time  int64
	Stats ping.Statistics
}

// Optimizer stores route optimizer configuration
type Optimizer struct {
	Targets             []string `yaml:"targets" description:"List of probe targets"`
	LatencyThreshold    uint     `yaml:"latency-threshold" description:"Maximum allowable latency in milliseconds" default:"100"`
	PacketLossThreshold float64  `yaml:"packet-loss-threshold" description:"Maximum allowable packet loss (percent)" default:"0.5"`
	LocalPrefModifier   uint     `yaml:"modifier" description:"Amount to lower local pref by for depreferred peers" default:"20"`

	PingCount   int `yaml:"probe-count" description:"Number of pings to send in each run" default:"5"`
	PingTimeout int `yaml:"probe-timeout" description:"Number of seconds to wait before considering the ICMP message unanswered" default:"1"`
	Interval    int `yaml:"probe-interval" description:"Number of seconds wait between each optimizer run" default:"120"`
	CacheSize   int `yaml:"cache-size" description:"Number of probe results to store per peer" default:"15"`

	ProbeUDPMode bool `yaml:"probe-udp" description:"Use UDP probe (else ICMP)" default:"false"`

	AlertScript string `yaml:"alert-script" description:"Script to call on optimizer event"`

	ExitOnCacheFull bool `yaml:"exit-on-cache-full" description:"Exit optimizer on cache full" default:"false"`

	Db map[string][]ProbeResult `yaml:"-" description:"-"`
}

// Config stores the global configuration
type Config struct {
	PeeringDBQueryTimeout uint   `yaml:"peeringdb-query-timeout" description:"PeeringDB query timeout in seconds" default:"10"`
	PeeringDBAPIKey       string `yaml:"peeringdb-api-key" description:"PeeringDB API key"`
	PeeringDBCache        bool   `yaml:"peeringdb-cache" description:"Cache PeeringDB results" default:"true"`
	IRRQueryTimeout       uint   `yaml:"irr-query-timeout" description:"IRR query timeout in seconds" default:"30"`
	BIRDDirectory         string `yaml:"bird-directory" description:"Directory to store BIRD configs" default:"/etc/bird/"`
	BIRDBinary            string `yaml:"bird-binary" description:"Path to BIRD binary" default:"/usr/sbin/bird"`
	BIRDSocket            string `yaml:"bird-socket" description:"UNIX control socket for BIRD" default:"/run/bird/bird.ctl"`
	CacheDirectory        string `yaml:"cache-directory" description:"Directory to store runtime configuration cache" default:"/var/run/pathvector/cache/"`
	KeepalivedConfig      string `yaml:"keepalived-config" description:"Configuration file for keepalived" default:"/etc/keepalived.conf"`
	WebUIFile             string `yaml:"web-ui-file" description:"File to write web UI to (disabled if empty)" default:""`
	LogFile               string `yaml:"log-file" description:"Log file location" default:"syslog"`

	PortalHost string `yaml:"portal-host" description:"Peering portal host (disabled if empty)" default:""`
	PortalKey  string `yaml:"portal-key" description:"Peering portal API key" default:""`
	Hostname   string `yaml:"hostname" description:"Router hostname (default system hostname)" default:""`

	ASN      int      `yaml:"asn" description:"Autonomous System Number" validate:"required" default:"0"`
	Prefixes []string `yaml:"prefixes" description:"List of prefixes to announce"`

	RouterID              string `yaml:"router-id" description:"Router ID (dotted quad notation)" validate:"required"`
	IRRServer             string `yaml:"irr-server" description:"Internet routing registry server" default:"rr.ntt.net"`
	RTRServer             string `yaml:"rtr-server" description:"RPKI-to-router server" default:"rtr.rpki.cloudflare.com:8282"`
	BGPQArgs              string `yaml:"bgpq-args" description:"Additional command line arguments to pass to bgpq4" default:""`
	KeepFiltered          bool   `yaml:"keep-filtered" description:"Should filtered routes be kept in memory?" default:"false"`
	KernelLearn           bool   `yaml:"kernel-learn" description:"Should routes from the kernel be learned into BIRD?" default:"false"`
	KernelExport          bool   `yaml:"kernel-export" description:"Export routes to kernel routing table" default:"true"`
	KernelRejectConnected bool   `yaml:"kernel-reject-connected" description:"Don't export connected routes (RTS_DEVICE) to kernel?'" default:"false"`
	MergePaths            bool   `yaml:"merge-paths" description:"Should best and equivalent non-best routes be imported to build ECMP routes?" default:"false"`
	Source4               string `yaml:"source4" description:"Source IPv4 address"`
	Source6               string `yaml:"source6" description:"Source IPv6 address"`
	DefaultRoute          bool   `yaml:"default-route" description:"Add a default route" default:"true"`
	AcceptDefault         bool   `yaml:"accept-default" description:"Should default routes be accepted? Setting to false adds 0.0.0.0/0 and ::/0 to the global bogon list." default:"false"`
	KernelTable           int    `yaml:"kernel-table" description:"Kernel table"`
	RPKIEnable            bool   `yaml:"rpki-enable" description:"Enable RPKI RTR session" default:"true"`

	NoAnnounce bool `yaml:"no-announce" description:"Don't announce any routes to any peer" default:"false"`
	NoAccept   bool `yaml:"no-accept" description:"Don't accept any routes from any peer" default:"false"`
	Stun       bool `yaml:"stun" description:"Don't accept or announce any routes from any peer (sets no-announce and no-accept)" default:"false"`

	Peers         map[string]*Peer         `yaml:"peers" description:"BGP peer configuration"`
	Templates     map[string]*Peer         `yaml:"templates" description:"BGP peer templates"`
	VRRPInstances map[string]*VRRPInstance `yaml:"vrrp" description:"List of VRRP instances"`
	BFDInstances  map[string]*BFDInstance  `yaml:"bfd" description:"BFD instances"`
	MRTInstances  map[string]*MRTInstance  `yaml:"mrt" description:"MRT instances"`
	Augments      *Augments                `yaml:"augments" description:"Custom configuration options"`
	Optimizer     *Optimizer               `yaml:"optimizer" description:"Route optimizer options"`
	Plugins       map[string]string        `yaml:"plugins" description:"Plugin-specific configuration"`

	RTRServerHost string   `yaml:"-" description:"-"`
	RTRServerPort int      `yaml:"-" description:"-"`
	Prefixes4     []string `yaml:"-" description:"-"`
	Prefixes6     []string `yaml:"-" description:"-"`
	QueryNVRS     bool     `yaml:"-" description:"-"`
	NVRSASNs      []uint32 `yaml:"-" description:"-"`
}

// Init initializes a Config with embedded structs prior to calling config.Default
func (c *Config) Init() {
	c.Peers = map[string]*Peer{}
	c.Templates = map[string]*Peer{}
	c.VRRPInstances = map[string]*VRRPInstance{}
	c.BFDInstances = map[string]*BFDInstance{}
	c.MRTInstances = map[string]*MRTInstance{}
	c.Augments = &Augments{}
	c.Optimizer = &Optimizer{}
	c.Plugins = map[string]string{}
}

// Default sets a Config's default values
func (c *Config) Default() error {
	// Set global config defaults
	return defaults.Set(c)
}
