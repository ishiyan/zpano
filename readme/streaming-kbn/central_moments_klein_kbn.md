# Pébay central-moment streaming statistics

`CentralMomentsKleinKBN` computes mean, variance, skewness, and kurtosis directly from running central moments, using Pébay's formulas with Klein second-order KBN accumulation. It avoids the raw-power conversion cancellation that affects higher moments when the mean is large relative to the spread. See the [package overview](streaming_kbn.md) for all five implementations and API conventions.

## State and settings

| State | Initial value | Purpose |
| --- | --- | --- |
| `n` | 0 | Sample count |
| `M1` | Zero-valued KBN accumulator | Running mean |
| `M2` | Zero-valued KBN accumulator | Sum of squared deviations |
| `M3` | Zero-valued KBN accumulator | Sum of cubed deviations |
| `M4` | Zero-valued KBN accumulator | Sum of fourth powers of deviations |

`M1` stores the mean. `M2`, `M3`, and `M4` store **sums**, not normalized moments:

$$
M_k = \sum_i (x_i - \bar{x})^k,
\qquad
\mu_k = M_k / n
\quad (k = 2, 3, 4).
$$

Every accumulator is a [KleinKBNAccumulator](klein_kbn_accumulator.md). Every operation and query takes O(1) time; storage is O(1).

