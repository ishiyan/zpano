# Pébay Central-Moment Streaming Statistics (`CentralMomentsKleinKBN`)

## Domain: Streaming Mean, Variance, Skewness, Kurtosis

Computing the first four moments of a data stream in O(1) per element can be done either via raw power sums (Σx, Σx², …) or by maintaining running central moments (m₁, m₂, m₃, m₄) directly. `CentralMomentsKleinKBN` implements the latter using Pébay's update formulas, with all accumulators backed by `KleinKBNAccumulator` for KBN double-compensated summation.

| Property | RawMomentsKleinKBN | CentralMomentsKleinKBN |
| --- | --- | --- |
| Accumulates | Σx, Σx², Σx³, Σx⁴ | m₁, m₂, m₃, m₄ |
| Numerical stability | ❌ Cancellation for large mean | ✅ Direct central moments |
| Revert | ✅ FIFO via `update(-x)` | ⚠️ LIFO only, compensation reset |
| Preferred use | Rolling windows with moderate values | Forward-only, high-precision |

### Why Central Moments Are More Stable

When data has a non-zero mean, raw power sums like Σx² contain a large `n·x̄²` term that must be subtracted to get the variance. The subtraction `Σx² − (Σx)²/n` can catastrophically cancel, losing precision. Central moment updates avoid this by working directly with deviations from the running mean.

## Algorithm

### Pébay Forward Update

Pébay's formulas (also sometimes called the "online" or "parallel" algorithm) update the central moments in O(1) per sample without computing raw power sums:

```pseudocode
n_new = n_old + 1
δ     = x − m₁                    [deviation from old mean]
δₙ    = δ / n_new                 [normalized deviation]
δₙ²   = δₙ · δₙ
term  = δ · δₙ · n_old

m₁   += δₙ                        [mean update]
m₄   += term·δₙ²·(n_new² − 3·n_new + 3)
       + 6·δₙ²·m₂ − 4·δₙ·m₃      [4th moment update]
m₃   += term·δₙ·(n_new − 2)
       − 3·δₙ·m₂                  [3rd moment update]
m₂   += term                      [2nd moment / variance sum]
```

Each of `m₁, m₂, m₃, m₄` is stored as a `KleinKBNAccumulator`, providing KBN compensation for the running sums.

### Inverse Pébay Revert (LIFO)

Only the most recently added sample can be reverted (LIFO stack). The inverse formulas restore the state to exactly what it would be had the last sample never been added:

```pseudocode
n_new = n (before revert), n_old = n − 1

m₁_old = (n_new · m₁_new − x) / n_old
δ      = x − m₁_old
δₙ     = δ / n_new
δₙ²    = δₙ · δₙ
term   = δ · δₙ · n_old

m₂_old = m₂_new − term
m₃_old = m₃_new − (term·δₙ·(n_new−2) − 3·δₙ·m₂_old)
m₄_old = m₄_new − (term·δₙ²·(n_new²−3·n_new+3)
                   + 6·δₙ²·m₂_old − 4·δₙ·m₃_old)
```

After computing the restored values, each accumulator is set via `set(value)`, which resets the KBN compensation terms (`_cs`, `_ccs`) to zero. This means subsequent updates rebuild compensation from the restored value — a minor loss of error correction per revert.

Because of this compensation reset, `CentralMomentsKleinKBN` is not well-suited for frequent FIFO rolling-window use. Prefer `RawMomentsKleinKBN` for rolling windows.

### Query Formulas

Moments are derived directly from the running central moment accumulators:

```pseudocode
variance     = m₂ / (n − ddof)
skewness     = √n · m₃ / m₂^1.5                       [bias=True]
kurtosis     = n · m₄ / m₂² − 3                       [bias=True, fisher=True]
```

Bias and Fisher corrections follow the same pattern as `RawMomentsKleinKBN`.

Guard clauses: skewness returns `0.0` when `n < 3` or `m₂ ≤ 0`; kurtosis returns `0.0` when `n ≤ 3` or `m₂ ≤ 0`.

## Implementation: `CentralMomentsKleinKBN`

