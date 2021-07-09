#!/bin/bash

# Download ioxclient if it doesn't already exist
if [ ! -f ../vendorbuild/cisco/ioxclient ]; then
  wget https://pubhub.devnetcloud.com/media/iox/docs/artifacts/ioxclient/ioxclient-v1.13.0.0/ioxclient_1.13.0.0_linux_amd64.tar.gz
  tar -xvzf ioxclient_1.13.0.0_linux_amd64.tar.gz
  mv ioxclient_1.13.0.0_linux_amd64/ioxclient ../vendorbuild/cisco/ioxclient
  rm -rf ioxclient_*
fi

# Update version
version=$(git describe --tags "$(git rev-list --tags --max-count=1)" | cut -c2-)
sed "s/pathvector:version/$version/" ../vendorbuild/cisco/package-template.yaml > package.yaml

# Build docker image
cp ../vendorbuild/cisco/ioxclientcfg.yaml ~/.ioxclientcfg.yaml
../vendorbuild/cisco/ioxclient docker package pathvector-iox:$version .
rm ~/.ioxclientcfg.yaml

mv package.tar pathvector-$version-cisco-iox-amd64.tar
