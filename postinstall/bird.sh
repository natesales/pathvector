#!/bin/bash
sudo cp /etc/bird.conf /mnt/flash/bird.conf
sudo rm /etc/bird.conf
sudo ln -s /mnt/flash/bird.conf /etc/bird.conf
sudo systemctl start bird