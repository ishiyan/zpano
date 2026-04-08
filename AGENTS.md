# AGENTS.md

## Project Overview

Multi-language financial library (Python, Go, TypeScript, Zig, Rust) implementing four core modules:
1. **Day Counting** — Financial day count conventions (30/360, Actual/Actual, etc.) per ISO 20022 and ISDA standards, with Excel YEARFRAC compatibility.
2. **Performance Metrics** — Portfolio performance ratios (Sharpe, Sortino, Omega, Kappa, Calmar, Sterling, Burke, Pain, Ulcer, Martin, etc.).
3. **Roundtrips** — Trading round-trip tracking with execution matching, PnL computation (gross/net, long/short), and 100+ incremental performance statistics (ROI, Sharpe, Sortino, Calmar, drawdowns, MAE/MFE, efficiency, consecutive streaks, duration analytics).
4. **Symbology** — Financial security identifier validation (ISIN, CUSIP, SEDOL) with check digit calculation and validation, country code verification, and SEC 13F workaround support.

Python is the reference implementation. Go, TypeScript, Zig, and Rust are ports that must match Python output to 13+ decimal places. The `laptop/` directory is an older working copy; prefer editing files under `py/`, `go/`, `ts/`, `zig/`, and `rust/`. The `performatce/` directory name is an intentional typo — do not rename it.

## Build / Lint / Test Commands

No linter, formatter, or CI/CD pipeline is configured for any language.

### Python
Dependencies: Python 3.10+, `numpy`, `scipy`. Tests import from `accounts.daycounting` and `accounts.performances` (absolute paths).
```bash
python -m unittest discover -s py -p "test_*.py"                            # all tests
python -m unittest py.daycounting.test_daycounting.TestEur30360             # single class
python -m unittest py.daycounting.test_daycounting.TestEur30360.test_excel_basis_4  # single method
python -m unittest py.roundtrips.test_roundtrip                             # roundtrip tests
python -m unittest py.roundtrips.test_performance                           # roundtrip performance tests
python -m unittest py.symbology.test_isin py.symbology.test_cusip py.symbology.test_sedol  # symbology tests
```

### Go
Dependencies: Go 1.26+ (zero external deps). Run from the `go/` directory.
```bash
cd go && go test ./...                                                      # all tests
cd go && go test ./daycounting -run TestEur30360 -v                         # single test function
cd go && go test ./daycounting -run "TestEur30360/Excel_basis_4" -v         # single subtest
cd go && go test ./performance -run TestSharpeRatio -v                      # single perf test
cd go && go test ./roundtrips -run TestRoundtripPerformance -v              # roundtrip perf test
cd go && go test ./symbology -run TestValidateISIN -v                       # single symbology test
cd go && go test ./daycounting -bench=. -benchmem                           # benchmarks
```

### TypeScript
Dependencies: Node.js 20+, TypeScript 5.3+, Jasmine 5.1+. Four independent npm packages; build `ts/daycounting` before `ts/performance` or `ts/roundtrips`. `ts/symbology` is standalone.
```bash
cd ts/daycounting && npm install && npm test    # tsc && jasmine
cd ts/performance && npm install && npm test    # jasmine (must build daycounting first)
cd ts/roundtrips && npm install && npm test     # tsc && jasmine (must build daycounting first)
cd ts/symbology && npm install && npm test      # tsc && jasmine (standalone)
```

### Zig
Dependencies: Zig 0.16.0-dev. Installed at `/usr/local/zig/` with `/usr/local/bin/zig` on PATH.
```bash
cd zig && zig build test                        # all tests (329 tests across 14 modules)
cd zig && zig build test 2>&1 --summary all     # with per-module counts
```
Zig has no built-in way to run a single named test from the build system. To filter, use the `zig test` command directly with `--test-filter`:
```bash
cd zig && zig test src/daycounting/daycounting.zig --test-filter "act365Fixed" \
  --dep conventions -Mconventions=src/daycounting/conventions.zig -Mroot=src/daycounting/daycounting.zig
```

### Rust
Dependencies: Rust 1.75.0+ (installed via apt at `/usr/bin/rustc`, `/usr/bin/cargo`), zero external deps.
```bash
cd rust && cargo test                                                       # all tests (302 tests)
cd rust && cargo test --lib daycounting                                     # daycounting tests only
cd rust && cargo test --lib performance                                     # performance tests only
cd rust && cargo test --lib roundtrips                                      # roundtrips tests only
cd rust && cargo test --lib symbology                                       # symbology tests only
cd rust && cargo test --lib test_sharpe_rf0                                 # single test by name
cd rust && cargo test --lib test_sharpe -- --nocapture                      # with stdout output
```

