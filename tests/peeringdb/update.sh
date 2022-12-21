#!/bin/bash

curl "https://www.peeringdb.com/api/net?info_never_via_route_servers=1" -o nvrs
curl "https://www.peeringdb.com/api/netixlan" -o netixlan
curl "https://www.peeringdb.com/api/net" -o net
