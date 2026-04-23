---
name: talib-indicator-conversion
description: Step-by-step guide for converting TA-Lib C indicators to zpano Go and TypeScript. Load when converting a TaLib indicator or understanding the mapping between TaLib C source and zpano patterns.
---

# Converting TA-Lib C Indicators to Zpano

This guide provides recipes and tips for converting indicators from the TA-Lib C codebase
to the zpano multi-language library (Go and TypeScript).

Load the `indicator-architecture` and `indicator-conversion` skills alongside this one
for the full zpano architecture reference and internal conversion patterns.

Load the `mbst-indicator-conversion` skill as a companion — it covers the shared zpano
patterns (file layout, registration, output enums, metadata, etc.) in detail.

---

## Table of Contents

1. [Overview](#overview)
2. [TaLib Source Structure](#talib-source-structure)
3. [Test Data Extraction](#test-data-extraction)
4. [Algorithm Conversion](#algorithm-conversion)
5. [Composite Indicators and Circular Buffers](#composite-indicators-and-circular-buffers)
6. [Priming and NaN Count Calculation](#priming-and-nan-count-calculation)
7. [Test Strategy for TaLib Indicators](#test-strategy-for-talib-indicators)
8. [Worked Example: ADXR](#worked-example-adxr)
9. [Naming & Style Conventions](#naming--style-conventions)

---

## Naming & Style Conventions

All identifier, receiver, concurrency, style, and cross-language parity
rules are defined in the **`indicator-architecture`** skill and MUST be
followed during conversion. Summary (see that skill for the full tables
and rationale):

- **Abbreviations banned in identifiers** — always expand: `idx→index`,
  `tmp→temp`, `res→result`, `sig→signal`, `val→value`, `prev→previous`,
  `avg→average`, `mult→multiplier`, `buf→buffer`, `param→parameter`,
  `hist→histogram`.
- **TA-Lib–specific exception:** the C-source prefix conventions
  **`beg`** and **`out`** (as in `begIdx`, `outIdx`, `outReal`,
  `outBegIdx`, `outNbElement`) are **retained** — but the `Idx` part
  is normalized to `Index` (`begIdx` → `begIndex`, `outIdx` →
  `outIndex`). This keeps cross-reference with the TA-Lib C source
  straightforward while obeying the `idx→index` rule.
- **Go receivers** — compound type name → `s`; simple name →
  first-letter lowercased. All methods on a type use the same
  receiver. If a local would shadow `s`, rename the local to `str`.
- **Concurrency** — stateful public indicators MUST carry `mu
  sync.RWMutex`; writers `s.mu.Lock(); defer s.mu.Unlock()`, readers
  `s.mu.RLock(); defer s.mu.RUnlock()`.
- **Go style invariants** — no `var x T = zero`, no split var-then-
  assign; use `any`; no bare `make([]T, 0)`; grouped imports; every
  exported symbol has a doc comment.
- **Go ↔ TS local-variable parity** — same concept = same name in
  both languages. Canonical: `sum`, `epsilon`, `temp`/`diff`,
  `stddev`, `spread`, `amount`, `lengthMinOne`; loop counter
  `i`/`j`/`k`.

When porting a TA-Lib algorithm, keep the `beg*`/`out*` variable
**prefixes** for direct comparability with the C source, but
normalize `Idx` → `Index` and expand any other banned abbreviations
(`tmp`, `prev`, `sig`, `val`, etc.) encountered in the C source.

---

## Overview

TaLib indicators live in `mbst-to-convert/<author>/<indicator>/` as:
- **`.i` file** — the C implementation (often a `.c` file preprocessed for .NET Managed C++)
- **`test_*.c` file** — C test file with spot-check test cases using `TST_*` macros

The conversion produces:
- **Go:** `go/indicators/<author>/<indicator>/` (5 files: params, output, output_test, impl, impl_test)
- **TS:** `ts/indicators/<author>/<indicator>/` (4 files: params, output, impl, impl.spec)

Always convert Go first, then TypeScript.

### Key Difference from MBST Conversion

MBST provides C# source with `List<double>` arrays of 252 reference values that can be
directly embedded in tests. TaLib provides **spot-check test values** — individual
`(index, expected_value)` pairs at specific output positions, plus `expectedBegIdx` and
`expectedNbElement`. You typically need to **compute full expected arrays** from
already-verified inner indicator data or use the spot checks to validate.

---

## TaLib Source Structure

### The `.i` Implementation File

TaLib `.i` files are C source wrapped in .NET Managed C++ boilerplate. The actual
algorithm is buried inside. Key things to look for:

1. **Lookback function** — e.g., `AdxrLookback(int optInTimePeriod)` — tells you
   how many input bars are consumed before the first valid output.
2. **The main function** — e.g., `TA_ADXR(startIdx, endIdx, inHigh, inLow, inClose, optInTimePeriod, outBegIdx, outNbElement, outReal)`.
3. **Internal calls** — composite indicators call other TA functions internally
   (e.g., `TA_ADXR` calls `TA_ADX` internally).

Most of the `.i` file is auto-generated boilerplate (enum declarations, function
signatures for every TA function). **Scroll past this** to find the actual algorithm
for the indicator you're converting. Search for the function name (e.g., `Adxr`).

### The `test_*.c` Test File

TaLib test files use a table-driven approach:

```c
static TA_Test tableTest[] =
{
    { TST_ADXR, 1, 0, 0, 251, 14, TA_SUCCESS, 0,   19.8666, 40, 252-40 },
    { TST_ADXR, 0, 0, 0, 251, 14, TA_SUCCESS, 1,   18.9092, 40, 252-40 },
    { TST_ADXR, 0, 0, 0, 251, 14, TA_SUCCESS, 210, 21.5972, 40, 252-40 },
    { TST_ADXR, 0, 0, 0, 251, 14, TA_SUCCESS, 211, 20.4920, 40, 252-40 },
};
```

Each row is a `TA_Test` struct:

| Field | Meaning |
|---|---|
| `id` | Which indicator to test (e.g., `TST_ADXR`, `TST_ADX`, `TST_DX`) |
| `doRangeTestFlag` | Whether to run range tests (1=yes) |
| `unstablePeriod` | Unstable period override (usually 0) |
| `startIdx` | Start index of input range (usually 0) |
| `endIdx` | End index of input range (usually 251 for 252-bar history) |
| `optInTimePeriod` | The indicator's period parameter |
| `expectedRetCode` | Expected return code (`TA_SUCCESS`) |
| `oneOfTheExpectedOutRealIndex0` | Index into the **output array** to check |
| `oneOfTheExpectedOutReal0` | Expected value at that output index |
| `expectedBegIdx` | First valid output index (relative to input) |
| `expectedNbElement` | Number of valid output elements |

**Critical:** `oneOfTheExpectedOutRealIndex0` is an index into the output array
(starting from 0), NOT an index into the input array. The output array starts at
`expectedBegIdx` in input-space. So output[0] corresponds to input[expectedBegIdx].

### Test Data Source

TaLib tests use a shared `TA_History` with 252 bars of OHLCV data (the standard test
dataset). This data is NOT in the test file — it comes from the test framework. You
can obtain the same data from the existing zpano test files (e.g., the ADX test already
has 252-entry high/low/close arrays).

---

## Test Data Extraction

### From Spot Checks to Full Arrays

TaLib tests only verify specific output indices. For zpano, we test ALL output values.
Two approaches:

#### Approach 1: Compute from Inner Indicator (Preferred for Composite Indicators)

If the indicator wraps an already-verified inner indicator, compute expected values
from that inner indicator's verified output. Example: ADXR = (ADX[i] + ADX[i-(len-1)]) / 2.
Since ADX is already verified, compute ADXR expected values mathematically and validate
against the TaLib spot checks.

#### Approach 2: Extract from TaLib Reference Run

Run the TaLib C code against the standard 252-bar dataset and capture all output values.
This is more work but necessary for indicators without a simple derivation.

### Reusing Existing Test Data

Many TaLib test files test **multiple related indicators** in one file (e.g., `test_adxr.c`
tests MINUS_DM, PLUS_DM, MINUS_DI, PLUS_DI, DX, ADX, and ADXR). If the inner indicators
are already converted and tested in zpano, reuse their verified expected data.

---

## Algorithm Conversion

### C to Go/TS Translation

| C (TaLib) | Go | TS |
|---|---|---|
| `TA_REAL_MIN` | `-math.MaxFloat64` | `-Number.MAX_VALUE` |
| `TA_REAL_MAX` | `math.MaxFloat64` | `Number.MAX_VALUE` |
| `TA_INTEGER_DEFAULT` | `0` | `0` |
| `VALUE_HANDLE_DEREF(x)` | Direct field access | Direct property access |
| `TA_SetUnstablePeriod(...)` | N/A — zpano has no unstable period concept | N/A |
| Output buffer `outReal[outIdx++]` | Return value from `Update()` method | Return value from `update()` method |

### Batch vs Streaming

TaLib functions are **batch-oriented**: they take arrays and produce arrays. Zpano
indicators are **streaming**: they process one sample at a time via `Update()`.

The conversion requires restructuring:
1. **Identify state variables** — anything that persists between iterations becomes a
   struct field.
2. **Identify the loop body** — the inner loop of the TaLib function becomes the
   `Update()` method.
3. **Handle initialization** — the pre-loop setup (first N bars) becomes priming logic
   with a counter.

### Internal TA Function Calls

When TaLib calls another TA function internally (e.g., `TA_ADXR` calls `TA_ADX`),
zpano uses **composition**: the outer indicator holds an instance of the inner indicator
and delegates to it.

```go
// TaLib: TA_ADXR calls TA_ADX internally on the same input
// Zpano: ADXR holds an ADX instance
type AverageDirectionalMovementIndexRating struct {
    adx       *averagedirectionalmovementindex.AverageDirectionalMovementIndex
    adxBuffer []float64  // circular buffer for past ADX values
    // ...
}
```

---

## Composite Indicators and Circular Buffers

### The Lookback Pattern

Many TaLib composite indicators need to look back at past values of an inner indicator.
Example: ADXR averages ADX[current] with ADX[current - (length-1)].

**Zpano pattern:** Use a **circular buffer** (ring buffer) to store past inner indicator
values.

### Go Circular Buffer Implementation

```go
type AverageDirectionalMovementIndexRating struct {
    adx         *adx.AverageDirectionalMovementIndex
    adxBuffer   []float64   // size = length
    bufferIndex int         // current write position
    bufferCount int         // how many values stored so far
    length      int
}

func (s *AverageDirectionalMovementIndexRating) Update(close, high, low float64) float64 {
    adxValue := s.adx.Update(close, high, low)
    if math.IsNaN(adxValue) {
        return math.NaN()
    }

    // Store in circular buffer
    s.adxBuffer[s.bufferIndex] = adxValue
    s.bufferIndex = (s.bufferIndex + 1) % s.length
    s.bufferCount++

    // Need length values to compute ADXR
    if s.bufferCount < s.length {
        return math.NaN()
    }

    // ADXR = (currentADX + ADX from (length-1) periods ago) / 2
    oldIndex := s.bufferIndex % s.length  // oldest value in buffer
    return (adxValue + s.adxBuffer[oldIndex]) / 2.0
}
```

### TS Circular Buffer Implementation

Same pattern using a plain `number[]` array.

---

## Priming and NaN Count Calculation

### Computing Expected NaN Count

For composite indicators, the total NaN count (priming period) is the sum of:
1. The inner indicator's priming period
2. Additional bars needed to fill the lookback buffer

**Example (ADXR with length=14):**
- ADX primes at index 27 (first valid ADX at index 27, so 27 NaN values)
- ADXR needs `length - 1 = 13` additional ADX values to fill its circular buffer
- Total: first valid ADXR at index 27 + 13 = **40**
- With 252 input bars: `expectedBegIdx=40`, `expectedNbElement=252-40=212`

### CRITICAL: C# Test Data Offset Bug (Also Applies to TaLib)

TaLib batch functions return a compact output array starting at `outBegIdx`. There are
no NaN values in the output — the output simply starts later.

Zpano streaming indicators return NaN for every input before priming. When converting
test expectations:
- **TaLib output[0]** = zpano output at input index `expectedBegIdx`
- **TaLib output[N]** = zpano output at input index `expectedBegIdx + N`
- Zpano outputs at indices 0 through `expectedBegIdx - 1` are all NaN

When computing expected arrays for zpano tests, prepend `expectedBegIdx` NaN values
before the valid output values.

---

## Test Strategy for TaLib Indicators

### Spot-Check Validation

Use the TaLib test table values to validate key points:

```go
// From test_adxr.c: { TST_ADXR, ..., 0, 19.8666, 40, 212 }
// Output index 0 at begIdx 40 -> input index 40
require.InDelta(t, 19.8666, results[40], 1e-4)
```

**Tolerance:** TaLib test values are typically rounded to 4 decimal places. Use `1e-4`
tolerance for spot checks.

### Full Array Validation

For composite indicators where expected values can be computed:
1. Use already-verified inner indicator expected data
2. Apply the composite formula to generate full expected arrays
3. Validate ALL output values with tight tolerance (e.g., `1e-10`)
4. Cross-check against TaLib spot values with `1e-4` tolerance

### Test Categories (Same as MBST)

1. **Output value tests** — Full array + spot checks
2. **IsPrimed tests** — Verify priming transitions at the correct index
3. **Metadata tests** — Type, mnemonic, description, outputs
4. **Constructor validation** — Invalid params produce errors
5. **UpdateEntity tests** — Scalar, Bar, Quote, Trade routing
6. **NaN handling** — NaN input doesn't corrupt state
7. **Output enum tests** — String, IsKnown, MarshalJSON, UnmarshalJSON (Go only)

---

## Worked Example: ADXR

### Source Analysis

From `ta_ADXR.i`, the ADXR algorithm:
1. Computes ADX for the full input range
2. For each output: `ADXR[i] = (ADX[i] + ADX[i - (optInTimePeriod - 1)]) / 2`

Lookback: `ADX_Lookback(period) + (period - 1)`

### Test Data from `test_adxr.c`

```
expectedBegIdx = 40, expectedNbElement = 212
output[0]   = 19.8666  (first value)
output[1]   = 18.9092
output[210] = 21.5972
output[211] = 20.4920  (last value)
```

### Conversion Decisions

1. **Composite pattern**: ADXR wraps an internal ADX instance
2. **Circular buffer**: Size = `length` (14), stores past ADX values
3. **9 outputs**: ADXR value + all 8 ADX sub-outputs (ADX, DX, DI+, DI-, DM+, DM-, ATR, TR)
4. **Bar-based**: Takes high/low/close, not a single scalar component

### Expected Data Computation

Since ADX is already verified in zpano (with 252-entry expected arrays), ADXR expected
values were computed as:
```
adxr[i] = (adxExpected[i] + adxExpected[i - 13]) / 2.0   // for i >= 40
adxr[i] = NaN                                              // for i < 40
```

The computed values were validated against all 4 TaLib spot checks within 1e-4 tolerance.

### Files Produced

**Go (5 files, in `averagedirectionalmovementindexrating/` package):**
- `params.go` — `Params` struct with `Length`
- `output.go` — 9-value `Output` enum
- `output_test.go` — Table-driven output enum tests
- `averagedirectionalmovementindexrating.go` — Implementation with ADX composition + circular buffer
- `averagedirectionalmovementindexrating_test.go` — Full 252-entry tests + spot checks

**TS (4 files, in `average-directional-movement-index-rating/` folder):**
- `params.ts` — Params interface
- `output.ts` — 9-value output enum
- `average-directional-movement-index-rating.ts` — Implementation
- `average-directional-movement-index-rating.spec.ts` — Full tests + spot checks

### Registration

- Go: Added `AverageDirectionalMovementIndexRating` to `go/indicators/core/identifier.go`
  (enum constant, string, String(), MarshalJSON, UnmarshalJSON) and all 4 test tables
  in `identifier_test.go`. Registered a descriptor row in `go/indicators/core/descriptors.go`
  (taxonomy dimensions + per-output `Kind`/`Shape`); `Metadata()` calls `core.BuildMetadata(...)`.
- TS: Added `AverageDirectionalMovementIndexRating` to
  `ts/indicators/core/indicator-identifier.ts`. Registered the matching descriptor row in
  `ts/indicators/core/descriptors.ts`; `metadata()` calls `buildMetadata(...)`.

See `.opencode/skills/indicator-architecture/SKILL.md` section "Taxonomy & Descriptor
Registry" for descriptor-row guidance and field meanings.
