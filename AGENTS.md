# AGENTS.md

## Project Overview

Multi-language financial library (Python, Go, TypeScript, Zig, Rust) implementing two core modules:
1. **Day Counting** — Financial day count conventions (30/360, Actual/Actual, etc.) per ISO 20022 and ISDA standards, with Excel YEARFRAC compatibility.
2. **Performance Metrics** — Portfolio performance ratios (Sharpe, Sortino, Omega, Kappa, Calmar, Sterling, Burke, Pain, Ulcer, Martin, etc.).

Python is the reference implementation. Go, TypeScript, Zig, and Rust are ports that must match Python output to 13+ decimal places. The `laptop/` directory is an older working copy; prefer editing files under `py/`, `go/`, `ts/`, `zig/`, and `rust/`. The `performatce/` directory name is an intentional typo — do not rename it.

## Build / Lint / Test Commands

No linter, formatter, or CI/CD pipeline is configured for any language.

### Python
Dependencies: Python 3.10+, `numpy`, `scipy`. Tests import from `accounts.daycounting` and `accounts.performances` (absolute paths).
```bash
python -m unittest discover -s py -p "test_*.py"                            # all tests
python -m unittest py.daycounting.test_daycounting.TestEur30360             # single class
python -m unittest py.daycounting.test_daycounting.TestEur30360.test_excel_basis_4  # single method
```

### Go
Dependencies: Go 1.26+ (zero external deps). Run from the `go/` directory.
```bash
cd go && go test ./...                                                      # all tests
cd go && go test ./daycounting -run TestEur30360 -v                         # single test function
cd go && go test ./daycounting -run "TestEur30360/Excel_basis_4" -v         # single subtest
cd go && go test ./performance -run TestSharpeRatio -v                      # single perf test
cd go && go test ./daycounting -bench=. -benchmem                           # benchmarks
```

### TypeScript
Dependencies: Node.js 20+, TypeScript 5.3+, Jasmine 5.1+. Two independent npm packages; build `ts/daycounting` before `ts/performance`.
```bash
cd ts/daycounting && npm install && npm test    # tsc && jasmine
cd ts/performance && npm install && npm test    # jasmine (must build daycounting first)
```

### Zig
Dependencies: Zig 0.16.0-dev. The SDK is at `~/zig-sdk/zig-x86_64-linux-0.16.0-dev.2915+065c6e794/zig`.
```bash
export PATH="$HOME/zig-sdk/zig-x86_64-linux-0.16.0-dev.2915+065c6e794:$PATH"
cd zig && zig build test                        # all tests (102 tests across 5 modules)
cd zig && zig build test 2>&1 --summary all     # with per-module counts
```
Zig has no built-in way to run a single named test from the build system. To filter, use the `zig test` command directly with `--test-filter`:
```bash
cd zig && zig test src/daycounting/daycounting.zig --test-filter "act365Fixed" \
  --dep conventions -Mconventions=src/daycounting/conventions.zig -Mroot=src/daycounting/daycounting.zig
```

### Rust
Dependencies: Rust nightly (rustc 1.96.0+), zero external deps. The toolchain is at `~/.rust/bin`.
```bash
export PATH="$HOME/.rust/bin:$PATH"
cd rust && cargo test                                                       # all tests (66 tests)
cd rust && cargo test --lib daycounting                                     # daycounting tests only
cd rust && cargo test --lib performance                                     # performance tests only
cd rust && cargo test --lib test_sharpe_rf0                                 # single test by name
cd rust && cargo test --lib test_sharpe -- --nocapture                      # with stdout output
```

## Project Structure

