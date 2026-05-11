# Fuzzy Architecture Skill

Architecture, design decisions, and implementation reference for the fuzzy logic layer in zpano. This package provides shared fuzzy primitives used by candlestick pattern recognition and (future) indicator-based signal generation.

## Fuzzy Logic Fundamentals

### Core Concepts

**Crisp logic** is binary: a condition is True (1) or False (0). A bar's body either *is* short or it *isn't*.

**Fuzzy logic** replaces binary truth with a **degree of membership** μ ∈ [0, 1]. A bar's body might be "0.85 short" — mostly short but not fully. This captures the continuous nature of financial data where boundaries between categories are inherently blurry.

### Terminology

| Term | Definition |
|------|-----------|
| **Membership function** (μ) | A function mapping a crisp value to [0, 1]. E.g., μ_short(body_size) → 0.85 |
| **Linguistic variable** | A variable described by fuzzy sets. E.g., "body size" has sets {short, medium, long} |
| **T-norm** | Fuzzy AND operator. Combines membership degrees. Must satisfy: commutativity, associativity, monotonicity, identity (T(a, 1) = a) |
| **S-norm (T-conorm)** | Fuzzy OR operator. Dual of t-norm |
| **Fuzzy negation** | Fuzzy NOT: `1 - μ` |
| **Alpha-cut (α-cut)** | Threshold that converts fuzzy output back to crisp: if μ ≥ α then True |
| **Defuzzification** | Converting a fuzzy result to a crisp output value |
| **Transition zone** | The range over which membership transitions from 0 to 1 (controlled by `width` parameter) |
| **Fuzz ratio** | Width of transition zone as a fraction of the criterion's running average |

### Why Fuzzy for Financial Patterns?

In crisp candlestick recognition:
- A body of size 0.49 (threshold 0.50) → "not short" → pattern missed entirely
- RSI at 69.99 (threshold 70) → "not overbought" → signal missed

With fuzzy logic:
- Body of 0.49 vs threshold 0.50 → μ ≈ 0.52 → pattern detected with 52% confidence
- RSI at 69.99 vs threshold 70 → μ ≈ 0.49 → signal detected with 49% confidence

The output becomes a continuous confidence score rather than a binary hit/miss.

## Package Structure

```
py/fuzzy/                          # Shared fuzzy logic primitives
    __init__.py                    # Exports: mu_less, mu_greater, mu_less_equal,
                                   #   mu_greater_equal, mu_near, mu_direction,
                                   #   t_product, t_min, t_lukasiewicz,
                                   #   s_probabilistic, s_max, f_not,
                                   #   t_product_all, t_min_all, alpha_cut
    membership.py                  # Membership functions
    operators.py                   # T-norms, S-norms, negation
    defuzzify.py                   # Alpha-cut defuzzification
    test_membership.py             # Unit tests for membership functions
    test_operators.py              # Unit tests for operators
    test_defuzzify.py              # Unit tests for defuzzification

py/signals/                        # Indicator-based fuzzy signal generation
    __init__.py                    # Exports all signal functions
    threshold.py                   # mu_above, mu_below, mu_overbought, mu_oversold
    crossover.py                   # mu_crosses_above/below, mu_line_crosses_above/below
    band.py                        # mu_above_band, mu_below_band, mu_between_bands
    histogram.py                   # mu_turns_positive, mu_turns_negative
    compose.py                     # signal_and, signal_or, signal_not, signal_strength
    test_threshold.py              # Tests per file
    test_crossover.py
    test_band.py
    test_histogram.py
    test_compose.py
```

Both packages are implemented in all five languages (Python, Go, TypeScript, Rust, Zig). See "Cross-Language Implementation Reference" below for per-language details.

The `py/fuzzy/` and `py/signals/` packages have **zero dependencies** on other zpano modules. They are standalone math libraries. `signals/` imports from `fuzzy/`; consumers (candlestick patterns, trading systems) import from both.

## Design Decisions

### D1: Membership Function Shape — Sigmoid (default) with Linear option

**Choice**: Sigmoid is the default shape. Linear (trapezoidal ramp) is available as an option.

