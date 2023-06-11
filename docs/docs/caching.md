---
title: Caching
sidebar_position: 9
---

Pathvector relies on external datasources to generate configuration, such as PeeringDB, IRR databases, and the RPKI. There are various mechanisms to cache this data to decrease latency and reduce load on these external services.

## RPKI

Networks should already be running their own RTR (RPKI to Router) server such as [stayrtr](https://github.com/bgp/stayrtr) or [rtrtr](https://github.com/NLnetLabs/rtrtr).

## IRR

## PeeringDB

Pathvector has an internal PeeringDB cache that stores PeeringDB objects *for the duration of a single `pathvector generate` run*. This does not cache for longer than a single command invocation.

### PeeringDB Local Cache

To cache PeeringDB data persistently, you can set the global [`peeringdb-url`](https://pathvector.io/docs/configuration/#peeringdb-url) option to a local [PeeringDB cache server](https://github.com/natesales/peeringdb-cache).
