# Raw Power-Sum Streaming Statistics (`RawMomentsKleinKBN`)

## Domain: Streaming Mean, Variance, Skewness, Kurtosis

Computing the first four moments of a data stream in O(1) per element requires maintaining sufficient summary statistics from which mean, variance, skewness, and kurtosis can be derived at query time. Two broad strategies exist:

| Approach | Method | Revert support | Numerical stability |
| --- | --- | --- |
| **Raw power sums** | Accumulate Σx, Σx², Σx³, Σx⁴; convert to central moments on the fly | ✅ FIFO (additive revert) | ❌ Catastrophic cancellation when values have non-zero mean |
| **Central moments** (Pébay) | Maintain running m₁, m₂, m₃, m₄ directly via Pébay's update | ⚠️ LIFO only, with compensation reset | ✅ More stable for forward-only computation |

`RawMomentsKleinKBN` implements the **raw power-sum approach** with KBN (Klein second-order) compensated accumulation for all four power sums and the Welford variance tracker. Because power sums are linear (`Σ(x - a) = Σx - na`), reverting a sample is simply `update(-x)` — making FIFO rolling-window support natural.

### Numerical Stability Trade-off

Raw power sums suffer from catastrophic cancellation when converting to central moments for data with non-zero mean. For example, `[1e8, 1e8 + 1, 1e8 + 2]` produces `Σx²` of order `3e16`, while the variance is `1.0` — subtracting two nearly-equal large numbers loses many significant bits. The KBN compensation mitigates this but does not eliminate the problem.

For forward-only computation on data with large means, prefer `CentralMomentsKleinKBN` which avoids raw sums entirely.

## Algorithm

### Accumulation

Four `KleinKBNAccumulator` instances track the power sums:

```pseudocode
_x1 = Σx
_x2 = Σx²
_x3 = Σx³
_x4 = Σx⁴
```

A separate Welford-style variance tracker (`_mean` and `_s`, also KBN-compensated) maintains:

```pseudocode
δ    = x − mean
mean += δ / n
s    += δ · (x − mean)
```

The `_s` accumulator holds the sum of squared deviations from the mean, which converts directly to variance: `s² = _s / (n − ddof)`.

### Central Moment Conversion

At query time, raw power sums are converted to central moments. Let `A = Σx / n`, the sample mean:

```pseudocode
B = Σx²/n − A²                                          [2nd central ≈ variance × n]
C = Σx³/n − A³ − 3·A·B                                  [3rd central]
D = Σx⁴/n − A⁴ − 6·B·A² − 4·C·A                        [4th central]
```

Then:

```pseudocode
g₁ = √n · C / B^1.5                                     [skewness, bias=True]
G₁ = g₁ · √(n·(n−1)) / (n−2)                            [bias=False]

g₂ = n · D / B² − 3                                     [excess kurtosis, bias=True, fisher=True]
G₂ = ((n²−1) · (n·D/B²) − 3·(n−1)²) / ((n−2)·(n−3))   [bias=False, fisher=True]
```

Guard clauses: skewness requires `n ≥ 3` and `B > 1e-14`; kurtosis requires `n > 3` and `B > 1e-14`. Both return `nan` when guards fail.

### Why Revert Is Order-Independent (FIFO Support)

The revert operation removes a sample $x$ by calling `update(-x)$. Unlike the Pébay central moment revert (which uses the inverse of a sequential update), the raw power-sum revert works for **any** sample regardless of insertion order in the sequence.

**Power sums are linear and commutative:**

$$
\Sigma x' = \Sigma x - x_k, \qquad
\Sigma x^2{}' = \Sigma x^2 - x_k^2, \qquad
\Sigma x^3{}' = \Sigma x^3 - x_k^3, \qquad
\Sigma x^4{}' = \Sigma x^4 - x_k^4
$$

Each is a simple sum of terms — addition is commutative, so subtracting $x_k^p$ correctly removes it whether it was added first, last, or anywhere in between.

**The Welford variance revert is also order-independent.** Let the current state be $(n, \bar{x}, S_{xx})$ and let $x_k$ be *any* sample to remove. The revert formulas implement:

$$
\begin{aligned}
n'      &= n - 1 \\[2pt]
\bar{x}' &= \bar{x} - \frac{x_k - \bar{x}}{n'} \\[4pt]
\delta  &= x_k - \bar{x} \\
S_{xx}' &= S_{xx} - \frac{n}{n'}\, \delta^2
\end{aligned}
$$

These depend only on the current state and $x_k$ — not on which position $x_k$ occupies in the insertion sequence. The derivation starts from the raw definitions:

$$
\begin{aligned}
n' \bar{x}' &= \sum_{i \ne k} x_i
            = \Bigl(\sum_i x_i\Bigr) - x_k
            = n \bar{x} - x_k \\[4pt]
\bar{x}'   &= \frac{n \bar{x} - x_k}{n'}
            = \bar{x} - \frac{x_k - \bar{x}}{n'} \\[6pt]
S_{xx}     &= \sum_i (x_i - \bar{x})^2 \\[3pt]
S_{xx}'    &= \sum_{i \ne k} (x_i - \bar{x}')^2
\end{aligned}
$$

Substituting $\bar{x}'$ and expanding the sum of squares yields the formula for $S_{xx}'$, with no assumption about $x_k$ being the most recently added element. The same algebra drives the `_mean` and `_s` accumulators: $\delta = x_k - \bar{x}$ depends only on the current mean and the sample being removed, and the accumulator update subtracts the correct contribution regardless of when $x_k$ was added.

**KBN compensation is preserved.** Because `update(-x)$ uses the same KBN branch logic (the `|sum| ≥ |x|` comparison and correction path), the compensation terms `_cs` and `_ccs` are correctly maintained through the subtraction. No precision is lost on revert — unlike a `set()`-based approach which zeros out the compensation.

