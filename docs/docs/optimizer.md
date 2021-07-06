---
sidebar_position: 5
---

# Route Optimization

Pathvector can use latency and packet loss metrics to make routing decisions. The optimizer works by sending ICMP or UDP ping out different peer networks and modifying BGP local pref according to average latency and packet loss thresholds.

## Alert Scripts

To be notified of an optimization event, you can add a custom alert script that Pathvector will call when the latency or packet loss meet or exceed the configured thresholds.
