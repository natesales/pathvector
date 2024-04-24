---
title: Snippets
sidebar_position: 5
---

The `global-config` option can be used to specify a block of global BIRD configuration. Additionally, Pathvector will
load files that match `manual*.conf` in the BIRD configuration directory.

The peer level `pre-import`, `pre-export`, `pre-import-final`, and `pre-export-final` options can be used to specify a
block of BIRD configuration that will be inserted before or after the import or export filters. Each option can be
suffixed with `-file` to indicate a config file that the configuration snippet will be read from. All peer level options
support string replacements for peer level configuration options keyed by `<pathvector.$OPTION>`. For example,
`pre-import: 'print "Hello from AS<pathvector.asn>";'` will print "Hello from AS65530" where 65530 is
the peer's ASN.
