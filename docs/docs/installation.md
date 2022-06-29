---
sidebar_position: 2
---

# Installation

<!--
import {DownloadForm} from "../src/components/DownloadForm";

<DownloadForm/>
-->

All versions of Pathvector for all supported platforms are available to download from the
latest [release page](https://github.com/natesales/pathvector/releases).

It is recommended to run Pathvector every 12 hours to update IRR prefix lists and PeeringDB prefix limits.
Adding `0 */12 * * * pathvector generate` to your crontab will update the filters at 12 AM and PM every day.

The only required dependency is `bird >= 2.0.7`, but some features require additional dependencies:

- RPKI filtering: RTR server such as [gortr](https://github.com/cloudflare/gortr) or Cloudflare's public RTR server at `rtr.rpki.cloudflare.com:8282`
- IRR prefix list generation: [bgpq4](https://github.com/bgp/bgpq4)
- VRRP daemon: [keepalived](https://github.com/acassen/keepalived)

## Package Repository

Pathvector releases >= 5.1.2 are available in the https://repo.pathvector.io package repository. Packages will still
continue to be uploaded to the [natesales repo](https://github.com/natesales/repo) for compatibility with existing
installs, but for security it is recommended to use the repo.pathvector.io for all new installations due to increased
security by GPG signatures. Packages in repo.pathvector.io are signed
with [`0983 AC66 7B4F 0B54 F69D`](https://repo.pathvector.io/pgp.asc). Note that packages downloaded from GitHub
releases are not signed.

Pathvector on Linux is available for amd64, aarch64, and mips64 as binaries and deb and rpm packages
from [releases](https://github.com/natesales/pathvector/releases).

## Debian

```shell
curl https://repo.pathvector.io/pgp.asc > /usr/share/keyrings/pathvector.asc
echo "deb [signed-by=/usr/share/keyrings/pathvector.asc] https://repo.pathvector.io/apt/ stable main" > /etc/apt/sources.list.d/pathvector.list
apt update && apt install -y pathvector
```

## CentOS

```shell
yum install -y yum-utils
yum-config-manager --add-repo https://repo.pathvector.io/yum/pathvector.repo
yum install pathvector
```

## VyOS

VyOS is based on Debian, see [Debian](#debian) for more information.

## TNSR

TNSR is based on CentOS, see [CentOS](#centos) for more information.

## FreeBSD

Pathvector is available as an amd64/aarch64 binary for FreeBSD from
the [releases](https://github.com/natesales/pathvector/releases) page.

## Ubiquiti EdgeOS

Ubiquiti EdgeRouters are based on Debian. Pathvector supports the ER-8-XG, ER-12P, ER-12, ERPro-8, ER-8, ER-6P, ERPoe-5,
ER-4, and ERLite-3 routers based on the MIPS64 architecture. See [Debian](#debian) for more information.

Not Supported (MIPS1004Kc): ER-X, ER-X-SFP, ER-10X

## Arista EOS

Pathvector can run on Arista switches by installing a SWIX (Switch Extension) file
from [releases](https://github.com/natesales/pathvector/releases). First, copy the `pathvector-arista.swix` file
to `/mnt/flash/` on your switch.

From the EOS CLI:

1. Copy the extension to extensions: `copy flash:pathvector-arista.swix extensions:`
2. Install the extension: `extension pathvector-arista.swix`
3. Mark the extension to be installed on boot: `copy installed-extensions boot-extensions`
4. Add the pathvector alias: `alias pathvector "bash pathvector"`
5. Add Pathvector to scheduler (
   optional): `schedule pathvector interval 720 max-log-files 1 command pathvector -c /mnt/flash/pathvector.yml -o /mnt/flash/bird/ -s /run/bird.ctl`
6. Save changes: `wr mem`

## Juniper JunOS

Pathvector can run on Juniper devices by installing a signed JunOS extension package. First, download the Pathvector
Juniper release, either to your local machine and SCP it over to the Juniper router/switch, or pull it directly in
the `request system software` command by replacing the filename with the URL to the file.

Add the Pathvector extension provider:
`set system extensions providers pathvector license-type customer deployment-scope commercial`

Install the Pathvector package:
`request system software add pathvector-juniper.tgz` or `request vmhost software add pathvector-juniper.tgz`

## Cisco IOx

Pathvector can run on IOx compatible Cisco devices by installing a IOx package release, or directly to the device with
ioxclient. See https://developer.cisco.com/docs/iox/#!app-management/application-management for more information.

## Nokia Service Router (SR) Linux

Nokia SR Linux is based on CentOS, see [CentOS](#centos) for more information.

## Mikrotik RouterOS

Pathvector has experimental RouterOS support. Mikrotik has discontinued this feature. See the [RouterOS Container](https://help.mikrotik.com/docs/display/ROS/Container) reference for installation instructions
for the container package.

To build a Docker image for an alternate architecture:

```bash
docker build --output type=tar,dest=pathvector-mikrotik-arm64v8.tar -t pathvector-cron:arm64v8 --build-arg ARCH=arm64v8 -f ../vendorbuild/mikrotik/Dockerfile ..
```

## Building from source

Pathvector can be easily built from source for some, but not all, of
the [many supported Go platforms](https://github.com/golang/go/blob/master/src/go/build/syslist.go).

For example, to build Pathvector for M1 Macs:

```bash
git clone https://github.com/natesales/pathvector && cd pathvector
GOOS=darwin GOARCH=arm64 go build
```
