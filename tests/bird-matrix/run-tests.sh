#!/usr/bin/env bash
# Run the Pathvector generate/validate tests against one or more BIRD builds.
#
# Usage:
#   tests/bird-matrix/run-tests.sh BIRD [BIRD...]
#
# Each BIRD argument is either a bird binary or a directory containing one
# (for example tests/bird-matrix/v2.0.7 as produced by build-bird-versions.sh,
# or /usr/local/sbin inside the Docker image).
#
# For every BIRD binary and every test config (tests/generate-*.yml) this:
#   1. runs `pathvector generate --dry-run`, which renders the BIRD config and
#      validates it with `bird -c bird.conf -p` (pkg/bird.Validate),
#   2. re-runs `bird -p` on the generated config directly, so a parse failure
#      is reported with BIRD's own error message, and
#   3. (DAEMON_TEST=1 only) starts the BIRD daemon on the generated config and
#      checks `birdc show status` over a private control socket.
#
# Environment:
#   PATHVECTOR   pathvector binary to test (default: build one with `go build`)
#   CONFIGS      space separated list of configs (default: tests/generate-*.yml)
#   DAEMON_TEST  set to 1 to also run step 3 (needs root, used in Docker)
#   KEEP_WORK    set to 1 to keep the temporary work directory for debugging
#   XFAIL_FILE   known failures (default: tests/bird-matrix/xfail.txt)
#
# Known incompatibilities are listed in xfail.txt as "<version> <config>".
# A listed combination that fails is reported as XFAIL and does not fail the
# run; one that unexpectedly passes is reported as XPASS and does fail it, so
# the list is kept accurate.

set -euo pipefail

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
repo_dir="$(cd "$script_dir/../.." && pwd)"

