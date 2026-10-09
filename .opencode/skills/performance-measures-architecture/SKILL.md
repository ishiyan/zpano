---
name: performance-measures-architecture
description: Architecture, formulas, and cross-language implementation reference for the zpano streaming performance and risk measures package (Measures class, core helpers, PerformanceAnalytics reference data). Load when adding or changing a performance measure, porting the package across languages, or debugging a mismatch against R PerformanceAnalytics.
---

# Performance Measures Architecture

Streaming (one return observation at a time) calculation of ~135 time-series performance and risk measures for a portfolio and its benchmark: moments, VaR/ES, partial moments, Sharpe/Sortino/Omega/Kappa families, drawdowns and CDaR, single-factor-model (SFM) and benchmark-relative measures, capture ratios, and miscellaneous ratios. Most measures are O(1) reads from KBN-compensated streaming state; a documented minority recompute over the stored window.

**Reference implementation:** Python `py/performance/`. Go, TypeScript, Zig, and Rust are ports of it and must match it to 13+ decimal places. `py/performance_deprecated/` (old `Ratios` + `periodicity`) is legacy and has no ports. `py/performance/resampler/` is experimental, unused by `Measures`, and intentionally **not ported**.

## Module Dependencies

```
streaming_kbn/          KleinKBNAccumulator, KleinKBNSummator, RawMomentsKleinKBN, LinearRegressionKleinKBN
    |
    v
performance/core/       streaming helpers (partial moments, drawdowns, capture, VaR/ES, norm, …)
    |
    v
performance/measures    Measures class (public API)

performance/reference_data/   test-only expected values (PerformanceAnalytics / reference formulas)
```

No dependency on `daycounting`, `entities`, or `indicators`. Measures take no timestamps: periodicity is the explicit `periods_per_annum` argument.

## Package Structure

| Python `py/performance/` | Go `go/performance/` | TypeScript `ts/performance/` | Rust `rs/src/performance/` | Zig `zig/src/performance/` |
|---|---|---|---|---|
| `__init__.py` | `doc.go` | `index.ts` | `mod.rs` | `performance.zig` (barrel, build module `performance`) |
| `measures.py` | `measures.go` + `measures_{distribution,ratios,drawdowns,benchmark,misc}.go` (split by family) | `measures.ts` | `measures.rs` | `measures.zig` |
| `test_measures.py`, `test_review_regressions.py` | `measures_test.go`, `measures_helpers_test.go`, `measures_<family>_test.go`, `measures_review_test.go` | `measures.spec.ts`, `measures-drawdowns-benchmark.spec.ts`, `measures-test-helpers.ts`, `measures-review.spec.ts` | `measures_tests/` (`mod.rs` helpers + `general`, `risk`, `ratios`, `drawdowns`, `benchmark`, `formulas`, `review`) | `measures_test.zig`, `measures_test_2.zig`, `measures_review_test.zig` |
| `core/__init__.py` | `core/doc.go` (package `core`) | `core/index.ts` | `core/mod.rs` | `core/core.zig` |
| `core/drawdown_episodes.py` | `core/drawdownepisodes.go` | `core/drawdown-episodes.ts` | `core/drawdown_episodes.rs` | `core/drawdown_episodes.zig` |
| `core/test_*.py` | `core/*_test.go` (+ `helpers_test.go`) | `core/*.spec.ts` | inline `#[cfg(test)]` (+ `risk_helpers_tests.rs`, `test_support` in `core/mod.rs`) | inline `test` blocks (+ internal `core/fifo_buffer.zig`) |
| `reference_data/*.py` | `referencedata/*.go` (package `referencedata`) | `reference-data/*.ts` | `reference_data/*.rs` (`#[cfg(test)]`) | `reference_data/*.zig` |

Filename rules: Go drops underscores (`high_watermark_drawdown.py` → `highwatermarkdrawdown.go`), TS uses dashes, Rust/Zig keep underscores.

### Build Registration

