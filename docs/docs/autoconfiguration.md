---
title: Autoconfiguration
sidebar_position: 5
---

Pathvector can build certain peer configurations automatically. The `match` command finds common IXPs for an ASN and outputs a plaintext email body or a Pathvector YAML snippet.

## YAML Config

Generated YAML output will be indented 2 spaces by default to fit within the YAML `peers` key. Plaintext output is selected by default, add the `--generate-config|-g` flag to select YAML generation mode.

```bash
~ ▴ grep ^asn /etc/pathvector.yml
asn: 34553
~ ▴ pathvector match 13335 -y
  Cloudflare SIX Seattle:
    asn: 13335
    neighbors:
      - 206.81.81.10
      - 2001:504:16::3417

  Cloudflare NWAX:
    asn: 13335
    neighbors:
      - 198.32.195.95
      - 2620:124:2000::95

  Cloudflare KCIX:
    asn: 13335
    neighbors:
      - 206.51.7.34
      - 2001:504:1b:1::34

  Cloudflare Speed-IX:
    asn: 13335
    neighbors:
      - 185.1.95.191
      - 2001:7f8:b7::a501:3335:1
```

## Plaintext

```bash
~ ▴ grep ^asn /etc/pathvector.yml
asn: 34553
~ ▴ pathvector match 13335
SIX Seattle: MTU 1500
  AS34553
  206.81.80.97
  2001:504:16::86f9

  AS13335
  206.81.81.10
  2001:504:16::3417

NWAX: Primary Peering VLAN
  AS34553
  198.32.195.28
  2620:124:2000::28

  AS13335
  198.32.195.95
  2620:124:2000::95

KCIX
  AS34553
  206.51.7.39
  2001:504:1b:1::39

  AS13335
  206.51.7.34
  2001:504:1b:1::34

Speed-IX: SPEED-IX
  AS34553
  185.1.95.166
  2001:7f8:b7::a503:4553:1

  AS13335
  185.1.95.191
  2001:7f8:b7::a501:3335:1
```

## Arbitrary ASN

The first ASN will be read from the `pathvector.yml` config file by default, but you can supply an alternate ASN with the `--local-asn|-l` flag if you want to find common IXPs for two arbitrary networks.

```bash
~ ▴ pathvector match -l 44977 34553
ARIX: Primary
  AS44977
  44.190.42.2
  2602:801:30ff::2

  AS34553
  44.190.42.3
  2602:801:30ff::3
```
