---
name: indicator-conversion
description: Step-by-step guide for converting old-structure indicators to the new architecture in both Go and TypeScript. Load when migrating an existing indicator.
---

# Converting an Old-Structure Indicator to New Structure

This guide provides step-by-step instructions for converting an indicator from the old
architecture to the new architecture. It covers both **Go** and **TypeScript** implementations.

The new architecture introduces:

- **Per-indicator packages** (Go) / **per-indicator folders** (TS) instead of a flat shared package
- **`LineIndicator` embedding/inheritance** that eliminates `UpdateScalar/Bar/Quote/Trade` boilerplate
- **`ComponentTripleMnemonic`** (bar + quote + trade) replacing `componentPairMnemonic` (bar + quote only)
- **Zero-value default resolution** for components: zero/undefined = use default, don't show in mnemonic
- **Top-level `Mnemonic` and `Description`** on metadata
- **Per-indicator output enum** (TS) in a dedicated file
- **Renamed imports** from `data`/`indicator`/`output` to `entities`/`core`/`outputs`

Use the SMA indicator as the canonical reference for single-constructor indicators.
Use the EMA indicator as the canonical reference for multi-constructor indicators
(see [Advanced: Multi-Constructor Indicators](#advanced-multi-constructor-indicators)).

---

## Table of Contents

1. [Go Conversion](#go-conversion)
   1. [Create the new package directory](#go-step-1-create-the-new-package-directory)
   2. [Convert the params file](#go-step-2-convert-the-params-file)
   3. [Convert the output file](#go-step-3-convert-the-output-file)
   4. [Convert the output test file](#go-step-4-convert-the-output-test-file)
   5. [Convert the main indicator file](#go-step-5-convert-the-main-indicator-file)
   6. [Convert the test file](#go-step-6-convert-the-test-file)
   7. [Register the indicator type](#go-step-7-register-the-indicator-type)
   8. [Verify](#go-step-8-verify)
2. [TypeScript Conversion](#typescript-conversion)
   1. [Create the new folder](#ts-step-1-create-the-new-folder)
   2. [Convert the params file](#ts-step-2-convert-the-params-file)
   3. [Create the output file](#ts-step-3-create-the-output-file)
   4. [Convert the main indicator file](#ts-step-4-convert-the-main-indicator-file)
   5. [Convert the test file](#ts-step-5-convert-the-test-file)
   6. [Verify](#ts-step-6-verify)
3. [Quick Reference: Import Mapping](#quick-reference-import-mapping)
4. [Quick Reference: Symbol Renames](#quick-reference-symbol-renames)
5. [Advanced: Multi-Constructor Indicators](#advanced-multi-constructor-indicators)
6. [Advanced: Helper / Shared Component Families](#advanced-helper--shared-component-families)
7. [Naming & Style Conventions](#naming--style-conventions)
8. [Registering in the Factory](#registering-in-the-factory)

---

## Naming & Style Conventions

All identifier, receiver, concurrency, style, and cross-language parity
rules are defined in the **`indicator-architecture`** skill and MUST be
followed during conversion. Summary (see that skill for the full tables
and rationale):

- **Abbreviations banned in identifiers** — always expand: `idx→index`,
  `tmp→temp`, `res→result`, `sig→signal`, `val→value`, `prev→previous`,
  `avg→average`, `mult→multiplier`, `buf→buffer`, `param→parameter`,
  `hist→histogram`. Allowed: Go idioms (`err`, `len`, `cap`, `min`,
  `max`, `num`), the Go `Params`/`params` bundle type, TS `value_`
  ctor-param idiom.
- **Go receivers** — compound type name (2+ CamelCase words) → `s`;
  simple type name (single word) → first-letter of type, lowercased.
  All methods on a type MUST use the same receiver. If a local would
  shadow `s` (e.g. in `MarshalJSON`), rename the local to `str`.
- **Concurrency** — stateful public indicators MUST carry `mu
  sync.RWMutex`; writers `s.mu.Lock(); defer s.mu.Unlock()`, readers
  `s.mu.RLock(); defer s.mu.RUnlock()`, defer on the line immediately
  after Lock. Exceptions: pure delegating wrappers, internal engines
  guarded by their public wrapper.
- **Go style invariants** — no `var x T = zero`, no split
  `var x T; x = expr`; use `any` not `interface{}`; no
  `make([]T, 0)` (always include capacity); no `new`; grouped imports
  (stdlib, external, `zpano/*`); every exported symbol has a doc
  comment, even trivial passthroughs.
- **Go ↔ TS local-variable parity** — same concept = same name in
  both languages. Canonical: `sum`, `epsilon`, `temp`/`diff`,
  `stddev`, `spread`, `bw`, `pctB`, `amount`, `lengthMinOne`;
  loop counter `i`/`j`/`k`; `index` only when semantically a named
  index. Never introduce new short forms.

When porting, copy the other language's local-variable names verbatim
where the concept is identical.

---

## Go Conversion

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

### Go Step 8: Verify

```bash
cd go && go test ./indicators/common/<indicatorname>/...
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

Each constructor path gets its own param struct (Go) or interface (TS).

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

The mnemonics must match between Go and TS. This may require changing the TS format
from the old style to match Go (e.g., old TS alpha path used `ema(0.123)` with 3 decimal
places; new uses `ema(10, 0.18181818)` with 8 decimal places to match Go).

### Cross-Language Behavior Alignment

When converting multi-constructor indicators, check for behavioral differences between
Go and TS and decide whether to align them:

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

| Role                    | Go filename                          | TS filename                            |
|-------------------------|--------------------------------------|----------------------------------------|
| Shared interface        | `cycleestimator.go`                  | `cycle-estimator.ts`                   |
| Shared params           | `cycleestimatorparams.go`            | `cycle-estimator-params.ts`            |
| Variant type enum       | `cycleestimatortype.go`              | `cycle-estimator-type.ts`              |
| Common + dispatcher     | `estimator.go`                       | `common.ts`                            |
| One file per variant    | `<variant>estimator.go`              | `<variant>.ts`                         |
| Spec per variant + common | `<variant>estimator_test.go`       | `<variant>.spec.ts`                    |

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

- Interface getter names match (modulo language casing).
- Dispatcher default variant matches.
- Per-variant default params match exactly (same numeric values).
- `Primed()` / `primed` semantics match (both return `isWarmedUp`, not some
  other internal flag).
- Error/throw messages use the same textual prefix where practical.

### Shared Formatter / Moniker Helpers

When a family defines a formatter (e.g. a per-variant mnemonic builder like
`EstimatorMoniker(typ, params) -> "hd(4, 0.200, 0.200)"`), it **must be
exported from both Go and TS** helper packages. Indicators that consume the
family (MAMA, etc.) import and call this helper to build their own mnemonic
suffix — duplicating the formatting logic per-consumer-indicator is a bug
waiting to happen.

Checklist when converting a consumer indicator:

1. Does Go's helper package export a formatter used in the consumer's mnemonic?
2. Does TS's helper export an equivalent function with the same name (camelCase)
   and identical format string (e.g. `%.3f` ↔ `.toFixed(3)`)?
3. If the TS equivalent is missing, **add it to the helper's common file** first,
   then import it from the consumer. Do not inline the format in the consumer.

Reference: `estimatorMoniker` in
`ts/indicators/john-ehlers/hilbert-transformer/hilbert-transformer-common.ts`
mirrors Go's `hilberttransformer.EstimatorMoniker`.

### Numerical Tolerance for Recursive / Cascaded Indicators

The default test tolerance of `1e-12` (Go) / `1e-12` (TS) works for simple
moving averages, single-stage EMAs, etc. Indicators that stack multiple
recursive EMAs, Hilbert-transform feedback loops, or trigonometric feedback
(MAMA, Hilbert-based cycle estimators, multi-stage T3) accumulate enough
floating-point drift between Go and TS that `1e-12` is too tight.

**Use `1e-10`** as the tolerance for:

- MAMA / FAMA and anything consuming a Hilbert Transformer
- Any indicator whose `update` feeds its own previous output through an
  `atan`, `cos`, or multi-stage EMA chain before producing the output

Keep `1e-12` / `1e-13` for purely additive/multiplicative indicators.

---

## Registering in the Factory

After converting an indicator and registering its identifier and descriptor,
add it to the **factory** so it can be created from a JSON identifier string
and parameters at runtime. The factory lives at `indicators/factory/` in both
languages (see the `indicator-architecture` skill for full details).

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

Add an entry to `cmd/icalc/settings.json` (shared between Go and TS) so the
new indicator is exercised by the CLI tool:

```json
{ "identifier": "myIndicator", "params": { "length": 14 } }
```

### Verification

Run `icalc` in both languages to confirm the new indicator loads and processes
bars without error:

```bash
cd go && go run ./cmd/icalc settings.json
cd ts && npx tsx cmd/icalc/main.ts cmd/icalc/settings.json
```
