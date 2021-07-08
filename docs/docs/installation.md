---
sidebar_position: 2
---

# Installation

All versions of Pathvector for all supported platforms are available to download from the latest [release page](https://github.com/natesales/pathvector/releases).

It is recommended to run Pathvector every 12 hours to update IRR prefix lists and PeeringDB prefix limits. Adding `0 */12 * * * pathvector` to your crontab will update the filters at 12 AM and PM every day.

The only required dependency is `bird >= 2.0.7`, but some features require additional dependencies:

- RPKI filtering: RTR server such as [gortr](https://github.com/cloudflare/gortr) or Cloudflare's public RTR server at `rtr.rpki.cloudflare.com:8282`
- IRR prefix list generation: [bgpq4](https://github.com/bgp/bgpq4)
- VRRP daemon: [keepalived](https://github.com/acassen/keepalived)

## Linux

Pathvector on Linux is available for amd64, aarch64, and mips64 as binaries and deb and rpm packages from [releases](https://github.com/natesales/pathvector/releases).

Packages are also available in the [natesales](https://github.com/natesales/repo) APT/YUM repo.

## FreeBSD

Pathvector is available as an amd64/aarch64 binary for FreeBSD from the [releases](https://github.com/natesales/pathvector/releases) page.

## VyOS

VyOS is based on Debian, see [Linux](#linux) for more information.

## Ubiquiti EdgeOS

Ubiquiti EdgeRouters are based on Debian. Pathvector supports the ER-8-XG, ER-12P, ER-12, ERPro-8, ER-8, ER-6P, ERPoe-5, ER-4, and ERLite-3 routers based on the MIPS64 architecture. See [Linux](#linux) for more information.

Not Supported (MIPS1004Kc): ER-X, ER-X-SFP, ER-10X

## Arista EOS

Pathvector can run on Arista switches by installing a SWIX (Switch Extension) file from [releases](https://github.com/natesales/pathvector/releases). First, copy the `pathvector-*.swix` file to `/mnt/flash/` on your switch.

From the EOS CLI:

1. Copy the extension to extensions: `copy flash:pathvector-*.swix extensions:`
2. Install the extension: `extension pathvector-bundle.swix`
3. Mark the extension to be installed on boot: `copy installed-extensions boot-extensions`
4. Add the pathvector alias: `alias pathvector "bash pathvector"`
5. Add Pathvector to scheduler (optional): `schedule pathvector interval 720 max-log-files 1 command pathvector -c /mnt/flash/pathvector.yml -o /mnt/flash/bird/ -s /run/bird.ctl`
6. Save changes: `wr mem`

## Building from source

Pathvector can be easily built from source for some, but not all, of the [many supported Go platforms](https://github.com/golang/go/blob/master/src/go/build/syslist.go).

For example, to build Pathvector for M1 Macs:

```bash
git clone https://github.com/natesales/pathvector
GOOS=darwin GOARCH=arm64 go build
```
