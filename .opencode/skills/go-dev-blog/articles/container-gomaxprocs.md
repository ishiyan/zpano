# Container-Aware GOMAXPROCS (Go 1.25+)

Go 1.25 automatically sets `GOMAXPROCS` based on the container's CPU limit, avoiding CPU throttling in Kubernetes and similar platforms.

Reference: [go.dev/blog/container-aware-gomaxprocs](https://go.dev/blog/container-aware-gomaxprocs)

## The Problem (pre-1.25)

| Before | After (Go 1.25) |
|---|---|
| `GOMAXPROCS` = total machine cores (e.g., 128) | `GOMAXPROCS` = min(CPU limit, machine cores) |
| Container with 2 CPU limit on 128-core machine runs 128 threads | Container runs ~2 threads |
| Linux kernel throttles the process (100ms pauses) | No throttling — smoother tail latency |

## Behavior

- If a CPU limit is set (e.g., Kubernetes `resources.limits.cpu`), `GOMAXPROCS` defaults to that limit (rounded up to next integer).
- Go periodically re-checks the limit and adjusts `GOMAXPROCS` if it changes.
- Explicit `GOMAXPROCS` env var or `runtime.GOMAXPROCS()` call overrides this.

## CPU Limit vs CPU Request

| Control | Effect on GOMAXPROCS |
|---|---|
| **CPU limit** (`limits.cpu`) | Used by Go 1.25 for GOMAXPROCS default |
| **CPU request** (`requests.cpu`) | **Not used** — request is a minimum guarantee, not a cap; using it would prevent utilizing idle CPU |

## Subtleties

- CPU limits are **throughput-based** (e.g., 800ms CPU time per 100ms wall time), not parallelism-based. Spiky workloads may see latency increases from the parallelism cap.
- Fractional limits (e.g., 2.5 CPU) are rounded **up** to the next integer.
- The `go.uber.org/automaxprocs` package is no longer needed for most use cases.

## Migration

Just set `go 1.25` in `go.mod`. No code changes required.

If you were previously setting `GOMAXPROCS` explicitly or using `automaxprocs`, you can likely remove that now — the runtime handles it.
