#!/bin/bash

set -e

# Clone jetez if it doesn't already exist
if [ ! -d ../vendorbuild/juniper/jetez ]; then
  git clone https://github.com/Juniper/jetez ../vendorbuild/juniper/jetez
  cd ../vendorbuild/juniper/jetez && python3 setup.py install && cd ..
fi

# Link xorrisofs
sudo ln -s "$(which xorrisofs)" /usr/bin/mkisofs

# Copy binary
cp pathvector_freebsd_amd64/pathvector ../vendorbuild/juniper/src

# Get version
version=$(git describe --tags "$(git rev-list --tags --max-count=1)" | cut -c2-)

# Build package
python3 ../vendorbuild/juniper/jetez/jet/main.py \
  --source ../vendorbuild/juniper/src/ \
  --cert ../vendorbuild/juniper/cert.pem \
  --key ../vendorbuild/juniper/key.pem \
  --version "$version"

# Cleanup
sudo rm -rf /usr/bin/mkisofs
rm -rf ../vendorbuild/juniper/src/pathvector

# Rename file
mv pathvector*tgz pathvector-"$version"-juniper-amd64.tgz
