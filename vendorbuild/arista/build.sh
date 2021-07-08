#!/bin/bash

cd vendorbuild/arista/ || { echo "Invalid path, must be in pathvector root"; exit 1; }
echo "Building Arista package..."

echo "format: 1" > manifest.txt
echo "primaryRpm: $(ls ../../dist/pathvector*linux-amd64.rpm)" >> manifest.txt
for f in ../../dist/pathvector*linux-amd64.rpm; do echo "$f-sha1: $(sha1sum $f | cut -d " " -f 1)"; done >> manifest.txt
zip pathvector-"$(git describe --tags "$(git rev-list --tags --max-count=1)" | cut -c2-)"-arista-amd64.swix manifest.txt ../../dist/pathvector*linux-amd64.rpm
