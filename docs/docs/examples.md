---
title: Examples
---

# Examples

Every network is unique, so check out the [configuration docs](https://pathvector.io/docs/configuration) to write your
Pathvector config to suit your needs. Here's an example of using BGP communities to control routing policy:

## Communities

| Standard Community | Large Community | Meaning                   |
|:-------------------|:----------------|:--------------------------|
| 65510,12           | 65510:0:12      | Learned from upstream     |
| 65510,13           | 65510:0:13      | Learned from route server |
| 65510,14           | 65510:0:14      | Learned from peer         |
| 65510,15           | 65510:0:15      | Learned from downstream   |

## Config File

```yaml title="/etc/pathvector.yml"
asn: 65510
router-id: 192.0.2.1
source4: 192.0.2.1
source6: 2001:db8::1
prefixes:
  - 192.0.2.0/24
  - 2001:db8::/48

templates:
  upstream:
    allow-local-as: true
    announce: [ "65510,15", "65510:0:15" ]
    remove-all-communities: 65510
    local-pref: 80
    add-on-import: [ "65510,12", "65510:0:12" ]

  routeserver:
    filter-transit-asns: true
    auto-import-limits: true
    enforce-peer-nexthop: false
    enforce-first-as: false
    announce: [ "65510,15", "65510:0:15" ]
    remove-all-communities: 65510
    local-pref: 90
    add-on-import: [ "65510,13", "65510:0:13" ]

  peer:
    filter-irr: true
    filter-transit-asns: true
    auto-import-limits: true
    auto-as-set: true
    announce: [ "65510,15", "65510:0:15" ]
    remove-all-communities: 65510
    local-pref: 100
    add-on-import: [ "65510,14", "65510:0:14" ]

  downstream:
    filter-irr: true
    allow-blackhole-community: true
    filter-transit-asns: true
    auto-import-limits: true
    auto-as-set: true
    announce: [ "65510,15", "65510:0:15" ]
    announce-default: true
    remove-all-communities: 65510
    local-pref: 200
    add-on-import: [ "65510,15", "65510:0:15" ]

peers:
  Hurricane Electric:
    asn: 6939
    template: upstream
    neighbors:
      - 203.0.113.66
      - 2001:db8:55::66
      - 203.0.113.67
      - 2001:db8:55::67

  ARIX Route Servers:
    asn: 47192
    template: routeserver
    password: VoP5ViKtjvw4CMG
    neighbors:
      - 203.0.113.251
      - 2001:db8:55::251
      - 203.0.113.252
      - 2001:db8:55::252

  Cloudflare:
    asn: 13335
    template: peer
    neighbors:
      - 203.0.113.95
      - 2001:db8:55::95

  AS112:
    asn: 112
    template: downstream
    neighbors:
      - 192.0.2.20
      - 2001:db8::20
```