This order-independence is what makes `RawMomentsKleinKBN` suitable for FIFO rolling windows: the oldest sample can be removed from a full window and the new sample added, with both operations maintaining the full KBN error correction.

## Implementation: `RawMomentsKleinKBN`

### State Variables

| Variable | Type | Purpose |
| --- | --- | --- |
| `n` | `int` | Sample count |
| `_x1` | `KleinKBNAccumulator` | Σx |
| `_x2` | `KleinKBNAccumulator` | Σx² |
| `_x3` | `KleinKBNAccumulator` | Σx³ |
| `_x4` | `KleinKBNAccumulator` | Σx⁴ |
| `_mean` | `KleinKBNAccumulator` | Welford running mean |
| `_s` | `KleinKBNAccumulator` | Welford sum of squared differences |
| `ddof` | `int` | Delta degrees of freedom for variance |
| `bias` | `bool` | If True, population standardized moments |
| `fisher` | `bool` | If True, excess kurtosis (Gaussian → 0) |

### Parameters

- **`ddof`** (int, default=1) — Divisor for variance is `n − ddof`. `ddof=0` gives population variance, `ddof=1` gives sample.
- **`bias`** (bool, default=True) — If True, compute population skewness/kurtosis. If False, apply Fisher-Pearson bias correction.
- **`fisher`** (bool, default=True) — If True, return excess kurtosis (subtract 3). If False, return raw kurtosis (Gaussian → 3).

### Methods

- **`update(x)`** — Adds sample `x`, updating all four power sums and the Welford variance tracker via KBN-compensated accumulation.
- **`revert(x)`** — Removes sample `x` by calling `update(-x)`. Works for any sample regardless of insertion order (FIFO or LIFO). Power sums are linear and commutative, and the Welford variance revert depends only on the current state and the value being removed. KBN compensation is preserved through the subtraction.
- **`reset()`** — Resets all accumulators and count to zero.
- **`mean`** — Returns the running mean (`_mean.value`).
- **`variance`** — Returns `_s.value / (n − ddof)`. Returns `nan` if `n ≤ ddof` or if `_s` is negative (floating-point noise).
- **`skewness`** — Returns the computed skewness. Returns `nan` if `n < 3` or `B ≤ 1e-14`.
- **`kurtosis`** — Returns the computed (excess) kurtosis. Returns `nan` if `n ≤ 3` or `B ≤ 1e-14`.

### Usage

```python
from raw_moments_klein_kbn import RawMomentsKleinKBN

m = RawMomentsKleinKBN(ddof=1, bias=True, fisher=True)
for x in data:
    m.update(x)

print(m.mean, m.variance, m.skewness, m.kurtosis)
```

Rolling window:

```python
window = deque(maxlen=100)
for x in stream:
    if len(window) == 100:
        m.revert(window[0])
    m.update(x)
    window.append(x)
```

### References

- Higham, N. J. (1993). "The accuracy of floating point summation". *SIAM Journal on Scientific Computing*, 14(4), 783–799.
- Klein, A. (2006). "A generalized Kahan–Babuška-Summation-Algorithm". *Computing*, 76(3–4), 279–293.
- Welford, B. P. (1962). "Note on a method for calculating corrected sums of squares and products". *Technometrics*, 4(3), 419–420.
- Cook, J. D. [Skewness and kurtosis](https://www.johndcook.com/skewness_kurtosis.html).
- Kuiperzone. [Compensated-Accumulators](https://github.com/kuiperzone/Compensated-Accumulators).
- NumPy issue #8786 — [Badly conditioned sum](https://github.com/numpy/numpy/issues/8786).
