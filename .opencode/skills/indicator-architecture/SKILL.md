---
name: indicator-architecture
description: Architecture, folder layout, naming conventions, and design patterns for the zpano trading indicators library. Load when creating new indicators or understanding the codebase structure.
---

# Architecture & Scaffolding Guide

This document describes the design decisions, folder layout, and naming conventions
for the **zpano** trading indicators library. It is intended as a reference for both
human developers and AI agents creating new indicators.

## Scope

The rules in this document apply **only to the `indicators/` folder** within each
language directory (e.g., `ts/indicators/`, `go/indicators/`). Other folders at the
same level as `indicators/` (such as `entities/`) have their own conventions and are
not governed by this guide.

## Core Principles

1. **Multi-language library.** The same set of indicators is implemented in
   TypeScript, Go, Python, Rust, and Zig. Each language lives in its own
   top-level folder (`ts/`, `go/`, `python/`, `rust/`, `zig/`).
2. **Shared ideology, language-idiomatic scaffolding.** The organizational
   structure and design rules are the same across all languages. File and folder
   naming adapts to each language's conventions.
3. **Consistent depth.** Every indicator lives at exactly
   `indicators/<group>/<indicator>/`. No exceptions.

## Folder Layout

```
indicators/
├── common/                 # Well-known indicators with no specific/known author
│   ├── simple-moving-average/
│   ├── exponential-moving-average/
│   └── ...
├── core/                   # Shared types, interfaces, enums, base classes
│   ├── outputs/            # Indicator output data types (band, heatmap, ...)
│   └── frequency-response/ # Frequency response calculation utilities
├── custom/                 # Your own experimental/custom indicators
│   └── my-indicator/
├── <author-name>/          # Author-attributed indicators
│   └── <indicator-name>/
│       ├── implementation
│       ├── parameters
│       └── tests
└── ...
```

### The Three Special Folders

| Folder     | Purpose |
|------------|---------|
| `core/`    | Shared foundations: types, interfaces, enums, base abstractions, utilities. |
| `common/`  | Indicators whose author is unknown or not attributed to a single person (SMA, EMA, RSI, ...). |
| `custom/`  | Indicators you develop yourself. |

These three names are **reserved** and must never be used as author names.

### Author Folders

When an indicator was developed by a known author, it is placed under a folder
named after that author. Each indicator gets its own subfolder.

Example: Mark Jurik created JMA and CFB, so:

```
indicators/
└── mark-jurik/
    ├── jurik-moving-average/
    └── composite-fractal-behavior/
```

**Go: every author folder MUST contain a `doc.go`** that declares the
parent package with a one-line godoc comment identifying the author:

```go
// Package markjurik implements indicators developed by Mark Jurik.
package markjurik
```

This is the only file at the author-folder level; all actual indicator
code lives in subpackages. The same rule applies to the special
folders `common` (shared utilities across authors) and `custom`
(zpano-specific indicators without a single author). The `core`
folder already contains real framework code and needs no `doc.go`.

TypeScript has no equivalent requirement (no package-level doc
construct).

### Indicator Folders

Each indicator folder contains:

- **Implementation** -- the indicator logic.
- **Parameters** -- an interface/struct describing the indicator's configuration.
- **Tests** -- unit tests for the indicator.
- Optionally, additional files (output types, documentation, etc.).

### Indicator Names

Use **full, descriptive names** (not short mnemonics) for indicator folders:

| Correct                          | Incorrect |
|----------------------------------|-----------|
| `simple-moving-average/`         | `sma/`    |
| `jurik-moving-average/`          | `jma/`    |
| `bollinger-bands/`               | `bb/`     |

Short names (mnemonics) may be cryptic and can overlap between different
indicators. Full names are unambiguous and self-documenting.

### `core/` Internal Structure

```
core/
├── outputs/              # Output data types (band, heatmap, ...)
└── frequency-response/   # Frequency response calculation utilities
```

- **`outputs/`** contains indicator output shapes. An indicator can output a
  scalar, a band (upper/lower), a heatmap, etc. This folder holds the concrete
  types; the taxonomy enum lives in the `outputs/shape/` sub-package as
  `shape.Shape` (Go) / `Shape` (TS).
- **`frequency-response/`** contains utilities for computing the frequency
  response of a filter/indicator.

Other shared types (indicator interface, metadata, specification, line indicator
base class, indicator type enum) live directly in `core/`.

## Naming Conventions per Language

The structure is identical across languages; only the naming style changes.

| Language   | Folder naming       | Example author folder    | Example indicator folder      |
|------------|---------------------|--------------------------|-------------------------------|
| TypeScript | `kebab-case`        | `mark-jurik/`            | `jurik-moving-average/`       |
| Go         | `lowercase` (no separators) | `markjurik/`      | `jurikmovingaverage/`         |
| Python     | `snake_case`        | `mark_jurik/`            | `jurik_moving_average/`       |
| Rust       | `snake_case`        | `mark_jurik/`            | `jurik_moving_average/`       |
| Zig        | `snake_case`        | `mark_jurik/`            | `jurik_moving_average/`       |

### Symbol Naming (Types, Enums)

Go uses package-scoped names, so symbols can be shorter (the package provides
context). All other languages use fully-qualified symbol names.

| Concept           | Go (in `core` pkg)  | Go (in `outputs/shape` pkg) | TypeScript / Python / Rust / Zig |
|-------------------|---------------------|-----------------------------|----------------------------------|
| Indicator identity| `core.Identifier`   | --                          | `IndicatorIdentifier`            |
| Output shape      | --                  | `shape.Shape`               | `Shape`                          |

### Identifier Registry Asymmetry

Go and TypeScript do not always have the same number of registered identifiers.
As of writing, Go has **72** `core.Identifier` constants (iota 1–72) while TS
has **65** `IndicatorIdentifier` enum values (0-based). The gap exists because
some indicators have been converted in Go but not yet in TS (or vice versa).
When adding a new indicator, register the identifier in **both** languages even
if only one implementation exists yet — this keeps the registries aligned.

**KAMA identifier quirk:** The Go JSON string for KAMA is
`"kaufmanAdaptiveMovingAverageMovingAverage"` (with "MovingAverage" duplicated)
because `core/identifier.go` defines it that way. The TS enum is just
`KaufmanAdaptiveMovingAverage` (no duplicate). Any tooling that maps between
the two (e.g., the TS `icalc` CLI) must handle this mismatch explicitly.

### File Naming

| Language   | Style                     | Test files                  | Example                          |
|------------|---------------------------|-----------------------------|----------------------------------|
| TypeScript | `kebab-case.ts`           | `kebab-case.spec.ts`        | `simple-moving-average.ts`       |
| Go         | `lowercase.go`            | `lowercase_test.go`         | `simplemovingaverage.go`         |
| Python     | `snake_case.py`           | `test_snake_case.py`        | `simple_moving_average.py`       |
| Rust       | `snake_case.rs`           | `snake_case_test.rs` or inline `#[cfg(test)]` | `simple_moving_average.rs` |
| Zig        | `snake_case.zig`          | `snake_case_test.zig`       | `simple_moving_average.zig`      |

No type-suffix convention is used (no `.enum.ts`, `.interface.ts`, etc.).
This is consistent across all languages.

#### Main vs. auxiliary files

Within an indicator folder, **only the main implementation file and its test
keep the indicator name**. Every auxiliary file drops the indicator-name
prefix — the folder (and in Go, the package name) already provides context,
so prefixing would stutter.

For a `simple-moving-average/` folder:

| Role                         | Go                              | TypeScript                              |
|------------------------------|---------------------------------|-----------------------------------------|
| Main implementation          | `simplemovingaverage.go`        | `simple-moving-average.ts`              |
| Main test                    | `simplemovingaverage_test.go`   | `simple-moving-average.spec.ts`         |
| Data-driven test (if split)  | `data_test.go`                  | `data.spec.ts`                          |
| Parameters                   | `params.go`                     | `params.ts`                             |
| Output enum                  | `output.go`                     | `output.ts`                             |
| Output test                  | `output_test.go`                | `output.spec.ts`                        |
| Coefficients / helpers       | `coefficients.go`, `estimator.go` | `coefficients.ts`, `estimator.ts`     |
| Package docs (Go only)       | `doc.go`                        | —                                       |

