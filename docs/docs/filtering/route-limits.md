# Route Limits

## Peer configuration options

`import-limit4` and `import-limit6` specify many prefixes can be accepted from the peer after filtering. These may be
set automatically from the peer's PeeringDB page by enabling `auto-import-limits`.

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
