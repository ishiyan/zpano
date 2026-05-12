---
name: signals-architecture
description: Architecture, design decisions, and implementation reference for the zpano fuzzy signals library. Load when creating new signal types, porting signals across languages, or understanding the signal composition system.
---

# Signals Architecture

Architecture, algorithms, and implementation reference for the fuzzy signals library in zpano. This package provides reusable signal primitives for interpreting technical indicator output using fuzzy logic.

> **Related skill:** The `fuzzy-architecture` skill covers the underlying fuzzy logic primitives (membership functions, operators, defuzzification) that this module builds upon. Load it for details on sigmoid/linear membership functions, t-norms, s-norms, alpha-cut, and the `MembershipShape` enum.

## Overview

### What Are Fuzzy Signals?

Traditional technical analysis uses crisp boolean conditions:
- "RSI > 70" = overbought (yes/no)
- "MACD histogram crosses zero" = buy signal (yes/no)
- "Price above upper Bollinger Band" = breakout (yes/no)

These hard thresholds lose information. RSI at 69.9 produces the same "not overbought" as RSI at 50. A MACD histogram barely crossing zero is treated identically to a strong thrust through zero.

Fuzzy signals replace boolean conditions with **membership degrees** in [0, 1]:
- RSI at 69.9 vs threshold 70 produces mu ~ 0.49 (almost overbought)
- RSI at 85 vs threshold 70 produces mu ~ 0.997 (clearly overbought)
- RSI at 50 vs threshold 70 produces mu ~ 0.0003 (not overbought)

The continuous output preserves the distance-from-threshold information, enabling proportional position sizing, weighted signal combination, and smoother strategy behavior.

### Design Principles

1. **Stateless.** All signal functions are pure: they take raw indicator values and return membership degrees. No internal state, no bar history, no warmup period.
2. **Zero dependency on indicators.** Signals depend only on the `fuzzy/` module. They accept raw `float` values and don't know or care which indicator produced them.
3. **Composable.** Primitive signals combine via `signal_and`, `signal_or`, `signal_not` to build complex conditions. The composition operators are themselves fuzzy, preserving the continuous nature throughout.
4. **Shape-configurable.** All membership functions accept a `MembershipShape` parameter (sigmoid or linear), allowing the caller to control transition smoothness.
5. **Width-configurable.** The `width` parameter controls how gradual the transition is. Width = 0 degrades to crisp boolean logic, providing backward compatibility.

## Module Dependencies

```
fuzzy/                  (standalone — membership functions, operators)
  |
  v
signals/                (standalone — signal primitives for indicator interpretation)
  |
  v
(consumers)             (trading systems, strategy engines, backtesting)
```

The signals module imports from `fuzzy/` and has **zero other dependencies**. It does not import from `indicators/`, `entities/`, or any other zpano module.

## Package Structure

### Python (`py/signals/`)
```
__init__.py                    # Barrel exports (16 functions)
threshold.py                   # mu_above, mu_below, mu_overbought, mu_oversold
crossover.py                   # mu_crosses_above, mu_crosses_below, mu_line_crosses_above, mu_line_crosses_below
band.py                        # mu_above_band, mu_below_band, mu_between_bands
histogram.py                   # mu_turns_positive, mu_turns_negative
compose.py                     # signal_and, signal_or, signal_not, signal_strength
test_threshold.py              # 13 tests
test_crossover.py              # 12 tests
test_band.py                   # 12 tests
test_histogram.py              # 10 tests
test_compose.py                # 13 tests
```

### Go (`go/signals/`)
```
doc.go                         # Package documentation
threshold.go                   # MuAbove, MuBelow, MuOverbought, MuOversold
crossover.go                   # MuCrossesAbove, MuCrossesBelow, MuLineCrossesAbove, MuLineCrossesBelow
band.go                        # MuAboveBand, MuBelowBand, MuBetweenBands
histogram.go                   # MuTurnsPositive, MuTurnsNegative
compose.go                     # SignalAnd, SignalOr, SignalNot, SignalStrength
signals_test.go                # Combined test file (~30 tests)
```

### TypeScript (`ts/signals/`)
```
index.ts                       # Barrel re-exports
threshold.ts                   # muAbove, muBelow, muOverbought, muOversold
crossover.ts                   # muCrossesAbove, muCrossesBelow, muLineCrossesAbove, muLineCrossesBelow
band.ts                        # muAboveBand, muBelowBand, muBetweenBands
histogram.ts                   # muTurnsPositive, muTurnsNegative
compose.ts                     # signalAnd, signalOr, signalNot, signalStrength
threshold.spec.ts              # Per-file spec files
crossover.spec.ts
band.spec.ts
histogram.spec.ts
compose.spec.ts
```

