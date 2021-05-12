# Configuration

Wireframe is configured by a single file in YAML, JSON, or TOML. All config file formats have the same configuration options but follow a different capitalization structure. YAML and JSON use all lowercase parameter names and TOML uses a more "natural" capitalization structure. For example, `router-id` in YAML and JSON is `Router-ID` in TOML.

## Global Configuration Options

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
| interfaces        | Map of network interface name to config                                                                         |
