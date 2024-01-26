# benchRedis

Results:

```plain
❯ go test --bench .                 
goos: darwin
goarch: arm64
pkg: github.com/juev/bench-redis
Benchmark_readKeysWithPipe-10                 38          33315189 ns/op
Benchmark_readKeys-10                          1        68555998875 ns/op
Benchmark_readKeysWithGoRoutines-10            7         147565268 ns/op
Benchmark_readKeysWithConc-10                  1        3433085166 ns/op
PASS
ok      github.com/juev/bench-redis     81.608s
```