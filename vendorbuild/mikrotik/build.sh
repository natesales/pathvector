#!/bin/bash

set -e
version=$(git describe --tags "$(git rev-list --tags --max-count=1)" | cut -c2-)

# Download docker-buildx if it doesn't already exist
docker_buildx="../vendorbuild/mikrotik/docker-buildx"
if [ ! -d "$docker_buildx" ]; then
  curl -L https://github.com/docker/buildx/releases/download/v0.6.3/buildx-v0.6.3.linux-amd64 -o $docker_buildx
  chmod +x $docker_buildx
fi

for arch in amd64 arm64v8; do
  echo "Building for $arch..."
  # The build command is run in the directory at the root of the project, so the Dockerfile needs to reference files according to relative paths
  $docker_buildx build \
    --output type=tar,dest=pathvector-$version-mikrotik-$arch.tar \
    -t pathvector-mikrotik:$version-$arch \
    --build-arg ARCH=$arch \
    -f ../vendorbuild/mikrotik/Dockerfile \
    ..
done
