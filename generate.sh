#!/bin/bash
# This script generates documentation from type definitions

echo -n "Generating documentation..."
go build -o /tmp/
/tmp/wireframe generate-config-docs > docs/configuration/options.md
echo '# Runtime' > docs/configuration/runtime.md
/tmp/wireframe generate-cli-docs >> docs/configuration/runtime.md
echo -e '\n## Usage\n```' >> docs/configuration/runtime.md
/tmp/wireframe -h >> docs/configuration/runtime.md
echo '```' >> docs/configuration/runtime.md
rm /tmp/wireframe
echo "done"
