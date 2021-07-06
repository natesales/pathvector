---
title: CLI
sidebar_position: 4
---
## Usage
```
Pathvector is a declarative routing platform for BGP which automates route optimization and control plane configuration with secure and repeatable routing policies.

Usage:
  pathvector [flags]
  pathvector [command]

Available Commands:
  completion  generate the autocompletion script for the specified shell
  help        Help about any command
  optimizer   Start optimization daemon
  version     Show version information

Flags:
      --bird-binary string             Path to bird binary (default "/usr/sbin/bird")
      --bird-directory string          Directory to store BIRD configs (default "/etc/bird/")
      --bird-socket string             UNIX control socket for BIRD (default "/run/bird/bird.ctl")
      --cache-directory string         Directory to store runtime configuration cache (default "/var/run/pathvector/cache/")
  -c, --config string                  Configuration file in YAML, TOML, or JSON format (default "/etc/pathvector.yml")
  -d, --dry-run                        Don't modify configuration
  -h, --help                           help for pathvector
      --irr-query-timeout uint         IRR query timeout in seconds (default 30)
      --keepalived-config string       Configuration file for keepalived (default "/etc/keepalived.conf")
      --lock-file-directory string     Lock file directory (lockfile check disabled if empty
  -n, --no-configure                   Don't configure BIRD
      --peeringdb-query-timeout uint   PeeringDB query timeout in seconds (default 10)
  -v, --verbose                        Show verbose log messages
      --web-ui-file string             File to write web UI to (disabled if empty)

Use "pathvector [command] --help" for more information about a command.
```
