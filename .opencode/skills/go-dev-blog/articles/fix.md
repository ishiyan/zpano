# `go fix` — Modernize Go Code

Use `go fix` (Go 1.26+) to automatically modernize Go code by applying safe, mechanical transformations that use newer language and library features.

Reference: [go.dev/blog/gofix](https://go.dev/blog/gofix)

## Basic Usage

```bash
# Fix all packages under current directory
go fix ./...

# Preview changes without applying
go fix -diff ./...

# Run a specific fixer only
go fix -minmax ./...

# Run all fixers except one
go fix -any=false ./...

# List available fixers
go tool fix help

# Show docs for a specific fixer
go tool fix help forvar
```

Run from a **clean git state** so the resulting diff contains only `go fix` edits.

## Cross-Platform Coverage

`go fix` analyzes one build configuration per run. For full coverage:

```bash
GOOS=linux   GOARCH=amd64 go fix ./...
GOOS=darwin  GOARCH=arm64 go fix ./...
GOOS=windows GOARCH=amd64 go fix ./...
```

## Run Twice for Synergistic Fixes

One fix can unlock another. Run `go fix ./...` a second time to catch chained opportunities (twice is usually enough).

## Available Analyzers

| Analyzer | Minimum Go | What it does |
|---|---|---|
| `any` | 1.18 | Replace `interface{}` with `any` |
| `minmax` | 1.21 | Replace `if/else` clamping with `min`/`max` |
| `forvar` | 1.22 | Remove redundant `x := x` in `range` loops |
| `rangeint` | 1.22 | Replace 3-clause `for i := 0; i < n; i++` with `range n` |
| `stringscut` | 1.18 | Replace `strings.Index` + slicing with `strings.Cut` |
| `stringsbuilder` | 1.10 | Replace string concatenation in loops with `strings.Builder` |
| `fmtappendf` | — | Replace `[]byte(fmt.Sprintf(...))` with `fmt.Appendf` |
| `mapsloop` | — | Replace explicit map iteration loops with `maps` package calls |
| `newexpr` | 1.26 | Replace `new(T)` + assignment with `new(value)` |
| `hostport` | — | Check format of addresses passed to `net.Dial` |
| `buildtag` | — | Check `//go:build` and `// +build` directives |
| `inline` | — | Apply fixes based on `go:fix inline` comment directives |

Fixers only apply to files whose effective Go version (from `go.mod` or `//go:build` tags) meets the minimum requirement.

## Key Transformations

**`minmax`** — clamping with `min`/`max`:
```go
// Before
x := f()
if x < 0 { x = 0 }
if x > 100 { x = 100 }

// After
x := min(max(f(), 0), 100)
```

**`rangeint`** — range-over-int loops:
```go
// Before
for i := 0; i < n; i++ { f() }

// After
for range n { f() }
```

**`stringscut`** — `strings.Cut`:
```go
// Before
i := strings.Index(s, ":")
if i >= 0 { return s[:i] }

// After
before, _, ok := strings.Cut(s, ":")
if ok { return before }
```

**`newexpr`** — `new(value)` (Go 1.26):
```go
// Before
ptr := new(string)
*ptr = "hello"

// After
ptr := new("hello")
```

This is especially useful for optional pointer fields in JSON/protobuf structs, replacing helper functions like `newInt(10)` with `new(10)`.

## Notes

- `go fix` skips generated files (those with `// Code generated` headers).
- If fixes cause unused imports, `go fix` removes them automatically.
- Rare semantic conflicts (e.g., a variable becoming unused after multiple fixes) may require manual cleanup — they surface as compile errors.