**Exception:** when an auxiliary file holds a *concept* that would be
meaningless outside the folder name (e.g., `hilberttransformer/` contains
`cycleestimator.go`, `phaseaccumulator.go`, `dualdifferentiator.go` — each
representing a distinct Hilbert-transformer variant), keep the conceptual
names as-is. Do not try to mechanically strip a prefix that was never there.

### Identifier Abbreviation Convention

To keep naming consistent across indicators and languages, prefer the **full
word** in identifiers. The following abbreviations are considered **banned**
as CamelCase word-parts (whether standalone or compound) — always use the
long form:

| Banned short | Canonical long | Example (before → after)                |
|--------------|----------------|------------------------------------------|
| `idx`        | `index`        | `bufIdx` → `bufferIndex`, `highestIdx` → `highestIndex` |
| `tmp`        | `temp`         | `tmpR` → `tempR`                         |
| `res`        | `result`       | `res` → `result`, `defSpectrumRes` → `defSpectrumResult` |
| `sig`        | `signal`       | `expSig` → `expSignal`, `mnemonicSig` → `mnemonicSignal` |
| `val`        | `value`        | `mnemonicVal` → `mnemonicValue`, `prevVal` → `previousValue` |
| `prev`       | `previous`     | `prevHigh` → `previousHigh`, `samplePrev` → `samplePrevious` |
| `avg`        | `average`      | `AvgLength` → `AverageLength`            |
| `mult`       | `multiplier`   | `upperMult` → `upperMultiplier`          |
| `buf`        | `buffer`       | *(compound only, paired with `idx`)*     |
| `param`      | `parameter`    | `paramResolution` → `parameterResolution` *(in identifiers; Go struct type `Params` stays)* |
| `hist`       | `histogram`    | `expectedHist` → `expectedHistogram`     |

**Allowed exceptions (retained as-is):**

- **Go language idioms:** `err`, `len`, `cap`, `min`, `max`, `num` — these are
  builtins or deeply idiomatic Go short forms. Do not expand.
- **Go struct field `Params` / variable `params`** — this is the established
  parameter-bundle type name across every indicator (`type Params struct`,
  `Compute(params Params)`). The long form `Parameters` is used **only** in
  `core/specification.go`/`core/indicator-specification.ts` where it
  describes framework-level metadata.
- **TS constructor-parameter idiom `value_`** (trailing underscore) — used
  to avoid shadowing a member `value` during construction.
- **Domain-specific short forms that happen to collide** (e.g., `pos` as a
  loop variable is fine even though an unrelated indicator may use
  `position` in its name). These are different concepts, not abbreviation
  drift.
- **TA-Lib port conventions** (`begIdx` → now `begIndex`, `outIdx` → now
  `outIndex`): the `Idx` part is normalized but the `beg`/`out` prefixes
  are kept, matching TA-Lib's C-source variable names for easier
  cross-reference.

**Go↔TS parity:** the same identifier stem MUST be used in both languages.
When porting, do not introduce new short forms; use the canonical long form
from the table above.

### Go Receiver Naming Convention

Method receivers follow a **type-shape rule**, not a one-letter-for-all rule:

- **Compound type name** (2+ CamelCase words, e.g. `SimpleMovingAverage`,
  `BarComponent`, `CycleEstimatorType`) → receiver is **`s`**
  ("self"/"struct"). Short, consistent across all large types.
- **Simple type name** (single word, e.g. `Momentum`, `Trade`, `Quote`) →
  receiver is the **first letter of the type name, lowercased** (`m`, `t`,
  `q`). Short types deserve short mnemonic receivers.

All methods on a given type MUST use the same receiver name (enforced by
`go vet`/consistency). When adding a method, match the receiver already in
use on that type.

**Local-variable collision:** if a method body needs a local variable that
would shadow the receiver `s` (typical case: `s := someValue.String()` in
`MarshalJSON`), rename the **local** to `str` (not the receiver). Example:

```go
func (s BarComponent) MarshalJSON() ([]byte, error) {
    str := s.String()           // local renamed s → str
    if str == unknown { ... }
}
```

This keeps the receiver convention uniform; local names are flexible.

### Concurrency / Lock Convention

Stateful indicators (those exposing `Update(...)` that mutates internal
fields) MUST guard reads and writes with a `sync.RWMutex`:

- **Field:** `mu sync.RWMutex` (always named `mu`, always `RWMutex` —
  never plain `sync.Mutex`).
- **Writer methods** (`Update`, any mutator): acquire with
  `s.mu.Lock(); defer s.mu.Unlock()` on the two lines immediately
  following the function-signature brace (before any other logic).
- **Reader methods** (`IsPrimed`, accessors that read mutable state):
  use `s.mu.RLock(); defer s.mu.RUnlock()`.
- **Never** unlock manually (non-deferred). Always pair Lock with
  `defer Unlock` on the very next line.

**Exceptions — indicators without a mutex:**

1. **Pure delegating wrappers** that own no mutable state of their own
   and forward all `Update`/read calls to an embedded indicator which
   itself carries a mutex (e.g. `standarddeviation` wrapping
   `variance`). The embedded indicator provides the lock.
2. **Internal engines** that are not part of the public indicator
   surface and are only consumed by higher-level wrappers which hold
   the mutex (e.g. `corona` engine used by `coronaswingposition`,
   `coronaspectrum`, `coronasignaltonoiseratio`, `coronatrendvigor`).

Any new stateful public indicator MUST either carry its own `mu
sync.RWMutex` following the pattern above or fall under one of the two
exceptions.

### Go Style Invariants (enforced across the codebase)

The following idioms are invariants — audits return **zero** deviations
and new code MUST respect them:

- **Variable declarations.** No `var x T = zero-literal`, no split
  `var x T` followed immediately by `x = expr`. Use either `var x T`
  (zero value) or `x := expr`.
- **Empty / nil tests.** Slice/map emptiness is not tested by
  `len(x) == 0`; prefer direct ranging or `== nil` where the value is
  nil-able. Error checks always use `err != nil`.
- **Any.** Use `any`, never `interface{}`.
- **Slice construction.** Use `[]T{}` or `make([]T, n)` /
  `make([]T, 0, n)` — never `make([]T, 0)` (no capacity hint).
- **`new`.** Not used; use composite literals (`&T{...}`).
- **Imports.** Grouped by class with blank lines: stdlib, external,
  `zpano/*`. Each group contains only one class.
- **Doc comments.** Every exported `func`/`type`/`var`/`const`/method
  has a leading `// Name ...` doc comment. Trivial stubs (e.g.
  `passthrough.Update`) still get a one-line comment.

### Go Params: No JSON Tags

Go indicator params structs do **not** carry `json:"..."` struct tags. Go's
`encoding/json` unmarshaler performs **case-insensitive** key matching by
default, so JSON `"smoothingFactor"` matches Go field `SmoothingFactor`
without explicit tags. Do not add JSON tags to params structs — they are
unnecessary and would diverge from the established pattern.

### TypeScript Import Conventions

1. **No barrel files.** The TS indicators codebase does not use `index.ts`
   barrel re-exports. All imports use **direct file paths** to the specific
   module (e.g., `import { Foo } from '../common/simple-moving-average/simple-moving-average.js'`).

2. **`.js` extensions in import paths.** TypeScript source files use `.js`
   extensions in import specifiers (not `.ts`). This is required by the ESM
   module resolution used in the project (`"type": "module"` in
   `package.json`, `tsx` loader for tests/CLI).

3. **Renamed imports for disambiguation.** When multiple indicators export
   the same symbol name (e.g., every params file exports `defaultParams`),
   use renamed imports:
   ```typescript
   import { defaultParams as defaultSmaParams } from '../common/simple-moving-average/params.js';
   import { defaultParams as defaultEmaLengthParams } from '../common/exponential-moving-average/length-params.js';
   ```
   This pattern is used extensively in the factory.

