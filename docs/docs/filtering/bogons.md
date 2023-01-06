# Bogons

Bogons (AKA martians) are routes and ASNs that shouldn't be visible on the Internet.

`filter-bogon-asns` rejects routes with a bogon ASN in path.

`filter-bogon-routes` rejects bogon routes.

Pathvector comes preloaded with a [default set of bogons](https://github.com/natesales/pathvector/blob/main/pkg/config/config.go) which can be overridden with the [`bogon-asns`](/docs/configuration#bogon-asns), [`bogons4`](/docs/configuration#bogons4), and [`bogons4`](/docs/configuration#bogons6) global config options.