## Project Structure

```
py/daycounting/          — conventions.py, daycounting.py, fractional.py, tests
py/performatce/          — periodicity.py, ratios.py, tests (typo is intentional)
py/roundtrips/           — execution.py, side.py, matching.py, grouping.py, roundtrip.py, performance.py, tests
py/symbology/            — isin.py, cusip.py, sedol.py, tests
go/daycounting/          — daycounting.go, fractional.go, conventions/ subpackage
go/performance/          — periodicity.go, ratios.go
go/roundtrips/           — execution.go, side.go, matching.go, grouping.go, roundtrip.go, performance.go, tests
go/symbology/            — isin.go, cusip.go, sedol.go, tests
ts/daycounting/          — npm @zpano/daycounting (conventions.ts, daycounting.ts, fractional.ts)
ts/performance/          — npm @zpano/performance (periodicity.ts, ratios.ts)
ts/roundtrips/           — npm @zpano/roundtrips (execution.ts, side.ts, matching.ts, grouping.ts, roundtrip.ts, performance.ts)
ts/symbology/            — npm @zpano/symbology (isin.ts, cusip.ts, sedol.ts)
zig/src/daycounting/     — conventions.zig, daycounting.zig, fractional.zig
zig/src/performance/     — periodicity.zig, ratios.zig
zig/src/roundtrips/      — execution.zig, side.zig, matching.zig, grouping.zig, roundtrip.zig, performance.zig
zig/src/symbology/       — isin.zig, cusip.zig, sedol.zig
zig/build.zig            — build config: 14 modules, 14 test targets
rust/src/daycounting/    — conventions.rs, daycounting.rs, fractional.rs
rust/src/performance/    — periodicity.rs, ratios.rs
rust/src/roundtrips/     — mod.rs, execution.rs, side.rs, matching.rs, grouping.rs, roundtrip.rs, performance.rs
rust/src/symbology/      — mod.rs, isin.rs, cusip.rs, sedol.rs
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
- Package structure: `daycounting` with `conventions` subpackage; `performance`, `roundtrips`, and `symbology` as separate packages.

### TypeScript
- 4-space indent, strict mode in `tsconfig.json`.
- `camelCase` functions, `PascalCase` classes/enums, `UPPER_SNAKE_CASE` enum members.
- `number | null` for nullable results. `throw new Error(...)` for invalid inputs; return `null` for impossible computations.
- Jasmine 5: `describe`/`it`, `toBeCloseTo(expected, 13)`, `toBeNull()`.
- Spec naming: `.spec.ts` (daycounting, symbology), `_spec.ts` (performance).

### Zig
- 4-space indent (standard Zig formatting).
- `camelCase` for all functions (public and private), `PascalCase` for types/enums/error sets, `snake_case` for struct fields, enum members, and file-level constants.
- Imports: `const std = @import("std");` first, then module imports by build.zig name (`@import("conventions")`), then type aliases.
- Optionals (`?f64`) for impossible computations; `orelse return null` to chain. Error unions (`!void`, `!f64`) for allocation failures. Pure math functions return plain `f64`.
- `ArrayList(f64)` uses the Zig 0.16 unmanaged API: init with `.empty`, pass allocator to `.append(self.allocator, item)`, `.deinit(self.allocator)`, `.appendSlice(self.allocator, items)`. `.clearRetainingCapacity()` takes no allocator.
- The `Ratios` struct stores `allocator: std.mem.Allocator` and passes it to all ArrayList operations.
- Tests live at the bottom of source files. Use `test "descriptive name" { ... }` blocks. Assertions: `try std.testing.expect(almostEqual(...))`, `try std.testing.expectEqual(expected, actual)`. Use `testing.allocator` with `defer obj.deinit()`.
- Build.zig defines 14 modules with a dependency graph: `conventions` (no deps) -> `daycounting` -> `fractional`; `periodicity` (no deps); `ratios` (depends on all four); `execution` (depends on `fractional`); `side`, `matching`, `grouping` (no deps); `roundtrip` (depends on `execution`, `side`, `fractional`); `performance` (depends on `roundtrip`, `execution`, `side`, `fractional`); `isin`, `cusip`, `sedol` (no deps, standalone symbology modules).
- Modules are registered with `b.addModule()` and tests with `b.createModule()` + `b.addTest(.{ .root_module = mod })`.

### Rust
- Standard Rust formatting (4-space indent, `rustfmt` conventions).
- `snake_case` functions, `PascalCase` types/enums, `UPPER_SNAKE_CASE` constants, `snake_case` enum variants with `#[repr(u8)]` for the Convention enum.
- `Option<f64>` for impossible computations; plain `f64` for pure math. `Result<..., String>` or `panic!` for invalid inputs.
- `DateTime` struct with fields `year: i32, month: i32, day: i32, hour: i32, minute: i32, second: i32`.
- Tests live in `#[cfg(test)] mod tests { ... }` at the bottom of source files. Helper `almost_equal(a, b, epsilon)` with `epsilon = 1e-14` (daycounting) or `1e-13` (ratios).
- Zero external dependencies. All math uses `f64` methods (`.ln()`, `.sqrt()`, `.powf()`, `.abs()`, `.cbrt()`).
- Module structure: `daycounting`, `performance`, `roundtrips`, and `symbology` as submodules of the `portf` crate, each with `mod.rs` re-exporting contents.

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

