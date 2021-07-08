#!/bin/bash
# vendorbuild.sh Pathvector vendor platform builder

# Check for lockfile
if [ -f "dist/vendor.lock" ]; then
  exit 0
else
  touch dist/vendor.lock
fi

for d in vendorbuild/* ; do
    "$d"/build.sh
done
