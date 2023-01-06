---
title: About
sidebar_position: 1
---

<img alt="Pathvector Logo" src="https://pathvector.io/img/black-border.svg" height="200" />

Pathvector is a declarative edge routing platform that automates route optimization and control plane configuration with
secure and repeatable routing policy.

[![Docs](https://img.shields.io/static/v1?label=docs&message=pathvector.io&color=9407cd&style=for-the-badge)](https://pathvector.io)
[![Go Report](https://goreportcard.com/badge/github.com/natesales/pathvector?style=for-the-badge)](https://goreportcard.com/report/github.com/natesales/pathvector)
[![CII Best Practices](https://img.shields.io/static/v1?label=CII%20Best%20Practices&message=passing&color=green&style=for-the-badge)](https://bestpractices.coreinfrastructure.org/projects/5328)

Check out the [installation page](https://pathvector.io/docs/installation),
[practical examples](https://pathvector.io/docs/examples), and
[configuration manual](https://pathvector.io/docs/configuration).

## Features

* Robust BGP route filtering with RPKI, IRR, and downstream AS cone, ASPA, never-via-RS and more
* Automatic configuration from PeeringDB
* Automatic route optimization by enriching the standard set of BGP attributes with latency and packet loss metrics
* Declarative configuration model: Want to track your changes? Just commit your file to version control.
* Data-plane agnostic: Pathvector works on servers, network switches, embedded devices, etc
* BFD and VRRP support
* Extensible Go plugin API
