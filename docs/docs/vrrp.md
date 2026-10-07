---
title: VRRP
sidebar_position: 6
---

Pathvector can generate a [keepalived](https://github.com/acassen/keepalived) configuration for VRRP. Define instances under the global `vrrp` key; the config is written to the `keepalived-config` file (default `/etc/keepalived.conf`) on each `pathvector generate`. See [VRRPInstance](/docs/configuration#vrrpinstance) for all options.

```yaml
keepalived-config: /etc/keepalived/keepalived.conf
vrrp:
  lan:
    state: primary  # or backup
    interface: eth0  # Interface that sends and receives VRRP advertisements
    vrid: 1
    priority: 255
    vips:
      - 192.0.2.1/24
      - 2001:db8::1/64
```

IPv4 VIPs are rendered in keepalived's `virtual_ipaddress` block and IPv6 VIPs in `virtual_ipaddress_excluded`.

## Binding VIPs to a different interface

By default, keepalived adds the virtual IPs to the same interface that sends VRRP packets (`interface`). Set `vip-interface` to add them to a different interface instead, for example to run VRRP over a dedicated link or VLAN while the VIPs live on a loopback or customer-facing interface:

```yaml
vrrp:
  lan:
    state: primary
    interface: eth0  # VRRP advertisements are sent on eth0
    vip-interface: eth1  # VIPs are added to eth1
    vrid: 1
    priority: 255
    vips:
      - 192.0.2.1/24
      - 2001:db8::1/64
```

This renders each VIP with keepalived's `dev` keyword:

```
vrrp_instance VRRPlan {
    state MASTER
    interface eth0
    virtual_router_id 1
    priority 255
    advert_int 1
    virtual_ipaddress {
        192.0.2.1/24 dev eth1
    }
    virtual_ipaddress_excluded {
        2001:db8::1/64 dev eth1
    }
}
```

`vip-interface` applies to all VIPs in the instance. Create separate instances if VIPs need to be bound to different interfaces.
