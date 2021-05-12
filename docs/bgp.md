# BGP (Border Gateway Protocol)

## How does filtering work?

bcg applies a universal pre-filter to all BGP sessions before evaluating IRR or manual prefix lists which rejects importing any route that...

- Is locally originated as defined in the global `prefixes` list
- Has a bogon ASN in the AS path
- Has a total AS path length of more than 100
- Is RPKI invalid
- Is a bogon
- Is IPv4 and 8 <= length >= 24
- Is IPv6 and 12 <= length >= 48

Bogon lists can be found in [global.tmpl](https://github.com/natesales/bcg/blob/main/templates/global.tmpl)

All peers with a type of `peer` will apply further strict filtering by IRR using their as-set defined in PeeringDB. Max prefix limits are also enforced for every peer.

## Local Preference

All sessions have a default BGP local pref of 100, except for routes tagged with community `65535, 0` ([RFC8326 Graceful Shutdown](https://tools.ietf.org/html/rfc8326)) where it is lowered to 0. BGP local pref can be adjusted on a per-peer basis with the `local-pref` option under the peer block.

## Pre-import and Pre-export conditions

There are many features of BIRD that aren't part of bcg. If you find such a feature, please [open an issue](https://github.com/natesales/bcg/issues/new). If it's something that is highly specific to your use case, you can supply a BIRD config snippet in `pre-import` or `pre-export` in the peer block to include that snippet of BIRD code after the import prefilter or before the export filter respectively, or in `pre-import-final` or `pre-export-final` to include it immediately before the
final `accept`/`reject` of the filter.

## iBGP

Next hop self will be enabled on BGP sessions where the neighbor ASN and local ASN are the same (iBGP sessions). You can also enable next hop self manually with `next-hop-self`.

## Manual configuration

If bcg doesn't have a feature you need (and you can't use pre-import/pre-export conditions) then you can supply pure BIRD config in `/etc/bird/manual*.conf` and bcg will load the config before loading the rest of the config.

## Single-stack (IPv4 only / IPv6 only) support

bcg *should* have reasonable single-stack peering support but is not fully tested. Peers that don't have any route{,6} objects will not have sessions of that address family configured.

## Peerlock Lite

Peers with type `peer` or `downstream` reject any route with a Tier 1 ASN in path ([Peerlock Lite](https://github.com/job/peerlock)).

## Communities

bcg uses RFC 8092 BGP Large Communities

| Large     | Meaning                                            |
|-----------|--------------------------------------------------- |
| ASN,0,100 | Originated                                         |
| ASN,0,101 | Learned from upstream                              |
| ASN,0,102 | Learned from peer                                  |
| ASN,0,103 | Learned from downstream                            |
| ASN,1,200 | Don't export to upstreams                          |
| ASN,1,401 | Prepend once                                       |
| ASN,1,402 | Prepend twice                                      |
| ASN,1,403 | Prepend three times                                |
| ASN,1,666 | Blackhole (must be enabled with `allow-blackholes` |

## Neighbor Configuration

| Option               | Usage                                                                                                             |
| -------------------- | ----------------------------------------------------------------------------------------------------------------- |
| asn                  | Neighbor ASN                                                                                                      |
| type                 | Type of peer (upstream, peer, downstream, import-valid)                                                           |
| local-pref           | BGP LOCAL_PREF                                                                                                    |
| disabled             | Should neighbor sessions be disabled?                                                                             |
| passive              | Should neighbor sessions listen passively for BGP TCP connections?                                                |
| multihop             | Should neighbor sessions allow multihop?                                                                          |
| password             | BGP MD5 Password                                                                                                  |
| port                 | BGP Port (default 179)                                                                                            |
| listen               | BGP listen address                                                                                                |
| neighbors            | List of neighbor IP addresses                                                                                     |
| mp46-neighbors         | List of multi-protocol neighbor IP addresses                                                                    |
| as-set               | Manual override for peer's IRRDB as-set                                                                           |
| pre-import           | BIRD expression to evaluate after the prefilter and before the prefix filter                                      |
| pre-export           | BIRD expression to evaluate before the export filter                                                              |
| pre-import-final     | BIRD expression to evaluate right before the static return condition on import (accept or reject)                 |
| pre-export-final     | BIRD expression to evaluate right before the static return condition on export (accept or reject)                 |
| prepends             | Number of times to prepend local AS to on export                                                                  |
| import-limit4        | Maximum number of IPv4 prefixes to allow before disabling the session                                             |
| import-limit6        | Maximum number of IPv6 prefixes to allow before disabling the session                                             |
| skip-filter          | Disable the universal bogon filter (Dangerous!)                                                                   |
| rs-client            | Enable route server client                                                                                        |
| rr-client            | Enable route reflector client                                                                                     |
| bfd                  | Enable BFD                                                                                                        |
| session-global       | String to add to session global config                                                                            |
| enforce-first-as     | Reject routes that don't have the peer ASN as the first ASN in path                                               |
| enforce-peer-nexthop | Reject routes where the next hop doesn't match the neighbor address                                               |
| export-default       | Should a default route be sent over the session? (default false)                                                  |
| no-specifics         | Don't send specific routes (default false, make sure to enable export-default or else no routes will be exported) |
| allow-blackholes     | Accept community (ASN,1,666) to blackhole /32 and /128 prefixes                                                   |
| communities          | List of BGP communities to add on export (two comma-separated values per list element; example `0,0`)             |
| large-communities    | List of BGP large communities to add on export (three comma-separated values per list element; example `0,0,0`)   |
| description          | Description string (just for human reference)                                                                     |
| max-prefix-action    | Max prefix violation action                                                                                       |
| no-peeringdb         | Don't query PeeringDB for peering information                                                                     |
| next-hop-self        | Enable next hop self                                                                                              |