### Zig (`zig/src/signals/`)
```
threshold.zig                  # muAbove, muBelow, muOverbought, muOversold + inline tests
crossover.zig                  # muCrossesAbove, muCrossesBelow, muLineCrossesAbove, muLineCrossesBelow + inline tests
band.zig                       # muAboveBand, muBelowBand, muBetweenBands + inline tests
histogram.zig                  # muTurnsPositive, muTurnsNegative + inline tests
compose.zig                    # signalAnd, signalOr, signalNot, signalStrength + inline tests
```

### Rust (`rs/src/signals/`)
```
mod.rs                         # Module root with pub use re-exports
threshold.rs                   # mu_above, mu_below, mu_overbought, mu_oversold + inline tests
crossover.rs                   # mu_crosses_above, mu_crosses_below, mu_line_crosses_above, mu_line_crosses_below + inline tests
band.rs                        # mu_above_band, mu_below_band, mu_between_bands + inline tests
histogram.rs                   # mu_turns_positive, mu_turns_negative + inline tests
compose.rs                     # signal_and, signal_or, signal_not, signal_strength + inline tests
```

## Signal Types

### 1. Threshold Signals (`threshold`)

Measure the degree to which a value is above or below a fixed level. The fundamental building block for overbought/oversold conditions.

#### Functions

| Function | Description | Algorithm |
|----------|-------------|-----------|
| `mu_above(value, threshold, width, shape)` | Degree value > threshold | `mu_greater(value, threshold, width, shape)` |
| `mu_below(value, threshold, width, shape)` | Degree value < threshold | `mu_less(value, threshold, width, shape)` |
| `mu_overbought(value, level, width, shape)` | Semantic alias for mu_above | `mu_greater(value, level, width, shape)` |
| `mu_oversold(value, level, width, shape)` | Semantic alias for mu_below | `mu_less(value, level, width, shape)` |

#### Algorithm

Delegates directly to fuzzy membership functions. For sigmoid shape:
```
mu_above(x, t, w) = 1 / (1 + exp(-k * (x - t)))    where k = 12 / w
mu_below(x, t, w) = 1 - mu_above(x, t, w)
```

At `x = threshold`: mu = 0.5 exactly. The transition zone spans `threshold +/- width/2`.

#### Key Properties

- `mu_above + mu_below = 1.0` for any value (exact complement)
- Monotonically increasing (mu_above) / decreasing (mu_below) in value
- Width = 0 produces crisp step function: 1.0 if above, 0.0 if below, 0.5 if equal

#### Default Parameters

- `mu_overbought`: level = 70.0 (standard RSI overbought)
- `mu_oversold`: level = 30.0 (standard RSI oversold)
- Both: width = 5.0 (moderate fuzzy transition)

#### Usage Examples

```python
from py.signals import mu_above, mu_below, mu_overbought, mu_oversold

# RSI overbought with default level 70
confidence = mu_overbought(rsi_value)  # mu in [0, 1]

# Custom Stochastic oversold at 20 with wide transition
confidence = mu_oversold(stoch_value, level=20.0, width=10.0)

# Generic: is CCI above +100?
confidence = mu_above(cci_value, threshold=100.0, width=20.0)
```

### 2. Crossover Signals (`crossover`)

Detect when a value or line pair crosses a level between two consecutive time steps. A crossover is the conjunction of two conditions: "was on one side" AND "is now on the other side."

#### Functions

| Function | Description | Algorithm |
|----------|-------------|-----------|
| `mu_crosses_above(prev, curr, threshold, width, shape)` | Value crossed above threshold | `mu_below(prev) * mu_above(curr)` |
| `mu_crosses_below(prev, curr, threshold, width, shape)` | Value crossed below threshold | `mu_above(prev) * mu_below(curr)` |
| `mu_line_crosses_above(prev_fast, curr_fast, prev_slow, curr_slow, width, shape)` | Fast line crossed above slow line | Reduces to `mu_crosses_above(prev_diff, curr_diff, 0)` |
| `mu_line_crosses_below(prev_fast, curr_fast, prev_slow, curr_slow, width, shape)` | Fast line crossed below slow line | Reduces to `mu_crosses_below(prev_diff, curr_diff, 0)` |