**Rationale**:
- Sigmoid is smooth and differentiable everywhere — important if anyone wants gradients later
- Sigmoid has a natural probabilistic interpretation (logistic function)
- Linear is simpler and more interpretable but has discontinuous derivatives at the transition boundaries
- Both degrade to crisp step functions when width → 0

**Sigmoid parameterization**: For `mu_less(x, threshold, width)`:
```
μ = 1 / (1 + exp(k * (x - threshold)))
k = 12 / width
```

The constant 12 is chosen so that:
- At `threshold ± width/2`: μ ≈ 0.997 / 0.003 (effectively 0 and 1)
- At `threshold`: μ = 0.5 exactly (natural crossover point)
- The transition spans `width` from ~0 to ~1

**Linear parameterization**: For `mu_less(x, threshold, width)`:
```
μ = 1.0           if x ≤ threshold - width/2
μ = 0.0           if x ≥ threshold + width/2
μ = linear interp  otherwise
```

### D2: Fuzz Ratio — 0.2 (configurable)

**Choice**: `fuzz_ratio = 0.2` means the transition zone width is 20% of the criterion's running average.

**Rationale**:
- Too narrow (< 0.1) ≈ effectively crisp, defeats the purpose
- Too wide (> 0.5) = everything is "sort of" a pattern, too many weak signals
- 0.2 provides meaningful differentiation: a value 10% below threshold gets μ ≈ 0.85, a value 10% above gets μ ≈ 0.15

**Adaptive behavior**: Because width = fuzz_ratio × running_average, the transition zone automatically scales with:
- Price magnitude (a $500 stock has wider absolute zones than a $5 stock)
- Market volatility (volatile markets produce larger running averages → wider zones)

**Constructor parameter**: `CandlestickPatterns(fuzz_ratio=0.2, shape=MembershipShape.SIGMOID)`. The fuzz_ratio can be tuned per application.

### D3: T-Norm — Product (primary)

**Choice**: Product t-norm (`a * b`) as the primary fuzzy AND.

**Available t-norms**:

| T-norm | Formula | Behavior |
|--------|---------|----------|
| **Product** | `a * b` | All conditions contribute proportionally. Moderate — rewards strong satisfaction of all conditions |
| **Minimum** | `min(a, b)` | Dominated by weakest link. One bad condition kills the whole result |
| **Łukasiewicz** | `max(0, a + b - 1)` | Very strict. Both conditions must have high membership for nonzero output |

**Rationale for product**:
- Product rewards bars that strongly satisfy ALL conditions while giving partial credit
- It's multiplicative: 5 conditions each at μ=0.9 → confidence = 0.9⁵ = 0.59 (reasonable)
- 5 conditions each at μ=0.5 → confidence = 0.5⁵ = 0.03 (correctly weak)
- One condition at μ=0.1 with others at μ=1.0 → 0.1 (correctly dominated by the weak link)
- Minimum would give 0.1 in the last case too, but wouldn't distinguish between "one weak + four strong" vs "all at 0.1"

### D4: Direction — Fuzzy (continuous)

**Choice**: Candle direction is a continuous value ∈ [-1, +1], not binary.

**Implementation**: `mu_direction(o, c, body_avg, steepness=2.0)`:
```
direction = tanh(steepness * (c - o) / body_avg)
```

Where `body_avg` is typically from `self._avg(self._short_body, shift)` — the running average body size provides scale normalization.

**Behavior**:
- Large white body (c >> o): direction ≈ +1.0
- Large black body (c << o): direction ≈ -1.0
- Doji (c ≈ o): direction ≈ 0.0
- Tiny white body: direction ≈ +0.1 (barely bullish)

**Where direction applies** (3 categories):

**Category A — Fixed direction (36 patterns)**: Direction is defined by the pattern itself (e.g., hammer is always bullish, dark_cloud_cover is always bearish). The `is_black`/`is_white` preconditions become fuzzy membership degrees that feed into the confidence product, but the output sign is fixed.
```python
# Example: dark_cloud_cover (always bearish)
μ_black1 = f_not(mu_direction(...))  # fuzzy "is bearish"
confidence = t_product_all(μ_black1, μ_rb, ...)
return -confidence * 100  # sign always negative
```

