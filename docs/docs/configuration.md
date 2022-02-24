---
title: Configuration
sidebar_position: 3
---

# Configuration
## Config
### `peeringdb-query-timeout`

PeeringDB query timeout in seconds

| Type | Default | Validation |
|------|---------|------------|
| uint   | 10      |          |

### `irr-query-timeout`

IRR query timeout in seconds

| Type | Default | Validation |
|------|---------|------------|
| uint   | 30      |          |

### `bird-directory`

Directory to store BIRD configs

| Type | Default | Validation |
|------|---------|------------|
| string   | /etc/bird/      |          |

### `bird-binary`

Path to BIRD binary

| Type | Default | Validation |
|------|---------|------------|
| string   | /usr/sbin/bird      |          |

### `bird-socket`

UNIX control socket for BIRD

| Type | Default | Validation |
|------|---------|------------|
| string   | /run/bird/bird.ctl      |          |

### `cache-directory`

Directory to store runtime configuration cache

| Type | Default | Validation |
|------|---------|------------|
| string   | /var/run/pathvector/cache/      |          |

### `keepalived-config`

Configuration file for keepalived

| Type | Default | Validation |
|------|---------|------------|
| string   | /etc/keepalived.conf      |          |

### `web-ui-file`

File to write web UI to (disabled if empty)

| Type | Default | Validation |
|------|---------|------------|
| string   |       |          |

### `log-file`

Log file location

| Type | Default | Validation |
|------|---------|------------|
| string   | syslog      |          |

### `portal-host`

Peering portal host (disabled if empty)

| Type | Default | Validation |
|------|---------|------------|
| string   |       |          |

### `portal-key`

Peering portal API key

| Type | Default | Validation |
|------|---------|------------|
| string   |       |          |

### `hostname`

Router hostname (default system hostname)

| Type | Default | Validation |
|------|---------|------------|
| string   |       |          |

### `asn`

Autonomous System Number

| Type | Default | Validation |
|------|---------|------------|
| int   | 0      | required         |

### `prefixes`

List of prefixes to announce

| Type | Default | Validation |
|------|---------|------------|
| []string   |       |          |

### `router-id`

Router ID (dotted quad notation)

| Type | Default | Validation |
|------|---------|------------|
| string   |       | required         |

### `irr-server`

Internet routing registry server

| Type | Default | Validation |
|------|---------|------------|
| string   | rr.ntt.net      |          |

### `rtr-server`

RPKI-to-router server

| Type | Default | Validation |
|------|---------|------------|
| string   | rtr.rpki.cloudflare.com:8282      |          |

### `bgpq-args`

Additional command line arguments to pass to bgpq4

| Type | Default | Validation |
|------|---------|------------|
| string   |       |          |

### `keep-filtered`

Should filtered routes be kept in memory?

| Type | Default | Validation |
|------|---------|------------|
| bool   | false      |          |

### `kernel-learn`

Should routes from the kernel be learned into BIRD?

| Type | Default | Validation |
|------|---------|------------|
| bool   | false      |          |

### `kernel-export`

Export routes to kernel routing table

| Type | Default | Validation |
|------|---------|------------|
| bool   | true      |          |

### `kernel-reject-connected`

Don't export connected routes (RTS_DEVICE) to kernel?'

| Type | Default | Validation |
|------|---------|------------|
| bool   | false      |          |

### `merge-paths`

Should best and equivalent non-best routes be imported to build ECMP routes?

| Type | Default | Validation |
|------|---------|------------|
| bool   | false      |          |

### `source4`

Source IPv4 address

| Type | Default | Validation |
|------|---------|------------|
| string   |       |          |

### `source6`

Source IPv6 address

| Type | Default | Validation |
|------|---------|------------|
| string   |       |          |

### `default-route`

Add a default route

| Type | Default | Validation |
|------|---------|------------|
| bool   | true      |          |

### `accept-default`

Should default routes be accepted? Setting to false adds 0.0.0.0/0 and ::/0 to the global bogon list.

| Type | Default | Validation |
|------|---------|------------|
| bool   | false      |          |

### `kernel-table`

Kernel table

