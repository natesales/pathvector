---
title: CLI
sidebar_position: 4
---
## Usage
```
Usage:
  pathvector [OPTIONS]

Application Options:
  -c, --config=                  Configuration file in YAML, TOML, or JSON
                                 format (default: /etc/pathvector.yml)
      --lock-file-directory=     Lock file directory (lockfile check disabled
                                 if empty)
  -v, --verbose                  Show verbose log messages
  -d, --dry-run                  Don't modify configuration
  -n, --no-configure             Don't configure BIRD
  -V, --version                  Show version and exit
      --bird-directory=          Directory to store BIRD configs (default:
                                 /etc/bird/)
      --bird-binary=             Path to bird binary (default: /usr/sbin/bird)
      --cache-directory=         Directory to store runtime configuration cache
                                 (default: /var/run/pathvector/cache/)
      --bird-socket=             UNIX control socket for BIRD (default:
                                 /run/bird/bird.ctl)
      --keepalived-config=       Configuration file for keepalived (default:
                                 /etc/keepalived.conf)
      --web-ui-file=             File to write web UI to (disabled if empty)
      --peeringdb-query-timeout= PeeringDB query timeout in seconds (default:
                                 10)
      --irr-query-timeout=       IRR query timeout in seconds (default: 30)
  -m, --mode=                    Should this run generate a config or start the
                                 optimization daemon? (generate or daemon)
                                 (default: generate)

Help Options:
  -h, --help                     Show this help message

```
