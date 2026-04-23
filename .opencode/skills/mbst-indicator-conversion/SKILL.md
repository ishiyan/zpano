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
11. [Naming & Style Conventions](#naming--style-conventions)

---

## Naming & Style Conventions

All identifier, receiver, concurrency, style, and cross-language parity
rules are defined in the **`indicator-architecture`** skill and MUST be
followed during conversion. Summary (see that skill for the full tables
and rationale):

- **Abbreviations banned in identifiers** — always expand: `idx→index`,
  `tmp→temp`, `res→result`, `sig→signal`, `val→value`, `prev→previous`,
  `avg→average`, `mult→multiplier`, `buf→buffer`, `param→parameter`,
  `hist→histogram`. MBST C# source frequently uses `avg`, `prev`, `tmp`,
  `sig`, `hist`, `mult`, `val` — **expand these** when porting; do not
  preserve the C# abbreviation. Keep the **Go struct type `Params`** and
  local `params` variable as-is.
- **Go receivers** — compound type name (2+ CamelCase words, as most
  MBST ports are) → `s`; single-word type → first-letter lowercased.
  All methods on a type use the same receiver. If a local would shadow
  `s`, rename the local to `str`.
- **Concurrency** — stateful public indicators MUST carry `mu
  sync.RWMutex`; writers `s.mu.Lock(); defer s.mu.Unlock()`, readers
  `s.mu.RLock(); defer s.mu.RUnlock()`. Exceptions: pure delegating
  wrappers, and internal engines consumed only by higher-level
  indicators that hold the lock (e.g. the `corona` engine behind the
  four corona* indicators).
- **Go style invariants** — no `var x T = zero`, no split var-then-
  assign; use `any`; no bare `make([]T, 0)`; grouped imports; every
  exported symbol has a doc comment.
- **Go ↔ TS local-variable parity** — same concept = same name in
  both languages. Canonical: `sum`, `epsilon`, `temp`/`diff`,
  `stddev`, `spread`, `bw`, `pctB`, `amount`, `lengthMinOne`;
  loop counter `i`/`j`/`k`.

When converting the C# algorithm, translate the operations but rename
the locals to the canonical zpano vocabulary — do not preserve MBST's
C# identifier names verbatim.

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

**Go:** `output.go` (in `centerofgravityoscillator/` package)
```go
type Output int
const ( Value Output = iota; Trigger )
```

**TS:** `output.ts` (in `center-of-gravity-oscillator/` folder)
```typescript
export enum CenterOfGravityOscillatorOutput { Value = 0, Trigger = 1 }
```

### Metadata for Multi-Output

Each output gets its own `core.OutputText` entry passed to `core.BuildMetadata`. The `Kind`
and `Shape` are sourced from the descriptor registry — the indicator file only supplies
per-output mnemonic + description text. The number and order of entries must match the
descriptor row registered in `go/indicators/core/descriptors.go`:

```go
func (s *CenterOfGravityOscillator) Metadata() core.Metadata {
    return core.BuildMetadata(
        core.CenterOfGravityOscillator,
        s.mnemonic,
        s.description,
        []core.OutputText{
            {Mnemonic: s.mnemonic,     Description: s.description},
            {Mnemonic: s.mnemonicTrig, Description: s.descriptionTrig},
        },
    )
}
```

TS mirrors this with `buildMetadata(IndicatorIdentifier.CenterOfGravityOscillator, ..., [{mnemonic, description}, {mnemonic, description}])`.

Facade-specific mnemonic patterns (like `cogTrig(10)`) are preserved as the Trigger
output mnemonic.

---

## Metadata Conversion

| MBST | Zpano |
|---|---|
| `Name` (e.g., `"cog"`) | First argument to `core.BuildMetadata` — a `core.Identifier` enum value |
| `Moniker` (e.g., `"cog(10)"`) | `Metadata().Mnemonic` — includes component triple |
| `Description` (e.g., `"Center of Gravity oscillator"`) | `Metadata().Description` — typically `"Full Name " + mnemonic` |

### Register the Identifier and Descriptor

Before implementing, register the indicator in:

- **Go identifier:** `go/indicators/core/identifier.go` — add enum constant, string, `String()`, `MarshalJSON`, `UnmarshalJSON`
- **Go descriptor:** `go/indicators/core/descriptors.go` — add a taxonomy row with per-output `Kind`/`Shape`
- **TS identifier:** `ts/indicators/core/indicator-identifier.ts` — add enum member
- **TS descriptor:** `ts/indicators/core/descriptors.ts` — add the matching row

`BuildMetadata` / `buildMetadata` panic / throw if the identifier has no descriptor row or
if the output-count mismatches. See `.opencode/skills/indicator-architecture/SKILL.md`
section "Taxonomy & Descriptor Registry" for field meanings and guidance.

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

**Go (5 files, in `centerofgravityoscillator/` package):**
- `params.go` — `Params` struct with `Length`, `BarComponent`, `QuoteComponent`, `TradeComponent`
- `output.go` — `Output` enum: `Value`, `Trigger`
- `output_test.go` — Output enum string tests
- `centerofgravityoscillator.go` — Main implementation (273 lines)
- `centerofgravityoscillator_test.go` — Tests with 252-entry data (659 lines)

**TS (4 files, in `center-of-gravity-oscillator/` folder):**
- `params.ts` — `CenterOfGravityOscillatorParams` interface
- `output.ts` — `CenterOfGravityOscillatorOutput` enum
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

### Composing a Multi-Output Inner Indicator

When the inner indicator is itself multi-output (implements `Indicator` directly and
returns a tuple), the outer indicator typically wants to **re-expose some of the inner
outputs alongside its own**. Example: **SineWave** composes **DominantCycle** and
publishes 5 outputs: `Value`, `Lead`, `Band` (its own) + `DominantCyclePeriod`,
`DominantCyclePhase` (forwarded from the inner).

Rules:

1. **Instantiate the inner using the `Params` constructor**, not `Default`, so the outer
   can forward parameters correctly (estimator params, `alphaEmaPeriodAdditional`,
   components, etc.).
2. **Resolve component defaults in the outer first**, then pass the **resolved** values
   explicitly to the inner. This keeps mnemonics aligned across outputs — e.g., if the
   user overrides `BarComponent = Median`, **both** `sw(0.330, hl/2)` and the inner's
   `dcp(0.330, hl/2)` carry the suffix. If you skip this, the inner resolves its own
   defaults and the two mnemonics drift apart.
3. **Call the inner's `Update()` before acquiring the outer lock** (inner manages its
   own mutex).
4. **Forward outputs by index** into the outer's output tuple — don't re-compute. The
   outer's `updateEntity` assembles all N `Scalar`/`Band` entities from the inner's
   period/phase plus its own Value/Lead/Band.
