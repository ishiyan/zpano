---
name: mbst-indicator-conversion
description: Step-by-step guide for converting MBST C# indicators to zpano Go and TypeScript. Load when converting an MBST indicator or understanding the mapping between MBST and zpano patterns.
---

# Converting MBST C# Indicators to Zpano

This guide provides recipes and tips for converting indicators from the MBST C# codebase
(`Mbst.Trading.Indicators`) to the zpano multi-language library (Go and TypeScript).

Load the `indicator-architecture` and `indicator-conversion` skills alongside this one
for the full zpano architecture reference and internal conversion patterns.

Load the `mbst-indicator-architecture` skill for the full MBST type hierarchy reference.

---

## Table of Contents

1. [Overview](#overview)
2. [Determine the Indicator Pattern](#determine-the-indicator-pattern)
3. [Component Mapping](#component-mapping)
4. [Constructor Conversion](#constructor-conversion)
5. [Algorithm Conversion](#algorithm-conversion)
6. [Multi-Output Indicators](#multi-output-indicators)
7. [Metadata Conversion](#metadata-conversion)
8. [What to Drop](#what-to-drop)
9. [Test Conversion](#test-conversion)
10. [Worked Example: CenterOfGravityOscillator](#worked-example-centerofgravityoscillator)

---

## Overview

MBST indicators live in `mbst-to-convert/<author>/<indicator>/` as C# classes.
Each indicator has a `.cs` implementation and a `Test.cs` test file.

The conversion produces:
- **Go:** `go/indicators/<author>/<indicator>/` (5 files: params, output, output_test, impl, impl_test)
- **TS:** `ts/indicators/<author>/<indicator>/` (4 files: params, output, impl, impl.spec)

Always convert Go first, then TypeScript.

---

## Determine the Indicator Pattern

Read the MBST source and identify which base class/interface the indicator uses:

| MBST Pattern | Zpano Pattern |
|---|---|
| `class X : LineIndicator` | Single output. Use `LineIndicator` embedding (Go) / `extends LineIndicator` (TS). Only implement `Update(float64)` / `update(number)`. |
| `class X : Indicator, ILineIndicator` | Multi-output (e.g., value + trigger). Implement `Indicator` interface directly (Go) / `implements Indicator` (TS). Must write `UpdateScalar/Bar/Quote/Trade` manually. |
| `class X : BandIndicator` | Band output. Map to zpano's band indicator pattern. |
| `class X : Indicator, IBandIndicator` | Custom band output. Implement `Indicator` interface directly. |
| `class X : ... IHeatmapIndicator` | Heatmap output. Handle case-by-case. |

### How to tell if it's multi-output

Look for:
1. **Facade properties** (`ValueFacade`, `TriggerFacade`, etc.) — the indicator exposes
   multiple outputs via facades.
2. **Multiple named value fields** (e.g., `value` + `valuePrevious` both exposed publicly).
3. **The class extends `Indicator` directly** (not `LineIndicator`) but implements
   `ILineIndicator` — this is a strong signal of multi-output.

**Exception:** Some MBST indicators use `: Indicator, ILineIndicator` but only have a single
output (`Value`). Example: `SuperSmoother` has this signature but no facades or additional
outputs. In this case, use `LineIndicator` embedding/inheritance in zpano (not the direct
`Indicator` interface). Always check the actual outputs — the class hierarchy alone is not
sufficient to determine multi-output vs single-output.

If multi-output: use the direct `Indicator` interface approach (like FRAMA, CoG).
If single-output: use `LineIndicator` embedding/inheritance.

---

## Component Mapping

### OhlcvComponent to Component Triple

MBST uses a single `OhlcvComponent` for OHLCV bars only. Zpano uses three separate
component types: `BarComponent`, `QuoteComponent`, `TradeComponent`.

| MBST `OhlcvComponent` | Zpano `BarComponent` | Go Constant | TS Enum |
|---|---|---|---|
| `ClosingPrice` (default) | `Close` (default) | `entities.BarClosePrice` | `BarComponent.Close` |
| `OpeningPrice` | `Open` | `entities.BarOpenPrice` | `BarComponent.Open` |
| `HighPrice` | `High` | `entities.BarHighPrice` | `BarComponent.High` |
| `LowPrice` | `Low` | `entities.BarLowPrice` | `BarComponent.Low` |
| `MedianPrice` | `Median` | `entities.BarMedianPrice` | `BarComponent.Median` |
| `TypicalPrice` | `Typical` | `entities.BarTypicalPrice` | `BarComponent.Typical` |
| `WeightedPrice` | `Weighted` | `entities.BarWeightedPrice` | `BarComponent.Weighted` |
| `Volume` | `Volume` | `entities.BarVolume` | `BarComponent.Volume` |

### Default Component Handling

- **MBST default:** `OhlcvComponent.ClosingPrice` — set in `Indicator` base class constructor.
- **Zpano default:** `BarClosePrice` / `BarComponent.Close` — aka `DefaultBarComponent`.
- **Some indicators override the default** (e.g., CoG defaults to `MedianPrice` instead of
  `ClosingPrice`). Check the MBST constructor signature for non-default values.

When converting, QuoteComponent and TradeComponent are **always new** (MBST doesn't have
them). Use the zpano defaults (`DefaultQuoteComponent`, `DefaultTradeComponent`) when the
params zero-value/undefined is provided.

### Non-Default Bar Component Rule

If the MBST indicator uses a non-default `OhlcvComponent` (anything other than
`ClosingPrice`), the zpano constructor must resolve the zero-value to that component
instead of `DefaultBarComponent`. This causes `ComponentTripleMnemonic` to include the
component in the mnemonic even for default parameters.

**Go example (CoG defaults to MedianPrice):**
```go
bc := params.BarComponent
if bc == 0 {
    bc = entities.BarMedianPrice  // NOT entities.DefaultBarComponent
}
```

**TS example:**
```typescript
const bc = params.barComponent ?? BarComponent.Median;  // NOT DefaultBarComponent
```

---

## Constructor Conversion

### MBST Constructor Pattern

```csharp
public CenterOfGravityOscillator(int length = 10,
    OhlcvComponent ohlcvComponent = OhlcvComponent.MedianPrice)
    : base(cog, cogFull, ohlcvComponent)
{
    if (1 > length) throw new ArgumentOutOfRangeException(argumentLength);
    this.length = length;
    // ...
    moniker = string.Concat(cog, "(", length.ToString(...), ")");
}
```

### Zpano Conversion Rules

1. **Parameters become a struct/interface** — not individual constructor args.
2. **Validation stays** — translate `ArgumentOutOfRangeException` to `error` (Go) / `throw Error` (TS).
3. **`moniker` becomes `mnemonic`** — built using `ComponentTripleMnemonic` for the triple.
4. **`description` is auto-generated** — typically `"<Full Name> " + mnemonic`.
5. **`base(name, description, component)` call is dropped** — zpano doesn't have a base
   class constructor chain for this.
6. **Component function resolution** — create `barFunc`/`quoteFunc`/`tradeFunc` in constructor.

---

## Algorithm Conversion

### The Only Method That Matters

In MBST, `Update(double sample)` contains the entire algorithm. All other `Update`
overloads (`Update(Scalar)`, `Update(Ohlcv)`) are pure delegation boilerplate.

**Only convert `Update(double sample)`.** The entity-level update methods are generated
by `LineIndicator` embedding (for single-output) or written manually using a simple
`updateEntity` helper (for multi-output).

### C# to Go/TS Algorithm Translation

| C# | Go | TS |
|---|---|---|
| `double.NaN` | `math.NaN()` | `Number.NaN` |
| `double.IsNaN(x)` | `math.IsNaN(x)` | `Number.isNaN(x)` |
| `double.Epsilon` | `math.SmallestNonzeroFloat64` | `Number.MIN_VALUE` |
| `Math.Abs(x)` | `math.Abs(x)` | `Math.abs(x)` |
| `Array.Copy(src, srcIdx, dst, dstIdx, len)` | Manual loop or `copy()` | Manual loop |
| `lock (updateLock) { ... }` | `s.mu.Lock(); defer s.mu.Unlock()` | Drop (single-threaded) |

### Priming Logic

MBST `primed` field is set inside `Update()` after enough samples.
Convert directly — zpano uses the same `primed` boolean pattern.

### NaN Guard

MBST typically has `if (double.IsNaN(sample)) return sample;` at the top of `Update`.
Preserve this — it's important for correctness.

---

## Multi-Output Indicators

### MBST: Facade Pattern

MBST exposes multiple outputs via facade classes:

```csharp
// In MBST indicator:
public LineIndicatorFacade ValueFacade =>
    new LineIndicatorFacade(cog, moniker, cogFull, () => IsPrimed, () => Value);
public LineIndicatorFacade TriggerFacade =>
    new LineIndicatorFacade(cogTrig, ...moniker..., cogTrigFull, () => IsPrimed, () => Trigger);
```

### Zpano: Output Array

Zpano replaces facades with an `Output` array (Go `core.Output` / TS `IndicatorOutput`):

**Go:**
```go
func (s *X) updateEntity(time time.Time, sample float64) core.Output {
    output := make([]any, 2)
    cog := s.Update(sample)
    trig := s.valuePrevious
    if math.IsNaN(cog) { trig = math.NaN() }
    output[0] = entities.Scalar{Time: time, Value: cog}
    output[1] = entities.Scalar{Time: time, Value: trig}
    return output
}
```

**TS:**
```typescript
private updateEntity(time: Date, sample: number): IndicatorOutput {
    const cog = this.update(sample);
    let trig = this.valuePrevious;
    if (Number.isNaN(cog)) { trig = Number.NaN; }
    const s1 = new Scalar(); s1.time = time; s1.value = cog;
    const s2 = new Scalar(); s2.time = time; s2.value = trig;
    return [s1, s2];
}
```

### Output Enum

Each multi-output indicator gets a per-indicator output enum:

**Go:** `centerofgravityoscillatoroutput.go`
```go
type Output int
const ( Value Output = iota; Trigger )
```

**TS:** `center-of-gravity-oscillator-output.ts`
```typescript
export enum CenterOfGravityOscillatorOutput { Value = 0, Trigger = 1 }
```

### Metadata for Multi-Output

Each output gets its own entry in `Metadata().Outputs`:

```go
Outputs: []outputs.Metadata{
    { Kind: int(Value), Type: outputs.ScalarType, Mnemonic: s.mnemonic, Description: s.description },
    { Kind: int(Trigger), Type: outputs.ScalarType, Mnemonic: s.mnemonicTrig, Description: s.descriptionTrig },
},
```

Facade-specific mnemonic patterns (like `cogTrig(10)`) are preserved as the Trigger
output mnemonic.

---

## Metadata Conversion

| MBST | Zpano |
|---|---|
| `Name` (e.g., `"cog"`) | `Metadata().Type` — use the registered `core.IndicatorType` enum |
| `Moniker` (e.g., `"cog(10)"`) | `Metadata().Mnemonic` — includes component triple |
| `Description` (e.g., `"Center of Gravity oscillator"`) | `Metadata().Description` — typically `"Full Name " + mnemonic` |

### Register the Indicator Type

Before implementing, register the indicator in:
- **Go:** `go/indicators/core/type.go` — add enum constant, string, `String()`, `MarshalJSON`, `UnmarshalJSON`
- **TS:** `ts/indicators/core/indicator-type.ts` — add enum member

---

## What to Drop

### From the Implementation

| MBST Element | Action |
|---|---|
| `[DataContract]`, `[DataMember]` | Drop — no serialization |
| `Reset()` method | Drop — zpano indicators are immutable |
| `ToString()` override | Drop — debug formatting |
| `lock (updateLock)` | Go: replace with `sync.RWMutex`. TS: drop entirely |
| `Update(Scalar)`, `Update(Ohlcv)` | Drop — handled by `LineIndicator` or `updateEntity` |
| `Update(double, DateTime)` | Drop — convenience overload |
| Facade properties (`ValueFacade`, `TriggerFacade`) | Drop — replaced by Output array |
| `OhlcvComponent` property setter | Drop — component is immutable after construction |
| C# regions (`#region ... #endregion`) | Drop — organizational noise |
| XML doc comments (`/// <summary>`) | Convert to Go doc comments / JSDoc |

### From Tests

| MBST Test | Action |
|---|---|
| `ToStringTest()` | Drop |
| `SerializeTo()` / `SerializeFrom()` | Drop |
| `SerializationTest()` | Drop |
| Facade-specific tests | Drop (test via Output array instead) |
| `Reset()` tests | Drop |
| `[TestMethod]` / `[TestClass]` | Convert to Go `Test*` / TS Jasmine `describe`/`it` |
| `Assert.IsTrue(x == y)` | Go: require assertions. TS: Jasmine `expect(x).toBe(y)` |

---

## Test Conversion

### Reference Data

MBST tests typically include large arrays of reference data (from TA-Lib Excel
simulations). **Preserve this data exactly** — it's the ground truth for numerical
verification.

Typical MBST test structure:
```csharp
private readonly List<double> rawInput = new List<double> { ... };  // 252 entries
private readonly List<double> expected = new List<double> { ... };  // expected output
```

Convert to:
- **Go:** `var` block with `[]float64` slices
- **TS:** `const` arrays of `number`

### Test Categories to Convert

1. **Output value tests** — Feed `rawInput`, compare against `expected` values.
   Use tolerance comparison: Go `require.InDelta(t, expected, actual, 1e-10)`,
   TS `expect(Math.abs(actual - expected)).toBeLessThan(1e-10)`.

2. **IsPrimed tests** — Verify `IsPrimed()` is false during warmup and true after.

3. **Metadata tests** — Verify `Metadata()` returns correct type, mnemonic, description, outputs.

4. **Constructor validation** — Verify invalid params (e.g., `length < 1`) produce errors.

5. **UpdateEntity tests** — Test `UpdateScalar`, `UpdateBar`, `UpdateQuote`, `UpdateTrade`
   with a few samples to verify entity routing works.

6. **NaN handling** — Verify that NaN input produces NaN output without corrupting state.

### Test for Multi-Output

For multi-output indicators, test **both** (all) outputs against reference data:
```go
// After feeding sample:
output := indicator.UpdateScalar(scalar)
cogScalar := output[int(Value)].(entities.Scalar)
trigScalar := output[int(Trigger)].(entities.Scalar)
require.InDelta(t, expectedCog, cogScalar.Value, tolerance)
require.InDelta(t, expectedTrig, trigScalar.Value, tolerance)
```

### MBST Test Data with Separate High/Low Arrays

Some MBST tests provide raw high and low arrays separately, then compute median price
in the test. In zpano tests, either:
- Pre-compute the median values and store as a single input array, or
- Feed `Bar` entities with high/low set and test via `UpdateBar`.

The CoG tests use a pre-computed `rawInput` array of median prices (already `(high + low) / 2`).

---

## Worked Example: CenterOfGravityOscillator

### Source Analysis

```csharp
// MBST: extends Indicator directly, implements ILineIndicator
// Has Value + Trigger facades = multi-output
// Default OhlcvComponent: MedianPrice (non-default!)
public sealed class CenterOfGravityOscillator : Indicator, ILineIndicator
```

### Decision: Multi-output, direct Indicator interface

Because CoG has two outputs (Value, Trigger) and extends `Indicator` directly (not
`LineIndicator`), it uses the direct `Indicator` interface pattern in zpano.

### Key Conversion Points

1. **Default component**: `OhlcvComponent.MedianPrice` -> `BarMedianPrice` / `BarComponent.Median`.
   Not the framework default, so it always appears in the mnemonic (`, hl/2`).

2. **Facades dropped**: `ValueFacade` and `TriggerFacade` replaced by Output array
   `[Scalar{cog}, Scalar{trigger}]`.

3. **Output enum created**: `Value = 0`, `Trigger = 1`.

4. **Algorithm**: `Update(double)` converted line-by-line. `Calculate()` helper preserved.
   `Array.Copy` replaced with explicit loop.

5. **Priming**: Requires `length + 1` samples. First `length` fill the window; at index
   `length - 1` the initial CoG is computed and stored as `valuePrevious`; at index
   `length` the indicator becomes primed.

6. **Mnemonic**: `cog(10, hl/2)` for default params. The `, hl/2` comes from
   `BarMedianPrice` being non-default relative to `DefaultBarComponent` (which is
   `BarClosePrice`).

7. **Tests**: 252-entry reference data preserved. Serialization/ToString/Reset tests dropped.
   Added entity update tests, metadata tests, NaN tests.

### Files Produced

**Go (5 files):**
- `centerofgravityoscillatorparams.go` — `Params` struct with `Length`, `BarComponent`, `QuoteComponent`, `TradeComponent`
- `centerofgravityoscillatoroutput.go` — `Output` enum: `Value`, `Trigger`
- `centerofgravityoscillatoroutput_test.go` — Output enum string tests
- `centerofgravityoscillator.go` — Main implementation (273 lines)
- `centerofgravityoscillator_test.go` — Tests with 252-entry data (659 lines)

**TS (4 files):**
- `center-of-gravity-oscillator-params.ts` — `CenterOfGravityOscillatorParams` interface
- `center-of-gravity-oscillator-output.ts` — `CenterOfGravityOscillatorOutput` enum
- `center-of-gravity-oscillator.ts` — Main implementation (227 lines)
- `center-of-gravity-oscillator.spec.ts` — Tests with 252-entry data

---

## Bar-Based Indicators (Non-LineIndicator Pattern)

Some indicators require bar data (high, low, close) rather than a single scalar component.
These indicators implement the `Indicator` interface directly without using `LineIndicator`
embedding/inheritance. Example: **TrueRange**.

### Key Differences from LineIndicator

| Aspect | LineIndicator | Bar-Based (e.g., TrueRange) |
|---|---|---|
| Base type | Embeds `LineIndicator` (Go) / `extends LineIndicator` (TS) | Implements `Indicator` directly |
| Core method | `Update(float64)` / `update(number)` — single scalar | `Update(close, high, low float64)` — multiple values |
| Bar handling | Extracts one component via `barComponentFunc` | Extracts H, L, C directly from bar |
| Scalar/Quote/Trade | Routed through component function | Use single value as H=L=C substitute |
| Params | Typically has `Length`, component fields | May be empty (parameterless) |
| Components | `BarComponent`, `QuoteComponent`, `TradeComponent` | None — bar fields accessed directly |

### Go Pattern

```go
type TrueRange struct {
    mu            sync.RWMutex
    previousClose float64
    value         float64
    primed        bool
}

func (tr *TrueRange) Update(close, high, low float64) float64 { ... }
func (tr *TrueRange) UpdateSample(sample float64) float64 {
    return tr.Update(sample, sample, sample)
}
func (tr *TrueRange) UpdateBar(sample *entities.Bar) core.Output {
    output := make([]any, 1)
    output[0] = entities.Scalar{Time: sample.Time, Value: tr.Update(sample.Close, sample.High, sample.Low)}
    return output
}
func (tr *TrueRange) UpdateScalar(sample *entities.Scalar) core.Output {
    v := sample.Value
    output := make([]any, 1)
    output[0] = entities.Scalar{Time: sample.Time, Value: tr.Update(v, v, v)}
    return output
}
```

### TS Pattern

```typescript
export class TrueRange implements Indicator {
    public update(close: number, high: number, low: number): number { ... }
    public updateSample(sample: number): number {
        return this.update(sample, sample, sample);
    }
    public updateBar(sample: Bar): IndicatorOutput {
        const scalar = new Scalar();
        scalar.time = sample.time;
        scalar.value = this.update(sample.close, sample.high, sample.low);
        return [scalar];
    }
}
```

### Test Data Extraction

For large test datasets (e.g., 252-entry TA-Lib arrays), extract data **programmatically**
from the C# test file using a Python script rather than manual transcription:

```python
import re
pattern = rf'readonly List<double> {name} = new List<double>\s*\{{(.*?)\}};'
# Remove C-style comments before extracting numbers
body = re.sub(r'/\*.*?\*/', '', body)
```

This avoids transcription errors that can be very hard to debug.

---

## Composite Indicators (Indicator-inside-Indicator Pattern)

Some indicators internally create and use other indicator instances. Example: **AverageTrueRange**
creates an internal **TrueRange** instance and delegates bar processing to it.

### Key Points

- The inner indicator is a private field, created in the constructor
- The outer indicator calls the inner indicator's `Update()` method and processes the result
- Both Go and TS follow the same pattern: import the inner indicator's package and instantiate it
- The outer indicator manages its own priming state independently of the inner indicator
- In Go, the inner indicator's mutex is separate — no nested locking issues because `Update()`
  on the inner indicator acquires/releases its own lock before the outer lock is held

### Go Example (AverageTrueRange)

```go
import "zpano/indicators/welleswilder/truerange"

type AverageTrueRange struct {
    mu        sync.RWMutex
    trueRange *truerange.TrueRange  // internal indicator instance
    // ... other fields
}

func NewAverageTrueRange(length int) (*AverageTrueRange, error) {
    return &AverageTrueRange{
        trueRange: truerange.NewTrueRange(),
        // ...
    }, nil
}

func (a *AverageTrueRange) Update(close, high, low float64) float64 {
    trueRangeValue := a.trueRange.Update(close, high, low)  // delegate to inner
    // ... apply Wilder smoothing to trueRangeValue
}
```

### TS Example (AverageTrueRange)

```typescript
import { TrueRange } from '../true-range/true-range';

export class AverageTrueRange implements Indicator {
    private readonly trueRange: TrueRange;

    constructor(length: number) {
        this.trueRange = new TrueRange();
    }

    public update(close: number, high: number, low: number): number {
        const trueRangeValue = this.trueRange.update(close, high, low);
        // ... apply Wilder smoothing to trueRangeValue
    }
}
```

**Important:** Call the inner indicator's `Update()` **before** acquiring the outer lock in Go,
or call it inside the lock if the inner indicator uses its own separate mutex (as TrueRange does).
In the ATR implementation, `trueRange.Update()` is called inside the outer lock — this works
because TrueRange's mutex is independent.

---

## Volume-Aware Indicators (UpdateWithVolume Pattern)

Some indicators require both a price sample and volume. Since `LineIndicator` only
supports `Update(float64)` / `update(number)`, these indicators need special handling.

**Example: MoneyFlowIndex (MFI)**

### Pattern

1. **Embed `LineIndicator`** as usual (single scalar output).
2. **Add `UpdateWithVolume(sample, volume float64)`** as the real computation method.
3. **`Update(sample)`** delegates to `UpdateWithVolume(sample, 1)` (volume=1 fallback).
4. **Shadow/override `UpdateBar`** to extract both price (via `barFunc`) AND volume from
   the bar, then call `UpdateWithVolume`.

### Go: Shadow `UpdateBar`

```go
// MoneyFlowIndex shadows LineIndicator.UpdateBar to extract volume.
func (s *MoneyFlowIndex) UpdateBar(sample *entities.Bar) core.Output {
    s.mu.Lock()
    defer s.mu.Unlock()
    price := s.barFunc(sample)
    v := s.updateWithVolume(price, sample.Volume)
    output := make([]any, 1)
    output[0] = entities.Scalar{Time: sample.Time, Value: v}
    return output
}
```

The shadowed `UpdateBar` on the concrete type takes precedence over `LineIndicator.UpdateBar`
when called on `*MoneyFlowIndex` directly or through the `Indicator` interface.

### TS: Override `updateBar`

```typescript
public override updateBar(sample: Bar): IndicatorOutput {
    const price = this.barFunc(sample);
    const v = this.updateWithVolume(price, sample.volume);
    const scalar = new Scalar();
    scalar.time = sample.time;
    scalar.value = v;
    return [scalar];
}
```

### Default BarComponent

MFI defaults to `BarTypicalPrice` / `BarComponent.Typical` (not `BarClosePrice`),
matching the C# default of `OhlcvComponent.TypicalPrice`.

### Mnemonic

MFI uses `mfi(LENGTH)` — no component suffix, matching C# behavior.

### Testing Volume-Aware Indicators

Test with two datasets:
1. **Real volume** — `updateWithVolume(price, volume)` against expected values.
2. **Volume=1** — `update(price)` (which uses volume=1) against a separate expected dataset.
3. **UpdateBar test** — Feed `Bar` entities with real OHLCV data, verify first computed
   value matches the real-volume expected data.

---

## Test Data from Julia/CSV Reference (No C# Tests)

Some MBST indicators lack C# unit tests but have Julia reference implementations and CSV
test data files. Example: `SuperSmoother` has `ehlers_super_smoother.jl` and
`test_3-3_Supersmoother.csv`.

### Key Differences: MBST vs Julia Priming

- **MBST priming:** Seeds previous filter values to the first sample value on count==1.
- **Julia priming:** Seeds previous filter values to zero (array initialized to zeros).
- After sufficient samples (~30-60), both converge to the same values.

### CSV Test Data Strategy

1. **Go tests:** Read the CSV file at test time using `encoding/csv`. Reference the CSV
   via a relative path (e.g., `../../../../mbst-to-convert/.../file.csv`).
2. **TS tests:** Embed a representative subset (e.g., first 200 rows) directly in the
   spec file as `const` arrays (CSV reading at test time is more complex in TS).
3. **Skip early rows:** Due to priming differences, skip the first N rows (e.g., 60)
   where MBST and Julia outputs diverge.
4. **Tolerance:** Use a generous tolerance (e.g., 2.5) to account for:
   - CSV data rounded to 2 decimal places
   - Different priming initialization (MBST seeds to first sample, Julia to zero)
   - Convergence lag in the early samples
5. **Julia test reference:** Check the Julia test file (`*_test.jl`) for the skip count
   and rounding precision used in the original Julia validation.
