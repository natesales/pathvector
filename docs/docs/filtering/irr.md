# IRR

IRR filtering uses [bgpq4](https://github.com/bgp/bgpq4) to generate sets of prefixes and ASNs.

## Global configuration

`irr-server` sets the IRR server address

`bgpq-args` adds additional arguments to pass to `bgpq4` (for example to limit IRR sources with `-S RIPE`)

### IRR sources

An as-set may be prefixed with the IRR database (source) that it should be looked up in, using the `SOURCE::AS-SET` syntax, for example `RIPE::AS-EXAMPLE`. This syntax is commonly found in PeeringDB `irr_as_set` values, so it is also used by `auto-as-set`.

How the source prefix is handled depends on whether `bgpq-args` already restricts the sources with a `-S` flag:

- If `bgpq-args` does **not** contain `-S`, the as-set's source is passed to bgpq4, so `RIPE::AS-EXAMPLE` is queried as `bgpq4 ... -S RIPE AS-EXAMPLE`.
- If `bgpq-args` **does** contain `-S`, your source list takes precedence: the `SOURCE::` prefix is stripped and no extra `-S` flag is added, so `RIPE::AS-EXAMPLE` is queried as `bgpq4 -S RPKI,RADB,RIPE ... AS-EXAMPLE`. (bgpq4 only honours the last `-S` flag, so adding the as-set's source would silently replace your list and can return an empty result.)

This applies to all bgpq4 queries: prefix sets (`filter-irr`) and AS set members (`auto-as-set-members`).

```yaml
bgpq-args: -S RPKI,AFRINIC,ARIN,APNIC,LACNIC,RADB,RIPE
peers:
  Example:
    asn: 65530
    auto-as-set: true # PeeringDB returns RIPE::AS-EXAMPLE, queried as AS-EXAMPLE using the sources above
    filter-irr: true
```

## Peer configuration

Enable `filter-irr` to enable IRR filtering.

`filter-irr` and `auto-as-set-members` both need the peer's as-set: either set `as-set` directly or enable
`auto-as-set` to get it from PeeringDB. Loading a config where one of them is enabled without either fails with an
error such as `[Example] filter-irr requires as-set or auto-as-set`.

Enable `filter-as-members` to reject routes that aren't originated from an ASN within the peer's `as-members` list.
Enable `auto-as-set-members` to retrieve that list automatically from their PeeringDB IRR object.

## Peer policy verification

Enable `verify-irr-policy` on a peer to check that the peer's own routing policy, as published in its RPSL `aut-num` object, still documents the session with you. This is useful to automatically shut down sessions with peers that have stopped peering with you (or never documented it), for example after a peer has moved to a different upstream.

Pathvector queries the peer's `aut-num` object (`AS<peer ASN>`) from the [`irr-server`](https://pathvector.io/docs/configuration/#irr-server) over WHOIS (TCP port 43, or the port given as `host:port`), using the `irr-query-timeout`. If `bgpq-args` restricts the IRR sources with `-S`, the WHOIS query is restricted to the same sources. The session is considered valid if the object contains both:

- an import rule `from <X> ... accept ...` (`import:` or `mp-import:`), and
- an export rule `to <X> ... announce ...` (`export:` or `mp-export:`),

where `<X>` is your global `asn` as `AS<asn>`, `AS-ANY`, or an as-set whose members (expanded recursively with bgpq4) include your ASN. `accept NOT ANY` and `announce NOT ANY` don't count.

The check is address family aware and is done for each family the peer has `neighbors` in:

- IPv4 is covered by plain `import:`/`export:` rules, and by `mp-import:`/`mp-export:` rules with `afi ipv4`, `ipv4.unicast`, `any` or `any.unicast` (or no `afi` at all).
- IPv6 is covered by `mp-import:`/`mp-export:` rules with `afi ipv6`, `ipv6.unicast`, `any` or `any.unicast` (or no `afi` at all).

For example, if your ASN is 6939, both of these `aut-num` objects document an IPv6 session with you (the first one assuming `AS199514:AS-UPSTREAMS` contains AS924 and AS6939); only the second one also documents an IPv4 session:

```
aut-num:   AS207960
mp-export: afi ipv6.unicast to AS199514:AS-UPSTREAMS announce AS-RAPDODGE
mp-import: afi ipv6.unicast from AS199514:AS-UPSTREAMS accept ANY
```

```
aut-num:   AS207960
mp-export: afi any.unicast to AS6939 announce AS-ROUTE48
mp-import: afi any.unicast from AS6939 accept ANY
```

```yaml
asn: 6939
peers:
  Example:
    asn: 207960
    verify-irr-policy: true
    neighbors:
      - 2001:db8::1
```

If policy is missing, the peer is disabled (`disabled: true`) and a warning names what is missing:

```
level=warning msg="[Example] disabling peer: aut-num AS207960 is missing IRR policy: import from AS6939 accept (IPv4), export to AS6939 announce (IPv4)"
```

A peer has a single `disabled` option for all of its neighbors, so **the whole peer is disabled if any address family it has neighbors in lacks policy**. To keep an IPv6 session up when the peer only documents IPv6, configure the IPv4 and IPv6 neighbors as separate peers.

If the WHOIS query fails, or an as-set needed for the decision can't be expanded, Pathvector logs a warning and leaves the peer as configured. The check is skipped with `pathvector generate --offline`. If the server has several `aut-num` objects for the ASN (for example mirrored from different IRR databases), the rules of all of them are combined. Only the peer expression, the `accept`/`announce` keyword and the `afi` list are interpreted; filters and actions are not evaluated.

## Failure handling

IRR lookups depend on an external service, so a single unreachable IRR server, a timeout, or a broken as-set must not prevent the rest of the router configuration from being generated. If a bgpq4 query fails for a peer, Pathvector logs an error and keeps going with the other peers. The affected peer **fails safe**: its `import` is set to `false`, so every route received from it is rejected until the next successful run. Sessions stay up and routes are still exported to the peer.

```
level=error msg="[Example] IRR prefix set generation failed: unable to get IPv4 IRR prefix list from AS-EXAMPLE: exit status 1: ...; rejecting all imports from AS65530"
```

The peer also rejects all imports when:

- `filter-irr` is enabled but IRR returns no IPv4 *and* no IPv6 prefixes for the peer. (If only one address family is empty, that family's prefix set is empty and rejects all routes of that family, while the other family is filtered normally.)
- `auto-as-set-members` fails, or returns no members while `filter-as-set` is enabled. `filter-as-set` is then turned off for that peer, because BIRD can't use an empty AS set (imports are rejected anyway).

If an earlier run cached IRR data for the peer, Pathvector uses that data with a warning instead of rejecting imports. See [Caching](../caching.md#irr).
