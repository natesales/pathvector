# Local Preference

## Setting local preference on import

By default, Pathvector sets the BGP local preference of every route imported from a peer to the peer's `local-pref`
(default `100`). Related options:

- `local-pref4` / `local-pref6` override `local-pref` for one address family.
- `set-local-pref: false` disables setting the local preference on import.
- `default-local-pref` sets BIRD's `default bgp_local_pref`, which only applies to routes received without a local
  preference (eBGP), instead of overwriting it.

## iBGP sessions

On iBGP, the local preference received from the neighbor usually carries policy decided at the network edge, so
Pathvector doesn't overwrite it by default. A session is treated as iBGP when the peer's `asn` equals the global `asn`
(or the peer's `local-asn`).

On an iBGP session, Pathvector only sets the local preference when one of `local-pref`, `local-pref4`, `local-pref6` or
`set-local-pref` is configured on the peer (or its template), or when `optimize-inbound` is enabled for the peer.

```yaml
asn: 65530
peers:
  Core:            # iBGP, local preference received from the neighbor is kept
    asn: 65530
    neighbors:
      - 192.0.2.2
  Core override:   # iBGP, local preference explicitly set to 50
    asn: 65530
    local-pref: 50
    neighbors:
      - 192.0.2.3
```

## Local preference by community

`community-prefs` maps a standard or large community to a local preference: routes carrying the community get that
local preference on import.

```yaml
community-prefs:
  "65530:0:200": 95
  "64496:100": 110
```

## Local preference by AS

`as-prefs` maps an ASN to a local preference: routes whose AS path contains the ASN get that local preference on
import. `as-prefs4` and `as-prefs6` do the same for only IPv4 or only IPv6 routes, and take precedence over `as-prefs`
for their address family.

```yaml
peers:
  Transit:
    asn: 64496
    neighbors:
      - 203.0.113.1
      - 2001:db8::1
    as-prefs:
      174: 90      # depreference paths via AS174 in both address families
    as-prefs6:
      6939: 120    # prefer IPv6 paths via AS6939
```

## Local preference by prefix

`prefix-prefs` maps a prefix to a local preference: routes matching the prefix get that local preference on import.
Keys use BIRD prefix pattern syntax, so `198.51.100.0/24` matches exactly that prefix, `198.51.100.0/24+` also matches
more specifics and `2001:db8::/32{32,48}` matches lengths 32 to 48. IPv4 and IPv6 prefixes can be mixed.

Together with `as-prefs` (AS path contains an ASN), this allows simple traffic engineering for a single peer without
custom BIRD configuration:

```yaml
peers:
  Transit A:
    asn: 64496
    neighbors:
      - 203.0.113.1
    prefix-prefs:
      "198.51.100.0/24+": 200        # prefer Transit A for this destination
      "2001:db8:1000::/36{36,48}": 200
    as-prefs:
      64511: 80                      # avoid paths through AS64511 via Transit A
```

For more complex matching (for example AS path regular expressions), use `pre-import-accept` with custom BIRD filter
statements.

## Precedence

When several options match the same route, they are evaluated in this order, and the last match wins:

1. `local-pref` (or the received value on iBGP, see above)
2. `local-pref4` / `local-pref6`
3. `community-prefs`
4. `as-prefs`
5. `as-prefs4` / `as-prefs6`
6. `prefix-prefs`

For example, with `community-prefs: {"65530:0:200": 95}` and `as-prefs: {6939: 150}`, a route via AS6939 tagged with
`65530:0:200` gets a local preference of 150.

:::note
Before this ordering was documented, `as-prefs` was evaluated before `community-prefs`, so community preferences took
precedence over AS preferences.
:::
