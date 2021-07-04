# Monitoring

Pathvector exposes a Prometheus metrics API for monitoring. This is completely separate from the [control API](/api) and only provides the single `/metrics` route. The default port is [`9785`](https://github.com/prometheus/prometheus/wiki/Default-port-allocations) and can be configured with the `--metrics` flag. For more information, see the [runtime configuration](/configuration/runtime) page.