#### Algorithm

Crossover = product t-norm of two membership degrees:
```
mu_crosses_above(prev, curr, t, w) = mu_less(prev, t, w) * mu_greater(curr, t, w)
```

This naturally produces:
- **1.0** when prev is clearly below AND curr is clearly above
- **0.0** when either condition fails (no crossover, or wrong direction)
- **0.25** when both values are exactly at threshold (0.5 * 0.5)
- **Partial values** for near-threshold crossovers with width > 0

Line crossover reduces to threshold crossover of the difference:
```
mu_line_crosses_above(pf, cf, ps, cs, w) = mu_crosses_above(pf - ps, cf - cs, 0, w)
```

#### Key Properties

- `mu_crosses_above(a, b, t, w) == mu_crosses_below(b, a, t, w)` (symmetry)
- With width = 0: crisp crossover detection (1.0 or 0.0, with 0.25 at exact threshold)
- The product t-norm ensures BOTH conditions must hold; one-sided membership can't produce a false positive

#### Default Parameters

- width = 0.0 (crisp crossover by default — most crossover strategies want exact detection)

#### Usage Examples

```python
from py.signals import mu_crosses_above, mu_line_crosses_above

# RSI crossing above 30 (leaving oversold)
signal = mu_crosses_above(prev_rsi, curr_rsi, threshold=30.0, width=5.0)

# Golden cross: 50-day MA crossing above 200-day MA
signal = mu_line_crosses_above(
    prev_fast=prev_ma50, curr_fast=curr_ma50,
    prev_slow=prev_ma200, curr_slow=curr_ma200,
    width=0.5
)

# Death cross: 50-day MA crossing below 200-day MA
signal = mu_line_crosses_below(
    prev_fast=prev_ma50, curr_fast=curr_ma50,
    prev_slow=prev_ma200, curr_slow=curr_ma200,
    width=0.5
)
```

### 3. Band Signals (`band`)

Measure the degree to which a value is above, below, or between dynamic bands (Bollinger Bands, Keltner Channels, Donchian, envelopes).

#### Functions

| Function | Description | Algorithm |
|----------|-------------|-----------|
| `mu_above_band(value, upper, width, shape)` | Degree above upper band | `mu_greater(value, upper, width, shape)` |
| `mu_below_band(value, lower, width, shape)` | Degree below lower band | `mu_less(value, lower, width, shape)` |
| `mu_between_bands(value, lower, upper, shape)` | Degree inside the channel | `mu_greater(value, lower, w) * mu_less(value, upper, w)` where `w = spread * 0.5` |

#### Algorithm — `mu_between_bands`

This is the most interesting function in the module. It computes band containment as a product of two sided memberships:

```
spread = upper - lower
width = spread * 0.5
mu_between = mu_greater(value, lower, width) * mu_less(value, upper, width)
```

Key design decision: the `width` is derived from the band spread itself (not caller-supplied). This means:
- Narrow bands produce sharp transitions (tight containment)
- Wide bands produce gradual transitions (loose containment)
- The membership peaks at the center of the bands and falls smoothly toward the edges
- At either band edge: mu ~ 0.5 * 1.0 = 0.5 (half-contained)

Edge case: if `upper <= lower`, returns 0.0 (degenerate bands).

#### Key Properties

- `mu_between_bands` is bell-shaped: peaks at band center, falls toward edges
- Centered value (exactly between bands): mu > 0.8
- At band edge: mu ~ 0.5
- Well outside bands: mu ~ 0.0
- Monotonically decreasing from center outward in both directions

#### Usage Examples

```python
from py.signals import mu_above_band, mu_below_band, mu_between_bands

# Bollinger Band breakout
breakout_up = mu_above_band(price, bb_upper, width=bb_width * 0.1)
breakout_down = mu_below_band(price, bb_lower, width=bb_width * 0.1)

# Price contained within bands (mean-reversion regime)
contained = mu_between_bands(price, bb_lower, bb_upper)
```

### 4. Histogram Signals (`histogram`)

Detect sign changes in oscillator histograms (MACD histogram, Awesome Oscillator, etc.). These are specialized crossover signals with a fixed threshold at zero.

#### Functions

