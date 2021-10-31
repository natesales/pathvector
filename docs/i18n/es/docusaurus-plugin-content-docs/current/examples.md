---
title: Ejemplos
---

# Ejemplos

Cada red esta única, así que revisa los [documentos de configuración](https://pathvector.io/docs/configuration) para escribir tu Pathvector para que se adapte a sus necesidades. Aquí hay un ejemplo de uso de comunidades BGP para controlar la política de enrutamiento:

## Comunidades

| Comunidad estándar | Comunidad grande | Significado               |
| :----------------- | :--------------  | :------------------------ |
| 65510,12           | 65510:0:12       | Aprendida de la corriente ascendente |
| 65510,13           | 65510:0:13       | Aprendida del servidor de rutas |
| 65510,14           | 65510:0:14       | Aprendida de los pares |
| 65510,15           | 65510:0:15       | Aprendida de la parte inferior de la red |

## Archivo de configuración (pathvector.yml)

```yaml
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
    announce-communities: [ "65510,15", "65510:0:15" ]
    remove-all-communities: 65510
    local-pref: 80
    import-communities: [ "65510,12", "65510:0:12" ]

  routeserver:
    filter-transit-asns: true
    auto-import-limits: true
    enforce-peer-nexthop: false
    enforce-first-as: false
    announce-communities: [ "65510,15", "65510:0:15" ]
    remove-all-communities: 65510
    local-pref: 90
    import-communities: [ "65510,13", "65510:0:13" ]

  peer:
    filter-irr: true
    filter-transit-asns: true
    auto-import-limits: true
    auto-as-set: true
    announce-communities: [ "65510,15", "65510:0:15" ]
    remove-all-communities: 65510
    local-pref: 100
    import-communities: [ "65510,14", "65510:0:14" ]

  downstream:
    filter-irr: true
    allow-blackhole-community: true
    filter-transit-asns: true
    auto-import-limits: true
    auto-as-set: true
    announce-communities: [ "65510,15", "65510:0:15" ]
    announce-default: true
    remove-all-communities: 65510
    local-pref: 200
    import-communities: [ "65510,15", "65510:0:15" ]

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
