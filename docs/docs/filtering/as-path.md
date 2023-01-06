# AS Path (ASPA and Downstream Cone)

Pathvector supports two types of AS path filtering:

## Downstream Cone

A peer's `filter-as-path` option enables downstream AS cone filtering. If a route's origin ASN isn't contained in the peer's AS-set, then it will be rejected. The AS-set can be defined manually with `as-set` or retrieved automatically from PeeringDB with `auto-as-set`.

## Transit Locking

The `transit-lock` peer option defines a list of authorized transit providers for the peer. If a route's AS path contains an ASN that isn't in the list, it will be rejected. ***WARNING: This option rejects routes originated directly by the peer (i.e, the AS path cannot only contain the peer's ASN). Be careful with this option.***