| Function | Description | Algorithm |
|----------|-------------|-----------|
| `mu_turns_positive(prev, curr, width, shape)` | Histogram went from <=0 to >0 | `mu_less(prev, 0) * mu_greater(curr, 0)` |
| `mu_turns_negative(prev, curr, width, shape)` | Histogram went from >=0 to <0 | `mu_greater(prev, 0) * mu_less(curr, 0)` |

#### Algorithm

Identical to `mu_crosses_above` / `mu_crosses_below` with threshold fixed at 0:
```
mu_turns_positive(prev, curr, w) = mu_less(prev, 0, w) * mu_greater(curr, 0, w)
mu_turns_negative(prev, curr, w) = mu_greater(prev, 0, w) * mu_less(curr, 0, w)
```

#### Key Properties

- `mu_turns_positive(-a, a, w) == mu_turns_negative(a, -a, w)` (symmetry around zero)
- From exactly zero to positive: mu = 0.5 * 1.0 = 0.5 (prev was on boundary)
- These are convenience wrappers; the same result can be achieved with `mu_crosses_above(prev, curr, 0.0, width)`

#### Usage Examples

```python
from py.signals import mu_turns_positive, mu_turns_negative

# MACD histogram turning bullish
bullish = mu_turns_positive(prev_macd_hist, curr_macd_hist, width=0.5)

# Awesome Oscillator turning bearish
bearish = mu_turns_negative(prev_ao, curr_ao, width=0.1)
```

### 5. Composition (`compose`)

Combine multiple signal outputs using fuzzy logic operators. These are thin wrappers over `fuzzy/operators` with signal-domain naming for readability.

#### Functions

| Function | Description | Algorithm |
|----------|-------------|-----------|
| `signal_and(*signals)` | All conditions must hold | `t_product_all(signals)` = product of all |
| `signal_or(a, b)` | Either condition suffices | `s_probabilistic(a, b)` = a + b - a*b |
| `signal_not(signal)` | Negate a condition | `f_not(signal)` = 1 - signal |
| `signal_strength(signal, min_strength)` | Filter weak signals to zero | `signal if signal >= min_strength else 0` |

#### Algorithm — `signal_and` (Product T-Norm)

```
signal_and(s1, s2, ..., sn) = s1 * s2 * ... * sn
```

Behavior with multiple conditions:
- All at 0.9: `0.9^n` — degrades with more conditions (5 conditions at 0.9 = 0.59)
- One at 0.0: result = 0.0 — any failed condition kills the combination
- All at 1.0: result = 1.0 — identity
- Empty input: 1.0 (identity element of product)

#### Algorithm — `signal_or` (Probabilistic S-Norm)

```
signal_or(a, b) = a + b - a*b
```

Properties:
- `signal_or(a, b) >= max(a, b)` — always at least as strong as the strongest input
- `signal_or(a, 0) = a` — identity element is 0
- `signal_or(1, 1) = 1` — bounded
- Not associative in the strict algebraic sense, but can be chained: `signal_or(signal_or(a, b), c)`

#### Algorithm — `signal_strength` (Soft Alpha-Cut)

```
signal_strength(s, alpha) = s    if s >= alpha
                          = 0.0  otherwise
```

**Not the same as `alpha_cut` from defuzzify.** Key differences:
- `signal_strength` preserves the continuous value for signals above the threshold
- `alpha_cut` discretizes to integer multiples of scale (e.g., {-200, -100, 0, +100, +200})
- `signal_strength` operates on [0, 1] membership degrees; `alpha_cut` operates on [-200, +200] pattern outputs

#### Usage Examples

```python
from py.signals import (
    mu_oversold, mu_crosses_above, mu_turns_positive,
    signal_and, signal_or, signal_not, signal_strength
)

# Complex buy condition: RSI oversold AND MACD crosses up AND price not overbought
buy = signal_and(
    mu_oversold(rsi, level=30, width=5),
    mu_crosses_above(prev_macd, curr_macd, threshold=0, width=0.5),
    signal_not(mu_overbought(rsi, level=70, width=5))
)

# Alternative entry: either RSI oversold OR MACD histogram turns positive
entry = signal_or(
    mu_oversold(rsi, level=30, width=5),
    mu_turns_positive(prev_hist, curr_hist, width=0.1)
)

# Only act on strong signals
if signal_strength(buy, min_strength=0.6) > 0:
    execute_trade()
```

## Cross-Language Naming Conventions

