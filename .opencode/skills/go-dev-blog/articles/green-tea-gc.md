# Green Tea Garbage Collector (Go 1.25 experiment, Go 1.26 default)

Green Tea is a new GC mark phase that works with memory **pages** instead of individual objects, reducing GC CPU cost by 10–40%.

Reference: [go.dev/blog/greenteagc](https://go.dev/blog/greenteagc)

## Enabling

```bash
# Go 1.25: opt-in experiment
GOEXPERIMENT=greenteagc go build ./...

# Go 1.26: default (opt-out with nogreenteagc)
GOEXPERIMENT=nogreenteagc go build ./...
```

No code changes required. No API changes.

## What Changed

| Aspect | Old GC | Green Tea |
|---|---|---|
| Work unit | Individual objects | Whole 8 KiB pages |
| Work list | Object pointers | Page references |
| Memory access pattern | Random jumps between objects | Sequential scans within pages |
| CPU cache utilization | Poor (frequent cache misses) | Good (sequential, predictable) |
| Vector hardware (AVX-512) | Not applicable | Used for page-level bitmap operations |

## How It Works (simplified)

1. Instead of scanning one object at a time, Green Tea **accumulates** objects to scan per page.
2. When a page is taken from the work list, all pending objects in that page are scanned in a single sequential pass.
3. Two bits per object slot: "seen" (found a pointer to it) and "scanned" (already processed). The difference identifies what to scan next.
4. Pages can re-enter the work list as new objects are discovered.

## Performance

- **Modal improvement**: ~10% reduction in GC CPU time.
- **Best case**: up to 40% reduction.
- **Typical impact**: if your app spends 10% in GC, expect 1–4% overall CPU reduction.
- **Vector acceleration** (AVX-512, Go 1.26+): additional ~10% GC CPU reduction on supported hardware.

## When It Helps Less

- Workloads where only 1 object per page needs scanning at a time (irregular heap structures).
- The implementation has a fast path for single-object pages to minimize regressions.
- Even scanning just 2% of a page at a time can outperform the old graph flood.

## Production Status

Already deployed at Google at scale. Production-ready in Go 1.25 as an experiment.
