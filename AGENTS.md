# AGENTS.md

## Project Overview

Multi-language financial library (Python, Go, TypeScript, Zig, Rust) implementing seven core modules:
1. **Day Counting** — Financial day count conventions (30/360, Actual/Actual, etc.) per ISO 20022 and ISDA standards, with Excel YEARFRAC compatibility.
2. **Entities** — Financial trading data types (Bar, Quote, Trade, Scalar) with computed properties and component extraction. Also provides component enums (BarComponent, QuoteComponent, TradeComponent) with factory functions, mnemonics, and default constants used by the indicators module. See the `entities-architecture` skill for the full cross-language reference.
3. **Performance Metrics** — Portfolio performance ratios (Sharpe, Sortino, Omega, Kappa, Calmar, Sterling, Burke, Pain, Ulcer, Martin, etc.).
4. **Roundtrips** — Trading round-trip tracking with execution matching, PnL computation (gross/net, long/short), and 100+ incremental performance statistics (ROI, Sharpe, Sortino, Calmar, drawdowns, MAE/MFE, efficiency, consecutive streaks, duration analytics).
5. **Symbology** — Financial security identifier validation (ISIN, CUSIP, SEDOL) with check digit calculation and validation, country code verification, and SEC 13F workaround support.
6. **Indicators** — 63 technical analysis indicators (SMA, EMA, RSI, MACD, Bollinger Bands, etc.) organized by author, with a shared `core/` framework providing the `Indicator` interface, `LineIndicator` base, metadata, descriptor registry, output types, and frequency response utilities. Implemented in all five languages (Go, TypeScript, Python, Zig, Rust). See the `indicator-architecture` skill for the full design reference.
7. **Cmd** — Three CLI tools (`icalc`, `iconf`, `ifres`) that exercise the indicators module: indicator calculation against 252-bar reference data, chart configuration generation, and frequency response analysis. Implemented in all five languages.