**Category B — Direction from color (12 patterns)**: Direction comes from a candle's color. The fuzzy direction value directly scales the output.
```python
# Example: engulfing
direction = mu_direction(o2, c2, body_avg)  # ∈ [-1, +1]
confidence = t_product_all(μ_cond1, μ_cond2, ...)
return direction * confidence * 100  # sign from direction
```

Patterns: counterattack, doji_star, engulfing, harami, harami_cross, kicking, kicking_by_length, long_line, marubozu, rising_falling_three_methods.

**Category C — Branched direction (13 patterns)**: Separate bullish and bearish branches with different conditions. Both branches are evaluated; the stronger signal wins.
```python
# Example: abandoned_baby
μ_bull = mu_bullish(o1, c1, avg)  # fuzzy is_white for 1st candle
μ_bear = mu_bearish(o1, c1, avg)

conf_bull = t_product_all(μ_bull, μ_bull_c1, μ_bull_c2, ...)
conf_bear = t_product_all(μ_bear, μ_bear_c1, μ_bear_c2, ...)

bull_signal = conf_bull * 100
bear_signal = -conf_bear * 100

return bull_signal if abs(bull_signal) >= abs(bear_signal) else bear_signal
```

Patterns: abandoned_baby, belt_hold, breakaway, closing_marubozu, high_wave, separating_lines, short_line, spinning_top, tasuki_gap, three_inside, three_line_strike, three_outside, up_down_gap_side_by_side_white_lines, x_side_gap_three_methods.

### D5: Output Format — Continuous Float

**Choice**: All pattern methods return continuous `float` values in [-200.0, +200.0].

**Previous crisp values**: {-200, -100, -80, 0, +80, +100, +200}

**New fuzzy values**: Any float in the range. E.g., -87.3, +42.1, -156.8.

**Scale interpretation**:
- |value| ∈ [0, 100]: primary pattern confidence
- |value| ∈ (100, 200]: confirmed pattern (hikkake, hikkake_modified only)
- Sign: direction (positive = bullish, negative = bearish)

**±80 vs ±100 distinction dropped**: Engulfing/harami previously returned ±80 for edge-touching and ±100 for strict containment. In fuzzy mode, the enclosure membership degree naturally captures this — a barely-touching engulfing gets lower confidence than a fully-engulfing one. The 80/100 distinction is subsumed by the continuous confidence.

### D6: Alpha-Cut for Backward Compatibility

**Choice**: `alpha_cut()` is a standalone utility function, not embedded in patterns.

**Implementation**:
```python
def alpha_cut(value: float, alpha: float = 0.5, scale: float = 100.0) -> int:
    """Convert continuous fuzzy output to crisp discrete value.
    
    If abs(value) / scale >= alpha, returns the nearest crisp level
    with the original sign. Otherwise returns 0.
    """
```

**Usage in tests**: Tests apply `alpha_cut()` to fuzzy output and compare against TA-Lib's crisp values. This validates that the fuzzy implementation + alpha_cut reproduces TA-Lib behavior.

**Usage by consumers**: Applications that need crisp signals (e.g., backtesting engines) call `alpha_cut()` on the fuzzy output with their preferred threshold.

### D7: What Stays Crisp

Some operations are inherently binary and are not fuzzified:

- **`_enough()` warmup check**: Either we have enough bars or we don't. Returns 0.0 immediately if not enough data.
- **Gap existence**: `is_high_low_gap_up(h1, l2)` checks if `h1 < l2`. A gap either exists or doesn't (there's no "partial gap"). However, gap SIZE could be fuzzified in the future.
- **Hikkake stateful index tracking**: The bar counting and pattern index management is inherently discrete.

Note: Even `is_white`/`is_black` (color checks) ARE fuzzified — they become `mu_direction` feeding into the confidence product. The only truly crisp operations are structural checks (enough bars, gap existence, index tracking).

## Membership Function API

### `MembershipShape` Enum

```python
from enum import IntEnum

class MembershipShape(IntEnum):
    SIGMOID = 0  # Smooth logistic curve (default)
    LINEAR = 1   # Piecewise-linear ramp (trapezoidal/triangular)
```

