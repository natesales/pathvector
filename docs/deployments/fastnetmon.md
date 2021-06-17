# FastNetMon for DDoS mitigation with Pathvector on Linux

[FastNetMon](https://fastnetmon.com) is a software-based DoS/DDoS detection tool that analyzes flows and signals a mitigation action over BGP such as traffic redirection or selective destination blackholes.

To get started with Pathvector and FastNetMon, first install [Pathvector](/installation) and [FastNetMon Advanced](https://fastnetmon.com/docs-fnm-advanced/advanced-install-guide/).

## Configure FastNetMon

With Pathvector, BIRD will be listening on the default BGP port (179) so FastNetMon needs to listen on a different port.

```shell
fcli set main gobgp_bgp_listen_port 1179
```

## Configure Pathvector

Aside from the standard fields like `asn` and `neighbors`, the Pathvector config needs a few extra options for the FastNetMon session. By default, /32 and /128 routes will be filtered by prefix length, so `filter-prefix-length` must be disabled. ROAs may have a maxLength that would cause the routes to be filtered, so `filter-rpki` must be disabled as well. 

```yaml
FastNetMon:
  asn: 65530
  filter-rpki: false
  filter-prefix-length: false
  neighbor-port: 1179
  import-communities:
    - 65530,666
  neighbors:
    - 127.0.0.1
    - ::1
```
