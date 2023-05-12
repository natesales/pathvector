# Blocklists

Pathvector supports a global blocklist of prefixes and ASNs.

The blocklist can be populated with a list of ASNs and prefixes in the config file, from a text file, or fetched from a URL. See [`blocklist`](https://pathvector.io/docs/configuration#blocklist), [`blocklist-urls`](https://pathvector.io/docs/configuration#blocklist-urls), [`blocklist-files`](https://pathvector.io/docs/configuration#blocklist-files) for more information.

All peers honor the blocklist by default. This behavior can be disabled by setting [`filter-blocklist`](https://pathvector.io/docs/configuration#filter-blocklist) to `false`.
