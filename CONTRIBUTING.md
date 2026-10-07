# Contributing to Pathvector

Contributions to Pathvector are always appreciated! Please open an [Issue](https://github.com/natesales/pathvector/issues/new/choose) for bug reports or a [Pull Request](https://github.com/natesales/pathvector/compare) for code contributions.

## Code Style

Pathvector follows Go's conventional code style. Please run `go fmt` before opening a PR.

## Testing

Pathvector relies on Go's [testing](https://pkg.go.dev/testing) package. All new code contributions should have associated unit and integration tests where applicable.

`make test-sequence` runs the unit and integration tests (it sets up a dummy interface and the local PeeringDB test API first, see the `Makefile`).

### BIRD version matrix

Changes to the BIRD templates should be checked against all supported BIRD versions, not only the one installed locally. The BIRD test matrix builds several BIRD releases from source in Docker and validates the generated configuration with each of them:

```sh
make bird-matrix                                  # all versions in tests/bird-matrix/versions.txt
make bird-matrix BIRD_VERSIONS="v2.0.7 v2.15.1"   # selected versions
```

It also runs in CI for every push and pull request and before every release. See [tests/bird-matrix/README.md](tests/bird-matrix/README.md) for running it without Docker and for adding versions.
