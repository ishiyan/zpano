# Raw power-sum streaming statistics

`RawMomentsKleinKBN` computes mean, variance, skewness, and kurtosis while retaining the first four raw power sums. Mean and variance come from a separate compensated Welford tracker; higher moments are converted from power sums at query time. For source files and language-specific APIs, see the [package overview](streaming_kbn.md).

## State and settings

| State | Initial value | Purpose |
| --- | --- | --- |
| `n` | 0 | Sample count |
| `P1`, `P2`, `P3`, `P4` | Zero-valued KBN accumulators | Sums of x, x², x³, x⁴ |
| `A` | Zero-valued KBN accumulator | Welford running mean |
| `S` | Zero-valued KBN accumulator | Sum of squared deviations from the mean |

All six accumulators are [KleinKBNAccumulator](klein_kbn_accumulator.md) instances. Every update, removal, and query takes O(1) time; storage is O(1).

The conventional settings are `ddof = 1`, `bias = true`, `fisher = true`. They are mutable; reset retains them. `ddof` must be a nonnegative integer. Constructor syntax and validation differ by language as detailed in the [overview](streaming_kbn.md#moment-settings).

| Operation | Behavior |
| --- | --- |
| `update(x)` | Add a sample to the power sums and Welford tracker |
| `revert(x)` | Remove a present sample; fail if empty; reset after final removal |
| `reset()` | Clear count and all six accumulators, retaining settings |

## Update and reset

The power sequence uses repeated multiplication. The mean update precedes the variance update, which uses both the previous and updated mean.

```pseudocode
PROCEDURE UPDATE(x)
    n ← n + 1
    P1.UPDATE(x)
    x2 ← x × x
    P2.UPDATE(x2)
    x3 ← x2 × x
    P3.UPDATE(x3)
    x4 ← x3 × x
    P4.UPDATE(x4)

    delta ← x − A.VALUE()
    A.UPDATE(delta / n)
    S.UPDATE(delta × (x − A.VALUE()))
END PROCEDURE

PROCEDURE RESET()
    n ← 0
    FOR EACH accumulator IN [P1, P2, P3, P4, A, S] DO
        accumulator.RESET()
    END FOR
END PROCEDURE
```

## Removal and FIFO windows

`revert(x)` removes any sample still present, regardless of insertion position. Power sums subtract its powers; the inverse Welford formula restores the mean and variance sum of the remaining samples.

```pseudocode
PROCEDURE REVERT(x)
    IF n = 0 THEN
        ERROR("Cannot revert from an empty accumulator")
    END IF
    IF n = 1 THEN
        RESET()
        RETURN
    END IF
    n ← n − 1

    P1.REVERT(x)
    x2 ← x × x
    P2.REVERT(x2)
    x3 ← x2 × x
    P3.REVERT(x3)
    x4 ← x3 × x
    P4.REVERT(x4)

    delta ← x − A.VALUE()
    A.REVERT(delta / n)
    S.REVERT(delta × (x − A.VALUE()))
END PROCEDURE
```

In exact arithmetic, with `N` samples before removal, mean $\bar{x}$, and variance sum $S$, removing $x$ gives:

$$
\bar{x}' = \bar{x} - \frac{x - \bar{x}}{N - 1},
\qquad
S' = S - (x - \bar{x})(x - \bar{x}').
$$

These identities depend on the sample set and the removed value, not on when it was added. FIFO removal is therefore supported. Nonfinal removals retain KBN compensation by adding negative contributions; floating-point rounding can still accumulate. The final removal resets all state exactly. Membership is not checked.

## Mean, variance, and raw queries

Define `V(d)` as follows. Check counts before subtracting them, including in implementations with unsigned counts.

```pseudocode
FUNCTION V(d)
    IF n ≤ d THEN
        RETURN NaN
    END IF
    divisor ← FLOAT(n − d)
    RETURN MAX(S.VALUE(), 0) / divisor
END FUNCTION
```

Negative variance sums caused by rounding are clamped to zero. Standard deviation is the square root of the corresponding variance and remains NaN when variance is unavailable.

| Query | Result | Availability |
| --- | --- | --- |
| `n` | Sample count | Always |
| `mean` | `A.VALUE()` | 0.0 when empty |
| `variance` | `V(ddof)` | n > ddof |
| `variance_ddof_0` | `V(0)` | n > 0 |
| `variance_ddof_1` | `V(1)` | n > 1 |
| `standard_deviation` | `SQRT(V(ddof))` | n > ddof |
| `standard_deviation_ddof_0` | `SQRT(V(0))` | n > 0 |
| `standard_deviation_ddof_1` | `SQRT(V(1))` | n > 1 |
| `x1_sum`, `x2_sum`, `x3_sum`, `x4_sum` | `P1.VALUE()` through `P4.VALUE()` | 0.0 when empty |
| `x1`, `x2`, `x3`, `x4` | Corresponding power sum divided by n | NaN when empty |

`mean` is the compensated Welford mean; `x1` is the mean obtained from the first power sum. They are mathematically equal but can differ in their final rounded bits. Fixed-ddof queries do not modify the instance's `ddof`.

## Conversion to central moments

Higher-moment queries first convert raw sums to population central moments. The internal `UNAVAILABLE` result becomes NaN in public queries.

