#!/bin/bash
set -ex
apt install -y socat
bird -d &
socat TCP-LISTEN:5002,fork UNIX-CONNECT:/usr/local/var/run/bird.ctl
