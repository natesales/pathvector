#!/bin/bash

cd vendorbuild/cisco/ || { echo "Invalid path, must be in pathvector root"; exit 1; }
echo "Building Cisco package..."

# Remove old builds
rm -f ./*.tar

# Download ioxclient if it doesn't already exist
if [ ! -f ioxclient ]; then
  wget https://pubhub.devnetcloud.com/media/iox/docs/artifacts/ioxclient/ioxclient-v1.13.0.0/ioxclient_1.13.0.0_linux_amd64.tar.gz
  tar -xvzf ioxclient_1.13.0.0_linux_amd64.tar.gz
  mv ioxclient_1.13.0.0_linux_amd64/ioxclient .
  rm -rf ioxclient_*
fi

# Update version
version=$(git describe --tags "$(git rev-list --tags --max-count=1)" | cut -c2-)
rm package.yaml
sed "s/pathvector:version/$version/" package-template.yaml > package.yaml

# Build docker image
cp ioxclientcfg.yaml ~/.ioxclientcfg.yaml
./ioxclient docker package pathvector-iox:$version .
rm ~/.ioxclientcfg.yaml

mv package.tar pathvector-$version-cisco-iox-amd64.tar