- **Go:** directory packages `zpano/performance`, `zpano/performance/core`, `zpano/performance/referencedata`.
- **TypeScript:** `ts/tsconfig.json` includes `performance/**/*.ts`.
- **Rust:** `rs/src/lib.rs` has `pub mod performance;`; `mod.rs` declares `core`, `measures`, and `#[cfg(test)] pub(crate) mod reference_data;`.
- **Zig:** `zig/build.zig` registers one library module `performance` and one test module, both rooted at `src/performance/performance.zig`, importing `klein_kbn_accumulator`, `klein_kbn_summator`, `raw_moments_klein_kbn`, `linear_regression_klein_kbn` by module name. Files inside the package use relative `@import("x.zig")`. (`rt_performance` in build.zig is the unrelated roundtrips performance module.)

## Measures: Construction and Streaming

```
Measures(periods_per_annum=252, annual_risk_free_rate=0, annual_target_return=0, rolling_window_size=0)
```

- `periods_per_annum` (P) must be > 0, else error. Constants: `PERIODS_PER_ANNUM_YEAR=1, _QUARTER=4, _MONTH=12, _WEEK=52, _DAY=252, _MINUTE_US_EQUITIES=98280 (390×252), _MINUTE_CRYPTO=525600 (1440×365)`.
- Periodic rates: `rf = (1+rf_a)^(1/P) − 1`, `T = (1+T_a)^(1/P) − 1`; the annual value is used as-is when it is 0 or P == 1. The annual rf is kept for `jensen_alpha` and `m_squared`.
- `rolling_window_size` ≤ 0 (or None) → unbounded window. **Both return series are always stored** (deque / ring buffer), so memory is O(n) in both modes.
- `reset()` clears all state, keeps configuration.

### Internal state

| Field | Type | Tracks |
|---|---|---|
| `_returns`, `_returns_benchmark` | deque(maxlen=w) | window of r and b |
| `_returns_kbn` | RawMomentsKleinKBN(ddof=1, bias=True, fisher=True) | r |
| `_excess_returns_kbn` | RawMoments | e = r − rf |
| `_target_returns_kbn` | RawMoments | r − T |
| `_benchmark_returns_kbn`, `_benchmark_excess_returns_kbn` | RawMoments | b, b − rf |
| `_active_returns_kbn` | RawMoments | r − b |
| `_sfm_regression` | core.SFMRegression | OLS y=r−rf on x=b−rf: all / bull (x>0) / bear (x<0) |
| `_target_partial_moments`, `_benchmark_target_partial_moments` | core.PartialMoments(T) | LPM1–4, HPM1–4 about T (÷n), upper/lower subsets |
| `_raw_partial_moments` | core.RawPartialMoments | **sums** Σmax(r,0), Σmax(−r,0), counts |
| `_win_loss` | core.WinLoss | Σ/count/mean of r≠0, r>0, r<0 |
| `_capture` | core.Capture | up (b>0) / down (b≤0) log-sums, arithmetic sums, number/percentage counts |
| `_cumulative_return`, `_cumulative_excess_return`, `_benchmark_cumulative_return` | core.CumulativeReturn | Σlog1p(·), count |
| `_drawdown_continuous_runs` | core.ContinuousDrawdownRuns | runs of r<0 (Burke) |
| `_drawdown_high_watermark` (+ benchmark) | core.HighWaterMarkDrawdown(w) | per-observation drawdowns, Σdd, Σdd² |
| `_drawdown_episodes` (+ benchmark) | core.DrawdownEpisodes | closed + open episodes, running depth/length/recovery sums |

### `add_return(ret, ret_bench)`

1. `evicted = w > 0 and len(window) == w`. If evicted, pop the oldest (r, b) and `revert` it from every revertible accumulator (moments, partial moments, win/loss, capture, cumulative returns, SFM, continuous runs).
2. `update` every accumulator with the new pair, append to the windows.
3. High-watermark drawdown and drawdown episodes have **no revert**: `HighWaterMarkDrawdown.update` evicts internally and returns `recalculated=True` when it had to recompute the window. If `recalculated or evicted`, episodes are rebuilt with `episodes.recalculate(hwm.drawdowns)` (O(w), keeps indices window-relative); otherwise `episodes.update(hwm.drawdown)` (O(1)).

