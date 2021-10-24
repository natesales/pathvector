![Banner](https://pathvector.io/img/black-border.svg)

Pathvector is a declarative edge routing platform that automates route optimization and control plane configuration with secure and repeatable routing policy.

[![Docs](https://img.shields.io/static/v1?label=docs&message=pathvector.io&color=9407cd&style=for-the-badge)](https://pathvector.io)
[![Go Report](https://goreportcard.com/badge/github.com/natesales/pathvector?style=for-the-badge)](https://goreportcard.com/report/github.com/natesales/pathvector)
[![Codecov](https://img.shields.io/codecov/c/github/natesales/pathvector?style=for-the-badge)](https://app.codecov.io/gh/natesales/pathvector)
[![CII Best Practices](https://img.shields.io/static/v1?label=CII%20Best%20Practices&message=passing&color=green&style=for-the-badge)](https://bestpractices.coreinfrastructure.org/projects/5328)

## Features

* Robust BGP route filtering with RPKI, IRR, and import limits automatically configured from PeeringDB.
* Automatic route optimization by enriching the standard set of BGP attributes with latency and packet loss metrics.
* Declarative configuration model: Want to track your changes? Just commit your file to version control.
* Data-plane Agnostic: Pathvector works on servers, network switches, embedded devices, etc.
