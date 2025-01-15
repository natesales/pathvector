# Matrix Test Suite

## Session establishment test

- Run BIRD on both router containers
- Namespaced routing tables, control sockets, and cache directories

### Assertions
- Session established
- Route received
- Route installed in kernel
- Ping

## Complex config generation test

- Generate with `generate-complex` on different versions of BIRD

## Test Suite

Go test flag to use container test target