### State Variables

| Variable | Type | Purpose |
| --- | --- | --- |
| `n` | `int` | Sample count |
| `m1` | `KleinKBNAccumulator` | Running mean m₁ |
| `m2` | `KleinKBNAccumulator` | Sum of squared deviations m₂ |
| `m3` | `KleinKBNAccumulator` | Sum of cubed deviations m₃ |
| `m4` | `KleinKBNAccumulator` | Sum of quartic deviations m₄ |
| `ddof` | `int` | Delta degrees of freedom for variance |
| `bias` | `bool` | If True, population standardized moments |
| `fisher` | `bool` | If True, excess kurtosis (Gaussian → 0) |

### Parameters

- **`ddof`** (int, default=1) — Divisor for variance is `n − ddof`. `ddof=0` gives population variance, `ddof=1` gives sample.
- **`bias`** (bool, default=True) — If True, compute population skewness/kurtosis. If False, apply Fisher-Pearson bias correction.
- **`fisher`** (bool, default=True) — If True, return excess kurtosis (subtract 3). If False, return raw kurtosis (Gaussian → 3).

### Methods

- **`update(x)`** — Adds sample `x` using Pébay's O(1) central moment formulas with KBN-compensated accumulation.
- **`revert(x)`** — LIFO revert: removes the most recently added sample `x` using inverse Pébay formulas. Resets KBN compensation on the restored values. Raises `ValueError` if `n = 0`.
- **`reset()`** — Resets all accumulators and count to zero.
- **`mean`** — Returns `m1.value`.
- **`variance`** — Returns `m2.value / (n − ddof)`. Returns `0.0` if `n ≤ ddof`.
- **`standard_deviation`** — Returns `√variance`.
- **`skewness`** — Returns the computed skewness. Returns `0.0` if `n < 3` or `m₂ ≤ 0`.
- **`kurtosis`** — Returns the computed (excess) kurtosis. Returns `0.0` if `n ≤ 3` or `m₂ ≤ 0`.

### Usage

```python
from central_moments_klein_kbn import CentralMomentsKleinKBN

m = CentralMomentsKleinKBN(ddof=0, bias=True, fisher=True)
for x in data:
    m.update(x)

print(m.mean, m.variance, m.skewness, m.kurtosis)
```

LIFO revert (rare use — only for undoing the most recent sample):

```python
m.update(x)
# ...
m.revert(x)  # restores prior state
```

For rolling FIFO windows, use `RawMomentsKleinKBN` instead.

## Comparison: `CentralMomentsKleinKBN` vs `RawMomentsKleinKBN`

| Aspect | CentralMomentsKleinKBN | RawMomentsKleinKBN |
| --- | --- | --- |
| Forward accuracy | ✅ Best (no raw-sum cancellation) | ⚠️ Good with KBN, but \|x̄\| ≫ 0 erodes precision |
| Revert | ⚠️ LIFO only, compensation reset | ✅ FIFO via `update(-x)` |
| Rolling window | ❌ Not recommended | ✅ Natural |
| Skewness/kurtosis guards | Returns `0.0` | Returns `nan` |
| Complexity | Higher (Pébay formulas) | Lower (raw sums) |

Choose `CentralMomentsKleinKBN` for forward-only streaming where numerical accuracy matters. Choose `RawMomentsKleinKBN` when you need FIFO rolling-window support or prefer raw-sum semantics.

### References

- Pébay, P. (2008). "Formulas for robust, one-pass parallel computation of covariances and arbitrary-order statistical moments". *Sandia Report SAND2008-6212*.
- Cook, J. D. [Skewness and kurtosis](https://www.johndcook.com/skewness_kurtosis.html).
- Klein, A. (2006). "A generalized Kahan–Babuška-Summation-Algorithm". *Computing*, 76(3–4), 279–293.
- Kuiperzone. [Compensated-Accumulators](https://github.com/kuiperzone/Compensated-Accumulators).
- Higham, N. J. (1993). "The accuracy of floating point summation". *SIAM Journal on Scientific Computing*, 14(4), 783–799.
