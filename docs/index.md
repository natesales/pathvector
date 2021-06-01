# Welcome to Wireframe

Wireframe is a declarative routing platform integrating BGP with robust filtering support, an XDP dataplane, and VRRP for high availability. It's best used in the core and peering edge, but is flexible enough to adapt to a multitude of network architectures.

[![Go Report](https://goreportcard.com/badge/github.com/natesales/wireframe?style=for-the-badge)](https://goreportcard.com/report/github.com/natesales/wireframe)
[![License](https://img.shields.io/github/license/natesales/wireframe?style=for-the-badge)](https://github.com/natesales/wireframe/blob/main/LICENSE)
[![Release](https://img.shields.io/github/v/release/natesales/wireframe?style=for-the-badge)](https://github.com/natesales/wireframe/releases)

## Overview

* Single configuration file (YAML, JSON, or TOML): Want to track your changes? Just commit your file to version control.
* Platform agnostic: Wireframe works on servers, switches, SBCs, etc.
* Free and Open Source: In addition to Wireframe itself, it's dependencies such as [bird](https://gitlab.nic.cz/labs/bird/), [xdprtr](https://github.com/natesales/xdprtr), [keepalived](https://github.com/acassen/keepalived), [gortr](https://github.com/cloudflare/gortr) and [bgpq4](https://github.com/bgp/bgpq4) are open source and free to use.

## Installation

Wireframe depends on [bird2](https://gitlab.nic.cz/labs/bird/). Make sure the `bird` and `gortr` daemons are running and `bgpq4` is in path before running Wireframe. Releases can be downloaded from GitHub and from my public code repositories - see https://github.com/natesales/repo for more info. You can also build from source by cloning the repo and running `go generate && go build`. It's recommended to run Wireframe every 12 hours to update IRR prefix lists and PeeringDB prefix limits. Adding `0 */12 * * * /usr/local/bin/wireframe` to your crontab will update the filters at 12 AM and PM. If you're using ZSH you might also be interested in my [birdc completion](https://github.com/natesales/zsh-bird-completions).

Some features require additional dependencies:

- [gortr](https://github.com/cloudflare/gortr)
- [bgpq4](https://github.com/bgp/bgpq4)
- [keepalived](https://github.com/acassen/keepalived)
- [xdprtr](https://github.com/natesales/xdprtr)

## Quick Example

Here's an example of a core router with BGP filtering by RPKI, IRR, and prefix limits, paired with VRRP for HA and XDP for fast packet forwarding: all in less than 30 lines.

```yaml
asn: 65530
router-id: 192.0.2.1
pref-src4: 192.0.2.1
pref-src6: 2001:db8::1
prefixes:
  - 192.0.2.0/24
  - 2001:db8::/48

interfaces:
  eth0:
    xdprtr: true

vrrp:
  - interface: eth0
    state: primary
    vrrid: 1
    priority: 255
    vips:
      - 192.0.2.1/24
      - 2001:db8::1/48

peers:
  Nate Sales:
    asn: 34553
    type: peer
    neighbors:
      - "203.0.113.34"
      - "2001:db8::34"
```
