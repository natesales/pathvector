# Team Cymru UTRS

The [Team Cymru Unwanted Traffic Removal Service (UTRS)](https://www.team-cymru.com/ddos-mitigation-services) is a free, community-run remote triggered blackhole (RTBH) service. Participants announce /32 and /128 blackhole routes for their own address space to UTRS, which redistributes them to every other participant so the attack traffic can be dropped closer to its source.

This guide is based on Team Cymru's [UTRS Peering Guide](https://github.com/team-cymru/network-security-templates/blob/master/UTRS-Peering-Guide/README.md) (BIRD 2 section). Always check that guide and your UTRS provisioning details for the current requirements.

## UTRS session requirements

- UTRS uses ASN **64496**.
- There are two UTRS route servers (an active/active pair). You'll receive their addresses and a TCP MD5 password when your participation is provisioned. Peer with both.
- Sessions are **eBGP multihop** with a **TCP MD5 password**. UTRS initiates the connection, so your side should be **passive**.
- IPv4 and IPv6 routes are exchanged over the same (IPv4) session.
- You may announce IPv4 prefixes from /25 to /32 and IPv6 prefixes from /49 to /128, and only for address space you originate or are the upstream for.
- UTRS accepts up to 25 active announcements per participant. Routes expire after 7 days unless they're re-announced.
- Routes received from UTRS carry the `64496:0` and `NO_EXPORT` communities and should be sent to a null/discard next hop.

UTRS also distributes FlowSpec rules. Pathvector doesn't generate FlowSpec channels, so this guide only covers RTBH (unicast) routes.

## Tagging routes to send to UTRS

Pathvector exports routes to UTRS based on a BGP community. In this guide, any route carrying the `65530:666` community will be sent to UTRS. Set it on routes learned from your DDoS detection system, for example with `add-on-import: [ "65530:666" ]` on the session from [FastNetMon](/docs/integrations/fastnetmon), or with your own `pre-import-accept` logic on an internal session.

## Configure Pathvector

The example below uses `65530` as the local ASN, `198.51.100.0/24` and `2001:db8::/48` as the local prefixes, and `203.0.113.10` and `203.0.113.11` as the UTRS route servers. Replace them with your own values.

```yaml
asn: 65530
prefixes:
  - 198.51.100.0/24
  - 2001:db8::/48

peers:
  UTRS:
    asn: 64496
    neighbors:
      - 203.0.113.10
      - 203.0.113.11
    description: Team Cymru UTRS
    password: "provided-by-team-cymru"
    multihop: true  # UTRS route servers are not directly connected
    passive: true  # UTRS initiates the session
    mp-unicast-46: true  # IPv4 and IPv6 routes over the same session

    # Import: accept blackhole routes from UTRS and null route them
    filter-prefix-length: false  # Allow /25-/32 and /49-/128 routes
    filter-rpki: false  # Blackhole routes are often more specific than the ROA maxLength
    enforce-first-as: false  # UTRS is a route server and doesn't prepend its ASN
    enforce-peer-nexthop: false  # Next hops are blackhole addresses, not the route server address
    blackhole-in: true  # Rewrite the next hop to the local null route (192.0.2.1 / 100::1)
    pre-import-filter: |
      if !((64496, 0) ~ bgp_community) then reject;
      if (net.type = NET_IP4 && net.len < 25) then reject;
      if (net.type = NET_IP6 && net.len < 49) then reject;
    import-limit4: 10000
    import-limit6: 10000
    import-limit-violation: block  # Keep the session up if the limit is reached

    # Export: only announce tagged host routes within our own address space
    announce-originated: false  # Don't send our aggregates to UTRS
    announce: [ "65530:666" ]  # Send routes with our blackhole trigger community
    only-announce:  # Only host routes within our own address space
      - 198.51.100.0/24{32,32}
      - 2001:db8::/48{128,128}
    export-limit4: 25
    export-limit6: 25
    export-limit-violation: block  # UTRS allows 25 active announcements per participant
```

Blackhole routes learned from UTRS are imported with their next hop set to `192.0.2.1` (IPv4) or `100::1` (IPv6), which Pathvector routes to a blackhole, so they're dropped once exported to the kernel. They keep the `NO_EXPORT` community, and since none of your other peers announce routes tagged `64496:0`, they won't be re-advertised.

If you'd rather announce /25-/31 or /49-/127 prefixes as well, change the length ranges in `only-announce`, for example `198.51.100.0/24{25,32}`. IPv4 and IPv6 prefixes can be mixed in `only-announce`; each address family only checks its own prefixes (see [Announcements](/docs/policy/announcements#restricting-announcements-by-prefix)).

Run `pathvector generate`, then check the sessions with `birdc show protocols` and the routes sent to UTRS with `birdc show route export UTRS_AS64496_v4`.
