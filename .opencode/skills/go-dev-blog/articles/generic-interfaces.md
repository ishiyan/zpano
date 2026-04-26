# Generic Interface Patterns in Go

Interfaces can have type parameters, enabling powerful constraint patterns for generic code. These patterns are non-obvious but commonly needed.

Reference: [go.dev/blog/generic-interfaces](https://go.dev/blog/generic-interfaces)

## Self-Referential Constraints

To require that a type can operate on *itself* (e.g., compare itself), define a generic interface and use it self-referentially:

```go
// Generic interface — T is unconstrained here
type Comparer[T any] interface {
    Compare(T) int
}

// Self-referential use: E must be able to compare with itself
type Tree[E Comparer[E]] struct {
    root *node[E]
}
```

`time.Time` satisfies `Comparer[time.Time]` because it has `Compare(Time) int`. This works for any type with a matching method — no explicit interface registration needed.

## Keep Interface Type Params as `any`

When defining a generic interface for abstraction, use `any` as the constraint:

```go
// Good — any implementation can satisfy this
type Set[E any] interface {
    Insert(E)
    Delete(E)
    Has(E) bool
    All() iter.Seq[E]
}
```

Different implementations need different constraints (`comparable` for map-based sets, `Comparer[E]` for tree-based sets). Putting constraints on the interface itself would exclude valid implementations. Leave constraints to concrete types, not interfaces.

## Combining Methods and Type Sets

When a type must satisfy both a method constraint and `comparable` (e.g., for use as a map key), you have three options:

**Option 1: Embed in the original interface** (restricts all users):
```go
type Comparer[E any] interface {
    comparable
    Compare(E) int
}
```

**Option 2: New named constraint** (clean, but another name):
```go
type ComparableComparer[E any] interface {
    comparable
    Comparer[E]
}
```

**Option 3: Inline at use site** (no extra names, can be verbose):
```go
type OrderedSet[E interface {
    comparable
    Comparer[E]
}] struct { ... }
```

Prefer option 2 or 3 to avoid over-constraining the base interface.

## Pointer Receiver Constraint

When you need to instantiate a value inside a generic function, but the type's methods use pointer receivers:

```go
// PtrTo constrains PS to be *S and implement Set[E]
type PtrTo[S, E any] interface {
    *S
    Set[E]
}

func Process[E, S any, PS PtrTo[S, E]](seq iter.Seq[E]) {
    seen := PS(new(S)) // creates *S, which has the methods
    for v := range seq {
        seen.Insert(v)
    }
}
```

The trailing type parameter is inferred — callers write `Process[int, OrderedSet[int]](seq)`, not all three.

**General pattern:**
```go
func Fn[T any, PT interface{ *T; SomeMethods }]()
```

## When to Avoid the Pointer Pattern

The pointer receiver constraint adds complexity. Before reaching for it, consider:

- **Accept the interface as a value** instead of instantiating inside the function:

```go
// Simpler — caller provides a valid Set, no pointer gymnastics
func InsertAll[E any](set Set[E], seq iter.Seq[E]) {
    for v := range seq {
        set.Insert(v)
    }
}
```

This also works with non-pointer implementations (e.g., `map`-based sets with value receivers).

## Summary

| Pattern | When to use |
|---|---|
| `[E Comparer[E]]` | Type must operate on itself (compare, equal, clone) |
| Interface type param = `any` | Defining abstract interfaces — leave constraints to implementations |
| `interface{ comparable; Comparer[E] }` | Need both method and type-set constraints |
| `[T any, PT interface{*T; M}]` | Must instantiate inside function + pointer receivers — try to avoid if possible |
