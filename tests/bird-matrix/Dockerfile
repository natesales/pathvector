# BIRD test matrix image: one BIRD version built from source + pathvector.
#
# Build from the repository root (the build context must contain the source):
#   docker build -f tests/bird-matrix/Dockerfile \
#     --build-arg BIRD_VERSION=v2.0.7 -t pathvector-bird-matrix:v2.0.7 .
#   docker run --rm pathvector-bird-matrix:v2.0.7
#
# or simply: make bird-matrix BIRD_VERSIONS="v2.0.7 v2.15.1"
#
# See tests/bird-matrix/README.md for details.

# Ubuntu LTS is used for building and running BIRD (it matches the GitHub
# Actions runners); the Go stage only produces a static binary.
ARG UBUNTU_RELEASE=24.04
ARG GO_VERSION=1.22

# ---------------------------------------------------------------------------
# Stage 1: build BIRD from a git tag
# ---------------------------------------------------------------------------
FROM ubuntu:${UBUNTU_RELEASE} AS bird
RUN apt-get update \
 && apt-get install -y --no-install-recommends \
      build-essential autoconf flex bison m4 git ca-certificates \
      libreadline-dev libncurses-dev libssh-dev \
 && rm -rf /var/lib/apt/lists/*

# Any tag of the BIRD repository, e.g. v2.0.7, v2.14, v2.17.7
ARG BIRD_VERSION=v2.0.7
# GitHub mirror by default; gitlab.nic.cz is not reachable from every CI
# runner. Override with --build-arg BIRD_REPO=https://gitlab.nic.cz/labs/bird.git
ARG BIRD_REPO=https://github.com/CZ-NIC/bird.git

# Reuse the same build script that is used for local (non-Docker) builds, so
# both paths build BIRD identically. Only the script is copied here so that
# the (slow) BIRD build stays cached when pathvector sources change.
COPY tests/bird-matrix/build-bird-versions.sh /build/
RUN BIRD_REPO="${BIRD_REPO}" /build/build-bird-versions.sh -o /opt/bird "${BIRD_VERSION}" \
 && mv /opt/bird/"${BIRD_VERSION}" /opt/bird/current

# ---------------------------------------------------------------------------
# Stage 2: build pathvector from the build context
# ---------------------------------------------------------------------------
FROM golang:${GO_VERSION} AS pathvector
WORKDIR /src
# Download modules first so they are cached independently of source changes
COPY go.mod go.sum ./
RUN go mod download
COPY . .
# The docs/ templates consumed via go:generate are committed, so a plain
# build is enough. CGO is not needed and a static binary runs anywhere.
RUN CGO_ENABLED=0 go build -o /out/pathvector .

# ---------------------------------------------------------------------------
# Stage 3: test runner
# ---------------------------------------------------------------------------
FROM ubuntu:${UBUNTU_RELEASE}
# Runtime libraries of bird (libssh) and birdc (readline, tinfo),
# python3-flask for the local PeeringDB test API (tests/peeringdb/) and
# ca-certificates for blocklist-urls in the configs.
# (libreadline8t64 is the Ubuntu 24.04 name of libreadline8.)
RUN apt-get update \
 && apt-get install -y --no-install-recommends \
      libreadline8t64 libtinfo6 libssh-4 \
      python3 python3-flask ca-certificates \
 && rm -rf /var/lib/apt/lists/*

COPY --from=bird /opt/bird/current/ /usr/local/sbin/
COPY --from=pathvector /out/pathvector /usr/local/bin/pathvector
# run-tests.sh locates the test configs and the PeeringDB fixtures relative
# to its own location, so keep the repository layout below /src
COPY tests/ /src/tests/

ARG BIRD_VERSION=v2.0.7
ENV BIRD_VERSION=${BIRD_VERSION} \
    PATHVECTOR=/usr/local/bin/pathvector \
    DAEMON_TEST=1
# Fail the image build early if BIRD does not even start
RUN bird --version

WORKDIR /src
ENTRYPOINT ["/src/tests/bird-matrix/run-tests.sh"]
CMD ["/usr/local/sbin/bird"]
