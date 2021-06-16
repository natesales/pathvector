#!/bin/bash
# This script generates documentation from type definitions

echo -n "Generating documentation..."
go build -o /tmp/
/tmp/pathvector generate-config-docs > docs/configuration/options.md
echo '# Runtime' > docs/configuration/runtime.md
/tmp/pathvector generate-cli-docs >> docs/configuration/runtime.md
echo -e '\n## Usage\n```' >> docs/configuration/runtime.md
/tmp/pathvector -h >> docs/configuration/runtime.md
echo '```' >> docs/configuration/runtime.md
rm /tmp/pathvector
echo "done"
