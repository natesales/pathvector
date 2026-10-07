# Communities

Communities can be written as standard communities (`65530:100`, or `65530,100`) or large communities
(`65530:0:100`) in every community option.

## Named communities

The global `communities` option names communities, so the name can be used in place of the community in any community
option. This keeps large configurations readable and lets you change a community in one place.

```yaml
communities:
  BLACKHOLE: 65535:666
  FROM-CUSTOMER: 65530:0:10
  FROM-PEER: 65530:0:20
  FROM-TRANSIT: 65530:0:30

templates:
  customer:
    add-on-import: [ FROM-CUSTOMER ]
  peer:
    add-on-import: [ FROM-PEER ]
    announce: [ FROM-CUSTOMER ]

peers:
  Transit:
    asn: 64496
    neighbors:
      - 203.0.113.1
    add-on-import: [ FROM-TRANSIT ]
    announce: [ FROM-CUSTOMER ]
    community-prefs:
      BLACKHOLE: 0
```

Names are supported in these options:

- Global: `origin-communities`, `local-communities`, `add-on-import`, `add-on-export` and `kernel.srd-communities`
- Peer: `add-on-import`, `add-on-export`, `announce`, `remove-communities`, the community lists of
  `prefix-communities` and the keys of `community-prefs`

A name must not itself look like a community, and every name must map to a valid standard or large community. Using a
name that isn't defined is an error, since it isn't a valid community either.
