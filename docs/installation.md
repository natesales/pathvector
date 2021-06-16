# Installation

Pathvector uses [bird2](https://gitlab.nic.cz/labs/bird/) as it's BGP daemon and supports version `bird >= 2.0.7`. Pathvector releases can be downloaded as prebuilt binaries or packages from GitHub or as deb/rpm packages from https://github.com/natesales/repo. You can also build Pathvector from source by cloning the repo and running `go generate && go build`.

It's recommended to run Pathvector every 12 hours to update IRR prefix lists and PeeringDB prefix limits. Adding `0 */12 * * * pathvector` to your crontab will update the filters at 12 AM and PM every day.

Some features require additional dependencies:

- RPKI filtering: RTR server such as [gortr](https://github.com/cloudflare/gortr) or Cloudflare's public RTR server at `rtr.rpki.cloudflare.com:8282`
- IRR prefix list generation: [bgpq4](https://github.com/bgp/bgpq4)
- VRRP daemon: [keepalived](https://github.com/acassen/keepalived)
