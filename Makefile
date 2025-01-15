dummy-iface:
	# Allow UDP ping. For more information, see https://github.com/go-ping/ping#linux
	sudo sysctl -w net.ipv4.ping_group_range="0 2147483647"
	sudo ip link add dev dummy0 type dummy
	sudo ip addr add dev dummy0 192.0.2.1/24
	sudo ip addr add dev dummy0 2001:db8::1/64
	sudo ip link set dev dummy0 up

build-pdb:
	docker build -t peeringdb-test-api tests/peeringdb

run-pdb:
	docker rm -f peeringdb-test-api || true
	docker run --name peeringdb-test-api -d -p 5001:5001 peeringdb-test-api

pdb-api: build-pdb run-pdb

test-setup: dummy-iface pdb-api

test:
	go test -v -race -coverprofile=coverage.txt -covermode=atomic ./pkg/... ./cmd/...

test-teardown:
	sudo ip link del dev dummy0

test-sequence: test-setup test test-teardown

snapshot:
	goreleaser --snapshot --clean