Inputs are not validated. Python raises from `log1p` if the portfolio return,
benchmark return, or their risk-free-adjusted returns are ≤ −1; the ports use
IEEE NaN/inf instead. Successful numerical parity assumes these log arguments
are in domain. Zig allocation failure can leave `Measures` partially updated;
call `reset()` before reusing it, as documented on `addReturn`.

## Core Helpers (`core/`)

| Helper | Purpose |
|---|---|
| `norm_cdf`, `norm_pdf`, `norm_ppf` | Standard normal; `norm_ppf` = Acklam rational approximation, error for p ≤ 0 or p ≥ 1 (NaN passes and yields NaN) |
| `percentile(window, q)` | NumPy "linear" percentile; error if q ∉ [0,1] or window empty |
| `var_historical/gaussian/cornish_fisher` | VaR from window / moments (Gaussian and CF use ddof=0 σ; CF falls back to Gaussian when S or K is NaN) |
| `es_historical/gaussian/cornish_fisher` | Expected shortfall, same conventions |
| `probabilistic_sharpe_ratio(moments, sr, reference_sr, …)` | Φ((SR−SR*)√(n−1)/√(1 − S·SR + (K−1)SR²/4)) |
| `SFMRegression(risk_free_rate)` | 3 × LinearRegressionKleinKBN: alpha, beta, beta_bull, beta_bear |
| `PartialMoments(threshold)` | LPM/HPM orders 1–4 about threshold, frequencies, subset sums |
| `RawPartialMoments` | sums/counts of positive and negative returns |
| `WinLoss` | non-zero / winning / losing sums, counts, means |
| `Capture` | up/down capture sums and number/percentage counts |
| `CumulativeReturn` | Σlog1p; cumulative, geometric-mean, annualized geometric-mean return; revert on empty is an error |
| `ContinuousDrawdownRuns`, `dd_percent` | runs of consecutive losses in R's percent convention (Burke) |
| `HighWaterMarkDrawdown(window_size)` | rolling cumulative-equity drawdowns `W_t / max(W_start, W_1..W_t) − 1` |
| `DrawdownEpisode`, `DrawdownEpisodes` | episodes (depth, from, trough, to, recovered) and their averages |

## Measures by Family

Notation: r return, e = r − rf, T periodic target, b benchmark, n window size, P periods per annum, `LPMk/HPMk` partial moments about T (÷n), `Gm = expm1(Σlog1p r / n)`, `Gann = expm1(Σlog1p r · P/n)`. **P** = property, **M** = method (Python defaults shown; explicit arguments in Go/Rust/Zig). Unless noted: O(1), not annualized, NaN when undefined.

### Moments and normality (`_returns_kbn`)
`skewness` (g1), `skewness_moment`, `skewness_fisher` (n≥3), `skewness_sample` (n≥3), `kurtosis` (β2−3), `kurtosis_excess`, `kurtosis_moment`, `kurtosis_sample_excess` (n≥4), `kurtosis_sample_corrected` (PA "sample"), `kurtosis_sample` — all from `RawMomentsKleinKBN` (see the streaming-kbn skill). `skewness_kurtosis_ratio = g1/β2`; `jarque_bera_normality_test_statistic = n/6·(g1² + (β2−3)²/4)`; M `is_normal_distribution(confidence=0.95) → bool` = JB ≤ −2·ln(1−c) (validates c ∈ (0,1), rejecting NaN, before returning False if JB is NaN).

### Return and growth (`_cumulative_return`)
`cumulative_geometric_return` (0.0 empty), `geometric_mean_return` (Gm), `compound_annual_growth_rate` (Gann).

### VaR, ES, reward-to-risk (confidence = 0.95)
M `var_historical / var_gaussian / var_cornish_fisher`, `es_historical / es_gaussian / es_cornish_fisher` (raw returns); M `reward_to_var_ratio_*`, `reward_to_es_ratio_*` = mean(e) / VaR|ES(**raw** r); P `mean_absolute_deviation_ratio` (O(n)). Historical variants sort the window (O(n log n)).

