---
name: indicator-conversion
description: Step-by-step guide for converting old-structure indicators to the new architecture in all five languages (Go, TypeScript, Python, Zig, Rust). Load when migrating an existing indicator or implementing one in any language.
---

# Converting an Old-Structure Indicator to New Structure

This guide provides step-by-step instructions for converting an indicator from the old
architecture to the new architecture. It covers all five languages: **Go**, **TypeScript**,
**Python**, **Zig**, and **Rust**.

The new architecture introduces:

- **Per-indicator packages** (Go) / **per-indicator folders** (TS/Python) instead of a flat shared package
- **`LineIndicator` embedding/inheritance/composition** that eliminates `UpdateScalar/Bar/Quote/Trade` boilerplate
- **`ComponentTripleMnemonic`** (bar + quote + trade) replacing `componentPairMnemonic` (bar + quote only). The mnemonic format gains a `%s` placeholder for the component triple string, e.g. `"sma(%d%s)"` where `%s` is the output of `ComponentTripleMnemonic(bc, qc, tc)` — an empty string when all components are defaults, or a suffix like `:h` when non-default components are selected
- **Zero-value default resolution** for components: zero/undefined/None = use default, don't show in mnemonic
- **Top-level `Mnemonic` and `Description`** on metadata
- **Per-indicator output enum** (TS/Python) in a dedicated file
- **Renamed imports** from `data`/`indicator`/`output` to `entities`/`core`/`outputs`