Cross-language mapping: Python `IntEnum`, Go `int` with constants, TypeScript numeric `enum`, Zig `enum(u8)`, Rust `#[repr(u8)] enum`.

### `mu_less(x, threshold, width, shape=MembershipShape.SIGMOID) → float`

Degree to which `x < threshold`. μ ∈ [0, 1].

| x relative to threshold | μ (sigmoid) | μ (linear) |
|-------------------------|-------------|------------|
| x ≤ threshold - width/2 | ≈ 0.997 | 1.0 |
| x = threshold - width/4 | ≈ 0.95 | 0.75 |
| x = threshold | 0.5 | 0.5 |
| x = threshold + width/4 | ≈ 0.05 | 0.25 |
| x ≥ threshold + width/2 | ≈ 0.003 | 0.0 |

When `width = 0`: returns 1.0 if `x < threshold`, 0.5 if `x == threshold`, 0.0 if `x > threshold`.

### `mu_less_equal(x, threshold, width, shape=MembershipShape.SIGMOID) → float`

Same as `mu_less` but shifted so μ = 0.5 at `x = threshold`. In practice identical to `mu_less` for continuous values. The distinction matters conceptually and for documentation clarity.

### `mu_greater(x, threshold, width, shape=MembershipShape.SIGMOID) → float`

Complement of `mu_less`: `1 - mu_less(x, threshold, width, shape)`.

### `mu_greater_equal(x, threshold, width, shape=MembershipShape.SIGMOID) → float`

Complement of `mu_less_equal`.

### `mu_near(x, target, width, shape=MembershipShape.SIGMOID) → float`

Bell-shaped membership centered at `target`. μ = 1.0 at `x = target`, falling to ≈0 at `x = target ± width`.

- **Sigmoid shape**: Product of two sigmoids: `mu_greater_equal(x, target - width/2) * mu_less_equal(x, target + width/2)`.
- **Linear shape**: Triangular peak at target, base spanning `[target - width, target + width]`.

### `mu_direction(o, c, body_avg, steepness=2.0) → float`

Fuzzy candle direction ∈ [-1, +1]. Uses `tanh(steepness * (c - o) / body_avg)`.

When `body_avg ≤ 0`: returns +1.0 if `c ≥ o`, -1.0 otherwise (crisp fallback).

### Derived Helpers (on CandlestickPatterns)

```python
def _mu_less(self, value, cs, shift) -> float:
    avg = self._avg(cs, shift)
    width = self._fuzz_ratio * avg if avg > 0 else 0.0
    return mu_less(value, avg, width, self._shape)

def _mu_greater(self, value, cs, shift) -> float:
    avg = self._avg(cs, shift)
    width = self._fuzz_ratio * avg if avg > 0 else 0.0
    return mu_greater(value, avg, width, self._shape)

def _mu_near_value(self, value, target, cs, shift) -> float:
    avg = self._avg(cs, shift)
    width = self._fuzz_ratio * avg if avg > 0 else 0.0
    return mu_near(value, target, width, self._shape)

def _mu_bullish(self, o, c, shift=1) -> float:
    """Fuzzy degree of bullishness ∈ [0, 1]."""
    d = self._mu_direction_raw(o, c, shift)
    return max(0.0, d)  # positive part of direction

def _mu_bearish(self, o, c, shift=1) -> float:
    """Fuzzy degree of bearishness ∈ [0, 1]."""
    d = self._mu_direction_raw(o, c, shift)
    return max(0.0, -d)  # negative part of direction, made positive

def _mu_direction_raw(self, o, c, shift=1) -> float:
    """Raw fuzzy direction ∈ [-1, +1]."""
    avg = self._avg(self._short_body, shift)
    return mu_direction(o, c, avg, steepness=2.0)
```

## Operator API

### T-Norms (Fuzzy AND)

```python
t_product(a, b) → a * b              # Default choice
t_min(a, b) → min(a, b)              # Weakest-link behavior
t_lukasiewicz(a, b) → max(0, a+b-1)  # Very strict

# Variadic
t_product_all(*args) → reduce(product, args)
t_min_all(*args) → reduce(min, args)
```

