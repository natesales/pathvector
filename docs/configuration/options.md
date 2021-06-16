# Options
<!-- Code generated DO NOT EDIT -->
## config
| Option | Type | Default | Validation | Description |
|--------|------|---------|------------|-------------|
| asn | int | 0 | required | Autonomous System Number |
| prefixes | []string |  |  | List of prefixes to announce |
| communities | []string |  |  | List of RFC1997 BGP communities |
| large-communities | []string |  |  | List of RFC8092 large BGP communities |
| router-id | string |  | required | Router ID (dotted quad notation) |
| irr-server | string | rr.ntt.net |  | Internet routing registry server |
| rtr-server | string | rtr.rpki.cloudflare.com:8282 |  | RPKI-to-router server |
| keep-filtered | bool | false |  | Should filtered routes be kept in memory? |
| merge-paths | bool | false |  | Should best and equivalent non-best routes be imported to build ECMP routes? |
| source4 | string |  |  | Source IPv4 address |
| source6 | string |  |  | Source IPv6 address |
| accept-default | bool | false |  | Should default routes be added to the bogon list? |
| kernel-table | int |  |  | Kernel table |
| peers | map[string]*peer |  |  | BGP peer configuration |
| templates | map[string]*peer |  |  | BGP peer templates |
| vrrp | []vrrpInstance |  |  | List of VRRP instances |
| augments | augments |  |  | Custom configuration options |
| optimizer | optimizer |  |  | Route optimizer options |

<!-- Code generated DO NOT EDIT -->
## augments
| Option | Type | Default | Validation | Description |
|--------|------|---------|------------|-------------|
| accept4 | []string |  |  | List of BIRD protocols to import into the IPv4 table |
| accept6 | []string |  |  | List of BIRD protocols to import into the IPv6 table |
| reject4 | []string |  |  | List of BIRD protocols to not import into the IPv4 table |
| reject6 | []string |  |  | List of BIRD protocols to not import into the IPv6 table |
| statics | map[string]string |  |  | List of static routes to include in BIRD |

<!-- Code generated DO NOT EDIT -->
## optimizer
| Option | Type | Default | Validation | Description |
|--------|------|---------|------------|-------------|
| targets | []string |  |  | List of probe targets |
| probe-count | int |  |  | Number of pings to send in each run |
| probe-timeout | int |  |  | Number of seconds to wait before considering the ICMP message unanswered |
| probe-interval | int |  |  | Time to wait between each optimizer run |
| cache-size | int |  |  | Number of probe results to store per peer |

<!-- Code generated DO NOT EDIT -->
## *peer
| Option | Type | Default | Validation | Description |
|--------|------|---------|------------|-------------|
| template | *string |  |  | Configuration template |
| description | *string |  |  | Peer description |
| disabled | *bool | false |  | Should the sessions be disabled? |
| asn | *int | 0 | required | Local ASN |
| neighbors | *[]string |  | required,ip | List of neighbor IPs |
| prepends | *int | 0 |  | Number of times to prepend local AS on export |
| local-pref | *int | 100 |  | BGP local preference |
| multihop | *bool | false |  | Should BGP multihop be enabled? (255 max hops) |
| listen | *string |  |  | BGP listen address |
| local-port | *int | 179 |  | Local TCP port |
| neighbor-port | *int | 179 |  | Neighbor TCP port |
| passive | *bool | false |  | Should we listen passively? |
| next-hop-self | *bool | false |  | Should BGP next-hop-self be enabled? |
| bfd | *bool | false |  | Should BFD be enabled? |
| password | *string |  |  | BGP MD5 password |
| rs-client | *bool | false |  | Should this peer be a route server client? |
| rr-client | *bool | false |  | Should this peer be a route reflector client? |
| remove-private-asns | *bool | true |  | Should private ASNs be removed from path before exporting? |
| mp-unicast-46 | *bool | false |  | Should this peer be configured with multiprotocol IPv4 and IPv6 unicast? |
| allow-local-as | *bool | false |  | Should routes originated by the local ASN be accepted? |
| import-communities | *[]string |  |  | List of communities to add to all imported routes |
| export-communities | *[]string |  |  | List of communities to add to all exported routes |
| announce-communities | *[]string |  |  | Announce all routes matching these communities to the peer |
| as-set | *string |  |  | Peer's as-set for filtering |
| import-limit4 | *int | 1000000 |  | Maximum number of IPv4 prefixes to import |
| import-limit6 | *int | 200000 |  | Maximum number of IPv6 prefixes to import |
| enforce-first-as | *bool | true |  | Should we only accept routes who's first AS is equal to the configured peer address? |
| enforce-peer-nexthop | *bool | true |  | Should we only accept routes with a next hop equal to the configured neighbor address? |
| max-prefix-action | *string | disable |  | What action should be taken when the max prefix limit is tripped? |
| allow-blackhole-community | *bool | false |  | Should this peer be allowed to send routes with the blackhole community? |
| filter-irr | *bool | false |  | Should IRR filtering be applied? |
| filter-rpki | *bool | true |  | Should RPKI invalids be rejected? |
| filter-max-prefix | *bool | true |  | Should max prefix filtering be applied? |
| filter-bogon-routes | *bool | true |  | Should bogon prefixes be rejected? |
| filter-bogon-asns | *bool | true |  | Should paths containing a bogon ASN be rejected? |
| filter-transit-asns | *bool | false |  | Should paths containing transit-free ASNs be rejected? (Peerlock Lite)' |
| filter-prefix-length | *bool | true |  | Should too large/small prefixes (IPv4: len > 24 || len < 8; IPv6: len > 48 || len < 12) be rejected? |
| auto-import-limits | *bool | false |  | Get import limits automatically from PeeringDB? |
| auto-as-set | *bool | false |  | Get as-set automatically from PeeringDB? |
| honor-graceful-shutdown | *bool | true |  | Should RFC8326 graceful shutdown be enabled? |
| prefixes | *[]string |  |  | Prefixes to accept |
| announce-default | *bool | false |  | Should a default route be exported to this peer? |
| announce-originated | *bool | true |  | Should locally originated routes be announced to this peer? |
| session-global | *string |  |  | Configuration to add to each session before any defined BGP protocols |
| pre-import | *string |  |  | Configuration to add at the beginning of the import filter |
| pre-export | *string |  |  | Configuration to add at the beginning of the export filter |
| pre-import-final | *string |  |  | Configuration to add immediately before the final accept/reject on import |
| pre-export-final | *string |  |  | Configuration to add immediately before the final accept/reject on export |
| probe-sources | *[]string |  |  | Optimizer probe source addresses |
| optimize | *bool | false |  | Should the optimizer be enabled for this peer? |
| optimize-inbound | *bool | false |  | Should the optimizer modify inbound policy? |
| optimize-outbound | *bool | false |  | Should the optimizer modify outbound policy? |

<!-- Code generated DO NOT EDIT -->
## vrrpInstance
| Option | Type | Default | Validation | Description |
|--------|------|---------|------------|-------------|
| state | string |  | required | VRRP instance state ('primary' or 'backup') |
| interface | string |  | required | Interface to send VRRP packets on |
| vrid | uint |  | required | RFC3768 VRRP Virtual Router ID (1-255) |
| priority | uint |  | required | RFC3768 VRRP Priority |
| vips | []string |  | required,cidr | List of virtual IPs |