Go and TypeScript are the reference implementations for modules 6 and 7; Python, Zig, and Rust implementations are complete and must match reference output to 13+ decimal places. Python is the reference for modules 1–5. All ports must match reference output to 13+ decimal places. The `laptop/` directory is an older working copy; prefer editing files under `py/`, `go/`, `ts/`, `zig/`, and `rs/`. The `performatce/` directory name is an intentional typo — do not rename it.

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
python -m unittest py.entities.test_entities                                # entities tests
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
cd go && go test ./entities -run TestBar -v                                 # single entities test
cd go && go test ./indicators/...                                           # all indicator tests
cd go && go test ./indicators/common/simplemovingaverage -v                 # single indicator test
cd go && go test ./daycounting -bench=. -benchmem                           # benchmarks
```

### TypeScript
Dependencies: Node.js 20+, TypeScript 5.3+, Jasmine 5.1+. Single unified npm package at `ts/` with all modules.
```bash
cd ts && npm install && npm test                                             # all tests (8935 specs)
cd ts && npm run build                                                       # build only (tsc)
```

### Zig
Dependencies: Zig 0.16.0-dev. Installed at `/usr/local/zig/` with `/usr/local/bin/zig` on PATH.
```bash
cd zig && zig build test                        # all tests (367 tests across 21 modules)
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
cd rs && cargo test                                                       # all tests (342 tests)
cd rs && cargo test --lib daycounting                                     # daycounting tests only
cd rs && cargo test --lib performance                                     # performance tests only
cd rs && cargo test --lib roundtrips                                      # roundtrips tests only
cd rs && cargo test --lib symbology                                       # symbology tests only
cd rs && cargo test --lib entities                                        # entities tests only
cd rs && cargo test --lib test_sharpe_rf0                                 # single test by name
cd rs && cargo test --lib test_sharpe -- --nocapture                      # with stdout output
```

## Project Structure

```
py/daycounting/          — conventions.py, daycounting.py, fractional.py, tests
py/performatce/          — periodicity.py, ratios.py, tests (typo is intentional)
py/roundtrips/           — execution.py, side.py, matching.py, grouping.py, roundtrip.py, performance.py, tests
py/symbology/            — isin.py, cusip.py, sedol.py, tests
py/entities/             — bar.py, quote.py, trade.py, scalar.py, bar_component.py, quote_component.py, trade_component.py, tests
py/indicators/           — 63 indicators organized by author (common/, john_ehlers/, welles_wilder/, etc.)
py/indicators/core/      — shared types: indicator.py, line_indicator.py, identifier.py, metadata.py, descriptor.py, outputs/, frequency_response/
py/indicators/factory/   — identifier + JSON params → indicator instance
py/cmd/icalc/            — CLI indicator calculator
py/cmd/iconf/            — CLI chart configuration generator
py/cmd/ifres/            — CLI frequency response calculator
go/daycounting/          — daycounting.go, fractional.go, conventions/ subpackage
go/performance/          — periodicity.go, ratios.go
go/roundtrips/           — execution.go, side.go, matching.go, grouping.go, roundtrip.go, performance.go, tests
go/symbology/            — isin.go, cusip.go, sedol.go, tests
go/entities/             — bar.go, quote.go, trade.go, scalar.go, barcomponent.go, quotecomponent.go, tradecomponent.go, tests
go/indicators/           — 63 indicators organized by author (common/, johnehlers/, welleswilder/, etc.)
go/indicators/core/      — shared types: indicator.go, lineindicator.go, identifier.go, metadata.go, descriptor.go, outputs/, frequency-response/
go/indicators/factory/   — identifier + JSON params → indicator instance
go/cmd/icalc/            — CLI indicator calculator
go/cmd/iconf/            — CLI chart configuration generator
go/cmd/ifres/            — CLI frequency response calculator
ts/daycounting/          — conventions.ts, daycounting.ts, fractional.ts
ts/performance/          — periodicity.ts, ratios.ts
ts/roundtrips/           — execution.ts, side.ts, matching.ts, grouping.ts, roundtrip.ts, performance.ts
ts/symbology/            — isin.ts, cusip.ts, sedol.ts
ts/entities/             — bar.ts, quote.ts, trade.ts, scalar.ts, bar-component.ts, quote-component.ts, trade-component.ts
ts/indicators/           — 63 indicators organized by author (common/, john-ehlers/, welles-wilder/, etc.)
ts/indicators/core/      — shared types: indicator.ts, line-indicator.ts, indicator-identifier.ts, indicator-metadata.ts, descriptor.ts, outputs/, frequency-response/
ts/indicators/factory/   — identifier + JSON params → indicator instance
ts/cmd/icalc/            — CLI indicator calculator
ts/cmd/iconf/            — CLI chart configuration generator
ts/cmd/ifres/            — CLI frequency response calculator
zig/src/daycounting/     — conventions.zig, daycounting.zig, fractional.zig
zig/src/performance/     — periodicity.zig, ratios.zig
zig/src/roundtrips/      — execution.zig, side.zig, matching.zig, grouping.zig, roundtrip.zig, performance.zig
zig/src/symbology/       — isin.zig, cusip.zig, sedol.zig
zig/src/entities/        — bar.zig, quote.zig, trade.zig, scalar.zig, bar_component.zig, quote_component.zig, trade_component.zig, entities.zig (barrel)
zig/src/indicators/      — 63 indicators organized by author (common/, john_ehlers/, welles_wilder/, etc.)
zig/src/indicators/core/ — shared types: indicator.zig, line_indicator.zig, identifier.zig, metadata.zig, descriptor.zig, outputs/, frequency_response/
zig/src/indicators/factory/ — identifier + JSON params → indicator instance
zig/src/cmd/icalc/       — CLI indicator calculator
zig/src/cmd/iconf/       — CLI chart configuration generator
zig/src/cmd/ifres/       — CLI frequency response calculator
zig/build.zig            — build config
rs/src/daycounting/      — conventions.rs, daycounting.rs, fractional.rs
rs/src/performance/      — periodicity.rs, ratios.rs
rs/src/roundtrips/       — mod.rs, execution.rs, side.rs, matching.rs, grouping.rs, roundtrip.rs, performance.rs
rs/src/symbology/        — mod.rs, isin.rs, cusip.rs, sedol.rs
rs/src/entities/         — mod.rs, bar.rs, quote.rs, trade.rs, scalar.rs, bar_component.rs, quote_component.rs, trade_component.rs
rs/src/indicators/       — 67 indicators organized by author (common/, john_ehlers/, welles_wilder/, etc.)
rs/src/indicators/core/  — shared types: indicator.rs, line_indicator.rs, identifier.rs, metadata.rs, descriptor.rs, outputs/, frequency_response/
rs/src/indicators/factory/ — identifier + JSON params → indicator instance
rs/src/cmd/icalc/        — CLI indicator calculator
rs/src/cmd/iconf/        — CLI chart configuration generator
rs/src/cmd/ifres/        — CLI frequency response calculator
rs/Cargo.toml            — minimal config, zero external deps
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
- Package structure: `daycounting` with `conventions` subpackage; `performance`, `roundtrips`, `symbology`, and `entities` (package name `entities`) as separate packages.