### S-Norms (Fuzzy OR)

```python
s_probabilistic(a, b) → a + b - a*b  # Dual of product t-norm
s_max(a, b) → max(a, b)              # Dual of min t-norm
```

### Negation

```python
f_not(a) → 1 - a
```

## Defuzzification API

### `alpha_cut(value, alpha=0.5, scale=100.0) → int`

Converts continuous fuzzy output to crisp discrete output.

**Algorithm**:
1. Compute `confidence = abs(value) / scale`
2. If `confidence >= alpha`: return `sign(value) * round(confidence) * scale` (nearest crisp level)
3. Else: return 0

**Examples**:
- `alpha_cut(-87.3, 0.5, 100)` → `-100` (87.3% confidence ≥ 50% threshold → crisp -100)
- `alpha_cut(-32.1, 0.5, 100)` → `0` (32.1% confidence < 50% threshold → no signal)
- `alpha_cut(156.8, 0.5, 100)` → `200` (156.8% → round to nearest 100 multiple = 200)
- `alpha_cut(-87.3, 0.9, 100)` → `0` (87.3% < 90% threshold → filtered out)

## Pattern Conversion Guide

### Step-by-Step Conversion

1. **Identify category** (A, B, or C) from the return expressions
2. **Replace color checks** with fuzzy direction/bullish/bearish
3. **Replace comparison predicates** with membership functions
4. **Combine with t_product_all** instead of `and`
5. **Compute output** as `direction * confidence * scale`

### Predicate Mapping

| Crisp predicate | Fuzzy equivalent |
|----------------|-----------------|
| `real_body(o, c) < self._avg(self._short_body, 1)` | `self._mu_less(real_body(o, c), self._short_body, 1)` |
| `real_body(o, c) > self._avg(self._long_body, 2)` | `self._mu_greater(real_body(o, c), self._long_body, 2)` |
| `real_body(o, c) <= self._avg(self._doji_body, 1)` | `self._mu_less(real_body(o, c), self._doji_body, 1)` |
| `lower_shadow(...) > self._avg(self._long_shadow, 1)` | `self._mu_greater(lower_shadow(...), self._long_shadow, 1)` |
| `upper_shadow(...) < self._avg(self._very_short_shadow, 1)` | `self._mu_less(upper_shadow(...), self._very_short_shadow, 1)` |
| `abs(c2 - l1) <= eq` | `self._mu_near_value(c2, l1, self._equal, shift)` |
| `min(o, c) >= h1 - near_avg` | `self._mu_greater_equal(min(o, c), h1 - near_avg)` — width from near criterion |
| `c2 > c1 + rb1 * 0.5` (penetration) | `mu_greater(c2, c1 + rb1 * 0.5, fuzz_ratio * rb1 * 0.5)` |
| `is_white(o, c)` | `self._mu_bullish(o, c, shift)` |
| `is_black(o, c)` | `self._mu_bearish(o, c, shift)` |
| `is_high_low_gap_up(h1, l2)` | **Stays crisp**: `1.0 if h1 < l2 else 0.0` |
| `is_real_body_gap_up(...)` | **Stays crisp**: `1.0 if gap else 0.0` |
| `o2 < l1` (gap down open) | **Stays crisp** or fuzzified: `mu_less(o2, l1, width)` |
| `max(o2, c2) < max(o1, c1)` (strict containment) | `mu_less(max(o2, c2), max(o1, c1), width)` |

### Special Cases

**Penetration checks**: Constants like `0.5` in `c2 > c1 + rb1 * 0.5` are fuzzified with `width = fuzz_ratio * rb1 * 0.5`. The threshold itself is `c1 + rb1 * 0.5`.

**Engulfing/Harami ±80/±100**: The strict/edge-touching distinction is dropped. The continuous containment membership naturally encodes how fully the second body is contained within the first. A barely-touching harami gets lower confidence than a fully-contained one.

**Hikkake ±200 confirmation**: The confirmation output becomes `direction * (100 + confirmation_confidence * 100)`, ranging up to ±200 for fully confirmed patterns.