5. **Keep a single top-level `Mnemonic` / `Description`** on the outer's `Metadata`
   (the primary output's). Each forwarded inner output carries its own per-output
   mnemonic unchanged.

```go
// Go — inside newSineWave, after resolving bc/qc/tc to non-zero defaults:
dc, err := dominantcycle.NewDominantCycleParams(&dominantcycle.Params{
    AlphaEmaPeriodAdditional: p.AlphaEmaPeriodAdditional,
    EstimatorType:            p.EstimatorType,
    EstimatorParams:          p.EstimatorParams,
    BarComponent:             bc, // resolved, not p.BarComponent
    QuoteComponent:           qc,
    TradeComponent:           tc,
})
```

```ts
// TS — inside the private SineWave constructor:
this.dominantCycle = DominantCycle.fromParams({
    alphaEmaPeriodAdditional: params.alphaEmaPeriodAdditional,
    estimatorType: params.estimatorType,
    estimatorParams: params.estimatorParams,
    barComponent: bc, quoteComponent: qc, tradeComponent: tc,
});
```

### Composing a Raw HTCE vs Wrapping `DominantCycle`

When porting an MBST indicator that internally owns a `HilbertTransformerCycleEstimator`
(HTCE), you must decide whether to:

**(a) Compose the already-merged `DominantCycle`** — when the MBST indicator feeds the
HTCE's **smoothed** output (`htce.Smoothed`) into its own buffers, same as MBST's
`DominantCyclePeriod`/`DominantCyclePhase` do. SineWave is the canonical example.

**(b) Compose a raw `HilbertTransformerCycleEstimator` directly** — when the MBST
indicator pushes the **raw, unsmoothed** input sample into its own buffer, not
`htce.Smoothed`. `HilbertTransformerInstantaneousTrendLine` (HTITL) is the canonical
example: its buffer holds raw prices and only uses the HTCE for its period output.

### How to Tell Which Pattern Applies

Read the MBST indicator's `Update(double sample)` carefully. Look at what gets pushed
into the indicator's own circular buffer:

| MBST buffer fed with… | Zpano pattern |
|---|---|
| `htce.Smoothed` (WMA-smoothed price) | Compose `DominantCycle` — the smoothing is already inside DC |
| Raw `sample` (the un-smoothed input) | Compose raw HTCE via `hilberttransformer.NewCycleEstimator` / `createEstimator` |
| A different transform (e.g., hi-low range) | Compose raw HTCE; replicate transform in the outer |

### Implications for Pattern (b)

When composing a raw HTCE:

1. **Replicate the α-EMA period smoothing inline** — DC does `Periodᵢ = α·RawPeriodᵢ +
   (1−α)·Periodᵢ₋₁` internally; when you're not wrapping DC you must write that loop in
   the outer. Fields: `alphaEmaPeriodAdditional`, `oneMinAlphaEmaPeriodAdditional`,
   `smoothedPeriod`.
2. **Expose `DominantCyclePeriod` (smoothed) directly** — the outer owns the smoothed
   period and publishes it as its own output, not forwarded from DC. Mnemonic template
   stays `dcp(α%s%s)` for consistency with DC/SineWave.
3. **Use `htce.Period()` on each bar past priming** — no `htce.Smoothed()` involvement
   in the averaging window; that's pattern (a).
4. **Warm-up period is still `MaxPeriod*2 = 100`** — MBST convention for anything
   driven by an HTCE with the default auto-warmup. See "Warm-Up Period Defaults When
   Wrapping an HTCE" below.
5. **The component convention from the containing MBST class still applies** — e.g.,
   HTITL defaults to `BarMedianPrice` (like CoG/SineWave), so the component always shows
   in the mnemonic: `htitl(0.330, 4, 1.000, hl/2)`.

### Go Sketch (HTITL-style)

```go
type X struct {
    htce                           hilberttransformer.CycleEstimator
    alphaEmaPeriodAdditional       float64
    oneMinAlphaEmaPeriodAdditional float64
    smoothedPeriod                 float64
    input                          []float64 // RAW samples, not smoothed
    // ...
}

func (s *X) Update(sample float64) (float64, float64) {
    s.htce.Update(sample)
    s.pushInput(sample) // RAW sample, not s.htce.Smoothed()

    if s.primed {
        s.smoothedPeriod = s.alphaEmaPeriodAdditional*s.htce.Period() +
            s.oneMinAlphaEmaPeriodAdditional*s.smoothedPeriod
        // ... use s.input buffer + s.smoothedPeriod ...
    }
}
```

### TS Sketch

```typescript
this.htce = createEstimator(params.estimatorType, params.estimatorParams);
// ...
this.htce.update(sample);
this.pushInput(sample); // RAW

if (this.primed) {
    this.smoothedPeriod = this.alpha * this.htce.period + this.oneMinAlpha * this.smoothedPeriod;
    // ...
}
```

### Don't Mix Patterns

Do **not** wrap `DominantCycle` and then also create a separate `HilbertTransformerCycleEstimator`
to get a raw input path — that would double the HTCE cost and desync the period
smoothing. Pick one pattern per indicator based on what MBST's buffer holds.

---

### Band Output Semantics When Wrapping MBST's `Band`

MBST's `Band(DateTime, firstValue, secondValue)` carries no upper/lower semantic — it's
just a pair. When a merged zpano indicator emits a `Band` output, the **wrapping
indicator must choose** which value becomes `Upper` and which becomes `Lower`.

Convention: **primary/fast value → `Upper`, secondary/slow value → `Lower`**. This
matches MAMA (`Upper: mama, Lower: fama`) and SineWave (`Upper: value, Lower: lead`).
Document the choice in a short comment above the `Band{}` construction so reviewers
don't have to cross-reference MBST to confirm ordering.

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

---

## Merging Multiple MBST Indicators Into One Zpano Indicator

Some MBST indicators are closely related (share a common internal computation) but are
exposed as **separate** C# classes. In zpano they should be merged into a **single
multi-output indicator** to avoid duplicated state/warm-up and simplify the API.

**Example: MBST `DominantCyclePeriod` + `DominantCyclePhase` → zpano `DominantCycle`** with
three outputs: `RawPeriod`, `Period` (EMA-smoothed), `Phase`.

### When to Merge

Merge when all of the following hold:
1. The MBST classes each own their own `HilbertTransformerCycleEstimator` (or other
   expensive internal state) but conceptually **compute the same thing**.
2. They share the same primary parameter(s) (e.g., `alphaEmaPeriodAdditional` and
   estimator params).
3. Users typically want multiple outputs together (reading phase requires the period
   anyway).

### Pattern

- **One class, N outputs** — the indicator implements `Indicator` directly (MAMA-style),
  owns `barFunc`/`quoteFunc`/`tradeFunc`, and has an `updateEntity` helper emitting N
  `Scalar` entities (or `Band`, etc.) in the fixed output order.
- **One estimator instance**, one priming flag, one warm-up period for the whole
  indicator. The `primed` transition seeds all derived state in one shot.
- **`Update(sample)` returns a tuple** (Go: multi-return `(a, b, c)`, TS:
  `[a, b, c]` tuple). Each output has its own mnemonic; the indicator's top-level
  `Mnemonic()` returns the "primary" output's mnemonic (e.g., Period, not RawPeriod).
- **Output enum** lists all N outputs. Go: `RawPeriod Output = iota + 1, Period, Phase,
  outputLast`. TS: numeric enum with explicit `= 0, = 1, = 2`.
- **Single entry in `core.Identifier`** — one `DominantCycle` constant, not three.

### Constructor Flavors

Follow MAMA: provide `NewDominantCycleDefault` / `static default()` and
`NewDominantCycleParams` / `static fromParams()`. A private constructor does the actual
work. Avoid length-vs-smoothing-factor duplication when the indicator has only a single
scalar parameter — use one `fromParams` flavor.

### Merging on Top of Another Merged Indicator

The MBST indicators you're merging may themselves each instantiate an already-merged
zpano indicator (e.g., `SineWave`, `SineWaveLead`, `SineWaveBand` each own a
`DominantCyclePeriod`/`DominantCyclePhase` pair — already merged in zpano as
`DominantCycle`). In that case:

- **Don't duplicate the inner estimator in the outer.** Compose the already-merged
  inner indicator (see "Composing a Multi-Output Inner Indicator" above).
- **Forward the inner's relevant outputs** as additional outputs of the outer, so users
  don't need to instantiate both. SineWave's 5 outputs = its own 3 (Value/Lead/Band) +
  2 forwarded from DominantCycle (Period/Phase).
- **Reuse the inner's reference test arrays** where MBST's tests happen to embed them
  (MBST's SineWave test file contains the same `dcPeriod`/`dcPhase` arrays as the
  DominantCycle tests — port them once into the outer's `_test`).

---

## Warm-Up Period Defaults When Wrapping an HTCE

The `HilbertTransformerCycleEstimator` has two different default warm-up values depending
on how it's created:

| Code path | Default `WarmUpPeriod` |
|---|---|
| MBST C# constructor when caller passes `0` | `MaxPeriod * 2 = 100` |
| zpano HTCE internal auto-default (Go/TS) when caller passes `0` | `smoothingLengthPlus3HtLength ≈ 25` |

**This means the zpano HTCE primes earlier by default than MBST.** When porting an MBST
indicator that wraps HTCE and relies on the MBST default (`warmUpPeriod = MaxPeriod * 2`),
the zpano indicator's default factory **must explicitly set `warmUpPeriod: 100`** in the
estimator params — otherwise the port primes ~75 samples earlier and diverges from the
MBST reference test data.

```go
// Go: inside NewDominantCycleDefault
&hilberttransformer.CycleEstimatorParams{
    SmoothingLength:           4,
    AlphaEmaQuadratureInPhase: 0.2,
    AlphaEmaPeriod:            0.2,
    WarmUpPeriod:              100, // NOT 0 — MBST default is MaxPeriod*2
}
```

```typescript
// TS: inside static default()
estimatorParams: {
    smoothingLength: 4,
    alphaEmaQuadratureInPhase: 0.2,
    alphaEmaPeriod: 0.2,
    warmUpPeriod: 100, // NOT omitted — MBST default is MaxPeriod*2
}
```

When writing tests that call `NewDominantCycleParams(...)` with arbitrary estimator
params, passing `WarmUpPeriod: 0` is fine — the HTCE auto-default kicks in and the test
isn't comparing against MBST reference data anyway (it's comparing mnemonic shape, error
handling, etc.).

---

## Handling Structurally-Divergent Reference Data

Occasionally the reference data you're validating against comes from an implementation
that is **structurally different** from MBST's — e.g., TA-Lib Excel templates that
produce output from bar 0 by smoothing through zeros, while the MBST port follows a
strict priming convention (NaN until HTCE primed, then seed-EMA).

When this happens, an exact match is **algorithmically impossible** for the early portion
of the output series. Both implementations converge later because each subsequent EMA
step shrinks the seed error by `(1-α)`.

### Detection

- You've ported the algorithm line-by-line from MBST/C# and the code is correct.
- Early output values diverge from reference by a lot (e.g., 20+ units); late values
  match to 1e-5 or better.
- Error decays geometrically, not randomly.

### Test Strategy

1. **Add a `settleSkip` constant** (tuned empirically) — the index past which
   convergence is tight enough to assert.
2. **Sanity-check earlier indices** — assert output is finite / non-NaN, but skip the
   value comparison.
3. **Use a realistic tolerance** (e.g., `1e-4`) that accommodates both residual seed
   error at `settleSkip` and late-series floating-point accumulation.
4. **Document the rationale in a block comment** above `settleSkip` — future maintainers
   must understand this is algorithmic, not a port bug.

```go
const (
    skip       = 9   // TradeStation convention — skip first N bars.
    settleSkip = 177 // Samples required for EMA to converge past structural reference mismatch.
)

for i := skip; i < len(input); i++ {
    _, period, _ := dc.Update(input[i])
    if math.IsNaN(period) || i < settleSkip {
        continue
    }
    if math.Abs(expPeriod[i]-period) > tolerance {
        t.Errorf(...)
    }
}
```

### Phase/Angle Comparisons — Modulo 360

If the reference produces phase values **outside** the MBST port's `(-90°, 360°]` range
(e.g., 639.09°, -42.95°), you cannot compare directly. Use a `phaseDiff` helper that
computes the shortest signed angular difference modulo 360:

```go
func phaseDiff(a, b float64) float64 {
    d := math.Mod(a-b, 360)
    if d > 180 {
        d -= 360
    } else if d <= -180 {
        d += 360
    }
    return d
}
```

Then assert `math.Abs(phaseDiff(expected, actual)) < tolerance`.

Same helper applies in TS, using `(a - b) % 360` (JavaScript's `%` is remainder, which
preserves the sign of the dividend — same semantics as `math.Mod` for this purpose).

---

## Exposing Wrapped-Indicator Internals via Accessors

When a new indicator wraps an existing merged indicator (e.g., `TrendCycleMode` wraps
`DominantCycle`) and needs access to internal state the inner indicator already computes
— **add read-only accessors on the inner indicator** rather than duplicating the
computation in the outer.

Common examples on `DominantCycle`:

| Need | Add to inner | Avoid |
|---|---|---|
| WMA-smoothed price (`htce.Smoothed`) for the trendline | `SmoothedPrice() float64` / `get smoothedPrice(): number` | Re-instantiating a second HTCE in the outer |
| `MaxPeriod` constant for sizing the raw-input buffer | `MaxPeriod() int` / `get maxPeriod(): number` | Hard-coding `50` in the outer |

Rules:

1. Accessors are **read-only** and do not acquire locks when returning immutable
   constants (e.g., `MaxPeriod`). For mutable state (e.g., `SmoothedPrice` updated each
   `Update` call), use a `RLock`/`RUnlock` in Go; TS has no locking.
2. Add matching unit tests on the inner indicator confirming the accessor returns the
   expected value after priming.
3. Prefer accessors over widening the inner's `Update()` return tuple — adding more
   return values would force every existing caller to change.

---

## Go Error-Wrapping Depth When Wrapping Sub-Indicators

When an outer constructor (e.g., `newTrendCycleMode`) internally calls another
indicator's constructor (`dominantcycle.NewDominantCycleParams`), any validation error
returned by the inner gets **double-wrapped** by the outer's own `fmt.Errorf("invalid X
parameters: %w", err)` prefix.

For component errors (bar/quote/trade), the expected error string in tests becomes:

```
invalid trend cycle mode parameters: invalid dominant cycle parameters: 9999: unknown bar component
```

Not:
```
invalid trend cycle mode parameters: 9999: unknown bar component
```

Write the Go tests' expected error constants with the full double-prefix so the assertion
matches reality. TypeScript doesn't have this issue because TS component helpers
silently default to close (see next section).

---

## TS Component Helpers Don't Throw

A cross-platform portability trap: Go's `entities.BarComponentValue(c)`,
`QuoteComponentValue(c)`, `TradeComponentValue(c)` return an error for unknown
component enum values (typically `9999` in tests). **TypeScript's equivalents
(`barComponentValue`, `quoteComponentValue`, `tradeComponentValue`) do not throw** — they
silently fall back to the close/mid/price default.

Implications:

1. **Skip "invalid bar/quote/trade component" tests in TS specs** — there's no error
   path to exercise. Port all other param validation tests (α, length, ranges, etc.)
   as usual.
2. **Existing TS specs follow this convention** (SineWave, DominantCycle, CenterOfGravity
   all omit invalid-component tests). Don't add them to new indicators either.
3. **Go specs should keep the component error tests** — they exercise a real code path.

---

## Tuple-Output Indicator Spec Template (8 Outputs, TrendCycleMode Pattern)

When the outer indicator emits more than the typical 3-5 outputs (e.g., TrendCycleMode's
8-output tuple), the TS spec pattern is:

1. **Destructure with positional blanks** — use `const [, , , , sine] = x.update(v)` for
   5th element; avoid naming unused tuple elements.
2. **Single `checkOutput(out: any[])` helper** verifies `out.length === 8` and each is a
   `Scalar` with matching `time`. No `Band`-specific branch needed when all outputs are
   Scalar.
3. **Data arrays ported verbatim from Go test file** — replace `math.NaN()` with
   `Number.NaN`, drop Go `//nolint` tags.
4. **Phase array comparison uses mod-360 helper** (see above). Don't skip NaN-expected
   indices globally; skip per-element with `if (Number.isNaN(expected[i])) continue`.
5. **Reference-value array can be shorter than input array** — e.g., TCM's `expectedValue`
   has 201 entries while `input` has 252. Guard with `if (i >= limit) continue` inside the
   loop instead of shortening the outer iteration range (keeps the indicator fed with the
   full input sequence).

---

## Heatmap-Output Indicators (Corona Suite Pattern)

Some Ehlers indicators emit a **`Heatmap`** output (a 2-D intensity grid over time) in
addition to scalar outputs. The MBST Corona family (`CoronaSpectrum`,
`CoronaSignalToNoiseRatio`, `CoronaSwingPosition`, `CoronaTrendVigor`) is the canonical
reference. Key conventions learned porting them to zpano:

### Shared Helper (Not a Registered Indicator)

The MBST `Corona` base class isn't an indicator itself — it's a reusable helper
encapsulating the highpass filter, filter-bank, amplitude-squared matrix, and dominant-
cycle-median logic. Port it as a plain class/struct in a `corona/` subpackage, **do not
register a `core.Identifier`** for it, and do not give it an `update…Entity` path.
Downstream indicators instantiate the helper by composition.

Expose the internals the wrappers need as **public accessors**, same rule as
"Exposing Wrapped-Indicator Internals via Accessors":

| Wrapper needs | Helper exposes |
|---|---|
| Sizing the sample buffer (CTV) | `maximalPeriodTimesTwo` (Go: `MaximalPeriodTimesTwo()`, TS: `maximalPeriodTimesTwo` getter) |
| DC median for scalar output | `dominantCycleMedian` |
| Filter-bank amplitude matrix | `maximalAmplitudeSquared` |
| Per-period filter outputs | `filterBank`, `filterBankLength` |

### Empty Heatmap Invariant

When the indicator isn't primed yet (or the first sample for store-only indicators),
`update()` must still return a **well-formed heatmap with axis metadata but empty
values**:

- Go: `outputs.NewEmptyHeatmap(xAxis, yAxis, resolution, time)` — populates axes, leaves
  `Values` nil/empty, `IsEmpty()` returns true.
- TS: `Heatmap.newEmptyHeatmap(xAxis, yAxis, resolution, time)` — same semantics, plus a
  `.isEmpty()` method.

**Never return a null/zero heatmap.** Consumers rely on axes/resolution being present
across all bars so the UI can size the grid correctly from bar 0.

Scalar outputs during warm-up return `NaN` (Go `math.NaN()`, TS `Number.NaN`).

### First-Sample Store-Only Pattern

CSNR/CSwing/CTV all share an "isStarted" flag: the **very first** sample is stored into
the buffer but no computation happens (empty heatmap + NaN scalar). From the second
sample onward, normal primed-or-not logic takes over.

Tested with the standard 252-entry TA-Lib MAMA series; snapshot indices `{11, 12, 50,
100, 150, 200, 251}` give good coverage of warm-up, early-primed, mid-series, and
tail-series behavior.

### Parameter Resolution Quirks (Per Indicator)

- **CSpectrum** rounds user-supplied `minParam` **up** (`ceil`) and `maxParam` **down**
  (`floor`) — preserves the integer raster count. Example: `min=8.7, max=40.4` →
  `9, 40`.
- **CSwing / CTV** substitute their defaults (`±5` / `±10`) for Min/Max **only when
  both are zero** (unconfigured sentinel). If either is non-zero, both user values are
  honored. The "both zero" detection matches MBST's `if (min == 0 && max == 0)`.
- **CSNR** coefficient sum in the signal EMA is `0.2 + 0.9 = 1.1`, **not** 1.0. This is
  intentional per Ehlers; don't "fix" it.

### Heatmap Resolution Formula

Resolution is `(rasterLength - 1) / (maxRaster - minRaster)` for raster-based heatmaps,
or `(length - 1) / (maxParam - minParam)` for parameter-indexed heatmaps. Examples:

| Indicator | rasterLength | min/max | resolution |
|---|---|---|---|
| CSpectrum | maxParam-minParam+1 = 25 | 6..30 | (25-1)/24 = 1.0 (but the indicator uses raster=24, producing (len-1)/24 = 2.0 over a len-49 raster) |
| CSNR | 50 | 1..11 | 49/10 = 4.9 |
| CSwing | 50 | -5..5 | 49/10 = 4.9 |
| CTV | 50 | -10..10 | 49/20 = 2.45 |

When in doubt, **replicate the Go reference value exactly** — the formula varies
slightly per indicator and the Go implementation is the source of truth.

### TrendVigor Lookback Edge Case

CTV computes a lookback window as `int(DCM - 1)` where `DCM` is the dominant-cycle
median. Two guards are required (MBST has the first but missed the second):

1. **Lower bound:** guard `cyclePeriod >= 1` before using it (avoid zero-length loop).
2. **Upper bound:** clamp at `sampleBuffer.length` — otherwise a long cycle period
   during warm-up can over-index the buffer.

Go's `int(x)` and TS's `Math.trunc(x)` both truncate toward zero for both positive
and negative floats (unlike `Math.floor`, which rounds toward -∞). Safe to translate
directly.

### DominantCycleBuffer Sentinel

Initialize the DCM 5-element median buffer to `math.MaxFloat64` / `Number.MAX_VALUE`
sentinels (MBST convention) so the partial median is well-defined before the buffer
fills. Once primed, all 5 slots hold real values.

### Heatmap Snapshot Testing

For heatmap outputs, snapshot tests validate:

1. **Axis metadata** — `xAxis`, `yAxis` arrays, `resolution`, `time`.
2. **`isEmpty()` transition** — empty during warm-up, non-empty once primed.
3. **Scalar co-outputs** — DC/DCM/SNR/SP/TV values at fixed indices against Go
   reference (tolerance `1e-4`).

Don't snapshot the full `Values` grid verbatim — it's high-dimensional and noisy. Trust
the scalar co-output snapshots + axis/resolution invariants + `isEmpty()` transitions.

### Component Triple Mnemonic on the Main Output

Corona indicators default to `BarMedianPrice` (hl/2). The main heatmap output's
mnemonic always carries the `, hl/2` suffix (pass `BarComponent.Median` explicitly to
`ComponentTripleMnemonic`). Scalar co-outputs reuse the same component suffix —
they're computed from the same component, so the mnemonic must match.

Examples (defaults):
- `cspect(6, 20, 6, 30, 30, hl/2)` / `cspect-dc(30, hl/2)` / `cspect-dcm(30, hl/2)`
- `csnr(50, 20, 1, 11, 30, hl/2)` / `csnr-snr(30, hl/2)`
- `cswing(50, 20, -5, 5, 30, hl/2)` / `cswing-sp(30, hl/2)`
- `ctv(50, 20, -10, 10, 30, hl/2)` / `ctv-tv(30, hl/2)`

### Update Signature Quirk: CSNR Takes (sample, low, high)

`CoronaSignalToNoiseRatio.update(sample, sampleLow, sampleHigh, time)` — note the
**low-then-high** order, not high-then-low. `updateBar` naturally has both, so it
passes `(barFunc(bar), bar.low, bar.high)`. `updateScalar`/`updateQuote`/`updateTrade`
have no high/low, so they pass `(v, v, v)` — collapsing H=L=sample causes SNR to fall
back to `MinParameterValue` (SNR=0 would be invalid logarithmically).

---

## Suite-as-One-Indicator Merging Pattern (FIR-Bank / ATCF)

Some MBST suites consist of several **independent** indicators that share nothing but a
computational shape (all FIR filters on the same input series) plus a few composite
outputs derived from pairs of them. The canonical reference is Vladimir Kravchuk's
**Adaptive Trend and Cycle Filter (ATCF)** suite — 5 FIRs (FATL, SATL, RFTL, RSTL, RBCI)
+ 3 composites (FTLM=FATL−RFTL, STLM=SATL−RSTL, PCCI=sample−FATL). Merge the whole
suite into **one** zpano indicator with N scalar outputs.

### When This Pattern Applies

Merge when all of the following hold:
1. The MBST classes each compute a single scalar from the **same input component** using
   independent internal state (no shared estimator, unlike DominantCycle).
2. The classes have **no tunable parameters** beyond the input component (coefficients
   are hard-coded).
3. One or more "composite" outputs are obvious pair-wise combinations (A−B, sample−A).
4. Users virtually always plot multiple lines from the suite together.

This differs from the DominantCycle merge pattern (which merges because the components
**share expensive state**). Here the driver is API surface and ergonomics.

### Structure

- **One `core.Identifier`** (e.g., `AdaptiveTrendAndCycleFilter`).
- **One file per usual concern** (params, output enum, impl, coefficients, tests).
- **A private `firFilter`/`FirFilter` type in the impl file** — too small to warrant its
  own package/module. Holds `window []float64`, `coeffs []float64`, `count int`,
  `primed bool`, `value float64`. `Update(sample)` shifts the window left by one, appends
  the new sample at the last index, and computes `Σ window[i]·coeffs[i]`.
- **Coefficients in a separate file** (`…coefficients.go` / `…-coefficients.ts`) as
  exported package-level `var` / `Object.freeze`'d `readonly number[]`. Use only the
  normalized arrays from MBST; drop the commented-out originals.
- **N-tuple Update()** — Go multi-return `(fatl, satl, rftl, rstl, rbci, ftlm, stlm,
  pcci float64)`, TS `[number, number, …]` tuple. Order matches the output enum order
  exactly.

### Priming Semantics

- **Each FIR primes independently** when its own window fills (FATL at i=38, RFTL at
  i=43, RBCI at i=55, SATL at i=64, RSTL at i=90 for 39/44/56/65/91-tap windows).
- **Per-output NaN until that output's own dependencies are primed**:
  - FIR outputs are NaN until their individual FIR primes.
  - Composite outputs are NaN until **both** their inputs are primed (e.g., FTLM stays
    NaN until both FATL at i=38 AND RFTL at i=43 are primed → FTLM primes at i=43).
- **Indicator-level `IsPrimed()` mirrors the longest pole** (RSTL at i=90 for ATCF).
  This is the "indicator is fully useful" signal; individual outputs are already useful
  earlier.

### Mnemonic Convention

- **Top-level mnemonic uses the suite acronym:** `atcf(<components>)`. When all
  components are defaults (Close for Bar, standard for Quote/Trade),
  `componentTripleMnemonic` returns `""` and the mnemonic becomes `atcf()`.
- **Per-output mnemonics use the individual acronym:** `fatl(…)`, `satl(…)`, `rftl(…)`,
  etc. — all sharing the same component suffix as the top-level.
- **Default bar component is `Close`** (MBST's `ClosingPrice`), so `atcf()` has no
  suffix; `atcf(hl/2)` appears only on user override.

### Mnemonic Helper Trap

`componentTripleMnemonic` (Go and TS) returns either `""` or `", <bar>[, <quote>[,
<trade>]]"` **with a leading `", "`**. When building `atcf(<arg>)` you must strip the
leading `", "` — otherwise you get `atcf(, hl/2)`:

```go
cm := core.ComponentTripleMnemonic(p.BarComponent, p.QuoteComponent, p.TradeComponent)
arg := ""
if cm != "" {
    arg = cm[2:] // strip leading ", "
}
mnemonic := fmt.Sprintf("atcf(%s)", arg)
```

```typescript
const cm = componentTripleMnemonic(params.barComponent, params.quoteComponent, params.tradeComponent);
const arg = cm === '' ? '' : cm.substring(2);
const mnemonic = `atcf(${arg})`;
```

### Coefficient Transcription Workflow

1. **Copy only the normalized `readonly double[]`** from each MBST `.cs` file — ignore
   the commented-out originals (sum != 1).
2. **Verify total tap count** matches the first-priming index in MBST's `Update()`
   (e.g., FATL's 39-tap array should prime the FATL output at i=38).
3. **Use `var` block in Go** (package-level, immutable by convention — Go has no const
   arrays) and `Object.freeze(…) as readonly number[]` in TS.

### Snapshot Capture Workflow

Since this pattern has no reference test data in MBST (MBST has no ATCF test file), use
a **one-time capture** process:

1. Write the Go implementation.
2. Write a throwaway `cmd/<name>capture/main.go` that runs the 252-bar TA-Lib MAMA
   reference series through the indicator and prints outputs at the priming boundaries
   + a few mid/late indices.
3. **Hand-verify at least one index** by computing the FIR output directly in Python
   (`sum(coeffs[i]*input[i] for i in range(N))`) — confirms coefficient correctness.
4. Paste the captured tuples into `…_test.go` as the locked snapshots with `1e-10`
   tolerance.
5. Delete the throwaway capture program.
6. TS spec reuses the **same snapshot values** at `1e-10` tolerance — no independent TS
   capture needed.

Snapshot indices should include: `{0, first-FIR-primes, first+1, next-FIR-primes,
next+1, … longest-FIR-primes, longest+1, mid-1, mid-2, last}`. For ATCF that's `{0, 38,
39, 43, 44, 55, 56, 64, 65, 90, 91, 100, 150, 200, 251}`.

### Metadata Per Output

All N outputs have `Shape.Scalar` (registered in the descriptor row); list them in enum
order. The top-level `Description` is `"<Full Suite Name> <mnemonic>"`; per-output
descriptions are `"<Full Line Name> <per-output mnemonic>"` (e.g., `"Fast Adaptive Trend
Line fatl()"`).

### Files Produced (ATCF)

**Go (6 files, in `adaptivetrendandcyclefilter/` package):**
- `params.go` — `Params{BarComponent, QuoteComponent, TradeComponent}`.
- `output.go` — 8-member enum.
- `output_test.go` — enum round-trip.
- `adaptivetrendandcyclefilter.go` — private `firFilter` + main type.
- `coefficients.go` — 5 `var []float64`.
- `adaptivetrendandcyclefilter_test.go` — 252-bar snapshot + priming + NaN + metadata + UpdateEntity.

**TS (5 files, in `adaptive-trend-and-cycle-filter/` folder):**
- `params.ts`
- `output.ts`
- `adaptive-trend-and-cycle-filter.ts` — private `FirFilter` + main class.
- `coefficients.ts` — 5 frozen arrays.
- `adaptive-trend-and-cycle-filter.spec.ts` — mirrors Go tests at `1e-10` tolerance.

---

## Single-Heatmap-Output Indicators (GoertzelSpectrum Pattern)

Some MBST indicators emit exactly **one heatmap output and nothing else** — no scalar
co-outputs, no band, no dominant-cycle side channel. The canonical reference is
**GoertzelSpectrum** (`mbst-to-convert/custom/goertzel-spectrum/`). This is a simpler
shape than the Corona suite and deserves its own recipe.

### Shape

- **One registered `core.Identifier`** (e.g., `GoertzelSpectrum`), one package /
  folder, one primary output.
- **Output enum has exactly one member.** Go uses the `iota + 1 / outputLast` sentinel
  pattern for `IsKnown()` consistency with other indicators. TS uses an explicit
  `Value = 0` (no trailing sentinel needed — TS has no JSON Stringer plumbing).
- **`Update(sample, time)` returns `*outputs.Heatmap` / `Heatmap` directly** (not a
  tuple). The entity wrappers wrap it in a 1-element `core.Output` / `IndicatorOutput`
  array.
- **Descriptor row has one output** with `Shape.Heatmap` (registered in
  `descriptors.go` / `descriptors.ts`). The indicator-level mnemonic and the single
  output's mnemonic are identical.

### Unexported Estimator Pattern

When the MBST source uses a helper class (e.g., `GoertzelSpectrumEstimator`) that's only
ever instantiated by the indicator itself and has no separate public API, **port it as
an unexported/private type in the same package** — don't give it its own folder:

- Go: lowercase `estimator` struct in `estimator.go`, lowercase
  `newEstimator(...)` constructor, no separate test file. The public indicator is the
  only caller.
- TS: exported `GoertzelSpectrumEstimator` class (TS has no unexported classes) living
  next to the main class in the same module folder. It's **not** re-exported from any
  barrel/index — effectively internal.

This differs from the **Corona** pattern (where the helper is a shared base used by four
sibling indicators — worth its own subpackage). Use your judgment: if only one
indicator consumes the helper, keep it private.

### Axis-Reversal Convention

The MBST estimator fills its internal `spectrum[]` array **descending by period**:
`spectrum[0]` at `maxPeriod`, `spectrum[last]` at `minPeriod`. The zpano `Heatmap` axis,
however, runs `ParameterFirst = minPeriod → ParameterLast = maxPeriod` — **ascending**.

**Reverse on output.** Inside `Update`, walk the output values array `i = 0..lengthSpectrum-1`
and read from `spectrum[lengthSpectrum - 1 - i]`:

```go
for i := 0; i < lengthSpectrum; i++ {
    v := (s.estimator.spectrum[lengthSpectrum-1-i] - minRef) / spectrumRange
    values[i] = v
    // track valueMin/valueMax
}
```

Always document this in a comment — it's the single most common source of confusion
when comparing outputs against an MBST dump.

### Floating vs Fixed Normalization Semantics

MBST's `FloatingNormalization` boolean picks one of two normalization modes:

| Mode | `minRef` used in `(v - minRef) / (maxRef - minRef)` |
|---|---|
| **Floating** (default) | `spectrumMin` — both ends adapt to the current spectrum |
| **Fixed** | `0` — only the upper end adapts; lower end clamped to zero |

The AGC (`spectrumMax = max(decayFactor · previousMax, currentMax)`) still controls
`maxRef` in both modes. Don't conflate "AGC off" with "fixed normalization" — they are
orthogonal knobs:

- AGC off + floating → `maxRef = currentMax`, `minRef = spectrumMin`.
- AGC off + fixed    → `maxRef = currentMax`, `minRef = 0`.
- AGC on  + floating → `maxRef = AGC-smoothed`, `minRef = spectrumMin` (default).
- AGC on  + fixed    → `maxRef = AGC-smoothed`, `minRef = 0`.

### Inverted-Sentinel Boolean Params (Zero = MBST Default)

Go's zero value for `bool` is `false`. For an indicator where **true means "enabled"**
in MBST (SDC, AGC, floating normalization), a naive `SpectralDilationCompensation bool`
param would default to `false` = disabled — wrong. Two fixes:

1. **(Preferred) Invert the name** so zero-value = MBST default:
   `DisableSpectralDilationCompensation`, `DisableAutomaticGainControl`,
   `FixedNormalization`. Then inside the constructor: `sdcOn := !cfg.Disable…`.
2. An `AllFlagsZero` sentinel that flips every bool when all are zero — **don't do
   this.** It's clever but fragile; it breaks the moment a user sets one flag
   intentionally.

Apply the same naming in TS for cross-language consistency, even though TS's
`boolean | undefined` makes the sentinel trick unnecessary. Symmetric names make the
indicator easier to document and port.

First-order vs second-order Goertzel is a true binary choice (not
"enable/disable something"), so `IsFirstOrder bool` naturally defaults to false =
MBST default (second-order).

### Flag-Tag Mnemonic Pattern

When an indicator has many boolean flags / secondary knobs that are **rarely
overridden**, listing them all in the mnemonic is noisy:

```
gspect(64, 2, 64, 1, second-order, sdc-on, agc-on, agc-decay=0.991, floating-norm, hl/2)
```

Instead, emit **override-only terse tags** for non-default values, in a fixed order, and
omit everything at default. The GoertzelSpectrum format:

```
gspect(length, minPeriod, maxPeriod, spectrumResolution[, fo][, no-sdc][, no-agc][, agc=<f>][, no-fn][, <components>])
```

Tag rules:

| Tag | Emitted when |
|---|---|
| `fo` | `IsFirstOrder = true` (override of second-order default) |
| `no-sdc` | SDC disabled |
| `no-agc` | AGC disabled |
| `agc=<g-format>` | AGC on **and** decay differs from MBST default by > 1e-12 |
| `no-fn` | Fixed normalization (override of floating default) |

All tags are a **leading `", "`** plus the tag (no inner spaces). The component triple
mnemonic appends last with its own leading `", "`. Default: `gspect(64, 2, 64, 1, hl/2)`
— flags section is empty because all flags are at MBST defaults; `, hl/2` is present
only because `BarMedianPrice` is non-default framework-wide.

Factor this into a `buildFlagTags(...)` helper rather than inlining in `NewX` — the
helper is easier to unit-test and keeps the constructor readable.

### Files Produced (GoertzelSpectrum)

**Go (6 files, in `goertzelspectrum/` package):**
- `params.go` — `Params` with inverted-sentinel bool fields.
- `output.go` — single-member enum with Stringer/IsKnown/JSON.
- `output_test.go` — enum round-trip.
- `estimator.go` — unexported `estimator` struct (port of
  `GoertzelSpectrumEstimator.cs`).
- `goertzelspectrum.go` — main indicator + `buildFlagTags` helper.
- `goertzelspectrum_test.go` — snapshot + priming + NaN + metadata + mnemonic flag
  matrix + validation matrix + UpdateEntity.

**TS (5 files, in `goertzel-spectrum/` folder):**
- `params.ts` — mirrors Go field names (e.g.,
  `disableSpectralDilationCompensation?`).
- `output.ts` — `{ Value = 0 }`.
- `estimator.ts` — internal class, not barrel-exported.
- `goertzel-spectrum.ts` — main class + `buildFlagTags` helper.
- `goertzel-spectrum.spec.ts` — reuses Go snapshots at `1e-10` tolerance.

### Snapshot & Verification Workflow

No MBST reference test file existed. Workflow:

1. Implement Go indicator against MBST source line-by-line.
2. Hand-verify spot values at the priming index (`i = length-1`) against an independent
   **Python** reimplementation of the Goertzel second-order recurrence. Target agreement
   better than `1e-14` — floating-point noise only.
3. Capture snapshots at `{length-1, length, mid-1, mid-2, late}` indices (e.g.,
   `{63, 64, 100, 150, 200}` for default length=64). For each: `ValueMin`, `ValueMax`,
   and a handful of spot `Values[i]` covering first/mid/last bins.
4. TS spec reuses the same snapshots verbatim at the same `1e-10` tolerance (with a
   relaxed `1e-9` tolerance for `ValueMin`/`ValueMax` where only 10-sig-figs were
   captured).
5. Delete any throwaway capture program before committing.

### Second Exemplar: MaximumEntropySpectrum

**MaximumEntropySpectrum** (`mbst-to-convert/custom/maximum-entropy-spectrum/`, MBST
`Mbst.Trading.Indicators.SpectralAnalysis.MaximumEntropySpectrum`) follows the same
shape as GoertzelSpectrum with two MES-specific twists worth capturing:

1. **Burg AR estimator** — the internal estimator ports Paul Bourke's zero-based `ar.c`
   reference directly (`per/pef` working buffers, `g[1..degree]` coefficient
   accumulator, final sign flip `coefficients[i] = -g[i+1]`). Verify the Go port
   against the coefficient constants baked into `MaximumEntropySpectrumEstimatorTest.cs`
   (sinusoids / test1 / test2 / test3 — 7 cases, rounded to 0 or 1 decimal places as
   MBST does). Tolerance is MBST's rounding: `math.Round(v*10^dec)/10^dec`, **not** a
   delta. A dedicated `_coef_test.go` / `coefCases` block mirrors MBST's 4 test
   methods in one loop.

2. **Large MBST input arrays need to live in their own file, in both languages.**
   The coefficient tests depend on 4 MBST reference series (`inputFourSinusoids`
   = 999 floats, `inputTest1/2/3` = 1999 each). Transcribing these by hand is
   error-prone, so:

   - **Go:** put the arrays in a `<indicator>_data_test.go` sibling file (`//nolint:testpackage`
     same package, one `func testInput<Name>() []float64` per array). Keeps the main
     `_test.go` readable.
   - **TS:** put them in a `<indicator>.data.spec.ts` sibling file (Jasmine picks up
     `*.spec.ts`; the `.data.` infix signals data-only). Export as
     `readonly number[]` constants. The main `.spec.ts` imports from
     `./<indicator>.data.spec`.
   - **Generate the TS file from the Go file with a Python script** (regex `func (\w+)\(\) \[\]float64 \{\s*return \[\]float64\{(.*?)\}\s*\}`
     over the Go source, rewrite as `export const … = [...]`). Preserves the exact
     10-per-line grouping and avoids transcription drift. Same approach as the
     Bar-Based Indicators "Test Data Extraction" section.

**Flag tags for MES:** simpler than Goertzel because MES has no SDC and no
first/second-order choice. Only three tags: `no-agc`, `agc=<f>`, `no-fn`. Default
mnemonic: `mespect(60, 30, 2, 59, 1, hl/2)` (Length, Degree, MinPeriod, MaxPeriod,
SpectrumResolution, then flags, then component triple — `, hl/2` present because
MES also defaults to `BarMedianPrice`).

**Validation matrix** (both Go and TS):
`Length>=2`, `0<Degree<Length`, `MinPeriod>=2`, `MaxPeriod>MinPeriod`,
`MaxPeriod<=2*Length`, `SpectrumResolution>=1`, and (AGC on ⇒ decay in `(0,1)`).
Mirror identical error prefixes (`invalid maximum entropy spectrum parameters: …`)
so Go `errors.Is`-style asserts and TS `toThrowError(/…/)` assertions can share the
same wording.

**Priming:** at sample index `length - 1` (60 − 1 = 59 for default params), identical
semantics to Goertzel. Snapshot indices used: `{59, 60, 100, 150, 200}`.

### Third Exemplar: DiscreteFourierTransformSpectrum

**DiscreteFourierTransformSpectrum** (`mbst-to-convert/john-ehlers/discrete-fourier-transform-spectrum/`,
MBST `Mbst.Trading.Indicators.JohnEhlers.DiscreteFourierTransformSpectrum`) is the
third heatmap-family indicator and introduces the `no-sdc` flag.

1. **MBST ≠ Ehlers EL listing 9-1 ≠ Julia `dfts.jl`.** These are three different
   algorithms. MBST drops Ehlers' HP+SuperSmoother pre-filter and normalizes with
   `(v−min)/(max−min)` (min floating or 0), not `v/MaxPwr`. **Port MBST faithfully;
   do not attempt numerical parity with EL/Julia.** Document the divergence in the
   package doc comment.

2. **Synthetic-sine sanity test** replaces the Burg-coefficient tests MES uses.
   Inject `100 + sin(2π·i/period)` with the period chosen to integer-divide the
   window length (e.g. `period=16, length=48` → 3 full cycles, no DFT leakage),
   disable AGC/SDC/floating-norm so the peak reflects raw DFT magnitude, then
   assert peak bin = `period - ParameterFirst`. Picking a non-integer-divisor
   period (e.g. 20 in a 48-window) causes leakage and shifts the peak to an
   adjacent bin — always pick integer divisors.

3. **Four flags, full ordering:** `no-sdc, no-agc, agc=<f>, no-fn` (SDC is new
   relative to MES). Default mnemonic: `dftps(48, 10, 48, 1, hl/2)` (Length,
   MinPeriod, MaxPeriod, SpectrumResolution, flags, component triple — `hl/2`
   present because BarMedianPrice is default).

4. **Priming at `length - 1 = 47`** for defaults. Snapshot indices used:
   `{47, 60, 100, 150, 200}`.

**Validation matrix:** `Length>=2`, `MinPeriod>=2`, `MaxPeriod>MinPeriod`,
`MaxPeriod<=2*Length`, `SpectrumResolution>=1`, and (AGC on ⇒ decay in `(0,1)`).
No Degree check (DFTS has no AR order parameter). Error prefix: `invalid discrete
Fourier transform spectrum parameters: …`.

### Fourth Exemplar: CombBandPassSpectrum

**CombBandPassSpectrum** (`mbst-to-convert/john-ehlers/comb-band-pass-spectrum/`,
MBST `Mbst.Trading.Indicators.JohnEhlers.CombBandPassSpectrum`) is the fourth
heatmap-family indicator and the first one that **breaks away from MBST**: the
MBST C# implementation is misnamed — it actually computes a plain DFT identical
to DFTS. Port the **EasyLanguage listing 10-1 algorithm** from Ehlers' "Cycle
Analytics for Traders" instead. Document the MBST trap prominently in the
package doc comment and explicitly point users to DFTS for the MBST DFT.

1. **MBST-misnamed-DFT trap.** `CombBandPassSpectrumEstimator.cs` runs the same
   mean-subtracted DFT as DFTS — the "comb band-pass" name is a leftover label.
   Verifying the trap: check that the C# file imports no band-pass state (no
   per-period `bp[N,m]` buffers, no HP/SuperSmoother pre-filters, no β/γ/α
   coefficients). When this happens, the **EasyLanguage listing is the source
   of truth**, not the C# file.

2. **EL listing 10-1 pipeline** (in order, per sample):
   - 2-pole Butterworth highpass, cutoff = MaxPeriod:
     `α = (cos(ω) + sin(ω) − 1)/cos(ω)` with `ω = 0.707·2π/MaxPeriod`;
     `HP = (1−α/2)²·(c−2c[1]+c[2]) + 2(1−α)·HP[1] − (1−α)²·HP[2]`.
   - 2-pole SuperSmoother on HP, cutoff = MinPeriod:
     `a₁ = exp(−1.414π/MinPeriod)`, `b₁ = 2a₁·cos(1.414π/MinPeriod)`;
     `c₁ = 1−b₁+a₁²`, `c₂ = b₁`, `c₃ = −a₁²`;
     `Filt = c₁·(HP+HP[1])/2 + c₂·Filt[1] + c₃·Filt[2]`.
   - Bank of 2-pole band-pass filters, one per integer N in `[MinPeriod..MaxPeriod]`:
     `β = cos(2π/N)`, `γ = 1/cos(2π·bw/N)`, `α = γ − √(γ²−1)`;
     `BP[N,0] = 0.5(1−α)·(Filt−Filt[2]) + β(1+α)·BP[N,1] − α·BP[N,2]`.
   - Power per bin: `Pwr[N] = Σ_{m∈[0..N)} (BP[N,m]/Comp)²` with `Comp=N` when
     SDC is on, `1` otherwise.

3. **EL degrees → radians conversion.** EL's `Cosine`/`Sine` take degrees. The
   standard substitution is `cos(k·360/N deg) ≡ cos(k·2π/N rad)`. In particular
   `.707·360/N deg → .707·2π/N rad` and `1.414·180/N deg → 1.414·π/N rad`.
   Do the conversion once at the coefficient-building site; never mix units.

4. **EL normalization is exactly our AGC+FixedNormalization.** EL does:
   `MaxPwr = 0.995·MaxPwr; MaxPwr = max(MaxPwr, currentMax); Pwr/MaxPwr`.
   That is identical to our pattern `spectrumMax = decay·previousSpectrumMax`
   then max-scan, with `FixedNormalization = true` (min ref = 0). So EL-exact
   behavior is `fixedNormalization: true` with all other defaults.

5. **BP buffer is 2D and shift-then-write per bin per bar.** `bp[i]` is a
   `maxPeriod`-long ring-flavored array indexed by lag (0 = current,
   `maxPeriod-1` = oldest). Shift rightward (`bp[i][m] = bp[i][m-1]` for
   `m = maxPeriod-1 … 1`), **then** compute and store `bp[i][0]`. Inline the
   shift inside the main loop — it touches `O(lengthSpectrum·maxPeriod)` memory
   per bar. A ring buffer is a valid optimization but sacrifices direct lag
   indexing and complicates the power sum.

6. **Spectrum is stored in axis order, not reversed.** Bin `i` = period
   `minPeriod + i`, matching the heatmap's `minParameterValue → maxParameterValue`
   axis. Unlike DFTS (which fills `spectrum[0]` at MaxPeriod and reverses on
   output), CBPS writes `spectrum[i]` directly and the main file's output loop
   does **not** reverse.

7. **`estimator.update()` runs every bar**, even pre-prime, so the HP/SS/BP
   state warms up through zeros. The main file gates only the **output**:
   `windowCount++; if (windowCount >= primeCount) primed = true`. Prime at
   index `primeCount − 1 = maxPeriod − 1 = 47` for defaults.

8. **`spectrumRange > 0` guard** on the normalizing division. At prime bar and
   during long flat inputs, `max == min == 0`. Output 0 rather than NaN.

9. **Parameters:** no `Length` (always tied to MaxPeriod-sized BP history), no
   `SpectrumResolution` (integer periods only). New param: `Bandwidth` in `(0,1)`
   (Ehlers default 0.3). Default mnemonic: `cbps(10, 48, hl/2)`. Flag order:
   `bw=<f>, no-sdc, no-agc, agc=<f>, no-fn`. Default BarComponent = Median (hl/2),
   always shown in mnemonic.

10. **Synthetic-sine sanity test uses period 20, bars 400.** BP filters need far
    longer to settle than DFT; a 48-bar / 200-bar setup like DFTS will not
    place the peak cleanly. Using 20 (not an integer divisor of 48) is fine
    because BP filters don't leak like DFT — the peak lands precisely at bin
    `period − ParameterFirst`. Disable AGC/SDC/floating-norm for the test.

11. **Snapshot indices:** `{47, 60, 100, 150, 200}` (same as DFTS). `valueMin`
    is always 0 in the reference run because AGC+floating-norm clamps the scale.

**Validation matrix:** `MinPeriod>=2`, `MaxPeriod>MinPeriod`, `Bandwidth∈(0,1)`,
and (AGC on ⇒ decay `∈(0,1)`). No Length or SpectrumResolution checks. Error
prefix: `invalid comb band-pass spectrum parameters: …`.

### Fifth Exemplar: Autocorrelation Suite (ACI + ACP)

The **AutoCorrelationIndicator** (`aci`) and **AutoCorrelationPeriodogram** (`acp`)
from `mbst-to-convert/john-ehlers/auto-correlation-spectrum/` are the fifth and
sixth heatmap-family ports and repeat the CBPS "MBST misnamed algorithm" trap,
this time for a **two-indicator suite** where the second builds on the first.

1. **Second MBST misnamed-algo trap.** MBST's `AutoCorrelationCoefficients.cs`
   and `AutoCorrelationSpectrum.cs`:
   - Omit the HP+SuperSmoother pre-filter entirely.
   - Use a different Pearson formulation than EL listings 8-2 / 8-3.
   - Smooth **raw `SqSum`** rather than **`SqSum²`** (the EL default squares the
     Fourier magnitude before exponential smoothing: `R[P] = 0.2·SqSum² + 0.8·R_prev[P]`).
   - Invert the `AvgLength = 0` convention.

   **Port the EasyLanguage listings (8-2 and 8-3), not the MBST classes.** The
   trap is identical in spirit to CBPS but affects a two-class suite. Document
   the divergence in both package doc comments and in the skill.

2. **Two-stage composition, not inheritance.** ACP is a superset of ACI's Pearson
   correlation bank: ACI emits `0.5·(r+1)` per lag, ACP takes the same `corr[]`
   array and runs a DFT + smoothing + AGC on top. **Do not compose** — ACI isn't
   a reusable sub-indicator the way DominantCycle is. Instead, duplicate the
   HP → SS → Filt buffer → Pearson-bank preamble in both estimators. The Pearson
   logic is ~20 lines and the cost of a shared helper isn't worth the coupling.

3. **Pearson with variable M.** EL 8-2's `AvgLength = 0` means "use the lag as M"
   (i.e. `M = max(lag, 1)`). Ehlers' intent: at lag `L`, correlate `L` consecutive
   samples with `L` lagged samples, so longer lags use longer windows. EL 8-3
   hardcodes `M = 3` (a fixed short window for all lags). Params:
   - ACI: `AveragingLength` default `0` → `M = lag`; non-zero `N` overrides to `M = N`.
   - ACP: `AveragingLength` default `3` (fixed).
   Denominator guard: `denom = (M·Σx² − (Σx)²)·(M·Σy² − (Σy)²)`; if `denom <= 0`
   emit `r = 0` (flat window, undefined correlation).

4. **Filt history buffer sizing.** Pearson at lag `L` with window `M` reads
   `filt[0..M-1]` as x-series and `filt[L..L+M-1]` as y-series. Buffer length =
   `maxLag + mMax` (ACI, where `mMax = maxLag` when `AveragingLength = 0`, else
   `N`) or `maxPeriod + averagingLength` (ACP). Prime count = filt buffer length;
   every bar shifts rightward then writes `filt[0] = SS output`.

5. **EL smoothing target differs from MBST.** EL 8-3 smooths `SqSum²`
   (squared-then-squared-again); MBST smooths raw `SqSum`. Parameterize with a
   `DisableSpectralSquaring` flag (default `false`, i.e. squaring on matches
   EL). This is distinct from the SDC flag in CBPS — squaring is a post-DFT
   transform on each bin, not a per-bin normalization.

6. **ACI output scaling ≠ ACP output scaling.** ACI estimator already emits
   `0.5·(r+1)` in `[0, 1]` — no further normalization in the main file
   (`valueMin = spectrum.min, valueMax = spectrum.max` as-is). ACP estimator
   AGC-normalizes in-place to `[0, 1]`; main file only applies optional
   floating-minimum subtraction (`maxRef = 1.0, minRef = spectrumMin` when
   floating, else `0`). Do **not** reapply `(v − min)/(max − min)` to ACI.

7. **ACP DFT lag range starts at 3.** The inner loop over correlation bins for
   the DFT runs `n = 3..maxPeriod` (skipping lags 0/1/2 where the autocorrelation
   is dominated by noise/aliasing). Hard-code this as `DFT_LAG_START = 3` in
   the estimator; don't parameterize. The cos/sin basis tables must also skip
   bins 0..2 (fill with 0 or leave untouched — the inner loop never reads them).

8. **Synthetic-sine peak test caveat for pure autocorrelation.** Unlike DFT/BP
   indicators, an autocorrelation function peaks at **every integer multiple**
   of the true period (r[kP] ≈ 1 for all k). The ACI test must pick a period
   such that only **one multiple** fits in `[minLag, maxLag]`. For defaults
   `[3, 48]`, `period = 35` works (70 > 48, so only 35 itself is in range);
   `period = 20` does **not** (20 and 40 both in range, peak can land at
   either). ACP avoids this because the DFT concentrates power at the true
   frequency regardless of autocorrelation multiples — use `period = 20`
   (= bin `20 − 10 = 10`) with `DisableAGC + FixedNormalization`, same recipe
   as DFTS/CBPS.

9. **Parameter defaults and mnemonics.**
   - ACI: `MinLag=3, MaxLag=48, SmoothingPeriod=10, AveragingLength=0` (M=lag).
     Default mnemonic: `aci(3, 48, 10, hl/2)`. Only flag: `avg=N` when non-zero.
   - ACP: `MinPeriod=10, MaxPeriod=48, AveragingLength=3` (fixed M).
     Default mnemonic: `acp(10, 48, hl/2)`. Flags in Params field order:
     `avg=<n>, no-sqr, no-smooth, no-agc, agc=<f>, no-fn`.
   Both default to `BarComponent.Median` (hl/2), always shown in the mnemonic.

10. **Snapshot indices and priming.**
    - ACI: prime at `windowCount >= filtBufferLen` (defaults: 48 + 48 = 96).
      Snapshots at `{120, 150, 200, 250}`. `valueMin > 0` because ACI output is
      in [0, 1] centered around 0.5 (no normalization artifacts clamp to 0).
    - ACP: prime at `windowCount >= filtBufferLen` (defaults: 48 + 3 = 51).
      Snapshots at `{120, 150, 200, 250}`. `valueMin = 0` in the reference run
      because AGC + floating-norm clamp the scale (same as CBPS).

**Validation matrix (ACI):** `MinLag>=1`, `MaxLag>MinLag`, `SmoothingPeriod>=2`,
`AveragingLength>=0`. Error prefix: `invalid autocorrelation indicator parameters: …`.

**Validation matrix (ACP):** `MinPeriod>=2`, `MaxPeriod>MinPeriod`,
`AveragingLength>=1`, and (AGC on ⇒ decay `∈(0,1)`). Error prefix:
`invalid autocorrelation periodogram parameters: …`.
