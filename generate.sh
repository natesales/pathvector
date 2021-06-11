#!/bin/bash
# This script generates documentation from type definitions

echo -n "Generating documentation..."
go build -o /tmp/
/tmp/wireframe generate-config-docs > docs/configuration.md
/tmp/wireframe generate-cli-docs > docs/runtime.md
echo -e '# Usage\n```' >> docs/runtime.md
/tmp/wireframe -h >> docs/runtime.md
echo '```' >> docs/runtime.md
rm /tmp/wireframe
echo "done"
