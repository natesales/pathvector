# Peering Portal

[![Docs](https://img.shields.io/static/v1?label=ABOUT&message=pathvector.io&color=9407cd&style=for-the-badge)](https://pathvector.io)
[![Docker](https://img.shields.io/docker/v/natesales/pathvector-portal?style=for-the-badge)](https://hub.docker.com/r/natesales/pathvector-portal)
[![Go Report](https://goreportcard.com/badge/github.com/natesales/pathvector-portal?style=for-the-badge)](https://goreportcard.com/report/github.com/natesales/pathvector-portal)
[![License](https://img.shields.io/github/license/natesales/pathvector-portal?style=for-the-badge)](https://github.com/natesales/pathvector-portal/blob/main/LICENSE)

The Pathvector Peering Portal is a web interface for multi-router peering session management. Peers can log in with PeeringDB OAuth to see the status of their current sessions and configure new sessions at common internet exchanges.

This project is part of Pathvector and works best on Pathvector routers. The API is generic enough though, that it's possible to integrate other vendors the portal.

## Setup

### PeeringDB OAuth

Create a new PeeringDB OAuth application at https://peeringdb.com/oauth2/applications/ with client type `Confidential` and authorization grant type `Authorization code`. The redirect URI is in the format of `https://peering.example.com/auth/redirect`

### Peering Portal

The easiest way to deploy the peering portal is with docker-compose. An example compose file is available [here](https://github.com/natesales/pathvector-portal/blob/main/docker-compose.yml). You will also need a reverse proxy in front of the portal container to forward requests from the internet to http://peering-portal:8080 on the container.

### Routers

Once you have the portal running, just add the `portal-host`, `portal-key`, and optionally `hostname` fields to your Pathvector config on each router that you want to use. Pathvector will then push BGP session status to the portal server on each run.
