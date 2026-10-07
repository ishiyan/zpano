# Streaming KBN package

The package computes streaming sums, means, variance, skewness, kurtosis, and ordinary least squares regression with Klein second-order Kahan–Babuška–Neumaier (KBN) compensation. Python is the reference implementation; Go, Rust, Zig, and TypeScript implement the same algorithms without external dependencies.

## Classes and dependencies

| Class | Computes | Internal dependencies |
| --- | --- | --- |
| [KleinKBNAccumulator](klein_kbn_accumulator.md) | Compensated sum, with replacement and subtraction | None |
| [KleinKBNSummator](klein_kbn_summator.md) | Compensated sum, sample count, mean | One accumulator |
| [RawMomentsKleinKBN](raw_moments_klein_kbn.md) | Raw power sums, mean, variance, skewness and kurtosis variants | Six accumulators |
| [CentralMomentsKleinKBN](central_moments_klein_kbn.md) | Mean, variance, skewness, kurtosis from central moments | Four accumulators |
| [LinearRegressionKleinKBN](linear_regression_klein_kbn.md) | Means, variances, co-moment, covariance, slope, intercept, correlation | Two raw-moment trackers and one accumulator |

Each class uses O(1) storage and O(1) time per update, removal, or query. Samples are not stored internally. A caller maintaining a rolling window must retain its samples separately, requiring O(window size) storage. There is no batch or merge API.

## Pseudocode conventions

All algorithm and usage blocks in these pages are language-neutral pseudocode:

- `←` assigns a value; `×`, `/`, and `SQRT` perform arithmetic.
- `PROCEDURE`, `FUNCTION`, `IF`, and `FOR EACH` describe control flow.
- `CREATE Class(...)` constructs an empty instance. Named settings describe the intended configuration, independent of constructor syntax.
- `READ object.query` reads a query through the language's property or getter method.
- `object.UPDATE(x)`, `REVERT(x)`, `SET(x)`, and `RESET()` invoke operations. `VALUE()` reads an accumulator's compensated sum.
- `ERROR(message)` signals the language-specific failure listed below. `UNAVAILABLE` is an internal marker; unavailable public numeric results are **NaN**.

State names and query names are descriptive labels. Arithmetic uses binary64 values, with ordinary precedence and left association for chained additions and multiplications. Parentheses and the sequence of intermediate calculations matter: algebraically equivalent formulas can round differently.

Sample counts are integers. In moment formulas, convert counts to floating point **before** products such as `n × n` or `n × (n − 1)`. For variance, first check `n ≤ ddof`, then subtract the integer counts and convert the difference to floating point. This avoids integer product overflow and unsigned subtraction underflow. Counts must remain representable in the implementation's count type; conversion of very large counts to binary64 can lose integer precision.

## Shared behavior

### Inputs and removal

Inputs are assumed to be finite binary64 values, with finite intermediate calculations. The implementations do not validate NaN or infinity and do not recover after intermediate overflow. Raw moments form powers through the fourth power, including inside the regression trackers.

`REVERT` can remove any sample still present, including the oldest in a FIFO window. The formulas are independent of insertion position in exact arithmetic; floating-point results can depend on operation order. Membership is not checked, so callers must remove only samples that are present, with the correct multiplicity. Regression removes complete `(x, y)` pairs.

Removing the final sample from a summator, either moment tracker, or regression calls `RESET()`, restoring an empty state exactly. The plain accumulator has no sample count and implements removal by adding the negated value.

Compensation improves accuracy but does not guarantee a correctly rounded result or exact reversal. Repeated rolling-window shifts can accumulate rounding error. Raw-moment removal retains compensation until the final reset. Central-moment removal writes restored moments with `SET`, clearing compensation at each removal.

### Empty and unavailable results

| Query | Empty result |
| --- | --- |
| Sample count `n` | 0 |
| Accumulator/summator `value` | 0.0 |
| Summator `mean` | NaN |
| Moment tracker `mean`; regression `mean_x`, `mean_y` | 0.0 |
| Raw power sums `x1_sum` through `x4_sum` | 0.0 |
| Raw moments `x1` through `x4` | NaN |
| Variance, standard deviation, skewness, kurtosis | NaN |
| Regression `co_moment` | 0.0 |
| Regression covariance, slope, intercept, correlation | NaN |

