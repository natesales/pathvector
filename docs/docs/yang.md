---
title: YANG Model
sidebar_position: 8
---

# YANG Model

Pathvector publishes a [YANG 1.1](https://www.rfc-editor.org/rfc/rfc7950) data model of its configuration, so the configuration can be described, validated and exchanged with standards based tooling.

| Module       | Namespace               | Prefix |
|--------------|-------------------------|--------|
| `pathvector` | `urn:pathvector:config` | `pv`   |

The module is generated directly from Pathvector's configuration structs, so it always matches the options documented on the [configuration](configuration) page. Print it for your installed version with:

```bash
pathvector yang > pathvector.yang
```

A copy of the module generated from the latest source is also published at [/pathvector.yang](pathname:///pathvector.yang), and can be checked with standard tools such as `pyang --strict pathvector.yang`.

## Mapping from the YAML configuration

| YAML configuration                                              | YANG model                                            |
|-----------------------------------------------------------------|-------------------------------------------------------|
| Boolean, string, and integer options                            | `leaf` of type `boolean`, `string`, `int64`/`uint64`/`uint32` |
| Decimal options (e.g. `packet-loss-threshold`)                  | `leaf` of type `decimal64` (4 fraction digits)        |
| Lists (e.g. `prefixes`, `neighbors`)                            | `leaf-list` (`ordered-by user`)                       |
| Named sections (`peers`, `templates`, `vrrp`, `bfd`, `mrt`)     | `list` keyed by `name`                                |
| Other maps (e.g. `kernel.statics`, `as-prefs`, `plugins`)       | `list` keyed by `key`, with a `value` leaf/leaf-list  |
| Nested sections (`kernel`, `optimizer`)                         | `container`                                           |

Option descriptions and default values are carried over as `description` and `default` statements.

## JSON configuration

Pathvector accepts JSON configuration files in two forms:

1. **Plain JSON** using the same structure as the YAML configuration (YAML is a superset of JSON):

```json
{
  "asn": 65530,
  "router-id": "192.0.2.1",
  "peers": {
    "Example": {"asn": 65510, "neighbors": ["203.0.113.25"]}
  }
}
```

2. **[RFC 7951](https://www.rfc-editor.org/rfc/rfc7951) JSON** instance data of the YANG model, as produced by NETCONF/RESTCONF tooling. A document is treated as RFC 7951 JSON when its top-level members are qualified with the `pathvector:` module name. Lists are converted back to the YAML map structure and string encoded 64-bit numbers are accepted:

```json
{
  "pathvector:asn": "65530",
  "pathvector:router-id": "192.0.2.1",
  "pathvector:peers": [
    {"name": "Example", "asn": "65510", "neighbors": ["203.0.113.25"]}
  ]
}
```

## Limitations and future work

- Validation rules beyond basic types (such as IP address and prefix formats) are enforced by Pathvector when loading the configuration and are not yet expressed in the YANG model.
- YANG 1.1 requires configuration `leaf-list` values to be unique, while some Pathvector lists (such as `prepend-path`) may intentionally contain repeated values.
- Serving the model over NETCONF or RESTCONF (a management server that edits and applies Pathvector configuration) is future work.
