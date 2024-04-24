---
title: Interactive CLI
sidebar_position: 7
---

Pathvector supports an interactive CLI for configuration.

```
$ pathvector cli
pathvector [empty] > enable
pathvector [empty] # init AS65530 192.0.2.1
Are you sure you want to create a new config with AS65530 (192.0.2.1)? [y/N] y
Config created
pathvector (altair) # set prefixes 192.0.2.0/24,2001:db8::/48
pathvector (altair) # create peers Example
pathvector (altair) # set peers Example asn 65510
pathvector (altair) # set peers Example neighbors 192.0.2.1,2001:db8::1
pathvector (altair) # commit
Persistent configuration updated
pathvector (altair) # run
Starting Pathvector...<truncated>
```
