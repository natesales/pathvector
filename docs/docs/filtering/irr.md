# IRR

IRR filtering uses [bgpq4](https://github.com/bgp/bgpq4) to generate sets of prefixes and ASNs.

## Global configuration

`irr-server` sets the IRR server address

`bgpq-args` adds additional arguments to pass to `bgpq4` (for example to limit IRR sources with `-S RIPE`)

## Peer configuration

Enable `filter-irr` to enable IRR filtering.

Enable `filter-as-members` to reject routes that aren't originated from an ASN within the peer's `as-members` list.
Enable `auto-as-set-members` to retrieve that list automatically from their PeeringDB IRR object.
