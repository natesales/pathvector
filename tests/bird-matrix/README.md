# BIRD test matrix

Pathvector generates BIRD 2 configuration and supports every BIRD release
from 2.0.7 (`supportedMin` in `pkg/bird/bird.go`) onwards. The regular test
suite only runs against whatever `bird2` package the CI runner ships, so this
harness builds several BIRD versions from source and checks that the
configuration Pathvector generates is accepted by each of them.

For every BIRD version and every test config (`tests/generate-*.yml`) the
harness:

1. runs `pathvector generate --dry-run`, which renders the config and
   validates it with `bird -p` (see `bird.Validate` in `pkg/bird/bird.go`),
2. runs `bird -p` on the generated config again to report BIRD's own error
   message, and
3. inside Docker (`DAEMON_TEST=1`): starts the BIRD daemon on the generated
   config and checks `birdc show status`.

The test configs query PeeringDB, so the harness starts the local PeeringDB
test API from `tests/peeringdb/` (the same one `make test-setup` uses) and no
access to peeringdb.com is needed. The configs also list a blocklist URL;
failing to fetch it only logs a warning.

## Running it

With Docker (recommended), from the repository root:

```sh
make bird-matrix                                  # all versions in versions.txt
make bird-matrix BIRD_VERSIONS="v2.0.7 v2.15.1"   # selected versions
```

This builds one image per version (`pathvector-bird-matrix:<version>`) from
`tests/bird-matrix/Dockerfile` and runs it. The BIRD build layer is cached,
so later runs only rebuild Pathvector. Extra `docker build` flags (for
example `--network host` or proxy settings) can be passed with
`DOCKER_BUILD_ARGS="..."`, and a different Dockerfile with
`BIRD_MATRIX_DOCKERFILE=...`. A single version by hand:

```sh
docker build -f tests/bird-matrix/Dockerfile --build-arg BIRD_VERSION=v2.0.7 \
  -t pathvector-bird-matrix:v2.0.7 .
docker run --rm pathvector-bird-matrix:v2.0.7
```

Without Docker, you need the BIRD build dependencies (`build-essential
autoconf flex bison m4 libreadline-dev libncurses-dev`, optionally
`libssh-dev`), Go and `python3-flask`:

```sh
make bird-matrix-local BIRD_VERSIONS="v2.0.7 v2.15.1"
# or step by step:
tests/bird-matrix/build-bird-versions.sh v2.0.7 v2.15.1  # -> tests/bird-matrix/v2.0.7/bird ...
tests/bird-matrix/run-tests.sh tests/bird-matrix/v2.0.7 tests/bird-matrix/v2.15.1
tests/bird-matrix/run-tests.sh /usr/sbin/bird            # test the system BIRD
```

`run-tests.sh` takes these environment variables: `PATHVECTOR` (binary to
test, built with `go build` if unset), `CONFIGS` (configs to use),
`DAEMON_TEST=1` (daemon check, needs root), `KEEP_WORK=1` (keep the generated
configs for debugging) and `XFAIL_FILE`.

## Continuous integration

`.github/workflows/bird-matrix.yml` runs the harness as a GitHub Actions
matrix with one job per line of `versions.txt` (`make bird-matrix
BIRD_VERSIONS=<version>` in each job). It runs on every branch push and pull
request, can be started by hand (workflow_dispatch), and is called by
`release.yml`: the goreleaser job `needs` the matrix, so a release tag is
only published when every BIRD version passes.

## Choosing the BIRD source

BIRD is cloned from the GitHub mirror `https://github.com/CZ-NIC/bird.git` by
default, because `gitlab.nic.cz` is not reachable from every CI runner. Set
`BIRD_REPO` (script) or `--build-arg BIRD_REPO=...` (Docker) to use
`https://gitlab.nic.cz/labs/bird.git` or another mirror.

## Tested versions

`versions.txt` lists the default versions (one git tag per line):

| Version  | Why                                         |
|----------|---------------------------------------------|
| v2.0.7   | oldest supported version (`supportedMin`)   |
| v2.0.12  | last 2.0.x release                          |
| v2.14    | `bird2` package of Ubuntu 24.04 LTS         |
| v2.15.1  | 2.15 release line                           |
| v2.17.7  | 2.17 release line                           |
| v2.19.3  | latest 2.x release when the matrix was added |

## Known failures

`xfail.txt` lists combinations of BIRD version and config that are known not
to work, one per line as `<version> <config-name>` (config name without
`.yml`). They are reported as `XFAIL` and do not fail the run. If a listed
combination starts passing it is reported as `XPASS` and the run fails, so the
list cannot go stale. Currently:

- `2.0.7 generate-complex`: `transit-lock` generates AS path masks using the
  `+` operator, which BIRD only supports since 2.0.8.

Prefer fixing the template or raising `supportedMin` over adding entries.

## Adding a version

1. Find the tag: `git ls-remote --tags https://github.com/CZ-NIC/bird.git 'v2*'`.
2. Try it: `make bird-matrix BIRD_VERSIONS=v2.x.y`.
3. Add the tag to `versions.txt`. The CI workflow builds its matrix from
   this file, so no workflow change is needed.
4. If it fails because of a real incompatibility, fix the template or add an
   entry with the reason to `xfail.txt`.

When `supportedMin` changes, update the oldest version in `versions.txt`.
