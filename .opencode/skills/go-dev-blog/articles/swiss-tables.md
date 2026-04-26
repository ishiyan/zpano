# Swiss Table Maps (Go 1.24+)

Go 1.24 replaces the built-in `map` implementation with Swiss Tables, an open-addressed hash table design using SIMD-accelerated group probing.

Reference: [go.dev/blog/swisstable](https://go.dev/blog/swisstable)

## Impact

- **No code changes needed.** This is a runtime-internal change to the built-in `map` type.
- Microbenchmarks: up to **60% faster** individual map operations.
- Full application benchmarks: ~**1.5% geometric mean CPU improvement**.

## How It Works

| Concept | Description |
|---|---|
| **Groups** | Slots are organized into groups of 8 |
| **Control word** | 64-bit metadata per group; each byte holds 7 bits of hash (`h2`) or empty/deleted marker |
| **Probe sequence** | `h1` selects starting group; all 8 slots checked in parallel via SIMD byte comparison |
| **Higher load factor** | Parallel probing enables higher max load → lower average memory footprint |

### Lookup flow

1. Compute `hash(key)` → split into `h1` (upper 57 bits) and `h2` (lower 7 bits).
2. Use `h1` to select the starting group.
3. Compare `h2` against all 8 control bytes simultaneously (SIMD or portable arithmetic).
4. Check candidate slots (where `h2` matched) by comparing full keys.
5. If no match and no empty slot, probe the next group.

## Go-Specific Design

### Incremental growth
Each map is split into multiple Swiss Tables (max 1024 entries each). Growing one table copies at most 1024 entries, bounding tail latency.

### Iteration during modification
Go allows map modification during `range`. The iterator keeps a reference to the pre-growth table for ordering, but consults the grown table for current values.

## What Developers Should Know

- **No behavioral changes** — the `map` type works exactly as before.
- **Potential edge cases** — iteration order is intentionally randomized and may differ from previous versions (as always specified).
- Future: group size may increase to 16 with wider SIMD (16 hash comparisons in parallel).