### Go ↔ TypeScript Local-Variable Parity

The same indicator MUST use the **same local/field names** in both
languages. When porting, copy the other language's names verbatim
where semantically identical. Observed canonical vocabulary:

| Concept                    | Canonical local name |
|----------------------------|----------------------|
| Sample difference / delta  | `temp` (short-lived delta) / `diff` (explicit diff) |
| Accumulator sum            | `sum` (never `total`, `acc`, `accumulator`) |
| Small-comparison tolerance | `epsilon` |
| Running standard deviation | `stddev` |
| MACD fast/slow/signal line | `fast`, `slow`, `signal`, `macd`, `histogram` |
| Variance (scratch)         | `v` |
| NaN sentinel (Go)          | `nan := math.NaN()` |
| Spread (upper-lower)       | `spread` |
| Bollinger band-width       | `bw` *(local; field is `bandWidth`)* |
| Bollinger percent-band     | `pctB` *(local; field is `percentBand`)* |
| Length minus one           | `lengthMinOne` |
| Amount = price × volume    | `amount` |

Primary loop counter is `i`; secondary `j`, `k`; only use `index`
when the counter semantically labels a named index (e.g. filter-bank
bin index). Never introduce new short forms (`idx`, `res`, etc. are
banned per the Abbreviation Convention).

## Adding a New Indicator -- Checklist

1. **Determine the group.** Is the author known? Use `<author-name>/`. Unknown
   author? Use `common/`. Your own? Use `custom/`.
2. **Create the indicator subfolder** at `indicators/<group>/<indicator-name>/`
   using full descriptive names and language-appropriate casing.
3. **Create the required files:**
   - **Parameters** -- define a params struct with a `DefaultParams()` /
     `defaultParams()` function (see the DefaultParams section below). Where the indicator can be
     driven by different price components, include optional `BarComponent`,
     `QuoteComponent`, `TradeComponent` fields (zero = default in Go,
     `undefined` = default in TS). Document the defaults in field comments.
     For indicators that require fixed inputs (e.g., HLCV for Advance-Decline,
     OHLC for Balance of Power, high/low for Parabolic SAR), the params struct
     may be empty or contain only non-component parameters (length, thresholds).
   - **Output enum** -- define a per-indicator output enum. In **Go**, the type
     is bare `Output` and constants are bare concept names (package name
     provides scoping — `simplemovingaverage.Output`, `simplemovingaverage.Value`).
     In **TypeScript**, the enum is named `<IndicatorName>Output` (e.g.,
     `SimpleMovingAverageOutput`) because TS imports by symbol, not by module.
   - **Implementation** -- embed `core.LineIndicator` (Go) or extend
     `LineIndicator` (TS). In the constructor:
     1. Resolve zero-value components → defaults for the component functions.
        In Go, check `if bc == 0 { bc = DefaultBarComponent }`.
        In TS, set the protected setters (`this.barComponent = params.barComponent`),
        which resolve `undefined` → default internally.
     2. Build the mnemonic using `ComponentTripleMnemonic` / `componentTripleMnemonic`
        with the **resolved** values (Go) or raw param values (TS — the function
        checks `!== undefined && !== Default*` itself). **Exception:** indicators
        with no component params (empty or non-component-only Params) hand-write
        their mnemonic (e.g., `bop()`, `ad()`, `psar(0.02,0.2)`) — the mnemonic
        still includes non-component parameters when non-default.
     3. Initialize the `LineIndicator` with mnemonic, description, component
        functions, and the indicator's `Update` method.
   - **`Metadata()`** -- return `core.BuildMetadata(core.<Identifier>, mnemonic, description, []core.OutputText{...})` (Go) or
     `buildMetadata(IndicatorIdentifier.<Name>, mnemonic, description, [...])` (TS). The helper sources each output's `kind` and
     `shape` from the descriptor registry — the caller supplies only per-output mnemonic/description. See the **Taxonomy &
     Descriptor Registry** section below for details and the mandatory descriptor-row step.
   - **Test file** -- include mnemonic sub-tests covering: all components zero,
     each component set individually, and combinations of two components.
4. **Register the indicator** in `core/identifier` (add a new enum variant in both Go `core.Identifier` and TS `IndicatorIdentifier`).
5. **Register the descriptor** in `core/descriptors.{go,ts}` — see the Taxonomy section below. A missing descriptor row causes
   `BuildMetadata` to panic at runtime.
6. **Follow the consistent depth rule** -- the indicator must be exactly two
   levels below `indicators/`.

## Design Decisions Log

| Decision | Rationale |
|----------|-----------|
| No underscore prefix on `core/`, `common/`, `custom/` | Leading underscores conflict with conventions in Python (`_` = private) and are non-idiomatic in Go. Dropping them gives true consistency across all five languages. The names are unambiguous enough on their own. |
| Full indicator names, not mnemonics | Short names (SMA, JMA) can be cryptic and may overlap. Full names are self-documenting. |
| `outputs/` instead of `entities/` in core | These types represent indicator *output shapes*, not general domain entities. `outputs` is more precise about their purpose. |
| `core/` instead of `types/` | `core` is broad enough to hold types, enums, base classes, and utilities without needing a rename later. `types` would be too narrow. |
| `common/` instead of `unknown/`, `classic/`, `standard/` | `common` is neutral -- doesn't imply age (`classic`), official status (`standard`), or sound dismissive (`unknown`). |
| `custom/` instead of `experimental/` | `custom` describes *origin* (you made it), not *maturity*. A custom indicator can be stable; an experimental one implies instability. If you need to track maturity, use metadata in code, not folder structure. |
| Go uses scoped names (`core.Identifier`, `shape.Shape`); other languages use `IndicatorIdentifier`/`Shape` at the top level | Go's package system provides scoping, making prefixes redundant and stuttery (`core.IndicatorIdentifier`, `outputs.Shape`). Other languages import symbols directly, so self-describing names are necessary; `Shape` is scoped to its own `outputs/shape` sub-package in Go to avoid clashing with entity/indicator shapes. |
| Per-indicator subfolders even in `common/` | Without subfolders, all indicator files in `common/` would land in one flat directory (and in Go, one package). This doesn't scale and is inconsistent with the author-folder structure. |
| Component enums start at `iota + 1` | Makes the zero value (`0`) explicitly "not set", which is used as the sentinel for "use default". This is idiomatic Go for optional enum fields in structs (zero value of `int` is `0`). |
| TypeScript enums start at `0`, components are `undefined` when not set | TS enums default to `0`, and optional params use `undefined` (not `0`) as the sentinel for "not set". The `componentTripleMnemonic` function checks `!== undefined && !== Default*` to decide whether to show a component. The `LineIndicator` setters use `=== undefined` (not truthiness) to decide when to apply the default, so explicitly passing a zero-valued enum (e.g., `BarComponent.Open`) works correctly. |
| Zero-value component = default, omitted from mnemonic | Keeps the common case clean (`sma(14)` instead of `sma(14, c, ba/2, p)`) while still allowing explicit overrides when needed. `ComponentTripleMnemonic` checks against `Default*` constants (not zero), so explicitly passing the default value also omits it — the same indicator always gets the same mnemonic. |
| Entity `Mnemonic()` method separate from `String()` | `String()` returns full words (`"close"`, `"mid"`) for JSON serialization and debugging. `Mnemonic()` returns short codes (`"c"`, `"ba/2"`) for compact display in chart labels. Separating them avoids overloading `String()` with a display concern. |
| Mnemonic sub-tests for every indicator | Mnemonic correctness matters for charting UIs. Each indicator must have sub-tests covering: all components zero, each component individually, and pairwise combinations. This prevents regressions when changing the format string or `ComponentTripleMnemonic`. |

## LineIndicator Design Pattern