```pseudocode
FUNCTION CENTRAL_MOMENTS()
    IF n < 2 THEN
        RETURN UNAVAILABLE
    END IF
    mu1 ← P1.VALUE() / n
    r ← mu1 × mu1
    meanX2 ← P2.VALUE() / n
    mu2 ← meanX2 − r
    IF mu2 ≤ 1e−14 × meanX2 THEN
        RETURN UNAVAILABLE
    END IF
    r ← r × mu1
    mu3 ← P3.VALUE() / n − r − 3 × mu1 × mu2
    r ← r × mu1
    mu4 ← P4.VALUE() / n − r − 6 × mu2 × mu1 × mu1 − 4 × mu3 × mu1
    RETURN (mu2, mu3, mu4)
END FUNCTION
```

The relative guard detects cancellation or a nonpositive second central moment. It is independent of the `ddof` setting and applies to every skewness and kurtosis variant.

For an available conversion, define:

$$
g_1 = \frac{\mu_3}{\mu_2\sqrt{\mu_2}},
\qquad
\beta_2 = \frac{\mu_4}{\mu_2^2},
\qquad
G_2 = \frac{(n^2-1)\beta_2 - 3(n-1)^2}{(n-2)(n-3)}.
$$

Count products in these formulas are evaluated in floating-point arithmetic after conversion, as described in the [pseudocode conventions](streaming_kbn.md#pseudocode-conventions).

### Explicit skewness and kurtosis variants

All variants return NaN when the conversion fails or the minimum sample count is unmet. Explicit variants ignore the instance's `bias` and `fisher` settings.

| Query | Formula | Minimum count | Convention |
| --- | --- | --- | --- |
| `skewness_moment` | g₁ | 2 | Population moment skewness |
| `skewness_fisher` | g₁ × √(n(n − 1)) / (n − 2) | 3 | Adjusted Fisher–Pearson skewness |
| `skewness_sample` | g₁ × n² / ((n − 1)(n − 2)) | 3 | PerformanceAnalytics “sample” skewness |
| `kurtosis_moment` | β₂ | 2 | Pearson moment kurtosis |
| `kurtosis_excess` | β₂ − 3 | 2 | Excess moment kurtosis |
| `kurtosis_sample_excess` | G₂ | 4 | Bias-corrected excess kurtosis |
| `kurtosis_sample` | G₂ + 3 | 4 | Bias-corrected Pearson kurtosis |
| `kurtosis_sample_corrected` | β₂ × (n² − 1) / ((n − 2)(n − 3)) | 4 | PerformanceAnalytics “sample” kurtosis |

### Setting-selected queries

| bias | fisher | `skewness` selects | `kurtosis` selects |
| --- | --- | --- | --- |
| true | true | `skewness_moment` | `kurtosis_excess` |
| true | false | `skewness_moment` | `kurtosis_moment` |
| false | true | `skewness_fisher` | `kurtosis_sample_excess` |
| false | false | `skewness_fisher` | `kurtosis_sample` |

These choices correspond to the SciPy skewness and kurtosis conventions, subject to this package's sample-size and cancellation guards. The PerformanceAnalytics “sample” variants are separate explicit queries.

## Usage and numerical limits

```pseudocode
moments ← CREATE RawMomentsKleinKBN(ddof: 1, bias: true, fisher: true)
FOR EACH x IN [1, 2, 3, 4] DO
    moments.UPDATE(x)
END FOR
OUTPUT READ moments.mean                 // 2.5
OUTPUT READ moments.variance_ddof_0       // 1.25
OUTPUT READ moments.variance_ddof_1       // approximately 1.6666666666666667
OUTPUT READ moments.skewness             // 0.0
OUTPUT READ moments.kurtosis             // approximately −1.36

moments.REVERT(1)                         // remove the oldest sample
OUTPUT READ moments.n                    // 3
OUTPUT READ moments.mean                 // 3.0
```

For streaming FIFO usage, see the [rolling-window pattern](streaming_kbn.md#rolling-window-pattern).

For `[1e8, 1e8 + 1, 1e8 + 2]`, Welford population variance is approximately `2/3`, but raw-power conversion loses the second central moment and higher-moment queries return NaN. A finite result can also have reduced accuracy before the cancellation guard triggers. [CentralMomentsKleinKBN](central_moments_klein_kbn.md) avoids this conversion for skewness and kurtosis.

Inputs and intermediate products must be finite. KBN compensation cannot recover bits lost when forming powers or subtracting nearly equal raw moments, and does not guarantee exact or indefinitely accurate removal.

## References

- Higham, N. J. (1993). "The accuracy of floating point summation". *SIAM Journal on Scientific Computing*, 14(4), 783–799.
- Klein, A. (2006). "A generalized Kahan–Babuška-Summation-Algorithm". *Computing*, 76(3–4), 279–293.
- Welford, B. P. (1962). "Note on a method for calculating corrected sums of squares and products". *Technometrics*, 4(3), 419–420.
- Cook, J. D. [Skewness and kurtosis](https://www.johndcook.com/skewness_kurtosis.html).
- Kuiperzone. [Compensated-Accumulators](https://github.com/kuiperzone/Compensated-Accumulators).
- NumPy issue #8786 — [Badly conditioned sum](https://github.com/numpy/numpy/issues/8786).
