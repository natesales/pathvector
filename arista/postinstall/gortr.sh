#!/bin/bash
sudo cp /usr/share/gortr/cf.pub /mnt/flash/cf.pub
sudo rm -rf /usr/share/gortr/cf.pub
sudo systemctl start gortr