A **line indicator** takes a single numeric input and produces a single scalar
output. Most indicators (SMA, EMA, JMA, ...) are line indicators. The
`LineIndicator` abstraction eliminates boilerplate by providing
`UpdateScalar/Bar/Quote/Trade` methods (and array variants) that delegate to the
indicator's core `Update(sample) → value` function.

### What Each Indicator Implements

Each concrete indicator is responsible for:

1. **`Update(sample float64) float64`** (Go) / **`update(sample: number): number`** (TS) --
   the core calculation logic.
2. **`Metadata()`** / **`metadata()`** -- returns indicator-level metadata with
   an explicit per-indicator output enum (not a hardcoded `kind: 0`).
3. **`IsPrimed()`** / **`isPrimed()`** -- whether the indicator has received
   enough data to produce meaningful output.

### What LineIndicator Provides (Eliminating Boilerplate)

The `LineIndicator` base provides:

- **`UpdateScalar/Bar/Quote/Trade`** -- all delegate to the indicator's `Update()`
  method via a function reference (Go) or abstract method (TS). The bar/quote/trade
  variants extract the relevant component value first.
- **Array update methods** -- `UpdateScalars/Bars/Quotes/Trades` (TS only;
  Go has equivalent free functions in `core/indicator.go`).
- **Storage of component functions** -- bar, quote, and trade component
  extraction functions are stored and used by the Update* methods.

### Go Implementation

In Go, `LineIndicator` is an **embedded struct**. The concrete indicator stores a
`core.LineIndicator` field and initializes it via `core.NewLineIndicator()`,
passing its own `Update` method as a function reference:

```go
type SimpleMovingAverage struct {
    mu sync.RWMutex
    core.LineIndicator   // embedded -- promotes UpdateScalar/Bar/Quote/Trade
    // ... indicator-specific fields
}

func NewSimpleMovingAverage(p *Params) (*SimpleMovingAverage, error) {
    sma := &SimpleMovingAverage{ /* ... */ }
    sma.LineIndicator = core.NewLineIndicator(
        mnemonic, desc, barFunc, quoteFunc, tradeFunc, sma.Update,
    )
    return sma, nil
}
```

