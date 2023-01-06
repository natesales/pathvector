# AS Path

Pathvector supports a few types of AS path filtering:

## Downstream AS Cone

A peer's `filter-as-path` option enables downstream AS cone filtering. If a route's origin ASN isn't contained in the peer's AS-set, then it will be rejected. The AS-set can be defined manually with `as-set` or retrieved automatically from PeeringDB with `auto-as-set`.

## AS Provider Authorization (ASPA)

The global `authorized-providers` option defines a network's authorized transit providers. For example, the following snippet will enforce that routes originated by AS65510 may only be transited by AS65511 and AS65512. Similarly, routes originated by AS65500 may only be transited by AS65540.

```yaml
authorized-providers:
    65510: [65511, 65512]
    65500: [65540]
```

To enable ASPA filtering, set `filter-aspa` on a peer. If a route's origin ASN isn't contained in the peer's authorized providers, then it will be rejected unless the path only contains the peer's ASN (no providers in path).

## Transit ASNs

`filter-transit-asns` enables filtering of known transit ASNs. If a route's path contains a transit ASN, it will be rejected. Pathvector is preloaded with a [default set of transit ASNs](https://github.com/natesales/pathvector/blob/main/pkg/config/config.go), which can be overridden with the global `transit-asns` list.

## Transit Locking

The `transit-lock` peer option defines a list of authorized transit providers for the peer. If a route's AS path contains an ASN that isn't in the list, it will be rejected unless the AS path only contains the peer's ASN (no providers in path).
