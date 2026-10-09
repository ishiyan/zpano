# Performance measures port review and fixes

Reviewed and corrected on 2026-10-09 against `py/performance/`, including
`Measures`, streaming `core` helpers, test mappings, reference-data translation,
and `.opencode/skills/performance-measures-architecture/SKILL.md`. The
experimental Python resampler and deprecated packages are excluded from the
port scope.

The four ports reproduce the reference formulas and rolling-state updates on
valid finite inputs. The confirmed formula, validation, fixture, documentation,
and Rust summation issues have been corrected. The remaining numerical limit
is platform-dependent rounding of transcendental functions in large annualized
outputs, which requires relative rather than absolute tolerance.

## Corrected behavior

| Issue | Correction across all five languages |
|---|---|
| Gain-to-pain divided mean return by the sum of losses | Uses compensated sum of returns / compensated sum of loss magnitudes; preserves NaN without losses |
| Modified information ratio used arithmetic active mean for its branch | Returns IR if annualized geometric active premium is positive, otherwise −IR |
| Normality accepted NaN confidence and skipped validation without data | Validates `0 < confidence < 1` before the unavailable-statistic fallback |
| Bias ratio accepted NaN multiplier | Rejects nonpositive and NaN multipliers before checking data; positive infinity remains accepted |
| Reward-to-conditional-drawdown lacked confidence validation | Validates `0 < confidence < 1` before reading state; rejects NaN, infinities, and finite out-of-range inputs |
| Volatility and variability fixtures had reversed labels | Labels and test lookups now match the measures; removed accidental extra trailing zeros |
| Rust used ordinary sums for three tails | Uses CPython 3.12+ compensated summation for both Rachev tails and conditional-drawdown tail |

The Python drawdown docstring now describes R's percent convention and its
actual example output. The down-number docstring now includes zero benchmark
returns. These two calculations retain their R-compatible behavior.

## Public API compatibility

Reward-to-conditional-drawdown now reports invalid confidence through each
language's error API:

- Python and TypeScript raise their existing exception types.
- Go `RewardToConditionalDrawdown(confidence)` returns `(float64, error)`.
- Rust `reward_to_conditional_drawdown(confidence)` returns `Result<f64, String>`.
- Zig `rewardToConditionalDrawdown(confidence)` adds `error.InvalidArgument` to
  its existing allocation error union.

All repository call sites and generic test evaluators have been updated. Valid
confidence on an empty instance still produces NaN. Invalid confidence on an
empty instance now produces an error for both normality and reward-to-drawdown.
Invalid-confidence errors use `confidence must be between 0 and 1` in
Python, Go, Rust, and TypeScript, and `error.InvalidArgument` in Zig.

Gain-to-pain numerical results change, and modified information ratio can
change sign when arithmetic and geometric active performance disagree. Fixture
identifiers retain their spelling, but now contain the correctly named data.

## Regressions and original failure examples

### Gain-to-pain normalization

For `[0.1, -0.05]`, the previous result was approximately `0.5`; the corrected
sum/sum ratio is `1.0`. Repeating the pair preserves `1.0`, including with rolling
windows of two and four observations. Regression tests cover repetition,
rolling eviction, reset, and the no-loss NaN policy in every language.

The Bacon gain-to-pain fixtures were calculated independently using exact
rational sums of the input binary floats, with numerator and denominator each
rounded to float before division. They were then translated into the four port
languages with `notes/performance/gen.py`. Expected values were not generated
by calling the measure under test.

### Modified information ratio sign source

For P=1, portfolio returns `[0.5, -0.3]`, and benchmarks `[0.05, 0.05]`:

- arithmetic active mean is approximately `+0.05`;
- annualized geometric active premium is `-0.02530492340404014`;
- conventional IR is `-0.04473320734100739`;
- corrected modified IR is `+0.04473320734100739`.

Every language tests this opposing-sign example against the fixed expected
value. The implementation follows the documented active-premium branch.
Documentation no longer claims that simple sign adjustment implements
Israelson's tracking-error exponent modification or removes inverse
tracking-error scaling.

### Argument validation

Every language checks invalid confidence values `-1e20`, `-1`, `0`, `1`, `2`,
NaN, and both infinities, on empty and populated instances. Bias regressions
check zero, negative, NaN, and negative-infinity multipliers.

Previously, non-finite reward-to-drawdown confidence raised conversion errors
in Python/TypeScript, selected a computed tail in Go/Rust, and yielded NaN or
computed tails in Zig. Large finite confidence also overflowed Go's integer
conversion: for returns `[-0.1, 0.2, -0.03]` and confidence `-1e20`, Python
returned `0.3604931266917538`, while Go returned `0.15621368823309331` on the
reviewed platform. All languages now reject these arguments before conversion.

### Rust long-series summation

