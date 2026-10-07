# Route Limits

## Peer configuration options

`import-limit4` and `import-limit6` specify how many prefixes can be accepted from the peer after filtering. They
default to 1,000,000 IPv4 and 300,000 IPv6 prefixes. These may be set automatically from the peer's PeeringDB page by
enabling `auto-import-limits`.

:::note
Pathvector 6.3.2 and earlier used a default `import-limit6` of 200,000, which is below the size of the full IPv6 table
carried by some transit providers. If you peer with a full-table IPv6 provider on those versions, set `import-limit6`
explicitly.
:::

`receive-limit4` and `receive-limit6` are like import limits but before filtering. `keep-filtered` must be enabled for
these to work.

`export-limit4` and `export-limit6` set the maximum number of prefixes to export to a peer.

## Policy violation actions

`import-limit-violation`, `receive-limit-violation`, and `export-limit-violation` control what happens when a route
limit is tripped. The default is `disable`.

`warn` logs a warning

`block` stops sending or accepting route updates after the configured number of routes have been processed

`restart` restarts the session

`disable` disables the session until it's manually enabled

See [Troubleshooting](/docs/troubleshooting#automatic-shutdown-route-limit-exceeded) if a session is shut down with `Automatic shutdown: Route limit exceeded`.
