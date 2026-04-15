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

**File:** `<indicatorname>params.go`

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

### Go Step 3: Convert the output file

**File:** `<indicatorname>output.go`

**Only change:** `package indicators` -> `package <indicatorname>`

The output enum type, constants, `String()`, `IsKnown()`, `MarshalJSON()`, and `UnmarshalJSON()`
methods are identical between old and new. No logic changes needed.

### Go Step 4: Convert the output test file

**File:** `<indicatorname>output_test.go`

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

Remove the `"mbg/trading/indicators/indicator/output"` import entirely -- it is only needed if
the old code references `output.Scalar` in `Metadata()`, which changes to `outputs.ScalarType`.

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

#### 5e. Metadata: update return type and fields

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

New:

```go
func (s *SimpleMovingAverage) Metadata() core.Metadata {
    return core.Metadata{
        Type:        core.SimpleMovingAverage,
        Mnemonic:    s.LineIndicator.Mnemonic,
        Description: s.LineIndicator.Description,
        Outputs: []outputs.Metadata{
            {
                Kind:        int(SimpleMovingAverageValue),
                Type:        outputs.ScalarType,
                Mnemonic:    s.LineIndicator.Mnemonic,
                Description: s.LineIndicator.Description,
            },
        },
    }
}
```

Key changes:

| Old | New |
|-----|-----|
| `indicator.Metadata` | `core.Metadata` |
| (no top-level Mnemonic/Description) | `Mnemonic: s.LineIndicator.Mnemonic` |
| (no top-level Mnemonic/Description) | `Description: s.LineIndicator.Description` |
| `[]output.Metadata` | `[]outputs.Metadata` |
| `output.Scalar` | `outputs.ScalarType` |
| `Name: s.name` | `Mnemonic: s.LineIndicator.Mnemonic` |
| `Description: s.description` | `Description: s.LineIndicator.Description` |

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
check("Type", core.SimpleMovingAverage, act.Type)
check("Outputs[0].Type", outputs.ScalarType, act.Outputs[0].Type)
check("Outputs[0].Mnemonic", "sma(5)", act.Outputs[0].Mnemonic)
```

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

### Go Step 7: Register the indicator type

If the indicator type constant does not already exist in `core/indicator.go` (or wherever
indicator type constants are defined), add it. Example:

```go
const (
    SimpleMovingAverage IndicatorType = iota + 1
    // ...
)
```

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
| `simple-moving-average-params.interface.ts` | `simple-moving-average-params.ts` |
| `simple-moving-average.ts` | `simple-moving-average.ts` |
| `simple-moving-average.spec.ts` | `simple-moving-average.spec.ts` |
| (none) | `simple-moving-average-output.ts` |

### TS Step 2: Convert the params file

**File:** `<indicator-name>-params.ts` (renamed from `<indicator-name>-params.interface.ts`)

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

### TS Step 3: Create the output file

**File:** `<indicator-name>-output.ts` (NEW file -- does not exist in old structure)

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
import { IndicatorMetadata } from '../../core/indicator-metadata';
import { IndicatorType } from '../../core/indicator-type';
import { OutputType } from '../../core/outputs/output-type';
import { SimpleMovingAverageOutput } from './simple-moving-average-output';
```

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

Add a new `metadata()` method to the class:

```typescript
/** Describes the output data of the indicator. */
metadata(): IndicatorMetadata {
    return {
        type: IndicatorType.SimpleMovingAverage,
        mnemonic: this.mnemonic,
        description: this.description,
        outputs: [
            {
                kind: SimpleMovingAverageOutput.Value,
                type: OutputType.Scalar,
                mnemonic: this.mnemonic,
                description: this.description,
            },
        ],
    };
}
```

Adapt `IndicatorType.*`, output enum, and output entries to your specific indicator.

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

### TS Step 6: Verify

Run the tests for the specific indicator:

```bash
# Adjust the path and test runner command to your project setup
ng test --include='**/indicators/common/<indicator-name>/**/*.spec.ts'
```

All tests must pass.

---

## Quick Reference: Import Mapping

### Go

| Old import path | New import path |
|-----------------|-----------------|
| `"mbg/trading/data"` | `"zpano/entities"` |
| `"mbg/trading/indicators/indicator"` | `"zpano/indicators/core"` |
| `"mbg/trading/indicators/indicator/output"` | `"zpano/indicators/core/outputs"` |

### TypeScript

| Old import path (relative from indicator) | New import path (relative from indicator) |
|------------------------------------------|------------------------------------------|
| `'../indicator/component-pair-mnemonic'` | `'../../core/component-triple-mnemonic'` |
| `'../indicator/line-indicator'` | `'../../core/line-indicator'` |
| `'./<name>-params.interface'` | `'./<name>-params'` |
| `'../../../data/entities/bar-component.enum'` | `'../../../entities/bar-component'` |
| `'../../../data/entities/quote-component.enum'` | `'../../../entities/quote-component'` |
| (none) | `'../../../entities/trade-component'` |
| (none) | `'../../core/indicator-metadata'` |
| (none) | `'../../core/indicator-type'` |
| (none) | `'../../core/outputs/output-type'` |
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
| `indicator.SimpleMovingAverage` | `core.SimpleMovingAverage` |
| `output.Metadata` | `outputs.Metadata` |
| `output.Scalar` | `outputs.ScalarType` |

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
| (no metadata method) | `metadata(): IndicatorMetadata { ... }` |
| (no output file) | `<name>-output.ts` with output enum |

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
