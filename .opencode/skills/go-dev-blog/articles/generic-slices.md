# Robust Generic Slice Functions (Go 1.22+)

Go 1.22 fixes a class of memory leaks in `slices.Delete`, `Compact`, `Replace`, and related functions by zeroing obsolete elements in the tail of the underlying array.

Reference: [go.dev/blog/generic-slice-functions](https://go.dev/blog/generic-slice-functions)

## The Problem (pre-1.22)

Functions like `slices.Delete` shift elements left but left stale pointers in the gap between the new and old lengths. These stale pointers prevented GC from collecting the referenced objects — a memory leak.

## The Fix

Since Go 1.22, these functions use `clear()` to zero obsolete elements after shifting:

| Function | Behavior change in 1.22 |
|---|---|
| `slices.Delete` | Zeros tail slots |
| `slices.DeleteFunc` | Zeros tail slots |
| `slices.Compact` | Zeros tail slots |
| `slices.CompactFunc` | Zeros tail slots |
| `slices.Replace` | Zeros tail slots |

## Critical Usage Rules

**Always use the return value.** These functions return a new slice header (potentially different length).

```go
// CORRECT:
s = slices.Delete(s, 2, 5)

// INCORRECT — ignores return value, s has stale/nil elements:
slices.Delete(s, 2, 5)

// INCORRECT — old slice is invalidated:
u := slices.Delete(s, 2, 3)
// s is now invalid, do not use s

// INCORRECT — accidental shadow:
s := slices.Delete(s, 2, 3)  // := creates new variable
```

## How Slices Work (reminder)

A slice is `(pointer, length, capacity)`. Functions like `Delete` modify the underlying array in place and return a slice with a shorter length. The original slice (same pointer, old length) now has zeroed/nil elements in the tail.

## Key Takeaway

After calling `Delete`, `Compact`, or `Replace`, treat the original slice as invalid. Only use the returned slice.