```
py/daycounting/          — conventions.py, daycounting.py, fractional.py, tests
py/performatce/          — periodicity.py, ratios.py, tests (typo is intentional)
go/daycounting/          — daycounting.go, fractional.go, conventions/ subpackage
go/performance/          — periodicity.go, ratios.go
ts/daycounting/          — npm @portf/daycounting (conventions.ts, daycounting.ts, fractional.ts)
ts/performance/          — npm @portf_py/performance (periodicity.ts, ratios.ts)
zig/src/daycounting/     — conventions.zig, daycounting.zig, fractional.zig
zig/src/performance/     — periodicity.zig, ratios.zig
zig/build.zig            — build config: 5 modules, 5 test targets
rust/src/daycounting/    — conventions.rs, daycounting.rs, fractional.rs
rust/src/performance/    — periodicity.rs, ratios.rs
rust/Cargo.toml          — minimal config, zero external deps
laptop/                  — older working copy, do not prefer
readme/performance/      — R validation scripts, reference PDFs, CSV data, SVG charts
```

## Code Style Guidelines

### Python
- 4-space indent. Backslash `\` for line continuation.
- `snake_case` functions, `PascalCase` classes, `UPPER_SNAKE_CASE` enum members/constants, leading `_` for private.
- Relative imports within packages (`from .conventions import ...`); tests use absolute (`from accounts.daycounting import ...`).
- Type annotations use built-in generics (3.10+): `tuple[int, int, int]`, `float`, `bool`.
- `ValueError` for invalid inputs; return `None` for impossible computations. Guard clauses for zero denominators.
- Testing: `unittest.TestCase`, `assertAlmostEqual(x, y, places=13)`, `assertIsNone()`.

### Go
- `gofmt` formatting (tabs). `PascalCase` exported, `camelCase` unexported.
- `(float64, error)` for fallible APIs; `*float64` (nil) for impossible computations; plain `float64` for pure math.
- Table-driven tests with `t.Run()`. Helper `almostEqual(a, b, tolerance)` with `epsilon = 1e-14`.
- Package-level `doc.go` files. Zero external dependencies.
- Package structure: `daycounting` with `conventions` subpackage; `performance` as a separate package.

### TypeScript
- 4-space indent, strict mode in `tsconfig.json`.
- `camelCase` functions, `PascalCase` classes/enums, `UPPER_SNAKE_CASE` enum members.
- `number | null` for nullable results. `throw new Error(...)` for invalid inputs; return `null` for impossible computations.
- Jasmine 5: `describe`/`it`, `toBeCloseTo(expected, 13)`, `toBeNull()`.
- Spec naming: `.spec.ts` (daycounting), `_spec.ts` (performance).

### Zig
- 4-space indent (standard Zig formatting).
- `camelCase` for all functions (public and private), `PascalCase` for types/enums/error sets, `snake_case` for struct fields, enum members, and file-level constants.
- Imports: `const std = @import("std");` first, then module imports by build.zig name (`@import("conventions")`), then type aliases.
- Optionals (`?f64`) for impossible computations; `orelse return null` to chain. Error unions (`!void`, `!f64`) for allocation failures. Pure math functions return plain `f64`.
- `ArrayList(f64)` uses the Zig 0.16 unmanaged API: init with `.empty`, pass allocator to `.append(self.allocator, item)`, `.deinit(self.allocator)`, `.appendSlice(self.allocator, items)`. `.clearRetainingCapacity()` takes no allocator.
- The `Ratios` struct stores `allocator: std.mem.Allocator` and passes it to all ArrayList operations.
- Tests live at the bottom of source files. Use `test "descriptive name" { ... }` blocks. Assertions: `try std.testing.expect(almostEqual(...))`, `try std.testing.expectEqual(expected, actual)`. Use `testing.allocator` with `defer obj.deinit()`.
- Build.zig defines 5 modules with a dependency graph: `conventions` (no deps) -> `daycounting` -> `fractional`; `periodicity` (no deps); `ratios` (depends on all four).
- Modules are registered with `b.addModule()` and tests with `b.createModule()` + `b.addTest(.{ .root_module = mod })`.

### Rust
- Standard Rust formatting (4-space indent, `rustfmt` conventions).
- `snake_case` functions, `PascalCase` types/enums, `UPPER_SNAKE_CASE` constants, `snake_case` enum variants with `#[repr(u8)]` for the Convention enum.
- `Option<f64>` for impossible computations; plain `f64` for pure math. `Result<..., String>` or `panic!` for invalid inputs.
- `DateTime` struct with fields `year: i32, month: i32, day: i32, hour: i32, minute: i32, second: i32`.
- Tests live in `#[cfg(test)] mod tests { ... }` at the bottom of source files. Helper `almost_equal(a, b, epsilon)` with `epsilon = 1e-14` (daycounting) or `1e-13` (ratios).
- Zero external dependencies. All math uses `f64` methods (`.ln()`, `.sqrt()`, `.powf()`, `.abs()`, `.cbrt()`).
- Module structure: `daycounting` and `performance` as submodules of the `portf` crate, each with `mod.rs` re-exporting contents.

