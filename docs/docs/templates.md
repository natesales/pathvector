---
title: Templates
sidebar_position: 5
---

# Templates

Templates hold peer options shared by several peers. A peer uses a template with the `template` option and inherits
every option it doesn't set itself.

```yaml
templates:
  upstream:
    local-pref: 80
    filter-irr: false

peers:
  Transit A:
    asn: 64496
    template: upstream
    neighbors:
      - 203.0.113.1
  Transit B:
    asn: 64497
    template: upstream
    local-pref: 90 # overrides the template's local-pref
    neighbors:
      - 203.0.113.5
```

## Template inheritance

A template can set a parent template with `template`. It inherits every option of its parent (and its parent's
parents) that it doesn't set itself, so common settings can live in one base template:

```yaml
templates:
  base:
    filter-transit-asns: true
    add-on-import: [ "65530:0:1" ]
  ix:
    template: base
    local-pref: 110
  ix-rs:
    template: ix
    enforce-first-as: false
```

Inheritance loops and references to undefined templates are reported as configuration errors.

## Merging lists

By default, a list or map option set on a peer replaces the template's value entirely. Set `merge-template-lists: true`
on the peer (or on a template, to apply it to all peers and child templates using it) to merge them instead:

- lists (for example `add-on-import`, `announce` or `prefixes`) are combined, template entries first, and
- map entries (for example `community-prefs` or `as-prefs`) are combined, with the peer's entries overriding the
  template's for the same key.

This allows a peer to add one or two communities to the communities of its exchange's template:

```yaml
templates:
  ix:
    add-on-import: [ "65530:0:100" ]
    announce: [ "65530:0:10" ]

peers:
  IX Peer:
    asn: 64496
    template: ix
    merge-template-lists: true
    announce: [ "65530:0:20" ] # announces routes tagged 65530:0:10 and 65530:0:20
    neighbors:
      - 198.51.100.2
```