The mutex lives on the concrete indicator (not on `LineIndicator`), and is
acquired inside `Update()`. The promoted `UpdateScalar/Bar/Quote/Trade` methods
call `updateFn` (which is the concrete indicator's `Update`), so the threading
model is preserved.

### TypeScript Implementation

In TypeScript, `LineIndicator` is an **abstract class**. The concrete indicator
extends it and implements `update()` and `metadata()`:

```typescript
export abstract class LineIndicator implements Indicator {
    protected mnemonic!: string;
    protected description!: string;
    protected primed!: boolean;

    public abstract metadata(): IndicatorMetadata;
    public abstract update(sample: number): number;
    // updateScalar/Bar/Quote/Trade provided by base class
}
```

**Concrete indicator constructors must set the three component setters** to
initialize the component functions used by `updateBar()`, `updateQuote()`, and
`updateTrade()`:

```typescript
this.barComponent = params.barComponent;       // undefined → DefaultBarComponent
this.quoteComponent = params.quoteComponent;   // undefined → DefaultQuoteComponent
this.tradeComponent = params.tradeComponent;   // undefined → DefaultTradeComponent
```

The `LineIndicator` setters use `=== undefined` (not truthiness) to apply
defaults, so explicitly passing a zero-valued enum like `BarComponent.Open`
works correctly.

Each entity file also exports a `Default*` constant (`DefaultBarComponent`,
`DefaultQuoteComponent`, `DefaultTradeComponent`) for use in
`componentTripleMnemonic` and anywhere else defaults need to be checked.

### Per-Indicator Output Enums

Each indicator defines its own output enum describing what it produces. The
naming convention is **language-asymmetric** because the two languages scope
symbols differently:

- **Go**: use bare `Output` / `Value`. The package name
  (`simplemovingaverage.Output`, `simplemovingaverage.Value`) provides scoping,
  so repeating the indicator name in the type would stutter.
- **TypeScript**: use long-form `<IndicatorName>Output` /
  `<IndicatorName>Value`. TS imports by symbol rather than by module, so the
  indicator name must be baked into the identifier to stay unambiguous at
  call sites.

```go
// Go — file: simplemovingaverage/output.go
package simplemovingaverage

type Output int
const (
    Value Output = iota
)
```

```typescript
// TypeScript — file: simple-moving-average/output.ts
export enum SimpleMovingAverageOutput {
    SimpleMovingAverageValue = 0,
}
```

For multi-output indicators, strip the indicator-name prefix on the Go side
and keep the descriptive suffix:

```go
// Go — directionalmovementindex/output.go
type Output int
const (
    PlusDirectionalIndicator Output = iota
    MinusDirectionalIndicator
    AverageDirectionalIndex
    // ...
)
```

```typescript
// TypeScript — directional-movement-index/output.ts
export enum DirectionalMovementIndexOutput {
    PlusDirectionalIndicator = 0,
    MinusDirectionalIndicator = 1,
    AverageDirectionalIndex = 2,
    // ...
}
```

The `metadata()` method uses this enum for the `kind` field instead of a
hardcoded `0`, making the output semantics explicit and type-safe.

### Metadata Structure

The `Metadata` / `IndicatorMetadata` type includes:

| Field         | Description |
|---------------|-------------|
| `identifier`  | The indicator identifier enum variant (e.g., `SimpleMovingAverage`). |
| `mnemonic`    | A short name like `sma(5)` or `jma(7, -1)`. |
| `description` | A human-readable description like `Simple moving average sma(5)`. |
| `outputs`     | An array of per-output metadata (kind, shape, mnemonic, description). |

The `mnemonic` and `description` live at the indicator level (on `Metadata`)
and are also duplicated per-output for convenience (each output carries its own
mnemonic/description).

Each output carries:

| Field         | Description |
|---------------|-------------|
| `kind`        | The integer value of the indicator's per-output enum. |
| `shape`       | The output's data shape (`Scalar`, `Band`, `Heatmap`, `Polyline`). Sourced from the descriptor registry. |
| `mnemonic`    | Short name for this output. |
| `description` | Human-readable description of this output. |

In practice, **indicators never construct `Metadata` object literals directly**.
They call `core.BuildMetadata` (Go) / `buildMetadata` (TS), which pulls `kind`
and `shape` from the descriptor registry. See the Taxonomy section below.

### Mnemonic Prefix Convention

Every indicator is identified by a short lowercase **mnemonic prefix** (e.g.
`sma`, `rsi`, `bb`, `macd`). These prefixes must be consistent between Go and
TypeScript, because downstream tooling keys off them.

**Rules:**

1. **Lowercase only.** Digits are allowed (`t2`, `t3`, `rocr100`).
   Leading `+`/`-` is allowed for paired directional indicators (`+di`, `-di`,
   `+dm`, `-dm`).
2. **Kept short** — typically 2-7 chars. The mnemonic is a compact label, not a
   full name; the full name goes in `description`.
3. **Go ↔ TS parity.** The same prefix must appear on both sides. When adding
   an indicator, add its prefix to both languages in the same PR.
4. **Format.** `<prefix>(<param1>, <param2>, ...<componentSuffix>)` where the
   component suffix comes from `ComponentTripleMnemonic` (empty when all
   components are at their default).
5. **Configurable components → include the suffix.** Any indicator that
   exposes `BarComponent` / `QuoteComponent` / `TradeComponent` on its params
   struct **must** append `ComponentTripleMnemonic(bc, qc, tc)` to its
   mnemonic. Indicators with hardcoded components (e.g. balance-of-power)
   omit the suffix.
6. **Parameterless indicators** use the bare prefix (`obv`, `ad`, `bop`) when
   defaults are in effect, or wrap the component suffix in parens when needed
   (e.g. `obv(hl/2)`).

**Canonical inventory (65 indicators with mnemonics, plus 4 paired directional):**

| Prefix    | Indicator                                            |
|-----------|------------------------------------------------------|
| `aci`     | autocorrelation indicator                            |
| `acp`     | autocorrelation periodogram                          |
| `ad`      | advance/decline                                      |
| `adosc`   | advance/decline oscillator                           |
| `adx`     | average directional movement index                   |
| `adxr`    | average directional movement index rating            |
| `apo`     | absolute price oscillator                            |
| `aroon`   | Aroon                                                |
| `atcf`    | adaptive trend and cycle filter                      |
| `atr`     | average true range                                   |
| `bb`      | Bollinger bands                                      |
| `bbtrend` | Bollinger bands trend                                |
| `bop`     | balance of power                                     |
| `cbps`    | comb band-pass spectrum                              |
| `cc`      | cyber cycle                                          |
| `cci`     | commodity channel index                              |
| `cmo`     | Chande momentum oscillator                           |
| `cog`     | center-of-gravity oscillator                         |
| `correl`  | Pearson's correlation coefficient                    |
| `csnr`    | corona signal-to-noise ratio                         |
| `cspect`  | corona spectrum                                      |
| `cswing`  | corona swing position                                |
| `ctv`     | corona trend vigor                                   |
| `dcp`     | dominant cycle                                       |
| `dema`    | double exponential moving average                    |
| `dftps`   | discrete Fourier transform power spectrum            |
| `+di`/`-di` | directional indicator plus / minus                 |
| `+dm`/`-dm` | directional movement plus / minus                  |
| `dx`      | directional movement index                           |
| `ema`     | exponential moving average                           |
| `frama`   | fractal adaptive moving average                      |
| `gspect`  | Goertzel spectrum                                    |
| `htitl`   | Hilbert transformer instantaneous trend line         |
| `iTrend`  | instantaneous trend line                             |
| `jma`     | Jurik moving average                                 |
| `kama`    | Kaufman adaptive moving average                      |
| `linreg`  | linear regression                                    |
| `macd`    | moving average convergence/divergence                |
| `mama`    | MESA adaptive moving average                         |
| `mespect` | maximum entropy spectrum                             |
| `mfi`     | money flow index                                     |
| `mom`     | momentum                                             |
| `natr`    | normalized average true range                        |
| `obv`     | on-balance volume                                    |
| `ppo`     | percentage price oscillator                          |
| `roc`     | rate of change                                       |
| `rocp`    | rate of change percent                               |
| `rocr`/`rocr100` | rate of change ratio (unit or ×100)           |
| `roof`    | roofing filter                                       |
| `rsi`     | relative strength index                              |
| `sar`     | parabolic stop and reverse                           |
| `sma`     | simple moving average                                |
| `ss`      | super smoother                                       |
| `stdev`   | standard deviation (`.s` sample / `.p` population)   |
| `stoch`   | stochastic                                           |
| `stochrsi`| stochastic relative strength index                   |
| `sw`      | sine wave                                            |
| `tcm`     | trend cycle mode                                     |
| `tema`    | triple exponential moving average                    |
| `tr`      | true range                                           |
| `trima`   | triangular moving average                            |
| `trix`    | triple exponential moving average oscillator         |
| `t2`      | T2 exponential moving average                        |
| `t3`      | T3 exponential moving average                        |
| `ultosc`  | ultimate oscillator                                  |
| `var`     | variance (`.s` sample / `.p` population)             |
| `willr`   | Williams %R                                          |
| `wma`     | weighted moving average                              |
| `zecema`  | zero-lag error-correcting exponential moving average |
| `zema`    | zero-lag exponential moving average                  |

The two utility modules `corona/` and `frequencyresponse/` are support helpers,
not indicators, and have no mnemonic of their own.

### Component Triple Mnemonic

A line indicator can be fed bar, quote, or trade data. Each data type has a
**component** that selects which field to extract (e.g., close price from a bar,
mid-price from a quote). The `ComponentTripleMnemonic` utility (in `core/`)
generates the component suffix for an indicator's mnemonic. There are three
components (bar, quote, trade) -- hence "triple", not "pair".

#### Default Components

Each component type has a default constant defined in `entities/`:

| Component type   | Default constant         | Default value    | Mnemonic |
|------------------|--------------------------|------------------|----------|
| `BarComponent`   | `DefaultBarComponent`    | `BarClosePrice`  | `c`      |
| `QuoteComponent` | `DefaultQuoteComponent`  | `QuoteMidPrice`  | `ba/2`   |
| `TradeComponent` | `DefaultTradeComponent`  | `TradePrice`     | `p`      |

#### Zero-Value Convention

All component enums start at `iota + 1`, making the zero value `0` explicitly
"not set". In indicator params, **a zero-value component means "use the default
and omit from the mnemonic"**:

```go
// SimpleMovingAverageParams
type SimpleMovingAverageParams struct {
    Length         int
    BarComponent   entities.BarComponent   // zero → DefaultBarComponent, omitted from mnemonic
    QuoteComponent entities.QuoteComponent // zero → DefaultQuoteComponent, omitted from mnemonic
    TradeComponent entities.TradeComponent // zero → DefaultTradeComponent, omitted from mnemonic
}
```

#### Constructor Resolution Logic

Each indicator constructor follows this pattern:

1. Read the component from params.
2. If zero, resolve to the default constant.
3. Get the component function using the resolved value.
4. Pass the **resolved** values to `ComponentTripleMnemonic` (which omits
   defaults automatically).

```go
// Resolve defaults for component functions.
bc := p.BarComponent
if bc == 0 {
    bc = entities.DefaultBarComponent
}
// ... same for qc, tc

// Build mnemonic using resolved components — defaults are omitted by ComponentTripleMnemonic.
mnemonic := fmt.Sprintf("sma(%d%s)", length,
    core.ComponentTripleMnemonic(bc, qc, tc))
```

This means that `sma(14)` is the mnemonic whether the user left `BarComponent`
as zero or explicitly set it to `BarClosePrice` — both produce the same
indicator with the same mnemonic.

#### ComponentTripleMnemonic Function

The function compares each component against its `Default*` constant. If a
component equals its default, it is omitted. Non-default components are
appended as `", "` followed by the component's `Mnemonic()` output:

```go
func ComponentTripleMnemonic(
    bc entities.BarComponent, qc entities.QuoteComponent, tc entities.TradeComponent,
) string {
    var s string
    if bc != entities.DefaultBarComponent   { s += ", " + bc.Mnemonic() }
    if qc != entities.DefaultQuoteComponent { s += ", " + qc.Mnemonic() }
    if tc != entities.DefaultTradeComponent { s += ", " + tc.Mnemonic() }
    return s
}
```

The result is inserted into the indicator format string after the numeric
parameters: `sma(14%s)` → `sma(14)` or `sma(14, hl/2)`.

#### Mnemonic Examples

| Params                                                                | Mnemonic              |
|-----------------------------------------------------------------------|-----------------------|
| `{Length: 5}`                                                         | `sma(5)`              |
| `{Length: 5, BarComponent: BarMedianPrice}`                           | `sma(5, hl/2)`        |
| `{Length: 5, QuoteComponent: QuoteBidPrice}`                          | `sma(5, b)`           |
| `{Length: 5, TradeComponent: TradeVolume}`                            | `sma(5, v)`           |
| `{Length: 5, BarComponent: BarOpenPrice, QuoteComponent: QuoteBidPrice}` | `sma(5, o, b)`     |
| `{Length: 5, BarComponent: BarHighPrice, TradeComponent: TradeVolume}` | `sma(5, h, v)`       |
| `{Length: 7, Phase: 0}`                                               | `jma(7, 0)`           |
| `{Length: 7, Phase: 0, BarComponent: BarMedianPrice}`                 | `jma(7, 0, hl/2)`    |

The `description` is always the full name followed by the mnemonic:
`"Simple moving average sma(5, hl/2)"`.

#### Entity Mnemonic Methods

Each component enum has two display methods: `String()` (full word, used for
JSON serialization) and `Mnemonic()` (short code, used in indicator mnemonics).

**BarComponent mnemonics:**

| Enum value        | `String()`   | `Mnemonic()` |
|-------------------|-------------|---------------|
| `BarOpenPrice`    | `open`      | `o`           |
| `BarHighPrice`    | `high`      | `h`           |
| `BarLowPrice`     | `low`       | `l`           |
| `BarClosePrice`   | `close`     | `c`           |
| `BarVolume`       | `volume`    | `v`           |
| `BarMedianPrice`  | `median`    | `hl/2`        |
| `BarTypicalPrice` | `typical`   | `hlc/3`       |
| `BarWeightedPrice`| `weighted`  | `hlcc/4`      |
| `BarAveragePrice` | `average`   | `ohlc/4`      |

**QuoteComponent mnemonics:**

| Enum value             | `String()`     | `Mnemonic()`          |
|------------------------|----------------|-----------------------|
| `QuoteBidPrice`        | `bid`          | `b`                   |
| `QuoteAskPrice`        | `ask`          | `a`                   |
| `QuoteBidSize`         | `bidSize`      | `bs`                  |
| `QuoteAskSize`         | `askSize`      | `as`                  |
| `QuoteMidPrice`        | `mid`          | `ba/2`                |
| `QuoteWeightedPrice`   | `weighted`     | `(bbs+aas)/(bs+as)`   |
| `QuoteWeightedMidPrice`| `weightedMid`  | `(bas+abs)/(bs+as)`   |
| `QuoteSpreadBp`        | `spreadBp`     | `spread bp`           |

**TradeComponent mnemonics:**

| Enum value    | `String()`  | `Mnemonic()` |
|---------------|-------------|---------------|
| `TradePrice`  | `price`     | `p`           |
| `TradeVolume` | `volume`    | `v`           |

## Helper / Shared Component Families

Some indicators are driven by an interchangeable internal component that has
several variants (e.g., different algorithms producing the same shape of
output). The canonical example is John Ehlers' **Hilbert Transformer cycle
estimator**, which has four variants (HomodyneDiscriminator,
HomodyneDiscriminatorUnrolled, PhaseAccumulator, DualDifferentiator) that all
produce a cycle period estimate from a price sample.

These are **not indicators themselves**. They are helper components used by
actual indicators (e.g., MESA Adaptive Moving Average). They live in the
indicator folder of their conceptual owner and follow a specific structure.

### Folder Layout

A helper family lives in a single folder at the standard indicator depth
(`indicators/<group>/<family-name>/`) and contains:

- **Shared interface** — the contract every variant implements.
- **Shared params struct/interface** — a single params shape common to all variants.
- **Variant type enum** — discriminator for the family (one constant per variant).
- **Common utilities file** — shared constants, helper functions (`push`,
  `fillWmaFactors`, ...) and the **factory/dispatcher** function.
- **One file per variant** implementing the shared interface.
- **One spec/test file per variant** plus one for the common file.

Example (TS — `ts/indicators/john-ehlers/hilbert-transformer/`):

```
cycle-estimator.ts                 # interface
cycle-estimator-params.ts          # shared params
cycle-estimator-type.ts            # variant enum
common.ts                          # utilities + createEstimator()
homodyne-discriminator.ts          # variant
homodyne-discriminator-unrolled.ts
phase-accumulator.ts
dual-differentiator.ts
# + matching .spec.ts for each
```

Go mirrors the same structure with lowercase filenames (`cycleestimator.go`,
`cycleestimatorparams.go`, `cycleestimatortype.go`, `estimator.go` for common
utilities and `NewCycleEstimator`, and one `<variant>estimator.go` per variant
— e.g., `homodynediscriminatorestimator.go`, `phaseaccumulatorestimator.go`).
These are the **conceptual-naming exception** noted in the File Naming section:
the folder name (`hilberttransformer/`) is never prefixed onto the member
files, because each file represents an independent concept within the family.

### Factory / Dispatcher Pattern

The common file exposes a dispatcher that constructs a variant by type.
Each variant has its own **default params** baked into the dispatcher — callers
that omit params for a specific type get sensible defaults for that variant.

```go
// Go — go/indicators/johnehlers/hilberttransformer/estimator.go
func NewCycleEstimator(typ CycleEstimatorType, params *CycleEstimatorParams) (CycleEstimator, error) {
    switch typ {
    case HomodyneDiscriminator:         return NewHomodyneDiscriminatorEstimator(params)
    case HomodyneDiscriminatorUnrolled: return NewHomodyneDiscriminatorEstimatorUnrolled(params)
    case PhaseAccumulator:              return NewPhaseAccumulatorEstimator(params)
    case DualDifferentiator:            return NewDualDifferentiatorEstimator(params)
    }
    return nil, fmt.Errorf("invalid cycle estimator type: %s", typ)
}
```

```ts
// TS — ts/indicators/john-ehlers/hilbert-transformer/hilbert-transformer-common.ts
export function createEstimator(
    estimatorType?: HilbertTransformerCycleEstimatorType,
    estimatorParams?: HilbertTransformerCycleEstimatorParams,
): HilbertTransformerCycleEstimator {
    if (estimatorType === undefined) {
        estimatorType = HilbertTransformerCycleEstimatorType.HomodyneDiscriminator;
    }
    switch (estimatorType) {
        case HilbertTransformerCycleEstimatorType.HomodyneDiscriminator:
            estimatorParams ??= { smoothingLength: 4, alphaEmaQuadratureInPhase: 0.2, alphaEmaPeriod: 0.2 };
            return new HilbertTransformerHomodyneDiscriminator(estimatorParams);
        // ... other variants, each with its own default params
        default:
            throw new Error('Invalid cycle estimator type: ' + estimatorType);
    }
}
```

The dispatcher must have a default variant (when `type` is omitted) and throw
on unknown types.

### Shared Interface Design

The shared interface exposes:

- Primed state (`primed`) and warm-up configuration (`warmUpPeriod`).
- Construction parameters as readable properties (`smoothingLength`,
  `alphaEmaQuadratureInPhase`, `alphaEmaPeriod`).
- Intermediate state visible for testing/debugging (`smoothed`, `detrended`,
  `inPhase`, `quadrature`, `period`).
- An `update(sample)` method returning the primary output.

TypeScript naming: **do not** suffix getters with `Value`. Prefix the backing
private field with `_` so the public getter can take the unadorned name:

```ts
// Correct — public getter is `period`, backing field is `_period`.
private _period = 0;
public get period(): number { return this._period; }

// Incorrect — do not use `periodValue` just to avoid a name clash.
```

### File Naming — No Type Suffixes

This module follows the same rule as the rest of the codebase: **no
`.interface.ts`, `.enum.ts`, or similar type-suffixes**. Interfaces, enums,
params, and implementations all use plain `.ts` filenames differentiated by
descriptive stem names.

### Warm-Up Period Contract

Each variant has an intrinsic minimum priming length (the sample count needed
for its state to be meaningful). If `params.warmUpPeriod` is provided and
larger than that minimum, it overrides the minimum; otherwise the intrinsic
minimum is used. Tests must cover both the default path and the custom
`warmUpPeriod` override.

## Taxonomy & Descriptor Registry

Every indicator carries a **Descriptor** — a taxonomic classification used by
charting UIs and discovery tooling to filter, group, and lay out indicators.
Descriptors are the single source of truth for output `kind` and `shape`;
indicator `Metadata()` implementations consume the registry via
`BuildMetadata` and MUST NOT hand-write `kind`/`shape`.

### Dimensions

| Primitive           | Go type                 | TS type                 | Purpose |
|---------------------|-------------------------|-------------------------|---------|
| `Identifier`        | `core.Identifier`       | `IndicatorIdentifier`   | Unique ID of the indicator. |
| `Role`              | `core.Role`             | `Role`                  | Semantic role of an **output** (Smoother, Envelope, Overlay, Polyline, Oscillator, BoundedOscillator, Volatility, VolumeFlow, Directional, CyclePeriod, CyclePhase, FractalDimension, Spectrum, Signal, Histogram, RegimeFlag, Correlation). |
| `Pane`              | `core.Pane`             | `Pane`                  | Chart pane where an output is drawn: `Price`, `Own`, `OverlayOnParent`. |
| `Adaptivity`        | `core.Adaptivity`       | `Adaptivity`            | `Static` or `Adaptive` (indicator-level). |
| `InputRequirement`  | `core.InputRequirement` | `InputRequirement`      | Minimum input data type: `ScalarInput`, `QuoteInput`, `BarInput`, `TradeInput`. |
| `VolumeUsage`       | `core.VolumeUsage`      | `VolumeUsage`           | `NoVolume`, `AggregateBarVolume`, `PerTradeVolume`, `QuoteLiquidityVolume`. |
| `OutputDescriptor`  | `core.OutputDescriptor` | `OutputDescriptor`      | `{Kind, Shape, Role, Pane}` per output. |
| `Descriptor`        | `core.Descriptor`       | `Descriptor`            | `{Identifier, Family, Adaptivity, InputRequirement, VolumeUsage, Outputs}`. |

**VolumeUsage ⇔ InputRequirement validity:**

- `AggregateBarVolume` ⇒ `BarInput`
- `PerTradeVolume` ⇒ `TradeInput`
- `QuoteLiquidityVolume` ⇒ `QuoteInput`
- `NoVolume` is valid with any input

### Registry Files

| Location                                              | Content |
|-------------------------------------------------------|---------|
| `go/indicators/core/descriptors.go`                   | Go descriptor map keyed by `Identifier`. |
| `go/indicators/core/descriptors_test.go`              | Coverage test: every `Identifier` has a descriptor row. |
| `go/indicators/core/metadata_descriptor_test.go`      | Invariants: outputs are non-empty, Kinds are strictly ascending starting at 1, shapes are known. |
| `ts/indicators/core/descriptors.ts`                   | TS descriptor `Map` keyed by `IndicatorIdentifier`. Uses `out()`/`desc()` helpers and `S`/`R`/`P`/`A`/`I`/`V` aliases for terse one-line entries. |

**Kind-numbering asymmetry:** Go per-indicator output enums are 1-based
(`iota + 1`), TS enums are 0-based. The descriptor tables mirror that — Go
`Kind` values are `1, 2, 3, ...`, TS `kind` values are `0, 1, 2, ...`. A row's
`Outputs` slice MUST appear in Kind-ascending order (same as the enum's
declared order), because `BuildMetadata` pairs `Outputs[i]` with
`outputTexts[i]` positionally.

### BuildMetadata Helper

Indicators call `BuildMetadata` / `buildMetadata` from within their
`Metadata()` method. The helper:

1. Looks up the descriptor for the given `Identifier`.
2. Panics (Go) / throws (TS) if no descriptor is registered.
3. Panics/throws if the number of `OutputText` entries does not match the
   number of descriptor outputs.
4. Returns a fully-populated `Metadata` with `kind` and `shape` taken from the
   registry and `mnemonic`/`description` taken from the caller.

**Go:**

```go
func (s *BollingerBands) Metadata() core.Metadata {
    desc := "Bollinger Bands " + s.mnemonic
    return core.BuildMetadata(
        core.BollingerBands,
        s.mnemonic,
        desc,
        []core.OutputText{
            {Mnemonic: s.mnemonic + " lower",       Description: desc + " Lower"},
            {Mnemonic: s.mnemonic + " middle",      Description: desc + " Middle"},
            {Mnemonic: s.mnemonic + " upper",       Description: desc + " Upper"},
            {Mnemonic: s.mnemonic + " bandWidth",   Description: desc + " Band Width"},
            {Mnemonic: s.mnemonic + " percentBand", Description: desc + " Percent Band"},
            {Mnemonic: s.mnemonic + " band",        Description: desc + " Band"},
        },
    )
}
```

**TypeScript:**

```ts
public metadata(): IndicatorMetadata {
  const description = `Bollinger Bands ${this.mnemonic}`;
  return buildMetadata(
    IndicatorIdentifier.BollingerBands,
    this.mnemonic,
    description,
    [
      { mnemonic: `${this.mnemonic} lower`,       description: `${description} Lower` },
      { mnemonic: `${this.mnemonic} middle`,      description: `${description} Middle` },
      { mnemonic: `${this.mnemonic} upper`,       description: `${description} Upper` },
      { mnemonic: `${this.mnemonic} bandWidth`,   description: `${description} Band Width` },
      { mnemonic: `${this.mnemonic} percentBand`, description: `${description} Percent Band` },
      { mnemonic: `${this.mnemonic} band`,        description: `${description} Band` },
    ],
  );
}
```

### Adding a Descriptor Row

When adding a new indicator you MUST add its descriptor in **both** Go and TS:

1. **Go** — `go/indicators/core/descriptors.go`:

   ```go
   MyIndicator: {
       Identifier: MyIndicator, Family: "<Author or Common/Custom>",
       Adaptivity: Static, InputRequirement: ScalarInput, VolumeUsage: NoVolume,
       Outputs: []OutputDescriptor{
           {Kind: 1 /* Output1 */, Shape: shape.Scalar, Role: Smoother, Pane: Price},
           {Kind: 2 /* Output2 */, Shape: shape.Scalar, Role: Signal,   Pane: Price},
       },
   },
   ```

2. **TS** — `ts/indicators/core/descriptors.ts` (single-line using helpers):

   ```ts
   [IndicatorIdentifier.MyIndicator, desc(IndicatorIdentifier.MyIndicator, '<family>', A.Static, I.ScalarInput, V.NoVolume, [
     out(0, S.Scalar, R.Smoother, P.Price),
     out(1, S.Scalar, R.Signal,   P.Price),
   ])],
   ```

3. Confirm the output order matches the per-indicator output enum order
   (`iota + 1` in Go, `0, 1, 2, ...` in TS).

4. Run the suites:
   ```
   cd go && go test ./indicators/core/...
   cd ts && npm test
   ```
   `TestDescriptorCoverage` and `TestDescriptorOutputsWellFormed` will catch
   missing or malformed rows on the Go side; the TS `buildMetadata` contract
   will throw at runtime if a row is missing.

### Family Naming

`Family` is the human-readable grouping label. Conventions:

- Author folder → author's full name (e.g., `john-ehlers` / `johnehlers`
  → `"John Ehlers"`, `welles-wilder` / `welleswilder` → `"Welles Wilder"`).