## Cross-Language Rules

- All five implementations must produce identical results to 13+ decimal places.
- Test tolerances: Python `places=13`, Go `epsilon=1e-14`, TypeScript `toBeCloseTo(x, 13)`, Zig `epsilon = 1e-13`, Rust `epsilon = 1e-14` (daycounting) / `1e-13` (ratios).
- Reference validation: Excel YEARFRAC (daycounting), R PerformanceAnalytics (performance metrics).
- Known deviations from Excel are documented inline with `Error:` comments in tests.
- 15 day count conventions share the same enum values (0–14) across all languages.
- `kurtosis` uses **population excess kurtosis** (`m4/m2^2 - 3`), matching `scipy.stats.kurtosis(bias=True, fisher=True)`.
- The `autocorrPenalty` / `_autocorr_penalty` method is a stub returning 1 in all implementations.
- Impossible computations return the language-idiomatic nullable: Python `None`, Go `*float64` nil, TypeScript `null`, Zig `?f64` null, Rust `Option<f64>` None.

## Architecture Notes

### Day Counting Module
Each language implements the same 15 conventions as an enum (values 0–14). The `dispatch` function routes a convention enum to one of 14 calculation functions (`.raw` returns null/None/nil). `fractional.py` / `fractional.go` / `fractional.ts` / `fractional.zig` / `fractional.rs` provide `frac`/`yearFrac`/`dayFrac`/`year_frac`/`day_frac` wrappers that call dispatch and convert between year fractions and day counts.

Date representation: Python uses `datetime.datetime`, Go uses `time.Time`, TypeScript uses `Date`, Zig uses a custom `DateTime` struct with fields `year: i32, month: u8, day: u8, hour: u8, minute: u8, second: u8`, Rust uses a `DateTime` struct with fields `year: i32, month: i32, day: i32, hour: i32, minute: i32, second: i32`.

### Performance Module
The `Ratios` struct/class accumulates portfolio returns incrementally via `addReturn()` and computes 20+ financial ratios at each step. All ratio methods are read-only accessors that derive values from internal state. The test dataset ("Bacon data") is a 24-element array of returns with corresponding dates, shared across all languages.

### Zig-Specific Pitfalls
- **ArrayList API (Zig 0.16):** `std.ArrayList(T)` resolves to `array_list.Aligned(T, null)`, the **unmanaged** type. It has no `.init(allocator)` method. Initialize with `.empty`, and pass the allocator to `.append(alloc, item)`, `.deinit(alloc)`, `.appendSlice(alloc, items)`. Only `.clearRetainingCapacity()` and `.shrinkRetainingCapacity()` take no allocator. The deprecated managed wrapper (`array_list.Managed`) should not be used.
- **Module imports:** Use `@import("conventions")` (the build.zig module name), not `@import("conventions.zig")` (a file path). The build.zig wires modules by name.
- **Large test data:** Writing Zig source files with many struct literals (backslash-heavy syntax) can cause JSON encoding issues in editor tools. If a file write fails, split the write into sections or use shell heredoc for the test data portion.
- **The conventions enum uses lowercase snake_case** (`.raw`, `.thirty_360_us`, `.act_365_fixed`, etc.) NOT UPPER_SNAKE_CASE.
- **Build.zig module registration:** Library modules use `b.addModule()`, test targets use `b.createModule()` + `b.addTest(.{ .root_module = mod })`. The `addTest` API in Zig 0.16 requires `root_module`, not `root_source_file`.
