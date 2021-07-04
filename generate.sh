#!/bin/bash
# This script generates documentation from type definitions

# Setup
go build -o /tmp/

# API routes
blocks=$(grep -B 2 'http.HandleFunc' cmd/root.go | sed -e 's/^[ \t]*//')

delimiter=--
s=$blocks$delimiter
routes=();
while [[ $s ]]; do
    routes+=("${s%%"$delimiter"*}")
    s=${s#*"$delimiter"}
done

echo "# API

Pathvector exposes an API for control and monitoring of the service. It doesn't authenticate requests, so you should protect the API endpoint behind an isolated network or reverse proxy if you want it to listen on something other than loopback (the default).

## Routes" > docs/api.md

for i in "${routes[@]}"; do
    echo -e "### \`$(echo "$i" | grep "http.HandleFunc" | cut -d '"' -f 2)\`\n" >> docs/api.md
    echo -e "$(echo "$i" | grep -E "^\/\/ Usage: " | sed 's/^\/\/ Usage: //')\n" >> docs/api.md
    echo -e "CLI: \`$(echo "$i" | grep -E "^\/\/ CLI: " | sed 's/^\/\/ CLI: //')\`\n" >> docs/api.md
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
