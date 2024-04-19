package config

import (
	"github.com/go-ping/ping"
)

var defaultTransitASNs = []uint32{
	174, // Cogent
	//  209, // Qwest (HE carries this on IXPs IPv6 (Jul 12 2018))
	701,  // UUNET
	702,  // UUNET
	1239, // Sprint
	1299, // Telia
	2914, // NTT Communications
	3257, // GTT Backbone
	3320, // Deutsche Telekom AG (DTAG)
	3356, // Level3 / Lumen
	3491, // PCCW
	3549, // Level3
	3561, // Savvis / CenturyLink
	4134, // Chinanet
	5511, // Orange opentransit
	6453, // Tata Communications
	6461, // Zayo Bandwidth
	6762, // Seabone / Telecom Italia
	6830, // Liberty Global
	7018, // AT&T
}

var defaultBogons4 = []string{
	// {{if not.AcceptDefault -}}0.0.0.0/0, # Default route{{end}}
	"0.0.0.0/8{8,32}",        // IANA - Local Identification
	"10.0.0.0/8{8,32}",       // RFC 1918 - Private Use
	"100.64.0.0/10{10,32}",   // RFC 6598 - Shared Address Space
	"127.0.0.0/8{8,32}",      // IANA - Loopback
	"169.254.0.0/16{16,32}",  // RFC 3927 - Link Local
	"172.16.0.0/12{12,32}",   // RFC 1918 - Private Use
	"192.0.2.0/24{24,32}",    // RFC 5737 - TEST-NET-1
	"192.88.99.0/24{24,32}",  // RFC 3068 - 6to4 prefix
	"192.168.0.0/16{16,32}",  // RFC 1918 - Private Use
	"198.18.0.0/15{15,32}",   // RFC 2544 - Network Interconnect Device Benchmark Testing
	"198.51.100.0/24{24,32}", // RFC 5737 - TEST-NET-2
	"203.0.113.0/24{24,32}",  // RFC 5737 - TEST-NET-3
	"224.0.0.0/3{3,32}",      // RFC 5771 - Multicast (formerly Class D)
}

var defaultBogons6 = []string{
	// {{ if not .AcceptDefault -}}::/0,                     # Default route{{ end }}
	"::/8{8,128}",              // loopback, unspecified, v4-mapped
	"64:ff9b::/96{96,128}",     // RFC 6052 - IPv4-IPv6 Translation
	"100::/8{8,128}",           // RFC 6666 - reserved for Discard-Only Address Block
	"200::/7{7,128}",           // RFC 4048 - Reserved by IETF
	"400::/6{6,128}",           // RFC 4291 - Reserved by IETF
	"800::/5{5,128}",           // RFC 4291 - Reserved by IETF
	"1000::/4{4,128}",          // RFC 4291 - Reserved by IETF
	"2001::/33{33,128}",        // RFC 4380 - Teredo prefix
	"2001:0:8000::/33{33,128}", // RFC 4380 - Teredo prefix
	"2001:2::/48{48,128}",      // RFC 5180 - Benchmarking
	"2001:3::/32{32,128}",      // RFC 7450 - Automatic Multicast Tunneling
	"2001:10::/28{28,128}",     // RFC 4843 - Deprecated ORCHID
	"2001:20::/28{28,128}",     // RFC 7343 - ORCHIDv2
	"2001:db8::/32{32,128}",    // RFC 3849 - NON-ROUTABLE range to be used for documentation purpose
	"2002::/16{16,128}",        // RFC 3068 - 6to4 prefix
	"3ffe::/16{16,128}",        // RFC 5156 - used for the 6bone but was returned
	"4000::/3{3,128}",          // RFC 4291 - Reserved by IETF
	"5f00::/8{8,128}",          // RFC 5156 - used for the 6bone but was returned
	"6000::/3{3,128}",          // RFC 4291 - Reserved by IETF
	"8000::/3{3,128}",          // RFC 4291 - Reserved by IETF
	"a000::/3{3,128}",          // RFC 4291 - Reserved by IETF
	"c000::/3{3,128}",          // RFC 4291 - Reserved by IETF
	"e000::/4{4,128}",          // RFC 4291 - Reserved by IETF
	"f000::/5{5,128}",          // RFC 4291 - Reserved by IETF
	"f800::/6{6,128}",          // RFC 4291 - Reserved by IETF
	"fc00::/7{7,128}",          // RFC 4193 - Unique Local Unicast
	"fe80::/10{10,128}",        // RFC 4291 - Link Local Unicast
	"fec0::/10{10,128}",        // RFC 4291 - Reserved by IETF
	"ff00::/8{8,128}",          // RFC 4291 - Multicast
}

