# Streaming linear regression

`LinearRegressionKleinKBN` fits the ordinary least squares model $y = a + b x$, where $b$ is the slope and $a$ is the intercept. It supports streaming insertion and removal of complete observation pairs, including FIFO removal. For all five source implementations and their API naming, see the [package overview](streaming_kbn.md).

## State and sufficient statistics

| State | Initial value | Purpose |
| --- | --- | --- |
| `n` | 0 | Number of observation pairs |
| `X`, `Y` | Empty raw-moment trackers with ddof = 0, bias = true, fisher = true | Means and population variances of x and y |
| `C` | Zero-valued KBN accumulator | Co-moment Sxy |

The internal trackers are [RawMomentsKleinKBN](raw_moments_klein_kbn.md) instances. Their compensated Welford statistics provide the means and sums of squared deviations. A dedicated [KleinKBNAccumulator](klein_kbn_accumulator.md) tracks the cross-product sum.

$$
S_{xx} = \sum_i (x_i-\bar{x})^2,
\qquad
S_{yy} = \sum_i (y_i-\bar{y})^2,
\qquad
S_{xy} = \sum_i (x_i-\bar{x})(y_i-\bar{y}).
$$

The public formulas are:

$$
b = S_{xy}/S_{xx},
\qquad
a = \bar{y} - b\bar{x},
\qquad
r = S_{xy}/\sqrt{S_{xx}S_{yy}}.
$$

Every update, removal, reset, and query takes O(1) time; internal storage is O(1). There are no public regression settings.

| Operation | Behavior |
| --- | --- |
| `update(x, y)` | Add an observation pair |
| `revert(x, y)` | Remove a present pair; fail if empty; reset after final removal |
| `reset()` | Clear count, both marginal trackers, and the co-moment accumulator |

## Update and reset

The co-moment update uses the means **before** adding the pair. Its contribution is accumulated with KBN compensation.

```pseudocode
PROCEDURE UPDATE(x, y)
    nOld ← n
    n ← nOld + 1
    term ← (READ X.mean − x) × (READ Y.mean − y) × nOld / (nOld + 1)
    C.UPDATE(term)
    X.UPDATE(x)
    Y.UPDATE(y)
END PROCEDURE

PROCEDURE RESET()
    n ← 0
    X.RESET()
    Y.RESET()
    C.RESET()
END PROCEDURE
```

In exact arithmetic, the cross-product identity is:

$$
S_{xy}^{new} = S_{xy}^{old}
+ (\bar{x}_{old}-x)(\bar{y}_{old}-y)\frac{n_{old}}{n_{old}+1}.
$$

## Removal

The marginal trackers are reverted first. The co-moment subtraction then uses the means of the **remaining** observations, which need not equal any historical means from before insertion.

```pseudocode
PROCEDURE REVERT(x, y)
    IF n = 0 THEN
        ERROR("Cannot revert from an empty regression")
    END IF
    IF n = 1 THEN
        RESET()
        RETURN
    END IF
    X.REVERT(x)
    Y.REVERT(y)
    remaining ← n − 1
    term ← (READ X.mean − x) × (READ Y.mean − y) × remaining / (remaining + 1)
    C.REVERT(term)
    n ← remaining
END PROCEDURE
```

For current count $N$ and remaining means $\bar{x}'$, $\bar{y}'$, the inverse identity is:

$$
S_{xy}' = S_{xy}
- (\bar{x}'-x)(\bar{y}'-y)\frac{N-1}{N}.
$$

It depends on the remaining sample set and removed pair, not the pair's insertion position. This permits removing the oldest pair in a FIFO window. The pair must still be present; membership is not validated.

Nonfinal removal retains the compensation in the marginal trackers and the co-moment accumulator. Floating-point rounding can still accumulate, and historical accumulator state is not restored. Final removal resets everything exactly. Empty-removal failures follow the [shared language mapping](streaming_kbn.md#errors-and-validation).

## Public queries and guards

| Query | Result | Empty or degenerate behavior |
| --- | --- | --- |
| `n` | Number of observation pairs | 0 when empty |
| `mean_x`, `mean_y` | Means from X and Y | 0.0 when empty |
| `variance_x`, `variance_y` | Population variances, Sxx / n and Syy / n | NaN when empty |
| `co_moment` | `C.VALUE()`, the unnormalized Sxy | 0.0 when empty |
| `covariance` | Population covariance, Sxy / n | NaN when empty |
| `slope` | Sxy / (variance_x × n) | NaN when n < 2 or variance_x × n = 0 |
| `intercept` | mean_y − slope × mean_x | NaN when slope is NaN |
| `correlation` | Sxy / (standard_deviation_x × standard_deviation_y × n), clamped to [−1, 1] | NaN when n < 2 or the standard-deviation product is zero |

The internal standard deviations use ddof = 0. The implementation evaluates their product and then multiplies by the count, matching the following pseudocode:

```pseudocode
FUNCTION CORRELATION()
    IF n < 2 THEN
        RETURN NaN
    END IF
    product ← (READ X.standard_deviation) × (READ Y.standard_deviation)
    IF product = 0 THEN
        RETURN NaN
    END IF
    r ← C.VALUE() / (product × n)
    RETURN MAX(−1, MIN(1, r))
END FUNCTION
```

Clamping absorbs small excursions beyond the correlation interval from rounding. Unavailable slope and correlation return NaN in every implementation. With one finite pair, population variances and covariance are zero; slope, intercept, and correlation remain NaN. Constant x prevents slope calculation, while constant y can still give a valid zero slope if x varies.

## Usage

```pseudocode
regression ← CREATE LinearRegressionKleinKBN()
FOR EACH (x, y) IN [(1, 3), (2, 5), (3, 7)] DO
    regression.UPDATE(x, y)
END FOR
OUTPUT READ regression.slope             // 2.0
OUTPUT READ regression.intercept         // 1.0
OUTPUT READ regression.correlation       // approximately 1.0
OUTPUT READ regression.covariance        // approximately 4/3

regression.REVERT(1, 3)                   // remove the oldest pair
OUTPUT READ regression.n                 // 2
OUTPUT READ regression.slope             // 2.0
```

For a rolling regression, retain complete pairs in a queue and use the [rolling-window pattern](streaming_kbn.md#rolling-window-pattern).

Mean and variance come from Welford updates rather than raw-power conversion, so the higher-moment cancellation guard does not participate in regression queries. The internal raw-moment trackers still calculate powers through x⁴ and y⁴. Inputs and intermediate calculations are assumed finite; compensation does not guarantee exact subtraction or indefinitely accurate window shifts.

## References

- Cook, J. D. [Running regression](https://www.johndcook.com/running_regression.html).
- Welford, B. P. (1962). "Note on a method for calculating corrected sums of squares and products". *Technometrics*, 4(3), 419–420.
- Klein, A. (2006). "A generalized Kahan–Babuška-Summation-Algorithm". *Computing*, 76(3–4), 279–293.
- Higham, N. J. (1993). "The accuracy of floating point summation". *SIAM Journal on Scientific Computing*, 14(4), 783–799.
- Kuiperzone. [Compensated-Accumulators](https://github.com/kuiperzone/Compensated-Accumulators).
