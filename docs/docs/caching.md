---
title: Caching
sidebar_position: 9
---

Pathvector relies on external datasources to generate configuration, such as PeeringDB, IRR databases, and the RPKI. There are various mechanisms to cache this data to decrease latency and reduce load on these external services.

## RPKI

Networks should already be running their own RTR (RPKI to Router) server such as [stayrtr](https://github.com/bgp/stayrtr) or [rtrtr](https://github.com/NLnetLabs/rtrtr).

## IRR

Pathvector stores the results of every successful IRR query on disk, per peer, in the [`cache-directory`](https://pathvector.io/docs/configuration/#cache-directory):

- `<cache-directory>/irr/AS<asn>_<peer name>.json` holds the IPv4 and IPv6 prefix sets (`filter-irr`) and the AS set members (`auto-as-set-members`) from the last successful query, along with the time of that query.

If a later IRR query for that peer fails (for example because the IRR server is unreachable or times out), Pathvector logs a warning showing how old the cached data is and uses it instead:

```
level=warning msg="unable to get IPv4 IRR prefix list from AS-EXAMPLE: exit status 1: ...; using cached IRR prefix sets for AS-EXAMPLE from 2024-01-01T00:00:00Z (26h0m0s old)"
```

Cached data is only used for exactly the same query: if the peer's as-set, `bgpq-args` or `irr-accept-child-prefixes` changed since the data was cached, it is ignored. If there is no usable cached data, the peer fails safe and rejects all imports (see [IRR failure handling](filtering/irr.md#failure-handling)).

## PeeringDB

Pathvector has an internal PeeringDB cache that stores PeeringDB objects *for the duration of a single `pathvector generate` run*. This does not cache for longer than a single command invocation.

The in-memory cache is controlled by the global [`peeringdb-cache`](https://pathvector.io/docs/configuration/#peeringdb-cache) option (enabled by default). When enabled, peers that share an ASN only query PeeringDB once per run. `peeringdb-cache` is a global option, so it must be set at the top level of the config file, not under a peer or template.

### Disabling PeeringDB queries for a peer

Setting `peeringdb-cache: false` doesn't stop PeeringDB queries, it only disables the in-memory cache. Pathvector queries PeeringDB for a peer only when [`auto-import-limits`](https://pathvector.io/docs/configuration/#auto-import-limits) or [`auto-as-set`](https://pathvector.io/docs/configuration/#auto-as-set) is enabled for that peer (directly, or through a template). To stop querying PeeringDB for a single peer, disable both options and set the values manually:

```yaml
peers:
  Example:
    asn: 65510
    neighbors:
      - 203.0.113.12
    auto-import-limits: false
    auto-as-set: false
    import-limit4: 100
    import-limit6: 50
    as-set: AS-EXAMPLE
```

Separately from per-peer queries, Pathvector makes one global PeeringDB query per run for the list of networks that should never be reachable via route servers when a peer has [`filter-never-via-route-servers`](https://pathvector.io/docs/configuration/#filter-never-via-route-servers) set.

In addition, the results of successful PeeringDB queries are stored on disk in the [`cache-directory`](https://pathvector.io/docs/configuration/#cache-directory), and used with a warning when PeeringDB can't be queried:

- `<cache-directory>/peeringdb/AS<asn>.json` holds a network's PeeringDB data (used by `auto-import-limits` and `auto-as-set`).
- `<cache-directory>/peeringdb/never-via-route-servers.json` holds the never via route servers list (used by `filter-never-via-route-servers`).

If a network has no PeeringDB page, that is treated as a definitive answer and no cached data is used. If PeeringDB data is required and neither a live query nor cached data is available, generation stops with an error, as before.

## Offline generation

`pathvector generate --offline` skips all live IRR and PeeringDB queries and uses only the data cached on disk by previous runs. This is useful when the router can't reach the internet but its configuration still needs to be regenerated, for example after a local config change. Peers without cached IRR data reject all imports, and peers that need PeeringDB data without cached data cause an error.

:::note
The default `cache-directory` (`/var/run/pathvector/cache/`) is usually on a tmpfs and is cleared on reboot. To keep the cached IRR and PeeringDB data across reboots, set `cache-directory` to a persistent location such as `/var/cache/pathvector/`.
:::

### PeeringDB Local Cache

To cache PeeringDB data persistently, you can set the global [`peeringdb-url`](https://pathvector.io/docs/configuration/#peeringdb-url) option to a local [PeeringDB cache server](https://github.com/natesales/peeringdb-cache).