### Upside / downside partial moments (about T)
`upside_potential_ratio = HPM1/√LPM2`, `upside_potential_ratio_subset`, `upside_frequency`, `upside_potential` (HPM1), `upside_potential_subset`, `upside_variance` (HPM2), `upside_variance_subset`, `upside_risk` (√HPM2), `upside_risk_subset`, `semi_deviation` (about the mean, ÷n, O(n)), `downside_deviation` (√LPM2), `downside_deviation_subset`, `downside_frequency`, `downside_potential` (LPM1), `variability_skewness = HPM2/LPM2`, `volatility_skewness = √(HPM2/LPM2)`. The `*_subset` variants return 0 (not NaN) when the subset is empty.

### Sharpe family (excess returns)
`sharpe_ratio = mean(e)/σ1(e)`; M `sharpe_ratio_var_{historical,gaussian,cornish_fisher}`, `sharpe_ratio_es_{…}` = mean(e)/VaR|ES(**excess** e); `downside_sharpe_ratio = mean(e)/(√2·semi_deviation)` (±inf if semi-deviation is 0); `adjusted_sharpe_ratio = SR(1 + S·SR/6 − K·SR²/24)`, `adjusted_sharpe_ratio_skew_only`; M `probabilistic_sharpe_ratio[_full|_symmetric|_gaussian](reference_sr=0.0)` (S = g1 or 0, K = 3 or excess+3).

### Sortino, Omega, Kappa and friends
`sortino_ratio = mean(r−T)/√LPM2` (= `sortino_satchell_ratio` = `kappa_2_ratio`), `sortino_ratio_sqrt2`, `omega_ratio = mean(r−T)/LPM1 + 1`, `omega_sharpe_ratio = omega − 1` (= `kappa_1_ratio`), `omega_excess_return` (annualized, O(n)), `kappa_{1..4}_ratio = mean(r−T)/LPMk^(1/k)`, M `prospect_ratio(lambda_loss=2.25)`, `prospect_ratio_performance_analytics`, `bernardo_ledoit_ratio = Σr⁺/Σr⁻`, `d_ratio` (inf when no positive returns, including empty), `gain_loss_ratio`, M `farinelli_tibiletti_ratio(upper_order=2, lower_order=2)` (orders ∈ 1..4, else error), M `rachev_ratio(alpha=0.1, beta=0.1)` (PA non-parametric, O(n log n), error if α/β ∉ (0,1)).

### Win / loss (`_win_loss`, zeros excluded)
`mean_non_zero_return`, `mean_win_return`, `mean_loss_return`, `win_rate`, `loss_rate`.

### Drawdowns
- Lists: `drawdowns_cumulative` = `drawdowns_high_watermark` (copy of the per-observation drawdowns, PA `Drawdowns()`), M `drawdowns_continuous_runs(max_runs=None)` (R percent convention; chronological by default, sorted worst-first and truncated when `max_runs > 0`).
- `min_drawdowns_cumulative` (≤ 0), `worst_drawdowns_cumulative` (|min|) — O(n).
- `calmar_ratio = Gm/|MDD|`, M `sterling_ratio(excess=0.1) = Gm/(|MDD| + excess)` (not annualized, PA scale=1), `burke_ratio = (Gm − rf)/√ΣDD²`, `burke_ratio_modified = burke·√n`, `pain_index = −mean(D)`, `pain_ratio`, `ulcer_index = √mean(D²)`, `martin_ratio`.
- Episodes (0.0 when none): `drawdown_average`, `drawdown_average_length`, `drawdown_average_peak_to_trough`, `drawdown_average_recovery`, `drawdown_deviation = √(Σdepth²/n_obs)`.
- CDaR (confidence = 0.95, error if ∉ (0,1)): M `cdar_average` (continuous path), `cdar_discrete` (episode depths, PA default), `cdar_beta` (benchmark episodes vs portfolio returns), `cdar_alpha` (annualized with P), `reward_to_conditional_drawdown` (Gm / mean magnitude of the worst `max(1, int(n·(1-c)))` path drawdowns).

