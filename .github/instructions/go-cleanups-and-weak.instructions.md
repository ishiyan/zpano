---
description: 'Guidance on runtime.AddCleanup and weak.Pointer in Go 1.24+'
applyTo: '**/*.go'
---

# `runtime.AddCleanup` and `weak.Pointer` (Go 1.24+)

Go 1.24 introduces two low-level GC primitives that replace `runtime.SetFinalizer`. Prefer these in all new code.

Reference: [go.dev/blog/cleanups-and-weak](https://go.dev/blog/cleanups-and-weak)

**These are advanced tools.** Most Go code should not use them directly. Reach for them only when managing external resources (syscalls, cgo, mapped memory) or building deduplication/interning caches.

## `runtime.AddCleanup` — replaces `runtime.SetFinalizer`

Schedules a function to run after an object becomes unreachable:

```go
runtime.AddCleanup(obj, cleanupFunc, arg)
```

- `obj` — pointer to the object to watch
- `cleanupFunc` — called with `arg` after `obj` is unreachable
- `arg` — value passed to `cleanupFunc` (must NOT reference `obj`)

### Example: memory-mapped file

```go
mf := &MemoryMappedFile{data: data}
runtime.AddCleanup(mf, func(data []byte) {
    syscall.Munmap(data)
}, data)
```

### Why not `runtime.SetFinalizer`

| Problem with `SetFinalizer` | Fixed by `AddCleanup` |
|---|---|
| Resurrects the object (passes it to the finalizer) | Cleanup receives a separate argument, not the object |
| Breaks on reference cycles (even self-references) | Object can be in cycles — cleanup arg is independent |
| Delays reclamation by 2+ GC cycles | Object memory reclaimed immediately |
| Only one finalizer per object | Multiple independent cleanups per object |
| Not composable | Composable — attach cleanups from different subsystems |

## `weak.Pointer[T]` — weak references

A pointer the GC ignores when determining reachability:

```go
wp := weak.Make(obj)     // create weak pointer
p := wp.Value()          // returns *T or nil if collected
```

Key properties:
- **Comparable**: weak pointers have stable identity even after the target is collected
- **Safe**: `.Value()` returns `nil` (not a dangling pointer) once the object is gone

### Example: deduplication cache

Combine `weak.Pointer` and `AddCleanup` to build a cache that auto-evicts when values are no longer referenced:

```go
var cache sync.Map // map[string]weak.Pointer[T]

func GetOrCreate(key string, create func() *T) *T {
    // Try existing entry
    if val, ok := cache.Load(key); ok {
        if p := val.(weak.Pointer[T]).Value(); p != nil {
            return p
        }
        cache.CompareAndDelete(key, val) // stale entry
    }

    // Create new value
    newVal := create()
    wp := weak.Make(newVal)
    if actual, loaded := cache.LoadOrStore(key, wp); loaded {
        if p := actual.(weak.Pointer[T]).Value(); p != nil {
            return p
        }
    }

    // Auto-evict when value is collected
    runtime.AddCleanup(newVal, func(key string) {
        cache.CompareAndDelete(key, wp)
    }, key)
    return newVal
}
```

## Pitfalls

1. **Cleanup must not reference the object.** Neither the cleanup function (as a closure capture) nor its argument may keep the object reachable — otherwise the cleanup never runs. `AddCleanup` panics if the argument is exactly the watched pointer.

2. **Weak map keys must not be reachable from map values.** If a map value holds a strong reference to the weakly-pointed object, it stays alive forever.

3. **Non-deterministic.** Cleanup timing depends on GC scheduling. The runtime is permitted to never run cleanups (e.g., at program exit). Do not rely on cleanups for correctness — only for resource efficiency.

4. **Testing.** Use `runtime.GC()` in tests to trigger collection, but be aware of subtleties. See the [GC guide on testing object death](https://go.dev/doc/gc-guide#Testing_object_death).

## Quick Reference

| API | Package | Since |
|---|---|---|
| `runtime.AddCleanup(obj, fn, arg)` | `runtime` | Go 1.24 |
| `weak.Make[T](ptr)` | `weak` | Go 1.24 |
| `weak.Pointer[T].Value()` | `weak` | Go 1.24 |
| `runtime.SetFinalizer` | `runtime` | Go 1.0 (avoid in new code) |
