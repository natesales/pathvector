<!-- Code generated DO NOT EDIT -->
# Configuration
## config
| Option | Type | Default | Validation | Description |
|--------|------|---------|------------|-------------|
| `asn` | `uint` |  | required | Autonomous System Number |
| `prefixes` | `[]string` |  |  | List of prefixes to announce |
| `communities` | `[]string` |  |  | List of RFC1997 BGP communities |
| `large-communities` | `[]string` |  |  | List of RFC8092 large BGP communities |
| `router-id` | `string` |  | required | Router ID (dotted quad notation) |
| `irr-server` | `string` | `rr.ntt.net` |  | Internet routing registry server |
| `rtr-server` | `string` | `rtr.rpki.cloudflare.com` |  | RPKI-to-router server |
| `rtr-port` | `int` | `8282` |  | RPKI-to-router port |
| `keep-filtered` | `bool` | `false` |  | Should filtered routes be kept in memory? |
| `merge-paths` | `bool` | `false` |  | Should best and equivalent non-best routes be imported for ECMP? |
| `source4` | `string` |  |  | Source IPv4 address |
| `source6` | `string` |  |  | Source IPv6 address |
| `accept-default` | `bool` | `false` |  | Should default routes be added to the bogon list? |
| `bird-directory` | `string` | `/etc/bird/` |  | Directory to store BIRD configs |
| `bird-socket` | `string` | `/run/bird/bird.ctl` |  | UNIX control socket for BIRD |
| `keepalived-config` | `string` | `/etc/keepalived.conf` |  | Configuration file for keepalived |
| `web-ui-file` | `string` | `/run/wireframe.html` |  | File to write web UI to |
| `peeringdb-query-timeout` | `uint` | `10` |  | PeeringDB query timeout in seconds |
| `irr-query-timeout` | `uint` | `30` |  | IRR query timeout in seconds |
| `peers` | `map[string]*peer` |  |  | BGP peer configuration |
| `interfaces` | `map[string]iface` |  |  | Network interface configuration |
| `vrrp` | `[]vrrpInstance` |  |  | List of VRRP instances |
| `augments` | `augments` |  |  | Custom configuration options |

