#!/bin/bash

set -e

# Clone jetez if it doesn't already exist
if [ ! -d ../vendorbuild/juniper/jetez ]; then
  git clone https://github.com/Juniper/jetez ../vendorbuild/juniper/jetez
  cd ../vendorbuild/juniper/jetez && rm -rf .git && sudo python3 setup.py install && cd -
fi

# Link xorrisofs
if [ ! -e /usr/bin/mkisofs ]; then
  sudo ln -s "$(which xorrisofs)" /usr/bin/mkisofs
fi

# Copy binary
cp pathvector_freebsd_amd64/pathvector ../vendorbuild/juniper/src-amd64/
cp pathvector_freebsd_arm64/pathvector ../vendorbuild/juniper/src-arm64/

# Get version
version=$(git describe --tags "$(git rev-list --tags --max-count=1)" | cut -c2-)

# Build package
python3 ../vendorbuild/juniper/jetez/jet/main.py \
  --source ../vendorbuild/juniper/src-amd64/ \
  --cert ../vendorbuild/juniper/cert.pem \
  --key ../vendorbuild/juniper/key.pem \
  --version "$version"
python3 ../vendorbuild/juniper/jetez/jet/main.py \
  --source ../vendorbuild/juniper/src-arm64/ \
  --cert ../vendorbuild/juniper/cert.pem \
  --key ../vendorbuild/juniper/key.pem \
  --version "$version"

# Cleanup
sudo rm -rf /usr/bin/mkisofs
rm -rf ../vendorbuild/juniper/src-amd64/pathvector
rm -rf ../vendorbuild/juniper/src-arm64/pathvector

# Rename file
if [ ! -e pathvector-"$version"-juniper-amd64.tgz ]; then
  mv pathvector-x86*tgz pathvector-"$version"-juniper-amd64.tgz
fi

if [ ! -e pathvector-"$version"-juniper-arm64.tgz ]; then
  mv pathvector-arm*tgz pathvector-"$version"-juniper-arm64.tgz
fi
