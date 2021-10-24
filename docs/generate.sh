#!/bin/bash
# This script generates documentation from type definitions

echo -n "Generating documentation..."
go build -o /tmp/
echo -e '---
title: Configuration
sidebar_position: 3
---\n' >docs/docs/configuration.md
/tmp/pathvector docs >>docs/docs/configuration.md
echo -e '---
title: CLI
sidebar_position: 4
---\n## Usage\n```' >docs/docs/cli.md
/tmp/pathvector -h >>docs/docs/cli.md
echo '```' >>docs/docs/cli.md
echo -e '---
title: About
sidebar_position: 1
---
' >docs/docs/about.md
cat README.md >>docs/docs/about.md
rm /tmp/pathvector
curl -s https://raw.githubusercontent.com/natesales/pathvector-portal/main/README.md >docs/docs/portal.md
sed -i 's/# Pathvector Peering Portal/# Peering Portal/' docs/docs/portal.md
echo "done"
