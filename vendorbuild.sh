#!/bin/bash
# vendorbuild.sh Pathvector vendor platform builder

# Check for lockfile
if [ -f "dist/vendor.lock" ]; then
  exit 0
else
  touch dist/vendor.lock
fi

rm -rf dist/vendor/
mkdir dist/vendor/

# Arista
echo "format: 1" > dist/vendor/manifest.txt
echo "primaryRpm: $(ls dist/pathvector*linux-amd64.rpm)" >> dist/vendor/manifest.txt
for f in dist/pathvector*linux-amd64.rpm; do echo "$f-sha1: $(sha1sum $f | cut -d " " -f 1)"; done >> dist/vendor/manifest.txt
zip dist/vendor/pathvector-"$(git describe --tags "$(git rev-list --tags --max-count=1)" | cut -c2-)"-arista-amd64.swix dist/vendor/manifest.txt dist/pathvector*linux-amd64.rpm