### SFM / benchmark (x = b − rf, y = r − rf)
`sfm_risk_premium`, `sfm_alpha`, `sfm_beta`, `sfm_beta_bull`, `sfm_beta_bear`, `timing_ratio = β_bull/β_bear`, `sfm_r2`, `jensen_alpha = Gann_p − (β·Gann_b + (1−β)·rf_a)` (annual), `fama_beta = σ0(r)/σ0(b)`, `modigliani` (periodic), `tracking_error = σ1(r−b)·√P`, `active_premium = Gann_p − Gann_b`, `information_ratio`, `information_ratio_modified`, `systematic_risk = |β|·σ1(b−rf)·√P`, `treynor_ratio`, `treynor_ratio_modified`, `specific_risk` (O(n)), `total_risk = √(sys² + spec²)`, `appraisal_ratio`, `jensen_alpha_modified`, `jensen_alpha_alternative`, `m_squared`, `m_squared_excess`, `m_squared_sortino`. Annualized: Jensen family, tracking error, active premium, IR, systematic/specific/total risk, Treynor, M² family.

### Capture (`_capture`)
M `upside_capture_ratio(geometric=True)`, `downside_capture_ratio(geometric=True)`, `overall_capture_ratio(geometric=True)`; `up_number_ratio`, `down_number_ratio` (down bucket b ≤ 0), `up_percentage_ratio`, `down_percentage_ratio` (strict b < 0).

### Miscellaneous
`autocorrelation_penalty` (Lo 2002, lag horizon from P, 1.0 when undefined, O(n·q)), M `tail_ratio(cutoff=0.95)` (not in R; error unless 0.5 < cutoff < 1), `kelly_ratio_full = mean(e)/Var1(e)`, `kelly_ratio` (half Kelly), `hurst_exponent` (single-scale R/S, O(n)), M `bias_ratio(std_dev_multiplier=1.0)` (O(n), error if nonpositive or NaN), `k_ratio` (Kestner, O(n), plain sums), `gain_to_pain_ratio`.

### Behavioral contracts and compatibility

- `gain_to_pain_ratio = Σr / Σmax(-r, 0)`, using compensated sums for both numerator and denominator. It returns NaN without losses and is independent of observation count when a series is repeated.
- `drawdowns_continuous_runs` uses `(prod(1 + r_i * 0.01) - 1) * 100`, following R's Burke percent convention. Without a positive `max_runs`, runs remain chronological; only truncated results are sorted worst-first.
- `down_number_ratio` includes zero benchmark returns (`b ≤ 0`). `down_percentage_ratio` uses strict `b < 0`.
- `information_ratio_modified` returns IR when the annualized geometric `active_premium > 0`, otherwise −IR. It does not branch on arithmetic active mean. This sign adjustment retains inverse tracking-error scaling and is not Israelson's tracking-error exponent modification.
- `is_normal_distribution` and `reward_to_conditional_drawdown` validate `0 < confidence < 1` before data-availability fallbacks, rejecting NaN and infinities even on empty instances. With valid confidence and an undefined JB statistic, normality returns False.
- `bias_ratio` requires a positive multiplier, rejecting NaN before data-availability checks. Positive infinity remains accepted.
- In `reference_data/volatility_skewness.py`, `…_VOLATILITY` contains `sqrt(HPM2/LPM2)` and `…_VARIABILITY` contains `HPM2/LPM2`. Each MAR fixture contains exactly 24 observations; tests use the matching label.
- `omega_excess_return` recomputes benchmark downside deviation over the window.

The review corrections changed gain-to-pain numerical results, modified-IR
sign selection, and invalid-argument behavior in every language. Go
`RewardToConditionalDrawdown` now returns `(float64, error)`; Rust
`reward_to_conditional_drawdown` returns `Result<f64, String>`; Zig's method
adds `error.InvalidArgument` to its error union. Update callers accordingly.
See [the port review and fix record](../../../notes/performance/performance_port_review.md)
for regressions, validation evidence, and numerical limits.