var defaultBogonASNs = []string{
	"0",                      // Reserved. RFC7607
	"23456",                  // AS_TRANS. RFC6793
	"64496..64511",           // Reserved for use in documentation and sample code. RFC5398
	"64512..65534",           // Reserved for Private Use. RFC6996
	"65535",                  // Reserved. RFC7300
	"65536..65551",           // Reserved for use in documentation and sample code. RFC5398
	"65552..131071",          // Reserved.
	"4200000000..4294967294", // Reserved for Private Use. [RFC6996]
	"4294967295",             // Reserved. RFC7300
}

// Peer stores a single peer config
type Peer struct {
	Template *string `yaml:"template" description:"Configuration template" default:"-"`

	Description *string   `yaml:"description" description:"Peer description" default:"-"`
	Tags        *[]string `yaml:"tags" description:"Peer tags" default:"-"`
	Disabled    *bool     `yaml:"disabled" description:"Should the sessions be disabled?" default:"false"`

	Import *bool `yaml:"import" description:"Import routes from this peer" default:"true"`
	Export *bool `yaml:"export" description:"Export routes to this peer" default:"true"`

	// BGP Attributes
	ASN                    *int      `yaml:"asn" description:"Local ASN" validate:"required" default:"0"`
	NeighborIPs            *[]string `yaml:"neighbors" description:"List of neighbor IPs" validate:"required,ip" default:"-"`
	Prepends               *int      `yaml:"prepends" description:"Number of times to prepend local AS on export" default:"0"`
	PrependPath            *[]uint32 `yaml:"prepend-path" description:"List of ASNs to prepend" default:"-"`
	ClearPath              *bool     `yaml:"clear-path" description:"Remove all ASNs from path (before prepends and prepend-path)" default:"false"`
	LocalPref              *int      `yaml:"local-pref" description:"BGP local preference" default:"100"`
	LocalPref4             *int      `yaml:"local-pref4" description:"IPv4 BGP local preference (overrides local-pref, not included in optimizer)" default:"-"`
	LocalPref6             *int      `yaml:"local-pref6" description:"IPv6 BGP local preference (overrides local-pref, not included in optimizer)" default:"-"`
	SetLocalPref           *bool     `yaml:"set-local-pref" description:"Should an explicit local pref be set?" default:"true"`
	Multihop               *bool     `yaml:"multihop" description:"Should BGP multihop be enabled? (255 max hops)" default:"false"`
	Listen4                *string   `yaml:"listen4" description:"IPv4 BGP listen address" default:"-"`
	Listen6                *string   `yaml:"listen6" description:"IPv6 BGP listen address" default:"-"`
	LocalASN               *int      `yaml:"local-asn" description:"Local ASN as defined in the global ASN field" default:"-"`
	LocalPort              *int      `yaml:"local-port" description:"Local TCP port" default:"179"`
	NeighborPort           *int      `yaml:"neighbor-port" description:"Neighbor TCP port" default:"179"`
	Passive                *bool     `yaml:"passive" description:"Should we listen passively?" default:"false"`
	Direct                 *bool     `yaml:"direct" description:"Specify that the neighbor is directly connected" default:"false"`
	NextHopSelf            *bool     `yaml:"next-hop-self" description:"Should BGP next-hop-self be enabled?" default:"false"`
	NextHopSelfEBGP        *bool     `yaml:"next-hop-self-ebgp" description:"Should BGP next-hop-self for eBGP be enabled?" default:"false"`
	NextHopSelfIBGP        *bool     `yaml:"next-hop-self-ibgp" description:"Should BGP next-hop-self for iBGP be enabled?" default:"false"`
	BFD                    *bool     `yaml:"bfd" description:"Should BFD be enabled?" default:"false"`
	Password               *string   `yaml:"password" description:"BGP MD5 password" default:"-"`
	RSClient               *bool     `yaml:"rs-client" description:"Should this peer be a route server client?" default:"false"`
	RRClient               *bool     `yaml:"rr-client" description:"Should this peer be a route reflector client?" default:"false"`
	RemovePrivateASNs      *bool     `yaml:"remove-private-asns" description:"Should private ASNs be removed from path before exporting?" default:"true"`
	MPUnicast46            *bool     `yaml:"mp-unicast-46" description:"Should this peer be configured with multiprotocol IPv4 and IPv6 unicast?" default:"false"`
	AllowLocalAS           *bool     `yaml:"allow-local-as" description:"Should routes originated by the local ASN be accepted?" default:"false"`
	AddPathTx              *bool     `yaml:"add-path-tx" description:"Enable BGP additional paths on export?" default:"false"`
	AddPathRx              *bool     `yaml:"add-path-rx" description:"Enable BGP additional paths on import?" default:"false"`
	ImportNextHop          *string   `yaml:"import-next-hop" description:"Rewrite the BGP next hop before importing routes learned from this peer" default:"-"`
	ExportNextHop          *string   `yaml:"export-next-hop" description:"Rewrite the BGP next hop before announcing routes to this peer" default:"-"`
	Confederation          *int      `yaml:"confederation" description:"BGP confederation (RFC 5065)" default:"-"`
	ConfederationMember    *bool     `yaml:"confederation-member" description:"Should this peer be a member of the local confederation?" default:"false"`
	TTLSecurity            *bool     `yaml:"ttl-security" description:"RFC 5082 Generalized TTL Security Mechanism" default:"false"`
	InterpretCommunities   *bool     `yaml:"interpret-communities" description:"Should well-known BGP communities be interpreted by their intended action?" default:"true"`
	DefaultLocalPref       *int      `yaml:"default-local-pref" description:"Default value for local preference" default:"-"`
	AdvertiseHostname      *bool     `yaml:"advertise-hostname" description:"Advertise hostname capability" default:"false"`
	DisableAfterError      *bool     `yaml:"disable-after-error" description:"Disable peer after error" default:"false"`
	PreferOlderRoutes      *bool     `yaml:"prefer-older-routes" description:"Prefer older routes instead of comparing router IDs (RFC 5004)" default:"false"`
	IRRAcceptChildPrefixes *bool     `yaml:"irr-accept-child-prefixes" description:"Accept prefixes up to /24 and /48 from covering parent IRR objects" default:"false"`

	ImportCommunities    *[]string `yaml:"add-on-import" description:"List of communities to add to all imported routes" default:"-"`
	ExportCommunities    *[]string `yaml:"add-on-export" description:"List of communities to add to all exported routes" default:"-"`
	AnnounceCommunities  *[]string `yaml:"announce" description:"Announce all routes matching these communities to the peer" default:"-"`
	RemoveCommunities    *[]string `yaml:"remove-communities" description:"List of communities to remove before from routes announced by this peer" default:"-"`
	RemoveAllCommunities *int      `yaml:"remove-all-communities" description:"Remove all standard and large communities beginning with this value" default:"-"`

	ASPrefs *map[uint32]uint32 `yaml:"as-prefs" description:"Map of ASN to import local pref (not included in optimizer)" default:"-"`

	CommunityPrefs         *map[string]uint32 `yaml:"community-prefs" description:"Map of community to import local pref (not included in optimizer)" default:"-"`
	StandardCommunityPrefs *map[string]uint32 `yaml:"-" description:"-" default:"-"`
	LargeCommunityPrefs    *map[string]uint32 `yaml:"-" description:"-" default:"-"`

	// Filtering
	ASSet *string `yaml:"as-set" description:"Peer's as-set for filtering" default:"-"`

	ImportLimit4          *int    `yaml:"import-limit4" description:"Maximum number of IPv4 prefixes to import after filtering" default:"1000000"`
	ImportLimit6          *int    `yaml:"import-limit6" description:"Maximum number of IPv6 prefixes to import after filtering" default:"300000"`
	ImportLimitTripAction *string `yaml:"import-limit-violation" description:"What action should be taken when the import limit is tripped?" default:"disable"`

	ReceiveLimit4          *int    `yaml:"receive-limit4" description:"Maximum number of IPv4 prefixes to accept (including filtered routes, requires keep-filtered)" default:"-"`
	ReceiveLimit6          *int    `yaml:"receive-limit6" description:"Maximum number of IPv6 prefixes to accept (including filtered routes, requires keep-filtered)" default:"-"`
	ReceiveLimitTripAction *string `yaml:"receive-limit-violation" description:"What action should be taken when the receive limit is tripped?" default:"disable"`

	ExportLimit4          *int    `yaml:"export-limit4" description:"Maximum number of IPv4 prefixes to export" default:"-"`
	ExportLimit6          *int    `yaml:"export-limit6" description:"Maximum number of IPv6 prefixes to export" default:"-"`
	ExportLimitTripAction *string `yaml:"export-limit-violation" description:"What action should be taken when the export limit is tripped?" default:"disable"`

	EnforceFirstAS          *bool `yaml:"enforce-first-as" description:"Should we only accept routes who's first AS is equal to the configured peer address?" default:"true"`
	EnforcePeerNexthop      *bool `yaml:"enforce-peer-nexthop" description:"Should we only accept routes with a next hop equal to the configured neighbor address?" default:"true"`
	ForcePeerNexthop        *bool `yaml:"force-peer-nexthop" description:"Rewrite nexthop to peer address" default:"false"`
	AllowBlackholeCommunity *bool `yaml:"allow-blackhole-community" description:"Should this peer be allowed to send routes with the blackhole community?" default:"false"`
	BlackholeIn             *bool `yaml:"blackhole-in" description:"Should imported routes be blackholed?" default:"false"`
	BlackholeOut            *bool `yaml:"blackhole-out" description:"Should exported routes be blackholed?" default:"false"`

	// Filtering
	FilterIRR                  *bool `yaml:"filter-irr" description:"Should IRR filtering be applied?" default:"false"`
	FilterRPKI                 *bool `yaml:"filter-rpki" description:"Should RPKI invalids be rejected?" default:"true"`
	StrictRPKI                 *bool `yaml:"strict-rpki" description:"Should only RPKI valids be accepted?" default:"false"`
	FilterMaxPrefix            *bool `yaml:"filter-max-prefix" description:"Should max prefix filtering be applied?" default:"true"`
	FilterBogonRoutes          *bool `yaml:"filter-bogon-routes" description:"Should bogon prefixes be rejected?" default:"true"`
	FilterBogonASNs            *bool `yaml:"filter-bogon-asns" description:"Should paths containing a bogon ASN be rejected?" default:"true"`
	FilterTransitASNs          *bool `yaml:"filter-transit-asns" description:"Should paths containing transit-free ASNs be rejected? (Peerlock Lite)'" default:"false"`
	FilterPrefixLength         *bool `yaml:"filter-prefix-length" description:"Should too large/small prefixes (IPv4 8 > len > 24 and IPv6 12 > len > 48) be rejected?" default:"true"`
	FilterNeverViaRouteServers *bool `yaml:"filter-never-via-route-servers" description:"Should routes containing an ASN reported in PeeringDB to never be reachable via route servers be filtered?" default:"false"`
	FilterASSet                *bool `yaml:"filter-as-set" description:"Reject routes that aren't originated by an ASN within this peer's AS set" default:"false"`
	FilterASPA                 *bool `yaml:"filter-aspa" description:"Reject routes that aren't originated by an ASN within the authorized-providers map" default:"false"`
	FilterBlocklist            *bool `yaml:"filter-blocklist" description:"Reject ASNs, prefixes, and IPs in the global blocklist" default:"true"`

	TransitLock *[]string `yaml:"transit-lock" description:"Reject routes that aren't transited by an AS in this list" default:"-"`

	DontAnnounce *[]string `yaml:"dont-announce" description:"Don't announce these prefixes to the peer" default:"-"`
	OnlyAnnounce *[]string `yaml:"only-announce" description:"Only announce these prefixes to the peer" default:"-"`

	PrefixCommunities         *map[string][]string `yaml:"prefix-communities" description:"Map of prefix to community list to add to the prefix" default:"-"`
	PrefixStandardCommunities *map[string][]string `yaml:"-" description:"-" default:"-"`
	PrefixLargeCommunities    *map[string][]string `yaml:"-" description:"-" default:"-"`

	AutoImportLimits *bool `yaml:"auto-import-limits" description:"Get import limits automatically from PeeringDB?" default:"false"`
	AutoASSet        *bool `yaml:"auto-as-set" description:"Get as-set automatically from PeeringDB? If no as-set exists in PeeringDB, a warning will be shown and the peer ASN used instead." default:"false"`
	AutoASSetMembers *bool `yaml:"auto-as-set-members" description:"Get AS set members automatically from the peer's IRR as-set? (independent from auto-as-set)" default:"false"`

	HonorGracefulShutdown *bool `yaml:"honor-graceful-shutdown" description:"Should RFC8326 graceful shutdown be enabled?" default:"true"`

	Prefixes     *[]string `yaml:"prefixes" description:"Prefixes to accept" default:"-"`
	ASSetMembers *[]uint32 `yaml:"as-set-members" description:"AS set members (For filter-as-set)" default:"-"`

	Role         *string `yaml:"role" description:"RFC 9234 Local BGP role" default:"-"`
	RequireRoles *bool   `yaml:"require-roles" description:"Require RFC 9234 BGP roles" default:"false"`

	// Export options
	AnnounceDefault    *bool `yaml:"announce-default" description:"Should a default route be exported to this peer?" default:"false"`
	AnnounceOriginated *bool `yaml:"announce-originated" description:"Should locally originated routes be announced to this peer?" default:"true"`
	AnnounceAll        *bool `yaml:"announce-all" description:"Should all routes be exported to this peer?" default:"false"`

	// Custom daemon configuration
	SessionGlobal *string `yaml:"session-global" description:"Configuration to add to each session before any defined BGP protocols" default:"-"`

	PreImportFilter  *string `yaml:"pre-import-filter" description:"Configuration to add before the filtering section of the import policy" default:"-"`
	PostImportFilter *string `yaml:"post-import-filter" description:"Configuration to add after the filtering section of the import filter" default:"-"`
	PreImportAccept  *string `yaml:"pre-import-accept" description:"Configuration to add immediately before the final accept term import" default:"-"`
	PreExport        *string `yaml:"pre-export" description:"Configuration to add before the export policy" default:"-"`
	PreExportFinal   *string `yaml:"pre-export-final" description:"Configuration to add after the export policy before the final accept/reject term" default:"-"`

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

// Kernel stores options that relate to the OS kernel
type Kernel struct {
	Accept4         []string          `yaml:"accept4" description:"List of BIRD protocols to import into the IPv4 table"`
	Accept6         []string          `yaml:"accept6" description:"List of BIRD protocols to import into the IPv6 table"`
	Reject4         []string          `yaml:"reject4" description:"List of BIRD protocols to not import into the IPv4 table"`
	Reject6         []string          `yaml:"reject6" description:"List of BIRD protocols to not import into the IPv6 table"`
	Statics         map[string]string `yaml:"statics" description:"List of static routes to include in BIRD"`
	SRDCommunities  []string          `yaml:"srd-communities" description:"List of communities to filter routes exported to kernel (if list is not empty, all other prefixes will not be exported)"`
	Learn           bool              `yaml:"learn" description:"Should routes from the kernel be learned into BIRD?" default:"false"`
	Export          bool              `yaml:"export" description:"Export routes to kernel routing table" default:"true"`
	RejectConnected bool              `yaml:"reject-connected" description:"Don't export connected routes (RTS_DEVICE) to kernel?'" default:"false"`
	Table           int               `yaml:"table" description:"Kernel table"`
	ScanTime        int               `yaml:"scan-time" description:"Time in seconds between scans of the kernel routing table" default:"10"`

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
	GlobalConfig          string `yaml:"global-config" description:"Global BIRD configuration" default:""`
	PeeringDBURL          string `yaml:"peeringdb-url" description:"PeeringDB API URL, can be set to a local PeeringDB cache server" default:"https://peeringdb.com/api/"`

	Blocklist      []string `yaml:"blocklist" description:"List of ASNs, prefixes, and IP addresses to block" default:""`
	BlocklistURLs  []string `yaml:"blocklist-urls" description:"List of URLs to fetch blocklists from" default:""`
	BlocklistFiles []string `yaml:"blocklist-files" description:"List of files to fetch blocklists from" default:""`

	BlocklistASNs     []uint32 `yaml:"-" description:"-"`
	BlocklistPrefixes []string `yaml:"-" description:"-"`

	OriginCommunities []string `yaml:"origin-communities" description:"List of communities to accept as locally originated routes" default:""`
	LocalCommunities  []string `yaml:"local-communities" description:"List of communities to add to locally originated prefixes" default:""`
	ImportCommunities []string `yaml:"add-on-import" description:"List of communities to add to all imported routes" default:"-"`
	ExportCommunities []string `yaml:"add-on-export" description:"List of communities to add to all exported routes" default:"-"`

	Hostname string `yaml:"hostname" description:"Router hostname (default system hostname)" default:""`

	ASN      int      `yaml:"asn" description:"Autonomous System Number" validate:"required" default:"0"`
	Prefixes []string `yaml:"prefixes" description:"List of prefixes to announce"`

	RouterID      string `yaml:"router-id" description:"Router ID (dotted quad notation)" validate:"required"`
	IRRServer     string `yaml:"irr-server" description:"Internet routing registry server" default:"rr.ntt.net"`
	RTRServer     string `yaml:"rtr-server" description:"RPKI-to-router server" default:"rtr.rpki.cloudflare.com:8282"`
	BGPQArgs      string `yaml:"bgpq-args" description:"Additional command line arguments to pass to bgpq4" default:""`
	KeepFiltered  bool   `yaml:"keep-filtered" description:"Should filtered routes be kept in memory?" default:"false"`
	MergePaths    bool   `yaml:"merge-paths" description:"Should best and equivalent non-best routes be imported to build ECMP routes?" default:"false"`
	Source4       string `yaml:"source4" description:"Source IPv4 address"`
	Source6       string `yaml:"source6" description:"Source IPv6 address"`
	DefaultRoute  bool   `yaml:"default-route" description:"Add a default route" default:"true"`
	AcceptDefault bool   `yaml:"accept-default" description:"Should default routes be accepted? Setting to false adds 0.0.0.0/0 and ::/0 to the global bogon list." default:"false"`
	RPKIEnable    bool   `yaml:"rpki-enable" description:"Enable RPKI protocol" default:"true"`

	TransitASNs        []uint32 `yaml:"transit-asns" description:"List of ASNs to consider transit providers for filter-transit-asns (default list in config)" default:""`
	Bogons4            []string `yaml:"bogons4" description:"List of IPv4 bogons (default list in config)" default:""`
	Bogons6            []string `yaml:"bogons6" description:"List of IPv6 bogons (default list in config)" default:""`
	BogonASNs          []string `yaml:"bogon-asns" description:"List of ASNs to consider bogons (default list in config)" default:""`
	BlackholeBogonASNs bool     `yaml:"blackhole-bogon-asns" description:"Should routes containing bogon ASNs be blackholed?" default:"false"`

	NoAnnounce bool `yaml:"no-announce" description:"Don't announce any routes to any peer" default:"false"`
	NoAccept   bool `yaml:"no-accept" description:"Don't accept any routes from any peer" default:"false"`
	Stun       bool `yaml:"stun" description:"Don't accept or announce any routes from any peer (sets no-announce and no-accept)" default:"false"`

	AuthorizedProviders map[uint32][]uint32 `yaml:"authorized-providers" description:"Map of origin ASN to authorized provider ASN list" default:"-"`

	Peers         map[string]*Peer         `yaml:"peers" description:"BGP peer configuration"`
	Templates     map[string]*Peer         `yaml:"templates" description:"BGP peer templates"`
	VRRPInstances map[string]*VRRPInstance `yaml:"vrrp" description:"List of VRRP instances"`
	BFDInstances  map[string]*BFDInstance  `yaml:"bfd" description:"BFD instances"`
	MRTInstances  map[string]*MRTInstance  `yaml:"mrt" description:"MRT instances"`
	Kernel        *Kernel                  `yaml:"kernel" description:"Kernel routing configuration options"`
	Optimizer     *Optimizer               `yaml:"optimizer" description:"Route optimizer options"`
	Plugins       map[string]string        `yaml:"plugins" description:"Plugin-specific configuration"`

	RTRServerHost             string   `yaml:"-" description:"-"`
	RTRServerPort             int      `yaml:"-" description:"-"`
	Prefixes4                 []string `yaml:"-" description:"-"`
	Prefixes6                 []string `yaml:"-" description:"-"`
	QueryNVRS                 bool     `yaml:"-" description:"-"`
	NVRSASNs                  []uint32 `yaml:"-" description:"-"`
	OriginStandardCommunities []string `yaml:"-" description:"-"`
	OriginLargeCommunities    []string `yaml:"-" description:"-"`
	LocalStandardCommunities  []string `yaml:"-" description:"-"`
	LocalLargeCommunities     []string `yaml:"-" description:"-"`
	ImportStandardCommunities []string `yaml:"-" description:"-" default:"-"`
	ImportLargeCommunities    []string `yaml:"-" description:"-" default:"-"`
	ExportStandardCommunities []string `yaml:"-" description:"-" default:"-"`
	ExportLargeCommunities    []string `yaml:"-" description:"-" default:"-"`
}

// Init initializes a Config with embedded structs prior to calling config.Default
func (c *Config) Init() {
	c.Peers = map[string]*Peer{}
	c.Templates = map[string]*Peer{}
	c.VRRPInstances = map[string]*VRRPInstance{}
	c.BFDInstances = map[string]*BFDInstance{}
	c.MRTInstances = map[string]*MRTInstance{}
	c.Kernel = &Kernel{}
	c.Optimizer = &Optimizer{}
	c.Plugins = map[string]string{}

	if c.TransitASNs == nil {
		c.TransitASNs = defaultTransitASNs
	}
	if c.Bogons4 == nil {
		c.Bogons4 = defaultBogons4
	}
	if c.Bogons6 == nil {
		c.Bogons6 = defaultBogons6
	}
	if c.BogonASNs == nil {
		c.BogonASNs = defaultBogonASNs
	}
}
