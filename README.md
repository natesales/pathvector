# bcg

[![Go Report](https://goreportcard.com/badge/github.com/natesales/bcg?style=for-the-badge)](https://goreportcard.com/report/github.com/natesales/bcg) 
[![License](https://img.shields.io/github/license/natesales/bcg?style=for-the-badge)](https://choosealicense.com/licenses/gpl-3.0/) 
[![Release](https://img.shields.io/github/v/release/natesales/bcg?style=for-the-badge)](https://github.com/natesales/bcg/releases) 

The automatic BIRD configuration generator with bogon, IRR, RPKI, and max prefix filtering support.

### Installation
bcg depends on [bird2](https://gitlab.nic.cz/labs/bird/), [GoRTR](https://github.com/cloudflare/gortr), and [bgpq4](https://github.com/bgp/bgpq4). Make sure the `bird` and `gortr` daemons are running and `bgpq4` is in path before running bcg. bcg is available for amd64 as a prebuilt deb package and binary for each release. Releases can be downloaded from Github or by adding `deb [trusted=yes] https://apt.fury.io/natesales/ /` to your `/etc/apt/source.list` file. You can also build from source by cloning the repo and running `go build`. It's recommended to run bcg every 12 hours to update IRR prefix lists and PeeringDB prefix limits. Adding `0 */12 * * * /usr/bin/bcg` to your crontab will update the filters at 12 AM and PM.

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
Usage for bcg https://github.com/natesales/bcg:
  -config string
        Configuration file in YAML, TOML, or JSON format (default "config.yml")
  -dryrun
        Skip modifying BIRD config. This can be used to test that your config syntax is correct.
  -output string
        Directory to write output files to (default "/etc/bird/")
  -socket string
        BIRD control socket (default "/run/bird/bird.ctl")
  -templates string
        Templates directory (default "/etc/bcg/templates/")
  -version
        Print bcg version and exit
```

#### How does filtering work?
bcg applies a universal pre-filter to all BGP sessions before evaluating IRR or manual prefix lists which rejects the following:
- Own prefixes as defined in the global `prefixes` list
- Have a [bogon ASN](https://github.com/natesales/bcg/blob/main/templates/global.tmpl#L176) anywhere in the AS_PATH
- Have a total AS_PATH length of more than 100
- IPv4 prefixes that are...
    - length > 24
    - length < 8
    - RPKI invalid
    - contained in the [bogons list](https://github.com/natesales/bcg/blob/main/templates/global.tmpl#L126)
- IPv6 prefixes that are...
    - length > 48
    - length < 12
    - RPKI invalid
    - contained in the [bogons list](https://github.com/natesales/bcg/blob/main/templates/global.tmpl#L143)

All peers with an import filter of `cone` will apply further strict filtering by either an AS Set or manual prefix list. Max-prefix limits are also enforced for every peer.

#### Local Preference
All sessions have a default BGP LOCAL_PREF of 100, except for routes tagged with community `65535, 0` ([RFC8326 Graceful Shutdown](https://tools.ietf.org/html/rfc8326)). Local pref can be adjusted on a per-peer basis with the `localpref` option under the peer block.

#### Pre-import and Pre-export conditions
There are many features of BIRD that aren't part of bcg. If you want to add a statement before importing or exporting of routes, you can supply a multiline in `preimport` or `preexport` in the peer block to include that snippet of BIRD code after the import prefilter or before the export filter respectively.

#### BGP Communities
bcg uses RFC 8092 BGP Large Communities

| Large     | Meaning                 |
|-----------|-------------------------|
| ASN,0,100 | Originated              |
| ASN,0,101 | Learned from upstream   |
| ASN,0,102 | Learned from peer       |
| ASN,0,103 | Learned from downstream |

#### Global Configuration Options

| Option    | Usage                                                                                                         |
| --------- | ------------------------------------------------------------------------------------------------------------- |
| asn       | ASN of this router                                                                                            |
| router-id | Router ID of this router                                                                                      |
| prefixes  | List of prefixes to originate                                                                                 |
| irrdb     | IRRDB to query prefix sets from (default is rr.ntt.net which includes generated route objects from RPKI ROAs) |
| rtrserver | IP address or hostname of RPKI RTR server (default is 127.0.0.1)                                              |
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
| preimport      | List of BIRD expressions to execute after the prefilter and before the prefix filter                      |
| prepends       | Number of times to prepend local AS to                                                                    |
