---
sidebar_position: 2
---

# Installation

Pathvector uses [bird2](https://gitlab.nic.cz/labs/bird/) as it's BGP daemon and supports version `bird >= 2.0.7`. Pathvector releases can be downloaded as prebuilt binaries or packages from GitHub or as deb/rpm packages from https://github.com/natesales/repo. You can also build Pathvector from source by cloning the repo and running `go generate && go build`.

It's recommended to run Pathvector every 12 hours to update IRR prefix lists and PeeringDB prefix limits. Adding `0 */12 * * * pathvector` to your crontab will update the filters at 12 AM and PM every day.

Some features require additional dependencies:

- RPKI filtering: RTR server such as [gortr](https://github.com/cloudflare/gortr) or Cloudflare's public RTR server at `rtr.rpki.cloudflare.com:8282`
- IRR prefix list generation: [bgpq4](https://github.com/bgp/bgpq4)
- VRRP daemon: [keepalived](https://github.com/acassen/keepalived)

## Linux

Pathvector can be installed on most Linux (and other UNIX-based) systems.

* Install pathvector from a [release](https://github.com/natesales/pathvector/releases) or add the [repo](https://github.com/natesales/repo) and install the `pathvector` deb/rpm package.
* Create the runtime directory `sudo mkdir -p /var/run/pathvector/cache/`
* Write your configuration to `/etc/pathvector.yml`
* Add pathvector to your crontab or other scheduler `0 */12 * * * pathvector`
* Run `pathvector` to validate your config

## Arista

Pathvector can run on Arista switches by installing a SWIX (Switch Extension) file. The SWIX bundle for each release contains Pathvector, BIRD2, GoRTR, and bgpq4.

To build the SWIX on your machine:

1. Download the latest .swix file from [releases](https://github.com/natesales/pathvector/releases), or build it manually with `cd arista && make`
2. Copy the resulting SWIX bundle extension to /mnt/flash/ on the switch

On the switch from the EOS CLI:

1. Copy the extension to extensions: `copy flash:pathvector-bundle.swix extensions:`
2. Install the extension: `extension pathvector-bundle.swix`
3. Mark the extension to be installed on boot: `copy installed-extensions boot-extensions`
4. Restart the EOS CLI: `bash sudo pkill Cli` and log back into the switch
5. Create the BIRD directory: `bash sudo mkdir /mnt/flash/bird/`
6. Write your pathvector config: `bash sudo nano /mnt/flash/pathvector.yml`
7. Run pathvector: `pathvector -c /mnt/flash/pathvector.yml -o /mnt/flash/bird/ --no-configure`
8. Restart bird: `bash sudo systemctl restart bird`
9. Add Pathvector to scheduler: `schedule test interval 720 max-log-files 1 command pathvector -c /mnt/flash/pathvector.yml -o /mnt/flash/bird/ -s /run/bird.ctl`
10. Save changes: `wr mem`

After installing the bundle extension, your switch will have a few new EOS CLI commands: `birdc`, `bgpq4`, and `pathvector`. These are just wrappers for the binaries installed on the underlying Linux system. 

## VyOS

TODO

## Juniper

TODO
