# Blackholing

Pathvector supports remotely triggered blackholing (RTBH).

## Accepting blackhole requests from a peer

Set `allow-blackhole-community: true` on a peer to let it request blackholing. A route is treated as a blackhole
request when it carries the large community `(<your ASN>, 1, 666)` and is a host route (`/32` for IPv4 or `/128` for
IPv6). Matching routes get their next hop rewritten to a null route (`192.0.2.1` / `100::1`, which are installed as
blackhole routes in BIRD).

Blackhole requests are by definition more specific than the `filter-prefix-length` bounds (IPv4 `/8`-`/24`, IPv6
`/12`-`/48`), so when `allow-blackhole-community` is enabled, blackhole requests skip the prefix length check. Every
other import filter (bogons, RPKI, IRR, prefix sets, AS sets, ...) still applies to them.

```yaml
peers:
  Customer:
    asn: 65510
    neighbors:
      - 192.0.2.10
    allow-blackhole-community: true
```

## Blackholing all routes from or to a peer

- `blackhole-in` blackholes every route imported from the peer.
- `blackhole-out` blackholes every route exported to the peer.
