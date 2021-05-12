# Runtime

`bcg` can take the following flags:

```
Usage:
  bcg [OPTIONS]

Application Options:
  -c, --config=            Configuration file in YAML, TOML, or JSON format (default: /etc/bcg/config.yml)
  -o, --output=            Directory to write output files to (default: /etc/bird/)
  -s, --socket=            BIRD control socket (default: /run/bird/bird.ctl)
  -k, --keepalived-config= Configuration file for keepalived (default: /etc/keepalived/keepalived.conf)
  -u, --ui-file=           File to store web UI (default: /tmp/bcg-ui.html)
  -n, --no-ui              Don't generate web UI
  -v, --verbose            Show verbose log messages
  -d, --dry-run            Don't modify BIRD config
      --no-configure       Don't configure BIRD
      --version            Show version and exit

Help Options:
  -h, --help               Show this help message
```
