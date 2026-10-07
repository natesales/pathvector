---
sidebar_position: 5
---

# Route Optimization

Pathvector can use latency and packet loss metrics to make routing decisions. The optimizer works by sending ICMP or UDP ping out different peer networks and modifying BGP local pref according to average latency and packet loss thresholds.

## How it works

1. Run `pathvector generate` first. The optimizer works on the active BIRD configuration in the
   [`bird-directory`](/docs/configuration#bird-directory) (`/etc/bird/` by default) that `generate` writes.
2. Start the optimizer with `pathvector optimizer`. It probes the global `optimizer.targets` from each peer's
   `probe-sources` every `probe-interval` seconds and keeps the last `cache-size` results per peer.
3. When a peer's average latency or packet loss meets or exceeds `latency-threshold` or `packet-loss-threshold`, the
   optimizer runs the alert script (if configured) and, for peers with `optimize-inbound: true`, lowers the peer's local
   pref to `local-pref - modifier` by editing the peer's config file in the BIRD directory in place. The modified
   config is validated with BIRD and BIRD is reconfigured (unless `--no-configure` is set).

With `--dry-run`, the optimizer logs the local pref change it would make without modifying any files.

The next `pathvector generate` run restores the configured local pref.

```yaml
optimizer:
  targets: [ "192.0.2.53" ]
  latency-threshold: 100 # ms
  packet-loss-threshold: 0.5 # percent
  modifier: 20

peers:
  Transit:
    asn: 64496
    local-pref: 100
    neighbors:
      - 203.0.113.1
    probe-sources: [ "203.0.113.2" ]
    optimize-inbound: true
```

## Alert Scripts

To be notified of an optimization event, you can add a custom alert script that Pathvector will call when the latency or packet loss meet or exceed the configured thresholds.
