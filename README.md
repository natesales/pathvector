# bcg

[![Go Report](https://goreportcard.com/badge/github.com/natesales/bcg?style=for-the-badge)](https://goreportcard.com/report/github.com/natesales/bcg) 
[![License](https://img.shields.io/github/license/natesales/bcg?style=for-the-badge)](https://choosealicense.com/licenses/gpl-3.0/) 
[![Release](https://img.shields.io/github/v/release/natesales/bcg?style=for-the-badge)](https://github.com/natesales/bcg/releases) 

The automatic BIRD configuration generator with bogon, IRR, RPKI, and max prefix filtering support.

### Installation
bcg depends on [bird2](https://gitlab.nic.cz/labs/bird/), [GoRTR](https://github.com/cloudflare/gortr), and [bgpq4](https://github.com/bgp/bgpq4). Make sure the `bird` and `gortr` daemons are running and `bgpq4` is in path before running bcg. bcg is available for amd64 as a prebuilt deb package and binary for each release. You can also build from source by cloning the repo and running `go build` 

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

`bcg` can take the following flags:

```      
Usage of ./bcg:
  -config string
        Configuration file in YAML, TOML, or JSON format (default "config.yml")
  -output string
        Directory to write output files to (default "output/")
  -socket string
        BIRD control socket (default "/run/bird/bird.ctl")
  -templates string
        Templates directory (default "/etc/bcg/templates/")
```

#### BGP Communities
bcg uses standard communities for 16-bit ASNs and large communities for 32-bit ASNs.

| Standard | Large     | Meaning                 |
|----------|-----------|-------------------------|
| ASN,100  | ASN,0,100 | Originated              |
| ASN,100  | ASN,0,101 | Learned from upstream   |
| ASN,100  | ASN,0,102 | Learned from peer       |
| ASN,100  | ASN,0,103 | Learned from downstream |

#### Global Configuration Options

| Option    | Usage                                                                                                         |
| --------- | ------------------------------------------------------------------------------------------------------------- |
| asn       | ASN of this router                                                                                            |
| router-id | Router ID of this router                                                                                      |
| prefixes  | List of prefixes to originate                                                                                 |
| irrdb     | IRRDB to query prefix sets from (default is rr.ntt.net which includes generated route objects from RPKI ROAs) |
| peers     | Map of name to peer (see below)                                                                               |

#### Peer Configuration Options

| Option         | Usage                                                                                                     |
| -------------- | --------------------------------------------------------------------------------------------------------- |
| asn            | Neighbor ASN                                                                                              |
| as-set         | Neighbor IRR AS-SET                                                                                       |
| maxpfx4        | Maximum number of IPv4 prefixes to accept before enacting `pfxlimitaction`                                |
| maxpfx6        | Maximum number of IPv6 prefixes to accept before enacting `pfxlimitaction`                                |
| pfxlimitaction | Action to take when the max prefix limits are tripped (warn, disable, block, or restart) default: disable |
| pfxfilter4     | IPv4 prefix filter list in BIRD format                                                                    |
| pfxfilter6     | IPv6 prefix filter list in BIRD format                                                                    |
| localpref      | BGP LOCAL_PREF                                                                                            |
| import-policy  | Peer import policy (any, cone, none)                                                                      |
| export-policy  | Peer export policy (any, cone, none)                                                                      |
| automaxpfx     | Should max prefix limits be pulled from PeeringDB?                                                        |
| autopfxfilter  | Should prefix filters be pulled from IRR data?                                                            |
| disabled       | Should neighbor sessions be disabled?                                                                     |
| passive        | Should neighbor sessions listen passively for BGP TCP connections?                                        |
| multihop       | Should neighbor sessions allow multihop?                                                                  |
| neighbors      | List of neighbor IP addresses                                                                             |
