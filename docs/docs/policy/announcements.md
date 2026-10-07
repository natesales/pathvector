# Announcements

This page describes the peer options that control which routes are exported (announced) to a peer.

## Restricting announcements by prefix

`dont-announce` and `only-announce` take a list of prefixes in BIRD prefix pattern syntax, so each entry may be a plain
prefix (`192.0.2.0/24`), a prefix with a length range (`192.0.2.0/24{24,32}`), or a prefix followed by `+` (the prefix
and all more specifics) or `-` (the prefix and all less specifics).

- `dont-announce` rejects matching routes before any other export policy is evaluated.
- `only-announce` rejects every route that doesn't match the list.

IPv4 and IPv6 prefixes can be mixed in the same list. Pathvector splits each list by address family and every BGP
channel only checks the prefixes of its own family, so a single peer with both IPv4 and IPv6 neighbors (or with
`mp-unicast-46`) can share one list:

```yaml
peers:
  Downstream:
    asn: 65510
    neighbors:
      - 192.0.2.10
      - 2001:db8::10
    only-announce:
      - 198.51.100.0/24
      - 2001:db8:1000::/36+
    dont-announce:
      - 198.51.100.128/25
      - 2001:db8:1fff::/48
```

If `only-announce` is set but contains no prefixes of an address family, nothing is announced in that address family
(for example, an `only-announce` list with only IPv4 prefixes means no IPv6 routes are announced to the peer).
`only-announce` cannot be combined with `announce-all`.

## Locally originated routes

`announce-originated` (enabled by default) announces locally originated routes to the peer. A route is considered
locally originated when it is

- within one of the global `prefixes` (the matching `local-communities` are added to it), or
- tagged with one of the global `origin-communities`, for example routes originated by another router in your network
  and learned over iBGP.

If neither `prefixes` nor `origin-communities` are configured, there is nothing to originate and `announce-originated`
is disabled for all peers. A router without local `prefixes` can therefore still announce routes originated elsewhere in
the network by tagging them with an origin community:

```yaml
asn: 65530
origin-communities:
  - 65530:0:1  # added to originated routes by the router that originates them
peers:
  Transit:
    asn: 64496
    neighbors:
      - 203.0.113.1
```

## Default route

`announce-default: true` announces a default route to the peer. The announced route is the best default route in
BIRD's table: the locally originated `default4`/`default6` route (requires the global `default-route`, enabled by
default), or a default route learned from another peer when the global `accept-default` is enabled. See
[Kernel](kernel#default-routes) for how `accept-default` affects route selection.
