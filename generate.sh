#!/bin/bash
# This script generates documentation from type definitions

echo -n "Generating documentation..."
go build -o /tmp/
echo -e '---
title: Configuration
sidebar_position: 3
---\n' > docs/docs/configuration.md
/tmp/pathvector generate-config-docs >> docs/docs/configuration.md
echo -e '---
title: CLI
sidebar_position: 4
---\n## Usage\n```' > docs/docs/cli.md
/tmp/pathvector -h >> docs/docs/cli.md
echo '```' >> docs/docs/cli.md
rm /tmp/pathvector
echo "done"
