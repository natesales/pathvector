#!/bin/bash
# Pathvector test setup

# Allow UDP ping. For more information, see https://github.com/go-ping/ping#linux
sysctl -w net.ipv4.ping_group_range="0 2147483647"

ip link del dev dummy0
ip link add dev dummy0 type dummy
ip addr add dev dummy0 192.0.2.1/24
ip link set dev dummy0 up

nohup python3 tests/peeringdb/peeringdb-test-api.py &
