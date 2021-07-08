#!/bin/bash

rm -rf dist/arista/
mkdir dist/arista/
echo "format: 1" > dist/arista/manifest.txt
echo "primaryRpm: $(ls dist/pathvector*linux-amd64.rpm)" >> dist/arista/manifest.txt
for f in dist/pathvector*linux-amd64.rpm; do echo "$f-sha1: $(sha1sum $f | cut -d " " -f 1)"; done >> dist/arista/manifest.txt
zip dist/arista/pathvector-"$(git describe --tags "$(git rev-list --tags --max-count=1)" | cut -c2-)"-arista-amd64.swix dist/arista/manifest.txt dist/pathvector*linux-amd64.rpm