| Type | Default | Validation |
|------|---------|------------|
| int   |       |          |

### `rpki-enable`

Enable RPKI RTR session

| Type | Default | Validation |
|------|---------|------------|
| bool   | true      |          |

### `no-announce`

Don't announce any routes to any peer

| Type | Default | Validation |
|------|---------|------------|
| bool   | false      |          |

### `no-accept`

Don't accept any routes from any peer

| Type | Default | Validation |
|------|---------|------------|
| bool   | false      |          |

### `stun`

Don't accept or announce any routes from any peer (sets no-announce and no-accept)

| Type | Default | Validation |
|------|---------|------------|
| bool   | false      |          |

### `peers`

BGP peer configuration

| Type | Default | Validation |
|------|---------|------------|
| map[string]Peer   |       |          |

### `templates`

BGP peer templates

| Type | Default | Validation |
|------|---------|------------|
| map[string]Peer   |       |          |

### `vrrp`

List of VRRP instances

| Type | Default | Validation |
|------|---------|------------|
| map[string]VRRPInstance   |       |          |

### `bfd`

BFD instances

| Type | Default | Validation |
|------|---------|------------|
| map[string]BFDInstance   |       |          |

### `augments`

Custom configuration options

| Type | Default | Validation |
|------|---------|------------|
| [Augments](#augments-1)   |       |          |

### `optimizer`

Route optimizer options

| Type | Default | Validation |
|------|---------|------------|
| [Optimizer](#optimizer-1)   |       |          |

### `plugins`

Plugin-specific configuration

| Type | Default | Validation |
|------|---------|------------|
| map[string]string   |       |          |


## BFDInstance
### `neighbor`

Neighbor IP address

| Type | Default | Validation |
|------|---------|------------|
| string   |       |          |

### `interface`

Interface (pattern accepted)

| Type | Default | Validation |
|------|---------|------------|
| string   |       |          |

### `interval`

RX and TX interval

| Type | Default | Validation |
|------|---------|------------|
| uint   | 200      |          |

### `multiplier`

Number of missed packets for the state to be declared down

| Type | Default | Validation |
|------|---------|------------|
| uint   | 10      |          |


## Peer
### `template`

Configuration template

| Type | Default | Validation |
|------|---------|------------|
| string   |       |          |

### `description`

Peer description

| Type | Default | Validation |
|------|---------|------------|
| string   |       |          |

### `disabled`

Should the sessions be disabled?

| Type | Default | Validation |
|------|---------|------------|
| bool   | false      |          |

### `asn`

Local ASN

| Type | Default | Validation |
|------|---------|------------|
| int   | 0      | required         |

### `neighbors`

List of neighbor IPs

| Type | Default | Validation |
|------|---------|------------|
| []string   |       | required,ip         |

### `prepends`

Number of times to prepend local AS on export

| Type | Default | Validation |
|------|---------|------------|
| int   | 0      |          |

### `local-pref`

BGP local preference

| Type | Default | Validation |
|------|---------|------------|
| int   | 100      |          |

### `set-local-pref`

Should an explicit local pref be set?

| Type | Default | Validation |
|------|---------|------------|
| bool   | true      |          |

### `multihop`

Should BGP multihop be enabled? (255 max hops)

| Type | Default | Validation |
|------|---------|------------|
| bool   | false      |          |

### `listen4`

IPv4 BGP listen address

| Type | Default | Validation |
|------|---------|------------|
| string   |       |          |

### `listen6`

IPv6 BGP listen address

| Type | Default | Validation |
|------|---------|------------|
| string   |       |          |

### `local-asn`

Local ASN as defined in the global ASN field

| Type | Default | Validation |
|------|---------|------------|
| int   |       |          |

### `local-port`

Local TCP port

| Type | Default | Validation |
|------|---------|------------|
| int   | 179      |          |

### `neighbor-port`

Neighbor TCP port

| Type | Default | Validation |
|------|---------|------------|
| int   | 179      |          |

### `passive`

Should we listen passively?

| Type | Default | Validation |
|------|---------|------------|
| bool   | false      |          |

### `direct`

Specify that the neighbor is directly connected

| Type | Default | Validation |
|------|---------|------------|
| bool   | false      |          |

### `next-hop-self`

Should BGP next-hop-self be enabled?

| Type | Default | Validation |
|------|---------|------------|
| bool   | false      |          |

### `bfd`

Should BFD be enabled?

| Type | Default | Validation |
|------|---------|------------|
| bool   | false      |          |

### `password`

BGP MD5 password

| Type | Default | Validation |
|------|---------|------------|
| string   |       |          |

### `rs-client`

Should this peer be a route server client?

| Type | Default | Validation |
|------|---------|------------|
| bool   | false      |          |

### `rr-client`

Should this peer be a route reflector client?

| Type | Default | Validation |
|------|---------|------------|
| bool   | false      |          |

### `remove-private-asns`

Should private ASNs be removed from path before exporting?

| Type | Default | Validation |
|------|---------|------------|
| bool   | true      |          |

### `mp-unicast-46`

Should this peer be configured with multiprotocol IPv4 and IPv6 unicast?

| Type | Default | Validation |
|------|---------|------------|
| bool   | false      |          |

### `allow-local-as`

Should routes originated by the local ASN be accepted?

| Type | Default | Validation |
|------|---------|------------|
| bool   | false      |          |

### `add-path-tx`

Enable BGP additional paths on export?

| Type | Default | Validation |
|------|---------|------------|
| bool   | false      |          |

### `add-path-rx`

Enable BGP additional paths on import?

| Type | Default | Validation |
|------|---------|------------|
| bool   | false      |          |

### `import-next-hop`

Rewrite the BGP next hop before importing routes learned from this peer

| Type | Default | Validation |
|------|---------|------------|
| string   |       |          |

### `export-next-hop`

Rewrite the BGP next hop before announcing routes to this peer

| Type | Default | Validation |
|------|---------|------------|
| string   |       |          |

### `confederation`

BGP confederation (RFC 5065)

| Type | Default | Validation |
|------|---------|------------|
| int   |       |          |

### `confederation-member`

Should this peer be a member of the local confederation?

| Type | Default | Validation |
|------|---------|------------|
| bool   | false      |          |

### `ttl-security`

RFC 5082 Generalized TTL Security Mechanism

| Type | Default | Validation |
|------|---------|------------|
| bool   | false      |          |

### `interpret-communities`

Should well-known BGP communities be interpreted by their intended action?

| Type | Default | Validation |
|------|---------|------------|
| bool   | true      |          |

### `default-local-pref`

Default value for local preference

| Type | Default | Validation |
|------|---------|------------|
| int   |       |          |

### `import-communities`

List of communities to add to all imported routes

| Type | Default | Validation |
|------|---------|------------|
| []string   |       |          |

### `export-communities`

List of communities to add to all exported routes

| Type | Default | Validation |
|------|---------|------------|
| []string   |       |          |

### `announce-communities`

Announce all routes matching these communities to the peer

| Type | Default | Validation |
|------|---------|------------|
| []string   |       |          |

### `remove-communities`

List of communities to remove before from routes announced by this peer

| Type | Default | Validation |
|------|---------|------------|
| []string   |       |          |

### `remove-all-communities`

Remove all standard and large communities beginning with this value

| Type | Default | Validation |
|------|---------|------------|
| int   |       |          |

### `as-prefs`

Map of ASN to import local pref (not included in optimizer)

| Type | Default | Validation |
|------|---------|------------|
| map[uint32]uint32   |       |          |

### `as-set`

Peer's as-set for filtering

| Type | Default | Validation |
|------|---------|------------|
| string   |       |          |

### `import-limit4`

Maximum number of IPv4 prefixes to import

| Type | Default | Validation |
|------|---------|------------|
| int   | 1000000      |          |

### `import-limit6`

Maximum number of IPv6 prefixes to import

| Type | Default | Validation |
|------|---------|------------|
| int   | 200000      |          |

### `enforce-first-as`

Should we only accept routes who's first AS is equal to the configured peer address?

| Type | Default | Validation |
|------|---------|------------|
| bool   | true      |          |

### `enforce-peer-nexthop`

Should we only accept routes with a next hop equal to the configured neighbor address?

| Type | Default | Validation |
|------|---------|------------|
| bool   | true      |          |

### `force-peer-nexthop`

Rewrite nexthop to peer address

| Type | Default | Validation |
|------|---------|------------|
| bool   | false      |          |

### `max-prefix-action`

What action should be taken when the max prefix limit is tripped?

| Type | Default | Validation |
|------|---------|------------|
| string   | disable      |          |

### `allow-blackhole-community`

Should this peer be allowed to send routes with the blackhole community?

| Type | Default | Validation |
|------|---------|------------|
| bool   | false      |          |

### `blackhole-in`

Should imported routes be blackholed?

| Type | Default | Validation |
|------|---------|------------|
| bool   | false      |          |

### `blackhole-out`

Should exported routes be blackholed?

| Type | Default | Validation |
|------|---------|------------|
| bool   | false      |          |

### `filter-irr`

Should IRR filtering be applied?

| Type | Default | Validation |
|------|---------|------------|
| bool   | false      |          |

### `filter-rpki`

Should RPKI invalids be rejected?

| Type | Default | Validation |
|------|---------|------------|
| bool   | true      |          |

### `strict-rpki`

Should only RPKI valids be accepted?

| Type | Default | Validation |
|------|---------|------------|
| bool   | false      |          |

### `filter-max-prefix`

Should max prefix filtering be applied?

| Type | Default | Validation |
|------|---------|------------|
| bool   | true      |          |

### `filter-bogon-routes`

Should bogon prefixes be rejected?

| Type | Default | Validation |
|------|---------|------------|
| bool   | true      |          |

### `filter-bogon-asns`

Should paths containing a bogon ASN be rejected?

| Type | Default | Validation |
|------|---------|------------|
| bool   | true      |          |

### `filter-transit-asns`

Should paths containing transit-free ASNs be rejected? (Peerlock Lite)'

| Type | Default | Validation |
|------|---------|------------|
| bool   | false      |          |

### `filter-prefix-length`

Should too large/small prefixes (IPv4 8 > len > 24 and IPv6 12 > len > 48) be rejected?

| Type | Default | Validation |
|------|---------|------------|
| bool   | true      |          |

### `filter-never-via-route-servers`

Should routes containing an ASN reported in PeeringDB to never be reachable via route servers be filtered?

| Type | Default | Validation |
|------|---------|------------|
| bool   | false      |          |

### `auto-import-limits`

Get import limits automatically from PeeringDB?

| Type | Default | Validation |
|------|---------|------------|
| bool   | false      |          |

### `auto-as-set`

Get as-set automatically from PeeringDB? If no as-set exists in PeeringDB, a warning will be shown and the peer ASN used instead.

| Type | Default | Validation |
|------|---------|------------|
| bool   | false      |          |

### `honor-graceful-shutdown`

Should RFC8326 graceful shutdown be enabled?

| Type | Default | Validation |
|------|---------|------------|
| bool   | true      |          |

### `prefixes`

Prefixes to accept

| Type | Default | Validation |
|------|---------|------------|
| []string   |       |          |

### `announce-default`

Should a default route be exported to this peer?

| Type | Default | Validation |
|------|---------|------------|
| bool   | false      |          |

### `announce-originated`

Should locally originated routes be announced to this peer?

| Type | Default | Validation |
|------|---------|------------|
| bool   | true      |          |

### `announce-all`

Should all routes be exported to this peer?

| Type | Default | Validation |
|------|---------|------------|
| bool   | false      |          |

### `session-global`

Configuration to add to each session before any defined BGP protocols

| Type | Default | Validation |
|------|---------|------------|
| string   |       |          |

### `pre-import`

Configuration to add at the beginning of the import filter

| Type | Default | Validation |
|------|---------|------------|
| string   |       |          |

### `pre-export`

Configuration to add at the beginning of the export filter

| Type | Default | Validation |
|------|---------|------------|
| string   |       |          |

### `pre-import-final`

Configuration to add immediately before the final accept/reject on import

| Type | Default | Validation |
|------|---------|------------|
| string   |       |          |

### `pre-export-final`

Configuration to add immediately before the final accept/reject on export

| Type | Default | Validation |
|------|---------|------------|
| string   |       |          |

### `pre-import-file`

Configuration file to append to pre-import

| Type | Default | Validation |
|------|---------|------------|
| string   |       |          |

### `pre-export-file`

Configuration file to append to pre-export

| Type | Default | Validation |
|------|---------|------------|
| string   |       |          |

### `pre-import-final-file`

Configuration file to append to pre-import-final

| Type | Default | Validation |
|------|---------|------------|
| string   |       |          |

### `pre-export-final-file`

Configuration file to append to pre-export-final

| Type | Default | Validation |
|------|---------|------------|
| string   |       |          |

### `probe-sources`

Optimizer probe source addresses

| Type | Default | Validation |
|------|---------|------------|
| []string   |       |          |

### `optimize-inbound`

Should the optimizer modify inbound policy?

| Type | Default | Validation |
|------|---------|------------|
| bool   | false      |          |


## VRRPInstance
### `state`

VRRP instance state ('primary' or 'backup')

| Type | Default | Validation |
|------|---------|------------|
| string   |       | required         |

### `interface`

Interface to send VRRP packets on

| Type | Default | Validation |
|------|---------|------------|
| string   |       | required         |

### `vrid`

RFC3768 VRRP Virtual Router ID (1-255)

| Type | Default | Validation |
|------|---------|------------|
| uint   |       | required         |

### `priority`

RFC3768 VRRP Priority

| Type | Default | Validation |
|------|---------|------------|
| uint   |       | required         |

### `vips`

List of virtual IPs

| Type | Default | Validation |
|------|---------|------------|
| []string   |       | required,cidr         |


## Augments
### `accept4`

List of BIRD protocols to import into the IPv4 table

| Type | Default | Validation |
|------|---------|------------|
| []string   |       |          |

### `accept6`

List of BIRD protocols to import into the IPv6 table

| Type | Default | Validation |
|------|---------|------------|
| []string   |       |          |

### `reject4`

List of BIRD protocols to not import into the IPv4 table

| Type | Default | Validation |
|------|---------|------------|
| []string   |       |          |

### `reject6`

List of BIRD protocols to not import into the IPv6 table

| Type | Default | Validation |
|------|---------|------------|
| []string   |       |          |

### `statics`

List of static routes to include in BIRD

| Type | Default | Validation |
|------|---------|------------|
| map[string]string   |       |          |

### `srd-communities`

List of communities to filter routes exported to kernel (if list is not empty, all other prefixes will not be exported)

| Type | Default | Validation |
|------|---------|------------|
| []string   |       |          |


## Optimizer
### `targets`

List of probe targets

| Type | Default | Validation |
|------|---------|------------|
| []string   |       |          |

### `latency-threshold`

Maximum allowable latency in milliseconds

| Type | Default | Validation |
|------|---------|------------|
| uint   | 100      |          |

### `packet-loss-threshold`

Maximum allowable packet loss (percent)

| Type | Default | Validation |
|------|---------|------------|
| float64   | 0.5      |          |

### `modifier`

Amount to lower local pref by for depreferred peers

| Type | Default | Validation |
|------|---------|------------|
| uint   | 20      |          |

### `probe-count`

Number of pings to send in each run

| Type | Default | Validation |
|------|---------|------------|
| int   | 5      |          |

### `probe-timeout`

Number of seconds to wait before considering the ICMP message unanswered

| Type | Default | Validation |
|------|---------|------------|
| int   | 1      |          |

### `probe-interval`

Number of seconds wait between each optimizer run

| Type | Default | Validation |
|------|---------|------------|
| int   | 120      |          |

### `cache-size`

Number of probe results to store per peer

| Type | Default | Validation |
|------|---------|------------|
| int   | 15      |          |

### `probe-udp`

Use UDP probe (else ICMP)

| Type | Default | Validation |
|------|---------|------------|
| bool   | false      |          |

### `alert-script`

Script to call on optimizer event

| Type | Default | Validation |
|------|---------|------------|
| string   |       |          |

### `exit-on-cache-full`

Exit optimizer on cache full

| Type | Default | Validation |
|------|---------|------------|
| bool   | false      |          |