Additional sample-size and denominator guards are given on each class page. This package uses NaN for unavailable numeric results in all five languages.

### Moment settings

Both moment trackers have mutable `ddof`, `bias`, and `fisher` settings. Queries do not change settings or accumulated state. `RESET()` clears the samples and retains the settings.

| Setting | Conventional default | Meaning |
| --- | --- | --- |
| `ddof` | 1 | Nonnegative integer subtracted from the variance divisor; 0 gives population variance, 1 gives sample variance |
| `bias` | true | Use population standardized moments; false applies skewness/kurtosis bias corrections |
| `fisher` | true | Report excess kurtosis; false reports Pearson kurtosis, whose normal-distribution reference value is 3 |

`ddof` affects variance and standard deviation, not skewness or kurtosis. `fisher` affects kurtosis, not skewness. Fixed variant queries on raw moments ignore the corresponding instance settings.

## Language API mapping

The class pages use the names ending in `KBN`. Rust and TypeScript spell that suffix `Kbn`, for example `RawMomentsKleinKbn` and `KleinKbnSummator`.

| Convention | Python | Go | Rust | Zig | TypeScript |
| --- | --- | --- | --- | --- | --- |
| Update | `update(x)` | `Update(x)` | `update(x)` | `update(x)` | `update(x)` |
| Count | `n` property | `N()` | `n()` | `n()` | `n` getter |
| Example query | `standard_deviation` property | `StandardDeviation()` | `standard_deviation()` | `standardDeviation()` | `standardDeviation` getter |
| Fixed-ddof query | `variance_ddof_0` | `VarianceDdof0()` | `variance_ddof_0()` | `varianceDdof0()` | `varianceDdof0` |
| Settings | Public attributes | `Ddof()` / `SetDdof()` etc. | `ddof()` / `set_ddof()` etc. | Public fields | Properties |

### Construction

| Language | Moment trackers | Other classes |
| --- | --- | --- |
| Python | Constructor defaults to `(1, true, true)` | No-argument constructors |
| Go | `NewRawMomentsKleinKBN(ddof, bias, fisher)` and `NewCentralMomentsKleinKBN(ddof, bias, fisher)` require all settings | Zero value for the accumulator; `NewKleinKBNSummator()` and `NewLinearRegressionKleinKBN()` |
| Rust | `::new(ddof, bias, fisher)`; `Default` uses `(1, true, true)` | `::new()` or `Default` |
| Zig | Struct initialization; field defaults are `(1, true, true)` | Struct initialization with defaults |
| TypeScript | Constructor defaults to `(1, true, true)` | No-argument constructors |

In Go, the summator's zero value also works. A zero-valued moment tracker uses `(0, false, false)`; use its constructor to select the intended settings. Construct regression through its constructor to initialize its internal tracker pointers.

### Errors and validation

| Situation | Python | Go | Rust | Zig | TypeScript |
| --- | --- | --- | --- | --- | --- |
| Empty removal from summator, moments, or regression | `ValueError` | `panic` | `panic!` | `error.EmptyRevert`; removal returns an error union | `Error` exception |
| Invalid `ddof` | Constructor requires a nonnegative value of exact type `int`; later attribute assignment is not revalidated | Constructor and setter reject negative `int` values with `panic` | `usize` excludes negative/fractional values | `u32` excludes negative/fractional values | Constructor and setter require a nonnegative integer, otherwise throw `Error` |

The empty-removal messages are `Cannot revert from an empty summator`, `Cannot revert from an empty accumulator` (both moment trackers), and `Cannot revert from an empty regression`. The plain accumulator has no empty-removal error.

## Source files

