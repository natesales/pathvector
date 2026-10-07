---
title: CLI Usage
sidebar_position: 6
---
# Usage
```
Pathvector is a declarative edge routing platform that automates route optimization and control plane configuration with secure and repeatable routing policy.

Usage:
  pathvector [command]

Available Commands:
  birdsh      Lightweight BIRD shell
  completion  Generate the autocompletion script for the specified shell
  config      Export configuration, optionally sanitized with logknife
  dump        Dump configuration
  generate    Generate router configuration
  help        Help about any command
  match       Find common IXPs for a given ASN
  optimizer   Start optimization daemon
  restart     Restart BIRD protocols
  status      Show protocol status
  version     Show version information
  yang        Print YANG model of the configuration

Flags:
  -c, --config string   YAML configuration file (default "/etc/pathvector.yml")
  -d, --dry-run         Don't modify configuration
  -h, --help            help for pathvector
      --lock string     Lock file (check disabled if empty)
  -n, --no-configure    Don't configure BIRD
  -t, --trace           Show trace log messages
  -v, --verbose         Show verbose log messages

Use "pathvector [command] --help" for more information about a command.
```

## Logging

`--verbose` (`-v`) enables debug log messages. `--trace` (`-t`) enables trace log messages, which include everything shown by `--verbose` plus detailed internals such as every parsed config field.

## BIRD control socket

Every command that talks to BIRD (`birdsh`/`cli`, `status`, `restart`, `version`, `config`, `generate` and `optimizer`) connects to the UNIX control socket set by the `bird-socket` config option (default `/run/bird/bird.ctl`). `birdsh` also accepts `--socket` to override it.

## version

`pathvector version` prints the Pathvector build information and the version of the running BIRD daemon. The BIRD version is taken from the greeting BIRD sends when a client connects (e.g. `BIRD 2.14 ready.`). Some BIRD builds omit the version from the greeting (`BIRD ready.`), in which case Pathvector queries `show status` instead. If the version still can't be determined, it is reported as `unknown`. Pathvector logs a warning if the running BIRD is older than the minimum supported version (2.0.7).

## status

`pathvector status` shows the state of all BIRD protocols. Protocols are listed by the peer names from your Pathvector config, using the `protocols.json` name map that `pathvector generate` writes to the `bird-directory` (default `/etc/bird/`). Use `--real-protocol-names` (`-r`) to show BIRD's protocol names instead.

## restart

`pathvector restart <name> [name...]` restarts BIRD protocols, equivalent to running `restart <protocol>` in the BIRD shell.

Each name can be either:

- a peer name as shown by `pathvector status`, which restarts every BIRD protocol of that peer (for example both the IPv4 and IPv6 sessions), or
- a BIRD protocol name as shown by `pathvector status --real-protocol-names`.

Names are passed to BIRD as protocol patterns, so shell-style wildcards such as `EXAMPLE_*` also work. Use `--all` (`-a`) to restart every BIRD protocol.

```
$ pathvector restart Example
EXAMPLE_AS65510_v4: restarted
EXAMPLE_AS65510_v6: restarted
```

The command exits with an error if any of the protocols could not be restarted (for example `No protocols match` for an unknown name).

## birdsh

`pathvector birdsh` is a lightweight BIRD shell. Run it without arguments for an interactive `bird>` prompt, or pass a command to run it once:

```
$ pathvector birdsh show protocols
```

It connects to the socket given by `--socket` (`-s`), or to the `bird-socket` config option if that flag is empty.

`pathvector cli` is an alias of `birdsh` (`cli` was the original name of this command). If a plugin that provides its own `cli` command is installed, such as the [interactive configuration CLI](/docs/interactive), the plugin's command takes precedence.
