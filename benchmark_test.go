package haproxytime_test

import (
	"net/http"
	"os"
	"testing"

	"github.com/frobware/haproxytime"
)

func init() {
	// ListenAndServe used by the Makefile target
	// benchmark-profile only.
	go func() {
		if port := os.Getenv("BENCHMARK_PROFILE_PORT"); port != "" {
			if err := http.ListenAndServe("localhost:"+port, nil); err != nil {
				panic(err)
			}
		}
	}()
}

// How to interpret the benchmark results for:
//
// $ go test -bench=. -benchmem [-count=1] [-benchtime=1s]
// goos: linux
// goarch: amd64
// pkg: github.com/frobware/haproxytime
// cpu: 11th Gen Intel(R) Core(TM) i7-1165G7 @ 2.80GHz
// BenchmarkParseDurationMultiUnitMode-8   7338422 168.1 ns/op 0 B/op 0 allocs/op
// BenchmarkParseDurationSingleUnitMode-8 28579370 41.55 ns/op 0 B/op 0 allocs/op
// PASS
// ok github.com/frobware/haproxytime 2.648s
//
// Here's a breakdown:
//
//   - `BenchmarkParseDurationMultiUnitMode-8`: This tells you the
//     name of the benchmark function that was executed. The `-8`
//     specifies that the benchmark was run with 8 threads.
//
//   - `7338422`: This is the number of iterations that the benchmark
//     managed to run during its timed execution.
//
//   - `168.1 ns/op`: This tells you that each operation (in this case,
//     a call to `ParseDuration`) took an average of 168.1 nanoseconds.
//
//   - `0 B/op`: This indicates that the function did not allocate any
//     additional bytes of memory per operation. This is often a
//     crucial factor in performance-sensitive code, so having 0 here
//     is generally a good sign.
//
//   - `0 allocs/op`: This tells you that the function did not make
//     any heap allocations per operation. Fewer allocations often lead
//     to faster code and less pressure on the garbage collector, so
//     this is also a positive indicator.
//
//   - `PASS`: This tells you that the benchmark completed successfully
//     without any errors.
//
//   - `ok github.com/frobware/haproxytime 2.648s`: This indicates that
//     the entire test, including setup, tear-down, and the running of
//     the benchmark, completed in 2.648s.

func BenchmarkParseDurationMultiUnitMode(b *testing.B) {
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := haproxytime.ParseDuration("24d20h31m23s647ms", haproxytime.UnitMillisecond, haproxytime.ParseModeMultiUnit)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkParseDurationSingleUnitMode(b *testing.B) {
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := haproxytime.ParseDuration("2147483647ms", haproxytime.UnitMillisecond, haproxytime.ParseModeSingleUnit)
		if err != nil {
			b.Fatal(err)
		}
	}
}