| Concept | Python | Go | TypeScript | Zig | Rust |
|---------|--------|----|------------|-----|------|
| Above threshold | `mu_above` | `MuAbove` | `muAbove` | `muAbove` | `mu_above` |
| Crosses above | `mu_crosses_above` | `MuCrossesAbove` | `muCrossesAbove` | `muCrossesAbove` | `mu_crosses_above` |
| Line crosses above | `mu_line_crosses_above` | `MuLineCrossesAbove` | `muLineCrossesAbove` | `muLineCrossesAbove` | `mu_line_crosses_above` |
| Above band | `mu_above_band` | `MuAboveBand` | `muAboveBand` | `muAboveBand` | `mu_above_band` |
| Between bands | `mu_between_bands` | `MuBetweenBands` | `muBetweenBands` | `muBetweenBands` | `mu_between_bands` |
| Turns positive | `mu_turns_positive` | `MuTurnsPositive` | `muTurnsPositive` | `muTurnsPositive` | `mu_turns_positive` |
| Signal AND | `signal_and` | `SignalAnd` | `signalAnd` | `signalAnd` | `signal_and` |
| Signal OR | `signal_or` | `SignalOr` | `signalOr` | `signalOr` | `signal_or` |
| Signal NOT | `signal_not` | `SignalNot` | `signalNot` | `signalNot` | `signal_not` |
| Signal strength | `signal_strength` | `SignalStrength` | `signalStrength` | `signalStrength` | `signal_strength` |
| File naming | `snake_case.py` | `lowercase.go` | `kebab-case.ts` | `snake_case.zig` | `snake_case.rs` |

## Cross-Language Import Patterns

| Language | Signals imports fuzzy as |
|----------|--------------------------|
| Python | `from ..fuzzy import MembershipShape, mu_greater, mu_less, t_product` (relative parent) |
| Go | `import "zpano/fuzzy"` then `fuzzy.MuGreater(...)` |
| TypeScript | `import { MembershipShape, muGreater, muLess, tProduct } from '../fuzzy/index.ts';` |
| Zig | `const fuzzy = @import("fuzzy");` then `fuzzy.membership.muGreater(...)`, `fuzzy.operators.tProduct(...)` |
| Rust | `use crate::fuzzy::{MembershipShape, mu_greater, mu_less, t_product};` |

Zig uses a single `fuzzy` barrel module (`src/fuzzy/fuzzy.zig`) that re-exports `membership`, `operators`, and `defuzzify` as sub-namespaces. Signal files create local aliases: `const membership = fuzzy.membership;`

## Cross-Language Variadic vs Slice

The `signal_and` / `t_product_all` function has different signatures per language:

| Language | Signature | Call syntax |
|----------|-----------|-------------|
| Python | `signal_and(*signals: float)` | `signal_and(a, b, c)` |
| Go | `SignalAnd(signals ...float64)` | `SignalAnd(a, b, c)` |
| TypeScript | `signalAnd(...signals: number[])` | `signalAnd(a, b, c)` |
| Zig | `signalAnd(signals: []const f64)` | `signalAnd(&.{a, b, c})` |
| Rust | `signal_and(signals: &[f64])` | `signal_and(&[a, b, c])` |

Python, Go, and TypeScript use variadic parameters for ergonomic call syntax. Zig and Rust use slice parameters because variadic functions are not idiomatic in those languages.

Empty input returns 1.0 (identity element of product t-norm).

## Build System Wiring

### Zig (`build.zig`)

Each signal file is registered as a separate build.zig module, all importing the single `fuzzy` module:

```
b.addModule("fuzzy", ...)              — src/fuzzy/fuzzy.zig (barrel, no deps)
b.addModule("sig_threshold", ...)      — src/signals/threshold.zig (imports: fuzzy)
b.addModule("sig_crossover", ...)      — src/signals/crossover.zig (imports: fuzzy)
b.addModule("sig_band", ...)           — src/signals/band.zig (imports: fuzzy)
b.addModule("sig_histogram", ...)      — src/signals/histogram.zig (imports: fuzzy)
b.addModule("sig_compose", ...)        — src/signals/compose.zig (imports: fuzzy)
```

Test modules mirror the library modules with `b.createModule()` + `b.addTest(.{ .root_module = mod })`.

### Go

Single package `signals` importing `"zpano/fuzzy"`. All functions are exported (PascalCase).

### TypeScript

Barrel `index.ts` re-exports all signal functions. Files import from `'../fuzzy/index.ts'`.

### Rust

Module `signals` with `mod.rs` declaring `pub mod threshold; pub mod crossover; ...` and `pub use` re-exports. Files import via `use crate::fuzzy::{...};`.

