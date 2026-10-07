# Klein second-order KBN accumulator

`KleinKBNAccumulator` maintains a compensated floating-point sum. It is the numerical building block for all other streaming KBN classes. For language names, source links, and shared conventions, see the [package overview](streaming_kbn.md).

## Why compensate a sum?

Each floating-point addition rounds to the available significand. Adding small values to a large running total can lose their contribution. Cancellation can then expose these lost contributions in the final result.

The condition number of a sum is $\sum_i |x_i| / |\sum_i x_i|$, for a nonzero total. A large value means the result is sensitive to small errors. Compensation reduces rounding error; it cannot remove the sensitivity of the underlying problem.

Peters' example illustrates the benefit:

| Method applied to `[1, 1e100, 1, −1e100]` | Result |
| --- | --- |
| Exact sum | 2.0 |
| Naive summation | 0.0 |
| Standard Kahan summation | 0.0 |
| KBN and Klein second-order KBN | 2.0 |

### Algorithm progression

Kahan's algorithm maintains one correction, feeding it into the next addition. Neumaier's KBN algorithm branches on the larger operand to recover the residual of each addition and accumulates those residuals separately. Klein's second-order variant also compensates the accumulation of those residuals.

| Algorithm | Running state | Result |
| --- | --- | --- |
| Naive | Primary sum | Primary sum |
| Kahan | Primary sum and correction fed into the next addition | Primary sum |
| KBN | Primary sum and accumulated residuals | Sum plus correction |
| Klein second-order KBN | Primary sum and two correction levels | `(sum + cs) + ccs` |

## State and public API

| State | Initial value | Purpose |
| --- | --- | --- |
| `sum` | 0.0 | Primary running sum |
| `cs` | 0.0 | Sum of first-level residuals |
| `ccs` | 0.0 | Accumulated residuals from adding to `cs` |

| Operation/query | Behavior |
| --- | --- |
| `update(x)` | Add `x` using two compensation levels |
| `revert(x)` | Add `−x` using the same algorithm |
| `set(x)` | Replace the total with `x` and clear both corrections |
| `reset()` | Set all three state values to zero |
| `value` | Read `(sum + cs) + ccs` |

There is no sample count, mean, or empty-removal check. Every operation takes O(1) time and the state takes O(1) storage.

## Algorithm

```pseudocode
PROCEDURE UPDATE(x)
    s ← sum
    t ← s + x
    IF ABS(s) ≥ ABS(x) THEN
        c ← (s − t) + x
    ELSE
        c ← (x − t) + s
    END IF
    sum ← t

    first ← cs
    t ← first + c
    IF ABS(first) ≥ ABS(c) THEN
        second ← (first − t) + c
    ELSE
        second ← (c − t) + first
    END IF
    cs ← t
    ccs ← ccs + second
END PROCEDURE

FUNCTION VALUE()
    RETURN (sum + cs) + ccs
END FUNCTION

PROCEDURE SET(x)
    sum ← x
    cs ← 0
    ccs ← 0
END PROCEDURE

PROCEDURE RESET()
    SET(0)
END PROCEDURE

PROCEDURE REVERT(x)
    UPDATE(−x)
END PROCEDURE
```

The second-level residual is **added** to `ccs` on every update. Overwriting it would discard earlier corrections. The final value uses the left-associated expression `(sum + cs) + ccs`, matching all five implementations; reassociating it can change the rounded result.

## Usage and accuracy

```pseudocode
total ← CREATE KleinKBNAccumulator()
FOR EACH x IN [1, 1e100, 1, −1e100] DO
    total.UPDATE(x)
END FOR
OUTPUT READ total.value                  // 2.0

total.REVERT(1)                           // remove one occurrence
OUTPUT READ total.value                  // 1.0
total.SET(5)                              // replace all accumulated state
total.RESET()                             // value is now 0.0
```

Adding the negation permits removal in any order, provided the caller supplies a value still present in its logical sample set. It does not restore historical internal state or guarantee an exact floating-point inverse. With no count, the accumulator cannot automatically reset after the last removal; use [KleinKBNSummator](klein_kbn_summator.md) when that behavior and a mean are needed.

The sequence `[1e−16, −1e16, 1, 1e−16, −1, −1e−16, −1e−32, 1e16]` exercises the second correction level. The current implementation returns `9.999999999999999e−17`.

Inputs and intermediate calculations are assumed finite. Compensation does not guarantee correctly rounded sums, exact reversal, or indefinitely accurate rolling-window updates.

## References

- <a id="ref-higham93"></a> Higham, N. J. (1993). "The accuracy of floating point summation". *SIAM Journal on Scientific Computing*, 14(4), 783–799. [doi:10.1137/0914050](https://doi.org/10.1137/0914050)
- <a id="ref-kahan65"></a> Kahan, W. (1965). "Further remarks on reducing truncation errors". *Communications of the ACM*, 8(1), 40. [doi:10.1145/363707.363723](https://doi.org/10.1145/363707.363723)
- <a id="ref-neumaier74"></a> Neumaier, A. (1974). "Rundungsfehleranalyse einiger Verfahren zur Summation endlicher Summen". *Zeitschrift für Angewandte Mathematik und Mechanik*, 54(1), 39–51. [doi:10.1002/zamm.19740540106](https://doi.org/10.1002/zamm.19740540106)
- <a id="ref-klein06"></a> Klein, A. (2006). "A generalized Kahan–Babuška-Summation-Algorithm". *Computing*, 76(3–4), 279–293. [doi:10.1007/s00607-005-0139-x](https://doi.org/10.1007/s00607-005-0139-x)
- <a id="ref-wikipedia"></a> Wikipedia. [Kahan summation algorithm](https://en.wikipedia.org/wiki/Kahan_summation_algorithm).
- <a id="ref-2sum"></a> Wikipedia. [2Sum](https://en.wikipedia.org/wiki/2Sum).
- <a id="ref-kuiperzone"></a> Kuiperzone. [Compensated-Accumulators](https://github.com/kuiperzone/Compensated-Accumulators).
- <a id="ref-numpy-issue"></a> NumPy issue #8786 — [Badly conditioned sum](https://github.com/numpy/numpy/issues/8786).
- <a id="ref-peters"></a> Peters' example discussed in the [CPython `math.fsum` implementation](https://github.com/python/cpython/blob/main/Modules/mathmodule.c).
