# Type Construction and Cycle Detection (Go 1.26)

Go 1.26 improves the compiler's type checker with a simpler, more robust cycle detection algorithm for recursive types, fixing several compiler panics.

Reference: [go.dev/blog/type-construction-and-cycle-detection](https://go.dev/blog/type-construction-and-cycle-detection)

## What Changed

The type checker now uses a systematic approach to detect invalid cyclic type definitions, particularly involving:
- Array sizes computed from `unsafe.Sizeof`, `len`, or other constant expressions
- Values of incomplete (still-being-constructed) types used in those expressions

## Key Concepts

| Term | Meaning |
|---|---|
| **Complete type** | All fields populated, all referenced types also complete — safe to deconstruct |
| **Incomplete type** | Under construction; referencing it is fine, deconstructing (inspecting internals) is not |
| **Cycle error** | A type's definition depends on itself in a way that can't be resolved |

## Valid vs Invalid Recursive Types

```go
// VALID — pointer to incomplete type (pointers have known size)
type T [unsafe.Sizeof(new(T))]int

// INVALID — T{} requires knowing T's size, which depends on T
type T [unsafe.Sizeof(T{})]int

// VALID — simple recursive type via pointer indirection
type Node struct { next *Node }

// INVALID — direct self-reference without indirection
type T T
```

**Rule:** it is never sound to operate on an incomplete value whose type is a defined type, because the type name alone conveys no underlying type information.

## Practical Impact

- Fixes compiler panics on esoteric type definitions (issues #75918, #76383, #76384, #76478, and more).
- No behavioral change for normal Go code — only affects edge cases with recursive types in constant expressions.
- Sets the foundation for future type system improvements.

## For Most Developers

This is a compiler internals improvement. Unless you write recursive type definitions involving `unsafe.Sizeof` or similar constant expressions with self-referential types, you won't notice any change.
