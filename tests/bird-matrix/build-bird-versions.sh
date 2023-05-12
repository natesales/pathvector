#!/bin/bash
# Build the last 5 bird versions

if [ ! -d bird ]; then
  git clone https://gitlab.nic.cz/labs/bird.git
fi

cd bird || exit 1

for tag in $(git tag | grep "^v2.0." | sort -V | tail -n 5); do
  echo "Building $tag"
  git reset --hard HEAD
  git checkout "$tag"
  autoreconf
  ./configure
  make
  mkdir ../"$tag"
  mv bird ../"$tag"
  mv birdc ../"$tag"
done

rm -rf bird