Use the SMA indicator as the canonical reference for single-constructor indicators.
Use the EMA indicator as the canonical reference for multi-constructor indicators
(see [Advanced: Multi-Constructor Indicators](#advanced-multi-constructor-indicators)).

---

## Conversion Workflow (New Indicators from External Sources)

When converting a new indicator from an external reference implementation into the zpano
architecture, choose the workflow based on the source language:

| Source Language | Pipeline |
|----------------|----------|
| **Go** | Go (new arch) → TS, Python, Zig, Rust |
| **Python** | Python (new arch) → Go → TS, Zig, Rust |
| **Rust** | Rust (new arch) → Go → TS, Python, Zig |

In all cases, the first step requires converting non-streaming logic to streaming:
replace array-based loops in the indicator's core calculation logic (the loops that
iterate over all input samples to produce output values) with an `Update(sample)`
method that accepts one value at a time, maintaining state across calls in struct
fields (running sums, previous values, ring buffers, etc.).

### If the external reference is in Go:

1. Convert to "new architecture" Go, adapting non-streaming (array-based) logic to our streaming (sample-by-sample `Update()`) architecture
2. From converted Go, port to: TypeScript, Python, Zig, Rust

### If the external reference is in Python:

1. Convert to "new architecture" Python, adapting non-streaming logic to streaming
2. From converted Python, port to Go
3. From converted Go, port to: TypeScript, Zig, Rust

### If the external reference is in Rust:

1. Convert to "new architecture" Rust, adapting non-streaming logic to streaming
2. From converted Rust, port to Go
3. From converted Go, port to: TypeScript, Python, Zig

### Post-conversion steps (all languages):

1. **Test data in separate files** — place test input and expected arrays in dedicated test data files:
   - Go: `testdata_test.go`
   - Python: `test_testdata.py`
   - TypeScript: `testdata.ts` (in spec folder)
   - Zig: `testdata.zig` (or at bottom of test file if small)
   - Rust: `testdata.rs` (or inline in test module if small)
2. **Include ALL parameter combinations** from the reference test suite (typically 9–15 combos).
3. **Register in factory** — add the indicator to the factory in each implemented language.
4. **Add to `icalc/settings.json`** — add an entry with default params so the CLI exercises it.
5. **Run icalc** — verify the indicator produces output without crashing in each language:
   ```bash
   cd go && go run ./cmd/icalc settings.json
   cd ts && npx tsx cmd/icalc/main.ts cmd/icalc/settings.json
   PYTHONPATH=. python3 -m py.cmd.icalc py/cmd/icalc/settings.json
   cd zig && zig build icalc -- src/cmd/icalc/settings.json
   cd rs && cargo run --bin icalc -- src/cmd/icalc/settings.json
   ```

---

## Table of Contents

1. [Naming & Style Conventions](#naming--style-conventions)
2. [Go Conversion](#go-conversion)
3. [TypeScript Conversion](#typescript-conversion)
4. [Python Implementation Guide](#python-implementation-guide)
5. [Zig Implementation Guide](#zig-implementation-guide)
6. [Rust Implementation Guide](#rust-implementation-guide)
7. [Quick Reference: Import Mapping](#quick-reference-import-mapping)
8. [Quick Reference: Symbol Renames](#quick-reference-symbol-renames)
9. [Advanced: Multi-Constructor Indicators](#advanced-multi-constructor-indicators)
10. [Advanced: Wrapper Indicators](#advanced-wrapper-indicators)
11. [Indicators That Do Not Use LineIndicator](#indicators-that-do-not-use-lineindicator)
12. [Advanced: Helper / Shared Component Families](#advanced-helper--shared-component-families)
13. [Registering in the Factory](#registering-in-the-factory)

---

## Naming & Style Conventions

All identifier, receiver, concurrency, style, and cross-language parity
rules are defined in the **`indicator-architecture`** skill (the authoritative
reference). Below is a quick-reference checklist of the most critical rules:

| Rule | Key Points |
|------|------------|
| **Banned abbreviations** | Always expand: `idx→index`, `tmp→temp`, `res→result`, `sig→signal`, `val→value`, `prev→previous`, `avg→average`, `mult→multiplier`, `buf→buffer`, `param→parameter`, `hist→histogram`. Allowed: Go idioms (`err`, `len`, `cap`, `min`, `max`, `num`), `Params`/`params` bundle type, TS `value_` ctor-param idiom. |
| **Go receivers** | Compound type (2+ words) → `s`; single word → first letter lowercased. All methods on a type MUST use the same receiver. |
| **Concurrency (Go)** | Stateful public indicators carry `mu sync.RWMutex`; writers lock/defer-unlock, readers rlock/defer-runlock. Exceptions: pure wrappers, internal engines. |
| **Go style invariants** | No `var x T = zero`; no split `var x T; x = expr`; use `any` not `interface{}`; `make([]T, 0)` must include capacity; no `new`; grouped imports; doc comments on all exports. |
| **Cross-language variable parity** | Same concept = same name in all languages (adjusted for casing). Canonical: `sum`, `epsilon`, `temp`/`diff`, `stddev`, `spread`, `bw`, `pctB`/`pct_b`, `amount`, `lengthMinOne`/`length_min_one`; loop counter `i`/`j`/`k`; `index` only when semantically a named index. |

When implementing in any language, copy the reference language's
local-variable names verbatim (adjusted to that language's casing
convention) where the concept is identical.

---

## Go Conversion

**Steps at a glance:** (1) Create package dir → (2) Params file → (3) Output file → (4) Output test → (5) Main indicator file → (6) Test file → (6A) Test data → (6B) Output test → (7) Register identifier/descriptor → (8) Add to icalc settings → (9) Verify.

### Go Step 1: Create the new package directory

Old indicators live in a flat `package indicators` directory. New indicators each get their own
package under `indicators/<group>/<indicatorname>/`.

**Action:** Create a new directory using the full descriptive name in `lowercase` (no separators).

```
# Example for SMA in the "common" group:
indicators/common/simplemovingaverage/
```

All files in this directory will use `package simplemovingaverage` (the directory name).

### Go Step 2: Convert the params file

**File:** `params.go`

Changes:

1. **Package declaration:** `package indicators` -> `package <indicatorname>`
2. **Import path:** `"mbg/trading/data"` -> `"zpano/entities"`
3. **Type references:**
   - `data.BarComponent` -> `entities.BarComponent`
   - `data.QuoteComponent` -> `entities.QuoteComponent`
   - `data.TradeComponent` -> `entities.TradeComponent`
4. **Add doc comments** to each component field explaining the zero-value default behavior:

```go
// BarComponent indicates the component of a bar to use when updating the indicator with a bar sample.
//
// If zero, the default (BarClosePrice) is used and the component is not shown in the indicator mnemonic.
BarComponent entities.BarComponent
```

Repeat the pattern for `QuoteComponent` (default: `QuoteMidPrice`) and `TradeComponent`
(default: `TradePrice`).

5. **Add `DefaultParams()`** at the end of the file. Every params file must export a
   `DefaultParams()` function that returns a pointer to a fully-populated struct with
   sensible defaults. Component fields are omitted (zero = use default). For dual-variant
   indicators, export `DefaultLengthParams()` and `DefaultSmoothingFactorParams()` instead.
   For no-params indicators (empty struct), still export `DefaultParams()` returning `&Params{}`.

```go
// DefaultParams returns a Params value populated with conventional defaults.
func DefaultParams() *Params {
    return &Params{
        Length: 14,
    }
}
```

See the `indicator-architecture` skill's DefaultParams section for default value
sources and naming conventions.

### Go Step 3: Convert the output file

**File:** `output.go`

Two mechanical changes:

1. `package indicators` → `package <indicatorname>`.
2. **Rename the output type and constants to the bare Go convention.** The
   package name provides scoping, so the indicator-name prefix would stutter
   (`simplemovingaverage.SimpleMovingAverageOutput`). Rename:
   - Type `<IndicatorName>Output` → `Output`.
   - Constant `<IndicatorName>Value` → `Value`.
   - Multi-output: strip the `<IndicatorName>` prefix and keep the descriptive
     suffix (e.g. `DirectionalMovementIndexAverageDirectionalIndex` →
     `AverageDirectionalIndex`). If stripping leaves a trailing `Value` with
     more text before it, drop that too (e.g. `AverageTrueRangeValue` →
     `AverageTrueRange`).

   Update all references in the package (main file, test files, `String()`,
   `IsKnown()`, `MarshalJSON()`, `UnmarshalJSON()`). The string forms returned
   by `String()`/`MarshalJSON()` stay unchanged — only the Go identifiers move.

> **TypeScript keeps the long form** (`SimpleMovingAverageOutput` /
> `SimpleMovingAverageValue`) because TS imports by symbol, not by module.
> Only the Go side strips the prefix.

### Go Step 4: Convert the output test file

**File:** `output_test.go`

**Only change:** `package indicators` -> `package <indicatorname>`

All test logic is identical.

### Go Step 5: Convert the main indicator file

**File:** `<indicatorname>.go`

This is the most involved change. Follow these sub-steps carefully.

#### 5a. Package and imports

Replace:

```go
package indicators

import (
    "mbg/trading/data"
    "mbg/trading/indicators/indicator"
    "mbg/trading/indicators/indicator/output"
)
```

With:

```go
package <indicatorname>

import (
    "zpano/entities"
    "zpano/indicators/core"
    "zpano/indicators/core/outputs"
)
```

Keep standard library imports (`fmt`, `math`, `sync`, etc.) as-is.

Remove the `"mbg/trading/indicators/indicator/output"` import entirely -- `BuildMetadata`
sources per-output `Kind`/`Shape` from the descriptor registry, so indicator files no
longer need to import `"zpano/indicators/core/outputs"` or `".../outputs/shape"` just for
`Metadata()`. Keep the `outputs` import only if the file uses `outputs.NewBand` /
`outputs.NewEmptyBand`.

#### 5b. Struct: remove boilerplate fields, embed LineIndicator

Old struct has these fields that must be **removed**:

```go
name        string
description string
barFunc     data.BarFunc
quoteFunc   data.QuoteFunc
tradeFunc   data.TradeFunc
```

**Add** the `LineIndicator` embedding:

```go
core.LineIndicator
```

The resulting struct keeps only the indicator's own state (e.g., `window`, `windowSum`, `primed`,
`mu sync.RWMutex`) plus the embedded `core.LineIndicator`.

**Important:** The `mu sync.RWMutex` field stays on the indicator struct. The `core.LineIndicator`
field goes after it (before the indicator-specific fields).

Example (SMA):

```go
type SimpleMovingAverage struct {
    mu sync.RWMutex
    core.LineIndicator
    window       []float64
    windowSum    float64
    windowLength int
    windowCount  int
    lastIndex    int
    primed       bool
}
```

#### 5c. Constructor: add default resolution and ComponentTripleMnemonic

**Before** calling `entities.BarComponentFunc(...)` etc., add a zero-value resolution block:

```go
// Resolve defaults for component functions.
// A zero value means "use default, don't show in mnemonic".
bc := p.BarComponent
if bc == 0 {
    bc = entities.DefaultBarComponent
}

qc := p.QuoteComponent
if qc == 0 {
    qc = entities.DefaultQuoteComponent
}

tc := p.TradeComponent
if tc == 0 {
    tc = entities.DefaultTradeComponent
}
```

Then pass the **resolved** values `bc`, `qc`, `tc` (not `p.BarComponent` etc.) to
`entities.BarComponentFunc`, `entities.QuoteComponentFunc`, `entities.TradeComponentFunc`.

**Change the mnemonic format** from:

```go
fmtn = "sma(%d)"
// ...
name := fmt.Sprintf(fmtn, length)
```

To:

```go
fmtn = "sma(%d%s)"
// ...
mnemonic := fmt.Sprintf(fmtn, length, core.ComponentTripleMnemonic(bc, qc, tc))
```

The variable name changes from `name` to `mnemonic`. The description uses the mnemonic:

```go
desc := "Simple moving average " + mnemonic
```

**Change the function type declarations** from `data.*Func` to `entities.*Func`:

```go
var (
    err       error
    barFunc   entities.BarFunc
    quoteFunc entities.QuoteFunc
    tradeFunc entities.TradeFunc
)
```

#### 5d. Constructor: replace struct literal with LineIndicator assignment

Old pattern -- returning a struct literal with embedded fields:

```go
return &SimpleMovingAverage{
    name:         name,
    description:  desc,
    window:       make([]float64, length),
    windowLength: length,
    lastIndex:    length - 1,
    barFunc:      barFunc,
    quoteFunc:    quoteFunc,
    tradeFunc:    tradeFunc,
}, nil
```

New pattern -- create the struct first, then assign `LineIndicator`:

```go
sma := &SimpleMovingAverage{
    window:       make([]float64, length),
    windowLength: length,
    lastIndex:    length - 1,
}

sma.LineIndicator = core.NewLineIndicator(mnemonic, desc, barFunc, quoteFunc, tradeFunc, sma.Update)

return sma, nil
```

**Critical:** `sma.Update` is a method reference passed to `NewLineIndicator`. This is what
allows `LineIndicator` to implement `UpdateScalar/Bar/Quote/Trade` without the indicator
having to define them. The struct must be created first so that `sma.Update` is a valid reference.

#### 5e. Metadata: use BuildMetadata

Old:

```go
func (s *SimpleMovingAverage) Metadata() indicator.Metadata {
    return indicator.Metadata{
        Type: indicator.SimpleMovingAverage,
        Outputs: []output.Metadata{
            {
                Kind:        int(SimpleMovingAverageValue),
                Type:        output.Scalar,
                Name:        s.name,
                Description: s.description,
            },
        },
    }
}
```

New — `BuildMetadata` pulls per-output `Kind` and `Shape` from the descriptor registry
(`go/indicators/core/descriptors.go`), so the caller supplies only mnemonic/description:

```go
func (s *SimpleMovingAverage) Metadata() core.Metadata {
    return core.BuildMetadata(
        core.SimpleMovingAverage,
        s.LineIndicator.Mnemonic,
        s.LineIndicator.Description,
        []core.OutputText{
            {Mnemonic: s.LineIndicator.Mnemonic, Description: s.LineIndicator.Description},
        },
    )
}
```

Key changes:

| Old | New |
|-----|-----|
| `indicator.Metadata` (return type) | `core.Metadata` (return type) |
| Inline `core.Metadata{...}` literal | `core.BuildMetadata(...)` call |
| `Type: indicator.SimpleMovingAverage` | First argument `core.SimpleMovingAverage` (the `core.Identifier` value) |
| `Name: s.name` | `{Mnemonic: s.LineIndicator.Mnemonic, Description: s.LineIndicator.Description}` per output |
| `output.Scalar` | (no longer specified — sourced from the descriptor registry) |
| `int(SimpleMovingAverageValue)` | (no longer specified — sourced from the descriptor registry) |
| `[]output.Metadata` | `[]core.OutputText` |

The caller NO LONGER imports `zpano/indicators/core/outputs` or
`zpano/indicators/core/outputs/shape` just for `Metadata()`. Drop those imports if they
are not referenced elsewhere in the file.

**Descriptor row required.** If no descriptor is registered for the indicator's identifier,
`BuildMetadata` panics at runtime. See Step 7 below — descriptor registration is now part
of indicator registration.

#### 5f. Delete UpdateScalar/Bar/Quote/Trade methods

**Remove entirely** the four methods:

- `UpdateScalar(*data.Scalar) indicator.Output`
- `UpdateBar(*data.Bar) indicator.Output`
- `UpdateQuote(*data.Quote) indicator.Output`
- `UpdateTrade(*data.Trade) indicator.Output`

These are now provided by the embedded `core.LineIndicator`.

#### 5g. Keep Update() and IsPrimed() as-is

The `Update(sample float64) float64` method contains the core algorithm and does not change.
The `IsPrimed()` method does not change.

### Go Step 6: Convert the test file

**File:** `<indicatorname>_test.go`

#### 6a. Package and imports

Replace:

```go
package indicators

import (
    "mbg/trading/data"
    "mbg/trading/indicators/indicator"
    "mbg/trading/indicators/indicator/output"
)
```

With:

```go
package <indicatorname>

import (
    "zpano/entities"
    "zpano/indicators/core"
    "zpano/indicators/core/outputs"
)
```

#### 6b. Entity type references

| Old | New |
|-----|-----|
| `data.Scalar` | `entities.Scalar` |
| `data.Bar` | `entities.Bar` |
| `data.Quote` | `entities.Quote` |
| `data.Trade` | `entities.Trade` |
| `indicator.Output` | `core.Output` |

#### 6c. Update entity tests (UpdateEntity tests)

The `check` function parameter type changes from `indicator.Output` to `core.Output`.

**Quote test:** If the old test uses `data.Quote{Time: time, Bid: inp}`, the new test must use
`entities.Quote{Time: time, Bid: inp, Ask: inp}` because the default quote component is
`QuoteMidPrice` (mid = (bid+ask)/2), which requires both fields. Using just `Bid: inp` would
produce a wrong mid price.

**Floating-point comparison:** Use tolerance-based comparison (`math.Abs(s.Value-exp) > 1e-13`)
instead of exact equality (`s.Value != exp`) for the scalar value check. The multi-stage EMA
indicators (T2, T3) can produce values that differ in the last significant digit due to
floating-point arithmetic ordering.

#### 6c-1. Multi-seeding algorithm test convergence (T2, T3)

When an indicator has two seeding algorithms (`firstIsAverage = true` vs `false` / Metastock),
the expected test data typically comes from one algorithm (usually `firstIsAverage = true` from
the spreadsheet). The other algorithm produces different values initially but converges after
some number of samples.

In Go tests, use a `firstCheck` offset:
- `firstIsAverage = true` (SMA-seeded): `firstCheck = lprimed` (or `lprimed + 1` if the
  first primed value differs due to averaging)
- `firstIsAverage = false` (Metastock): `firstCheck = lprimed + N` where N is the number
  of samples needed for convergence (e.g., 43 for T2 with length 5)

In TS tests, use the same pattern: `if (i >= lenPrimed + N) { expect(act).toBeCloseTo(...) }`

#### 6d. Constructor test ("length > 1" sub-test)

Old checks internal fields directly:

```go
check("name", "sma(5)", sma.name)
check("description", "Simple moving average sma(5)", sma.description)
check("barFunc == nil", false, sma.barFunc == nil)
check("quoteFunc == nil", false, sma.quoteFunc == nil)
check("tradeFunc == nil", false, sma.tradeFunc == nil)
```

New checks `LineIndicator` fields and removes function nil checks:

```go
check("mnemonic", "sma(5, hl/2)", sma.LineIndicator.Mnemonic)
check("description", "Simple moving average sma(5, hl/2)", sma.LineIndicator.Description)
```

The `barFunc/quoteFunc/tradeFunc == nil` checks are removed because these are now internal to
`LineIndicator`.

Note: the mnemonic includes `hl/2` because the "length > 1" test uses `BarMedianPrice` which is
not the default. If your test uses non-default components, adjust the expected mnemonic accordingly.

#### 6e. Component references in test constants and params

| Old | New |
|-----|-----|
| `data.BarComponent` | `entities.BarComponent` |
| `data.BarMedianPrice` | `entities.BarMedianPrice` |
| `data.BarClosePrice` | `entities.BarClosePrice` |
| `data.QuoteComponent` | `entities.QuoteComponent` |
| `data.QuoteMidPrice` | `entities.QuoteMidPrice` |
| `data.QuoteBidPrice` | `entities.QuoteBidPrice` |
| `data.TradeComponent` | `entities.TradeComponent` |
| `data.TradePrice` | `entities.TradePrice` |
| `data.BarComponent(9999)` | `entities.BarComponent(9999)` |
| etc. | etc. |

#### 6f. Metadata test

Old:

```go
check("Outputs[0].Type", output.Scalar, act.Outputs[0].Type)
check("Outputs[0].Name", "sma(5)", act.Outputs[0].Name)
```

New:

```go
check("Identifier", core.SimpleMovingAverage, act.Identifier)
check("Outputs[0].Shape", shape.Scalar, act.Outputs[0].Shape)
check("Outputs[0].Mnemonic", "sma(5)", act.Outputs[0].Mnemonic)
```

(`shape` is `zpano/indicators/core/outputs/shape`. The test file is one of the few places
that still imports it, since the indicator file no longer does.)

Also verify that the top-level `Mnemonic` and `Description` are checked on `act`:

```go
// These checks are implicitly done via Type already existing in old tests,
// but verify the new Mnemonic and Description fields are present.
```

#### 6g. Test helper: use zero-value defaults

Old test helper passes explicit component values:

```go
func testSimpleMovingAverageCreate(length int) *SimpleMovingAverage {
    params := SimpleMovingAverageParams{
        Length: length, BarComponent: data.BarClosePrice, QuoteComponent: data.QuoteBidPrice, TradeComponent: data.TradePrice,
    }
    sma, _ := NewSimpleMovingAverage(&params)
    return sma
}
```

New test helper uses zero-value defaults (the standard usage pattern):

```go
func testSimpleMovingAverageCreate(length int) *SimpleMovingAverage {
    params := SimpleMovingAverageParams{
        Length: length,
    }
    sma, _ := NewSimpleMovingAverage(&params)
    return sma
}
```

**Important:** If the old helper passed `QuoteBidPrice`, the old "update quote" test could use
`data.Quote{Time: time, Bid: inp}`. The new helper uses the default `QuoteMidPrice`, so the quote
test must provide both `Bid` and `Ask`: `entities.Quote{Time: time, Bid: inp, Ask: inp}`.

#### 6h. Add mnemonic sub-tests

Add sub-tests that verify mnemonic generation for various component combinations. These test the
zero-value default omission behavior:

```go
t.Run("all components zero", func(t *testing.T) {
    t.Parallel()
    params := SimpleMovingAverageParams{Length: length}
    sma, err := NewSimpleMovingAverage(&params)
    check("err == nil", true, err == nil)
    check("mnemonic", "sma(5)", sma.LineIndicator.Mnemonic)
    check("description", "Simple moving average sma(5)", sma.LineIndicator.Description)
})

t.Run("only bar component set", func(t *testing.T) {
    t.Parallel()
    params := SimpleMovingAverageParams{Length: length, BarComponent: entities.BarMedianPrice}
    sma, err := NewSimpleMovingAverage(&params)
    check("err == nil", true, err == nil)
    check("mnemonic", "sma(5, hl/2)", sma.LineIndicator.Mnemonic)
    check("description", "Simple moving average sma(5, hl/2)", sma.LineIndicator.Description)
})

t.Run("only quote component set", func(t *testing.T) {
    t.Parallel()
    params := SimpleMovingAverageParams{Length: length, QuoteComponent: entities.QuoteBidPrice}
    sma, err := NewSimpleMovingAverage(&params)
    check("err == nil", true, err == nil)
    check("mnemonic", "sma(5, b)", sma.LineIndicator.Mnemonic)
    check("description", "Simple moving average sma(5, b)", sma.LineIndicator.Description)
})

t.Run("only trade component set", func(t *testing.T) {
    t.Parallel()
    params := SimpleMovingAverageParams{Length: length, TradeComponent: entities.TradeVolume}
    sma, err := NewSimpleMovingAverage(&params)
    check("err == nil", true, err == nil)
    check("mnemonic", "sma(5, v)", sma.LineIndicator.Mnemonic)
    check("description", "Simple moving average sma(5, v)", sma.LineIndicator.Description)
})

t.Run("bar and quote components set", func(t *testing.T) {
    t.Parallel()
    params := SimpleMovingAverageParams{
        Length: length, BarComponent: entities.BarOpenPrice, QuoteComponent: entities.QuoteBidPrice,
    }
    sma, err := NewSimpleMovingAverage(&params)
    check("err == nil", true, err == nil)
    check("mnemonic", "sma(5, o, b)", sma.LineIndicator.Mnemonic)
    check("description", "Simple moving average sma(5, o, b)", sma.LineIndicator.Description)
})

t.Run("bar and trade components set", func(t *testing.T) {
    t.Parallel()
    params := SimpleMovingAverageParams{
        Length: length, BarComponent: entities.BarHighPrice, TradeComponent: entities.TradeVolume,
    }
    sma, err := NewSimpleMovingAverage(&params)
    check("err == nil", true, err == nil)
    check("mnemonic", "sma(5, h, v)", sma.LineIndicator.Mnemonic)
    check("description", "Simple moving average sma(5, h, v)", sma.LineIndicator.Description)
})

t.Run("quote and trade components set", func(t *testing.T) {
    t.Parallel()
    params := SimpleMovingAverageParams{
        Length: length, QuoteComponent: entities.QuoteAskPrice, TradeComponent: entities.TradeVolume,
    }
    sma, err := NewSimpleMovingAverage(&params)
    check("err == nil", true, err == nil)
    check("mnemonic", "sma(5, a, v)", sma.LineIndicator.Mnemonic)
    check("description", "Simple moving average sma(5, a, v)", sma.LineIndicator.Description)
})
```

Adapt the mnemonics and component values to your specific indicator. The pattern is always:

- Default component = omitted from mnemonic
- Non-default component = shown as its mnemonic abbreviation (e.g., `hl/2`, `o`, `h`, `b`, `a`, `v`)

### Go Step 6-A: Create testdata_test.go (separate test data file)

**File:** `testdata_test.go`

When converting from Python, put all test input data and expected output arrays in a separate
`testdata_test.go` file rather than inlining them in the main `*_test.go` file. This keeps
the test logic readable and the data maintainable.

**Structure:**

```go
//nolint:testpackage
package <indicatorname>

import "math"

// testInput is the shared 252-bar close-price series used across all indicator tests.
var testInput = []float64{
    // 10 values per line, full 15-digit precision
    50.1, 51.2, ...
}

// expectedLen10 is the expected output for length=10.
var expectedLen10 = []float64{
    math.NaN(), math.NaN(), ...,
    0.123456789012345, ...
}

// expectedLen14 is the expected output for length=14.
var expectedLen14 = []float64{ ... }

// expectedDefault is the default-params expected output (alias).
var expectedDefault = expectedLen14
```

**Multiple parameter combinations:** You MUST include expected arrays for ALL parameter
combinations present in the Python `test_testdata.py` (typically 9–15 combos). Do not skip
any — the goal is full parity with the Python test suite. Name them descriptively:
`expectedLen10`, `expectedLo2Hi15`, `expectedSens05`, etc.

### Go Step 6-B: Create output_test.go

**File:** `output_test.go`

Every indicator package MUST have an `output_test.go` that tests the output enum's methods.
Follow the JMA pattern:

```go
//nolint:testpackage
package <indicatorname>

import (
    "encoding/json"
    "testing"
)

func TestOutputString(t *testing.T) {
    t.Parallel()
    tests := []struct {
        output Output
        want   string
    }{
        {OutputValue, "value"},
        // ... all valid outputs
        {outputLast, "unknown"},
        {Output(99), "unknown"},
    }
    for _, tt := range tests {
        if got := tt.output.String(); got != tt.want {
            t.Errorf("Output(%d).String() = %q, want %q", tt.output, got, tt.want)
        }
    }
}

func TestOutputIsKnown(t *testing.T) {
    t.Parallel()
    // Test all valid outputs return true, outputLast and beyond return false
}

func TestOutputMarshalJSON(t *testing.T) {
    t.Parallel()
    // Test valid outputs marshal to quoted strings, unknown returns error
}

func TestOutputUnmarshalJSON(t *testing.T) {
    t.Parallel()
    // Test valid strings unmarshal correctly, unknown/invalid returns error
}
```

### Go Step 7: Register the identifier and descriptor

1. **Register the identifier.** If the constant does not already exist in
   `go/indicators/core/identifier.go`, add it:

   ```go
   const (
       SimpleMovingAverage Identifier = iota + 1
       // ...
   )
   ```

2. **Register the descriptor.** Add a row to the registry in
   `go/indicators/core/descriptors.go`. The descriptor drives the taxonomy (role, pane,
   volume usage, adaptivity, family) **and** supplies per-output `Kind`/`Shape` consumed by
   `BuildMetadata`. See `.opencode/skills/indicator-architecture/SKILL.md` section
   "Taxonomy & Descriptor Registry" for field meanings and guidance. The order of the
   `Outputs` slice must match the order of the `[]OutputText` passed by the indicator's
   `Metadata()` method.

### Go Step 8: Add to icalc settings

Add the new indicator to `go/cmd/icalc/settings.json` so it is exercised by the CLI tool:

```json
{ "identifier": "myIndicator", "params": { "length": 14 } }
```

The `identifier` string uses **camelCase** (matching the factory's string mapping).

Then run icalc to verify it doesn't crash:

```bash
cd go && go run ./cmd/icalc settings.json
```

### Go Step 9: Verify

```bash
cd go && go test ./indicators/<group>/<indicatorname>/...
```

All tests must pass. If the old tests are still present in the `_old` folder, you may need to
exclude them or ensure they still compile (they reference the old `mbg/trading` module, so they
won't be picked up by the new module).

---

## TypeScript Conversion

### TS Step 1: Create the new folder

Old indicators live in a folder with no per-indicator output file and use `.interface` and `.enum`
suffixes. New indicators each get a clean folder.

**Action:** Create a new directory using the full descriptive name in `kebab-case`.

```
# Example for SMA in the "common" group:
indicators/common/simple-moving-average/
```

The new folder will contain 4 files (old had 3):

| Old files | New files |
|-----------|-----------|
| `simple-moving-average-params.interface.ts` | `params.ts` |
| `simple-moving-average.ts` | `simple-moving-average.ts` |
| `simple-moving-average.spec.ts` | `simple-moving-average.spec.ts` |
| (none) | `output.ts` |

### TS Step 2: Convert the params file

**File:** `params.ts` (renamed from `<indicator-name>-params.interface.ts` — prefix dropped, `.interface` suffix dropped)

Changes:

1. **Rename file:** Drop the `.interface` suffix.
2. **Import paths:** Update from `data/entities` to `entities`:

| Old | New |
|-----|-----|
| `'../../../data/entities/bar-component.enum'` | `'../../../entities/bar-component'` |
| `'../../../data/entities/quote-component.enum'` | `'../../../entities/quote-component'` |

3. **Add `TradeComponent`** if missing:

```typescript
import { TradeComponent } from '../../../entities/trade-component';
```

And add the field to the interface:

```typescript
/**
 * A component of a trade to use when updating the indicator with a trade sample.
 *
 * If _undefined_, the trade component will have a default value and will not be shown
 * in the indicator mnemonic.
 */
tradeComponent?: TradeComponent;
```

4. **Add `defaultParams()`** at the end of the file. Every params file must export a
   `defaultParams()` function that returns a plain object with sensible defaults.
   Component fields are omitted (`undefined` = use default). For dual-variant indicators,
   export `defaultLengthParams()` and `defaultSmoothingFactorParams()` instead. For
   no-params indicators (empty interface), still export `defaultParams()` returning `{}`.

```typescript
export function defaultParams(): SimpleMovingAverageParams {
    return {
        length: 14,
    };
}
```

See the `indicator-architecture` skill's DefaultParams section for default value
sources and naming conventions.

### TS Step 3: Create the output file

**File:** `output.ts` (NEW file -- does not exist in old structure)

Create a per-indicator output enum:

```typescript
/** Describes the outputs of the indicator. */
export enum SimpleMovingAverageOutput {
    /** The scalar value of the moving average. */
    Value = 'value',
}
```

Adapt the enum name and members to your specific indicator.

### TS Step 4: Convert the main indicator file

**File:** `<indicator-name>.ts`

#### 4a. Imports

Replace old imports with new ones:

| Old import | New import |
|------------|------------|
| `'../indicator/component-pair-mnemonic'` | `'../../core/component-triple-mnemonic'` |
| `'../indicator/line-indicator'` | `'../../core/line-indicator'` |
| `'./<name>-params.interface'` | `'./<name>-params'` |

Add new imports:

```typescript
import { buildMetadata } from '../../core/build-metadata';
import { IndicatorIdentifier } from '../../core/indicator-identifier';
import { IndicatorMetadata } from '../../core/indicator-metadata';
import { SimpleMovingAverageOutput } from './simple-moving-average-output';
```

Note: `OutputType`/`output-type` and `IndicatorType`/`indicator-type` are gone. `buildMetadata`
sources the per-output `Shape` and `Kind` from the descriptor registry
(`ts/indicators/core/descriptors.ts`), so the indicator file no longer imports them.

Change the mnemonic function import:

```typescript
// Old:
import { componentPairMnemonic } from '../indicator/component-pair-mnemonic';
// New:
import { componentTripleMnemonic } from '../../core/component-triple-mnemonic';
```

#### 4b. Constructor: mnemonic function

Replace:

```typescript
const m = componentPairMnemonic(params.barComponent, params.quoteComponent);
```

With:

```typescript
const m = componentTripleMnemonic(params.barComponent, params.quoteComponent, params.tradeComponent);
```

#### 4c. Constructor: add description

After the mnemonic assignment, add:

```typescript
this.description = 'Simple moving average ' + this.mnemonic;
```

(Adapt the description text to your indicator.)

#### 4d. Constructor: add component setter calls

After the description, add three component setter calls:

```typescript
this.barComponent = params.barComponent;
this.quoteComponent = params.quoteComponent;
this.tradeComponent = params.tradeComponent;
```

These call the `LineIndicator` setters which handle `undefined` -> default resolution internally.

#### 4e. Add metadata() method

Add a new `metadata()` method to the class. `buildMetadata` reads the per-output `Shape` and
`Kind` from the descriptor registry (`ts/indicators/core/descriptors.ts`), so the caller only
supplies mnemonic/description text:

```typescript
/** Describes the output data of the indicator. */
metadata(): IndicatorMetadata {
    return buildMetadata(
        IndicatorIdentifier.SimpleMovingAverage,
        this.mnemonic,
        this.description,
        [{ mnemonic: this.mnemonic, description: this.description }],
    );
}
```

Adapt `IndicatorIdentifier.*` and the per-output text entries to your specific indicator. The
number and order of the `OutputText` entries must match the descriptor row registered in
`descriptors.ts` — `buildMetadata` throws if they disagree.

**Descriptor row required.** If no descriptor is registered for the identifier, `buildMetadata`
throws at runtime. See TS Step 7 below — descriptor registration is now part of indicator
registration.

### TS Step 5: Convert the test file

**File:** `<indicator-name>.spec.ts`

#### 5a. Mnemonic assertion

Old:

```typescript
expect(sma.getMnemonic()).toBe('sma(7)');
```

New:

```typescript
expect(sma.metadata().mnemonic).toBe('sma(7)');
```

The `getMnemonic()` method is replaced by accessing `metadata().mnemonic`.

#### 5b. Other test logic

The `update()`, `isPrimed()`, and NaN tests remain **identical**. No changes needed for the core
calculation tests.

### TS Step 6: Register the identifier and descriptor

1. **Register the identifier.** If the member does not already exist in
   `ts/indicators/core/indicator-identifier.ts`, add it.

2. **Register the descriptor.** Add a row to the registry in
   `ts/indicators/core/descriptors.ts`. The descriptor drives the taxonomy and supplies
   per-output `kind`/`shape` consumed by `buildMetadata`. See
   `.opencode/skills/indicator-architecture/SKILL.md` section "Taxonomy & Descriptor
   Registry" for field meanings. The order of the `outputs` entries must match the order
   of the `OutputText[]` passed by the indicator's `metadata()` method.

### TS Step 7: Verify

Run the tests for the specific indicator:

```bash
# zpano uses Jasmine (not Jest) via tsx. From /ts:
npm test                                    # full suite
node --import tsx/esm ./node_modules/.bin/jasmine --config=jasmine.json --filter=SineWave
# --filter is a case-sensitive regex substring match against the spec describe() name.
```

All tests must pass.

---

## Quick Reference: Import Mapping

### Go

| Old import path | New import path |
|-----------------|-----------------|
| `"mbg/trading/data"` | `"zpano/entities"` |
| `"mbg/trading/indicators/indicator"` | `"zpano/indicators/core"` |
| `"mbg/trading/indicators/indicator/output"` | `"zpano/indicators/core/outputs"` (only if `outputs.NewBand`/`outputs.NewEmptyBand` is used; the `Shape` enum now lives at `"zpano/indicators/core/outputs/shape"`, but indicator files rarely import it directly — `BuildMetadata` sources `Shape` from the registry) |

### TypeScript

| Old import path (relative from indicator) | New import path (relative from indicator) |
|------------------------------------------|------------------------------------------|
| `'../indicator/component-pair-mnemonic'` | `'../../core/component-triple-mnemonic'` |
| `'../indicator/line-indicator'` | `'../../core/line-indicator'` |
| `'./<name>-params.interface'` | `'./<name>-params'` |
| `'../../../data/entities/bar-component.enum'` | `'../../../entities/bar-component'` |
| `'../../../data/entities/quote-component.enum'` | `'../../../entities/quote-component'` |
| (none) | `'../../../entities/trade-component'` |
| (none) | `'../../core/build-metadata'` |
| (none) | `'../../core/indicator-identifier'` |
| (none) | `'../../core/indicator-metadata'` |
| (none) | `'./<name>-output'` |

---

## Quick Reference: Symbol Renames

### Go

| Old symbol | New symbol |
|------------|------------|
| `data.Scalar` | `entities.Scalar` |
| `data.Bar` | `entities.Bar` |
| `data.Quote` | `entities.Quote` |
| `data.Trade` | `entities.Trade` |
| `data.BarFunc` | `entities.BarFunc` |
| `data.QuoteFunc` | `entities.QuoteFunc` |
| `data.TradeFunc` | `entities.TradeFunc` |
| `data.BarComponent` | `entities.BarComponent` |
| `data.QuoteComponent` | `entities.QuoteComponent` |
| `data.TradeComponent` | `entities.TradeComponent` |
| `data.BarClosePrice` | `entities.BarClosePrice` |
| `data.BarMedianPrice` | `entities.BarMedianPrice` |
| `data.QuoteMidPrice` | `entities.QuoteMidPrice` |
| `data.QuoteBidPrice` | `entities.QuoteBidPrice` |
| `data.TradePrice` | `entities.TradePrice` |
| `data.BarComponentFunc(...)` | `entities.BarComponentFunc(...)` |
| `data.QuoteComponentFunc(...)` | `entities.QuoteComponentFunc(...)` |
| `data.TradeComponentFunc(...)` | `entities.TradeComponentFunc(...)` |
| `indicator.Metadata` | `core.Metadata` |
| `indicator.Output` | `core.Output` |
| `indicator.SimpleMovingAverage` | `core.SimpleMovingAverage` (now a `core.Identifier`, not an `IndicatorType`) |
| `indicator.Type` / `IndicatorType` | `core.Identifier` |
| `output.Metadata` | `outputs.Metadata` |
| `output.Scalar` | `shape.Scalar` (package `zpano/indicators/core/outputs/shape`) — but indicator files no longer write this; sourced from the descriptor registry |
| `outputs.ScalarType` / `outputs.Type` | `shape.Scalar` / `shape.Shape` (same note — sourced from the registry) |

### Go: Struct field / method changes

| Old | New |
|-----|-----|
| `s.name` | `s.LineIndicator.Mnemonic` |
| `s.description` | `s.LineIndicator.Description` |
| `s.barFunc` | (removed -- inside LineIndicator) |
| `s.quoteFunc` | (removed -- inside LineIndicator) |
| `s.tradeFunc` | (removed -- inside LineIndicator) |
| `Outputs[i].Name` | `Outputs[i].Mnemonic` |

### Go: Metadata field additions

| Old (not present) | New |
|-------------------|-----|
| (none) | `Mnemonic: s.LineIndicator.Mnemonic` |
| (none) | `Description: s.LineIndicator.Description` |

### TypeScript

| Old symbol / pattern | New symbol / pattern |
|----------------------|----------------------|
| `componentPairMnemonic(bar, quote)` | `componentTripleMnemonic(bar, quote, trade)` |
| `sma.getMnemonic()` | `sma.metadata().mnemonic` |
| (no description) | `this.description = '...' + this.mnemonic` |
| (no component setters) | `this.barComponent = params.barComponent` |
| (no component setters) | `this.quoteComponent = params.quoteComponent` |
| (no component setters) | `this.tradeComponent = params.tradeComponent` |
| (no metadata method) | `metadata(): IndicatorMetadata { return buildMetadata(...); }` |
| `IndicatorType` | `IndicatorIdentifier` (from `core/indicator-identifier`) |
| `OutputType.Scalar` | `Shape.Scalar` (from `core/outputs/shape/shape`) — but indicator files no longer write this; sourced from the descriptor registry |
| (no output file) | `output.ts` with output enum |

---

## Advanced: Multi-Constructor Indicators

Some indicators have **multiple constructors** that create the same type with different
parameterization. The EMA (Exponential Moving Average) is the canonical example: it can be
constructed from a **length** or from a **smoothing factor (alpha)**.

The base conversion guide above assumes a single constructor. This section covers the
additional considerations for multi-constructor indicators.

### Multiple Param Structs / Interfaces

Each constructor path gets its own param struct (Go/Zig/Rust), interface (TS), or dataclass (Python).

**Go:**

```go
type ExponentialMovingAverageLengthParams struct {
    Length        int
    BarComponent  entities.BarComponent
    // ...
}

type ExponentialMovingAverageSmoothingFactorParams struct {
    SmoothingFactor float64
    BarComponent    entities.BarComponent
    // ...
}
```

**TS:**

```ts
export interface ExponentialMovingAverageLengthParams {
    length: number;
    barComponent?: BarComponent;
    // ...
}

export interface ExponentialMovingAverageSmoothingFactorParams {
    smoothingFactor: number;
    barComponent?: BarComponent;
    // ...
}
```

**Python:**

```python
@dataclass
class ExponentialMovingAverageLengthParams:
    length: int = 10
    bar_component: Optional[BarComponent] = None
    # ...

@dataclass
class ExponentialMovingAverageSmoothingFactorParams:
    smoothing_factor: float = 0.18181818
    bar_component: Optional[BarComponent] = None
    # ...
```

**Zig:**

```zig
pub const ExponentialMovingAverageLengthParams = struct {
    length: u32 = 10,
    bar_component: ?BarComponent = null,
    // ...
};

pub const ExponentialMovingAverageSmoothingFactorParams = struct {
    smoothing_factor: f64 = 0.18181818,
    bar_component: ?BarComponent = null,
    // ...
};
```

**Rust:**

```rust
pub struct ExponentialMovingAverageLengthParams {
    pub length: usize,
    pub bar_component: Option<BarComponent>,
    // ...
}

pub struct ExponentialMovingAverageSmoothingFactorParams {
    pub smoothing_factor: f64,
    pub bar_component: Option<BarComponent>,
    // ...
}
```

Component fields (`BarComponent`, `QuoteComponent`, `TradeComponent`) should be duplicated
across all param structs/interfaces with the same default-resolution doc comments.

### Shared Private Constructor (Go)

In Go, use a **private** shared constructor that both public constructors delegate to:

```go
func NewExponentialMovingAverageLength(p *ExponentialMovingAverageLengthParams) (*ExponentialMovingAverage, error) {
    // validate length, compute smoothingFactor from length
    return newExponentialMovingAverage(length, smoothingFactor, p.BarComponent, ...)
}

func NewExponentialMovingAverageSmoothingFactor(p *ExponentialMovingAverageSmoothingFactorParams) (*ExponentialMovingAverage, error) {
    // validate smoothingFactor, compute length from smoothingFactor
    return newExponentialMovingAverage(length, smoothingFactor, p.BarComponent, ...)
}

func newExponentialMovingAverage(length int, smoothingFactor float64, bc entities.BarComponent, ...) (*ExponentialMovingAverage, error) {
    // default resolution, ComponentTripleMnemonic, two-step construction
}
```

### Multiple Static Methods (TS)

In TypeScript, use **static factory methods** instead of overloading the constructor:

```ts
export class ExponentialMovingAverage extends LineIndicator {
    static fromLength(params: ExponentialMovingAverageLengthParams): ExponentialMovingAverage { ... }
    static fromSmoothingFactor(params: ExponentialMovingAverageSmoothingFactorParams): ExponentialMovingAverage { ... }

    private constructor(...) { ... }
}
```

Both static methods delegate to the private constructor with resolved parameters.

### Python: Static Factory Methods

```python
class ExponentialMovingAverage(Indicator):
    @staticmethod
    def from_length(params: ExponentialMovingAverageLengthParams) -> 'ExponentialMovingAverage': ...

    @staticmethod
    def from_smoothing_factor(params: ExponentialMovingAverageSmoothingFactorParams) -> 'ExponentialMovingAverage': ...
```

### Zig: Separate `init` Functions

```zig
pub fn initFromLength(allocator: Allocator, params: ExponentialMovingAverageLengthParams) !ExponentialMovingAverage { ... }
pub fn initFromSmoothingFactor(allocator: Allocator, params: ExponentialMovingAverageSmoothingFactorParams) !ExponentialMovingAverage { ... }
```

### Rust: Associated Functions

```rust
impl ExponentialMovingAverage {
    pub fn from_length(params: ExponentialMovingAverageLengthParams) -> Self { ... }
    pub fn from_smoothing_factor(params: ExponentialMovingAverageSmoothingFactorParams) -> Self { ... }
}
```

### Constructor-Specific Mnemonic Formats

Different constructor paths may produce different mnemonic formats. Define the format
at the point where you know which constructor path was taken:

```go
// length-based:
mnemonic := fmt.Sprintf("ema(%d%s)", length, core.ComponentTripleMnemonic(bc, qc, tc))
// => "ema(10)" or "ema(10, hl/2)"

// smoothing-factor-based (includes both computed length and the explicit factor):
mnemonic := fmt.Sprintf("ema(%d, %.8f%s)", length, smoothingFactor, core.ComponentTripleMnemonic(bc, qc, tc))
// => "ema(10, 0.18181818)" or "ema(10, 0.18181818, hl/2)"
```

The mnemonics must match across all five languages. This may require changing
a language's format from the old style to match the reference (e.g., old TS
alpha path used `ema(0.123)` with 3 decimal places; new uses
`ema(10, 0.18181818)` with 8 decimal places to match Go).

### Cross-Language Behavior Alignment

When converting multi-constructor indicators, check for behavioral differences
across languages and decide whether to align them:

- **Priming behavior:** If one language primes differently for a given constructor path,
  consider aligning. Example: old TS smoothing-factor EMA primed immediately (length=0),
  while Go computed length from alpha and waited. New TS computes
  `length = Math.round(2/alpha) - 1` to match Go.

- **Validation:** Validation differences may be intentional (stricter in one language) and
  can be preserved. Document any known differences.

### Test Considerations

Each constructor path needs its own test group:

```go
t.Run("from length", func(t *testing.T) { ... })
t.Run("from smoothing factor", func(t *testing.T) { ... })
```

```ts
describe('from length', () => { ... });
describe('from smoothing factor', () => { ... });
```

```python
def test_from_length(self): ...
def test_from_smoothing_factor(self): ...
```

```zig
test "ema from length" { ... }
test "ema from smoothing factor" { ... }
```

```rust
#[test]
fn test_from_length() { ... }
#[test]
fn test_from_smoothing_factor() { ... }
```

Mnemonic tests should cover both paths, including with non-default components, to verify
that each constructor path produces the correct format.

---

## Advanced: Wrapper Indicators

Some indicators **delegate their core calculation** to another indicator internally. The Standard
Deviation indicator is the canonical example: it wraps a Variance indicator and returns the square
root of the variance value.

### Key Differences from Standard Indicators

1. **Dual embedding:** The wrapper struct embeds both `core.LineIndicator` (for the update
   protocol) and holds a pointer to the wrapped indicator.

2. **The wrapper has its own `Update()` method** that:
   - Calls the wrapped indicator's `Update()`
   - Transforms the result (e.g., `math.Sqrt(varianceResult)`)
   - Returns the transformed value

3. **`IsPrimed()` delegates** to the wrapped indicator (the wrapper itself has no separate
   priming logic).

4. **Constructor creates the wrapped indicator internally** using the same params:

```go
// Go example: Standard Deviation wrapping Variance
type StandardDeviation struct {
    core.LineIndicator
    variance *variance.Variance
}

func NewStandardDeviation(p *StandardDeviationParams) (*StandardDeviation, error) {
    // Resolve component defaults
    // Create a variance.Variance internally
    vp := &variance.VarianceParams{
        Length: p.Length, IsUnbiased: p.IsUnbiased,
        BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
    }
    v, err := variance.NewVariance(vp)
    if err != nil {
        return nil, err
    }

    // Also create own component functions for LineIndicator
    // Build mnemonic and description
    sd := &StandardDeviation{variance: v}
    sd.LineIndicator = core.NewLineIndicator(mnemonic, desc, barFunc, quoteFunc, tradeFunc, sd.Update)
    return sd, nil
}
```

```typescript
// TS example: Standard Deviation wrapping Variance
export class StandardDeviation extends LineIndicator {
    private variance: Variance;

    public constructor(params: StandardDeviationParams) {
        super();
        // Validate, build mnemonic, set components
        this.variance = new Variance(params);
        this.primed = false;
    }

    public update(sample: number): number {
        const value = this.variance.update(sample);
        if (Number.isNaN(value)) return value;
        this.primed = this.variance.isPrimed();
        return Math.sqrt(value);
    }
}
```

### Important: Separate Packages / Folders

In the new architecture, the wrapper and wrapped indicators live in **separate packages** (Go)
or **separate folders** (TS), even if they were in the same package before:

- Go: `indicators/common/standarddeviation/` imports `zpano/indicators/common/variance`
- TS: `indicators/common/standard-deviation/` imports `'../variance/variance'`

### Old Multi-Output to New Single-Output

The old Standard Deviation was a **multi-output** indicator (SD value + optional variance value).
In the new architecture, it is a **single-output** `LineIndicator`:

- `Update(float64) float64` returns only the standard deviation value
- Users who want both SD and variance should create both indicators separately
- The per-indicator output enum has only one member (e.g., `StandardDeviationValue`)

This simplification is required because `LineIndicator.UpdateScalar` always returns a
single-element output. If the old indicator had multiple outputs, remove the extra outputs and
direct users to the standalone indicator for those values.

### Constructor Resolves Components Before Passing to Wrapped Indicator

The wrapper must resolve zero-value component defaults **before** creating the wrapped indicator.
This ensures both the wrapper's mnemonic and the wrapped indicator use the same resolved values:

```go
bc := p.BarComponent
if bc == 0 {
    bc = entities.DefaultBarComponent
}
// ... resolve qc, tc similarly

vp := &variance.VarianceParams{
    Length: p.Length, IsUnbiased: p.IsUnbiased,
    BarComponent: bc, QuoteComponent: qc, TradeComponent: tc,
}
```

### Mnemonic Pattern

Each indicator has its own mnemonic prefix. The wrapper does **not** reuse the wrapped indicator's
mnemonic:

- Variance: `var.s(5)` / `var.p(5)`
- Standard Deviation: `stdev.s(5)` / `stdev.p(5)`

The `s`/`p` suffix denotes sample (unbiased) vs. population (biased) estimation.

### Test Considerations

- The wrapper's test file defines its **own test data and expected values** (it cannot share test
  helpers from the wrapped indicator's package, since they are in different packages)
- Error messages from validation may come from the wrapped indicator (e.g.,
  `"invalid variance parameters: length should be greater than 1"`)
- The `IsPrimed()` test verifies delegation: the wrapper is primed when the wrapped indicator is

---

## Indicators That Do Not Use LineIndicator

Some indicators cannot use `LineIndicator` because they:
1. Have **multiple outputs** (e.g., FRAMA outputs both a value and a fractal dimension)
2. Need **high/low data** from bars/quotes, not just a single scalar input
3. The `Update(float64) float64` / `update(sample: number): number` signature doesn't fit

### Pattern: Implement `Indicator` Interface Directly

**Go**: The struct does NOT embed `core.LineIndicator`. Instead, it:
- Stores its own `barFunc`, `quoteFunc`, `tradeFunc` component functions
- Implements `UpdateScalar`, `UpdateBar`, `UpdateQuote`, `UpdateTrade` directly
- Has a private `updateEntity` helper that calls `Update` and builds the `core.Output` slice
- Uses `core.ComponentTripleMnemonic` for mnemonic construction (same as LineIndicator-based indicators)

**TS**: The class `implements Indicator` instead of `extends LineIndicator`. It:
- Stores its own `barComponentFunc`, `quoteComponentFunc`, `tradeComponentFunc` functions
- Resolves defaults with `?? DefaultBarComponent` etc. (same logic as `LineIndicator`'s protected setters)
- Implements `updateScalar`, `updateBar`, `updateQuote`, `updateTrade` directly
- Has a private `updateEntity` helper that calls `update` and builds the `IndicatorOutput` array
- Uses `componentTripleMnemonic` for mnemonic construction

**Python**: The class extends `Indicator` (ABC) without using `LineIndicator`. It:
- Stores `self._bar_func`, `self._quote_func`, `self._trade_func` directly
- Resolves defaults with `if x is None: x = DEFAULT_*`
- Implements `update_scalar`, `update_bar`, `update_quote`, `update_trade` directly
- Uses `component_triple_mnemonic` for mnemonic construction

**Zig**: The struct does NOT use `LineIndicator`. It:
- Stores `bar_func`, `quote_func`, `trade_func` function pointers directly
- Resolves defaults with `orelse default_*_component`
- Implements `updateBar`, `updateQuote`, `updateTrade`, `updateScalar` directly
- Builds `OutputArray` manually with multiple entries
- Uses `componentTripleMnemonic` for mnemonic construction

**Rust**: The struct does NOT use `LineIndicator`. It:
- Stores `bar_func: fn(&Bar) -> f64`, `quote_func`, `trade_func` directly
- Resolves defaults with `.unwrap_or(DEFAULT_*)`
- Implements `update_bar`, `update_quote`, `update_trade`, `update_scalar` directly
- Returns `Vec<Box<dyn Any>>` with multiple output entries
- Uses `component_triple_mnemonic` for mnemonic construction

### Example: FRAMA (Fractal Adaptive Moving Average)

FRAMA has two outputs (`Value` and `Fdim`), and its core `Update` takes three parameters
(`sample`, `sampleHigh`, `sampleLow`).

**Go signature**: `Update(sample, sampleHigh, sampleLow float64) float64`
**TS signature**: `update(sample: number, sampleHigh: number, sampleLow: number): number`

The `updateEntity` helper produces a 2-element output array: `[Scalar(frama), Scalar(fdim)]`.
If frama is NaN (not primed), fdim is also set to NaN.

**Entity mapping for high/low**:
- `UpdateBar`: `high = sample.High`, `low = sample.Low`
- `UpdateQuote`: `high = sample.Ask` (Go: `sample.Ask`; TS: `sample.askPrice`), `low = sample.Bid` (Go: `sample.Bid`; TS: `sample.bidPrice`)
- `UpdateScalar`/`UpdateTrade`: `high = low = sample value` (no separate high/low available)

### Reference Files

- Go: `go/indicators/johnehlers/fractaladaptivemovingaverage/`
- TS: `ts/indicators/john-ehlers/fractal-adaptive-moving-average/`

---

## Advanced: Helper / Shared Component Families

A **helper family** is a group of interchangeable internal components that all
implement the same interface (e.g., multiple cycle-estimator algorithms used
by MESA-style indicators). They are not indicators themselves and do not
extend `LineIndicator` — they are building blocks used *inside* indicators.

Canonical reference: **Hilbert Transformer cycle estimator**
(`go/indicators/johnehlers/hilberttransformer/`,
`ts/indicators/john-ehlers/hilbert-transformer/`).

### Folder Contents

A helper family lives in a single folder at the standard indicator depth and
contains one file per component role:

| Role                    | Go filename                          | TS filename                            | Python filename                        | Zig filename                           | Rust filename                          |
|-------------------------|--------------------------------------|----------------------------------------|----------------------------------------|----------------------------------------|----------------------------------------|
| Shared interface        | `cycleestimator.go`                  | `cycle-estimator.ts`                   | `cycle_estimator.py`                   | `cycle_estimator.zig`                  | `cycle_estimator.rs`                   |
| Shared params           | `cycleestimatorparams.go`            | `cycle-estimator-params.ts`            | `cycle_estimator_params.py`            | *(in same file)*                        | *(in same file)*                        |
| Variant type enum       | `cycleestimatortype.go`              | `cycle-estimator-type.ts`              | `cycle_estimator_type.py`              | *(in same file)*                        | *(in same file)*                        |
| Common + dispatcher     | `estimator.go`                       | `common.ts`                            | `common.py`                            | `estimator.zig`                         | `estimator.rs`                         |
| One file per variant    | `<variant>estimator.go`              | `<variant>.ts`                         | `<variant>_estimator.py`               | `<variant>_estimator.zig`              | `<variant>_estimator.rs`               |
| Spec per variant + common | `<variant>estimator_test.go`       | `<variant>.spec.ts`                    | `test_<variant>_estimator.py`          | *(tests in same file)*                  | *(tests in same file)*                  |

**No type suffixes.** Do not use `.interface.ts` / `.enum.ts` — plain `.ts`
files differentiated by stem name, same as every other indicator.

### Shared Interface

Expose:

- Primed flag and `warmUpPeriod` (read-only).
- Construction parameters as read-only getters.
- Intermediate state useful for tests (e.g., `smoothed`, `detrended`,
  `inPhase`, `quadrature`, `period`).
- A single `update(sample)` that returns the primary output.

TypeScript: getters take the plain name (`period`, not `periodValue`).
Backing private fields are prefixed with `_` (`_period`) to avoid the name
clash with the public getter.

### Dispatcher / Factory

The common file exposes a single entry point that constructs a variant from a
type enum value and optional params. Each variant case supplies its own
**default params** when the caller omits them:

```go
// Go
func NewCycleEstimator(typ CycleEstimatorType, params *CycleEstimatorParams) (CycleEstimator, error) {
    switch typ {
    case HomodyneDiscriminator: return NewHomodyneDiscriminatorEstimator(params)
    // ... other variants
    }
    return nil, fmt.Errorf("invalid cycle estimator type: %s", typ)
}
```

```ts
// TS
export function createEstimator(
    estimatorType?: HilbertTransformerCycleEstimatorType,
    estimatorParams?: HilbertTransformerCycleEstimatorParams,
): HilbertTransformerCycleEstimator {
    estimatorType ??= HilbertTransformerCycleEstimatorType.HomodyneDiscriminator;
    switch (estimatorType) {
        case HilbertTransformerCycleEstimatorType.HomodyneDiscriminator:
            estimatorParams ??= { smoothingLength: 4, alphaEmaQuadratureInPhase: 0.2, alphaEmaPeriod: 0.2 };
            return new HilbertTransformerHomodyneDiscriminator(estimatorParams);
        // ... other variants with their own defaults
        default: throw new Error('Invalid cycle estimator type: ' + estimatorType);
    }
}
```

Rules:

- Dispatcher picks a default variant when the type argument is omitted.
- Each `case` has its own default params literal — defaults can differ per
  variant and must match between Go and TS.
- Unknown types: Go returns an error, TS throws.

### Warm-Up Period

Every variant accepts an optional `warmUpPeriod` that, if larger than the
variant's intrinsic minimum priming length, overrides it. Tests must cover:

1. **Default priming** — `primed === false` for the first `intrinsicMinimum`
   samples, then `true`.
2. **Custom warm-up** — construct with `warmUpPeriod = N` (larger than
   intrinsic); `primed === false` for the first `N` samples, then `true`.

```ts
it('should respect custom warmUpPeriod', () => {
    const lprimed = 50;
    const est = new HilbertTransformerHomodyneDiscriminator(
        { smoothingLength: 4, alphaEmaQuadratureInPhase: 0.2, alphaEmaPeriod: 0.2, warmUpPeriod: lprimed }
    );
    expect(est.primed).toBe(false);
    for (let i = 0; i < lprimed; i++) { est.update(input[i]); expect(est.primed).toBe(false); }
    for (let i = lprimed; i < input.length; i++) { est.update(input[i]); expect(est.primed).toBe(true); }
});
```

### Cross-Language Alignment Checklist

- Interface getter names match across all five languages (modulo casing).
- Dispatcher default variant matches.
- Per-variant default params match exactly (same numeric values).
- `Primed()` / `isPrimed()` / `is_primed()` semantics match (all return
  `isWarmedUp`, not some other internal flag).
- Error/throw/panic messages use the same textual prefix where practical.

### Shared Formatter / Moniker Helpers

When a family defines a formatter (e.g. a per-variant mnemonic builder like
`EstimatorMoniker(typ, params) -> "hd(4, 0.200, 0.200)"`), it **must be
exported from all languages** that implement the indicator family. Indicators
that consume the family (MAMA, etc.) import and call this helper to build
their own mnemonic suffix — duplicating the formatting logic per-consumer
is a bug waiting to happen.

Checklist when converting a consumer indicator:

1. Does the reference language's helper package export a formatter used in the consumer's mnemonic?
2. Does the target language export an equivalent function with the same name
   (adjusted for casing) and identical format string?
3. If the target equivalent is missing, **add it to the helper's common file** first,
   then import it from the consumer. Do not inline the format in the consumer.

Reference: `estimatorMoniker` in
`ts/indicators/john-ehlers/hilbert-transformer/hilbert-transformer-common.ts`
mirrors Go's `hilberttransformer.EstimatorMoniker`.

### Numerical Tolerance for Recursive / Cascaded Indicators

The default test tolerance of `1e-12` works for simple moving averages,
single-stage EMAs, etc. Indicators that stack multiple recursive EMAs,
Hilbert-transform feedback loops, or trigonometric feedback (MAMA,
Hilbert-based cycle estimators, multi-stage T3) accumulate enough
floating-point drift across languages that `1e-12` is too tight.

**Use `1e-10`** as the tolerance for:

- MAMA / FAMA and anything consuming a Hilbert Transformer
- Any indicator whose `update` feeds its own previous output through an
  `atan`, `cos`, or multi-stage EMA chain before producing the output

Keep `1e-12` / `1e-13` for purely additive/multiplicative indicators.

---

## Registering in the Factory

After converting an indicator and registering its identifier and descriptor,
add it to the **factory** so it can be created from a JSON identifier string
and parameters at runtime. The factory lives at `indicators/factory/` in all
languages (see the `indicator-architecture` skill for full details).

Each language's implementation guide above includes factory registration steps:
- Go: [Step 7](#go-step-7-register-the-identifier-and-descriptor)
- TypeScript: built into the conversion steps
- Python: [Step 6](#step-6-register-in-the-factory)
- Zig: [Step 7](#step-7-register-in-factory)
- Rust: [Step 6](#step-6-register-in-factory)

Below are the Go and TypeScript factory patterns (as the reference implementations):

### Go

Add a `case` to the switch in `go/indicators/factory/factory.go` inside
`func New(identifier core.Identifier, paramsJSON []byte)`:

```go
case core.MyIndicator:
    var p myindicator.Params
    if err := unmarshal(paramsJSON, &p); err != nil {
        return nil, err
    }
    return myindicator.NewMyIndicator(&p)
```

For **multi-constructor** indicators (Length vs SmoothingFactor), use
`hasKey()` to detect which constructor to call:

```go
case core.MyIndicator:
    if hasKey(paramsJSON, "smoothingFactor") {
        var p myindicator.SmoothingFactorParams
        if err := unmarshal(paramsJSON, &p); err != nil {
            return nil, err
        }
        return myindicator.NewMyIndicatorSmoothingFactor(&p)
    }
    var p myindicator.LengthParams
    if err := unmarshal(paramsJSON, &p); err != nil {
        return nil, err
    }
    return myindicator.NewMyIndicatorLength(&p)
```

For **Default vs Params** indicators, use `isEmptyObject()`:

```go
case core.MyIndicator:
    if isEmptyObject(paramsJSON) {
        return myindicator.NewDefaultMyIndicator()
    }
    var p myindicator.Params
    if err := unmarshal(paramsJSON, &p); err != nil {
        return nil, err
    }
    return myindicator.NewMyIndicator(&p)
```

### TypeScript

Add a `case` to the switch in `ts/indicators/factory/factory.ts` inside
`createIndicator()`:

```ts
case IndicatorIdentifier.MyIndicator:
    return new MyIndicator(p as MyIndicatorParams);
```

For multi-constructor indicators, check for the discriminating key:

```ts
case IndicatorIdentifier.MyIndicator:
    if ('smoothingFactor' in p)
        return MyIndicator.fromSmoothingFactor(p as MyIndicatorSmoothingFactorParams);
    return MyIndicator.fromLength({ length: 14, ...p } as MyIndicatorLengthParams);
```

For Default vs Params indicators:

```ts
case IndicatorIdentifier.MyIndicator:
    if (Object.keys(p).length === 0) return MyIndicator.default();
    return MyIndicator.fromParams(p as MyIndicatorParams);
```

### Settings File

Add an entry to `cmd/icalc/settings.json` (shared across all languages) so the
new indicator is exercised by the CLI tool:

```json
{ "identifier": "myIndicator", "params": { "length": 14 } }
```

The `identifier` string uses **camelCase** (e.g., `jurikAdaptiveZeroLagVelocity`).

### Identifier String Mapping (Python)

Python's icalc has a manual string→enum mapping dict in `py/cmd/icalc/main.py`.
You MUST add a new entry mapping the camelCase JSON identifier to the
`Identifier` enum value:

```python
# In the IDENTIFIER_MAP dict in py/cmd/icalc/main.py:
'myIndicator': Identifier.MY_INDICATOR,
```

Without this entry, icalc will fail with `error: unknown indicator identifier`.

### Verification

Run `icalc` in **all implemented languages** to confirm the new indicator loads
and processes bars without error:

```bash
# Go
cd go && go run ./cmd/icalc settings.json

# TypeScript
cd ts && npx tsx cmd/icalc/main.ts cmd/icalc/settings.json

# Python
cd /home/dev/repos/chi/zpano && PYTHONPATH=. python3 -m py.cmd.icalc py/cmd/icalc/settings.json

# Zig
cd zig && zig build icalc -- src/cmd/icalc/settings.json

# Rust
cd rs && cargo run --bin icalc -- src/cmd/icalc/settings.json
```

If any language reports an unknown identifier or crashes, fix the registration
before considering the indicator complete.

---

## Python Implementation Guide

This section describes how to implement an indicator in Python. The Python
indicators module is **complete** (63 indicators, factory, frequency_response,
icalc/ifres/iconf — 803 tests passing). Use this guide as reference for future
indicators.

### Step 1: Create the folder structure

Create `py/indicators/<group>/<indicator_name>/` with:
- `__init__.py` — re-exports class, output enum, params, default_params
- `params.py` — `@dataclass` with `default_params()` function
- `output.py` — `IntEnum` class starting at 0
- `<indicator_name>.py` — main implementation
- `test_<indicator_name>.py` — unit tests

Also ensure `py/indicators/<group>/__init__.py` exists (empty).

### Step 2: Convert the params file

Map Go/TS params to a Python `@dataclass`. Key mappings:

| Go / TS | Python |
|---------|--------|
| `Length int` / `length: number` | `length: int = 20` |
| `BarComponent entities.BarComponent` (zero = not set) | `bar_component: Optional[BarComponent] = None` |
| `QuoteComponent` / `TradeComponent` | Same pattern with `Optional` + `None` |
| `SmoothingFactor float64` | `smoothing_factor: float = 0.0952` |
| `FirstIsAverage bool` | `first_is_average: bool = False` |
| Custom enum fields (e.g., `MovingAverageType`) | Python `IntEnum` in params.py |

Always include `default_params()` returning a default instance.

For multi-constructor indicators (EMA), create separate dataclass per variant:
`ExponentialMovingAverageLengthParams` and `ExponentialMovingAverageSmoothingFactorParams`.

### Step 3: Convert the output enum

```python
from enum import IntEnum

class SimpleMovingAverageOutput(IntEnum):
    VALUE = 0
```

Use `UPPER_SNAKE_CASE` members starting at 0 (matching TS, unlike Go's `iota`).

### Step 4: Convert the main indicator file

#### Single-output (line) indicators

1. Class extends `Indicator` (ABC), uses `LineIndicator` via composition:

```python
class SimpleMovingAverage(Indicator):
    def __init__(self, params: SimpleMovingAverageParams) -> None:
        # Resolve components: None → default
        bc = params.bar_component if params.bar_component is not None else DEFAULT_BAR_COMPONENT
        qc = params.quote_component if params.quote_component is not None else DEFAULT_QUOTE_COMPONENT
        tc = params.trade_component if params.trade_component is not None else DEFAULT_TRADE_COMPONENT

        bar_func = bar_component_value(bc)
        quote_func = quote_component_value(qc)
        trade_func = trade_component_value(tc)

        mnemonic = f"sma({length}{component_triple_mnemonic(bc, qc, tc)})"
        description = f"Simple moving average {mnemonic}"

        self._line = LineIndicator(mnemonic, description, bar_func, quote_func, trade_func, self.update)
        # ... indicator-specific state ...
```

2. Implement `update(self, sample: float) -> float` with the core logic.
3. Implement `is_primed(self) -> bool`.
4. Implement `metadata(self) -> Metadata` using `build_metadata()`.
5. Delegate `update_scalar/bar/quote/trade` to `self._line.*`.

#### Multi-output indicators (MACD, Bollinger Bands, etc.)

Do NOT use `LineIndicator`. Instead:
- Store component functions directly: `self._bar_func`, `self._quote_func`, `self._trade_func`
- Implement `update(self, sample: float)` returning a tuple of values
- Implement `update_scalar()` building a list of `Scalar`/`Band`/`Heatmap` objects
- Delegate `update_bar/quote/trade` through component extraction → `update_scalar`

#### Heatmap-output indicators

```python
from ...core.outputs.heatmap import Heatmap

# In update:
heatmap = Heatmap(time, param_first, param_last, param_resolution, value_min, value_max, values)
# or empty:
heatmap = Heatmap.empty(time, param_first, param_last, param_resolution)
```

#### Multi-constructor indicators (EMA pattern)

Use `@staticmethod` factory methods:

```python
class ExponentialMovingAverage(Indicator):
    @staticmethod
    def from_length(params: ExponentialMovingAverageLengthParams) -> 'ExponentialMovingAverage':
        ...

    @staticmethod
    def from_smoothing_factor(params: ExponentialMovingAverageSmoothingFactorParams) -> 'ExponentialMovingAverage':
        ...
```

#### Ehlers indicators with `create()` pattern

```python
class SuperSmoother(Indicator):
    @staticmethod
    def create(params: SuperSmootherParams) -> 'SuperSmoother':
        # Pre-compute coefficients from params
        ...
        return SuperSmoother(...)

    @staticmethod
    def create_default() -> 'SuperSmoother':
        return SuperSmoother.create(default_params())
```

### Step 5: Convert the test file

Key differences from Go/TS tests:
- Use `unittest.TestCase` (not Jasmine or `testing.T`)
- Use `assertAlmostEqual(result, expected, delta=1e-N)` — always use `delta=`, NOT `places=`
- Use `math.isnan()` for NaN checks, `self.assertTrue(math.isnan(result))`
- Test imports are absolute: `from py.indicators.common.simple_moving_average.simple_moving_average import SimpleMovingAverage`
- Test data arrays at module level: `INPUT = [...]`, `EXPECTED_3 = [...]`
- Call `is_primed()` with parentheses: `self.assertTrue(sma.is_primed())`

Tolerance mapping from Go to Python:
```
Go: epsilon = 1e-10  →  Python: delta=1e-10
Go: epsilon = 1e-12  →  Python: delta=1e-12
Go: math.Abs(exp-act) > 1e-N  →  Python: assertAlmostEqual(act, exp, delta=1e-N)
```

### Step 6: Register in the factory

Add to `py/indicators/factory/factory.py`:

```python
if identifier == Identifier.SIMPLE_MOVING_AVERAGE:
    from ..common.simple_moving_average.params import default_params
    from ..common.simple_moving_average.simple_moving_average import SimpleMovingAverage
    return SimpleMovingAverage(_apply(default_params(), params))
```

The factory uses lazy imports (inside the `if` block) to avoid circular dependencies.

Four construction patterns:
1. **Direct**: `Indicator(_apply(default_params(), params))` — most indicators
2. **`create()`**: `Indicator.create(_apply(default_params(), params))` — Ehlers
3. **Dual variant**: detect `_has_key(params, 'smoothing_factor')` — EMA family
4. **Raw constructor**: `WilliamsPercentR(params['length'])` — special cases

### Step 7: Register identifier and descriptor

If adding a new indicator (not just porting):
- Add to `py/indicators/core/identifier.py` — new `Identifier` IntEnum member
- Add to `py/indicators/core/descriptors.py` — new `_descriptors` entry

### Step 8: Verify

```bash
# Run all indicator tests:
cd /home/dev/repos/chi/zpano && PYTHONPATH=. python3 -m unittest discover -s py/indicators -p "test_*.py" -t .

# Run a single indicator's tests:
cd /home/dev/repos/chi/zpano && PYTHONPATH=. python3 -m unittest py.indicators.common.simple_moving_average.test_simple_moving_average

# Run icalc to verify factory + icalc mapping integration:
cd /home/dev/repos/chi/zpano && PYTHONPATH=. python3 -m py.cmd.icalc py/cmd/icalc/settings.json
```

**Important:** Running icalc is mandatory — it validates both the factory registration
AND the identifier string mapping in `py/cmd/icalc/main.py`. A passing unit test does
not guarantee icalc will work (the mapping dict is separate from the factory).

### Python Import Path Reference

From an indicator at `py/indicators/<group>/<indicator>/`:

```python
# Core framework
from ...core.indicator import Indicator
from ...core.line_indicator import LineIndicator
from ...core.metadata import Metadata
from ...core.build_metadata import build_metadata, OutputText
from ...core.identifier import Identifier
from ...core.component_triple_mnemonic import component_triple_mnemonic
from ...core.output import Output

# Entities (one more level up)
from ....entities.bar import Bar
from ....entities.quote import Quote
from ....entities.trade import Trade
from ....entities.scalar import Scalar
from ....entities.bar_component import BarComponent, DEFAULT_BAR_COMPONENT, bar_component_value
from ....entities.quote_component import QuoteComponent, DEFAULT_QUOTE_COMPONENT, quote_component_value
from ....entities.trade_component import TradeComponent, DEFAULT_TRADE_COMPONENT, trade_component_value

# Output types (when needed)
from ...core.outputs.band import Band
from ...core.outputs.heatmap import Heatmap
from ...core.outputs.polyline import Polyline, Point

# Sibling indicators
from ...common.exponential_moving_average.exponential_moving_average import ExponentialMovingAverage
from ...common.exponential_moving_average.params import ExponentialMovingAverageLengthParams
```

### Python ↔ Go/TS Name Mapping

See the consolidated [Cross-Language Name Mapping](#zig--gotstpythonrust-name-mapping)
table in the Zig section below for the full 5-language comparison.

---

## Zig Implementation Guide

This section describes how to implement an indicator in Zig. The Zig
indicators module is **complete** (63 indicators, factory, frequency_response,
icalc/ifres/iconf — 1014 tests passing). Use this guide as reference for future
indicators.

### Step 1: Create the folder structure

Create `zig/src/indicators/<group>/<indicator_name>/` with a single file:
- `<indicator_name>.zig` — params struct, output enum, indicator struct, impl, and `test` blocks

Zig uses a single-file-per-indicator pattern. No separate params/output/test files.

Register the module in the barrel file `zig/src/indicators/indicators.zig`:
```zig
// In the pub const exports section:
pub const simple_moving_average = @import("common/simple_moving_average/simple_moving_average.zig");

// In the comptime test-inclusion block:
comptime { _ = simple_moving_average; }
```

Also register in `build.zig` if the indicator needs its own test target (usually
not needed — the barrel file's comptime block includes tests automatically).

### Step 2: Convert the params struct

Map Go/TS params to a Zig struct with defaults:

| Go / TS | Zig |
|---------|-----|
| `Length int` / `length: number` | `length: u32 = 20` |
| `BarComponent entities.BarComponent` (zero = not set) | `bar_component: ?BarComponent = null` |
| `QuoteComponent` / `TradeComponent` | Same `?T = null` pattern |
| `SmoothingFactor float64` | `smoothing_factor: f64 = 0.0952...` |
| `FirstIsAverage bool` | `first_is_average: bool = false` |
| Custom enum fields (e.g., `MovingAverageType`) | Zig `enum(u8)` in same file |

```zig
pub const SimpleMovingAverageParams = struct {
    length: u32 = 20,
    bar_component: ?BarComponent = null,
    quote_component: ?QuoteComponent = null,
    trade_component: ?TradeComponent = null,
};
```

Component sentinel: `?BarComponent` with `null` meaning "use default", resolved
via `params.bar_component orelse default_bar_component`.

For multi-constructor indicators (EMA), create separate structs:
`ExponentialMovingAverageLengthParams` and `ExponentialMovingAverageSmoothingFactorParams`.

### Step 3: Convert the output enum

```zig
pub const SimpleMovingAverageOutput = enum(u8) {
    value = 1,
};
```

Output enums are **1-based** (matching Go's `iota+1` and the descriptor registry).
This differs from Python/TS which use 0-based.

### Step 4: Convert the main indicator file

#### Single-output (line) indicators

```zig
pub const SimpleMovingAverage = struct {
    // LineIndicator for entity routing
    line: LineIndicator,

    // Owned string buffers (for mnemonic/description after heap move)
    mnemonic_buf: [64]u8 = undefined,
    mnemonic_len: usize = 0,
    description_buf: [128]u8 = undefined,
    description_len: usize = 0,

    // Indicator-specific state
    primed: bool = false,
    window: []f64,
    window_sum: f64 = 0,
    length: u32,
    allocator: std.mem.Allocator,

    const vtable = Indicator.GenVTable(SimpleMovingAverage);
};
```

Key patterns:

1. **Component resolution** — `null` → default:
   ```zig
   const bc = params.bar_component orelse default_bar_component;
   const bar_func = barComponentValue(bc);
   ```

2. **LineIndicator composition** — extract sample then call core update:
   ```zig
   pub fn updateBar(self: *SimpleMovingAverage, bar: Bar) OutputArray {
       const sample = self.line.extractBar(bar);
       const value = self.update(sample);
       return self.line.wrapScalar(bar.time, value);
   }
   ```

3. **NaN handling** — Use `std.math.nan(f64)` and `std.math.isNan()`:
   ```zig
   if (!self.primed) return std.math.nan(f64);
   ```

4. **Metadata construction** — out-pointer pattern:
   ```zig
   pub fn getMetadata(self: *SimpleMovingAverage, out: *Metadata) void {
       buildMetadata(out, .simple_moving_average,
           self.line.mnemonic, self.line.description,
           &.{ .{ .mnemonic = "val", .description = "Value" } });
   }
   ```

5. **vtable exposure**:
   ```zig
   const vtable = Indicator.GenVTable(SimpleMovingAverage);
   pub fn indicator(self: *SimpleMovingAverage) Indicator {
       return .{ .ptr = self, .vtable = &vtable };
   }
   ```

6. **fixSlices()** — critical for heap-allocated indicators:
   ```zig
   pub fn fixSlices(self: *SimpleMovingAverage) void {
       self.line.mnemonic = self.mnemonic_buf[0..self.mnemonic_len];
       self.line.description = self.description_buf[0..self.description_len];
   }
   ```

#### Multi-output indicators

Do NOT use `LineIndicator`. Store component functions directly and build
`OutputArray` with multiple entries:

```zig
pub fn updateBar(self: *BollingerBands, bar: Bar) OutputArray {
    const sample = self.bar_func(bar);
    const result = self.update(sample);
    var out = OutputArray{};
    out.append(.{ .scalar = Scalar.init(bar.time, result.lower) });
    out.append(.{ .scalar = Scalar.init(bar.time, result.middle) });
    out.append(.{ .scalar = Scalar.init(bar.time, result.upper) });
    out.append(.{ .band = Band.init(bar.time, result.upper, result.lower) });
    out.append(.{ .scalar = Scalar.init(bar.time, result.bandwidth) });
    out.append(.{ .scalar = Scalar.init(bar.time, result.percent_band) });
    return out;
}
```

#### Ehlers indicators with precomputed coefficients

Use `init()` with coefficient computation at construction time:

```zig
pub fn init(params: SuperSmootherParams) !SuperSmoother {
    if (params.length < 2) return error.InvalidLength;
    const angle = 2.0 * std.math.pi / @as(f64, @floatFromInt(params.length));
    const a1 = @exp(-std.math.sqrt(2.0) * std.math.pi / @as(f64, @floatFromInt(params.length)));
    // ... compute coefficients
}
```

### Step 5: Constructor pattern — `init()` and `deinit()`

Two categories:

1. **No allocation needed** (most Ehlers, simple filters):
   ```zig
   pub fn init(params: SuperSmootherParams) !SuperSmoother { ... }
   // No deinit needed
   ```

2. **Allocation needed** (SMA, BB, anything with ring buffers):
   ```zig
   pub fn init(allocator: std.mem.Allocator, params: SimpleMovingAverageParams) !SimpleMovingAverage {
       const window = try allocator.alloc(f64, length);
       // ...
   }
   pub fn deinit(self: *SimpleMovingAverage) void {
       self.allocator.free(self.window);
   }
   ```

### Step 6: Convert the tests

```zig
test "simple moving average length 20" {
    const allocator = std.testing.allocator;
    var sma = try SimpleMovingAverage.init(allocator, .{ .length = 20 });
    defer sma.deinit();

    for (INPUT, 0..) |input, i| {
        const result = sma.update(input);
        if (std.math.isNan(EXPECTED_20[i])) {
            try std.testing.expect(std.math.isNan(result));
        } else {
            try std.testing.expect(@abs(result - EXPECTED_20[i]) < 1e-8);
        }
    }
    try std.testing.expect(sma.isPrimed());
}
```

Key test conventions:
- `test "descriptive name" { ... }` blocks at bottom of the indicator file
- `std.testing.allocator` for leak detection (with `defer indicator.deinit()`)
- `@abs(result - expected) < tolerance` for float comparisons
- `std.math.isNan()` for NaN checks
- Test data as module-level `const` arrays

### Step 7: Register in factory

Edit `zig/src/indicators/factory/factory.zig`:

1. Add import at top:
   ```zig
   const sma_mod = @import("../common/simple_moving_average/simple_moving_average.zig");
   ```

2. Add match arm in `create()` function. Two patterns based on allocator need:
   ```zig
   // Pattern A — needs allocator:
   .simple_moving_average => return createWithAllocParams(
       sma_mod.SimpleMovingAverage, sma_mod.SimpleMovingAverageParams,
       allocator, params_json),

   // Pattern B — no allocator:
   .super_smoother => return createWithParams(
       ss_mod.SuperSmoother, ss_mod.SuperSmootherParams,
       allocator, params_json),
   ```

### Step 8: Verify

```bash
# Build all (includes indicators):
cd zig && zig build

# Run all tests (includes indicator tests via barrel comptime block):
cd zig && zig build test --summary all

# Run icalc to verify factory integration:
cd zig && zig build icalc -- src/cmd/icalc/settings.json
```

### Zig Import Path Reference

From an indicator at `zig/src/indicators/<group>/<indicator>/`:

```zig
// Core framework (via barrel)
const indicators = @import("indicators");
const Indicator = indicators.Indicator;
const OutputArray = indicators.OutputArray;
const OutputValue = indicators.OutputValue;
const LineIndicator = indicators.LineIndicator;
const Metadata = indicators.Metadata;
const Identifier = indicators.Identifier;
const buildMetadata = indicators.buildMetadata;
const componentTripleMnemonic = indicators.componentTripleMnemonic;

// Entities (via barrel)
const entities = @import("entities");
const Bar = entities.Bar;
const BarComponent = entities.BarComponent;
const default_bar_component = entities.default_bar_component;
const barComponentValue = entities.barComponentValue;
const Quote = entities.Quote;
const QuoteComponent = entities.QuoteComponent;
const default_quote_component = entities.default_quote_component;
const quoteComponentValue = entities.quoteComponentValue;
const Trade = entities.Trade;
const TradeComponent = entities.TradeComponent;
const default_trade_component = entities.default_trade_component;
const tradeComponentValue = entities.tradeComponentValue;
const Scalar = entities.Scalar;

// Output types (when needed)
const Band = indicators.Band;
const Heatmap = indicators.Heatmap;
const Polyline = indicators.Polyline;

// Sibling indicators
const ema_mod = @import("../common/exponential_moving_average/exponential_moving_average.zig");
```

### Zig ↔ Go/TS/Python/Rust Name Mapping

This is the consolidated **cross-language name mapping** table for all five languages:

| Go | TypeScript | Python | Zig | Rust |
|----|-----------|--------|-----|------|
| `Update(sample float64) float64` | `update(sample: number): number` | `update(self, sample: float) -> float` | `fn update(self: *Self, sample: f64) f64` | `fn update(&mut self, sample: f64) -> f64` |
| `IsPrimed() bool` | `isPrimed(): boolean` | `is_primed(self) -> bool` | `fn isPrimed(self: *Self) bool` | `fn is_primed(&self) -> bool` |
| `Metadata() Metadata` | `metadata(): IndicatorMetadata` | `metadata(self) -> Metadata` | `fn getMetadata(self: *Self, out: *Metadata) void` | `fn metadata(&self) -> Metadata` |
| `UpdateBar(*Bar) Output` | `updateBar(bar: Bar): Output` | `update_bar(self, sample: Bar) -> Output` | `fn updateBar(self: *Self, bar: Bar) OutputArray` | `fn update_bar(&mut self, bar: &Bar) -> Output` |
| `core.LineIndicator` (embedded) | `LineIndicator` (extended) | `LineIndicator` (composed via `self._line`) | `LineIndicator` (composed via `self.line`) | Fields inlined (composition) |
| `core.BuildMetadata(...)` | `buildMetadata(...)` | `build_metadata(...)` | `buildMetadata(out, ...)` (out-pointer) | `build_metadata(...)` |
| `core.ComponentTripleMnemonic(...)` | `componentTripleMnemonic(...)` | `component_triple_mnemonic(...)` | `componentTripleMnemonic(...)` | `component_triple_mnemonic(...)` |
| `core.OutputText{...}` | `{ mnemonic, description }` | `OutputText(mnemonic, description)` | `.{ .mnemonic = "...", .description = "..." }` | `OutputText { mnemonic: "...", description: "..." }` |
| `entities.DefaultBarComponent` | `DefaultBarComponent` | `DEFAULT_BAR_COMPONENT` | `default_bar_component` | `DEFAULT_BAR_COMPONENT` |
| `entities.BarFunc` | `(bar: Bar) => number` | `Callable[[Bar], float]` | `BarFunc` (`*const fn(Bar) f64`) | `fn(&Bar) -> f64` |
| `math.NaN()` | `NaN` | `math.nan` | `std.math.nan(f64)` | `f64::NAN` |
| `math.IsNaN(x)` | `isNaN(x)` | `math.isnan(x)` | `std.math.isNan(x)` | `x.is_nan()` |
| `Params` struct | `params` interface | `@dataclass` class | `Params` struct with defaults | `struct` + `impl Default` |
| `DefaultParams()` | `defaultParams()` | `default_params()` | `.{}` (struct literal with defaults) | `Default::default()` |

### Zig-Specific Pitfalls

1. **fixSlices()** — Any indicator storing mnemonic/description slices pointing into
   owned `[N]u8` buffers MUST implement `fixSlices()`. The factory calls it after
   copying the indicator from stack to heap. Without it, slices become dangling pointers.

2. **Owned string buffers** — Use `[64]u8` + `usize` length for mnemonic, `[128]u8`
   for description. Build via `std.fmt.bufPrint`. Do NOT allocate strings on the heap.

3. **OutputArray is stack-based** — Max 9 outputs, no heap allocation. Use
   `OutputArray.fromScalar()` for single-output, or manual `append()` for multi-output.

4. **Metadata out-pointer** — `getMetadata` takes `*Metadata` instead of returning it
   (avoids large struct copy). This differs from all other languages.

5. **Allocator propagation** — Indicators needing heap allocation store the allocator
   as a field and pass it to `ArrayList`/`alloc` operations. Use `std.testing.allocator`
   in tests for automatic leak detection.

6. **Build.zig module wiring** — Indicators import the `entities` barrel module
   (`@import("entities")`), not individual entity modules. The barrel file
   re-exports everything; new indicators must be added to it.

---

## Rust Implementation Guide

This section describes how to implement an indicator in Rust. The Rust
indicators module is **complete** (67 indicators, factory, frequency_response,
icalc/ifres/iconf — 985 tests passing). Use this guide as reference for future
indicators.

### Step 1: Create the folder structure

Create `rs/src/indicators/<group>/<indicator_name>/` with:
- `mod.rs` — re-exports: `mod <indicator_name>; pub use <indicator_name>::*;`
- `<indicator_name>.rs` — params, output enum, indicator struct, impl, and `#[cfg(test)] mod tests`

Register the module in the group's `mod.rs` (e.g., `rs/src/indicators/common/mod.rs`):
```rust
pub mod simple_moving_average;
```

### Step 2: Convert the params struct

Map Go/TS params to a Rust struct with `impl Default`:

| Go / TS | Rust |
|---------|------|
| `Length int` / `length: number` | `pub length: usize` (default `20`) |
| `BarComponent entities.BarComponent` (zero = not set) | `pub bar_component: Option<BarComponent>` (default `None`) |
| `QuoteComponent` / `TradeComponent` | Same `Option` pattern |
| `SmoothingFactor float64` | `pub smoothing_factor: f64` (default `0.0952...`) |
| `FirstIsAverage bool` | `pub first_is_average: bool` (default `false`) |
| Custom enum fields (e.g., `MovingAverageType`) | Rust `enum` with `#[repr(u8)]` in same file |

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

For multi-constructor indicators (EMA), create separate structs:
`ExponentialMovingAverageLengthParams` and `ExponentialMovingAverageSmoothingFactorParams`.

MESA params do **not** implement `Default` — factory constructs them explicitly.

### Step 3: Convert the output enum

```rust
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
#[repr(u8)]
pub enum SimpleMovingAverageOutput {
    Value = 1,
}
```

Output enums are **1-based** (matching Go's `iota+1` and the descriptor registry).
This differs from Python/TS which use 0-based.

### Step 4: Convert the main indicator file

#### Single-output (line) indicators

```rust
pub struct SimpleMovingAverage {
    // LineIndicator fields inlined (Rust uses composition)
    bar_func: fn(&Bar) -> f64,
    quote_func: fn(&Quote) -> f64,
    trade_func: fn(&Trade) -> f64,
    meta: Metadata,

    // Indicator-specific state
    primed: bool,
    window: Vec<f64>,
    index: usize,
    sum: f64,
    length: usize,
}
```

Key patterns:

1. **Component resolution** — `None` → default:
   ```rust
   let bc = params.bar_component.unwrap_or(DEFAULT_BAR_COMPONENT);
   let bar_func = bar_component_value(bc);
   ```

2. **Borrow checker constraint** — Cannot pass `|v| self.update(v)` closure to
   `self.line` because both borrow `self`. Extract the sample first:
   ```rust
   fn update_bar(&mut self, bar: &Bar) -> Output {
       let sample = (self.bar_func)(bar);
       let value = self.update(sample);
       vec![Box::new(Scalar::new(bar.time, value))]
   }
   ```

3. **NaN handling** — Use `f64::NAN` and `f64::is_nan()`:
   ```rust
   if !self.primed { return f64::NAN; }
   ```

4. **Metadata construction**:
   ```rust
   let meta = build_metadata(
       Identifier::SimpleMovingAverage,
       &format!("sma({length}{ctm})"),
       &format!("Simple moving average sma({length}{ctm})"),
       &[OutputText { mnemonic: "val", description: "Value" }],
   );
   ```

#### Multi-output indicators

Return `Vec<Box<dyn Any>>` with multiple `Scalar`/`Band`/`Heatmap` outputs:

```rust
fn update_bar(&mut self, bar: &Bar) -> Output {
    let sample = (self.bar_func)(bar);
    let (signal, histogram) = self.update_internal(sample);
    vec![
        Box::new(Scalar::new(bar.time, signal)),
        Box::new(Scalar::new(bar.time, histogram)),
    ]
}
```

#### Ehlers indicators with `create()` pattern

Ehlers indicators use an associated function for construction:

```rust
impl MesaAdaptiveMovingAverage {
    pub fn create(params: MesaAdaptiveMovingAverageLengthParams) -> Self { ... }
    pub fn from_smoothing_factor(params: MesaAdaptiveMovingAverageSmoothingFactorParams) -> Self { ... }
}
```

### Step 5: Convert the tests

```rust
#[cfg(test)]
mod tests {
    use super::*;
    use crate::entities::bar_component::BarComponent;

    const INPUT: [f64; 50] = [ /* ... */ ];
    const EXPECTED_20: [f64; 50] = [ /* ... */ ];

    #[test]
    fn test_simple_moving_average_20() {
        let params = SimpleMovingAverageParams { length: 20, ..Default::default() };
        let mut sma = SimpleMovingAverage::new(params);
        for i in 0..INPUT.len() {
            let result = sma.update(INPUT[i]);
            if EXPECTED_20[i].is_nan() {
                assert!(result.is_nan(), "bar {i}: expected NaN, got {result}");
            } else {
                assert!(
                    (result - EXPECTED_20[i]).abs() < 1e-8,
                    "bar {i}: expected {}, got {result}", EXPECTED_20[i]
                );
            }
        }
        assert!(sma.is_primed());
    }
}
```

Key test conventions:
- `#[cfg(test)] mod tests` at bottom of the indicator `.rs` file
- `use super::*;` to import everything from the parent module
- Test data as `const` arrays (not `static`)
- Tolerance typically `1e-8` (matching Go test tolerances)
- NaN checks with `.is_nan()` + descriptive assertion messages with bar index

### Step 6: Register in factory

Edit `rs/src/indicators/factory/factory.rs`:

1. Add import at top (match visibility pattern A or B):
   ```rust
   // Pattern A (pub use * re-export):
   use crate::indicators::common::simple_moving_average::{SimpleMovingAverage, SimpleMovingAverageParams};
   // Pattern B (pub mod):
   use crate::indicators::custom::goertzel_spectrum::goertzel_spectrum::{GoertzelSpectrum, GoertzelSpectrumParams};
   ```

2. Add match arm in `create_indicator()`:
   ```rust
   Identifier::SimpleMovingAverage => {
       let mut p = SimpleMovingAverageParams::default();
       if !is_empty_object(json) {
           if let Some(v) = get_usize(json, "length") { p.length = v; }
           if let Some(v) = get_i32(json, "bar_component") { p.bar_component = Some(BarComponent::from_i32(v)); }
           // ... other fields
       }
       Box::new(SimpleMovingAverage::new(p))
   }
   ```

### Step 7: Register in identifier enum

Add to `rs/src/indicators/core/identifier.rs` and `rs/src/indicators/core/descriptors.rs`.

### Step 8: Verify

```bash
# Run all indicator tests:
cd rs && cargo test --lib indicators

# Run a single indicator's tests:
cd rs && cargo test --lib simple_moving_average

# Run factory tests:
cd rs && cargo test --lib factory

# Run icalc to verify factory integration:
cd rs && cargo run --bin icalc -- path/to/settings.json
```

### Rust Import Path Reference

From an indicator at `rs/src/indicators/<group>/<indicator>/`:

```rust
// Core framework
use crate::indicators::core::indicator::{Indicator, Output};
use crate::indicators::core::line_indicator::LineIndicator;  // if used
use crate::indicators::core::metadata::Metadata;
use crate::indicators::core::build_metadata::{build_metadata, OutputText};
use crate::indicators::core::identifier::Identifier;
use crate::indicators::core::component_triple_mnemonic::component_triple_mnemonic;

// Entities
use crate::entities::bar::Bar;
use crate::entities::quote::Quote;
use crate::entities::trade::Trade;
use crate::entities::scalar::Scalar;
use crate::entities::bar_component::{component_value as bar_component_value, BarComponent, DEFAULT_BAR_COMPONENT};
use crate::entities::quote_component::{component_value as quote_component_value, QuoteComponent, DEFAULT_QUOTE_COMPONENT};
use crate::entities::trade_component::{component_value as trade_component_value, TradeComponent, DEFAULT_TRADE_COMPONENT};

// Output types (when needed)
use crate::indicators::core::outputs::band::Band;
use crate::indicators::core::outputs::heatmap::Heatmap;
use crate::indicators::core::outputs::polyline::{Polyline, Point};

// Sibling indicators
use crate::indicators::common::exponential_moving_average::{ExponentialMovingAverage, ExponentialMovingAverageLengthParams};
```

### Rust Name Mapping

See the consolidated [Cross-Language Name Mapping](#zig--gotstpythonrust-name-mapping)
table in the Zig section above for the full 5-language comparison.
