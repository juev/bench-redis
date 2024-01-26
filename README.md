# benchRedis

Results:

```plain
 ❯ go test --bench .
goos: darwin
goarch: arm64
pkg: github.com/juev/bench-redis
Benchmark_readKeysWithPipe-10                 38          36000475 ns/op
Benchmark_readKeys-10                          1        67612201542 ns/op
Benchmark_readKeysWithGoRoutines-10            7         182168185 ns/op
PASS
ok      github.com/juev/bench-redis      76.912s
```