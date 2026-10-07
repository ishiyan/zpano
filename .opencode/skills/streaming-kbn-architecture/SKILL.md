---
name: streaming-kbn-architecture
description: Architecture, algorithms, and implementation reference for the zpano streaming Klein KBN compensated summation library. Load when implementing new streaming accumulators, porting across languages, or understanding the KBN-compensated numerical stability system.
---

# Streaming KBN Architecture

Architecture, algorithms, and implementation reference for the streaming KBN-compensated statistical accumulators in zpano. This package provides five classes for streaming (one-pass, O(1) per sample) computation of sums, mean, variance, skewness, kurtosis, and linear regression, all backed by Klein second-order Kahan-Babuška-Neumaier (KBN) double-compensated summation.

**Reference implementation:** Python `py/streaming_kbn/`. Go, TypeScript, Zig, and Rust are ports of it and must match it to 13+ decimal places. `py/streaming_kbn_deprecated/` is the legacy pre-rework version, kept only for history; do not port from it.

Inputs are assumed to be finite. No implementation validates NaN or infinity, or recovers when an intermediate product overflows.

## Module Dependencies

```
streaming_kbn/              (standalone — zero dependencies on other zpano modules)
    |
    v
(consumers)                 (indicators module, icalc CLI tool, arbitrary callers)
```

Inside the package:

```
KleinKBNAccumulator
    ├── KleinKBNSummator
    ├── RawMomentsKleinKBN ──┐
    ├── CentralMomentsKleinKBN
    └───────────────────────┴── LinearRegressionKleinKBN  (uses 2× RawMomentsKleinKBN(ddof=0) + 1 accumulator)
```

## Domain: KBN-Compensated Accumulation

### Floating-Point Summation Problem

Adding floating-point numbers naively accumulates round-off error because each addition rounds the result to the available significand. Low-order bits of the smaller operand are lost whenever the sum becomes large relative to the addend.

**Naive sum** `s += x`: worst-case relative error grows as `O(ε n)`. The bound is proportional to the **condition number**:

```
cond = Σ|xᵢ| / |Σxᵢ|
```

**Peters example** — `[1.0, 1e100, 1.0, -1e100]`:

| Method | Result |
|--------|--------|
| Exact | 2.0 |
| Naive / Kahan | 0.0 |
| **KBN / Klein KBN** | **2.0** |

### Algorithm Progression

**Kahan (1965):** Single-level compensated summation with `c = (t - s) - y`. Fails when sum and addend differ hugely.