**Multi-branch patterns**: Both bullish and bearish branches are evaluated. Each produces its own confidence. The branch with higher absolute confidence wins. This handles the case where a bar is ambiguously colored — both branches get partial scores.

## Testing Strategy

### Backward Compatibility Tests

All existing test data (TA-Lib reference + euronext + nyse, ~60 cases per pattern) continues to work. The test harness applies `alpha_cut(result, 0.5, 100)` to convert fuzzy output to crisp, then compares against the expected crisp value.

**Key invariant**: `alpha_cut(fuzzy_result, 0.5) == talib_crisp_result` for all existing test cases.

### Fuzzy-Specific Tests

1. **Near-miss cases**: Bars that barely fail a crisp condition should produce small nonzero fuzzy confidence
2. **Perfect matches**: Bars that strongly satisfy all conditions should produce confidence near 100
3. **Monotonicity**: As a bar more strongly satisfies a condition, confidence should increase
4. **Width=0 degrades to crisp**: With `fuzz_ratio=0`, fuzzy output should match crisp output exactly (modulo the 0.5 membership at exact thresholds)

## Signals Module

The `signals/` package provides fuzzy signal primitives for indicator-based signal generation. It imports from `fuzzy/` and has **zero other dependencies**. Implemented in all five languages.

### Signal Functions

Five files, each with focused responsibility:

**threshold** — Degree to which a value is above/below a level, with overbought/oversold convenience wrappers:
- `mu_above(value, threshold, width, shape)` — degree value > threshold
- `mu_below(value, threshold, width, shape)` — degree value < threshold
- `mu_overbought(value, threshold, width, shape)` — alias for mu_above (semantic clarity)
- `mu_oversold(value, threshold, width, shape)` — alias for mu_below (semantic clarity)

**crossover** — Degree to which a value or line pair crosses a level between two time steps:
- `mu_crosses_above(prev, curr, threshold, width, shape)` — product of "was below" × "is now above"
- `mu_crosses_below(prev, curr, threshold, width, shape)` — product of "was above" × "is now below"
- `mu_line_crosses_above(prev_a, curr_a, prev_b, curr_b, width, shape)` — line A crosses above line B
- `mu_line_crosses_below(prev_a, curr_a, prev_b, curr_b, width, shape)` — line A crosses below line B

**band** — Degree to which a value is outside or inside a band:
- `mu_above_band(value, upper, width, shape)` — degree value > upper band
- `mu_below_band(value, lower, width, shape)` — degree value < lower band
- `mu_between_bands(value, lower, upper, width, shape)` — degree value is between bands; uses `width = spread * 0.5` for both edges

**histogram** — Degree to which a histogram value changes sign:
- `mu_turns_positive(prev, curr, width, shape)` — product of "was negative" × "is now positive"
- `mu_turns_negative(prev, curr, width, shape)` — product of "was positive" × "is now negative"

**compose** — Combine and transform fuzzy signals:
- `signal_and(signals...)` — product t-norm of all signals (variadic in Python/Go/TS; slice in Rust/Zig)
- `signal_or(a, b)` — probabilistic s-norm: `a + b - a * b`
- `signal_not(a)` — fuzzy negation: `1 - a`
- `signal_strength(signal, alpha)` — soft alpha-cut preserving continuous values (returns 0.0 if signal < alpha, otherwise returns signal unchanged)

### Signal Composition Example

```python
# "Buy when RSI is oversold AND MACD crosses up AND price near support"
mu_rsi = mu_below(rsi, 30, width=5)
mu_macd = mu_crosses_above(prev_macd, curr_macd, threshold=0.0, width=0.5)
mu_support = mu_near(price, support_level, width=price * 0.01)
buy_confidence = signal_and(mu_rsi, mu_macd, mu_support)
```

### Key Design Decisions for Signals

**D8: Crossover = product of two memberships**. `mu_crosses_above(prev, curr, threshold, width)` = `mu_below(prev, threshold, width) * mu_above(curr, threshold, width)`. This naturally produces 0 when either condition fails, and peaks when both "was clearly below" and "is clearly above" are strong.