### TypeScript
- 4-space indent, strict mode in `tsconfig.json`.
- `camelCase` functions, `PascalCase` classes/enums, `UPPER_SNAKE_CASE` enum members.
- `number | null` for nullable results. `throw new Error(...)` for invalid inputs; return `null` for impossible computations.
- Jasmine 5: `describe`/`it`, `toBeCloseTo(expected, 13)`, `toBeNull()`.
- Spec naming: `.spec.ts` for all modules.

### Zig
- 4-space indent (standard Zig formatting).
- `camelCase` for all functions (public and private), `PascalCase` for types/enums/error sets, `snake_case` for struct fields, enum members, and file-level constants.
- Imports: `const std = @import("std");` first, then module imports by build.zig name (`@import("conventions")`), then type aliases.
- Optionals (`?f64`) for impossible computations; `orelse return null` to chain. Error unions (`!void`, `!f64`) for allocation failures. Pure math functions return plain `f64`.
- `ArrayList(f64)` uses the Zig 0.16 unmanaged API: init with `.empty`, pass allocator to `.append(self.allocator, item)`, `.deinit(self.allocator)`, `.appendSlice(self.allocator, items)`. `.clearRetainingCapacity()` takes no allocator.
- The `Ratios` struct stores `allocator: std.mem.Allocator` and passes it to all ArrayList operations.
- Tests live at the bottom of source files. Use `test "descriptive name" { ... }` blocks. Assertions: `try std.testing.expect(almostEqual(...))`, `try std.testing.expectEqual(expected, actual)`. Use `testing.allocator` with `defer obj.deinit()`.
- Build.zig defines 21+ modules with a dependency graph: `conventions` (no deps) -> `daycounting` -> `fractional`; `periodicity` (no deps); `ratios` (depends on all four); `execution` (depends on `fractional`); `side`, `matching`, `grouping` (no deps); `roundtrip` (depends on `execution`, `side`, `fractional`); `performance` (depends on `roundtrip`, `execution`, `side`, `fractional`); `isin`, `cusip`, `sedol` (no deps, standalone symbology modules); `bar`, `quote`, `trade`, `scalar`, `bar_component` (depends on `bar`), `quote_component` (depends on `quote`), `trade_component` (depends on `trade`) (standalone entities modules); `entities` (barrel, depends on all entity modules); `indicators` (depends on `entities`).
- Modules are registered with `b.addModule()` and tests with `b.createModule()` + `b.addTest(.{ .root_module = mod })`.

