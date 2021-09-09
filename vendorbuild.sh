#!/bin/bash
# vendorbuild.sh Pathvector vendor platform builder

cd dist/

# Check for lockfile
if [ -f "vendor.lock" ]; then
  exit 0
else
  touch vendor.lock
fi

for d in ../vendorbuild/*; do
  echo "Building for $d"
  "$d"/build.sh
done
