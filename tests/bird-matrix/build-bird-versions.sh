#!/usr/bin/env bash
# Build one or more BIRD 2.x releases from source.
#
# Usage:
#   tests/bird-matrix/build-bird-versions.sh [-o OUTDIR] [TAG...]
#
# Every TAG is a git tag of the BIRD repository (for example v2.0.7 or v2.15.1).
# The resulting binaries are written to OUTDIR/TAG/{bird,birdc,birdcl}.
# OUTDIR defaults to the directory of this script; tests/bird-matrix/v[0-9]* is
# ignored by git, so local builds never show up as untracked files.
#
# When no TAG is given, the default matrix in versions.txt is built.
#
# Environment:
#   BIRD_REPO  git URL to clone from. Defaults to the GitHub mirror of BIRD,
#              because gitlab.nic.cz is not reachable from every CI runner or
#              proxy. https://gitlab.nic.cz/labs/bird.git works as well.
#   JOBS       parallel make jobs (default: number of CPUs)
#
# Build dependencies (Debian/Ubuntu package names):
#   build-essential autoconf flex bison m4 libreadline-dev libncurses-dev git
#   (libssh-dev is optional; it enables RPKI-over-SSH and is picked up
#   automatically by ./configure when present)

set -euo pipefail

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
out_dir="$script_dir"
BIRD_REPO="${BIRD_REPO:-https://github.com/CZ-NIC/bird.git}"
JOBS="${JOBS:-$(nproc 2>/dev/null || echo 2)}"

while getopts "o:h" opt; do
  case "$opt" in
    o) out_dir="$OPTARG" ;;
    h)
      sed -n '2,24p' "$0"
      exit 0
      ;;
    *) exit 2 ;;
  esac
done
shift $((OPTIND - 1))

if [ $# -eq 0 ]; then
  # versions.txt holds one tag per line; '#' starts a comment
  mapfile -t tags < <(sed -e 's/#.*//' -e '/^[[:space:]]*$/d' "$script_dir/versions.txt")
else
  tags=("$@")
fi

mkdir -p "$out_dir"
out_dir="$(cd "$out_dir" && pwd)"

build_one() {
  local tag="$1"
  local src
  src="$(mktemp -d)"
  # shellcheck disable=SC2064
  trap "rm -rf '$src'" RETURN

  echo "==> Building BIRD $tag from $BIRD_REPO"
  # A shallow clone of a single tag is enough and much faster than a full clone
  git clone --quiet --depth 1 --branch "$tag" "$BIRD_REPO" "$src"

  (
    cd "$src"
    # Release tarballs ship a generated ./configure, git checkouts do not
    autoreconf -i
    # -fcommon: BIRD releases before 2.0.8 define some globals in headers,
    # which fails to link with GCC >= 10 (where -fno-common is the default).
    # It is harmless for newer releases.
    # --runstatedir/--sysconfdir only change compiled-in defaults; the test
    # harness always passes explicit -c/-s paths.
    ./configure --quiet CFLAGS="-O2 -fcommon" \
      --sysconfdir=/etc/bird --runstatedir=/run/bird
    # The build system of some releases (seen with v2.0.12) has a race in
    # parallel builds: a rule writes into obj/ before the directory exists
    # ("cannot create obj/nest/proto-build.c"). If the parallel build fails,
    # finish it serially; real compile errors still fail the serial run.
    make -j"$JOBS" >/dev/null || make -j1 >/dev/null
  )

  mkdir -p "$out_dir/$tag"
  for bin in bird birdc birdcl; do
    if [ -f "$src/$bin" ]; then
      install -m 0755 "$src/$bin" "$out_dir/$tag/$bin"
    fi
  done
  echo "==> $("$out_dir/$tag/bird" --version 2>&1) installed to $out_dir/$tag"
}

for tag in "${tags[@]}"; do
  build_one "$tag"
done
