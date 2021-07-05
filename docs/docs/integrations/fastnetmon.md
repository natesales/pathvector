# DDoS mitigation with FastNetMon

[FastNetMon](https://fastnetmon.com) is a software-based DoS/DDoS detection tool that analyzes flows and signals a mitigation action over BGP such as traffic redirection or selective destination blackholes.

To get started with Pathvector and FastNetMon, first install [Pathvector](/installation) and [FastNetMon Advanced](https://fastnetmon.com/docs-fnm-advanced/advanced-install-guide/).

## Configure FastNetMon

From the `fcli` prompt:

```shell
set main mirror_afpacket enable
set main interfaces bond0
set main process_ipv6_traffic enable
set main networks_list 198.51.100.0/24
set main networks_list 2001:db8::/48

set main email_notifications_enabled enable
set main email_notifications_tls enable
set main email_notifications_auth enable
set main email_notifications_port 587
set main email_notifications_host mail.example.com
set main email_notifications_from fnm@example.com
set main email_notifications_username fnm@example.com
set main email_notifications_password examplepassword
set main email_notifications_recipients noc@example.com

set hostgroup global threshold_mbps 900
set hostgroup global ban_for_bandwidth enable
set hostgroup global threshold_tcp_syn_mbps 10
set hostgroup global ban_for_tcp_syn_bandwidth enable
set hostgroup global threshold_icmp_mbps 10
set hostgroup global ban_for_icmp_bandwidth enable

set hostgroup global enable_ban enable
set main enable_ban_ipv6 enable
set main enable_ban enable
set main unban_enabled enable
set main ban_time 300  # Unban after 5 minutes

set main gobgp enable
set main gobgp_ipv6 enable
set main gobgp_announce_host enable
set main gobgp_announce_host_ipv6 enable
set main gobgp_next_hop 192.0.2.1
set main gobgp_next_hop_ipv6 100::1
set main gobgp_bgp_listen_port 1179

set bgp pathvector
set bgp pathvector local_asn 65530
set bgp pathvector remote_asn 65530
set bgp pathvector local_address 127.0.0.2
set bgp pathvector remote_address 127.0.0.1
set bgp pathvector ipv4_unicast enable
set bgp pathvector ipv6_unicast enable
set bgp pathvector active enable

commit
```

## Configure Pathvector

Pathvector config needs a few extra options for the FastNetMon session:

```yaml
FastNetMon:
  asn: 65530
  local-asn: 65530  # In this example the ASN and local ASN are set explicitly for iBGP
  listen: 127.0.0.1
  neighbors: [ "127.0.0.2" ]
  filter-rpki: false  # ROAs may have a maxLength that would cause the routes to be filtered
  filter-prefix-length: false  # Disable prefix length filter so /32 and /128 routes will be accepted
  enforce-first-as: false  # We don't care about the first AS in path
  enforce-peer-nexthop: false  # Peer nexthops will be set to blackhole addresses, not the BGP peer address
  announce-originated: false  # No need to announce anything to FNM. While it does support learning routes over BGP, the implementation requires a cronjob to run the updates: https://fastnetmon.com/docs-fnm-advanced/subnet-collection-from-bgp-peering-session/
  neighbor-port: 1179  # The default BGP port will conflict, so we'll use a different one for FastNetMon
  import-communities: [ "65530,666" ]  # More communities can be added here for other peers, or added on a per-peer basis
  mp-unicast-46: true  # FastNetMon will announce both IPv4 and IPv6 routes over this multiprotocol session
```