## Cross-Language Conventions

### Values

- NaN exactly where Python returns `math.nan`; 0.0 / 0 where Python returns 0; ±inf where Python returns inf (`d_ratio`, `downside_sharpe_ratio`); `autocorrelation_penalty` 1.0 fallback.
- Lists: Go `[]float64`, TS `number[]`, Rust `Vec<f64>`, Zig allocator-owned `[]f64` (caller frees).
- Formulas keep Python's operation order and constants so results agree to the last few ulps.

### Errors (functions that raise `ValueError` in Python)

| Language | Mechanism |
|---|---|
| Python | `ValueError` |
| Go | `(T, error)` returns; other measures plain `float64` |
| TypeScript | `throw new Error(<same message>)` |
| Rust | `Result<T, String>` |
| Zig | error unions (`!T`), plus `error.OutOfMemory` where allocation happens |

Fallible: constructor, `is_normal_distribution`, `farinelli_tibiletti_ratio`, `rachev_ratio`, `cdar_average/discrete/beta/alpha`, `tail_ratio`, `bias_ratio`, `reward_to_conditional_drawdown`, core `norm_ppf`, `percentile`, `CumulativeReturn.revert`. Measures that call these with always-valid internal arguments are not fallible.

### Naming

| Concept | Python | Go | TypeScript | Rust | Zig |
|---|---|---|---|---|---|
| Constructor | `Measures(252, 0, 0, 0)` | `NewMeasures(ppa, rf, mar, window) (*Measures, error)` | `new Measures(252, 0, 0, 0)` (defaults) | `Measures::new(..) -> Result<Measures, String>` | `Measures.init(allocator, ..) !Measures` + `deinit()` |
| Add | `add_return(r, b)` | `AddReturn(r, b)` | `addReturn(r, b)` | `add_return(r, b)` | `addReturn(r, b) !void` |
| Property | `m.sharpe_ratio` | `m.SharpeRatio()` | `m.sharpeRatio` (getter) | `m.sharpe_ratio()` | `m.sharpeRatio()` |
| Method with default | `m.var_historical()` | `m.VarHistorical(0.95)` | `m.varHistorical()` | `m.var_historical(0.95)` | `m.varHistorical(0.95)` |
| Constant | `PERIODS_PER_ANNUM_DAY` | `PeriodsPerAnnumDay` | `PERIODS_PER_ANNUM_DAY` | `PERIODS_PER_ANNUM_DAY` | `periods_per_annum_day` |

Rolling windows: Go slices (append + reslice), TS arrays (push/shift), Rust `VecDeque<f64>`, Zig `core/fifo_buffer.zig` (contiguous, allocator-owned).

Invalid `confidence` for VaR/ES-based measures (Python raises only indirectly, from `percentile`/`norm_ppf`): Go, Rust and Zig return NaN exactly where Python would raise (historical variants accept c = 0 and c = 1; Gaussian/Cornish-Fisher need the `norm_ppf` argument in (0,1)); TS throws like Python. Zig methods that need scratch memory (historical VaR/ES, Rachev, CDaR, tail ratio, list results) return `!f64` / `![]f64` (`error.OutOfMemory`); returned lists are freed with the `Measures` allocator.

All four ports reproduce Python's compensated built-in `sum()` (Python ≥ 3.12)
for Rachev and reward-to-conditional-drawdown, retaining input order and the
non-finite correction fallback. Do not replace these sums with ordinary
iterator sums: long series can otherwise miss the precision target.

`erf` is a private fdlibm port in TS, Rust and Zig, matching the glibc-based
CPython reference on the tested platform. Transcendental functions are platform
dependent: Go/Zig `pow` may differ from Python's libm by several ulps. Large
annualized outputs need relative tolerance; a blanket guarantee of 13 absolute
decimal places for arbitrarily large outputs is not achievable.

## Tests and Reference Data

