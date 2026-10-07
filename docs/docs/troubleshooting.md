---
title: Troubleshooting
sidebar_position: 8
---

## Automatic shutdown: Route limit exceeded

If a BGP session goes down and `birdc show protocols` reports

```
Automatic shutdown: Route limit exceeded
```

the peer sent more routes than the session's configured route limit allows. This is most common with transit providers and large networks that send a full table (for example, Hurricane Electric's IPv6 table), which can be larger than Pathvector's default limits.

### Default limits

Every peer gets an import limit, which caps how many prefixes BIRD will accept from the peer **after filtering**:

| Option          | Default   |
|-----------------|-----------|
| `import-limit4` | `1000000` |
| `import-limit6` | `300000`  |

Receive limits (`receive-limit4`/`receive-limit6`) count routes **before filtering** and require the global `keep-filtered` option. Export limits (`export-limit4`/`export-limit6`) cap the number of routes announced to the peer. Receive and export limits are not set by default.

If `auto-import-limits` is enabled for a peer, `import-limit4` and `import-limit6` are replaced with the `info_prefixes4` and `info_prefixes6` values from the peer's PeeringDB record. These values are maintained by the peer and are often set for what the network *originates*, not for a full table. Don't use `auto-import-limits` on transit sessions; set explicit limits instead.

See [Route Limits](/docs/filtering/route-limits) for more detail.

### Violation actions

`import-limit-violation`, `receive-limit-violation`, and `export-limit-violation` control what BIRD does when a limit is exceeded. The default is `disable`.

| Action    | Behavior                                                                                         |
|-----------|--------------------------------------------------------------------------------------------------|
| `warn`    | Log a warning and keep the session up                                                            |
| `block`   | Keep the session up but ignore any further routes beyond the limit                               |
| `restart` | Restart the session (it will be shut down again if the peer still sends too many routes)         |
| `disable` | Shut down the session and keep it down until it's manually re-enabled (`Automatic shutdown`)    |

### Recovering a session

With the `disable` action, the session stays down after the limit trips, even after you raise the limit. Once you've increased the limit and run `pathvector generate`, bring the session back up with `birdc`. Pathvector's protocol names are the peer name in uppercase with non-alphanumeric characters removed, followed by `_AS<asn>_v4` or `_AS<asn>_v6`:

```bash
birdc show protocols                 # find the protocol name
birdc enable HURRICANE_AS6939_v6     # re-enable a disabled session
birdc restart HURRICANE_AS6939_v6    # or restart it
```

### Example: full-table transit peer

```yaml
peers:
  Hurricane:
    asn: 6939
    neighbors:
      - 203.0.113.1
      - 2001:db8::1
    local-pref: 80
    # Full table sizes grow over time, leave plenty of headroom
    import-limit4: 1500000
    import-limit6: 500000
    # Don't take the session down if the limit is reached; log and stop accepting new routes instead
    import-limit-violation: block
    # Use explicit limits, PeeringDB values describe originated routes and are too small for a full table
    auto-import-limits: false
    # Transit providers send routes from the whole DFZ
    filter-irr: false
    filter-transit-asns: false
```