### Roundtrips Module
Six source files per language implement trading round-trip tracking:

1. **Execution** — `OrderSide` enum (BUY/SELL) and `Execution` struct (side, price, commission_per_unit, unrealized_price_high, unrealized_price_low, datetime).
2. **Side** — `RoundtripSide` enum (LONG/SHORT).
3. **Matching** — `RoundtripMatching` enum (FIFO/LIFO).
4. **Grouping** — `RoundtripGrouping` enum (FILL_TO_FILL/FLAT_TO_FLAT/FLAT_TO_REDUCED).
5. **Roundtrip** — Immutable struct computed from entry/exit Execution + quantity. 19 properties: side, quantity, entry/exit time and price, duration, highest/lowest price, commission, gross/net PnL, MAE/MFE (%), entry/exit/total efficiency. Different computation paths for LONG vs SHORT.
6. **Performance** — Incremental statistics tracker (100+ computed properties). Constructor takes initial_balance, annual_risk_free_rate, annual_target_return, day_count_convention. `add_roundtrip()` accumulates ~115 internal state variables. Properties cover: ROI stats, Sharpe/Sortino/Calmar ratios (regular and annualized), rate of return, profit ratios, counts (total/long/short/winning/losing by gross/net), PnL totals and averages, duration statistics, MAE/MFE/efficiency averages, consecutive streaks, drawdowns, recovery factor.

Key behavioral details:
- **"loosing" spelling** is intentional throughout the codebase (e.g., `loosing_count`, `net_loosing_pnl`) — must be preserved in all ports for API compatibility.
- **PnL quirk in `add_roundtrip()`**: When accumulating `net_long_winning_pnl`, `net_long_loosing_pnl`, `net_short_winning_pnl`, `net_short_loosing_pnl`, the code adds `gross_pnl` (not `net_pnl`). This is intentional and verified by tests.
- **Duration**: Python uses `timedelta.total_seconds()`, Go uses `time.Duration`, TypeScript stores milliseconds with a seconds getter, Zig/Rust store `duration_seconds` as `f64`.
- **Drawdown percent denominator**: `initial_balance + max_net_pnl`.
- **None vs 0.0 policy**: Risk-adjusted ratios (Sharpe, Sortino, Calmar) return null/None when denominator is 0. All other averages/ratios return `0.0` when denominator is 0.
- **Immutability**: Python uses `__setattr__` override, Go uses exported struct fields (convention), TypeScript uses `readonly`, Zig uses `const`, Rust uses private fields with getter methods.
- **`year_frac` with RAW convention**: `total_seconds / 31_556_952` (Gregorian year in seconds). Dates auto-sorted if inverted.

### Symbology Module
Three source files per language implement financial security identifier validation:

1. **ISIN** (International Securities Identification Number, ISO 6166) — 12-char identifier: 2-letter ISO 3166-1 country code + 9-char alphanumeric NSIN + 1 check digit. `validate()` checks country code + check digit. `validate_check_digit()` uses Luhn algorithm processing right-to-left. `validate_country()` covers ~250 ISO country codes via a switch/match statement. `calculate_check_digit()` processes characters with a multiply toggle: single-digit ordinals (0-9) and two-digit ordinals (10-35 from letters A-Z) have different doubling paths, but only one toggle per character regardless of digit count.

2. **CUSIP** (Committee on Uniform Security Identification Procedures) — 9-char North American security identifier: 6-char issuer + 2-char issue + 1 check digit. Uses "Modulus 10 Double Add Double" where every odd-positioned digit (0-indexed, i%2==1) is doubled. Characters map: 0-9 as-is, A-Z to 10-35, `*`=36, `@`=37, `#`=38. Has SEC 13F workaround: if check digit doesn't match but position 6 is '9' and position 7 is '0' or '5', validation passes (options CUSIPs).

