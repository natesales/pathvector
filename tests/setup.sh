#!/bin/bash
# Pathvector test setup
# Usage: ./setup.sh INTERNET_INTERFACE

INTERNET_INTERFACE="$1"

# Allow UDP ping. For more information, see https://github.com/go-ping/ping#linux
sysctl -w net.ipv4.ping_group_range="0 2147483647"

ip link del dev pathvector0
ip link add dev pathvector0 type dummy
ip addr add dev pathvector0 192.0.2.1/24
ip link set dev pathvector0 up

# Setup NAT
iptables -A FORWARD -i "$INTERNET_INTERFACE" -o pathvector0 -m state --state RELATED,ESTABLISHED -j ACCEPT
iptables -t nat -A POSTROUTING -o "$INTERNET_INTERFACE" -j MASQUERADE

# Check that we can ping Cloudflare DNS
echo -n "Testing IPv4 ping..."
ping -c 2 1.1.1.1 -I 192.0.2.1 &>/dev/null && echo OK || echo FAIL
