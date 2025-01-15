down:
	docker rm -f pathvector-peeringdb-test-api || true
	docker rm -f pathvector-bird || true
	sudo ip link del dev dummy0 || true

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
	docker rm -f pathvector-peeringdb-test-api || true
	docker run --name pathvector-peeringdb-test-api -d -p 5001:5001 peeringdb-test-api

run-bird:
	docker rm -f pathvector-bird || true
	rm -rf /tmp/bird-conf || true
	mkdir -p /tmp/bird-conf
	echo "protocol device {}" > /tmp/bird-conf/bird.conf
	docker run \
		-d \
		--privileged \
		--name pathvector-bird \
		-p 5002:5002 \
		-v $(shell pwd)/tests/bird-entrypoint.sh:/entrypoint.sh \
		-v /tmp/bird-conf/:/etc/bird/ \
		-v /tmp/test-cache/:/tmp/test-cache/ \
		pierky/bird:2.16 /entrypoint.sh

pdb-api: build-pdb run-pdb

test-setup: dummy-iface run-bird pdb-api

test:
	go test -v -p 1 -coverprofile=coverage.txt -covermode=atomic ./pkg/... ./cmd/...

test-sequence: test-setup test down

snapshot:
	goreleaser --snapshot --clean
