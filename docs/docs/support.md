---
title: Support
sidebar_position: 7
---

Best-effort community support is available via GitHub. Please [open an issue](https://github.com/pathvector/pathvector/issues/new/choose) for any questions, bug reports, or feature requests.

*Has Pathvector helped automate your network? Support the project on [GitHub Sponsors](https://github.com/sponsors/natesales). A small donation goes a long way to keep Pathvector sustainable and free for everyone.*

## Reporting Configuration

Run `pathvector config` to generate a configuration file report. Add the `--sanitize` flag to replace IP addresses and sensitive config options with innocuous placeholders.

```bash title="Example output"
$ pathvector config --sanitize
# Pathvector v6.2.0
# Built 05f3142b87ff03a3ff018d6693f0a77090167ea4 on 2023-04-03T20:29:59Z
# No plugins
# System Linux cr1.pdx4 5.10.0-20-amd64 #1 SMP Debian 5.10.158-2 (2022-12-13) x86_64 GNU/Linux
# Sanitized config exported from tests/generate-simple.yml on 03 Apr 23 02:53 -0400

asn: 34553
router-id: 10.181.208.176
source4: 10.181.208.176
source6: 2001:db8:5ded:33bc::1
<truncated>
```