**Kahan-Babuška-Neumaier (KBN, 1974):** Branches on which operand is larger — the term `(big - (big + small))` is exact via [2Sum](https://en.wikipedia.org/wiki/2Sum). The correction `c` accumulates losses and is applied as a final `s + c`.

**Klein second-order (2006):** Applies the same KBN trick to *the correction term itself*, and accumulates the second-level residual in `ccs`. The corrected value is `sum + cs + ccs`.

## Package Structure

### Documentation (`readme/streaming-kbn/`)
```
streaming_kbn.md                    # Package overview, all-language APIs + source mapping
klein_kbn_accumulator.md            # Algorithm documentation
klein_kbn_summator.md               # Counted sum, mean + final-sample reset
raw_moments_klein_kbn.md            # Algorithm + revert math
central_moments_klein_kbn.md        # Algorithm + Pébay formulas
linear_regression_klein_kbn.md      # Algorithm + cross-product revert math
```
The class pages use language-neutral pseudocode. The package overview documents
shared settings, NaN/error behavior, construction, entry files, and source links
for all five implementations.

### Python (`py/streaming_kbn/`) — reference
```
__init__.py                         # Package docstring + exports of all 5 classes
klein_kbn_accumulator.py            # KleinKBNAccumulator
klein_kbn_summator.py               # KleinKBNSummator
raw_moments_klein_kbn.py            # RawMomentsKleinKBN
central_moments_klein_kbn.py        # CentralMomentsKleinKBN
linear_regression_klein_kbn.py      # LinearRegressionKleinKBN
test_klein_kbn_accumulator.py       # 12 tests
test_klein_kbn_summator.py          # 10 tests
test_raw_moments_klein_kbn.py       # 19 tests
test_central_moments_klein_kbn.py   # 19 tests
test_linear_regression_klein_kbn.py # 15 tests
```
Tests use only the standard library (`math`, `random`, `statistics`, `unittest`); reference values are hard-coded.

### Go (`go/streamingkbn/`, package `streamingkbn`)
```
doc.go                              # Package documentation
kleinkbnaccumulator.go              # KleinKBNAccumulator
kleinkbnsummator.go                 # KleinKBNSummator
rawmomentskleinkbn.go               # RawMomentsKleinKBN
centralmomentskleinkbn.go           # CentralMomentsKleinKBN
linearregressionkleinkbn.go         # LinearRegressionKleinKBN
*_test.go                           # Co-located tests (shared helpers in one test file)
```

### TypeScript (`ts/streaming-kbn/`)
```
index.ts                            # Barrel re-exports (all 5 classes)
klein-kbn-accumulator.ts            # KleinKbnAccumulator
klein-kbn-summator.ts               # KleinKbnSummator
raw-moments-klein-kbn.ts            # RawMomentsKleinKbn
central-moments-klein-kbn.ts        # CentralMomentsKleinKbn
linear-regression-klein-kbn.ts      # LinearRegressionKleinKbn
*.spec.ts                           # Co-located Jasmine specs
```

### Zig (`zig/src/streaming_kbn/`)
```
streaming_kbn.zig                   # Barrel re-export
klein_kbn_accumulator.zig           # KleinKBNAccumulator + inline tests
klein_kbn_summator.zig              # KleinKBNSummator + inline tests
raw_moments_klein_kbn.zig           # RawMomentsKleinKBN + inline tests
central_moments_klein_kbn.zig       # CentralMomentsKleinKBN + inline tests
linear_regression_klein_kbn.zig     # LinearRegressionKleinKBN + inline tests
```

### Rust (`rs/src/streaming_kbn/`)
```
mod.rs                              # Module root with pub use re-exports
klein_kbn_accumulator.rs            # KleinKbnAccumulator + inline tests
klein_kbn_summator.rs               # KleinKbnSummator + inline tests
raw_moments_klein_kbn.rs            # RawMomentsKleinKbn + inline tests
central_moments_klein_kbn.rs        # CentralMomentsKleinKbn + inline tests
linear_regression_klein_kbn.rs      # LinearRegressionKleinKbn + inline tests
```

### Build Registration

- **Zig:** `build.zig` defines 5 library modules (`klein_kbn_accumulator`, `klein_kbn_summator`, `raw_moments_klein_kbn`, `central_moments_klein_kbn`, `linear_regression_klein_kbn`) + a barrel module (`streaming_kbn`). Each non-barrel source file also gets a test module (`b.createModule()` with the same imports) + `b.addTest(.{ .root_module = ..., .filters = filters })` + run artifact wired into `test_step`.
- **Rust:** `rs/src/lib.rs` requires `pub mod streaming_kbn;`; `mod.rs` declares and re-exports all 5.
- **TypeScript:** `ts/tsconfig.json` has `"streaming-kbn/**/*.ts"` in `include`; update `index.ts` when adding classes. Specs run from `dist/`, so clean `dist/` after deleting a spec.
- **Go/Python:** No registration needed — package boundaries are directory-based.

## Class Reference

### 1. `KleinKBNAccumulator`

Klein second-order KBN compensated sum. State `_sum`, `_cs`, `_ccs` (all 0.0 initially).

```
update(x):
    s = _sum;  t = s + x
    c = (s - t) + x  if |s| >= |x|  else  (x - t) + s
    _sum = t
    cs = _cs;  t = cs + c
    cc = (cs - t) + c  if |cs| >= |c|  else  (c - t) + cs
    _cs = t
    _ccs += cc                 # accumulate (the legacy version overwrote with `= cc` — a bug)

value:      _sum + _cs + _ccs
set(x):     _sum = x; _cs = _ccs = 0
reset():    set(0)
revert(x):  update(-x)         # removes any previously added value
```

### 2. `KleinKBNSummator`

Compensated sum plus sample count and mean. State: `_n` and one accumulator.

| Member | Semantics |
|--------|-----------|
| `update(x)` | `n += 1`; adds `x` to the sum only when `x != 0` (zeros are counted but never touch the compensation) |
| `revert(x)` | error if `n <= 0` ("Cannot revert from an empty summator"); `reset()` and return if `n == 1` (including a final zero sample); otherwise `n -= 1` and revert `x` only when `x != 0` |
| `reset()` | clears count and sum |
| `value` | compensated sum, 0.0 when empty |
| `mean` | `value / n`, **NaN when empty** |
| `n` | sample count |

### 3. `RawMomentsKleinKBN(ddof=1, bias=True, fisher=True)`

Mean/variance from a KBN-compensated Welford tracker; skewness/kurtosis converted from raw power sums Σx..Σx⁴ at query time. `ddof` must be a nonnegative integer. `ddof`, `bias`, `fisher` are public and mutable.

#### State

| Variable | Purpose |
|----------|---------|
| `_n` | sample count |
| `_x1.._x4` | accumulators for Σx, Σx², Σx³, Σx⁴ (powers built by repeated multiplication `x2=x*x; x3=x2*x; x4=x3*x`) |
| `_mean`, `_s` | accumulators for Welford mean and M₂ = Σ(x − x̄)² |

#### Update / revert

```
update(x):  n += 1; add x, x², x³, x⁴
            δ = x − mean;  mean += δ/n;  s += δ·(x − mean)
revert(x):  error if n <= 0 ("Cannot revert from an empty accumulator"); reset() if n == 1
            n -= 1; revert x, x², x³, x⁴
            δ = x − mean;  mean −= δ/n;  s −= δ·(x − mean)
```
Any previously added sample can be reverted (power sums and Welford's mean/M₂ are symmetric functions of the samples) → suitable for FIFO rolling windows.

#### Central moment conversion (query time)

```
n < 2 → none (NaN results)
μ₁ = Σx/n;  r = μ₁²;  mean_x2 = Σx²/n;  μ₂ = mean_x2 − r
μ₂ <= 1e-14 · mean_x2  → none           # relative cancellation guard (scale-invariant)
r *= μ₁;  μ₃ = Σx³/n − r − 3·μ₁·μ₂
r *= μ₁;  μ₄ = Σx⁴/n − r − 6·μ₂·μ₁·μ₁ − 4·μ₃·μ₁
g₁ = μ₃ / (μ₂·√μ₂)        β₂ = μ₄ / (μ₂·μ₂)
```

#### Properties (all floats unless noted)

| Property | Formula | Needs |
|----------|---------|-------|
| `mean` | Welford mean (**0.0 when empty**) | — |
| `variance`, `standard_deviation` | `max(s,0)/(n−ddof)` using instance `ddof`, no side effects | n > ddof |
| `variance_ddof_0/1`, `standard_deviation_ddof_0/1` | same with fixed ddof | n > ddof |
| `skewness_moment` | g₁ (scipy `skew(bias=True)`) | n ≥ 2 |
| `skewness_fisher` | `g₁·√(n(n−1))/(n−2)` (scipy `skew(bias=False)`) | n ≥ 3 |
| `skewness_sample` | `g₁·n²/((n−1)(n−2))` (PerformanceAnalytics "sample") | n ≥ 3 |
| `kurtosis_moment` | β₂ | n ≥ 2 |
| `kurtosis_excess` | β₂ − 3 | n ≥ 2 |
| `kurtosis_sample_excess` | G₂ = `((n²−1)·β₂ − 3(n−1)²)/((n−2)(n−3))` | n ≥ 4 |
| `kurtosis_sample` | G₂ + 3 | n ≥ 4 |
| `kurtosis_sample_corrected` | `β₂·(n²−1)/((n−2)(n−3))` (PerformanceAnalytics "sample") | n ≥ 4 |
| `x1_sum..x4_sum` | raw sums (0.0 when empty) | — |
| `x1..x4` | raw moments `sum/n` (**NaN when empty**) | — |
| `n` | int | — |

Dispatch (`skewness`, `kurtosis`) — matches scipy:

| bias | fisher | skewness | kurtosis |
|------|--------|----------|----------|
| true | true | `skewness_moment` | `kurtosis_excess` |
| true | false | `skewness_moment` | `kurtosis_moment` |
| false | true | `skewness_fisher` | `kurtosis_sample_excess` |
| false | false | `skewness_fisher` | `kurtosis_sample` |

Accuracy caveat: raw-sum conversion suffers catastrophic cancellation when the mean is large relative to the spread (e.g. `[1e8, 1e8+1, 1e8+2]` gives the correct variance 2/3 via Welford but NaN skewness/kurtosis). Use `CentralMomentsKleinKBN` for such data.

### 4. `CentralMomentsKleinKBN(ddof=1, bias=True, fisher=True)`

Pébay (2008) central-moment updates of M₁ (mean) and M₂, M₃, M₄ (sums of central powers), each a KBN accumulator. Same constructor validation as raw moments.

#### Forward update (n = count after adding x)

```
δ = x − M₁;  δₙ = δ/n;  δₙ² = δₙ·δₙ;  term = δ·δₙ·(n−1)
m2, m3 = M₂, M₃ (captured before updating)
M₁ += δₙ
M₄ += term·δₙ²·(n²−3n+3) + 6·δₙ²·m2 − 4·δₙ·m3
M₃ += term·δₙ·(n−2) − 3·δₙ·m2
M₂ += term
```

#### Inverse revert (any sample, any order)

```
error if n == 0 ("Cannot revert from an empty accumulator"); reset() if n_old == 0
M₁_old = (nₙ·M₁ − x)/nₒ;  δ = x − M₁_old;  δₙ = δ/nₙ;  term = δ·δₙ·nₒ
M₂_old = M₂ − term
M₃_old = M₃ − (term·δₙ·(nₙ−2) − 3·δₙ·M₂_old)
M₄_old = M₄ − (term·δₙ²·(nₙ²−3nₙ+3) + 6·δₙ²·M₂_old − 4·δₙ·M₃_old)
```
Restored values are written with `set()`, which clears the compensation terms; repeated reverts can accumulate rounding error (especially for large-offset data) but removal of the oldest sample (FIFO) is supported.

#### Properties

| Property | Formula | Needs |
|----------|---------|-------|
| `n`, `mean` (0.0 when empty) | | |
| `variance` / `standard_deviation` | `max(M₂,0)/(n−ddof)` | n > ddof |
| `skewness` | g₁ = `√n·M₃/(M₂·√M₂)`; bias=false → `g₁·√(n(n−1))/(n−2)` | n ≥ 2 and M₂ > 0; unbiased n ≥ 3 |
| `kurtosis` | β₂ = `n·M₄/(M₂·M₂)`; bias=true → β₂−3 (fisher) / β₂; bias=false → G₂ (fisher) / G₂+3 | n ≥ 2 and M₂ > 0; unbiased n ≥ 4 |

No separate variant getters (unlike raw moments).

#### Raw vs Central Moments

| Aspect | RawMomentsKleinKBN | CentralMomentsKleinKBN |
|--------|--------------------|------------------------|
| Forward accuracy (large mean) | ⚠️ higher moments lose precision | ✅ best |
| Revert | ✅ any sample, compensation preserved | ✅ any sample, compensation cleared |
| Variant getters | ✅ many | dispatch only |

### 5. `LinearRegressionKleinKBN()`

Streaming OLS `y = b·x + a`. State: `_n`, `_x_moments`/`_y_moments` = `RawMomentsKleinKBN(ddof=0)`, `_s_xy` accumulator.

```
update(x, y):  n_old = n; n += 1
               s_xy += (x̄ − x)·(ȳ − y)·n_old/(n_old+1)    # means before the sample
               x_moments.update(x); y_moments.update(y)
revert(x, y):  error if n == 0 ("Cannot revert from an empty regression"); reset() if n == 1
               x_moments.revert(x); y_moments.revert(y)    # means after removal
               n' = n − 1;  s_xy −= (x̄ − x)·(ȳ − y)·n'/(n'+1);  n = n'
```

| Property | Formula | NaN when |
|----------|---------|----------|
| `n` | | |
| `mean_x`, `mean_y` | (0.0 when empty) | |
| `variance_x`, `variance_y` | population variance | empty |
| `co_moment` | S_xy (0.0 when empty) | |
| `covariance` | S_xy / n | n < 1 |
| `slope` | S_xy / S_xx, S_xx = variance_x·n | n < 2 or S_xx = 0 |
| `intercept` | ȳ − slope·x̄ | slope NaN |
| `correlation` | `clamp(S_xy/(σx·σy·n), −1, 1)` | n < 2 or σx·σy = 0 |

## Cross-Language Conventions

### NaN / zero policy

All languages return NaN (not null/None/Option) where Python returns `math.nan`, and 0.0 where Python returns 0.0. Notable empty-state values: `mean`/`mean_x`/`mean_y`/`value`/`x*_sum`/`co_moment` → 0.0; `KleinKBNSummator.mean`, `x1..x4`, `variance*`, `covariance` → NaN.

### Errors

| Situation | Python | Go | TypeScript | Zig | Rust |
|-----------|--------|----|------------|-----|------|
| Revert on empty (summator, moments, regression) | `ValueError` | `panic` | `throw new Error` | `error.EmptyRevert` (`revert` returns `!void`) | `panic!` |
| Invalid ddof | `ValueError` | `panic` (negative) | `throw` (not non-negative integer) | unrepresentable (`u32`) | unrepresentable (`usize`) |

Messages mirror the Python text. `KleinKBNAccumulator.revert` never fails.

In Go, evaluate count products in moment formulas in floating-point arithmetic
after converting the count, so `int` products cannot overflow. For variance,
compare `n <= ddof` before subtracting unsigned counts; do not narrow Rust's
`usize` values to signed integers.

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
| Class names | `KleinKBNAccumulator`, `KleinKBNSummator`, `RawMomentsKleinKBN`, `CentralMomentsKleinKBN`, `LinearRegressionKleinKBN` | same as Python | `…Kbn` (`KleinKbnSummator`, `RawMomentsKleinKbn`, …) | same as Python (`…KBN`) | `…Kbn` |
| Moments constructor | `(ddof=1, bias=True, fisher=True)` | `NewRawMomentsKleinKBN(ddof, bias, fisher)` | `new RawMomentsKleinKbn(ddof = 1, bias = true, fisher = true)` | struct literal with field defaults `.{ .ddof = 1, .bias = true, .fisher = true }` | `::new(ddof, bias, fisher)`; `Default` = (1, true, true) |
| Other constructors | `X()` | `&KleinKBNAccumulator{}`, `NewKleinKBNSummator()`, `NewLinearRegressionKleinKBN()` | `new X()` | `X{}` | `X::new()` / `Default` |
| Property access | `m.skewness_moment` | `m.SkewnessMoment()` | `m.skewnessMoment` (getter) | `m.skewnessMoment()` | `m.skewness_moment()` (`&self`) |
| ddof/bias/fisher | public attrs | `Ddof()`/`SetDdof()` etc. | public properties | public fields | `ddof()`/`set_ddof()` etc. |

### Test Conventions

| Language | Framework | Location | Key assertion |
|----------|-----------|----------|---------------|
| Python | `unittest` | `test_*.py` next to sources | `assertAlmostEqual(x, y, places=N)` |
| Go | `testing` | `*_test.go`, `t.Parallel()` | `almostEqual(a, b, tol)` |
| TypeScript | Jasmine | `*.spec.ts` | `toBeCloseTo(x, N)` |
| Zig | built-in | inline at bottom of source | `try testing.expect(almostEqual(...))`, `expectError(error.EmptyRevert, ...)` |
| Rust | built-in | inline `#[cfg(test)] mod tests` | `assert!(almost_equal(..))`, `#[should_panic]` |

Every Python test method is ported in every language (except `test_invalid_ddof` in Zig/Rust). Tests that use Python's `math.fsum` use a test-local port of CPython's `math_fsum` (Shewchuk partials + half-even correction); tests that use `random.Random(42)` use each language's own seeded PRNG with the same distributions.

## Test Data & Expected Values

### Bacon Data (24 portfolio returns; Bacon 2008, p. 65)

```
[ 0.003,  0.026,  0.011, -0.010,  0.015,  0.025,  0.016,  0.067,
 -0.014,  0.040, -0.005,  0.081,  0.040, -0.037, -0.061,  0.017,
 -0.049, -0.022,  0.070,  0.058, -0.065,  0.024, -0.005, -0.009 ]
```

### Exact reference values for Bacon

Computed with exact rational arithmetic on the binary float inputs (square roots with 50-digit decimals), rounded to the nearest float. Shared by raw and central moments tests.

| Metric | Expected |
|--------|----------|
| mean | 0.009000000000000001 |
| variance_ddof_0 | 0.0014989166666666668 |
| variance_ddof_1 | 0.0015640869565217393 |
| standard_deviation_ddof_0 | 0.03871584516275819 |
| standard_deviation_ddof_1 | 0.039548539246370897 |
| skewness_moment (skew bias=True) | -0.08256245520856804 |
| skewness_fisher (skew bias=False) | -0.08817174934967535 |
| skewness_sample (PA "sample") | -0.09398413873544505 |
| kurtosis_moment | 2.4324537941078743 |
| kurtosis_excess | -0.5675462058921257 |
| kurtosis_sample_excess | -0.40766032118608714 |
| kurtosis_sample | 2.592339678813913 |
| kurtosis_sample_corrected (PA "sample") | 3.027404613878848 |
| x1_sum / x2_sum / x3_sum / x4_sum | 0.21600000000000003 / 0.037918 / 0.0008738040000000003 / 0.00014466403 |
| x1 / x2 / x3 / x4 | 0.009000000000000001 / 0.0015799166666666668 / 3.640850000000001e-05 / 6.0276679166666674e-06 |

Shifted data `1e4 + x` (central moments): skewness (biased) -0.08256245521966786, excess kurtosis -0.5675462058934164.

### Linear regression (x = benchmark, y = portfolio, 24 values each; see the Python test)

| Metric | Expected |
|--------|----------|
| slope | 0.9988502086225746 |
| intercept | -0.001030120844918352 |
| correlation | 0.9693858148753051 |
| co_moment | 0.033844 |
| covariance | 0.0014101666666666668 |
| variance_x | 0.0014117899305555557 |
| variance_y | 0.0014989166666666668 |

### Accumulator

- Peters `[1.0, 1e100, 1.0, -1e100]` → 2.0 (naive 0.0).
- NumPy issue 8786 data → `-0.377392919181026`.
- Second-level correction `[1e-16, -1e16, 1.0, 1e-16, -1.0, -1e-16, -1e-32, 1e16]` → `9.999999999999999e-17` (the legacy `_ccs = cc` bug gives `-1.0000000000000001e-16`).

## References

- Higham, N. J. (1993). "The accuracy of floating point summation". *SIAM Journal on Scientific Computing*, 14(4), 783–799.
- Kahan, W. (1965). "Further remarks on reducing truncation errors". *Communications of the ACM*, 8(1), 40.
- Neumaier, A. (1974). "Rundungsfehleranalyse einiger Verfahren zur Summation endlicher Summen". *ZAMM*, 54(1), 39–51.
- Klein, A. (2006). "A generalized Kahan–Babuška-Summation-Algorithm". *Computing*, 76(3–4), 279–293.
- Pébay, P. (2008). "Formulas for robust, one-pass parallel computation of covariances and arbitrary-order statistical moments". *Sandia Report SAND2008-6212*.
- Welford, B. P. (1962). "Note on a method for calculating corrected sums of squares and products". *Technometrics*, 4(3), 419–420.
- Bacon, C. R. (2008). *Practical Portfolio Performance Measurement and Attribution*, 2nd ed., Wiley.
- Cook, J. D. [Skewness and kurtosis](https://www.johndcook.com/skewness_kurtosis.html); [Running regression](https://www.johndcook.com/running_regression.html).
- Kuiperzone. [Compensated-Accumulators](https://github.com/kuiperzone/Compensated-Accumulators).
- Wikipedia. [Kahan summation algorithm](https://en.wikipedia.org/wiki/Kahan_summation_algorithm); [2Sum](https://en.wikipedia.org/wiki/2Sum).
- NumPy issue #8786 — [Badly conditioned sum](https://github.com/numpy/numpy/issues/8786).
- CPython `math.fsum` — [mathmodule.c](https://github.com/python/cpython/blob/main/Modules/mathmodule.c).
