dep:
	pip3 install flask

peeringdb-test-harness:
	nohup python3 tests/peeringdb/peeringdb-test-api.py &

test-setup: peeringdb-test-harness

test:
	export PATHVECTOR_TEST=1 && go test -v -race -coverprofile=coverage.txt -covermode=atomic ./pkg/... ./cmd/...

test-teardown:
	pkill -f tests/peeringdb/peeringdb-test-api.py
	rm -f nohup.out

test-sequence: test-setup test test-teardown