- `common/` → `"Common"`.
- `custom/` → `"Custom"`.

### Role Assignment Guidance

| Role                | Use for |
|---------------------|---------|
| `Smoother`          | A moving average or trend-following line plotted on price pane (SMA, EMA, JMA, KAMA, Middle Bollinger, ATCF trend lines, T2/T3, TEMA, ...). |
| `Envelope`          | Upper/lower bands or deviation channels (Bollinger upper/lower/band). |
| `Overlay`           | Other price-pane overlays that are not smoothers or envelopes (e.g., Parabolic SAR dots). |
| `Polyline`          | Outputs that produce an arbitrary polyline shape rather than a time series. |
| `Oscillator`        | Unbounded oscillator centred on 0 or arbitrary range (MACD, momentum, ATCF Ftlm/Stlm/Pcci). |
| `BoundedOscillator` | Oscillator bounded to 0..100 or similar (RSI, %R, Stochastic, ADX, PercentBand). |
| `Volatility`        | Volatility measure (TR, ATR, NATR, StdDev, BandWidth). |
| `VolumeFlow`        | Volume-based cumulative or flow measure (OBV, A/D, A/D Oscillator, MFI). |
| `Directional`       | Directional movement components (+DI, -DI, +DM, -DM). |
| `CyclePeriod`       | Dominant-cycle period estimate. |
| `CyclePhase`        | Cycle phase or sine-wave output. |
| `FractalDimension`  | Fractal-dimension based output. |
| `Spectrum`          | Spectrum / periodogram output (heatmap or polyline). |
| `Signal`            | The signal line of a multi-line indicator (MACD signal, Stochastic %D). |
| `Histogram`         | Difference bar between two lines (MACD histogram). |
| `RegimeFlag`        | Discrete regime indicator (Trend/Cycle mode). |
| `Correlation`       | Correlation coefficient in [-1, 1]. |

