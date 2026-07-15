---
name: streaming-kbn-architecture
description: Architecture, algorithms, and implementation reference for the zpano streaming Klein KBN compensated summation library. Load when implementing new streaming accumulators, porting across languages, or understanding the KBN-compensated numerical stability system.
---

# Streaming KBN Architecture

Architecture, algorithms, and implementation reference for the streaming KBN-compensated statistical accumulators in zpano. This package provides four classes for streaming O(1) computation of mean, variance, skewness, kurtosis, and linear regression, all backed by Klein second-order Kahan-Babuška-Neumaier (KBN) double-compensated summation.

## Module Dependencies

```
streaming_kbn/              (standalone — zero dependencies on other zpano modules)
    |
    v
(consumers)                 (indicators module, icalc CLI tool, arbitary callers)
```

The streaming_kbn package has **zero dependencies** on other zpano modules (`entities/`, `indicators/`, etc.). It is pure math operating on `f64` values.

## Domain: KBN-Compensated Accumulation

### Floating-Point Summation Problem

Adding floating-point numbers naively accumulates round-off error because each addition rounds the result to the available significand. Low-order bits of the smaller operand are lost whenever the sum becomes large relative to the addend.

**Naive sum** `s += x`: worst-case relative error grows as `O(ε n)`. The bound is proportional to the **condition number**:

```
cond = Σ|xᵢ| / |Σxᵢ|
```

A large condition number means the sum is intrinsically sensitive to round-off.

**Peters example** — `[1.0, 1e100, 1.0, -1e100]`:

| Method | Result |
|--------|--------|
| Exact | 2.0 |
| Naive / Kahan | 0.0 |
| **KBN / Klein KBN** | **2.0** |

Naive and standard Kahan both return 0.0 because the `1.0` additions are completely lost when `1e100` dominates the significand.

### Algorithm Progression

**Kahan (1965):** Single-level compensated summation with `c = (t - s) - y`. Reduces error to `O(ε + nε²)` but fails when sum and addend differ hugely.

