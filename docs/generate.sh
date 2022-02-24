#!/bin/bash
# This script generates documentation from type definitions

echo Generating documentation
go build -o /tmp/

echo Generating options page
echo -e '---
title: Configuration
sidebar_position: 3
---\n' >docs/docs/configuration.md
/tmp/pathvector docs >>docs/docs/configuration.md

echo Generating CLI preview
echo -e '---
title: CLI
sidebar_position: 4
---\n# Usage\n```' >docs/docs/cli.md
/tmp/pathvector -h >>docs/docs/cli.md
echo '```' >>docs/docs/cli.md

echo Copying README to index
echo -e '---
title: About
sidebar_position: 1
---
' >docs/docs/about.md
cat README.md >>docs/docs/about.md

rm /tmp/pathvector

# Add peering portal readme page
curl -s https://raw.githubusercontent.com/natesales/pathvector-portal/main/README.md >docs/docs/portal.md
sed -i 's/# Pathvector Peering Portal/# Peering Portal/' docs/docs/portal.md

# Generate PDF documentation
echo Generating PDF

commit=$(git rev-list --tags --max-count=1)
version=$(git describe --tags "$commit" | cut -c2-)

echo "
# Pathvector

Pathvector Edge Routing Platform version $version commit $commit" > release_full.md

for f in docs/docs/installation.md docs/docs/cli.md docs/docs/configuration.md; do
  sed '1 { /^---/ { :a N; /\n---/! ba; d} }' >> release_full.md < $f
done

echo "Copyright © 2022 Nate Sales" >> release_full.md

pandoc release_full.md \
    -f gfm \
    -V linkcolor:blue \
    -V geometry:a4paper \
    -V geometry:margin=2cm \
    --pdf-engine=xelatex \
    -o pathvector-$version-release.pdf

rm release_full.md
