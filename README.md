# bcg

[![Go Report](https://goreportcard.com/badge/github.com/natesales/bcg?style=for-the-badge)](https://goreportcard.com/badge/github.com/natesales/bcg) 
[![License](https://img.shields.io/github/license/natesales/bcg?style=for-the-badge)](https://choosealicense.com/licenses/gpl-3.0/) 

The automatic BIRD configuration generator with bogon, IRR, RPKI, and max prefix filtering support.

#### Configuration

BCG can be configured in YAML, TOML, or JSON. All config file formats have the same configuration parameters, but follow a different capitalization structure. YAML and JSON use all lowercase parameter names and TOML uses CamelCase.

An example to configure a peer with bogon, IRR, RPKI, and max prefix filtering.
```yaml
asn: 65530
router-id: 192.0.2.1
prefixes:
  - 192.0.2.0/24
  - 2001:db8::/48

peers:
  Cloudflare:
    asn: 13335
    import: cone
    export: cone
    automaxprefix: true
    autopfxfilter: true
    neighbors:
      - 203.0.113.39
      - 2001:db8:6939::39
```

Global Configuration Options

| Key |   |   |   |   |
|-----|---|---|---|---|
|     |   |   |   |   |
|     |   |   |   |   |
|     |   |   |   |   |

Peer Configuration Options

#### Installation

Download the latest release at
