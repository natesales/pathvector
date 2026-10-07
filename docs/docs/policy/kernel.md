# Kernel

Pathvector configures two BIRD kernel protocols that synchronise BIRD's routing tables with the operating system's
routing tables:

| BIRD protocol | Address family |
|---------------|----------------|
| `kernel4`     | IPv4           |
| `kernel6`     | IPv6           |

Use these names with `birdc`, for example `birdc show route export kernel4`. (Older versions left the protocols
unnamed, so BIRD called them `kernel1` and `kernel2`.)

Kernel options live under the global `kernel` key, see [the configuration reference](/docs/configuration#kernel):

```yaml
kernel:
  table: 100              # kernel table to sync with
  scan-time: 10           # seconds between kernel table scans
  learn: false            # learn routes from the kernel into BIRD
  export: true            # export routes to the kernel
  reject-connected: false # don't export connected (RTS_DEVICE) routes
  accept4: []             # BIRD protocols to always export to the kernel (IPv4)
  reject4: []             # BIRD protocols to never export to the kernel (IPv4)
```

## Multiple kernel tables

`kernel.tables` lists additional kernel routing tables to export routes to, for example for policy based routing or a
management VRF. Every additional table gets the same export policy as the main kernel table (`kernel.table`, SRD
communities, statics, source addresses, ...).

```yaml
kernel:
  table: 254      # main table
  tables: [ 100, 200 ]
```

Because BIRD only allows one kernel protocol per BIRD routing table, Pathvector creates, for each additional table `N`
and address family:

| BIRD object                 | Purpose                                          |
|-----------------------------|--------------------------------------------------|
| `kernel4_tableN_rib`        | BIRD routing table holding a copy of `master4`   |
| `kernel4_tableN_pipe`       | Pipe copying routes from `master4` into it       |
| `kernel4_tableN`            | Kernel protocol exporting to kernel table `N`    |

(and the same with `6` for IPv6). Routes are only exported to the additional tables, they are not learned from them.

## Static routes

`kernel.statics` maps a prefix to a next hop (optionally with an interface, `fe80::1%eth0`). Static routes are placed
in their own BIRD protocols, `statics4` and `statics6`, separate from the locally originated `prefixes` (`static4` /
`static6`), and are always exported to the kernel when `kernel.export` is enabled, including when `srd-communities` or
`source4`/`source6` are configured.

```yaml
kernel:
  statics:
    "198.51.100.0/24": 203.0.113.1
    "2001:db8:5::/48": "fe80::1%eth0"
```

## Selective route download (SRD)

`kernel.srd-communities` limits which BGP routes are installed in the kernel: if the list is not empty, only routes
carrying one of these communities (and `kernel.statics`) are exported to the kernel. This keeps the kernel FIB small
when BIRD holds full tables.

```yaml
kernel:
  srd-communities:
    - 65530:100
```

## Default routes

With `default-route: true` (the default), Pathvector originates `0.0.0.0/0` and `::/0` reject (unreachable) routes in
the BIRD protocols `default4` and `default6`. They are used by the `announce-default` peer option and are also exported
to the kernel, so traffic to destinations without a more specific route is dropped.

To install a default route learned from an upstream instead, enable `accept-default`. This

- lets default routes pass the prefix length filter (`filter-prefix-length`) on import, and
- lowers the preference of the `default4`/`default6` routes to 1, so a learned default route becomes the best route and
  is installed in the kernel. The locally originated default route remains as a fallback if no default is learned.

```yaml
accept-default: true
peers:
  Upstream:
    asn: 64496
    neighbors:
      - 203.0.113.1
```

## Built-in protocols

Besides the kernel protocols, Pathvector always configures a BIRD `device` protocol (interface discovery) and a
`direct` protocol (routes for directly connected networks). They can be tuned with global options:

- `device-scan-time`: seconds between interface scans (`scan time`), BIRD's default if unset.
- `direct-check-link: true`: only import connected routes of interfaces whose link is up (`check link yes`).
- `kernel.scan-time`, `kernel.learn`, `kernel.reject-connected` and the other `kernel` options described above.

When these options aren't enough, list built-in protocols in `disable-protocols` (any of `device`, `direct`,
`kernel4` and `kernel6`) to leave them out of the generated configuration, and define your own in `global-config` or
a `manual*.conf` file in the BIRD directory:

```yaml
device-scan-time: 10
direct-check-link: true
disable-protocols: [ kernel4, kernel6 ]
global-config: |
  protocol kernel kernel4 { ipv4 { import all; export where source != RTS_DEVICE; }; }
  protocol kernel kernel6 { ipv6 { import all; export where source != RTS_DEVICE; }; }
```

When you replace the kernel protocols, the `kernel` options (SRD communities, statics export, source addresses, ...)
no longer apply to them.