**D9: mu_between_bands width derivation**. Uses `width = spread * 0.5` where `spread = upper - lower`. The caller-supplied `width` parameter is ignored for the internal edge calculations — the band spread itself determines the transition zone. When spread ≤ 0, returns 0.0.

**D10: signal_strength is NOT alpha_cut**. `signal_strength` preserves the continuous fuzzy value (returns the original signal if it passes the alpha threshold). `alpha_cut` discretizes to crisp integer levels. They serve different purposes.

**D11: signal_and variadic vs slice**. Python, Go, and TypeScript use variadic parameters for ergonomic call syntax. Rust and Zig use slice parameters (`&[f64]`, `[]const f64`) because variadic functions are not idiomatic in those languages. Empty input returns 1.0 (identity element of product t-norm). Single input returns that value unchanged.

## Cross-Language Implementation Reference

### Package / Module Structure

```
py/fuzzy/                          # Python package with __init__.py barrel
    membership.py, operators.py, defuzzify.py
    test_membership.py, test_operators.py, test_defuzzify.py
py/signals/
    threshold.py, crossover.py, band.py, histogram.py, compose.py
    test_threshold.py, test_crossover.py, test_band.py, test_histogram.py, test_compose.py

go/fuzzy/                          # Go package "fuzzy"
    membership.go, operators.go, defuzzify.go
    membership_test.go, operators_test.go, defuzzify_test.go
go/signals/                        # Go package "signals"
    doc.go, threshold.go, crossover.go, band.go, histogram.go, compose.go
    signals_test.go                # single combined test file

ts/fuzzy/                          # TypeScript with index.ts barrel
    index.ts, membership.ts, operators.ts, defuzzify.ts
    membership.spec.ts, operators.spec.ts, defuzzify.spec.ts
ts/signals/
    index.ts, threshold.ts, crossover.ts, band.ts, histogram.ts, compose.ts
    threshold.spec.ts, crossover.spec.ts, band.spec.ts, histogram.spec.ts, compose.spec.ts

rs/src/fuzzy/                      # Rust module with mod.rs re-exports
    mod.rs, membership.rs, operators.rs, defuzzify.rs
rs/src/signals/
    mod.rs, threshold.rs, crossover.rs, band.rs, histogram.rs, compose.rs
    # Tests inline in each .rs file (#[cfg(test)] mod tests)

zig/src/fuzzy/                     # Zig — per-file build.zig modules
    membership.zig, operators.zig, defuzzify.zig
zig/src/signals/
    threshold.zig, crossover.zig, band.zig, histogram.zig, compose.zig
    # Tests inline at bottom of each .zig file
```

### Naming Conventions

| Concept | Python | Go | TypeScript | Rust | Zig |
|---------|--------|----|------------|------|-----|
| mu_above | `mu_above` | `MuAbove` | `muAbove` | `mu_above` | `muAbove` |
| signal_and | `signal_and` | `SignalAnd` | `signalAnd` | `signal_and` | `signalAnd` |
| MembershipShape | `MembershipShape(IntEnum)` | `MembershipShape int` | `enum MembershipShape` | `#[repr(u8)] enum MembershipShape` | `enum(u8) { sigmoid = 0, linear = 1 }` |
| SIGMOID member | `SIGMOID = 0` | `Sigmoid MembershipShape = 0` | `SIGMOID = 0` | `Sigmoid = 0` | `sigmoid = 0` |
| File naming | `snake_case.py` | `lowercase.go` | `kebab-case.ts` | `snake_case.rs` | `snake_case.zig` |

### Import Patterns (signals → fuzzy)

| Language | Import |
|----------|--------|
| Python | `from ..fuzzy import MembershipShape, mu_greater, mu_less` (relative parent) |
| Go | `import "zpano/fuzzy"` → `fuzzy.MuGreater(...)` |
| TypeScript | `import { MembershipShape, muGreater } from '../fuzzy/index.ts';` |
| Rust | `use crate::fuzzy::{MembershipShape, mu_greater, mu_less};` |
| Zig | `const membership = @import("membership");` — imports build.zig module names, NOT `@import("../fuzzy/...")` |