**Kahan-Babuška-Neumaier (KBN, 1974):** Branches on which operand is larger — the term `(big - (big + small))` is exact via [2Sum](https://en.wikipedia.org/wiki/2Sum). The correction `c` accumulates losses and is applied as a final `s + c`.

**Klein second-order (2006):** Applies the same KBN trick to *the correction term itself*:

```
Level 1 (KBN):      t = s + x;  if |s| >= |x|:  c = (s - t) + x
                    else:                        c = (x - t) + s

Level 2 (Klein):    same correction applied to cs + c
```

The corrected value is `sum + cs + ccs`.

## Package Structure

### Documentation (`readme/streaming_kbn/`)
```
klein_kbn_accumulator.md            # Algorithm documentation
raw_moments_klein_kbn.md            # Algorithm + revert math
central_moments_klein_kbn.md        # Algorithm + Pébay formulas
linear_regression_klein_kbn.md      # Algorithm + cross-product revert math
```

### Python (`py/streaming_kbn/`)
```
__init__.py                         # Package init
klein_kbn_accumulator.py            # KleinKBNAccumulator class
raw_moments_klein_kbn.py            # RawMomentsKleinKBN class
central_moments_klein_kbn.py        # CentralMomentsKleinKBN class
linear_regression_klein_kbn.py      # LinearRegressionKleinKBN class
test_klein_kbn_accumulator.py       # 5 tests
test_raw_moments_klein_kbn.py       # 7 tests
test_central_moments_klein_kbn.py   # 8 tests
test_linear_regression_klein_kbn.py # 9 tests
```

### Go (`go/streamingkbn/`)
```
kleinkbnaccumulator.go              # KleinKBNAccumulator
rawmomentskleinkbn.go               # RawMomentsKleinKBN
centralmomentskleinkbn.go           # CentralMomentsKleinKBN
linearregressionkleinkbn.go         # LinearRegressionKleinKBN
*_test.go                           # Co-located test files (30 tests)
```

### TypeScript (`ts/streaming-kbn/`)
```
index.ts                            # Barrel re-exports
klein-kbn-accumulator.ts            # KleinKbnAccumulator
raw-moments-klein-kbn.ts            # RawMomentsKleinKbn
central-moments-klein-kbn.ts        # CentralMomentsKleinKbn
linear-regression-klein-kbn.ts      # LinearRegressionKleinKbn
*.spec.ts                           # Co-located spec files (30 tests)
```

### Zig (`zig/src/streaming_kbn/`)
```
klein_kbn_accumulator.zig           # KleinKbnAccumulator + inline tests
raw_moments_klein_kbn.zig           # RawMomentsKleinKbn + inline tests
central_moments_klein_kbn.zig       # CentralMomentsKleinKbn + inline tests
linear_regression_klein_kbn.zig     # LinearRegressionKleinKbn + inline tests
streaming_kbn.zig                   # Barrel re-export
```

### Rust (`rs/src/streaming_kbn/`)
```
mod.rs                              # Module root with pub use re-exports
klein_kbn_accumulator.rs            # KleinKbnAccumulator + inline tests
raw_moments_klein_kbn.rs            # RawMomentsKleinKbn + inline tests
central_moments_klein_kbn.rs        # CentralMomentsKleinKbn + inline tests
linear_regression_klein_kbn.rs      # LinearRegressionKleinKbn + inline tests
```

### Build Registration

- **Zig:** `build.zig` defines 4 library modules (`klein_kbn_accumulator`, `raw_moments_klein_kbn`, `central_moments_klein_kbn`, `linear_regression_klein_kbn`) + a barrel module (`streaming_kbn`). Test modules wired via `b.createModule()` + `b.addTest()`.
- **Rust:** `rs/src/lib.rs` requires `pub mod streaming_kbn;`.
- **TypeScript:** `ts/tsconfig.json` needs `"streaming-kbn/**/*.ts"` in `include`.
- **Go/Python:** No registration needed — package boundaries are directory-based.

## Class Reference

### 1. `KleinKbnAccumulator`

Klein second-order KBN compensated summation. Maintains `_sum + _cs + _ccs`.

#### Algorithm

```
update(x):
    s = _sum
    t = s + x
    if |s| >= |x|:  c = (s - t) + x
    else:           c = (x - t) + s
    _sum = t

    cs = _cs
    t = cs + c
    if |cs| >= |c|:  cc = (cs - t) + c
    else:            cc = (c - t) + cs
    _cs = t
    _ccs = cc

value():    return _sum + _cs + _ccs
set(x):     _sum = x, _cs = 0, _ccs = 0
reset():    set(0)
revert(x):  update(-x)
```

#### Cross-Language API

| Operation | Python | Go | TypeScript | Zig | Rust |
|-----------|--------|----|------------|-----|------|
| Constructor | `KleinKBNAccumulator()` | `&KleinKBNAccumulator{}` | `new KleinKbnAccumulator()` | `KleinKbnAccumulator{}` / `.init` | `KleinKbnAccumulator::new()` |
| Update | `update(x)` | `Update(x)` | `update(x)` | `update(x)` | `update(x)` |
| Revert | `revert(x)` | `Revert(x)` | `revert(x)` | `revert(x)` | `revert(x)` |
| Set | `set(x)` | `Set(x)` | `set(x)` | `set(x)` | `set(x)` |
| Reset | `reset()` | `Reset()` | `reset()` | `reset()` | `reset()` |
| Value (getter) | `.value` | `.Value()` | `.value` | `.value()` | `.value()` |

### 2. `RawMomentsKleinKBN`

Streaming mean, variance, skewness, kurtosis via raw power sums (x¹..x⁴) with KBN compensation.

#### State

| Variable | Accumulator | Purpose |
|----------|-------------|---------|
| `n` | `int`/`usize` | Sample count |
| `x1` | KleinKBN | Σx |
| `x2` | KleinKBN | Σx² |
| `x3` | KleinKBN | Σx³ |
| `x4` | KleinKBN | Σx⁴ |
| `mean` | KleinKBN | Welford running mean |
| `s` | KleinKBN | Welford sum of squared deviations |
| `ddof` | int/`usize` | Delta degrees of freedom for variance |
| `bias` | bool | Population (true) vs bias-corrected (false) moments |
| `fisher` | bool | Excess kurtosis (true) vs raw kurtosis (false) |

#### Central Moment Conversion

```
A = Σx / n
B = Σx²/n − A²
R = √B
C = Σx³/n − A³ − 3·A·B
D = Σx⁴/n − A⁴ − 6·B·A² − 4·C·A

skewness (bias=true):  g1 = C / R³
skewness (bias=false): G1 = g1 · √(n·(n−1)) / (n−2)

kurtosis (bias=true, fisher=true):  g2 = n·D/B² − 3
kurtosis (bias=false, fisher=true):
  G2 = ((n²−1)·(n·D/B²) − 3·(n−1)²) / ((n−2)·(n−3))
```

Guard clauses: skewness requires `n ≥ 3` and `B > 1e-14`; kurtosis requires `n ≥ 4` and `B > 1e-14`.

#### FIFO Revert (Order-Independence)

`RawMomentsKleinKBN.revert(x)` works for **any** sample regardless of insertion order because:

1. **Power sums are linear and commutative:** removing `x_k^p` by `update(-x_k^p)` is exact — addition is order-independent.
2. **Welford variance revert depends only on current state and `x_k`:**

   ```
   n'      = n − 1
   x̄'     = x̄ − (x_k − x̄) / n'
   S_xx'  = S_xx − (n/n') · (x_k − x̄)²
   ```

   These formulas hold for any `x_k` in the set, regardless of insertion position.
3. **KBN compensation is preserved** because `revert` calls `update(-x)`, which uses the same KBN branch logic — the compensation terms remain intact.

This is what makes `RawMomentsKleinKBN` suitable for FIFO rolling windows (deque-based).

#### Cross-Language Notes

- **`variance` getter may reset `_s`:** If `_s.value < 0` (floating-point noise), the accumulator is reset to zero before returning NaN. In Rust this means `variance()` takes `&mut self`.
- **`standard_deviation`** is computed directly as `√(s/n)` to avoid calling `variance` (and its potential side-effect). Go/TS/Zig/Rust follow this pattern.

### 3. `CentralMomentsKleinKBN`

Streaming mean, variance, skewness, kurtosis via Pébay's central moment update with KBN compensation. Avoids the numerical cancellation inherent in raw power-sum conversion.

#### Pébay Forward Update

```
n_new = n_old + 1
δ     = x − m₁
δₙ    = δ / n_new
term  = δ · δₙ · n_old

m₁   += δₙ
m₂   += term
m₃   += term · δₙ · (n_new − 2)  −  3·δₙ·m₂
m₄   += term · δₙ² · (n_new² − 3·n_new + 3)  +  6·δₙ²·m₂  −  4·δₙ·m₃
```

#### Inverse Pébay Revert (LIFO only)

```
m₁_old = (n_new · m₁_new − x) / n_old
δ      = x − m₁_old
δₙ     = δ / n_new
term   = δ · δₙ · n_old

m₂_old = m₂_new − term
m₃_old = m₃_new − (term·δₙ·(n_new−2) − 3·δₙ·m₂_old)
m₄_old = m₄_new − (term·δₙ²·(n_new²−3·n_new+3) + 6·δₙ²·m₂_old − 4·δₙ·m₃_old)
```

After computing restored values, each accumulator is set via `set(value)`, which resets KBN compensation terms to zero. This means **only the most recent sample can be reverted** (LIFO stack) — not suitable for FIFO rolling windows.

#### Raw vs Central Moments

| Aspect | RawMomentsKleinKBN | CentralMomentsKleinKBN |
|--------|--------------------|------------------------|
| Forward accuracy | ⚠️ Good with KBN, but large mean erodes precision | ✅ Best (no raw-sum cancellation) |
| Revert | ✅ FIFO via `update(-x)` | ⚠️ LIFO only, compensation reset |
| Rolling window | ✅ Natural (deque-based) | ❌ Not recommended |
| Guard-clause return value | NaN | NaN |

### 4. `LinearRegressionKleinKBN`

Streaming simple linear regression (`y = β₁x + β₀`) with KBN-compensated accumulation.

#### State

| Variable | Type | Purpose |
|----------|------|---------|
| `n` | int | Sample count |
| `xMoments` | RawMomentsKleinKBN(ddof=0) | `x̄`, `S_xx` |
| `yMoments` | RawMomentsKleinKBN(ddof=0) | `ȳ`, `S_yy` |
| `sXY` | KleinKbnAccumulator | Cross-product sum `S_xy` |

#### Welford Cross-Product Update

```
n_old = n
n += 1
term = (x̄ − x) · (ȳ − y) · n_old / n
s_xy += term
x_moments.update(x)    # updates x̄, S_xx
y_moments.update(y)    # updates ȳ, S_yy
```

#### Revert Inverse Formula

```
x_moments.revert(x)    # restores x̄₀
y_moments.revert(y)    # restores ȳ₀
n -= 1
term = (x̄₀ − x) · (ȳ₀ − y) · n / (n+1)
s_xy -= term
```

FIFO revert works because:
1. `RawMomentsKleinKBN.revert()` is order-independent (restores `x̄₀`, `S_xx` for any removed sample).
2. The cross-product revert formula uses only restored means and the removed sample — quantities well-defined regardless of insertion order.
3. KBN compensation is preserved through KBN subtraction (`update(-term)`).

#### Property Formulas

```
β₁ = s_xy / S_xx        (slope, guard: n < 2 or S_xx = 0 → NaN)
β₀ = ȳ − β₁ · x̄        (intercept, NaN propagates from slope)
r  = s_xy / √(S_xx·S_yy)  (correlation, guard: n < 2 or σ_x·σ_y = 0 → NaN)
```

## Cross-Language Conventions

### NaN Returns for Impossible Computations

All languages return NaN (not null, None, or 0) for impossible-computation cases:
- `n ≤ ddof` for variance (all 4 classes)
- `n < 3` for skewness, `n < 4` for kurtosis
- `n < 2` or zero variance for slope/correlation/intercept
- Non-positive `B` (central moment guard) for skewness/kurtosis

### Filename Patterns

| Language | Convention | Example |
|----------|-----------|---------|
| Python | `snake_case.py` | `raw_moments_klein_kbn.py` |
| Go | `flatcase.go` | `rawmomentskleinkbn.go` |
| TypeScript | `kebab-case.ts` | `raw-moments-klein-kbn.ts` |
| Zig | `snake_case.zig` | `raw_moments_klein_kbn.zig` |
| Rust | `snake_case.rs` | `raw_moments_klein_kbn.rs` |

### Naming

| Concept | Python | Go | TypeScript | Zig | Rust |
|---------|--------|----|------------|-----|------|
| Class name | `KleinKBNAccumulator` | `KleinKBNAccumulator` | `KleinKbnAccumulator` | `KleinKbnAccumulator` | `KleinKbnAccumulator` |
| Class name | `RawMomentsKleinKBN` | `RawMomentsKleinKBN` | `RawMomentsKleinKbn` | `RawMomentsKleinKbn` | `RawMomentsKleinKbn` |
| Class name | `CentralMomentsKleinKBN` | `CentralMomentsKleinKBN` | `CentralMomentsKleinKbn` | `CentralMomentsKleinKbn` | `CentralMomentsKleinKbn` |
| Class name | `LinearRegressionKleinKBN` | `LinearRegressionKleinKBN` | `LinearRegressionKleinKbn` | `LinearRegressionKleinKbn` | `LinearRegressionKleinKbn` |
| Constructor | `(ddof=1, bias=True, fisher=True)` | `New*(ddof, bias, fisher)` | `(ddof=1, bias=true, fisher=true)` | `.init(ddof, bias, fisher)` | `::new(ddof, bias, fisher)` |
| Default ddof | 1 | varies by call site | 1 | varies by call site | varies by call site |

### Test Conventions

| Language | Framework | Location | Key assertion |
|----------|-----------|----------|---------------|
| Python | `unittest` | Separate `test_*.py` | `assertAlmostEqual(x, y, places=13)` |
| Go | `testing` | Same package `*_test.go` | `almostEqual(a, b, 1e-14)` |
| TypeScript | Jasmine | Same dir `*.spec.ts` | `toBeCloseTo(x, 13)` |
| Zig | built-in | Inline at bottom of source | `try testing.expect(almostEqual(...))` |
| Rust | built-in | Inline `#[cfg(test)] mod tests` | `assert!((a - b).abs() < eps)` |

### Default Parameter Values (Moments Classes)

- **Python:** `ddof=1, bias=True, fisher=True` (default constructor; tests override to `ddof=0`)
- **Go/TS/Zig/Rust:** No defaults — parameters always explicit at call site. Tests use `ddof=0, bias=true, fisher=true`.

## Test Data & Expected Values

### Bacon Data (24-element return series)

```
[ 0.003,  0.026,  0.011, -0.010,  0.015,  0.025,  0.016,  0.067,
 -0.014,  0.040, -0.005,  0.081,  0.040, -0.037, -0.061,  0.017,
 -0.049, -0.022,  0.070,  0.058, -0.065,  0.024, -0.005, -0.009 ]
```

### RawMomentsKleinKbn Expected Values (ddof=0, bias=true, fisher=true)

| Metric | Expected |
|--------|----------|
| Mean | 0.009000000000000001 |
| Variance | 0.0014989166666666666 |
| Skewness | -0.08256245520856798 |
| Kurtosis | -0.5675462058921257 |

### RawMomentsKleinKbn Expected Values (bias=false, fisher=true)

| Metric | Expected |
|--------|----------|
| Skewness | -0.08817174934967527 |
| Kurtosis | -0.40766032118608714 |

### CentralMomentsKleinKbn Expected Values (ddof=0, bias=true, fisher=true)

| Metric | Expected |
|--------|----------|
| Mean | 0.009000000000000001 |
| Variance | 0.0014989166666666668 |
| Skewness | -0.08256245520856803 |
| Kurtosis | -0.5675462058921261 |

### CentralMomentsKleinKbn Expected Values (bias=false, fisher=true)

| Metric | Expected |
|--------|----------|
| Skewness | -0.08817174934967532 |
| Kurtosis | -0.4076603211860876 |

Note: RawMoments and CentralMoments produce slightly different values due to different accumulation paths (raw power sums vs Pébay). Both match scipy within tolerance.

## References

- Higham, N. J. (1993). "The accuracy of floating point summation". *SIAM Journal on Scientific Computing*, 14(4), 783–799.
- Kahan, W. (1965). "Further remarks on reducing truncation errors". *Communications of the ACM*, 8(1), 40.
- Neumaier, A. (1974). "Rundungsfehleranalyse einiger Verfahren zur Summation endlicher Summen". *Zeitschrift für Angewandte Mathematik und Mechanik*, 54(1), 39–51.
- Klein, A. (2006). "A generalized Kahan–Babuška-Summation-Algorithm". *Computing*, 76(3–4), 279–293.
- Pébay, P. (2008). "Formulas for robust, one-pass parallel computation of covariances and arbitrary-order statistical moments". *Sandia Report SAND2008-6212*.
- Welford, B. P. (1962). "Note on a method for calculating corrected sums of squares and products". *Technometrics*, 4(3), 419–420.
- Cook, J. D. [Skewness and kurtosis](https://www.johndcook.com/skewness_kurtosis.html).
- Cook, J. D. [Running regression](https://www.johndcook.com/running_regression.html).
- Kuiperzone. [Compensated-Accumulators](https://github.com/kuiperzone/Compensated-Accumulators).
- Wikipedia. [Kahan summation algorithm](https://en.wikipedia.org/wiki/Kahan_summation_algorithm).
- Wikipedia. [2Sum](https://en.wikipedia.org/wiki/2Sum).
- NumPy issue #8786 — [Badly conditioned sum](https://github.com/numpy/numpy/issues/8786).
- CPython `math.fsum` implementation — [Peters' example](https://github.com/python/cpython/blob/main/Modules/mathmodule.c).