### Rust
- Standard Rust formatting (4-space indent, `rustfmt` conventions).
- `snake_case` functions, `PascalCase` types/enums, `UPPER_SNAKE_CASE` constants, `snake_case` enum variants with `#[repr(u8)]` for the Convention enum.
- `Option<f64>` for impossible computations; plain `f64` for pure math. `Result<..., String>` or `panic!` for invalid inputs.
- `DateTime` struct with fields `year: i32, month: i32, day: i32, hour: i32, minute: i32, second: i32`.
- Tests live in `#[cfg(test)] mod tests { ... }` at the bottom of source files. Helper `almost_equal(a, b, epsilon)` with `epsilon = 1e-14` (daycounting) or `1e-13` (ratios).
- Zero external dependencies. All math uses `f64` methods (`.ln()`, `.sqrt()`, `.powf()`, `.abs()`, `.cbrt()`).
- Module structure: `daycounting`, `performance`, `roundtrips`, `symbology`, `entities`, and `indicators` as submodules of the `zpano` crate, each with `mod.rs` re-exporting contents.

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

### Entities Module
Four source files per language implement financial trading data types, plus three component files for field extraction:

1. **Bar** — OHLCV candlestick: fields `time, open, high, low, close, volume`. Six computed methods: `isRising()` (close > open, strictly), `isFalling()` (close < open), `median()` (HL/2), `typical()` (HLC/3), `weighted()` (HLCC/4), `average()` (OHLC/4).
2. **Quote** — Bid/ask quote: fields `time, bid_price, ask_price, bid_size, ask_size`. Four computed methods: `mid()` ((bid+ask)/2), `weighted()` ((bid*bs+ask*as)/(bs+as)), `weightedMid()` ((bid*as+ask*bs)/(bs+as)), `spreadBp()` (20000*(ask-bid)/(ask+bid)). Zero-denominator guards return 0.0.
3. **Trade** — Single trade: fields `time, price, volume`.
4. **Scalar** — Time-value pair: fields `time, value`.
5. **BarComponent** — Enum (9 values: Open, High, Low, Close, Volume, Median, Typical, Weighted, Average). `componentValue()` returns a function/closure extracting the named component from a Bar. `componentMnemonic()` returns a short label (o, h, l, c, v, hl/2, hlc/3, hlcc/4, ohlc/4). Unknown components default to Close.
6. **QuoteComponent** — Enum (8 values: Bid, Ask, BidSize, AskSize, Mid, Weighted, WeightedMid, SpreadBp). Similar factory + mnemonic pattern. Unknown defaults to mid().
7. **TradeComponent** — Enum (2 values: Price, Volume). Similar factory + mnemonic pattern. Unknown defaults to price.

Key behavioral details:
- **No inter-entity dependencies**: Bar, Quote, Trade, Scalar are fully standalone. Component modules depend only on their parent entity type.
- **Go package name**: The `go/entities/` directory uses package name `entities`.
- **Field naming**: Go uses `Bid`/`Ask` (exported PascalCase); TS uses `bidPrice`/`askPrice` (camelCase); Python/Zig/Rust use `bid_price`/`ask_price` (snake_case).
- **Enum numbering**: Go uses `iota+1` (1-based); Python `IntEnum` starts at 0; TS numeric enum starts at 0; Zig `enum(u8)` starts at 0; Rust `#[repr(u8)]` starts at 0.
- **Component return types**: Go returns `(BarFunc, error)` for unknown components; Python/TS return closures with default fallback; Zig/Rust return function pointers with default fallback.
- **Mnemonics**: bar: o, h, l, c, v, hl/2, hlc/3, hlcc/4, ohlc/4; quote: b, a, bs, as, ba/2, (bbs+aas)/(bs+as), (bas+abs)/(bs+as), spread bp; trade: p, v. Unknown returns "??".
- **Default component constants**: Each component module exports a default: `DefaultBarComponent = Close`, `DefaultQuoteComponent = Mid`, `DefaultTradeComponent = Price`. Naming: Go `DefaultBarComponent`, TS `DefaultBarComponent`, Python `DEFAULT_BAR_COMPONENT`, Zig `default_bar_component`, Rust `DEFAULT_BAR_COMPONENT`.
- **Go `Mnemonic()` method**: Go component enums have a `Mnemonic() string` method on the enum type itself (in addition to the standalone `ComponentValue`/mnemonic functions). Unknown values return `"unknown"`. Other languages use standalone functions and return `"??"` for unknown values.

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

