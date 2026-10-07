---
title: BGP Sessions
sidebar_position: 5
---

# BGP Sessions

This page covers peer options that change how BGP sessions and their next hops are set up. See the
[configuration reference](/docs/configuration#peers) for every option.

## Gateway mode

BIRD resolves a BGP route's next hop either *directly* (the next hop must be on a directly connected network) or
*recursively* (the next hop is looked up in the routing table, like an IGP next hop). BIRD uses direct mode for directly
connected eBGP neighbors and recursive mode for multihop and iBGP sessions.

The `gateway` peer option (`direct` or `recursive`) overrides this. A common case is an upstream reached over an IPv6
link-local address: the session can't use `multihop`, but the routes' next hops may still need to be resolved
recursively.

```yaml
peers:
  Upstream:
    asn: 64496
    neighbors:
      - fe80::1%eth0
    gateway: recursive
```

## Next hop cost

In recursive gateway mode (mainly multihop iBGP), BIRD uses the IGP metric of the path to the BGP next hop as a tie
breaker in best path selection. Sessions in direct gateway mode have no IGP metric, so all of them look equally close.

The `cost` peer option sets the distance to the next hop for routes from a session in direct gateway mode, for example
to prefer one directly connected iBGP neighbor over another when no IGP provides metrics:

```yaml
peers:
  Core A:
    asn: 65530
    direct: true
    cost: 10
    neighbors:
      - 192.0.2.2
  Core B:
    asn: 65530
    direct: true
    cost: 20
    neighbors:
      - 192.0.2.3
```

`cost` is a per-channel option in BIRD, which is why it can't be set with `session-global`.

## Source addresses

Pathvector has two kinds of "source" address:

- **Session source address**: `listen4` and `listen6` set the local address a BGP session is sourced from (BIRD's
  `local` address), for example a loopback address for multihop or iBGP sessions.
- **Kernel route source address**: the global `source4` and `source6` set the preferred source address
  (`krt_prefsrc`) of BGP routes installed in the kernel, which is the address the router itself uses when sending
  traffic along those routes. The same options can be set per peer (or template) to override the global value for
  routes learned from that peer, for example when one peer is reached over an IXP interface and another over a transit
  link.

```yaml
source4: 192.0.2.1        # loopback, default for kernel routes
source6: 2001:db8::1

peers:
  IXP Peer:
    asn: 64496
    neighbors:
      - 198.51.100.2
      - 2001:db8:ffff::2
    source4: 198.51.100.1 # use the IXP interface address for routes from this peer
    source6: 2001:db8:ffff::1
  iBGP:
    asn: 65530
    neighbors:
      - 192.0.2.2
    listen4: 192.0.2.1    # source the session from the loopback
    multihop: true
```

The kernel source address must be configured on an interface of the router, otherwise the kernel rejects the routes.

## L3VPN (VPNv4/VPNv6)

Set `l3vpn: true` on a peer to exchange MPLS L3VPN routes (RFC 4364 VPNv4 and VPNv6) with it, in addition to unicast
routes. This lets Pathvector act as a BGP route reflector for an MPLS L3VPN network whose PE routers do the label
switching, for example cheap MPLS-capable switches with small FIBs.

VPN routes are stored in the BIRD tables `vpntab4` and `vpntab6` (created when at least one peer has `l3vpn` enabled)
and are reflected unchanged between peers: Pathvector doesn't filter VPN routes, install them in the kernel or
allocate MPLS labels. `import: false`, `export: false`, the global `no-accept`/`no-announce`, `next-hop-self` and the
`add-path-*` options also apply to the VPN channels.

```yaml
asn: 65530
peers:
  PE1:
    asn: 65530
    rr-client: true
    l3vpn: true
    neighbors:
      - 192.0.2.11
  PE2:
    asn: 65530
    rr-client: true
    l3vpn: true
    neighbors:
      - 192.0.2.12
```

Inspect VPN routes with `birdc show route table vpntab4`.
