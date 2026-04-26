# Profile-Guided Optimization (Go 1.21+)

PGO uses CPU profiles from production to guide compiler optimizations, typically yielding 2–7% CPU improvement.

Reference: [go.dev/blog/pgo](https://go.dev/blog/pgo)

## Workflow

1. **Collect a CPU profile** from production (e.g., via `net/http/pprof`):
   ```bash
   curl -o cpu.pprof "http://server:port/debug/pprof/profile?seconds=30"
   ```

2. **Place the profile** in the main package directory as `default.pgo`:
   ```bash
   mv cpu.pprof default.pgo
   ```

3. **Build normally** — the toolchain enables PGO automatically when `default.pgo` exists:
   ```bash
   go build -o myapp
   ```

4. **Verify PGO was used**:
   ```bash
   go version -m myapp | grep pgo
   ```

Alternatively, use `-pgo=path/to/profile.pprof` explicitly.

## Best Practices

| Practice | Details |
|---|---|
| Commit `default.pgo` to your repo | Ensures reproducible builds and automatic PGO for all builders |
| Profile from production, not benchmarks | Production profiles reflect real workload distribution |
| Update profiles periodically | Source drift reduces PGO effectiveness over time |
| Minor source changes are fine | PGO tolerates small code changes between profile collection and build |

## What PGO Optimizes

### Inlining
Hot function calls (identified by the profile) are inlined more aggressively, even if they exceed the normal size threshold. This can cascade into better escape analysis and fewer heap allocations.

### Devirtualization
If profiling shows an interface call usually targets a specific concrete type, the compiler inserts a type-check fast path with a direct call, falling back to the indirect call:

```go
// Compiler-generated (conceptual):
if f, ok := r.(*os.File); ok {
    f.Read(b)       // direct call, inlineable
} else {
    r.Read(b)       // indirect call fallback
}
```

## Impact

- Typical: **2–7% overall CPU reduction**.
- Improvements come from reduced allocations (via inlining → better escape analysis), cheaper calls, and better branch prediction.