The conventional settings are `ddof = 1`, `bias = true`, `fisher = true`. They are mutable, and reset retains them. `ddof` must be a nonnegative integer; [construction and validation](streaming_kbn.md#language-api-mapping) differ by language.

| Operation | Behavior |
| --- | --- |
| `update(x)` | Add a sample using the compensated central-moment formulas |
| `revert(x)` | Remove a present sample; fail if empty; reset after final removal |
| `reset()` | Clear count and all four accumulators, retaining settings |

## Forward update and reset

`m2` and `m3` are captured before updating their accumulators. Counts used in arithmetic products below are converted to floating point before multiplication.

```pseudocode
PROCEDURE UPDATE(x)
    nOld ← n
    nNew ← nOld + 1
    n ← nNew
    delta ← x − M1.VALUE()
    d ← delta / nNew
    d2 ← d × d
    term ← delta × d × nOld
    m2 ← M2.VALUE()
    m3 ← M3.VALUE()

    M1.UPDATE(d)
    M4.UPDATE(term × d2 × (nNew × nNew − 3 × nNew + 3)
              + 6 × d2 × m2 − 4 × d × m3)
    M3.UPDATE(term × d × (nNew − 2) − 3 × d × m2)
    M2.UPDATE(term)
END PROCEDURE

PROCEDURE RESET()
    n ← 0
    FOR EACH accumulator IN [M1, M2, M3, M4] DO
        accumulator.RESET()
    END FOR
END PROCEDURE
```

Updating M4 before M3 and M2 preserves the dependency on the previous central sums.

## Inverse removal

Any sample still present can be removed, including the oldest in a FIFO window. The algorithm reconstructs the moments of the remaining samples. In the notation below, `nNew` is the current count before removal and `nOld` is the remaining count, viewing removal as the inverse of insertion.

```pseudocode
PROCEDURE REVERT(x)
    IF n = 0 THEN
        ERROR("Cannot revert from an empty accumulator")
    END IF
    IF n = 1 THEN
        RESET()
        RETURN
    END IF
    nNew ← n
    nOld ← nNew − 1
    m1New ← M1.VALUE()
    m2New ← M2.VALUE()
    m3New ← M3.VALUE()
    m4New ← M4.VALUE()

    m1Old ← (nNew × m1New − x) / nOld
    delta ← x − m1Old
    d ← delta / nNew
    d2 ← d × d
    term ← delta × d × nOld
    m2Old ← m2New − term
    m3Old ← m3New − (term × d × (nNew − 2) − 3 × d × m2Old)
    m4Old ← m4New − (term × d2 × (nNew × nNew − 3 × nNew + 3)
                     + 6 × d2 × m2Old − 4 × d × m3Old)

    n ← nOld
    M1.SET(m1Old)
    M2.SET(m2Old)
    M3.SET(m3Old)
    M4.SET(m4Old)
END PROCEDURE
```

The restored values must be calculated in dependency order: mean, second sum, third sum, then fourth sum. `SET` clears each accumulator's compensation terms. Subsequent updates build compensation from these restored values; repeated removals can accumulate rounding error, especially for data with a large mean.

The identities do not depend on insertion position in exact arithmetic. Floating-point removal is not an exact inverse, and the implementation does not check sample membership. The final removal resets count and all accumulators exactly. See the [error mapping](streaming_kbn.md#errors-and-validation) for empty removal.

## Public queries

Define `m2 = M2.VALUE()`, `m3 = M3.VALUE()`, and `m4 = M4.VALUE()`. For valid positive variance and sufficient samples, the standardized moments are:

$$
g_1 = \frac{\sqrt{n}\,m3}{m2\sqrt{m2}},
\qquad
\beta_2 = \frac{n\,m4}{m2^2},
\qquad
G_2 = \frac{(n^2-1)\beta_2 - 3(n-1)^2}{(n-2)(n-3)}.
$$

| Query | Formula or behavior | Unavailable result |
| --- | --- | --- |
| `n` | Sample count | Always available |
| `mean` | `M1.VALUE()` | 0.0 when empty |
| `variance` | `MAX(m2, 0) / (n − ddof)` | NaN when n ≤ ddof |
| `standard_deviation` | Square root of variance | NaN when variance is unavailable |
| `skewness` | g₁ if bias is true; g₁ × √(n(n − 1)) / (n − 2) otherwise | NaN when n < 2 or m2 ≤ 0; additionally n < 3 when bias is false |
| `kurtosis` | Selected by the table below | NaN when n < 2 or m2 ≤ 0; additionally n < 4 when bias is false |

Variance checks `n ≤ ddof` before subtracting counts, then converts the difference to floating point. Negative second sums from rounding are clamped to zero for variance; skewness and kurtosis require a strictly positive second sum. There is no relative raw-power cancellation guard because these queries use central sums directly.

| bias | fisher | `kurtosis` |
| --- | --- | --- |
| true | true | β₂ − 3 |
| true | false | β₂ |
| false | true | G₂ |
| false | false | G₂ + 3 |

These selections follow the same statistical conventions as the setting-selected queries on raw moments. `ddof` affects variance and standard deviation only; `fisher` affects kurtosis only. Central moments has no separate fixed-ddof, raw-power, or explicit skewness/kurtosis variant queries.

## Usage and choice of tracker

```pseudocode
moments ← CREATE CentralMomentsKleinKBN(ddof: 0, bias: true, fisher: true)
FOR EACH x IN [1e8, 1e8 + 1, 1e8 + 2] DO
    moments.UPDATE(x)
END FOR
OUTPUT READ moments.mean                 // 100000001.0
OUTPUT READ moments.variance             // approximately 2/3
OUTPUT READ moments.skewness             // 0.0
OUTPUT READ moments.kurtosis             // −1.5

moments.REVERT(1e8)                       // FIFO removal
OUTPUT READ moments.n                    // 2
```

For FIFO streams, retain a queue as shown in the [rolling-window pattern](streaming_kbn.md#rolling-window-pattern).

| Aspect | RawMomentsKleinKBN | CentralMomentsKleinKBN |
| --- | --- | --- |
| Mean and variance | Compensated Welford tracker | Compensated central-moment updates |
| Higher moments | Convert raw power sums; cancellation can reduce accuracy or produce NaN | Use central sums directly; avoids that conversion |
| Nonfinal removal | Retains compensation | Clears compensation on restored moments |
| Query choices | Raw sums, fixed-ddof queries, explicit moment variants, setting-selected queries | Setting-selected queries |
| FIFO removal | Supported | Supported |

Use central moments for higher moments when the mean is large relative to the spread. Use [raw moments](raw_moments_klein_kbn.md) when raw sums or explicit statistical variants are needed. Neither tracker guarantees exact reversal or indefinitely accurate rolling-window shifts. Inputs and all intermediate calculations must remain finite.

## References

- Pébay, P. (2008). "Formulas for robust, one-pass parallel computation of covariances and arbitrary-order statistical moments". *Sandia Report SAND2008-6212*.
- Cook, J. D. [Skewness and kurtosis](https://www.johndcook.com/skewness_kurtosis.html).
- Klein, A. (2006). "A generalized Kahan–Babuška-Summation-Algorithm". *Computing*, 76(3–4), 279–293.
- Kuiperzone. [Compensated-Accumulators](https://github.com/kuiperzone/Compensated-Accumulators).
- Higham, N. J. (1993). "The accuracy of floating point summation". *SIAM Journal on Scientific Computing*, 14(4), 783–799.