### Pane Assignment Guidance

| Pane               | Use for |
|--------------------|---------|
| `Price`            | Drawn on the price chart (smoothers, envelopes, overlays, SAR). |
| `Own`              | Drawn in its own subchart (oscillators, volatility, volume flow, spectra). |
| `OverlayOnParent`  | Drawn on top of another indicator's pane (reserved for composite visualizations). |

### Adaptivity

- `Adaptive` — the indicator adjusts its coefficients, length, or alpha based
  on market state (KAMA, JMA, MAMA, FRAMA, ATCF, DominantCycle, SineWave,
  HilbertTransformerInstantaneousTrendLine, TrendCycleMode, all Corona
  indicators, AutoCorrelationPeriodogram).
- `Static` — everything else.

## Indicator Factory

The **factory** maps an indicator identifier string + JSON parameters to a fully
constructed `Indicator` instance at runtime. This enables data-driven indicator
creation from configuration files, user input, or serialized settings.

### Location & Rationale

| Language   | Path                                  |
|------------|---------------------------------------|
| Go         | `go/indicators/factory/factory.go`    |
| TypeScript | `ts/indicators/factory/factory.ts`    |

The factory **cannot live in `core/`** in Go due to circular imports: indicator
packages import `core`, so `core` cannot import them back. The `factory/`
package sits at `indicators/factory/` — a sibling of `core/` and the
author/group folders — and imports both `core` and every indicator package.

