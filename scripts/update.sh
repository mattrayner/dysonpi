#!/bin/bash
sudo apt-get update -y
sudo apt-get upgrade -y
cd /home/pi/go/src/github.com/mattrayner/dysonpi
git pull origin HEAD
make deps
make build
make update-scripts
cd ~
sudo systemctl restart dysonpi.service
