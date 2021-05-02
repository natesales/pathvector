# bcg

[![Go Report](https://goreportcard.com/badge/github.com/natesales/bcg?style=for-the-badge)](https://goreportcard.com/report/github.com/natesales/bcg)
[![License](https://img.shields.io/github/license/natesales/bcg?style=for-the-badge)](https://choosealicense.com/licenses/gpl-3.0/)
[![Release](https://img.shields.io/github/v/release/natesales/bcg?style=for-the-badge)](https://github.com/natesales/bcg/releases)

The automatic router configuration generator for BGP with bogon, IRR, RPKI, and max prefix filtering support.

### Installation

bcg depends on [bird2](https://gitlab.nic.cz/labs/bird/), [GoRTR](https://github.com/cloudflare/gortr), [bgpq4](https://github.com/bgp/bgpq4), and optionally [keepalived](https://github.com/acassen/keepalived). Make sure the `bird` and `gortr` daemons are running and `bgpq4` is in path before running bcg. Releases can be downloaded from GitHub and from my public code repositories - see https://github.com/natesales/repo for more info. You can also build from source by cloning the repo and running `go build`. It's recommended to run bcg every 12 hours to update IRR prefix lists and PeeringDB prefix limits. Adding `0 */12 * * * /usr/local/bin/bcg` to your crontab will update the filters at 12 AM and PM. If you're using ZSH you might also be interested in my [birdc completion](https://github.com/natesales/zsh-bird-completions).

#### Configuration

BCG can be configured in YAML, TOML, or JSON. All config file formats have the same configuration options but follow a different capitalization structure. YAML and JSON use all lowercase parameter names and TOML uses CapsCase with acronyms capitalized. For example, `router-id` in YAML and JSON is `Router-ID` in TOML.

An example to configure a peer with bogon, IRR, RPKI, and max prefix filtering:

```yaml
asn: 65530
router-id: 192.0.2.1
prefixes:
  - 192.0.2.0/24
  - 2001:db8::/48

peers:
  Cloudflare:
    asn: 13335
    type: peer
    neighbors:
      - 203.0.113.39
      - 2001:db8:6939::39
```

`bcg` can take the following flags:

```
Usage:
  bcg [OPTIONS]

Application Options:
  -c, --config=            Configuration file in YAML, TOML, or JSON format (default: /etc/bcg/config.yml)
  -o, --output=            Directory to write output files to (default: /etc/bird/)
  -s, --socket=            BIRD control socket (default: /run/bird/bird.ctl)
  -k, --keepalived-config= Configuration file for keepalived (default: /etc/keepalived/keepalived.conf)
  -u, --ui-file=           File to store web UI (default: /tmp/bcg-ui.html)
  -n, --no-ui              Don't generate web UI
  -v, --verbose            Show verbose log messages
  -d, --dry-run            Don't modify BIRD config
      --no-configure       Don't configure BIRD
      --version            Show version and exit

Help Options:
  -h, --help               Show this help message
```

#### How does filtering work?

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

#### Local Preference

All sessions have a default BGP local pref of 100, except for routes tagged with community `65535, 0` ([RFC8326 Graceful Shutdown](https://tools.ietf.org/html/rfc8326)) where it is lowered to 0. BGP local pref can be adjusted on a per-peer basis with the `local-pref` option under the peer block.

#### Pre-import and Pre-export conditions

There are many features of BIRD that aren't part of bcg. If you find such a feature, please [open an issue](https://github.com/natesales/bcg/issues/new). If it's something that is highly specific to your use case, you can supply a BIRD config snippet in `pre-import` or `pre-export` in the peer block to include that snippet of BIRD code after the import prefilter or before the export filter respectively, or in `pre-import-final` or `pre-export-final` to include it immediately before the final `accept`/`reject` of the filter.

#### iBGP

Next hop self will be enabled on BGP sessions where the neighbor ASN and local ASN are the same (iBGP sessions). You can also enable next hop self manually with `next-hop-self`.

#### Manual configuration

If bcg doesn't have a feature you need (and you can't use pre-import/pre-export conditions) then you can supply pure BIRD config in `/etc/bird/manual*.conf` and bcg will load the config before loading the rest of the config.

#### Single-stack (IPv4 only / IPv6 only) support

bcg *should* have reasonable single-stack peering support but is not fully tested. Peers that don't have any route{,6} objects will not have sessions of that address family configured.

#### Peerlock Lite

Peers with type `peer` or `downstream` reject any route with a Tier 1 ASN in path ([Peerlock Lite](https://github.com/job/peerlock)).

#### VRRP

bcg can build [keepalived](https://github.com/acassen/keepalived) configs for VRRP. To enable VRRP, add a `vrrp` config key containing a list of VRRP instances to your bcg config file.

#### Communities

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

#### Global Configuration Options

| Option            | Usage                                                                                                           |
| ----------------- | --------------------------------------------------------------------------------------------------------------- |
| asn               | ASN of this router                                                                                              |
| router-id         | Router ID of this router                                                                                        |
| prefixes          | List of prefixes to originate                                                                                   |
| statics           | Map of static route to nexthop                                                                                  |
| irrdb             | IRRDB to query prefix sets from (default is rr.ntt.net which includes generated route objects from RPKI ROAs)   |
| rtr-server        | IP address or hostname of RPKI RTR server (default is 127.0.0.1)                                                |
| keep-filtered     | Should BIRD keep filtered routes                                                                                |
| peers             | Map of name to peer (see below)                                                                                 |
| merge-paths       | Enable merge paths on kernel export                                                                             |
| pref-src4         | Preferred source IPv4 to export to kernel                                                                       |
| pref-src6         | Preferred source IPv6 to export to kernel                                                                       |
| filter-default    | Should default routes be denied?                                                                                |
| enable-default    | Add static default routes                                                                                       |
| communities       | List of BGP communities to add on export (two comma-separated values per list element; example `0,0`)           |
| large-communities | List of BGP large communities to add on export (three comma-separated values per list element; example `0,0,0`) |
| kernel-accept4    | List of protocols to accept into the kernel table                                                               |
| kernel-accept6    | List of protocols to accept into the kernel table                                                               |
| kernel-reject4    | List of protocols to reject from the kernel table                                                               |
| kernel-reject6    | List of protocols to reject from the kernel table                                                               |

#### BGP Peer Configuration Options

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
| next-hop-self        | Enable "next hop self;" for specific peers                                                                        |

#### VRRP instance config options

| Option      | Usage                                                                          |
| ----------- | ------------------------------------------------------------------------------ |
| state       | VRRP state (`primary` or `backup`)                                             |
| interface   | Interface to run VRRP on                                                       |
| vrrid       | VRRP Router ID (must be the same for multiple routers in the same VRRP domain  |
| priority    | VRRP router selection priority                                                 |
| vips        | List of Virtual IPs                                                            |
