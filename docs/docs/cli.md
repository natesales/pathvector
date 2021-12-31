---
title: CLI
sidebar_position: 4
---
## Usage
```
Pathvector is a declarative edge routing platform that automates route optimization and control plane configuration with secure and repeatable routing policy.

Usage:
  pathvector [command]

Available Commands:
  completion    generate the autocompletion script for the specified shell
  dump          Dump configuration
  generate      generate router configuration
  help          Help about any command
  match         Find common IXPs for a given ASN
  optimizer     Start optimization daemon
  portal-update Update portal status
  version       Show version information

Flags:
  -c, --config string   YAML configuration file (default "/etc/pathvector.yml")
  -d, --dry-run         Don't modify configuration
  -h, --help            help for pathvector
      --lock string     Lock file (check disabled if empty)
  -n, --no-configure    Don't configure BIRD
  -v, --verbose         Show verbose log messages

Use "pathvector [command] --help" for more information about a command.
```
