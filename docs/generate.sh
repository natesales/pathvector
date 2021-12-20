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

# Add plugin readmes
rm docs/docs/plugins/*.md
for p in plugins/*/; do
  plugin=$(echo "$p" | cut -d "/" -f 2)
  cp "plugins/$plugin/README.md" "docs/docs/plugins/$plugin.md"
done

echo "done"
