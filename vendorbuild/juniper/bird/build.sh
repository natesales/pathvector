#!/bin/bash
# TODO: Add this to pathvector repo

set -e

BIRD_VERSION="2.0.8"

echo "Building BIRD v$BIRD_VERSION..."

# Clone bird if it doesn't already exist
if [ ! -d "bird/" ]; then
  git clone https://gitlab.nic.cz/labs/bird -b v"$BIRD_VERSION" bird/
  # Build bird
  cd bird/ || {
    echo "Unable to find bird directory"
    exit 1
  }
  autoreconf
  ./configure
  sed -i 's/^LDFLAGS=.*/& -static/' Makefile
  make
fi

jetez_dir="../jetez"

# Clone jetez if it doesn't already exist
if [ ! -d "$jetez_dir" ]; then
  git clone https://github.com/Juniper/jetez "$jetez_dir"
  cd "$jetez_dir" && sudo python3 setup.py install && cd -
fi

# Link xorrisofs
if [ ! -e /usr/bin/mkisofs ]; then
  sudo ln -s "$(which xorrisofs)" /usr/bin/mkisofs
fi

# Copy binary
cp bird/{bird,birdc} src/

# Build package
python3 "$jetez_dir"/jet/main.py \
  --source src/ \
  --cert "$jetez_dir"/cert.pem \
  --key "$jetez_dir"/key.pem \
  --version "$BIRD_VERSION"

# Cleanup
sudo rm /usr/bin/mkisofs
rm -rf bird/
rm -rf src/bird*