An unbounded yearly instance fed 10,000 repetitions of returns `0.1` and
`-0.001`, each with benchmark zero, previously produced:

| Measure | Python 3.14 | Previous Rust | Corrected Rust |
|---|---:|---:|---:|
| `rachev_ratio(0.1, 0.1)` | 100.0 | 100.00000000001693 | 100.0 |
| `reward_to_conditional_drawdown(0.95)` | 48.28431257909963 | 48.28431257910049 | 48.28431257909963 |

The new Rust long-series regression checks both results to 13 decimal places.

### R-compatible documentation and fixture naming

The drawdown docstring input `[-0.01, -0.02, 0.01, -0.03, -0.04]` produces
approximately `[-0.029998, -0.069988]`, using
`(prod(1 + r_i * 0.01) - 1) * 100`. Default run order remains chronological;
positive `max_runs` sorts worst-first before truncating.

Down-number includes benchmarks equal to zero, while down-percentage retains
strictly negative benchmarks. Python regressions exercise zero benchmarks and
the corrected continuous-run example.

At the fourth Bacon observation with MAR=0, variability is `8.06`, while
volatility is `sqrt(8.06)`, approximately `2.83901391331568`. The fixture labels
and all five test lookups now agree with these definitions. Every MAR fixture
has exactly 24 observations, eliminating the prior trimming workaround.

## Numerical limit for large annualized outputs

Go/Zig `pow` can differ from Python's libm by several ulps. In direct
comparisons, only `cdar_alpha` and `omega_excess_return` exceeded `5e-14`
absolute error in Go/Zig, with maximum observed relative discrepancy
`1.94e-14`. A daily two-observation window produced CDaR alpha
`9272830523013.78` in Python and `9272830523013.705` in Go/Zig: approximately
`8e-15` relative error despite the visible absolute difference.

These formulas retain their operation order and annualization behavior. The
architecture skill and Go package documentation now qualify precision claims:
ordinary-scale results use absolute decimal-place tolerance; large values use
relative tolerance. Exact agreement for arbitrary magnitudes would require
common transcendental implementations across platforms.

## Architecture and provenance

The dependency graph, reference/port roles, family formulas, reset behavior,
rolling eviction, drawdown episode rebuilding, and language API mappings agree
with the reviewed code. The architecture skill now describes the corrected
contracts and links the R provenance and reference translator.

The R scripts in `performance_analytics_reference_data_generation.md` produced
the PerformanceAnalytics expected values; measures without R data use manual
or formula-based fixtures. `notes/performance/gen.py`, moved from
`/tmp/gen_perf_refdata/gen.py`, translates those Python fixtures. It does not
run R or independently calculate expected values. Its repository root is
resolved from the script path, so it also works from another working directory.

The repository AGENTS.md statement calling autocorrelation penalty a stub was
also corrected: all five implementations calculate the Lo (2002) penalty.
Zig's documented need to reset after an allocation failure during `addReturn`
remains recorded in the skill.

## Validation

Environment: Python 3.14.3, Go 1.27.1, Rust/Cargo 1.98.1, Zig 0.16.0,
Node 24.15.0. Post-fix validation passed:

| Suite | Result |
|---|---:|
| Python performance discovery, including core and resampler tests | 231 tests |
| Go performance + core (uncached) | 239 tests |
| Rust `cargo test --lib performance` | 404 tests; filter also includes roundtrips performance |
| TypeScript performance specs, plus TypeScript build | 237 specs |
| Zig whole repository | 1,922 tests |

A temporary differential harness compared all 144 public Python measures with
their port equivalents: 80 deterministic mixed finite return pairs, including
zeros; configurations `(P, rf, MAR)` of `(1,0,0)`, `(12,0.05,0.03)`,
`(252,0.001,-0.01)`; windows `0,1,2,5,17`; nine intermediate snapshots per
combination; method defaults. This produced 19,440 comparisons per language,
77,760 total, rerun after the fixes. Rust and TypeScript stayed within `5e-14`
absolute error. Go/Zig each had the 39 differences restricted to the two
annualized formulas described above. This is sampled evidence, not proof for
all floating-point inputs or argument combinations.

The fixture translator was verified by regenerating into a temporary tree and
comparing all four generated trees with the checked-in output (`gofmt` for Go).
The skill validator and `git diff --check` passed.

Useful reproduction commands, from the repository root:

```bash
python3 -m unittest discover -s py/performance -t . -p 'test_*.py'
go -C go test -count=1 ./performance/...
cargo test --manifest-path rs/Cargo.toml --lib performance
(cd ts && npm run build && node --loader ./loader.js node_modules/.bin/jasmine --config=jasmine.json 'dist/performance/**/*.spec.js')
(cd zig && zig build test --summary all)
python3 notes/performance/gen.py /tmp/perf-refdata
```
