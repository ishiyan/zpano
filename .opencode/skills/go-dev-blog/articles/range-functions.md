# Range Over Function Types — Iterators (Go 1.23+)

Go 1.23 adds `for/range` support for function types, enabling standard iterators over user-defined containers.

Reference: [go.dev/blog/range-functions](https://go.dev/blog/range-functions)

## Iterator Types (`iter.Seq` / `iter.Seq2`)

```go
import "iter"

type Seq[V any]      func(yield func(V) bool)
type Seq2[K, V any]  func(yield func(K, V) bool)
```

An iterator calls `yield` with each value. If `yield` returns `false`, stop early and return.

## Writing an Iterator

```go
func (s *Set[E]) All() iter.Seq[E] {
    return func(yield func(E) bool) {
        for v := range s.m {
            if !yield(v) {
                return
            }
        }
    }
}
```

**Convention:** container types should expose an `All()` method returning an iterator, so callers always use `for v := range x.All()`.

## Using an Iterator

```go
for v := range s.All() {
    fmt.Println(v)
}
```

The compiler creates the yield function and wires up `break`/`panic` automatically.

## Pull Iterators

Convert a push iterator to a pull iterator with `iter.Pull`:

```go
next, stop := iter.Pull(s.All())
defer stop()
for v, ok := next(); ok; v, ok = next() {
    fmt.Println(v)
}
```

Use pull iterators when you need to iterate two sequences in parallel or step through manually. Always call `stop()` when done.

## Standard Library Functions (Go 1.23+)

### `slices` package

| Function | Signature |
|---|---|
| `slices.All` | `([]E) iter.Seq2[int, E]` |
| `slices.Values` | `([]E) iter.Seq[E]` |
| `slices.Collect` | `(iter.Seq[E]) []E` |
| `slices.AppendSeq` | `([]E, iter.Seq[E]) []E` |
| `slices.Backward` | `([]E) iter.Seq2[int, E]` |
| `slices.Sorted` | `(iter.Seq[E]) []E` |
| `slices.Chunk` | `([]E, int) iter.Seq([]E)` |

### `maps` package

| Function | Signature |
|---|---|
| `maps.All` | `(map[K]V) iter.Seq2[K, V]` |
| `maps.Keys` | `(map[K]V) iter.Seq[K]` |
| `maps.Values` | `(map[K]V) iter.Seq[V]` |
| `maps.Collect` | `(iter.Seq2[K, V]) map[K, V]` |
| `maps.Insert` | `(map[K, V], iter.Seq2[K, V])` |

## Adapter Pattern

```go
func Filter[V any](f func(V) bool, s iter.Seq[V]) iter.Seq[V] {
    return func(yield func(V) bool) {
        for v := range s {
            if f(v) {
                if !yield(v) {
                    return
                }
            }
        }
    }
}
```

Chain adapters: `slices.Collect(Filter(isLong, maps.Values(m)))`.

## Rules

- Requires `go 1.23` or later in `go.mod`.
- Always check `if !yield(v) { return }` inside iterators.
- By convention, return an iterator from a method (e.g. `All()`) rather than making the method itself an iterator.
- Pull iterators: always `defer stop()`.
