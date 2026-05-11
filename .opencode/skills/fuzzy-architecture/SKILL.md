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

py/signals/                        # Future: indicator-based signal generation
    __init__.py                    # Placeholder (scaffolded)
```

The `py/fuzzy/` package has **zero dependencies** on other zpano modules. It is a standalone math library. Consumers (candlestick patterns, future signals) import from it.

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

## Future: Indicator Signals

The `py/fuzzy/` package is designed to serve both candlestick patterns and future indicator-based signals.

### Planned Signal Types

| Signal | Crisp | Fuzzy |
|--------|-------|-------|
| Threshold crossing | `rsi > 70` | `mu_greater(rsi, 70, width=5)` |
| Moving average crossover | `slow > fast and prev_slow <= prev_fast` | `mu_crosses_above(prev_diff, curr_diff, width)` |
| Bollinger band touch | `price > upper_band` | `mu_greater(price, upper_band, width)` |
| MACD histogram sign change | `hist > 0 and prev_hist <= 0` | Fuzzy crossover of zero line |

### Signal Composition

Multiple fuzzy signals combine using t-norms:
```python
# "Buy when RSI is oversold AND MACD crosses up AND price near support"
μ_rsi = mu_less(rsi, 30, width=5)
μ_macd = mu_crosses_above(prev_macd, curr_macd, width=0.5)
μ_support = mu_near(price, support_level, width=price * 0.01)
buy_confidence = t_product_all(μ_rsi, μ_macd, μ_support)
```

This will be implemented in `py/signals/` when the indicator signal work begins.

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
