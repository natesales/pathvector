dep:
	pip3 install flask

dummy-iface:
	# Allow UDP ping. For more information, see https://github.com/go-ping/ping#linux
	sudo sysctl -w net.ipv4.ping_group_range="0 2147483647"
	sudo ip link add dev dummy0 type dummy
	sudo ip addr add dev dummy0 192.0.2.1/24
	sudo ip addr add dev dummy0 2001:db8::1/64
	sudo ip link set dev dummy0 up

peeringdb-test-harness:
	nohup python3 tests/peeringdb/peeringdb-test-api.py &

test-setup: dummy-iface peeringdb-test-harness

test:
	export PATHVECTOR_TEST=1 && go test -v -race -coverprofile=coverage.txt -covermode=atomic ./pkg/... ./cmd/...

test-teardown:
	pkill -f tests/peeringdb/peeringdb-test-api.py
	sudo ip link del dev dummy0
	rm -f nohup.out

test-sequence: test-setup test test-teardown

# BIRD test matrix (see tests/bird-matrix/README.md)
# Builds one Docker image per BIRD version and runs the generate/validate
# tests in it. Defaults to the versions listed in tests/bird-matrix/versions.txt.
BIRD_VERSIONS ?= $(shell sed -e 's/\#.*//' tests/bird-matrix/versions.txt)
BIRD_MATRIX_DOCKERFILE ?= tests/bird-matrix/Dockerfile
BIRD_MATRIX_IMAGE ?= pathvector-bird-matrix
DOCKER_BUILD_ARGS ?=

bird-matrix:
	@# Keep going after a failing version so one run reports every failure
	@failed=""; for v in $(BIRD_VERSIONS); do \
		echo "### BIRD $$v"; \
		docker build $(DOCKER_BUILD_ARGS) -f $(BIRD_MATRIX_DOCKERFILE) \
			--build-arg BIRD_VERSION=$$v -t $(BIRD_MATRIX_IMAGE):$$v . \
		&& docker run --rm $(BIRD_MATRIX_IMAGE):$$v \
		|| failed="$$failed $$v"; \
	done; \
	if [ -n "$$failed" ]; then echo "### BIRD matrix failed for:$$failed"; exit 1; fi; \
	echo "### BIRD matrix passed for: $(strip $(BIRD_VERSIONS))"

# Same tests without Docker: builds BIRD into tests/bird-matrix/<version>/
# (needs the BIRD build dependencies, Go and python3-flask on the host)
bird-matrix-local:
	tests/bird-matrix/build-bird-versions.sh $(BIRD_VERSIONS)
	tests/bird-matrix/run-tests.sh $(addprefix tests/bird-matrix/,$(BIRD_VERSIONS))

.PHONY: bird-matrix bird-matrix-local