**Zig note**: Zig has no unified `fuzzy` module. Each file (`membership`, `operators`, `defuzzify`) is a separate build.zig module. Signals modules import them individually: `threshold.zig` imports `membership`; `crossover.zig` imports `membership`; `band.zig` imports `membership`; `histogram.zig` imports `membership`; `compose.zig` imports `operators`.

### Test Organization

| Language | Style | Details |
|----------|-------|---------|
| Python | Separate test files | `test_*.py` per source file, `unittest.TestCase` subclasses, `assertAlmostEqual(places=13)` |
| Go | Combined test file for signals | `signals_test.go` covers all signal functions; fuzzy has per-file test files. `t.Run()` subtests |
| TypeScript | Separate spec files | `*.spec.ts` per source file, Jasmine `describe`/`it`, `toBeCloseTo(expected, 13)` |
| Rust | Inline tests | `#[cfg(test)] mod tests` at bottom of each `.rs` file. Each module defines local `almost_equal` helper |
| Zig | Inline tests | `test "name" { }` blocks at bottom of each `.zig` file. Each file defines local `almostEqual` helper |

### Test Counts

| Language | Fuzzy Tests | Signals Tests | Total |
|----------|------------|---------------|-------|
| Python | 86 | 75 | 161 |
| Go | ~40 | ~30 | ~70 |
| TypeScript | 160 combined | — | 160 |
| Rust | 85 | 48 | 133 |
| Zig | 40 | 93 | 133 |

### Zig Build System Wiring

Zig requires explicit module registration in `build.zig`:

**Library modules** (for production imports):
```
b.addModule("membership", ...) — src/fuzzy/membership.zig (no deps)
b.addModule("operators", ...)  — src/fuzzy/operators.zig (no deps)
b.addModule("defuzzify", ...)  — src/fuzzy/defuzzify.zig (imports: operators)
b.addModule("sig_threshold", ...) — src/signals/threshold.zig (imports: membership)
b.addModule("sig_crossover", ...) — src/signals/crossover.zig (imports: membership)
b.addModule("sig_band", ...)      — src/signals/band.zig (imports: membership)
b.addModule("sig_histogram", ...) — src/signals/histogram.zig (imports: membership)
b.addModule("sig_compose", ...)   — src/signals/compose.zig (imports: operators)
```

**Test modules** require: `b.createModule()` → `b.addTest(.{ .root_module = mod })` → `b.addRunArtifact()` → `test_step.dependOn()`. Forgetting `addRunArtifact` + `dependOn` causes "unused local constant" errors and tests silently don't run.

### Exp Clamp Constant

All languages clamp the sigmoid exponent to `[-500, 500]` to avoid floating-point overflow in `exp()`. This is implemented in `membership` and produces deterministic 0.0 or 1.0 at extreme inputs.

### Width = 0 Behavior

All membership functions degrade to crisp step functions when `width = 0`:
- `mu_less(x, t, 0)` → 1.0 if x < t, 0.5 if x == t, 0.0 if x > t
- `mu_near(x, t, 0)` → 1.0 if x == t, 0.0 otherwise
- Consistent across all five languages

## References

### Fuzzy Logic Libraries Reviewed

Three libraries were studied during the design phase:

1. **LFLL** (C++, header-only): 17 membership function shapes, 7 t-norms, 7 s-norms, compile-time template unrolling. Inspired: comprehensive operator set, pre-computed constants.
2. **eFLL** (C++, Arduino): Trapezoidal-only MFs, geometric defuzzification (exact centroids). Inspired: simplicity, analytical approach.
3. **simpful** (Python): Natural-language rule parser, 9 MF shapes, AutoTriangle partitions. Inspired: dual function/point-based sets, pluggable aggregation.

### Key Academic Concepts

- **Zadeh (1965)**: Introduced fuzzy sets. Min/max t-norm/s-norm.
- **Product t-norm**: Also called "probabilistic AND." Treats memberships as independent probabilities.
- **Sigmoid membership**: Logistic function `1/(1+exp(-k*(x-c)))`. Smooth, differentiable, natural probabilistic interpretation.
- **Alpha-cut**: Fundamental defuzzification method. The α-cut of a fuzzy set A is the crisp set {x | μ_A(x) ≥ α}.
