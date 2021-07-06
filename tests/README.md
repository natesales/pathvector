# Integration Testing

YAML files in this directory are used for Pathvector integration tests. Each file has a category prefix that determines which tests it will be used to configure as defined in [main_test.go](https://github.com/natesales/pathvector/blob/main/main_test.go).

A test environment needs to be configured before running route optimizer integration tests. On a machine with IPv4 and IPv6 internet access, run `sudo ./test.sh INTERNET_INTERFACE` (substituting the internet-facing interface) will set up the test environment.