3. **SEDOL** (Stock Exchange Daily Official List) — 7-char UK/Ireland security identifier: 6-char alphanumeric + 1 check digit. Three styles: old-style (first char digit 0-8, digits only), user-defined (first char '9', vowels allowed), new-style (first char letter, no vowels AEIOU). Weight array `[1, 3, 1, 7, 3, 9]` for positions 0-5. Letters map A=10..Z=35.

Key behavioral details:
- **No inter-module dependencies**: Each symbology validator (ISIN, CUSIP, SEDOL) is fully standalone with no imports from other modules.
- **Error handling**: All languages use error returns (`Result`/`error`/exceptions) for invalid input, not nullable returns. `validate_country()` returns `bool`.
- **Immutability**: Go uses type aliases on `string` with methods; Python, TypeScript, Zig, and Rust use standalone functions accepting string/slice inputs.
- **Test data volume**: ISIN ~815 cases (validate ~260, check digit ~274, country ~280), CUSIP ~4475 cases, SEDOL ~3249 cases — all ported identically across all five languages.

### Zig-Specific Pitfalls
- **ArrayList API (Zig 0.16):** `std.ArrayList(T)` resolves to `array_list.Aligned(T, null)`, the **unmanaged** type. It has no `.init(allocator)` method. Initialize with `.empty`, and pass the allocator to `.append(alloc, item)`, `.deinit(alloc)`, `.appendSlice(alloc, items)`. Only `.clearRetainingCapacity()` and `.shrinkRetainingCapacity()` take no allocator. The deprecated managed wrapper (`array_list.Managed`) should not be used.
- **Module imports:** Use `@import("conventions")` (the build.zig module name), not `@import("conventions.zig")` (a file path). The build.zig wires modules by name.
- **Large test data:** Writing Zig source files with many struct literals (backslash-heavy syntax) can cause JSON encoding issues in editor tools. If a file write fails, split the write into sections or use shell heredoc for the test data portion.
- **The conventions enum uses lowercase snake_case** (`.raw`, `.thirty_360_us`, `.act_365_fixed`, etc.) NOT UPPER_SNAKE_CASE.
- **Build.zig module registration:** Library modules use `b.addModule()`, test targets use `b.createModule()` + `b.addTest(.{ .root_module = mod })`. The `addTest` API in Zig 0.16 requires `root_module`, not `root_source_file`.

## Quick Reference: Build & Test All Languages

All commands run from the project root (`~/repos/zpano/`).

| Language | Prerequisites | Command | Expected |
|----------|--------------|---------|----------|
| **Python** | Python 3.10+, `numpy`, `scipy`, symlink `ln -sf py accounts`, `touch py/__init__.py` | `PYTHONPATH=. python3 -m unittest discover -s py -p "test_*.py" -t .` | 291 tests |
| **Go** | Go 1.26+ | `cd go && go test ./...&& cd ..`| 5 packages OK |
| **TypeScript** | Node.js 20+, TypeScript 5.3+, Jasmine 5.1+ | `cd ts/daycounting && npm install && npm test && cd ../performance && npm install && npm test && cd ../roundtrips && npm install && npm test && cd ../symbology && npm install && npm test && cd ..` | 92 + 112 + 315 + 8539 specs |
| **Zig** | Zig 0.16.0-dev | `cd zig && zig build test --summary all && cd ..` | 329 tests |
| **Rust** | Rust 1.75.0+ (apt) | `cd rust && cargo test && cd ..` | 302 tests |

```bash
python3 -m unittest discover -s py -p "test_*.py" -t .
cd go && go test ./...&& cd ..
cd ts/daycounting && npm install && npm test && cd ../performance && npm install && npm test && cd ../roundtrips && npm install && npm test && cd ../symbology && npm install && npm test && cd ../..
cd zig && zig build test --summary all && cd ..
cd rust && cargo test && cd ..
```

### Build Only (no tests)

| Language | Command | Notes |
|----------|---------|-------|
| **Python** | *(interpreted — no build step)* | |
| **Go** | `cd go && go build ./...` | Compiles all packages |
| **TypeScript** | `cd ts/daycounting && npm run build && cd ../performance && npm run build && cd ../roundtrips && npm run build && cd ../symbology && npm run build` | Build daycounting first (performance and roundtrips depend on it) |
| **Zig** | `cd zig && zig build` | Build without running tests |
| **Rust** | `cd rust && cargo build` | Debug build; add `--release` for optimized |

```bash
cd go && go build ./... && cd ..
cd ts/daycounting && npm run build && cd ../performance && npm run build && cd ../roundtrips && npm run build && cd ../symbology && npm run build && cd ../..
cd zig && zig build && cd ..
cd rust && cargo build && cd ..
```
