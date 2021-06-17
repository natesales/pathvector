# FastNetMon for DDoS mitigation with Pathvector on Linux

[FastNetMon](https://fastnetmon.com) is a software-based DoS/DDoS detection tool that analyzes flows and signals a mitigation action over BGP such as traffic redirection or selective destination blackholes.

To get started with Pathvector and FastNetMon, first install [Pathvector](/installation) and [FastNetMon Advanced](https://fastnetmon.com/docs-fnm-advanced/advanced-install-guide/).

## Configure FastNetMon

With Pathvector, BIRD will be listening on the default BGP port (179) so FastNetMon needs to listen on a different port.

From the `fcli` prompt:

```shell
set main networks_list 198.51.100.0/24
set main mirror_afpacket enable
set main interfaces bond0

set main gobgp enable
set main gobgp_ipv6 enable
set main gobgp_announce_host enable
set main gobgp_announce_host_ipv6 enable
set main gobgp_next_hop 192.0.2.1
set main gobgp_next_hop_ipv6 100::1
set main gobgp_bgp_listen_port 1179

set bgp pathvector
set bgp pathvector local_asn 65530
set bgp pathvector remote_asn 65530
set bgp pathvector local_address 127.0.0.2
set bgp pathvector remote_address 127.0.0.1
set bgp pathvector ipv4_unicast enable
set bgp pathvector ipv6_unicast enable
set bgp pathvector active enable

commit
```

## Configure Pathvector

Aside from the standard fields like `asn` and `neighbors`, the Pathvector config needs a few extra options for the FastNetMon session. By default, /32 and /128 routes will be filtered by prefix length, so `filter-prefix-length` must be disabled. ROAs may have a maxLength that would cause the routes to be filtered, so `filter-rpki` must be disabled as well.

```yaml
peers:
  FastNetMon:
    asn: 65530
    filter-rpki: false
    filter-prefix-length: false
    enforce-first-as: false
    enforce-peer-nexthop: false
    neighbor-port: 1179
    import-communities:
      - 65530,666
    mp-unicast-46: true
    listen: 127.0.0.1
    neighbors:
      - 127.0.0.2
```