| Class | Python | Go | Rust | Zig | TypeScript |
| --- | --- | --- | --- | --- | --- |
| Accumulator | [klein_kbn_accumulator.py](../../py/streaming_kbn/klein_kbn_accumulator.py) | [kleinkbnaccumulator.go](../../go/streamingkbn/kleinkbnaccumulator.go) | [klein_kbn_accumulator.rs](../../rs/src/streaming_kbn/klein_kbn_accumulator.rs) | [klein_kbn_accumulator.zig](../../zig/src/streaming_kbn/klein_kbn_accumulator.zig) | [klein-kbn-accumulator.ts](../../ts/streaming-kbn/klein-kbn-accumulator.ts) |
| Summator | [klein_kbn_summator.py](../../py/streaming_kbn/klein_kbn_summator.py) | [kleinkbnsummator.go](../../go/streamingkbn/kleinkbnsummator.go) | [klein_kbn_summator.rs](../../rs/src/streaming_kbn/klein_kbn_summator.rs) | [klein_kbn_summator.zig](../../zig/src/streaming_kbn/klein_kbn_summator.zig) | [klein-kbn-summator.ts](../../ts/streaming-kbn/klein-kbn-summator.ts) |
| Raw moments | [raw_moments_klein_kbn.py](../../py/streaming_kbn/raw_moments_klein_kbn.py) | [rawmomentskleinkbn.go](../../go/streamingkbn/rawmomentskleinkbn.go) | [raw_moments_klein_kbn.rs](../../rs/src/streaming_kbn/raw_moments_klein_kbn.rs) | [raw_moments_klein_kbn.zig](../../zig/src/streaming_kbn/raw_moments_klein_kbn.zig) | [raw-moments-klein-kbn.ts](../../ts/streaming-kbn/raw-moments-klein-kbn.ts) |
| Central moments | [central_moments_klein_kbn.py](../../py/streaming_kbn/central_moments_klein_kbn.py) | [centralmomentskleinkbn.go](../../go/streamingkbn/centralmomentskleinkbn.go) | [central_moments_klein_kbn.rs](../../rs/src/streaming_kbn/central_moments_klein_kbn.rs) | [central_moments_klein_kbn.zig](../../zig/src/streaming_kbn/central_moments_klein_kbn.zig) | [central-moments-klein-kbn.ts](../../ts/streaming-kbn/central-moments-klein-kbn.ts) |
| Regression | [linear_regression_klein_kbn.py](../../py/streaming_kbn/linear_regression_klein_kbn.py) | [linearregressionkleinkbn.go](../../go/streamingkbn/linearregressionkleinkbn.go) | [linear_regression_klein_kbn.rs](../../rs/src/streaming_kbn/linear_regression_klein_kbn.rs) | [linear_regression_klein_kbn.zig](../../zig/src/streaming_kbn/linear_regression_klein_kbn.zig) | [linear-regression-klein-kbn.ts](../../ts/streaming-kbn/linear-regression-klein-kbn.ts) |

### Package entry files

These files document the package and expose the five public classes:

| Language | File | Role |
| --- | --- | --- |
| Python | [__init__.py](../../py/streaming_kbn/__init__.py) | Package documentation and class imports |
| Go | [doc.go](../../go/streamingkbn/doc.go) | Package documentation; exported types reside in the same package |
| Rust | [mod.rs](../../rs/src/streaming_kbn/mod.rs) | Public submodules and class re-exports; shared helpers compiled only for tests |
| Zig | [streaming_kbn.zig](../../zig/src/streaming_kbn/streaming_kbn.zig) | Module and type re-exports, with imports wired by [build.zig](../../zig/build.zig) |
| TypeScript | [index.ts](../../ts/streaming-kbn/index.ts) | Package documentation and class re-exports |

Python tests use `test_*.py`, Go tests use `*_test.go`, and TypeScript uses Jasmine `*.spec.ts`. Rust and Zig keep tests in the implementation files. Test helpers include compensated reference summation and seeded random inputs; they are not public package APIs.

## Rolling-window pattern

The following pattern applies to the summator and both moment trackers. For regression, retain pairs and call `UPDATE(x, y)` and `REVERT(x, y)`.

```pseudocode
tracker ← CREATE RawMomentsKleinKBN(ddof: 1, bias: true, fisher: true)
window ← CREATE empty FIFO queue
capacity ← 100

FOR EACH x IN stream DO
    IF SIZE(window) = capacity THEN
        oldest ← REMOVE FRONT(window)
        tracker.REVERT(oldest)
    END IF
    tracker.UPDATE(x)
    APPEND BACK(window, x)
    OUTPUT READ tracker.mean, READ tracker.variance
END FOR
```

For higher moments of data with a large mean relative to its spread, use central moments to avoid raw-power conversion cancellation. Use raw moments when raw sums or its explicit statistical variants are needed. Both support FIFO removal with the rounding limitations described above.
