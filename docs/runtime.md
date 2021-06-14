<!-- Code generated DO NOT EDIT -->
## CLI Flags
| Option | Type | Default | Description |
|--------|------|---------|-------------|
| -c,  --config | string | /etc/wireframe.yml | Configuration file in YAML, TOML, or JSON format |
|  --lock-file | string |  | Lock file (check disabled if empty) |
| -v,  --verbose | bool |  | Show verbose log messages |
| -d,  --dry-run | bool |  | Don't modify configuration |
| -n,  --no-configure | bool |  | Don't configure BIRD |
| -V,  --version | bool |  | Show version and exit |
|  --bird-directory | string | /etc/bird/ | Directory to store BIRD configs |
|  --bird-binary | string | /usr/sbin/bird | Path to bird binary |
|  --cache-directory | string | /var/run/wireframe/cache/ | Directory to store runtime configuration cache |
|  --bird-socket | string | /run/bird/bird.ctl | UNIX control socket for BIRD |
|  --keepalived-config | string | /etc/keepalived.conf | Configuration file for keepalived |
|  --web-ui-file | string |  | File to write web UI to (disabled if empty) |
|  --peeringdb-query-timeout | uint | 10 | PeeringDB query timeout in seconds |
|  --irr-query-timeout | uint | 30 | IRR query timeout in seconds |

# Usage
```
Usage:
  wireframe [OPTIONS]

Application Options:
  -c, --config=                  Configuration file in YAML, TOML, or JSON
                                 format (default: /etc/wireframe.yml)
      --lock-file=               Lock file (check disabled if empty)
  -v, --verbose                  Show verbose log messages
  -d, --dry-run                  Don't modify configuration
  -n, --no-configure             Don't configure BIRD
  -V, --version                  Show version and exit
      --bird-directory=          Directory to store BIRD configs (default:
                                 /etc/bird/)
      --bird-binary=             Path to bird binary (default: /usr/sbin/bird)
      --cache-directory=         Directory to store runtime configuration cache
                                 (default: /var/run/wireframe/cache/)
      --bird-socket=             UNIX control socket for BIRD (default:
                                 /run/bird/bird.ctl)
      --keepalived-config=       Configuration file for keepalived (default:
                                 /etc/keepalived.conf)
      --web-ui-file=             File to write web UI to (disabled if empty)
      --peeringdb-query-timeout= PeeringDB query timeout in seconds (default:
                                 10)
      --irr-query-timeout=       IRR query timeout in seconds (default: 30)

Help Options:
  -h, --help                     Show this help message

```