## Testing

### Test Organization

| Language | Style | File(s) | Approximate Count |
|----------|-------|---------|-------------------|
| Python | Separate `test_*.py` per source file | 5 test files | ~60 tests |
| Go | Combined `signals_test.go` | 1 test file | ~30 tests |
| TypeScript | Separate `*.spec.ts` per source file | 5 spec files | ~60 tests |
| Zig | Inline `test` blocks at bottom of source | 5 source files | ~45 tests |
| Rust | Inline `#[cfg(test)] mod tests` | 5 source files | ~48 tests |

### Test Patterns

All tests follow the same structure across languages:

1. **Well above/below**: extreme values produce mu near 1.0 or 0.0 (tolerance ~0.01)
2. **At threshold**: exact threshold produces mu = 0.5 (tolerance 1e-10)
3. **Zero width**: crisp behavior (1.0 or 0.0, with 0.5 at exact boundary)
4. **Monotonicity**: higher values produce higher mu_above membership
5. **Complement**: `mu_above(v) + mu_below(v) = 1.0` for any v
6. **Symmetry**: `crosses_below(a, b, t) == crosses_above(b, a, t)`
7. **Degenerate cases**: empty bands, zero spread, etc.
8. **Linear shape**: verify piecewise-linear membership at edges and center

### Test Tolerances

- Exact equality: `1e-10` (for values that should be exactly 0.5, 0.0, 1.0, or 0.25)
- Approximate: `0.01` (for "well above" / "well below" where exact value depends on sigmoid steepness)
- Cross-language parity: all five languages produce identical results to 13+ decimal places

## Design Decisions

### D1: Stateless Functions

Signals are pure functions with no internal state. This is a deliberate departure from the candlestick patterns engine (which maintains a ring buffer and criterion states).

**Rationale**: Indicator values are already computed by the indicators module. The signal layer just interprets them. State management belongs to the caller (strategy engine, backtest loop), not to the signal functions.

### D2: Crossover = Product of Two Memberships

`mu_crosses_above(prev, curr, t, w) = mu_below(prev, t, w) * mu_above(curr, t, w)`

**Rationale**: The product t-norm naturally enforces that BOTH conditions must hold:
- "Was below" — if prev was already above, this is 0 and the product is 0 (no false positives from one-sided membership)
- "Is above" — if curr is still below, this is 0

Alternative considered: `min(mu_below(prev), mu_above(curr))`. Rejected because min can't distinguish "one strong + one weak" from "both weak." Product captures the proportional confidence of each side.

### D3: Band Width from Spread

`mu_between_bands` uses `width = spread * 0.5` rather than a caller-supplied width.

**Rationale**: The band spread itself determines what "inside" means. Narrow bands should have sharp containment (you're either inside or outside a 1-point channel). Wide bands should have gradual transitions (a 50-point channel has fuzzy edges). Tying width to spread makes the function self-calibrating.

### D4: signal_strength is NOT alpha_cut

`signal_strength` preserves continuous values above the threshold. `alpha_cut` discretizes to integer levels.

**Rationale**: Signal composition chains should preserve continuous values as long as possible. Discretization is a final step for execution decisions, not an intermediate step in signal processing.

### D5: Default Width Values

- Threshold functions: width = 5.0 (moderate transition for typical 0-100 range indicators like RSI, Stochastic)
- Crossover functions: width = 0.0 (crisp by default — most crossover strategies want exact detection)
- Band/histogram functions: width = 0.0 (crisp by default — band edges are already dynamic)

**Rationale**: Thresholds benefit from fuzzy transitions because the exact level (70 vs 71) is inherently arbitrary. Crossovers and band touches are more structural events where fuzzy transitions are optional.

## Adding a New Signal Type

1. Create source file in each language's `signals/` directory following naming conventions
2. Import only from `fuzzy/` — never from `indicators/` or other modules
3. Keep functions pure and stateless
4. Accept `width` and `shape` parameters for membership function calls
5. Add tests covering: extreme values, threshold boundary, zero width (crisp), monotonicity, symmetry where applicable
6. Update barrel exports (`__init__.py`, `index.ts`, `mod.rs`)
7. For Zig: register as a new `b.addModule("sig_<name>", ...)` in `build.zig` with `fuzzy` import, plus test module + run artifact + test step dependency
8. Port to all five languages maintaining identical logic and results to 13+ decimal places