TypeScript has no circular-import issue but uses the same location for
consistency.

### Go API

```go
package factory

// New creates an indicator from its identifier and JSON-encoded parameters.
func New(identifier core.Identifier, paramsJSON []byte) (core.Indicator, error)
```

- `identifier` is a `core.Identifier` value (the same enum used in descriptors
  and metadata).
- `paramsJSON` is the raw JSON for the indicator's params struct. For
  no-params indicators, pass `nil` or `[]byte("{}}")`.
- Returns the constructed indicator or an error if the identifier is unknown or
  params are invalid.

### TypeScript API

```ts
export function createIndicator(
    identifier: IndicatorIdentifier,
    params?: Record<string, unknown>,
): Indicator
```

- `identifier` is an `IndicatorIdentifier` enum value.
- `params` is the plain object (already parsed from JSON). For no-params
  indicators, omit or pass `undefined` / `{}`.
- Throws on unknown identifier or invalid params.

### Constructor Pattern Categories

The factory handles five categories of indicator constructors:

| Category | Count (Go/TS) | JSON detection | Example |
|----------|---------------|----------------|---------|
| **Standard `*Params` struct** | ~35 | Unmarshal into the params struct directly | SMA, WMA, RSI, BB, MACD, CCI |
| **Length vs SmoothingFactor variants** | 9 | If JSON contains `"smoothingFactor"` key → SF constructor; else → Length constructor | EMA, DEMA, TEMA, T2, T3, KAMA, CyberCycle, ITL |
| **Default vs Params variants** | ~15 | If JSON is `{}` or empty → Default constructor; else → Params constructor | Ehlers spectra, Corona family, DominantCycle, SineWave, ATCF |
| **Bare `int` length** | 10 | Parse `{"length": N}` | ATR, NATR, DM±, DI±, DMI, ADX, ADXR, WilliamsPercentR |
| **No params** | 3 | Ignore JSON entirely | TrueRange, BalanceOfPower, AdvanceDecline |

**Special case — MAMA:** uses `"fastLimitSmoothingFactor"` /
`"slowLimitSmoothingFactor"` keys for the SmoothingFactor variant, and
`"fastestSmoothingFactor"` / `"slowestSmoothingFactor"` for KAMA.

### Auto-Detection Logic

The factory uses helper functions to detect which constructor variant to call:

- **Go:** `hasKey(json, "smoothingFactor")` checks for a specific key;
  `isEmptyObject(json)` checks if JSON is `{}`.
- **TS:** `"smoothingFactor" in params` checks for a key;
  `Object.keys(params).length === 0` checks for empty params.

No JSON tags exist on Go params structs — Go's default case-insensitive JSON
unmarshaling handles key matching (e.g., JSON `"smoothingFactor"` matches Go
field `SmoothingFactor`).

### Adding a New Indicator to the Factory

When adding a new indicator, add a `case` to the factory's switch statement in
**both** Go and TS:

1. **Go** — add a `case core.<Identifier>:` block in `factory.New()` that
   unmarshals `paramsJSON` into the indicator's params struct and calls its
   constructor.
2. **TS** — add a `case IndicatorIdentifier.<Name>:` block in
   `createIndicator()` that spreads defaults onto `params` and calls the
   constructor.

The case block should follow the pattern of existing entries in the same
constructor category. TS uses a `{ defaultField: value, ...p }` spread pattern
to fill in required defaults.

### `icalc` CLI Tool

The `icalc` ("indicator calculator") CLI is a reference consumer of the factory
in both languages. It reads a JSON settings file, creates indicators via the
factory, and runs them against embedded test bar data.

| Language | Path | Run command |
|----------|------|-------------|
| Go | `go/cmd/icalc/main.go` | `cd go && go run ./cmd/icalc settings.json` |
| TS | `ts/cmd/icalc/main.ts` | `cd ts && npx tsx cmd/icalc/main.ts cmd/icalc/settings.json` |

Settings file format (`settings.json`):

```json
[
    { "identifier": "simpleMovingAverage", "params": { "length": 14 } },
    { "identifier": "exponentialMovingAverage", "params": { "smoothingFactor": 0.1 } },
    { "identifier": "trueRange", "params": {} }
]
```

- `identifier` is the camelCase JSON string of the `core.Identifier` /
  `IndicatorIdentifier` enum.
- `params` is the JSON object passed to the factory.
- Both CLIs embed 252-entry H/L/C/V test bar data (from TA-Lib `test_data.c`)
  and print metadata + per-bar outputs for each configured indicator.

## DefaultParams / defaultParams

Every indicator params file exports a **`DefaultParams()`** (Go) /
**`defaultParams()`** (TS) function that returns a fully-populated params
struct/object with sensible default values. This provides programmatic access to
defaults for UIs, factories, and documentation generators.

### Convention

- **Go:** `func DefaultParams() *Params` (or `*StructName` if the struct name
  differs from `Params`). Returns a pointer to a new struct.
- **TS:** `export function defaultParams(): SomeParamsInterface`. Returns a
  plain object satisfying the interface.
- **Dual-variant indicators** (Length/SmoothingFactor) export two functions:
  `DefaultLengthParams()` / `defaultLengthParams()` and
  `DefaultSmoothingFactorParams()` / `defaultSmoothingFactorParams()`.
- **No-params indicators** (empty struct/interface) still export
  `DefaultParams()` / `defaultParams()` returning an empty struct/object for
  consistency.
- **Component fields** (`BarComponent`, `QuoteComponent`, `TradeComponent`)
  are omitted from `DefaultParams` — their zero/undefined values already mean
  "use default".

### Placement

The function is added at the **end** of the params file, after the struct/
interface definition and any validation logic.

### Doc Comment

Go:
```go
// DefaultParams returns a Params value populated with conventional defaults.
func DefaultParams() *Params {
    return &Params{
        Length: 20,
    }
}
```

TypeScript:
```typescript
export function defaultParams(): SimpleMovingAverageParams {
    return {
        length: 20,
    };
}
```

### Default Value Sources

Default values come from (in priority order):
1. Values documented in the params struct/interface field comments.
2. The original paper or reference implementation (Ehlers, TA-Lib, etc.).
3. Conventional industry defaults (e.g., length=14 for RSI/ATR, length=20
   for SMA/EMA).

### When Adding a New Indicator

Add `DefaultParams()` / `defaultParams()` to the params file as part of the
standard indicator creation checklist. The function should be present from day
one, not added retroactively.
