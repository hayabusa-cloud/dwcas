# Measuring 128-bit atomic operations

Run `make bench` for the package suite. It includes single-worker loads and stores, shared reads, and matching and mismatching CAS with both result forms.

## CAS retries under contention

Workers increment both words of one aligned pair. Each successful update provides the worker’s expected value for its next attempt. On a mismatch, `Pair` reuses the pair returned by `AcqRel`. `TryReload` calls `LoadRelaxed` after `TryAcqRel` returns false. Each run checks both final words against the number of completed updates.

```sh
go test -run='^$' -bench='^BenchmarkCAS_VersionedBump_Parallel$' -benchmem -benchtime=200ms -count=10 -cpu=1,2,4 .
```

One operation is one completed update. `ns/op` measures elapsed time divided by all workers’ completed updates. `retries/op` counts failed CAS attempts per completed update. Each failure in `TryReload` also costs one explicit pair load.

One worker measures uncontended updates. With two or four workers, writers retry immediately when another update makes their expected pair stale.

For revision comparisons, use identical benchmark source on the same native amd64 or LSE2 arm64 machine. Keep the Go version, build tags, CPU placement, and worker counts fixed. Save each run’s output and compare the samples with `benchstat before.txt after.txt`. [Go parallel benchmarks](https://pkg.go.dev/testing#B.RunParallel) · [benchstat](https://pkg.go.dev/golang.org/x/perf/cmd/benchstat)
