# Compensated sum, count, and mean

`KleinKBNSummator` combines a [KleinKBNAccumulator](klein_kbn_accumulator.md) with a sample count. It adds a mean query, checks for removal from an empty sample set, and resets exactly when the final sample is removed. It is available in all five languages; see the [package overview](streaming_kbn.md) for source files and API naming.

## State and public API

The initial state is `n = 0` and one zero-valued accumulator, `total`. Storage and the time for every operation or query are O(1).

| Operation/query | Behavior |
| --- | --- |
| `update(x)` | Count the sample; add it to the accumulator when `x ≠ 0` |
| `revert(x)` | Remove a present sample in any order; fail if empty; reset if removing the final sample |
| `reset()` | Clear count and accumulator |
| `n` | Number of samples, including zeros |
| `value` | Compensated sum; 0.0 when empty |
| `mean` | `value / n`; NaN when empty |

## Algorithm

```pseudocode
PROCEDURE UPDATE(x)
    n ← n + 1
    IF x ≠ 0 THEN
        total.UPDATE(x)
    END IF
END PROCEDURE

PROCEDURE REVERT(x)
    IF n = 0 THEN
        ERROR("Cannot revert from an empty summator")
    END IF
    IF n = 1 THEN
        RESET()
        RETURN
    END IF
    n ← n − 1
    IF x ≠ 0 THEN
        total.REVERT(x)
    END IF
END PROCEDURE

PROCEDURE RESET()
    n ← 0
    total.RESET()
END PROCEDURE

FUNCTION VALUE()
    RETURN total.VALUE()
END FUNCTION

FUNCTION MEAN()
    IF n = 0 THEN
        RETURN NaN
    END IF
    RETURN VALUE() / n
END FUNCTION
```

Zero samples increase or decrease the count without changing the accumulator's corrections. The singleton check precedes this zero shortcut, so removing a final zero sample also clears any rounding residual left by earlier removals.

The empty-removal failure maps to an exception, panic, or Zig error union as described in the [shared error table](streaming_kbn.md#errors-and-validation). Membership is not validated: callers must remove a sample still present, with the correct multiplicity.

## Usage

```pseudocode
summary ← CREATE KleinKBNSummator()
summary.UPDATE(2)
summary.UPDATE(0)
summary.UPDATE(4)
OUTPUT READ summary.n                    // 3
OUTPUT READ summary.value                // 6.0
OUTPUT READ summary.mean                 // 2.0

summary.REVERT(2)                         // remove the oldest sample
OUTPUT READ summary.mean                 // 2.0, from [0, 4]
summary.REVERT(4)
summary.REVERT(0)                         // final sample resets everything
OUTPUT READ summary.n                    // 0
OUTPUT READ summary.value                // exactly 0.0
OUTPUT READ summary.mean                 // NaN
```

A cancellation example adds `[0.1, 1e16, 1e32, 1e48]`, then removes `[1e16, 0.1, 1e32, 1e48]`. After the final removal, the count and sum are exactly zero and the mean is NaN. The final reset guarantees this empty state; it does not make the preceding subtractions exact.

For a FIFO stream, retain a queue and follow the [rolling-window pattern](streaming_kbn.md#rolling-window-pattern). Inputs and intermediate calculations must be finite. Repeated removals can accumulate rounding error even though nonfinal removals retain KBN compensation.

## References

- [Accumulator algorithm and references](klein_kbn_accumulator.md#references).
- Klein, A. (2006). “A generalized Kahan–Babuška-Summation-Algorithm.” *Computing*, 76(3–4), 279–293. [doi:10.1007/s00607-005-0139-x](https://doi.org/10.1007/s00607-005-0139-x)
