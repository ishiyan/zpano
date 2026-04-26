# `//go:fix inline` — Source-Level Inliner

Use the `//go:fix inline` directive (Go 1.26+) to mark functions, type aliases, and constants for automatic inlining. When users run `go fix`, every call/reference is replaced by the function body / underlying definition. `gopls` also shows diagnostics at call sites in real time.

Reference: [go.dev/blog/inliner](https://go.dev/blog/inliner)

## How to Use It

Add `//go:fix inline` directly above a function, type alias, or constant declaration. Then run `go fix ./...` to apply replacements across the codebase.

### Inline a function (rename / migrate)

```go
package ioutil

import "os"

// Deprecated: As of Go 1.16, this function simply calls os.ReadFile.
//go:fix inline
func ReadFile(filename string) ([]byte, error) {
    return os.ReadFile(filename)
}
```

Running `go fix` replaces all calls:

```diff
-import "io/ioutil"
+import "os"

-   data, err := ioutil.ReadFile("hello.txt")
+   data, err := os.ReadFile("hello.txt")
```

### Inline a type alias

```go
package oldmath

import "newmath"

//go:fix inline
type Rational = newmath.Rational
```

All references to `oldmath.Rational` become `newmath.Rational`.

### Inline a constant

```go
package oldmath

import "newmath"

//go:fix inline
const Pi = newmath.Pi
```

All references to `oldmath.Pi` become `newmath.Pi`.

## Use Cases

**Renaming/moving functions** — Write a wrapper in the old package that calls the new location, mark it `//go:fix inline`, and deprecate it.

**Fixing API design flaws** — Implement the old API in terms of a corrected new API:

```go
// Deprecated: parameter order is confusing.
//go:fix inline
func Sub(y, x int) int {
    return newmath.Sub(x, y)
}
```

After inlining, `oldmath.Sub(1, 10)` becomes `newmath.Sub(10, 1)` — arguments reordered correctly.

**Eliminating trivial helpers** — Mark one-liner wrappers (e.g., `newInt(x)` helpers for pointer fields) so users migrate to built-in alternatives like `new(x)` in Go 1.26.

## Running

```bash
# Apply all inliner fixes
go fix ./...

# Apply only the inliner (skip other fixers)
go fix -inline ./...

# Preview changes
go fix -inline -diff ./...
```

`gopls` shows diagnostics (e.g., "call of oldmath.Sub should be inlined") as soon as the directive is added — no need to wait for `go fix`.

## Limitations

- **`defer` in callee**: If the function uses `defer`, the inliner wraps the body in a function literal (`func() { ... }()`). In batch mode (`go fix`), such calls are skipped entirely.
- **Conservative output**: The inliner may produce parameter binding declarations (`var x = expr`) instead of direct substitution when it cannot prove safety (side effects, shadowing, multiple uses of non-trivial arguments). The result is correct but may benefit from manual cleanup.
- **Semantic conflicts**: Two independent inlinings can each remove a second-to-last variable reference, making the variable unused and causing a compile error. Fix manually if this occurs.
- **Unused imports**: `go fix` automatically removes imports that become unused after inlining.

## Writing Good Inlineable Functions

1. Keep the body simple — ideally a single `return` statement calling the new API.
2. Avoid `defer`, goroutines, or complex control flow in the wrapper.
3. Add a `// Deprecated:` comment above the directive to signal intent.
4. Implement the old API entirely in terms of the new one — the inliner substitutes the body as-is.