### Indicators Module
63 technical analysis indicators organized by author, with a shared `core/` framework. Implemented in all five languages (Go, TypeScript, Python, Zig, Rust).

See the `indicator-architecture` skill for the full design reference including:
- Folder layout and naming conventions per language
- `core/` internal structure (Indicator interface, LineIndicator base, metadata, descriptors, outputs, frequency response)
- Per-indicator file conventions (params, output enum, implementation, tests)
- Mnemonic format and component triple mnemonic
- Concurrency/lock conventions (Go)
- 72 registered identifiers, 63 concrete implementations across 19 author-based families

Key design points:
- **LineIndicator** is the base for most indicators. It routes `updateBar/Quote/Trade` to a core `update(sample)` method via stored component functions from the entities module.
- **Descriptor registry** classifies each indicator by role, pane, shape, adaptivity, input requirement, and volume usage. `BuildMetadata` pulls kind/shape from this registry.
- **Factory** maps `Identifier` + JSON params → indicator instance. Depends on all individual indicators.
- **Frequency response** computes spectral characteristics of filters. Used by spectrum indicators and the `ifres` CLI tool.
- **Component sentinel pattern**: Go uses zero-value (`if bc == 0`), TS uses `undefined`, Python uses `Optional[T]` with `None`, Zig uses `?T` with `null`, Rust uses `Option<T>` with `None`. See the `entities-architecture` skill for the cross-language mapping.

### Cmd Module
Three CLI tools exercising the indicators module. Implemented in all five languages.

1. **icalc** — Indicator calculator. Reads JSON settings, creates indicators via factory, feeds embedded 252-bar test data, prints metadata + per-bar output values.
2. **iconf** — Chart configuration generator. Same input as icalc but outputs JSON + TypeScript chart configuration with lines/bands/heatmaps for an OHLCV chart component.
3. **ifres** — Frequency response calculator. Creates indicators, detects warmup period, computes frequency response (power/amplitude/phase spectra) with signal length 1024.

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
| **Python** | Python 3.10+, `numpy`, `scipy`, symlink `ln -sf py accounts`, `touch py/__init__.py` | `PYTHONPATH=. python3 -m unittest discover -s py -p "test_*.py" -t .` | 339 tests |
| **Go** | Go 1.26+ | `cd go && go test ./...&& cd ..`| 84 packages OK |
| **TypeScript** | Node.js 20+, TypeScript 5.3+, Jasmine 5.1+ | `cd ts && npm install && npm test && cd ..` | 8935 specs |
| **Zig** | Zig 0.16.0-dev | `cd zig && zig build test --summary all && cd ..` | 1014 tests |
| **Rust** | Rust 1.75.0+ (apt) | `cd rs && cargo test && cd ..` | 985 tests |

```bash
python3 -m unittest discover -s py -p "test_*.py" -t .
cd go && go test ./...&& cd ..
cd ts && npm install && npm test && cd ..
cd zig && zig build test --summary all && cd ..
cd rs && cargo test && cd ..
```

### Build Only (no tests)

| Language | Command | Notes |
|----------|---------|-------|
| **Python** | *(interpreted — no build step)* | |
| **Go** | `cd go && go build ./...` | Compiles all packages |
| **TypeScript** | `cd ts && npm run build` | Single tsc build for all modules |
| **Zig** | `cd zig && zig build` | Build without running tests |
| **Rust** | `cd rs && cargo build` | Debug build; add `--release` for optimized |

```bash
cd go && go build ./... && cd ..
cd ts && npm run build && cd ..
cd zig && zig build && cd ..
cd rs && cargo build && cd ..
```