- **Dataset:** "portfolio_bacon" from PerformanceAnalytics: Bacon (2008, 2nd ed.) p. 65 portfolio and p. 66 benchmark returns, 24 observations each, defined at the top of `test_measures.py`.
- **Streaming checks:** most measure tests stream the 24 observations through a fresh `Measures` and compare each intermediate value with a 24-element expected list (`run_stream_method` / `run_stream_property`), for several rf/MAR/confidence values and daily/monthly/annual P. Rolling-window tests check that a window of size w equals a fresh instance fed with the last w observations.
- **Expected-value provenance:** the R scripts documented in [performance_analytics_reference_data_generation.md](../../../notes/performance/performance_analytics_reference_data_generation.md) produced the PerformanceAnalytics expected values using the online R interpreter described there. R output was recorded in `py/performance/reference_data/*.py`; measures without R data use manual/formula-based fixtures. The Python-to-port generator is a source translator, not the producer of R results or an independent mathematical oracle.
- **Reference-data translation:** `py/performance/reference_data/*.py` contain literal lists and dicts keyed by float/int/bool/str (or nested float→float dicts). [notes/performance/gen.py](../../../notes/performance/gen.py), formerly `/tmp/gen_perf_refdata/gen.py`, imports them and emits all four port trees. The port files are **generated**, not hand-edited: each data file starts with `Code generated … DO NOT EDIT.`, and floats use Python `repr` so they round-trip exactly. Shapes per language:
  - Go: package-level `var <Module><Name>` holding `[]float64` / `map[float64][]float64` / `map[bool]…` / `map[string]…` / nested maps.
  - TS: `export const NAME: readonly number[] | ReadonlyMap<…>`.
  - Rust: `pub const NAME: &[f64]` / `&[(K, &[f64])]` plus `reference_data::lookup(pairs, key)`.
  - Zig: `pub const name: []const f64` / `[]const Entry(K, V)` plus `reference_data.lookup(K, V, entries, key)`; the module `var` is imported as `@"var"`.

  After changing Python reference fixtures, run `python3 notes/performance/gen.py`
  from the repository root, format the Go output (Rust fixtures currently retain
  the generator's formatting), and rebuild every language.
  An optional output-root argument supports comparison without overwriting the
  checked-in trees: `python3 notes/performance/gen.py /tmp/perf-refdata`.
- **Python-only test mechanics:** the `inspect`-based `public_measures()` becomes an explicit list in the ports. The `PropertyMock` patch of the JB statistic becomes a direct test of the decision rule. `random.Random(seed)` becomes each language's own seeded PRNG with the same distributions; any assertion that depends on Python's exact random stream is adapted to an equivalent invariant.
- **Tolerances:** `assertAlmostEqual(places=N)` maps to Go/Rust/Zig `|a−b| ≤ 0.5·10⁻ᴺ` (or `1e-N`) and TS `toBeCloseTo(x, N)`. Large reference values use a relative tolerance (see `TestSeriesAssertions`).

### Commands

```bash
python3 -m unittest py.performance.test_measures
python3 -m unittest discover -s py/performance/core -t . -p 'test_*.py'
cd go && go test ./performance/...
cd ts && npm test
cd rs && cargo test --lib performance
cd zig && zig build test --summary all
```

## References

- Bacon, C. R. (2008). *Practical Portfolio Performance Measurement and Attribution*, 2nd ed., Wiley.
- Peterson, B. G., Carl, P. et al. *PerformanceAnalytics* R package (portfolio_bacon dataset, reference implementations).
- Lo, A. W. (2002). "The Statistics of Sharpe Ratios". *Financial Analysts Journal*, 58(4).
- Bailey, D. H., López de Prado, M. (2012). "The Sharpe Ratio Efficient Frontier" (probabilistic Sharpe ratio).
- Pezier, J., White, A. (2006). "The Relative Merits of Investable Hedge Fund Indices…" (adjusted Sharpe ratio).
- Chekhlov, A., Uryasev, S., Zabarankin, M. (2005). "Drawdown Measure in Portfolio Optimization" (CDaR).
- Acklam, P. J. "An algorithm for computing the inverse normal cumulative distribution function".
- See also the `streaming-kbn-architecture` skill for the underlying accumulators.
