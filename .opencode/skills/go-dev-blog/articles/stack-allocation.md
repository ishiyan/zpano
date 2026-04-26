# Stack Allocation of Slices in Go 1.25+

The Go compiler progressively improved slice allocation to favor the stack over the heap. Stack allocations are cheaper (often free), produce no GC pressure, and are cache-friendly. Write natural code and let the compiler optimize — avoid unnecessary hand-tuning.

Reference: [go.dev/blog/allocation-optimizations](https://go.dev/blog/allocation-optimizations)

## What the Compiler Does Automatically

| Go version | Optimization |
|---|---|
| All | `make([]T, 0, constant)` is stack-allocated if the backing store doesn't escape and the size is known at compile time |
| 1.25+ | `make([]T, 0, variable)` gets a speculative 32-byte stack buffer; used if the runtime size fits, otherwise falls back to heap |
| 1.26+ | `append` to a nil/empty slice uses a speculative 32-byte stack buffer, avoiding all heap allocations during the small-slice "startup phase" |
| 1.26+ | Escaping slices built via `append` use stack buffers for intermediates, then a single right-sized heap allocation at the escape point (`runtime.move2heap`) |

## Practical Guidance

**You no longer need to hand-optimize small slice preallocation in Go 1.26+.** This natural code:

```go
func process(c chan task) {
    var tasks []task
    for t := range c {
        tasks = append(tasks, t)
    }
    processAll(tasks)
}
```

automatically gets a stack-allocated backing store for the first few elements. No `make` preallocation needed if slices are typically small.

**When hand-optimization still helps:** If you know the expected size, preallocating with a constant is still the best option — zero heap allocations guaranteed:

```go
tasks := make([]task, 0, 10) // constant size → stack-allocated
```

With a variable size (Go 1.25+), the compiler adds a speculative stack buffer:

```go
tasks := make([]task, 0, lengthGuess) // variable → stack buffer if small enough
```

**Escaping slices (Go 1.26+):** Returning a slice from a function no longer means all intermediate allocations hit the heap. The compiler uses stack buffers internally and copies to a single right-sized heap allocation only at the return point:

```go
func extract(c chan task) []task {
    var tasks []task
    for t := range c {
        tasks = append(tasks, t)
    }
    return tasks // compiler inserts move-to-heap only here
}
```

## Debugging

If these optimizations cause correctness or performance issues, disable them:

```bash
go build -gcflags='all=-d=variablemakehash=n' ./...
```

If disabling helps, [file an issue](https://go.dev/issue/new).

## Key Takeaway

Write simple, idiomatic slice code. Avoid premature preallocation tricks — the Go 1.26 compiler handles the common cases. Reserve hand-optimization (`make` with a known constant capacity) for hot paths where you have a reliable size estimate.
