# Streaming Linear Regression (`LinearRegressionKleinKBN`)

## Domain: Streaming Linear Regression with Update/Revert

Fitting a linear model $y = \beta_1 x + \beta_0$ to a data stream in O(1) per observation requires maintaining sufficient statistics from which the slope, intercept, and correlation can be derived at query time. The statistics needed are:

| Statistic | Purpose |
| --- | --- |
| $n$ | Sample count |
| $\bar{x},\ \bar{y}$ | Running means of $x$ and $y$ |
| $S_{xx} = \sum (x_i - \bar{x})^2$ | Sum of squared $x$ deviations |
| $S_{yy} = \sum (y_i - \bar{y})^2$ | Sum of squared $y$ deviations |
| $S_{xy} = \sum (x_i - \bar{x})(y_i - \bar{y})$ | Sum of cross products |

Given these, the regression statistics are:

$$
\begin{aligned}
\beta_1 &= \frac{S_{xy}}{S_{xx}} \\
\beta_0 &= \bar{y} - \beta_1 \bar{x} \\
r      &= \frac{S_{xy}}{\sqrt{S_{xx} S_{yy}}}
\end{aligned}
$$

`LinearRegressionKleinKBN` maintains $\bar{x}$, $\bar{y}$, $S_{xx}$, $S_{yy}$ via two `RawMomentsKleinKBN` instances (which use KBN-compensated Welford variance tracking), and $S_{xy}$ via a dedicated `KleinKBNAccumulator` for the cross-product sum. All internal accumulators use Klein second-order Kahan-Babuška-Neumaier compensation for improved numerical stability.

## Algorithm

### Welford Cross-Product Update

The update formula for $S_{xy}$ is derived from Welford's covariance identity. Let $n_0$ be the count before adding $(x, y)$, and let $\bar{x}_0,\ \bar{y}_0$ be the means before the update:

$$
S_{xy}^{(1)} = S_{xy}^{(0)} + \frac{n_0}{n_0 + 1}(\bar{x}_0 - x)(\bar{y}_0 - y)
$$

Pseudocode:

```pseudocode
n_old ← n
n ← n + 1
term ← (x̄ − x) · (ȳ − y) · n_old / n
S_xy ← S_xy + term
x_moments.update(x)     # updates x̄, S_xx
y_moments.update(y)     # updates ȳ, S_yy
```

### Revert Inverse Formula

To remove an observation $(x, y)$, we first revert `_x_moments` and `_y_moments` (which restores the means $\bar{x}_0,\ \bar{y}_0$ to their values before $(x, y)$ was added), then subtract the same cross-product term that was added:

```pseudocode
x_moments.revert(x)     # restores x̄₀
y_moments.revert(y)     # restores ȳ₀
n ← n − 1
term ← (x̄₀ − x) · (ȳ₀ − y) · n / (n + 1)
S_xy ← S_xy − term
```

Edge case: reverting the last element ($n: 1 \to 0$) simply calls `reset()`, since a single point contributes nothing to covariance and the term is zero.

### Why FIFO Revert Works