<!-- Code generated DO NOT EDIT -->
# Configuration
## *peer
| Option | Type | Default | Validation | Description |
|--------|------|---------|------------|-------------|
| `description` | `string` |  |  | Peer description |
| `disabled` | `bool` |  |  | Should the sessions be disabled? |
| `asn` | `uint` |  | required | Local ASN |
| `neighbors` | `[]string` |  | required,ip | List of neighbor IPs |
| `prepends` | `uint` | `0` |  | Number of times to prepend local AS on export |
| `local-pref` | `uint` | `100` |  | BGP local preference |
| `multihop` | `bool` | `false` |  | Should BGP multihop be enabled? (255 max hops) |
| `listen` | `string` |  |  | BGP listen address |
| `local-port` | `uint16` | `179` |  | Local TCP port |
| `neighbor-port` | `uint16` | `179` |  | Neighbor TCP port |
| `passive` | `bool` | `false` |  | Should we listen passively? |
| `next-hop-self` | `bool` | `false` |  | Should BGP next-hop-self be enabled? |
| `bfd` | `bool` | `false` |  | Should BFD be enabled? |
| `password` | `string` |  |  | BGP MD5 password |
| `rs-client` | `bool` | `false` |  | Should this peer be a route server client? |
| `rr-client` | `bool` | `false` |  | Should this peer be a route reflector client? |
| `remove-private-as` | `bool` | `true` |  | Should private ASNs be removed from path before exporting? |
| `mp-unicast-46` | `bool` | `false` |  | Should this peer be configured with multiprotocol IPv4 and IPv6 unicast? |
| `import-communities` | `[]string` |  |  | List of communities to add to all imported routes |
| `export-communities` | `[]string` |  |  | List of communities to add to all exported routes |
| `announce-communities` | `[]string` |  |  | Announce all routes matching these communities to the peer |
| `as-set` | `string` |  |  | Peer's as-set for filtering |
| `import-limit4` | `uint` | `1000000` |  | Maximum number of IPv4 prefixes to import |
| `import-limit6` | `uint` | `100000` |  | Maximum number of IPv6 prefixes to import |
| `enforce-first-as` | `bool` | `true` |  | Should we only accept routes who's first AS is equal to the configured peer address? |
| `enforce-peer-nexthop` | `bool` | `true` |  | Should we only accept routes with a next hop equal to the configured neighbor address? |
| `max-prefix-action` | `string` | `disable` |  | What action should be taken when the max prefix limit is tripped? |
| `allow-blackhole-community` | `bool` | `false` |  | Should this peer be allowed to send routes with the blackhole community? |
| `filter-irr` | `bool` | `true` |  | Should IRR filtering be applied? |
| `filter-rpki` | `bool` | `true` |  | Should RPKI invalids be rejected? |
| `filter-max-prefix` | `bool` | `true` |  | Should max prefix filtering be applied? |
| `filter-bogons` | `bool` | `true` |  | Should bogon prefixes be rejected? |
| `filter-tier1-asns` | `bool` | `false` |  | Should paths containing 'Tier 1' ASNs be rejected (Peerlock Lite)?' |
| `auto-import-limits` | `bool` | `false` |  | Get import limits automatically from PeeringDB? |
| `auto-as-set` | `bool` | `false` |  | Get as-set automatically from PeeringDB? |
| `prefixes` | `[]string` |  |  | Prefixes to accept |
| `announce-default` | `bool` | `false` |  | Should a default route be exported to this peer? |
| `session-global` | `string` |  |  | Configuration to add to each session before any defined BGP protocols |
| `pre-import` | `string` |  |  | Configuration to add at the beginning of the import filter |
| `pre-export` | `string` |  |  | Configuration to add at the beginning of the export filter |
| `pre-import-final` | `string` |  |  | Configuration to add immediately before the final accept/reject on import |
| `pre-export-final` | `string` |  |  | Configuration to add immediately before the final accept/reject on export |

<!-- Code generated DO NOT EDIT -->
# Configuration
## iface
| Option | Type | Default | Validation | Description |
|--------|------|---------|------------|-------------|
| `mtu` | `uint` | `1500` |  | Interface MTU (Maximum Transmission Unit) |
| `xdprtr` | `bool` | `false` |  | Should XDPRTR be loaded on this interface? |
| `addresses` | `[]string` |  |  | List of addresses to add to this interface |
| `dummy` | `bool` | `false` |  | Should a new dummy interface be created with this configuration? |
| `down` | `bool` | `false` |  | Should the interface be set to a down state? |

<!-- Code generated DO NOT EDIT -->
# Configuration
## vrrpInstance
| Option | Type | Default | Validation | Description |
|--------|------|---------|------------|-------------|
| `state` | `string` |  | required | VRRP instance state ('primary' or 'backup') |
| `interface` | `string` |  | required | Interface to send VRRP packets on |
| `vrid` | `uint` |  | required | RFC3768 VRRP Virtual Router ID (1-255) |
| `priority` | `uint8` |  | required | RFC3768 VRRP Priority |
| `vips` | `[]string` |  | required,cidr | List of virtual IPs |

<!-- Code generated DO NOT EDIT -->
# Configuration
## augments
| Option | Type | Default | Validation | Description |
|--------|------|---------|------------|-------------|
| `accept4` | `[]string` |  |  | List of BIRD protocols to import into the IPv4 table |
| `accept6` | `[]string` |  |  | List of BIRD protocols to import into the IPv6 table |
| `reject4` | `[]string` |  |  | List of BIRD protocols to not import into the IPv4 table |
| `reject6` | `[]string` |  |  | List of BIRD protocols to not import into the IPv6 table |
| `statics` | `map[string]string` |  |  | List of static routes to include in BIRD |

