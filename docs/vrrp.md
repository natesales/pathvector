# VRRP (Virtual Router Redundancy Protocol)

Wireframe can build [keepalived](https://github.com/acassen/keepalived) configs for VRRP. To enable VRRP, add a `vrrp` config key containing a list of VRRP instances to your bcg config file.

| Option      | Usage                                                                          |
| ----------- | ------------------------------------------------------------------------------ |
| state       | VRRP state (`primary` or `backup`)                                             |
| interface   | Interface to run VRRP on                                                       |
| vrrid       | VRRP Router ID (must be the same for multiple routers in the same VRRP domain  |
| priority    | VRRP router selection priority                                                 |
| vips        | List of Virtual IPs                                                            |