`LinearRegressionKleinKBN.revert(x, y)$ works for **any** observation in the window, not just the most recent. This enables FIFO rolling-window semantics (remove the oldest, add the newest). The order-independence follows from two facts:

**1. `RawMomentsKleinKBN.revert()` is order-independent.** As shown in the [raw moments documentation](raw_moments_klein_kbn.md), both the power sums $\Sigma x^p$ and the Welford variance tracker use formulas that depend only on the current state and the value $x_k$ being removed — not on its insertion position. Together these correctly restore $\bar{x}_0$ and $S_{xx}$ for any $x_k$.

**2. The cross-product revert formula is also order-independent.** Let the current state before revert be $(n, \bar{x}, \bar{y})$ (the means include the observation $(x_k, y_k)$ being removed). After reverting `_x_moments` and `_y_moments`, the means are restored to $\bar{x}_0,\ \bar{y}_0$ — the means of the remaining $n-1$ samples. The term to subtract is:

$$
\text{term} = (\bar{x}_0 - x_k)\,(\bar{y}_0 - y_k)\,\frac{n-1}{n}
$$

This uses only the restored means and the removed sample — quantities that are well-defined regardless of when $(x_k, y_k)$ was added. The derivation from the definition of $S_{xy}$ confirms this:

$$
\begin{aligned}
S_{xy} &= \sum_{i=1}^n (x_i - \bar{x})(y_i - \bar{y}) \\[4pt]
S_{xy}' &= \sum_{i \ne k} (x_i - \bar{x}_0)(y_i - \bar{y}_0) \\[4pt]
        &= S_{xy} - (\bar{x}_0 - x_k)(\bar{y}_0 - y_k)\,\frac{n-1}{n}
\end{aligned}
$$

**3. KBN compensation is preserved.** Both `_x_moments.revert(x)$, `_y_moments.revert(y)$, and `_s_xy.update(-term)$ use KBN subtraction (via `update(-*)$), maintaining the double-compensated error correction. No precision is lost on revert.

Together these facts mean you can use a deque-based rolling window with `LinearRegressionKleinKBN` and get numerically correct results over arbitrarily many window shifts.

### Property Formulas

With `RawMomentsKleinKBN(ddof=0)`, the variance is $S_{xx} / n$ (population), so $S_{xx} = \text{variance}_x \cdot n$:

$$
\begin{aligned}
\beta_1 &= \frac{S_{xy}}{S_{xx}} \\
\beta_0 &= \bar{y} - \beta_1 \bar{x} \\
r      &= \frac{S_{xy}}{\sqrt{S_{xx} S_{yy}}}
\end{aligned}
$$

Guard clauses: slope returns $0.0$ when $n < 2$ or $S_{xx} = 0$; correlation returns $0.0$ when $n < 2$ or either standard deviation is zero.

## Implementation: `LinearRegressionKleinKBN`

### State Variables

| Variable | Type | Purpose |
| --- | --- | --- |
| `n` | `int` | Sample count |
| `_x_moments` | `RawMomentsKleinKBN(ddof=0)` | Running mean $\bar{x}$, variance $S_{xx}/n$ |
| `_y_moments` | `RawMomentsKleinKBN(ddof=0)` | Running mean $\bar{y}$, variance $S_{yy}/n$ |
| `_s_xy` | `KleinKBNAccumulator` | Sum of cross products $S_{xy}$ |

### Methods

- **`update(x, y)`** — Adds observation $(x, y)$ using the Welford cross-product identity. Updates $\bar{x},\ S_{xx},\ \bar{y},\ S_{yy},\ S_{xy}$ in O(1).
- **`revert(x, y)`** — Removes an observation $(x, y)$ using the inverse formula. Works for any sample in the window (not just the most recent), enabling FIFO rolling windows via a deque.
- **`reset()`** — Resets all accumulators and count to zero.
- **`slope`** — Returns $\beta_1 = S_{xy} / S_{xx}$. Returns $0.0$ if $n < 2$ or $S_{xx} = 0$.
- **`intercept`** — Returns $\beta_0 = \bar{y} - \beta_1 \bar{x}$.
- **`correlation`** — Returns Pearson's $r = S_{xy} / \sqrt{S_{xx} S_{yy}}$. Returns $0.0$ if $n < 2$ or either variance is zero.

### Usage

```python
from linear_regression_klein_kbn import LinearRegressionKleinKBN

reg = LinearRegressionKleinKBN()
for x, y in zip(xs, ys):
    reg.update(x, y)

print(reg.slope, reg.intercept, reg.correlation)
```

Rolling window via FIFO revert:

```python
from collections import deque

window = deque(maxlen=100)
reg = LinearRegressionKleinKBN()

for x, y in stream:
    if len(window) == 100:
        x_old, y_old = window[0]
        reg.revert(x_old, y_old)
    reg.update(x, y)
    window.append((x, y))
```

## Comparison: `LinearRegressionKleinKBN` vs `Regression`

| Aspect | `Regression` | `LinearRegressionKleinKBN` |
| --- | --- | --- |
| $S_{xx}, S_{yy}$ | `CentralMoments(ddof=1)` | `RawMomentsKleinKBN(ddof=0)` |
| $S_{xy}$ | Plain `float` | `KleinKBNAccumulator` (KBN-compensated) |
| `revert(x, y)` | ❌ Not implemented | ✅ Inverse formula with KBN |
| Rolling window | ❌ Requires full refit | ✅ FIFO via revert |
| Numerical stability (forward) | ✅ Pébay central moments | ✅ KBN-compensated raw moments |
| Numerical stability (revert) | N/A | ✅ KBN compensation preserved |

The key advantage of `LinearRegressionKleinKBN` is **revert support** enabling O(1) rolling-window regression. The KBN compensation on $S_{xy}$ also provides better accuracy on the cross-product sum compared to a plain float accumulator.

### References

- Cook, J. D. [Running regression](https://www.johndcook.com/running_regression.html).
- Welford, B. P. (1962). "Note on a method for calculating corrected sums of squares and products". *Technometrics*, 4(3), 419–420.
- Klein, A. (2006). "A generalized Kahan–Babuška-Summation-Algorithm". *Computing*, 76(3–4), 279–293.
- Higham, N. J. (1993). "The accuracy of floating point summation". *SIAM Journal on Scientific Computing*, 14(4), 783–799.
- Kuiperzone. [Compensated-Accumulators](https://github.com/kuiperzone/Compensated-Accumulators).
