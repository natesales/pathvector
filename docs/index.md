<!-- This empty header is there to make mkdocs remove the title -->
#

![Banner](assets/banner.png)

Pathvector is a declarative routing control-plane platform for BGP with robust filtering and route optimization.

[![Go Report](https://goreportcard.com/badge/github.com/natesales/wireframe?style=for-the-badge)](https://goreportcard.com/report/github.com/natesales/wireframe)
[![License](https://img.shields.io/github/license/natesales/wireframe?style=for-the-badge)](https://github.com/natesales/wireframe/blob/main/LICENSE)
[![Release](https://img.shields.io/github/v/release/natesales/wireframe?style=for-the-badge)](https://github.com/natesales/wireframe/releases)

## Features

* Robust BGP route filtering with RPKI, IRR, and import limits automatically configured from PeeringDB.
* Automatic route optimization by enriching the standard set of BGP attributes with latency and packet loss metrics. 
* Single YAML configuration file: Want to track your changes? Just commit your file to version control.
* Data-plane Agnostic: Pathvector works on servers, network switches, embedded devices, etc.
* Built on Open Source: In addition to Pathvector itself, it's dependencies such as [bird](https://gitlab.nic.cz/labs/bird/), [keepalived](https://github.com/acassen/keepalived), [gortr](https://github.com/cloudflare/gortr) and [bgpq4](https://github.com/bgp/bgpq4) are open source and free to use.
