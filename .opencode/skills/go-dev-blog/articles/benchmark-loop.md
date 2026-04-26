# `testing.B.Loop` — Preferred Benchmark Loop (Go 1.24+)

`b.Loop()` is the preferred way to write Go benchmarks. It replaces the `for range b.N` pattern with fewer pitfalls and more accurate results.

Reference: [go.dev/blog/testing-b-loop](https://go.dev/blog/testing-b-loop)

## Migration

**Before** (b.N-style):

```go
func BenchmarkFoo(b *testing.B) {
    // setup
    b.ResetTimer()
    for range b.N {
        result = foo()
    }
    b.StopTimer()
    // cleanup
}
```

**After** (b.Loop-style):

```go
func BenchmarkFoo(b *testing.B) {
    // setup (automatically excluded from timing)
    for b.Loop() {
        foo()
    }
    // cleanup (automatically excluded from timing)
}
```

## Benefits Over `b.N`

| Problem with `b.N` | Fixed by `b.Loop()` |
|---|---|
| Setup/cleanup included in timing unless `ResetTimer`/`StopTimer` are called | Timing starts at first `b.Loop()` call, stops after last iteration |
| Compiler can dead-code-eliminate unused results (fake sub-ns benchmarks) | Compiler prevents dead-code elimination inside `b.Loop()` loops |
| Benchmark function called multiple times with increasing `b.N` | Single invocation — `b.Loop()` ramps up internally |
| Benchmark code can accidentally depend on `b.N` value or iteration index | No iteration count or index exposed |

## Per-Iteration Setup

When each iteration needs its own setup (e.g., regenerating input), manually control the timer inside the loop:

```go
func BenchmarkSort(b *testing.B) {
    ints := make([]int, N)
    for b.Loop() {
        b.StopTimer()
        fillRandomInts(ints)
        b.StartTimer()
        slices.Sort(ints)
    }
}
```

## Rules

- Exactly **one** `b.Loop()` loop per benchmark function.
- Cannot mix `b.Loop()` and `b.N` in the same benchmark.
- Every iteration should do the same work (no branching on iteration count).
- No need for result sinks or accumulator variables — dead-code elimination is prevented automatically.