if [ $# -eq 0 ]; then
  sed -n '2,29p' "$0"
  exit 2
fi

work="$(mktemp -d)"
pdb_pid=""
cleanup() {
  if [ -n "$pdb_pid" ]; then
    kill "$pdb_pid" 2>/dev/null || true
  fi
  if [ "${KEEP_WORK:-0}" = 1 ]; then
    echo "Work directory kept at $work"
  else
    rm -rf "$work"
  fi
}
trap cleanup EXIT

# --- pathvector binary -------------------------------------------------------
if [ -z "${PATHVECTOR:-}" ]; then
  echo "==> Building pathvector"
  (cd "$repo_dir" && go build -o "$work/pathvector" .)
  PATHVECTOR="$work/pathvector"
fi
echo "==> Using pathvector binary $PATHVECTOR"

# --- PeeringDB test API --------------------------------------------------------
# The test configs set peeringdb-url to http://localhost:5000/api and
# generate-complex.yml uses filter-never-via-route-servers, which makes
# pathvector query PeeringDB and abort if that fails. Serve the canned
# responses from tests/peeringdb/ (same as `make peeringdb-test-harness`)
# unless something is already listening on port 5000.
pdb_up() {
  python3 -c 'import urllib.request; urllib.request.urlopen("http://127.0.0.1:5000/api/net?asn=34553", timeout=1)' 2>/dev/null
}
if ! pdb_up; then
  echo "==> Starting PeeringDB test API"
  # The API opens its fixtures relative to the working directory
  (cd "$repo_dir" && exec python3 tests/peeringdb/peeringdb-test-api.py) >"$work/peeringdb.log" 2>&1 &
  pdb_pid=$!
  for _ in $(seq 1 50); do
    pdb_up && break
    sleep 0.2
  done
  if ! pdb_up; then
    echo "PeeringDB test API did not start:" >&2
    cat "$work/peeringdb.log" >&2
    exit 1
  fi
fi
# Also makes pkg/peeringdb use the local endpoint, like `make test`
export PATHVECTOR_TEST=1

# --- test configs --------------------------------------------------------------
if [ -n "${CONFIGS:-}" ]; then
  read -r -a configs <<<"$CONFIGS"
else
  configs=("$repo_dir"/tests/generate-*.yml)
fi

# The configs reference ../tests/blocklist.txt (they are written to be run
# from cmd/ by `go test`). Run pathvector from $work/run with a tests symlink
# next to it so that relative path resolves.
mkdir -p "$work/run"
ln -s "$repo_dir/tests" "$work/tests"

failures=()
passes=0
xfails=0

XFAIL_FILE="${XFAIL_FILE:-$script_dir/xfail.txt}"
# is_xfail VERSION CONFIG: versions are compared without a leading "v"
is_xfail() {
  [ -f "$XFAIL_FILE" ] || return 1
  sed -e 's/#.*//' "$XFAIL_FILE" | awk -v v="${1#v}" -v c="$2" '
    { sub(/^v/, "", $1) } $1 == v && $2 == c { found = 1 } END { exit !found }'
}

daemon_test() {
  local bird="$1" conf_dir="$2" name="$3"
  local birdc
  birdc="$(dirname "$bird")/birdc"
  local sock="$work/$name.ctl"
  # -f: stay in foreground so we own the process; -s: private control socket
  # so a system BIRD is never touched; -P: pid file
  "$bird" -f -c "$conf_dir/bird.conf" -s "$sock" -P "$work/$name.pid" \
    >"$work/$name.daemon.log" 2>&1 &
  local pid=$!
  local ok=1
  for _ in $(seq 1 50); do
    [ -S "$sock" ] && break
    sleep 0.1
  done
  if [ -S "$sock" ] && "$birdc" -s "$sock" show status | grep -q "Daemon is up and running"; then
    ok=0
  else
    cat "$work/$name.daemon.log" >&2
  fi
  kill "$pid" 2>/dev/null || true
  wait "$pid" 2>/dev/null || true
  return $ok
}

for arg in "$@"; do
  bird="$arg"
  [ -d "$bird" ] && bird="$bird/bird"
  if [ ! -x "$bird" ]; then
    echo "BIRD binary $bird not found or not executable" >&2
    exit 1
  fi
  bird="$(cd "$(dirname "$bird")" && pwd)/$(basename "$bird")"
  # "BIRD version 2.17.7+branch..." -> "2.17.7"; the suffix comes from
  # building a shallow git checkout and carries no information here
  bird_version="$("$bird" --version 2>&1 | awk '{print $3}')"
  bird_version="${bird_version%%+*}"
  bird_version="${bird_version#v}"
  echo "==> Testing BIRD $bird_version ($bird)"

  for config in "${configs[@]}"; do
    cname="$(basename "$config" .yml)"
    name="$bird_version-$cname"
    cache="$work/$name"
    # Point the config at this BIRD binary and a private cache directory.
    # Drop any top-level setting of those keys first: duplicate keys are a
    # YAML error.
    {
      grep -Ev '^(bird-binary|cache-directory):' "$config"
      echo
      echo "bird-binary: $bird"
      echo "cache-directory: $cache"
    } >"$work/$name.yml"

    step="generate"
    if (cd "$work/run" && "$PATHVECTOR" generate --dry-run --config "$work/$name.yml") >"$work/$name.log" 2>&1; then
      step="bird -p"
      if "$bird" -p -c "$cache/bird.conf" >>"$work/$name.log" 2>&1; then
        step="daemon"
        if [ "${DAEMON_TEST:-0}" != 1 ] || daemon_test "$bird" "$cache" "$name" >>"$work/$name.log" 2>&1; then
          if is_xfail "$bird_version" "$cname"; then
            echo "    XPASS $cname (listed in $(basename "$XFAIL_FILE"), remove the entry)"
            failures+=("BIRD $bird_version: $cname (unexpected pass)")
          else
            echo "    PASS $cname"
            passes=$((passes + 1))
          fi
          continue
        fi
      fi
    fi
    if is_xfail "$bird_version" "$cname"; then
      echo "    XFAIL $cname ($step, known incompatibility)"
      xfails=$((xfails + 1))
      continue
    fi
    echo "    FAIL $cname ($step)"
    sed 's/^/      | /' "$work/$name.log"
    failures+=("BIRD $bird_version: $cname ($step)")
  done
done

echo
echo "==> $passes passed, $xfails known failures, ${#failures[@]} failed"
for f in "${failures[@]}"; do
  echo "    $f"
done
[ ${#failures[@]} -eq 0 ]
