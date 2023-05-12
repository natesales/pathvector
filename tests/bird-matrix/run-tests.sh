#!/bin/bash
#for tag in $(ls | grep "^v2.0." | sort -V); do
#  echo "Testing $tag"
#done

version=$1
user=$(whoami)

echo Starting BIRD "$version"
sudo mkdir -p /run/bird
sudo mkdir -p /etc/bird
echo "protocol device {}" | sudo tee /etc/bird/bird.conf >/dev/null
echo Starting bird

{
  sleep 1
  sudo chown "$user":"$user" /run/bird
  sudo chown "$user":"$user" /run/bird/bird.ctl
  birdc show status
} &

sudo "$version"/bird -s /run/bird/bird.ctl -c /etc/bird/bird.conf -d
