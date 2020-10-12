# bcg

[![Go Report](https://goreportcard.com/badge/github.com/natesales/bcg?style=for-the-badge)](https://goreportcard.com/badge/github.com/natesales/bcg) 
[![License](https://img.shields.io/github/license/natesales/bcg?style=for-the-badge)](https://choosealicense.com/licenses/gpl-3.0/) 

The automatic BIRD configuration generator with bogon, IRR, RPKI, and max prefix filtering support.

#### Configuration

BCG can be configured in YAML, TOML, or JSON. All config file formats have the same configuration options but follow a different capitalization structure. YAML and JSON use all lowercase parameter names and TOML uses CapsCase with acronyms capitalized. For example, `router-id` in YAML and JSON is `Router-ID` in TOML.

An example to configure a peer with bogon, IRR, RPKI, and max prefix filtering.
```yaml
asn: 65530
router-id: 192.0.2.1
prefixes:
  - 192.0.2.0/24
  - 2001:db8::/48

peers:
  Cloudflare:
    asn: 13335
    import: cone
    export: cone
    automaxprefix: true
    autopfxfilter: true
    neighbors:
      - 203.0.113.39
      - 2001:db8:6939::39
```

Global Configuration Options

| Option    | Usage                                                                                                         |
| --------- | ------------------------------------------------------------------------------------------------------------- |
| asn       | ASN of this router                                                                                            |
| router-id | Router ID of this router                                                                                      |
| prefixes  | List of prefixes to originate                                                                                 |
| irrdb     | IRRDB to query prefix sets from (default is rr.ntt.net which includes generated route objects from RPKI ROAs) |
| peers     | Map of name to peer (see below)                                                                               |

Peer Configuration Options

| Option         | Usage                                                                                                     |
| -------------- | --------------------------------------------------------------------------------------------------------- |
| asn            | Neighbor ASN                                                                                              |
| as-set         | Neighbor IRR AS-SET                                                                                       |
| maxpfx4        | Maximum number of IPv4 prefixes to accept before enacting `pfxlimitaction`                                |
| maxpfx6        | Maximum number of IPv6 prefixes to accept before enacting `pfxlimitaction`                                |
| pfxlimitaction | Action to take when the max prefix limits are tripped (warn, disable, block, or restart) default: disable |
| pfxfilter4     | IPv4 prefix filter list in BIRD format                                                                    |
| pfxfilter6     | IPv6 prefix filter list in BIRD format                                                                    |
| import-policy  | Peer import policy (any, cone, none)                                                                      |
| export-policy  | Peer export policy (any, cone, none)                                                                      |
| automaxpfx     | Should max prefix limits be pulled from PeeringDB?                                                        |
| autopfxfilter  | Should prefix filters be pulled from IRR data?                                                            |
| disabled       | Should neighbor sessions be disabled?                                                                     |
| passive        | Should neighbor sessions listen passively for BGP TCP connections?                                        |
| multihop       | Should neighbor sessions allow multihop?                                                                  |
| neighbors      | List of neighbor IP addresses                                                                             |
