#!/bin/bash
# This script generates documentation from type definitions

# Setup
go build -o /tmp/

# Copy index
cp README.md docs/index.md

# API routes
echo "# API

Pathvector exposes an API for control and monitoring of the service. It doesn't authenticate requests, so you should protect the API endpoint behind an isolated network or reverse proxy if you want it to listen on something other than loopback (the default).

## Routes" > docs/api.md

grep 'autodoc API route' cmd/root.go | sed -e 's/^[ \t]*//' | while read line; do
  echo "### \`$(echo $line | grep -o -P '(?<=route ).*(?=:)')\`" >> docs/api.md
  echo $line | sed -n -e 's/^.*: //p' >> docs/api.md
done

# Config file docs
/tmp/pathvector docs > docs/configuration/options.md

# CLI flags
echo '# Runtime' > docs/configuration/runtime.md
echo -e '\n## Usage\n```' >> docs/configuration/runtime.md
/tmp/pathvector -h >> docs/configuration/runtime.md
echo '```' >> docs/configuration/runtime.md

# Cleanup
rm /tmp/pathvector
