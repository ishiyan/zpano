# `unique` Package — Value Interning (Go 1.23+)

The `unique` package canonicalizes comparable values, deduplicating them behind a `Handle[T]` that enables cheap pointer-based equality.

Reference: [go.dev/blog/unique](https://go.dev/blog/unique)

## API

```go
import "unique"

h := unique.Make("some string")   // returns Handle[string]
v := h.Value()                     // returns "some string"
```

| Function / Method | Description |
|---|---|
| `unique.Make[T](v)` | Returns a `Handle[T]` for the canonical copy of `v` |
| `Handle[T].Value()` | Retrieves the canonical value |

## Key Properties

- **Cheap equality**: two `Handle[T]` values are equal iff the values used to create them are equal. Comparison is a pointer comparison — much faster than comparing large strings.
- **Automatic eviction**: when all `Handle[T]` values for a given canonical value are garbage collected, the internal map entry is cleaned up (via `runtime.AddCleanup` + `weak.Pointer` internally).
- **Thread-safe**: the internal map is a concurrent map, safe for use from multiple goroutines.
- **Generic**: works with any `comparable` type, not just strings.

## Use Cases

1. **Deduplicating strings** from parsed input (e.g., zone names in `net/netip.Addr`).
2. **Cheap map keys** — use `Handle[T]` as map key for O(1) equality instead of O(n) string comparison.
3. **Interning identifiers** in parsers, compilers, or protocol handlers.

## Example: Real-World Usage in `net/netip`

```go
type Addr struct {
    z unique.Handle[addrDetail]
    // ...
}

type addrDetail struct {
    isV6   bool
    zoneV6 string
}

var z6noz = unique.Make(addrDetail{isV6: true})
```

## Workaround for Transparent String Interning

If you don't want to retain handles:

```go
s := unique.Make("my string").Value()
```

The canonical copy persists at least until the next GC cycle, providing some deduplication.

## Pitfalls

1. **You must retain the `Handle[T]`** to keep the canonical entry alive. Discarding the handle allows eviction.
2. **Not a general-purpose cache** — it's for deduplication of values, not memoization of computations.
