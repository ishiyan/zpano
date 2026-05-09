---
name: indicator-architecture
description: Architecture, folder layout, naming conventions, and design patterns for the zpano trading indicators library. Load when creating new indicators or understanding the codebase structure.
---

# Architecture & Scaffolding Guide

This document describes the design decisions, folder layout, and naming conventions
for the **zpano** trading indicators library. It is intended as a reference for both
human developers and AI agents creating new indicators.

> **How to use this document:** This is a large reference — do NOT try to apply all
> rules at once. Follow these steps:
> 1. Identify your task: creating a new indicator, registering in factory, or naming review.
> 2. Identify your target language (Go, TS, Python, Zig, or Rust).
> 3. Use the navigation table below to jump directly to the relevant section.
> 4. Read only that section. Ignore all other language sections.

### Quick Navigation by Task

| Task | Go section | TS section | Python section | Zig section | Rust section |
|------|-----------|-----------|---------------|------------|-------------|
| Naming rules | [Go-Specific](#go-specific-conventions) | [TS-Specific](#typescript-specific-conventions) | [Python Reference](#python-reference) | [Zig Reference](#zig-reference) | [Rust Reference](#rust-reference) |
| LineIndicator pattern | [Go Impl](#go-implementation) | [TS Impl](#typescript-implementation) | [Python Impl](#python-implementation) | [Zig Impl](#zig-implementation) | [Rust Impl](#rust-implementation) |
| Factory registration | [Indicator Factory](#indicator-factory) | [Indicator Factory](#indicator-factory) | [Factory pattern (Py)](#factory-pattern) | [Factory pattern (Zig)](#factory-pattern-1) | [Factory pattern (Rust)](#factory-pattern-2) |
| Cross-language rules | [Cross-Language Local-Variable Parity](#cross-language-local-variable-parity) | — | — | — | — |

## Scope

The rules in this document apply **only to the `indicators/` folder** within each
language directory (e.g., `ts/indicators/`, `go/indicators/`, `py/indicators/`). Other folders at the
same level as `indicators/` (such as `entities/`) have their own conventions and are
not governed by this guide.

## Core Principles

1. **Multi-language library.** The same set of indicators is implemented in
   TypeScript, Go, Python, Rust, and Zig. Each language lives in its own
   top-level folder (`ts/`, `go/`, `python/`, `rs/`, `zig/`).
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
│   │   └── shape/          # Output shape enum (shape.Shape in Go, Shape in TS)
│   └── frequency-response/ # Frequency response calculation utilities
├── custom/                 # Your own experimental/custom indicators
│   └── my-indicator/
├── factory/                # Identifier + JSON params → indicator instance
├── <author-name>/          # Author-attributed indicators
│   └── <indicator-name>/
│       ├── implementation
│       ├── parameters
│       └── tests
└── ...
```

### The Four Special Folders

| Folder     | Purpose |
|------------|---------|
| `core/`    | Shared foundations: types, interfaces, enums, base abstractions, utilities. |
| `common/`  | Indicators whose author is unknown or not attributed to a single person (SMA, EMA, RSI, ...). |
| `custom/`  | Indicators you develop yourself. |
| `factory/` | Maps an indicator identifier + JSON parameters to a fully constructed indicator instance. Sits at `indicators/factory/` as a sibling of `core/` (not inside `core/`) because Go's circular-import rule prevents `core` from importing indicator packages. |

These four names are **reserved** and must never be used as author names.

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

**Author-folder boilerplate by language:**

- **Go:** every author folder MUST contain a `doc.go` that declares the
  parent package with a one-line godoc comment identifying the author:
  ```go
  // Package markjurik implements indicators developed by Mark Jurik.
  package markjurik
  ```
  This is the only file at the author-folder level; all actual indicator
  code lives in subpackages. The same rule applies to the special
  folders `common` and `custom`. The `core` folder already contains
  real framework code and needs no `doc.go`.

- **Python:** every directory (author, indicator, core, outputs) MUST
  contain an `__init__.py` file. Author-level `__init__.py` files are
  empty. Indicator-level `__init__.py` files re-export the main class,
  output enum, and params.

- **TypeScript, Zig, Rust:** no equivalent requirement (no package-level
  doc or init construct).

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
  response of a filter/indicator. See the **Frequency Response API** section
  below for the full API reference.

Other shared types (indicator interface, metadata, specification, line indicator
base class, indicator type enum) live directly in `core/`.

### Frequency Response API

The frequency response module computes the spectral characteristics of a
filter/indicator by feeding it an impulse signal, applying a real FFT, and
extracting power, amplitude, and phase spectra. The Go and TS implementations
are kept **feature-compatible** — both produce the same set of output
components with the same semantics.

#### Data Types

Both languages use a **component** type that bundles a data array with its
min/max bounds:

```go
// Go — frequencyresponse.Component
type Component struct {
    Data []float64
    Min  float64
    Max  float64
}
```

```typescript
// TS — FrequencyResponseComponent
interface FrequencyResponseComponent {
    data: number[];
    min: number;
    max: number;
}
```

The result struct/interface contains:

| Field                   | Go type      | TS type                       | Description |
|-------------------------|--------------|-------------------------------|-------------|
| `Label`                 | `string`     | `label: string`               | Filter mnemonic. |
| `NormalizedFrequency`   | `[]float64`  | `frequencies: number[]`       | Normalized frequencies in (0, 1], 1 = Nyquist. |
| `PowerPercent`          | `Component`  | `FrequencyResponseComponent`  | Spectrum power in percentages from max. |
| `PowerDecibel`          | `Component`  | `FrequencyResponseComponent`  | Spectrum power in decibels. |
| `AmplitudePercent`      | `Component`  | `FrequencyResponseComponent`  | Spectrum amplitude in percentages from max. |
| `AmplitudeDecibel`      | `Component`  | `FrequencyResponseComponent`  | Spectrum amplitude in decibels. |
| `PhaseDegrees`          | `Component`  | `FrequencyResponseComponent`  | Phase in degrees, range [-180, 180]. |
| `PhaseDegreesUnwrapped` | `Component`  | `FrequencyResponseComponent`  | Phase in degrees, unwrapped. |

#### Filter Interface

Both languages require the filter to expose metadata and an update function:

```go
// Go
type Updater interface {
    Metadata() core.Metadata
    Update(sample float64) float64
}
```

```typescript
// TS
interface FrequencyResponseFilter {
    metadata(): { mnemonic: string };
    update(sample: number): number;
}
```

#### Calculate Signature

```go
// Go
func Calculate(signalLength int, filter Updater, warmup int,
    phaseDegreesUnwrappingLimit float64) (*FrequencyResponse, error)
```

```typescript
// TS
static calculate(signalLength: number, filter: FrequencyResponseFilter,
    warmup: number, phaseDegreesUnwrappingLimit = 179,
    filteredSignal: number[] = []): FrequencyResponseResult
```

| Parameter                      | Description |
|--------------------------------|-------------|
| `signalLength`                 | Must be a power of 2 and ≥ 4 (realistic: 512–4096). |
| `filter`                       | The filter/indicator to analyze. |
| `warmup`                       | Number of zero-value updates before the impulse. |
| `phaseDegreesUnwrappingLimit`  | Threshold for phase unwrapping (use 179 as default). |
| `filteredSignal` (TS only)     | Optional pre-computed filtered signal; if provided and length matches, skips internal filtering. |

**Error handling:** Go returns `(*FrequencyResponse, error)`; TS throws on
invalid signal length.

#### Processing Pipeline

Both implementations follow the same pipeline:

1. Validate signal length (power of 2, ≥ 4).
2. Compute `spectrumLength = signalLength/2 - 1`.
3. Prepare frequency domain (normalized frequencies).
4. Prepare filtered signal (warmup with zeros, then impulse of 1000).
5. Apply direct real FFT.
6. Parse spectrum → extract power, amplitude, phase (with min/max tracking).
7. Unwrap phase degrees using the unwrapping limit.
8. Convert power/amplitude to decibels (with base normalization, min/max clamping).
9. Convert power/amplitude to percentages (with base normalization, max clamping).

#### Decibel Conversion

Decibels are computed relative to a base value (first element, or max if first
is near zero): `db = 20 * log10(value / base)`. The min is snapped to the
nearest [-100, -90), [-90, -80), ..., [-10, 0) interval boundary. Values below
-100 dB are clamped. The max is snapped to [0, 5) or [5, 10) interval
boundaries and clamped at 10 dB.

#### Percentage Conversion

Percentages are computed as `100 * value / base`. The max is snapped to
[100, 110), [110, 120), ..., [190, 200) interval boundaries and clamped at
200%. Min is always 0.

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

### Identifier Registry Parity

Go, TypeScript, Python, Zig, and Rust MUST have the same set of registered
identifiers. All five currently have **84** identifiers (Go: `core.Identifier`
iota 1–84; TS: `IndicatorIdentifier` enum 0–83; Python: `Identifier` IntEnum
0–83; Zig: `Identifier` enum(u8) 0–83; Rust: `Identifier` enum 0–83) with
identical names (PascalCase in Go/TS/Rust, UPPER_SNAKE_CASE in Python,
snake_case in Zig). When adding a new indicator, register the identifier in
**all** languages even if only one implementation exists yet.

#### Identifier Grouping by Author

Identifiers are grouped by author with comment dividers matching the style used
in `factory.go`. The **common** group comes first, author groups are arranged
alphabetically, and the **custom** group comes last.

Divider format per language:

| Language   | Divider style |
|------------|---------------|
| Go         | `// ── groupname ──────...──` (tab-indented) |
| TypeScript | `// ── groupname ──────...──` (4-space indent) |
| Python     | `# ── groupname ──────...──` (4-space indent) |
| Zig        | `// ── groupname ──────...──` (4-space indent) |
| Rust       | `// ── groupname ──────...──` (4-space indent) |

The 22 canonical groups in order:

1. common, 2. arnaud legoux, 3. donald lambert, 4. gene quong, 5. george lane,
6. gerald appel, 7. igor livshin, 8. jack hutson, 9. john bollinger,
10. john ehlers, 11. joseph granville, 12. larry williams, 13. manfred durschner,
14. marc chaikin, 15. mark jurik, 16. patrick mulloy, 17. perry kaufman,
18. tim tillson, 19. tushar chande, 20. vladimir kravchuk, 21. welles wilder,
22. custom

When adding a new indicator, **append** the new enum member at the end of its
author group (after the last existing member in that group), assigning the next
sequential number. Do **not** re-sort alphabetically within the group — this
avoids renumbering all subsequent members across 5 languages. If the author
group does not yet exist, insert a new group in alphabetical order among the
existing author groups (between "common" and "custom"), and assign sequential
numbers continuing from the previous group's last value.

### File Naming

| Language   | Style                     | Test files                  | Example                          |
|------------|---------------------------|-----------------------------|----------------------------------|
| TypeScript | `kebab-case.ts`           | `kebab-case.spec.ts`        | `simple-moving-average.ts`       |
| Go         | `lowercase.go`            | `lowercase_test.go`         | `simplemovingaverage.go`         |
| Python     | `snake_case.py`           | `test_snake_case.py`        | `simple_moving_average.py`       |
| Rust       | `snake_case.rs`           | `snake_case_test.rs` or inline `#[cfg(test)]` | `simple_moving_average.rs` |
| Zig        | `snake_case.zig`          | inline `test` blocks at bottom  | `simple_moving_average.zig`      |

No type-suffix convention is used (no `.enum.ts`, `.interface.ts`, etc.).
This is consistent across all languages.

#### Main vs. auxiliary files

Within an indicator folder, **only the main implementation file and its test
keep the indicator name**. Every auxiliary file drops the indicator-name
prefix — the folder (and in Go, the package name) already provides context,
so prefixing would stutter.

For a `simple-moving-average/` folder:

| Role                         | Go                              | TypeScript                              | Python                                   | Zig                                      | Rust                                     |
|------------------------------|---------------------------------|-----------------------------------------|------------------------------------------|------------------------------------------|------------------------------------------|
| Main implementation          | `simplemovingaverage.go`        | `simple-moving-average.ts`              | `simple_moving_average.py`               | `simple_moving_average.zig` (single file)| `simple_moving_average.rs` (single file) |
| Main test                    | `simplemovingaverage_test.go`   | `simple-moving-average.spec.ts`         | `test_simple_moving_average.py`          | inline `test` blocks at bottom           | inline `#[cfg(test)]` at bottom          |
| Data-driven test (if split)  | `data_test.go`                  | `data.spec.ts`                          | —                                        | —                                        | —                                        |
| Parameters                   | `params.go`                     | `params.ts`                             | `params.py`                              | inline struct in main file               | inline in main file                      |
| Output enum                  | `output.go`                     | `output.ts`                             | `output.py`                              | inline enum in main file                 | inline in main file                      |
| Output test                  | `output_test.go`                | `output.spec.ts`                        | —                                        | —                                        | —                                        |
| Coefficients / helpers       | `coefficients.go`, `estimator.go` | `coefficients.ts`, `estimator.ts`     | `coefficients.py`                        | `coefficients.zig`, `estimator.zig`      | inline in main file                      |
| Package docs (Go only)       | `doc.go`                        | —                                       | `__init__.py` (re-exports)               | —                                        | `mod.rs` (re-exports)                    |

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

**Cross-language parity:** the same identifier stem MUST be used in all five
languages. When porting, do not introduce new short forms; use the canonical
long form from the table above.

### Go-Specific Conventions

See the **Go Reference** section under Language-Specific Reference below for:
Go Receiver Naming Convention, Concurrency / Lock Convention, Go Style
Invariants, and Go Params: No JSON Tags.

### TypeScript-Specific Conventions

See the **TypeScript Reference** section under Language-Specific Reference
below for: TypeScript Import Conventions.

### Cross-Language Local-Variable Parity

The same indicator MUST use the **same local/field names** across all five
languages (adapted to each language's casing convention: camelCase in Go/TS,
snake_case in Python/Zig struct fields/Rust). When porting, copy the other
language's names verbatim where semantically identical. Observed canonical
vocabulary:

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
4. **Register the indicator** in `core/identifier` across all five languages.
   Append the new enum member at the end of its author group (do not re-sort
   alphabetically — just append after the last member). Assign the next
   sequential number. If the author group does not exist, insert a new group
   divider in alphabetical order among existing author groups. See the
   "Identifier Grouping by Author" section above for divider format.
   **Go:** also add the new identifier to all four test tables in
   `go/indicators/core/identifier_test.go` (String, IsKnown, MarshalJSON,
   UnmarshalJSON).
5. **Register the descriptor** in `core/descriptors.{go,ts,py,zig,rs}` — see the Taxonomy section below. A missing descriptor row causes
   `BuildMetadata` to panic at runtime.
6. **Register in the factory** — add a factory case mapping `Identifier` + JSON params → indicator instance.
7. **Add to `icalc/settings.json`** — add an entry with default params (camelCase identifier string).
8. **Run icalc** in all implemented languages to verify the indicator produces output without crashing.
9. **Test data in separate files** — place test input and expected arrays in dedicated files:
   - Go: `testdata_test.go`
   - Python: `test_testdata.py`
   - TypeScript: `testdata.ts`
   - Zig: `testdata.zig` (or bottom of test file)
   - Rust: `testdata.rs` (or inline in test module)
10. **Follow the consistent depth rule** -- the indicator must be exactly two
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

1. **`Update(sample float64) float64`** (Go) / **`update(sample: number): number`** (TS) /
   **`update(self, sample: float) -> float`** (Python) / **`update(self: *T, sample: f64) f64`** (Zig) /
   **`update(&mut self, sample: f64) -> f64`** (Rust) -- the core calculation logic.
2. **`Metadata()`** / **`metadata()`** -- returns indicator-level metadata with
   an explicit per-indicator output enum (not a hardcoded `kind: 0`).
3. **`IsPrimed()`** / **`isPrimed()`** / **`is_primed()`** / **`isPrimed()`** (Zig) / **`is_primed()`** (Rust) -- whether the indicator has received
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

### Python Implementation

In Python, `LineIndicator` is a **helper class used via composition**. The concrete
indicator creates a `LineIndicator` instance in `__init__` and delegates
`update_scalar/bar/quote/trade` calls to it. Unlike Go (embedding) and TS
(inheritance), the concrete indicator must explicitly define all four `update_*`
methods as one-line delegations:

```python
class SimpleMovingAverage(Indicator):
    def __init__(self, params: SimpleMovingAverageParams) -> None:
        bc = params.bar_component if params.bar_component is not None else DEFAULT_BAR_COMPONENT
        qc = params.quote_component if params.quote_component is not None else DEFAULT_QUOTE_COMPONENT
        tc = params.trade_component if params.trade_component is not None else DEFAULT_TRADE_COMPONENT

        bar_func = bar_component_value(bc)
        quote_func = quote_component_value(qc)
        trade_func = trade_component_value(tc)

        mnemonic = f"sma({length}{component_triple_mnemonic(bc, qc, tc)})"
        description = f"Simple moving average {mnemonic}"

        self._line = LineIndicator(mnemonic, description, bar_func, quote_func, trade_func, self.update)

    def update_scalar(self, sample: Scalar) -> Output:
        return self._line.update_scalar(sample)

    def update_bar(self, sample: Bar) -> Output:
        return self._line.update_bar(sample)

    def update_quote(self, sample: Quote) -> Output:
        return self._line.update_quote(sample)

    def update_trade(self, sample: Trade) -> Output:
        return self._line.update_trade(sample)
```

Component resolution uses `is not None` (not truthiness), so explicitly passing
a zero-valued enum like `BarComponent.OPEN` (value 0) works correctly.

Python entity modules export standalone functions `bar_component_value()`,
`quote_component_value()`, `trade_component_value()` that return
`Callable[[Bar], float]` (etc.), and constants `DEFAULT_BAR_COMPONENT`,
`DEFAULT_QUOTE_COMPONENT`, `DEFAULT_TRADE_COMPONENT`.

### Zig Implementation

In Zig, `LineIndicator` is used via **composition** (stored as a field). The concrete
indicator stores a `line: LineIndicator` field and delegates entity extraction:

```zig
pub const LineIndicator = struct {
    mnemonic: []const u8,
    description: []const u8,
    bar_func: BarFunc,
    quote_func: QuoteFunc,
    trade_func: TradeFunc,

    pub fn new(mnemonic, description, bc, qc, tc) LineIndicator { ... }
    pub fn extractBar(self: *const LineIndicator, bar: Bar) f64 { ... }
    pub fn wrapScalar(self: *const LineIndicator, time: i64, value: f64) OutputArray { ... }
};
```

Single-output indicators store `line: LineIndicator` and delegate entity extraction:
```zig
pub fn updateBar(self: *SuperSmoother, bar: Bar) OutputArray {
    const sample = self.line.extractBar(bar);
    const value = self.update(sample);
    return self.line.wrapScalar(bar.time, value);
}
```

The Zig `Indicator` interface uses a manually-constructed **vtable** for type erasure:

```zig
pub const Indicator = struct {
    ptr: *anyopaque,
    vtable: *const VTable,

    const VTable = struct {
        isPrimed: *const fn (*anyopaque) bool,
        metadata: *const fn (*anyopaque, *Metadata) void,
        updateScalar: *const fn (*anyopaque, Scalar) OutputArray,
        updateBar: *const fn (*anyopaque, Bar) OutputArray,
        updateQuote: *const fn (*anyopaque, Quote) OutputArray,
        updateTrade: *const fn (*anyopaque, Trade) OutputArray,
    };

    pub fn GenVTable(comptime T: type) VTable { ... }
};
```

Each concrete indicator exposes an `indicator()` method returning this interface:
```zig
const vtable = Indicator.GenVTable(SimpleMovingAverage);
pub fn indicator(self: *SimpleMovingAverage) Indicator {
    return .{ .ptr = self, .vtable = &vtable };
}
```

### Rust Implementation

In Rust, `LineIndicator` is used via **composition** (stored as a field), similar to
Python. Due to borrow-checker constraints, the update pattern extracts the component
value first, then calls `self.update()`:

```rust
pub trait Indicator {
    fn update(&mut self, sample: f64) -> f64;
    fn is_primed(&self) -> bool;
    fn metadata(&self) -> Metadata;
    fn update_bar(&mut self, bar: &Bar) -> Vec<Box<dyn std::any::Any>>;
    fn update_quote(&mut self, quote: &Quote) -> Vec<Box<dyn std::any::Any>>;
    fn update_trade(&mut self, trade: &Trade) -> Vec<Box<dyn std::any::Any>>;
}
```

`Output` is `Vec<Box<dyn std::any::Any>>`, downcast by callers.

The update pattern cannot use a closure delegation because the closure would borrow
`self` while `self.line` is already borrowed:

```rust
fn update_bar(&mut self, bar: &Bar) -> Vec<Box<dyn std::any::Any>> {
    let sample = (self.bar_func)(bar);
    let value = self.update(sample);
    vec![Box::new(Scalar::new(bar.time, value))]
}
```

### Per-Indicator Output Enums

Each indicator defines its own output enum describing what it produces. The
naming convention is **language-asymmetric** because the languages scope
symbols differently:

- **Go**: use bare `Output` / `Value`. The package name
  (`simplemovingaverage.Output`, `simplemovingaverage.Value`) provides scoping,
  so repeating the indicator name in the type would stutter.
- **TypeScript**: use long-form `<IndicatorName>Output` /
  `<IndicatorName>Value`. TS imports by symbol rather than by module, so the
  indicator name must be baked into the identifier to stay unambiguous at
  call sites.
- **Python**: use long-form `<IndicatorName>Output` (same as TS). `IntEnum`
  starting at 0 with `UPPER_SNAKE_CASE` members.
- **Zig**: use long-form `<IndicatorName>Output`. `enum(u8)` **1-based** with
  `snake_case` members.
- **Rust**: use long-form `<IndicatorName>Output`. `#[repr(u8)]` **1-based**
  with `PascalCase` members.

```go
// Go — file: simplemovingaverage/output.go
package simplemovingaverage

type Output int
const (
    Value Output = iota
)
```

```typescript
// TypeScript — simple-moving-average/output.ts
export enum SimpleMovingAverageOutput {
    SimpleMovingAverageValue = 0,
}
```

```python
# Python — simple_moving_average/output.py
class SimpleMovingAverageOutput(IntEnum):
    VALUE = 0
```

```zig
// Zig — simple_moving_average/simple_moving_average.zig
pub const SimpleMovingAverageOutput = enum(u8) {
    value = 1,
};
```

```rust
// Rust — simple_moving_average/simple_moving_average.rs
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
#[repr(u8)]
pub enum SimpleMovingAverageOutput {
    Value = 1,
}
```

**Cross-language output enum summary:**

| Language | Enum start | Naming style | Example |
|----------|-----------|--------------|---------|
| Go | 0 (`iota`) | Bare `Output` / bare concept names (pkg scoped) | `simplemovingaverage.Value` |
| TypeScript | 0 | `<IndicatorName>Output`, PascalCase members | `SimpleMovingAverageOutput.SimpleMovingAverageValue` |
| Python | 0 | `<IndicatorName>Output`, UPPER_SNAKE_CASE members | `SimpleMovingAverageOutput.VALUE` |
| Zig | 1 | `<IndicatorName>Output` enum(u8), snake_case members | `SimpleMovingAverageOutput.value` |
| Rust | 1 | `<IndicatorName>Output` #[repr(u8)], PascalCase members | `SimpleMovingAverageOutput::Value` |

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

This pattern is the same across all five languages. The helper function name
varies: `core.BuildMetadata` (Go), `buildMetadata` (TS), `build_metadata`
(Python), `buildMetadata` (Zig), `build_metadata` (Rust).

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
3. **Cross-language parity.** The same prefix must appear in all five languages.
   When adding an indicator, add its prefix to all languages simultaneously.
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

**Cross-language sentinel patterns:**

| Language | "Not set" representation | Resolution |
|----------|------------------------|------------|
| Go | `0` (zero value of int-based enum) | `if bc == 0 { bc = DefaultBarComponent }` |
| TypeScript | `undefined` (optional field) | `=== undefined` check in LineIndicator setter |
| Python | `None` (`Optional[BarComponent]`) | `if bc is None: bc = DEFAULT_BAR_COMPONENT` |
| Zig | `null` (`?BarComponent`) | `bc orelse default_bar_component` |
| Rust | `None` (`Option<BarComponent>`) | `bc.unwrap_or(DEFAULT_BAR_COMPONENT)` |

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

Python, Zig, and Rust follow the same dispatcher pattern with language-idiomatic
syntax. See the per-language reference sections below for specifics.

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
| `go/indicators/core/descriptors.go`                   | Go descriptor map keyed by `Identifier`. Grouped by author with `// ──` dividers. |
| `go/indicators/core/descriptors_test.go`              | Coverage test: every `Identifier` has a descriptor row. |
| `go/indicators/core/metadata_descriptor_test.go`      | Invariants: outputs are non-empty, Kinds are strictly ascending starting at 1, shapes are known. |
| `ts/indicators/core/descriptors.ts`                   | TS descriptor `Map` keyed by `IndicatorIdentifier`. Grouped by author with `// ──` dividers. Uses `out()`/`desc()` helpers and `S`/`R`/`P`/`A`/`I`/`V` aliases for terse one-line entries. |
| `py/indicators/core/descriptors.py`                   | Python descriptor dict keyed by `Identifier`. Grouped by author with `# ──` dividers. Uses `_o()`/`_d()` helpers. |
| `zig/src/indicators/core/descriptors.zig`             | Zig descriptor array. Grouped by author with `// ──` dividers. |
| `rs/src/indicators/core/descriptors.rs`               | Rust descriptor slice. Grouped by author with `// ──` dividers. |

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

When adding a new indicator you MUST add its descriptor in **all 5 languages**
(`go/indicators/core/descriptors.go`, `ts/indicators/core/descriptors.ts`,
`py/indicators/core/descriptors.py`, `zig/src/indicators/core/descriptors.zig`,
`rs/src/indicators/core/descriptors.rs`).

**Grouping rules (same as identifiers):**
- Descriptor entries are grouped by author with `// ──` / `# ──` comment dividers.
- "common" first, author groups alphabetical, "custom" last.
- **Append** new entry at end of its author group — do not re-sort alphabetically within a group.
- If a new author group is needed, insert a new divider in alphabetical order between existing groups.

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

3. **Python** — `py/indicators/core/descriptors.py` (using `_d()`/`_o()` helpers):

   ```python
   Id.MY_INDICATOR: _d(
       Id.MY_INDICATOR, "<family>", A.STATIC, I.SCALAR_INPUT, V.NO_VOLUME,
       [_o(0, S.SCALAR, R.SMOOTHER, P.PRICE),
        _o(1, S.SCALAR, R.SIGNAL,   P.PRICE)]),
   ```

4. **Zig** — `zig/src/indicators/core/descriptors.zig` (single-line struct literals):

   ```zig
   .{ .identifier = .my_indicator, .family = "<family>", .adaptivity = .static_, .input_requirement = .scalar_input, .volume_usage = .no_volume, .outputs = &[_]OD{.{ .kind = 1, .shape = .scalar, .role = .smoother, .pane = .price }, .{ .kind = 2, .shape = .scalar, .role = .signal, .pane = .price }} },
   ```

5. **Rust** — `rs/src/indicators/core/descriptors.rs`:

   ```rust
   Descriptor {
       identifier: MyIndicator, family: "<family>",
       adaptivity: Static, input_requirement: ScalarInput, volume_usage: NoVolume,
       outputs: &[OutputDescriptor { kind: 1, shape: Scalar, role: Smoother, pane: Price },
                  OutputDescriptor { kind: 2, shape: Scalar, role: Signal,   pane: Price }],
   },
   ```

6. Confirm the output order matches the per-indicator output enum order
   (`iota + 1` in Go, `1, 2, ...` in Zig/Rust, `0, 1, 2, ...` in TS/Python).

7. Run the suites:
   ```
   cd go && go test ./indicators/core/...
   cd ts && npm test
   cd zig && zig build test --summary all
   cd rs && cargo test --lib descriptors
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

| Language   | Path                                           |
|------------|------------------------------------------------|
| Go         | `go/indicators/factory/factory.go`             |
| TypeScript | `ts/indicators/factory/factory.ts`             |
| Python     | `py/indicators/factory/factory.py`             |
| Zig        | `zig/src/indicators/factory/factory.zig`       |
| Rust       | `rs/src/indicators/factory/factory.rs`         |

The factory **cannot live in `core/`** in Go due to circular imports: indicator
packages import `core`, so `core` cannot import them back. The `factory/`
package sits at `indicators/factory/` — a sibling of `core/` and the
author/group folders — and imports both `core` and every indicator package.

TypeScript has no circular-import issue but uses the same location for
consistency. The same layout is used in Python, Zig, and Rust.

### Factory Grouping

Factory switch/match cases are grouped by author with the same `// ──` / `# ──`
comment dividers used in identifiers and descriptors. The **common** group comes
first, author groups are arranged alphabetically, and the **custom** group comes
last. When adding a new indicator, **append** the new case at the end of its
author group — do not re-sort alphabetically within the group.

The 22 canonical groups in order:

1. common, 2. arnaud legoux, 3. donald lambert, 4. gene quong, 5. george lane,
6. gerald appel, 7. igor livshin, 8. jack hutson, 9. john bollinger,
10. john ehlers, 11. joseph granville, 12. larry williams, 13. manfred durschner,
14. marc chaikin, 15. mark jurik, 16. patrick mulloy, 17. perry kaufman,
18. tim tillson, 19. tushar chande, 20. vladimir kravchuk, 21. welles wilder,
22. custom

This order is identical across identifiers, descriptors, and factory files in
all 5 languages. Divider labels use the author's name with spaces — e.g.,
`// ── john ehlers ──...──`.

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

To determine which factory path applies for a given indicator, check these
conditions **in order** — the first match wins:

1. **No params** → Indicator has no params struct → ignore JSON, call bare constructor.
2. **SmoothingFactor variant** → JSON contains `"smoothingFactor"` key → call SF constructor.
3. **Default variant** → JSON is empty AND indicator has a Default constructor → call Default constructor.
4. **Bare int length** → Params struct has only a `length` field → parse `{"length": N}`.
5. **Standard params** → (all other cases) → unmarshal JSON into the full params struct.

The five categories with examples:

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

## Language-Specific Reference

This section documents language-specific conventions for each implementation.
All five ports are complete and produce identical results.

### Component Sentinel Pattern

Go uses `iota+1` (1-based enums) so `0` means "not set". TS uses `undefined`.
Python, Zig, and Rust keep **0-based enums** and use their idiomatic optional
types for the "not set" sentinel:

| Language | Params field type | "Not set" value | Resolution pattern |
|----------|------------------|-----------------|-------------------|
| **Go** | `entities.BarComponent` (int) | `0` | `if bc == 0 { bc = DefaultBarComponent }` |
| **TypeScript** | `BarComponent \| undefined` | `undefined` | `=== undefined` in setter |
| **Python** | `Optional[BarComponent]` | `None` | `if bc is None: bc = DEFAULT_BAR_COMPONENT` |
| **Zig** | `?BarComponent` | `null` | `bc orelse default_bar_component` |
| **Rust** | `Option<BarComponent>` | `None` | `bc.unwrap_or(DEFAULT_BAR_COMPONENT)` |

### Go Reference

#### Receiver Naming Convention

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

#### Concurrency / Lock Convention

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

#### Style Invariants (enforced across the codebase)

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

#### Params: No JSON Tags

Go indicator params structs do **not** carry `json:"..."` struct tags. Go's
`encoding/json` unmarshaler performs **case-insensitive** key matching by
default, so JSON `"smoothingFactor"` matches Go field `SmoothingFactor`
without explicit tags. Do not add JSON tags to params structs — they are
unnecessary and would diverge from the established pattern.

### TypeScript Reference

#### Import Conventions

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

### Folder Naming (repeated from above for completeness)

| Language | Author folder | Indicator folder | File naming |
|----------|--------------|-----------------|-------------|
| **Python** | `mark_jurik/` | `jurik_moving_average/` | `snake_case.py`, tests: `test_snake_case.py` |
| **Zig** | `mark_jurik/` | `jurik_moving_average/` | `snake_case.zig`, tests at bottom of source |
| **Rust** | `mark_jurik/` | `jurik_moving_average/` | `snake_case.rs`, tests in `#[cfg(test)]` at bottom |

### Python Reference

Python indicators live under `py/indicators/`. The Python port is **complete** (63 indicators,
factory, frequency_response, icalc/ifres/iconf cmd tools — 803 tests passing).

#### Directory structure

```
py/indicators/
├── __init__.py              # empty
├── core/
│   ├── __init__.py          # empty
│   ├── indicator.py         # Indicator ABC
│   ├── line_indicator.py    # LineIndicator helper (composition, not inheritance)
│   ├── identifier.py        # Identifier IntEnum (0-based, 72 values)
│   ├── metadata.py          # Metadata class
│   ├── build_metadata.py    # build_metadata() + OutputText
│   ├── output.py            # Output = list[Any]
│   ├── descriptor.py        # Descriptor class
│   ├── descriptors.py       # static registry, descriptor_of(), descriptors()
│   ├── output_descriptor.py # OutputDescriptor class
│   ├── specification.py     # Specification class
│   ├── component_triple_mnemonic.py
│   ├── adaptivity.py        # Adaptivity IntEnum
│   ├── input_requirement.py # InputRequirement IntEnum
│   ├── volume_usage.py      # VolumeUsage IntEnum
│   ├── role.py              # Role IntEnum
│   ├── pane.py              # Pane IntEnum
│   ├── frequency_response.py
│   ├── test_frequency_response.py
│   └── outputs/
│       ├── __init__.py
│       ├── metadata.py      # OutputMetadata class
│       ├── shape.py         # Shape IntEnum
│       ├── band.py          # Band class
│       ├── heatmap.py       # Heatmap class
│       └── polyline.py      # Polyline class (Point + Polyline)
├── common/
│   ├── __init__.py          # empty
│   └── simple_moving_average/
│       ├── __init__.py      # re-exports class, output, params, default_params
│       ├── simple_moving_average.py
│       ├── params.py
│       ├── output.py
│       └── test_simple_moving_average.py
├── <author_name>/           # e.g., john_ehlers/, mark_jurik/
│   ├── __init__.py          # empty
│   └── <indicator_name>/
│       └── ...
├── factory/
│   ├── __init__.py
│   └── factory.py           # create_indicator(identifier, params)
└── custom/
    ├── __init__.py
    └── ...
```

Every directory needs an `__init__.py` file. Author-level `__init__.py` files are empty.
Indicator-level `__init__.py` files re-export the main class, output enum, and params.

#### LineIndicator pattern — composition, not inheritance

Unlike Go (embedding) and TS (abstract class inheritance), Python uses **composition**.
The concrete indicator creates a `LineIndicator` instance and delegates `update_*` calls:

```python
class SimpleMovingAverage(Indicator):
    def __init__(self, params: SimpleMovingAverageParams) -> None:
        # ... resolve components, build mnemonic ...
        self._line = LineIndicator(mnemonic, description, bar_func, quote_func, trade_func, self.update)

    def update(self, sample: float) -> float:
        # core calculation logic
        ...

    def update_scalar(self, sample: Scalar) -> Output:
        return self._line.update_scalar(sample)

    def update_bar(self, sample: Bar) -> Output:
        return self._line.update_bar(sample)

    def update_quote(self, sample: Quote) -> Output:
        return self._line.update_quote(sample)

    def update_trade(self, sample: Trade) -> Output:
        return self._line.update_trade(sample)
```

Multi-output indicators (MACD, Bollinger Bands, etc.) do NOT use `LineIndicator`.
They implement `Indicator` directly, store component functions, and build output
lists manually in `update_scalar()`.

#### Params pattern

Parameters are `@dataclass` classes with defaults. Component fields use `Optional[BarComponent]`
with `None` meaning "use default". Every params file exports a `default_params()` function:

```python
@dataclass
class SimpleMovingAverageParams:
    length: int = 20
    bar_component: Optional[BarComponent] = None
    quote_component: Optional[QuoteComponent] = None
    trade_component: Optional[TradeComponent] = None

def default_params() -> SimpleMovingAverageParams:
    return SimpleMovingAverageParams()
```

Multi-constructor indicators (EMA) have multiple params classes and use `@staticmethod`
factory methods: `ExponentialMovingAverage.from_length(params)` and
`ExponentialMovingAverage.from_smoothing_factor(params)`.

Ehlers indicators with coefficient pre-computation use a `create(params)` static method
(e.g., `SuperSmoother.create(params)`) and a `create_default()` for parameterless defaults.

#### Output enum pattern

Per-indicator output enums are `IntEnum` starting at 0. Naming uses the long-form
`<IndicatorName>Output` (same as TypeScript):

```python
class SimpleMovingAverageOutput(IntEnum):
    VALUE = 0
```

#### Import conventions

- Relative imports within the indicators package (e.g., `from ...core.metadata import Metadata`)
- Tests use absolute imports (e.g., `from py.indicators.common.simple_moving_average...`)
- Entity imports: `from ....entities.bar_component import BarComponent, DEFAULT_BAR_COMPONENT, bar_component_value`

Key import paths from an indicator file at `py/indicators/<group>/<indicator>/`:
```python
from ...core.indicator import Indicator
from ...core.line_indicator import LineIndicator
from ...core.metadata import Metadata
from ...core.build_metadata import build_metadata, OutputText
from ...core.identifier import Identifier
from ...core.component_triple_mnemonic import component_triple_mnemonic
from ...core.output import Output
from ....entities.bar import Bar
from ....entities.bar_component import BarComponent, DEFAULT_BAR_COMPONENT, bar_component_value
from ....entities.quote_component import QuoteComponent, DEFAULT_QUOTE_COMPONENT, quote_component_value
from ....entities.trade_component import TradeComponent, DEFAULT_TRADE_COMPONENT, trade_component_value
```

For heatmap-producing indicators, also:
```python
from ...core.outputs.heatmap import Heatmap
```

#### Naming conventions

| Concept | Python |
|---------|--------|
| Identifier enum | `Identifier.SIMPLE_MOVING_AVERAGE` (UPPER_SNAKE_CASE, 0-based) |
| Output enum | `SimpleMovingAverageOutput.VALUE` (UPPER_SNAKE_CASE, 0-based) |
| Params class | `SimpleMovingAverageParams` (PascalCase `@dataclass`) |
| Indicator class | `SimpleMovingAverage` (PascalCase) |
| File naming | `simple_moving_average.py`, `test_simple_moving_average.py` |
| Folder naming | `simple_moving_average/`, `john_ehlers/` |
| Methods | `update()`, `is_primed()`, `metadata()`, `update_bar()` (snake_case) |
| Private fields | `self._primed`, `self._window`, `self._line` |
| Constants | `DEFAULT_BAR_COMPONENT`, `_EPSILON` |

#### Testing conventions

- `unittest.TestCase` with `assertAlmostEqual(result, expected, delta=1e-N)`
  (using `delta=`, NOT `places=`, to match Go's `math.Abs(exp-act) > 1e-N` semantics)
- `math.isnan()` checks for NaN expected values
- `self.assertTrue(sma.is_primed())` — always call `is_primed()` with parentheses
- Tests use absolute imports: `from py.indicators.common.simple_moving_average.simple_moving_average import SimpleMovingAverage`
- Test data arrays are module-level constants (INPUT, EXPECTED_3, EXPECTED_5, etc.)

#### Factory pattern

`py/indicators/factory/factory.py` exports `create_indicator(identifier, params)`.
It uses four construction patterns:

1. **Direct construction**: `Indicator(_apply(default_params(), params))` — 51 indicators
2. **`create()` static factory**: `Indicator.create(_apply(default_params(), params))` — Ehlers indicators (9)
3. **Dual `from_length()` / `from_smoothing_factor()`**: detected by `_has_key(params, 'smoothing_factor')` — EMA family (9)
4. **Raw constructor**: `WilliamsPercentR(length)` — 1 indicator

The `_apply()` helper overlays dict values onto a default dataclass instance.

#### Descriptor registry

`py/indicators/core/descriptors.py` uses 0-based `Kind` values (matching Python/TS output enums),
unlike Go which uses `iota` (0-based but starting from the `Output` const block which itself is 0).
Output kinds are integers directly matching the per-indicator `IntEnum` values.

#### Cmd tools

Three CLI tools at `py/cmd/{icalc,ifres,iconf}/`:
- Each has `__init__.py`, `__main__.py`, `main.py`, `settings.json`, `run-me.sh`
- Run via: `PYTHONPATH=. python3 -m py.cmd.icalc py/cmd/icalc/settings.json`
- `icalc/main.py` contains shared data: `_IDENTIFIER_MAP`, `_convert_params()`, 252-bar test arrays
- `ifres` and `iconf` import from `icalc/main.py` for shared functionality

#### No concurrency

Python indicators have no mutex/lock equivalent (unlike Go's `sync.RWMutex`).
The indicators are not thread-safe by design.

### Zig Reference

The Zig port is **complete**: 63 indicators, factory, frequency_response,
icalc/ifres/iconf CLI tools — 1014 tests passing.

#### Directory structure

```
zig/src/indicators/
├── indicators.zig           # barrel: re-exports all types + comptime test inclusion
├── core/
│   ├── indicator.zig        # Indicator interface (vtable-based), OutputValue, OutputArray
│   ├── line_indicator.zig   # LineIndicator composition helper
│   ├── identifier.zig       # Identifier enum(u8), 72 values (0-based)
│   ├── metadata.zig         # Metadata struct (fixed-capacity OutputMetadata buffer)
│   ├── build_metadata.zig   # buildMetadata() + OutputText
│   ├── descriptor.zig       # Descriptor struct
│   ├── descriptors.zig      # static registry, descriptorOf(), allDescriptors()
│   ├── output_descriptor.zig
│   ├── specification.zig
│   ├── component_triple_mnemonic.zig
│   ├── adaptivity.zig       # Adaptivity enum
│   ├── input_requirement.zig
│   ├── volume_usage.zig
│   ├── role.zig
│   ├── pane.zig
│   ├── frequency_response.zig  # FFT spectral analysis (582 lines)
│   └── outputs/
│       ├── shape.zig        # Shape enum
│       ├── band.zig         # Band struct
│       ├── heatmap.zig      # Heatmap struct (max 256 values)
│       ├── polyline.zig     # Point + Polyline
│       └── output_metadata.zig
├── common/
│   └── simple_moving_average/
│       └── simple_moving_average.zig   # params + output + impl + tests (single file)
├── <author_name>/           # e.g., john_ehlers/, mark_jurik/
│   └── <indicator_name>/
│       └── <indicator_name>.zig
├── factory/
│   └── factory.zig          # create(allocator, Identifier, params_json)
└── (20 author family directories, ~68 indicator subdirectories)
```

Every indicator is a single `.zig` file: params struct, output enum, indicator struct,
impl, and `test` blocks — all in one file. No separate params/output/test files.

#### Indicator interface — vtable-based type erasure

Unlike Go (interface), TS (class inheritance), Python (ABC), or Rust (trait object),
Zig uses a manually-constructed **vtable** for type erasure:

```zig
pub const Indicator = struct {
    ptr: *anyopaque,
    vtable: *const VTable,

    const VTable = struct {
        isPrimed: *const fn (*anyopaque) bool,
        metadata: *const fn (*anyopaque, *Metadata) void,
        updateScalar: *const fn (*anyopaque, Scalar) OutputArray,
        updateBar: *const fn (*anyopaque, Bar) OutputArray,
        updateQuote: *const fn (*anyopaque, Quote) OutputArray,
        updateTrade: *const fn (*anyopaque, Trade) OutputArray,
    };

    pub fn GenVTable(comptime T: type) VTable { ... }
};
```

Each concrete indicator exposes an `indicator()` method returning this interface:
```zig
const vtable = Indicator.GenVTable(SimpleMovingAverage);
pub fn indicator(self: *SimpleMovingAverage) Indicator {
    return .{ .ptr = self, .vtable = &vtable };
}
```

#### OutputValue — tagged union (not `Box<dyn Any>`)

Zig avoids heap allocation for outputs. `OutputValue` is a tagged union:
```zig
pub const OutputValue = union(enum) {
    scalar: Scalar,
    band: Band,
    heatmap: Heatmap,
    polyline: Polyline,
};
```

`OutputArray` is a fixed-capacity array (max 9 outputs) with stack storage:
```zig
pub const OutputArray = struct {
    buf: [max_output_count]OutputValue = undefined,
    len: usize = 0,
    pub fn fromScalar(s: Scalar) OutputArray { ... }
    pub fn append(self: *OutputArray, val: OutputValue) void { ... }
    pub fn slice(self: *const OutputArray) []const OutputValue { ... }
};
```

#### LineIndicator — composition pattern

Like Python and Rust, Zig uses **composition** (stored as a field):

```zig
pub const LineIndicator = struct {
    mnemonic: []const u8,
    description: []const u8,
    bar_func: BarFunc,
    quote_func: QuoteFunc,
    trade_func: TradeFunc,

    pub fn new(mnemonic, description, bc, qc, tc) LineIndicator { ... }
    pub fn extractBar(self: *const LineIndicator, bar: Bar) f64 { ... }
    pub fn wrapScalar(self: *const LineIndicator, time: i64, value: f64) OutputArray { ... }
};
```

Single-output indicators store `line: LineIndicator` and delegate entity extraction:
```zig
pub fn updateBar(self: *SuperSmoother, bar: Bar) OutputArray {
    const sample = self.line.extractBar(bar);
    const value = self.update(sample);
    return self.line.wrapScalar(bar.time, value);
}
```

#### Params pattern

Parameters are plain structs with optional component fields. No `Default` trait —
defaults are set in `init()`:

```zig
pub const SimpleMovingAverageParams = struct {
    length: u32 = 20,
    bar_component: ?BarComponent = null,
    quote_component: ?QuoteComponent = null,
    trade_component: ?TradeComponent = null,
};
```

Component sentinel pattern: `?BarComponent` with `null` meaning "use default",
resolved via `orelse default_bar_component`.

#### Output enum pattern

Per-indicator output enums are `enum(u8)` **1-based** (matching Go's `iota+1`):

```zig
pub const SimpleMovingAverageOutput = enum(u8) {
    value = 1,
};
```

#### Constructor: `init()` and `deinit()`

Indicators that need heap allocation (ring buffers, arrays) take an `allocator`
parameter in `init()` and must implement `deinit()`:

```zig
pub fn init(allocator: std.mem.Allocator, params: SimpleMovingAverageParams) !SimpleMovingAverage { ... }
pub fn deinit(self: *SimpleMovingAverage) void {
    self.allocator.free(self.window);
}
```

Simple indicators (e.g., SuperSmoother) that need no allocation take no allocator
and have no deinit.

#### The `fixSlices()` pattern — critical for heap allocation

When the factory copies an indicator from stack to heap (`heapAlloc`), any slices
pointing into the struct's owned `[N]u8` buffers become dangling. Every indicator
that stores mnemonic/description as slices into owned buffers must implement:

```zig
pub fn fixSlices(self: *SimpleMovingAverage) void {
    self.line.mnemonic = self.mnemonic_buf[0..self.mnemonic_len];
    self.line.description = self.description_buf[0..self.description_len];
}
```

This is called automatically by the factory's `heapAlloc` after copying.

#### Owned string buffer pattern

Indicators store mnemonics/descriptions in fixed-size owned buffers to avoid
heap-allocating strings:

```zig
mnemonic_buf: [64]u8 = undefined,
mnemonic_len: usize = 0,
description_buf: [128]u8 = undefined,
description_len: usize = 0,
```

Built via `std.fmt.bufPrint(&self.mnemonic_buf, "sma({d}{s})", .{length, ctm})`.

#### Factory pattern

`zig/src/indicators/factory/factory.zig` exports:

```zig
pub fn create(allocator: Allocator, id: Identifier, params_json: []const u8) FactoryError!FactoryResult
```

`FactoryResult` bundles the type-erased indicator with a cleanup function:
```zig
pub const FactoryResult = struct {
    indicator: Indicator,         // vtable-based interface
    deinit_fn: *const fn(*anyopaque, Allocator) void,
    ptr: *anyopaque,
    pub fn deinit(self: FactoryResult, allocator: Allocator) void { ... }
};
```

Two internal creation paths:
1. **`createWithParams`** — for indicators that need no allocator (most Ehlers)
2. **`createWithAllocParams`** — for indicators requiring heap allocation (SMA, BB, etc.)

Both use `heapAlloc` to copy the stack-initialized indicator to the heap and call
`fixSlices()` automatically.

JSON parsing uses `std.json` from the standard library — no custom parser needed
(unlike Rust which has its own `json.rs`).

#### Import conventions

All imports use `@import` with build.zig module names (not file paths):

From the barrel file `indicators.zig`, indicators import core types:
```zig
const indicators = @import("indicators");
const Identifier = indicators.Identifier;
const Metadata = indicators.Metadata;
const Indicator = indicators.Indicator;
const OutputArray = indicators.OutputArray;
const LineIndicator = indicators.LineIndicator;
```

Entity imports:
```zig
const bar_mod = @import("bar");
const Bar = bar_mod.Bar;
const BarComponent = bar_mod.BarComponent;
const default_bar_component = bar_mod.default_bar_component;
const barComponentValue = bar_mod.componentValue;
```

#### Naming conventions

| Concept | Zig |
|---------|-----|
| Identifier enum | `Identifier.simple_moving_average` (snake_case, 0-based) |
| Output enum | `SimpleMovingAverageOutput.value` (snake_case, **1-based**) |
| Params struct | `SimpleMovingAverageParams` (PascalCase) |
| Indicator struct | `SimpleMovingAverage` (PascalCase) |
| File naming | `simple_moving_average.zig` (snake_case) |
| Folder naming | `simple_moving_average/`, `john_ehlers/` (snake_case) |
| Methods | `update()`, `isPrimed()`, `getMetadata()`, `updateBar()` (camelCase) |
| Struct fields | `window_sum`, `last_index`, `mnemonic_buf` (snake_case) |
| Constants | `default_bar_component`, `max_output_count` (snake_case) |

Note: Zig uses `camelCase` for methods/functions and `snake_case` for struct fields
and module-level constants — this is standard Zig convention.

#### Testing conventions

- Inline `test "descriptive name" { ... }` blocks at bottom of indicator files
- `std.testing.expect(almostEqual(result, expected, 1e-8))` for float comparisons
- `std.math.isNan(result)` checks for NaN expected values
- `testing.allocator` with `defer indicator.deinit()` for leak detection
- Test data as module-level `const` arrays
- The barrel file `indicators.zig` has a `comptime` block that force-references all
  indicator modules so their tests are included in `zig build test`

#### No concurrency

Zig indicators have no mutex/lock equivalent (like Python and Rust, unlike Go's
`sync.RWMutex`). The vtable uses `*anyopaque` (mutable pointer), so thread safety
is the caller's responsibility.

#### Metadata — fixed-capacity buffer

Unlike Go (slice), TS (array), Python (list), Rust (`Vec`), Zig uses a
fixed-capacity buffer (max 9 outputs) to avoid heap allocation:

```zig
pub const Metadata = struct {
    identifier: Identifier,
    mnemonic: []const u8,
    description: []const u8,
    outputs_buf: [9]OutputMetadata,
    outputs_len: usize,
};
```

`buildMetadata` takes an `*Metadata` out-pointer (not return value) to avoid
large struct copies.

#### Cmd tools

Three CLI tools at `zig/src/cmd/{icalc,ifres,iconf}/`:
- Each has `main.zig`, `settings.json`, `run-me.sh`, and output files
- Build via: `cd zig && zig build icalc` (or `ifres`, `iconf`)
- Run via: `cd zig && zig build icalc -- src/cmd/icalc/settings.json`
- `icalc` — indicator calculator, outputs metadata + per-bar values to stdout
- `ifres` — frequency response calculator, wraps indicators in updater adapter
- `iconf` — chart config generator, writes `.json` and `.ts` files directly

All use Zig 0.16 I/O patterns (`Io.Writer.Allocating`, `Io.Dir.cwd()`,
`File.writePositionalAll(io, data, 0)`).

#### Frequency response

`zig/src/indicators/core/frequency_response.zig` — 582 lines, standalone module.
Computes spectral characteristics (power, amplitude, phase) using FFT. Returns
a `FrequencyResponse` struct with 7 allocated `[]f64` components.

Used by the `ifres` CLI tool. The tool wraps indicators in an updater that
implements the `Updater` interface expected by `frequency_response.calculate()`.

### Rust Reference

The Rust port is **complete**: 67 indicators, factory, frequency_response, icalc/ifres/iconf — 985 tests passing.

#### File layout

Rust uses a single-file pattern per indicator. Each indicator lives in
`rs/src/indicators/<group>/<indicator_name>/`:

- `mod.rs` — re-exports: `mod <indicator_name>; pub use <indicator_name>::*;`
- `<indicator_name>.rs` — contains params struct, output enum, indicator struct, impl, and `#[cfg(test)] mod tests`

No separate `params.rs`, `output.rs`, or test files. Everything is in one `.rs` file,
re-exported through `mod.rs`.

Group directories also have a `mod.rs` that lists all indicator submodules.

#### Module visibility patterns

Two patterns are used for re-exports in `mod.rs` files:

- **Pattern A** (preferred) — Private mod + wildcard re-export:
  ```rust
  mod simple_moving_average;
  pub use simple_moving_average::*;
  ```
  Types accessible as `crate::indicators::common::simple_moving_average::TypeName`.
  Used by `common/*`.

- **Pattern B** — Public module:
  ```rust
  pub mod goertzel_spectrum;
  ```
  Types accessible as `crate::indicators::custom::goertzel_spectrum::goertzel_spectrum::TypeName`.
  Used by `custom/*`, `gerald_appel/*`, `tim_tillson/*`.

The factory imports must match the visibility pattern of each indicator.

#### Params

Params are plain structs (not `#[derive(Clone)]` unless needed) with `impl Default`:

```rust
pub struct SimpleMovingAverageParams {
    pub length: usize,
    pub bar_component: Option<BarComponent>,
    pub quote_component: Option<QuoteComponent>,
    pub trade_component: Option<TradeComponent>,
}

impl Default for SimpleMovingAverageParams {
    fn default() -> Self {
        Self { length: 20, bar_component: None, quote_component: None, trade_component: None }
    }
}
```

Component fields use `Option<BarComponent>` (`None` = use default). This matches
the sentinel pattern: Go uses zero-value, TS uses `undefined`, Python uses `Optional[...] = None`,
Rust uses `Option<...>`.

Multi-constructor indicators (EMA family) have separate params structs:
`ExponentialMovingAverageLengthParams` and `ExponentialMovingAverageSmoothingFactorParams`.

MESA params do **not** implement `Default` due to complex nested types; the factory
constructs them explicitly.

#### Output enums

Output enums are **1-based** (matching Go's `iota+1` and the descriptor registry):

```rust
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
#[repr(u8)]
pub enum SimpleMovingAverageOutput {
    Value = 1,
}
```

#### Indicator trait and LineIndicator

The `Indicator` trait is the core interface:

```rust
pub trait Indicator {
    fn update(&mut self, sample: f64) -> f64;
    fn is_primed(&self) -> bool;
    fn metadata(&self) -> Metadata;
    fn update_bar(&mut self, bar: &Bar) -> Vec<Box<dyn std::any::Any>>;
    fn update_quote(&mut self, quote: &Quote) -> Vec<Box<dyn std::any::Any>>;
    fn update_trade(&mut self, trade: &Trade) -> Vec<Box<dyn std::any::Any>>;
}
```

`Output` is `Vec<Box<dyn std::any::Any>>`, downcast by callers.

`LineIndicator` is used via composition (like Python), stored as a field. Due to
borrow-checker constraints, the update pattern extracts the component value first:

```rust
fn update_bar(&mut self, bar: &Bar) -> Vec<Box<dyn std::any::Any>> {
    let sample = (self.bar_func)(bar);
    let value = self.update(sample);
    vec![Box::new(Scalar::new(bar.time, value))]
}
```

You **cannot** use `self.line.update_bar(bar, |v| self.update(v))` because the closure
borrows `self` while `self.line` is already borrowed.

#### Import conventions

All imports use absolute `crate::` paths:

```rust
use crate::entities::bar::Bar;
use crate::entities::bar_component::{component_value as bar_component_value, BarComponent, DEFAULT_BAR_COMPONENT};
use crate::indicators::core::build_metadata::{build_metadata, OutputText};
use crate::indicators::core::identifier::Identifier;
use crate::indicators::core::indicator::{Indicator, Output};
use crate::indicators::core::metadata::Metadata;
```

Note: `component_value` is imported with an alias (`as bar_component_value`) to avoid
name collisions when multiple component modules are imported.

#### Naming conventions

| Concept | Rust |
|---------|------|
| Identifier enum | `Identifier::SimpleMovingAverage` (PascalCase, 0-based) |
| Output enum | `SimpleMovingAverageOutput::Value` (PascalCase, **1-based**) |
| Params struct | `SimpleMovingAverageParams` (PascalCase, `impl Default`) |
| Indicator struct | `SimpleMovingAverage` (PascalCase) |
| File naming | `simple_moving_average.rs` (snake_case) |
| Folder naming | `simple_moving_average/`, `john_ehlers/` (snake_case) |
| Methods | `update()`, `is_primed()`, `metadata()`, `update_bar()` (snake_case) |
| Private fields | `primed`, `window`, `line` (no prefix, struct fields are private by default) |
| Constants | `DEFAULT_BAR_COMPONENT`, `EPSILON` (UPPER_SNAKE_CASE) |

Module renames from Go/TS: `cybercycle→cyber_cycle`, `sinewave→sine_wave`,
`supersmoother→super_smoother`, `stochastic_rsi→stochastic_relative_strength_index`.

#### Testing conventions

- Inline `#[cfg(test)] mod tests { use super::*; ... }` at bottom of indicator file
- `assert!((result - expected).abs() < 1e-8)` or custom `almost_equal()` helper
- `f64::is_nan()` checks for NaN expected values
- Test data arrays as module-level constants (`INPUT`, `EXPECTED_3`, etc.)

#### Factory pattern

`rs/src/indicators/factory/factory.rs` exports `create_indicator(id: Identifier, params: &str) -> Box<dyn Indicator>`.

Custom JSON parser in `factory/json.rs` (~250 lines, zero dependencies — no serde).
Provides `JsonValue` enum and helpers: `has_key()`, `get_f64()`, `get_usize()`, `get_bool()`, etc.

Factory uses a large `match` on `Identifier` enum (~152 arms). Construction patterns:
1. **Default**: `SomeIndicator(SomeIndicatorParams { ..overrides, ..Default::default() })`
2. **EMA-style dual**: detect `has_key("smoothing_factor")` → length vs smoothing_factor variant
3. **`create()` associated function**: Ehlers indicators with `MesaParams`
4. **Explicit construction**: MESA params built field-by-field (no `Default`)

#### Frequency response

`rs/src/indicators/core/frequency_response.rs` — standalone module, not tied to any indicator.
Uses `Updater` trait. The `ifres` cmd tool wraps indicators in an `IndicatorUpdater` adapter struct.

#### Cmd tools

Three binaries in `rs/src/bin/`:
- `icalc.rs` — indicator calculator (reads JSON settings, feeds 252-bar test data)
- `ifres.rs` — frequency response calculator (`IndicatorUpdater` adapter implements `Updater`)
- `iconf.rs` — chart config generator (custom `JVal` enum for JSON building, `PaneData` struct, color cycling, generates `.json` + `.ts`)

All share: JSON settings parsing, 252-bar embedded test data, `json_value_to_string()` helper.

Run via: `cd rust && cargo run --bin icalc -- settings.json`

#### No concurrency

Rust indicators have no `Mutex`/`RwLock` (like Python, unlike Go's `sync.RWMutex`).
The indicators are not `Send`/`Sync` by default due to internal mutability via `&mut self`.

#### Precision fix

`std::f64::consts::FRAC_PI_3` must be used instead of `(2.0 * PI) / 6.0` in
`shared.rs` `calculate_differential_phase()` to match Go's compile-time constant precision.

## Related Skills

- **`indicator-checklist`** — Compact (~80 lines) mechanically-verifiable rules for imports, naming, and structure. Use for quick verification or as a checklist during conversion.
- **`indicator-conversion`** — Full step-by-step conversion workflow for porting indicators between languages. Includes the multi-language conversion routing (Go→all, Python→Go→rest, Rust→Go→rest), streaming adaptation, test data generation, factory registration, and icalc verification.
- **`mbst-indicator-conversion`** — Converting from MBST C# source indicators.
- **`talib-indicator-conversion`** — Converting from TA-Lib C source indicators.
