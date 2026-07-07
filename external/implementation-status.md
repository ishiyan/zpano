# Implementation Status: Performance Ratios

This table maps Bacon 3rd edition chapters, Braverock R implementations, Python methods in `py/performance/ratios.py`, and test data availability.

## Column Descriptions

- **bacon3**: Link to chapter/section in `external/bacon3/toc.md`
- **braverock**: R source file in `external/braverock/R/`
- **ratios**: Method/function/property in `py/performance/ratios.py`
- **testdata**: Whether test data was generated from Braverock R (found in `external/r-tests/test_results_generation.initial.R`)

## Summary Table

| bacon3 | braverock | ratios | testdata |
|--------|-----------|--------|----------|
| [Adjusted Sharpe ratio](ch5-03.md#g) | AdjustedSharpeRatio.R | adjusted_sharpe_ratio | ✅ (rf 0-0.3 by 0.05) |
| [Sharpe ratio](ch5-03.md#b) | SharpeRatio.R | sharpe_ratio | ✅ (StdDev, VaR, ES, SemiSD; rf 0-0.3 by 0.05) |
| [Roy ratio](ch5-03.md#c) | - | - | - |
| [Risk-free rate](ch5-03.md#d) | - | risk_free_rate (property) | - |
| [Alternative Sharpe ratio](ch5-03.md#e) | - | - | - |
| [Revised Sharpe ratio](ch5-03.md#f) | - | - | - |
| [Skew-adjusted Sharpe ratio](ch5-03.md#h) | - | - | - |
| [Tracking error](ch5-03.md#l) | TrackingError.R | tracking_error | ❌ |
| [Information ratio](ch5-04.md#c) | InformationRatio.R | information_ratio | ✅ |
| [Geometric information ratio](ch5-04.md#d) | - | information_ratio_geometric | - |
| [Modified information ratio](ch5-04.md#f) | - | modified_information_ratio | - |
| [Regression beta](ch5-05.md#d) | CAPM.beta.R | beta | ❌ |
| [Jensen's alpha](ch5-05.md#i) | CAPM.jensenAlpha.R | alpha | ❌ |
| [Bull beta](ch5-05.md#l) | CAPM.beta.R (bull) | - | - |
| [Bear beta](ch5-05.md#m) | CAPM.beta.R (bear) | - | - |
| [Beta timing ratio](ch5-05.md#n) | - | timing_ratio | - |
| [Market timing](ch5-05.md#p) | MarketTiming.R | - | - |
| [Correlation](ch5-05.md#r) | - | - | - |
| [R²](ch5-05.md#s) | - | r_squared | - |
| [Treynor ratio](ch5-05.md#w) | TreynorRatio.R | treynor_ratio | ❌ |
| [Appraisal ratio](ch5-05.md#x) | AppraisalRatio.R | appraisal_ratio | ❌ |
| [Fama decomposition](ch5-06.md#b) | FamaBeta.R | fama_beta | ❌ |
| [Selectivity](ch5-06.md#c) | Selectivity.R | - | - |
| [Net selectivity](ch5-06.md#e) | NetSelectivity.R | - | - |
| [Fama-French three-factor](ch5-06.md#f) | SFM.coefficients.R | sfm_coefficients | ❌ |
| [Three-factor alpha](ch5-06.md#g) | SFM.coefficients.R | sfm_coefficients | ❌ |
| [Carhart four-factor](ch5-06.md#h) | SFM.coefficients.R | sfm_coefficients | ❌ |
| [Four-factor alpha](ch5-06.md#i) | SFM.coefficients.R | sfm_coefficients | ❌ |
| [Multi-factor models](ch5-06.md#j) | SFM.coefficients.R | sfm_coefficients | ❌ |
| [Average drawdown](ch5-07.md#b) | - | - | - |
| [Maximum drawdown](ch5-07.md#c) | maxDrawdown.R / Drawdowns.R | worst_drawdowns_cumulative | ✅ |
| [Largest individual drawdown](ch5-07.md#d) | - | - | - |
| [Recovery time](ch5-07.md#e) | - | - | - |
| [Drawdown deviation](ch5-07.md#f) | DownsideDeviation.R | downside_deviation | ❌ |
| [Ulcer index](ch5-07.md#g) | UlcerIndex.R | ulcer_index | ✅ |
| [Pain index](ch5-07.md#h) | PainIndex.R | pain_index | ✅ |
| [Calmar ratio](ch5-07.md#i) | CalmarRatio.R | calmar_ratio | ✅ |
| [MAR ratio](ch5-07.md#j) | - | - | - |
| [Sterling ratio](ch5-07.md#k) | CalmarRatio.R (Sterling) | sterling_ratio | ✅ (excess 0-0.1 by 0.02) |
| [Sterling-Calmar ratio](ch5-07.md#l) | - | - | - |
| [Burke ratio](ch5-07.md#m) | BurkeRatio.R | burke_ratio | ✅ (modified True/False; rf 0-0.1 by 0.02) |
| [Modified Burke ratio](ch5-07.md#n) | BurkeRatio.R | burke_ratio(modified=True) | ✅ |
| [Martin ratio](ch5-07.md#o) | MartinRatio.R | martin_ratio | ✅ (rf 0-0.1 by 0.02) |
| [Pain ratio](ch5-07.md#p) | PainRatio.R | pain_ratio | ✅ (rf 0-0.1 by 0.02) |
| [Downside risk](ch5-08.md#b) | DownsideDeviation.R | downside_deviation | ❌ |
| [Downside potential](ch5-08.md#c) | DownsideDeviation.R | downside_potential | ❌ |
| [Pure downside risk](ch5-08.md#d) | DownsideDeviation.R | - | - |
| [Half variance](ch5-08.md#e) | SemiDeviation.R | semi_deviation | ❌ |
| [Upside risk](ch5-08.md#f) | UpsideRisk.R | - | - |
| [Mean absolute moment](ch5-08.md#g) | - | - | - |
| [Omega ratio](ch5-08.md#h) | Omega.R | omega_ratio | ✅ (L 0-0.1 by 0.02) |
| [Bernardo and Ledoit ratio](ch5-08.md#i) | BernadoLedoitratio.R | bernardo_ledoit_ratio | ✅ |
| [Omega–Sharpe ratio](ch5-08.md#j) | OmegaSharpeRatio.R | omega_sharpe_ratio | - |
| [Sortino ratio](ch5-08.md#k) | SortinoRatio.R | sortino_ratio | ✅ (MAR 0-0.3 by 0.05) |
| [Reward to half-variance](ch5-08.md#l) | - | - | - |
| [Downside-risk Sharpe ratio](ch5-08.md#m) | DownsideSharpeRatio.R | downside_sharpe_ratio | - (commented in R tests) |
| [Sortino–Satchell ratio](ch5-08.md#n) | - | sortino_satchell_ratio | - |
| [Kappa ratio](ch5-08.md#n) | Kappa.R | kappa_ratio | ✅ (l=1,2,3,4; MAR 0-0.3 by 0.05) |
| [Upside potential ratio](ch5-08.md#o) | UpsidePotentialRatio.R | upside_potential_ratio | ✅ (full/subset; MAR 0-0.1 by 0.02) |
| [Volatility skewness](ch5-08.md#p) | VolatilitySkewness.R | volatility_skewness | ❌ |
| [Variability skewness](ch5-08.md#q) | VolatilitySkewness.R | volatility_skewness(stat="variability") | ❌ |
| [Farinelli–Tibiletti ratio](ch5-08.md#s) | - | farinelli_tibiletti_ratio | - |
| [Prospect ratio](ch5-08.md#t) | ProspectRatio.R | prospect_ratio | - |
| [LPM](ch5-08.md) | lpm.R | lpm, hpm | - |
| [Downside frequency](ch5-08.md) | DownsideFrequency.R | downside_frequency | ❌ |
| [Downside potential](ch5-08.md#c) | DownsideDeviation.R | downside_potential | ❌ |
| [Duration](ch5-09.md#e) | - | - | - |
| [Macaulay duration](ch5-09.md#f) | - | - | - |
| [Macaulay–Weil duration](ch5-09.md#g) | - | - | - |
| [Modified duration](ch5-09.md#h) | - | - | - |
| [Effective duration](ch5-09.md#k) | - | - | - |
| [Convexity](ch5-09.md#n) | - | - | - |
| [Modified convexity](ch5-09.md#o) | - | - | - |
| [Effective convexity](ch5-09.md#p) | - | - | - |
| [Duration beta](ch5-09.md#u) | - | - | - |
| [Reward to duration](ch5-09.md#v) | - | - | - |
| [Hurst index](ch5-10.md#b) | HurstIndex.R | hurst_index | ❌ |
| [Bias ratio](ch5-10.md#c) | - | bias_ratio | - |
| [Active share](ch5-10.md#d) | - | - | DROPPED |
| [Value at risk](ch5-10.md#e) | VaR.R | var (historical, gaussian, modified) | ❌ |
| [CVaR / Expected Shortfall](ch5-10.md#e) | ES.R | cvar (historical, gaussian, modified) | ❌ |
| [M²](ch5-11.md#b) | Modigliani.R / MSquared.R | modigliani_modigliani | ❌ |
| [M² excess return](ch5-11.md#c) | MSquaredExcess.R | - | - |
| [Differential return](ch5-11.md#d) | - | - | - |
| [Adjusted M²](ch5-11.md#e) | - | - | - |
| [Skew-adjusted M²](ch5-11.md#f) | - | - | - |
| [Adjusted Sharpe ratio](ch5-03.md#g) | AdjustedSharpeRatio.R | adjusted_sharpe_ratio | ✅ |
| [Skew-adjusted Sharpe](ch5-03.md#h) | - | adjusted_sharpe_ratio | - |
| [Adjusted Sharpe ratio](ch5-03.md#g) | AdjustedSharpeRatio.R | adjusted_sharpe_ratio | ✅ |
| [Omega–Sharpe ratio](ch5-08.md#j) | OmegaSharpeRatio.R | omega_sharpe_ratio | ❌ |
| [Reward to VaR](ch5-10.md#e) | - | reward_to_var | - |
| [Reward to ES](ch5-10.md#e) | - | - | - |
| [Conditional Sharpe ratio](ch5-13.md) | - | - | - |
| [Rachev ratio](ch5-13.md) | RachevRatio.R | rachev_ratio | ❌ |
| [Generalised Rachev ratio](ch5-13.md) | RachevRatio.R | - | - |
| [Generalised Z ratio](ch5-13.md) | - | - | - |
| [Reward to conditional drawdown](ch5-13.md) | CDAR.alpha.R | cdar_alpha | ❌ |
| [Reward to conditional drawdown](ch5-13.md) | CDAR.beta.R | cdar_beta | ❌ |
| [M² for VaR](ch5-13.md) | - | - | - |
| [Alpha](ch5-13.md) | CAPM.alpha.R | alpha | - |
| [K ratio](ch5-13.md) | - | k_ratio | - |
| [Upside capture](ch5-13.md) | UpDownRatios.R | upside_capture | ❌ |
| [Downside capture](ch5-13.md) | UpDownRatios.R | downside_capture | ❌ |
| [Capture ratio](ch5-13.md) | UpDownRatios.R | capture_ratio | ❌ |
| [R²](ch5-13.md) | - | r_squared | - |
| [Bias ratio](ch5-13.md) | - | bias_ratio | - |
| [Active share](ch5-13.md) | - | - | DROPPED |
| [Omega excess](ch5-13.md) | OmegaExcessReturn.R | omega_excess_return | ❌ |
| [Percentile rank](ch4-05.md#b) | - | percentile_rank | - |
| [Modified information ratio](ch5-04.md#f) | - | modified_information_ratio | - |
| [Geometric information ratio](ch5-04.md#d) | - | information_ratio_geometric | - |
| [Kappa](ch5-08.md#n) | Kappa.R | kappa_ratio | ✅ |
| [Farnelli–Tibiletti ratio](ch5-08.md#s) | - | farinelli_tibiletti_ratio | ❌ |
| [Prospect ratio](ch5-08.md#t) | ProspectRatio.R | prospect_ratio | ❌ |
| [Skewness](ch5-02.md#p) | skewness.R | skew | ❌ |
| [Sample skewness](ch5-02.md#q) | - | - | - |
| [Kurtosis](ch5-02.md#r) | kurtosis.R | kurtosis | ✅ (method=excess) |
| [Excess kurtosis](ch5-02.md#s) | kurtosis.R | kurtosis | ✅ |
| [Sample kurtosis](ch5-02.md#t) | - | - | - |
| [Bera-Jarque statistic](ch5-02.md#u) | SkewnessKurtosisRatio.R | bera_jarque_statistic | ❌ |
| [Covariance](ch5-02.md#x) | - | - | - |
| [Sample covariance](ch5-02.md#y) | - | - | - |
| [Correlation](ch5-02.md#z) | - | - | - |
| [Sample correlation](ch5-02.md#0) | - | - | - |
| [Mean absolute deviation](ch5-02.md#c) | MeanAbsoluteDeviation.R | mad_ratio | ❌ |
| [Variance](ch5-02.md#d) | StdDev.R | - | - |
| [Sample variance](ch5-02.md#h) | StdDev.R | - | - |
| [Standard deviation](ch5-02.md#i) | StdDev.R / StdDev.annualized.R | - | - |
| [Annualised risk](ch5-02.md#j) | StdDev.annualized.R | - | - |
| [Beta](ch5-05.md#h) | CAPM.beta.R | beta | ❌ |
| [Jensen's alpha](ch5-05.md#i) | CAPM.jensenAlpha.R | alpha | ❌ |
| [Bull beta](ch5-05.md#l) | CAPM.beta.R | - | - |
| [Bear beta](ch5-05.md#m) | CAPM.beta.R | - | - |
| [Timing ratio](ch5-05.md#n) | - | timing_ratio | - |
| [Market timing](ch5-05.md#p) | MarketTiming.R | - | - |
| [Systematic risk](ch5-05.md#q) | SystematicRisk.R | - | - |
| [Specific risk](ch5-05.md#u) | SpecificRisk.R | - | - |
| [Total risk](ch5-05.md#q) | TotalRisk.R | - | - |
| [Treynor ratio](ch5-05.md#w) | TreynorRatio.R | treynor_ratio | ❌ |
| [Appraisal ratio](ch5-05.md#x) | AppraisalRatio.R | appraisal_ratio | ❌ |
| [Net selectivity](ch5-06.md#e) | NetSelectivity.R | - | - |
| [Fama beta](ch5-06.md#b) | FamaBeta.R | fama_beta | ❌ |
| [Kelly ratio](ch5-13.md) | KellyRatio.R | kelly_ratio | ❌ |
| [Probabilistic Sharpe ratio](ch5-13.md) | ProbSharpeRatio.R | probabilistic_sharpe_ratio | ❌ |
| [K ratio](ch5-13.md) | - | k_ratio | - |
| [Downside frequency](ch5-08.md) | DownsideFrequency.R | downside_frequency | ❌ |
| [Downside Sharpe ratio](ch5-08.md#m) | DownsideSharpeRatio.R | downside_sharpe_ratio | - (commented in R tests) |
| [Reward to duration](ch5-09.md#v) | - | - | - |
| [Reward to conditional drawdown](ch5-13.md) | CDAR.alpha.R | cdar_alpha | ❌ |
| [Reward to conditional drawdown](ch5-13.md) | CDAR.beta.R | cdar_beta | ❌ |
| [Reward to VaR](ch5-13.md) | - | reward_to_var | - |
| [R²](ch5-13.md) | - | r_squared | - |
| [Tail ratio](ch5-13.md) | - | tail_ratio | - |
| [Convexity](ch5-13.md) | - | - | - |
| [Bera–Jarque statistic](ch5-13.md) | SkewnessKurtosisRatio.R | bera_jarque_statistic | ❌ |
| [Active share](ch5-13.md) | - | - | DROPPED |
| [Omega excess](ch5-13.md) | OmegaExcessReturn.R | omega_excess_return | ❌ |
| [Factor alpha](ch5-13.md) | SFM.coefficients.R | sfm_coefficients | ❌ |
| [Mean](ch5-02.md#b) | mean.utils.R | - | - |
| [Cumulative return](ch5-02.md) | Return.cumulative.R | cumulative_return | ✅ (geometric=True/False) |
| [Compound growth rate](ch5-02.md) | - | compound_growth_rate | - |
| [Drawdowns](ch5-07.md) | Drawdowns.R | drawdowns_cumulative | ✅ |
| [Drawdown peaks](ch5-07.md) | DrawdownPeak.R | drawdowns_peaks | ❌ |
| [Continuous drawdowns](ch5-07.md) | - | drawdowns_continuous | - |
| [LPM](ch5-08.md) | lpm.R | lpm | ❌ |
| [LPM](ch5-08.md) | lpm.R | hpm | ❌ |
| [Downside deviation](ch5-08.md#b) | DownsideDeviation.R | downside_deviation | ❌ |
| [Semi-deviation](ch5-08.md#d) | SemiDeviation.R | semi_deviation | ❌ |
| [Upside risk](ch5-08.md#f) | UpsideRisk.R | - | - |
| [Upside potential](ch5-08.md#f) | UpsidePotentialRatio.R | - | - |
| [Downside potential](ch5-08.md#c) | DownsideDeviation.R | downside_potential | ❌ |
| [Sortino ratio](ch5-08.md#k) | SortinoRatio.R | sortino_ratio | ✅ |
| [Upside potential ratio](ch5-08.md#o) | UpsidePotentialRatio.R | upside_potential_ratio | ✅ |
| [Omega ratio](ch5-08.md#h) | Omega.R | omega_ratio | ✅ |
| [Bernardo and Ledoit ratio](ch5-08.md#i) | BernadoLedoitratio.R | bernardo_ledoit_ratio | ✅ |
| [Omega–Sharpe ratio](ch5-08.md#j) | OmegaSharpeRatio.R | omega_sharpe_ratio | ❌ |
| [Gain–loss ratio](ch5-08.md#i) | - | gain_loss_ratio | ❌ |
| [Reward to half-variance](ch5-08.md#l) | - | - | - |
| [Reward to VaR](ch5-13.md) | - | reward_to_var | - |
| [Reward to conditional drawdown](ch5-13.md) | CDAR.alpha.R | cdar_alpha | ❌ |
| [Reward to conditional drawdown](ch5-13.md) | CDAR.beta.R | cdar_beta | ❌ |
| [Gain-to-Pain ratio](ch5-08.md) | - | gain_to_pain_ratio | - |
| [Risk of Ruin](ch5-13.md) | - | risk_of_ruin | - |
| [Risk-Return ratio](ch5-13.md) | - | risk_return_ratio | - |
| [MAD ratio](ch5-13.md) | MeanAbsoluteDeviation.R | mad_ratio | ❌ |
| [Omega–Sharpe ratio](ch5-13.md) | OmegaSharpeRatio.R | omega_sharpe_ratio | ❌ |
| [Variability skewness](ch5-13.md) | VolatilitySkewness.R | volatility_skewness | ❌ |
| [Gain–loss skewness](ch5-13.md) | - | - | - |
| [Skew-adjusted prospect ratio](ch5-13.md) | - | - | - |
| [New prospect ratio](ch5-13.md) | - | - | - |
| [SkewnessKurtosisRatio](ch5-13.md) | SkewnessKurtosisRatio.R | skewness_kurtosis_ratio | ❌ |
| [K ratio](ch5-13.md) | - | k_ratio | - |
| [Upside capture](ch5-13.md) | UpDownRatios.R | upside_capture | ❌ |
| [Downside capture](ch5-13.md) | UpDownRatios.R | downside_capture | ❌ |
| [Capture ratio](ch5-13.md) | UpDownRatios.R | capture_ratio | ❌ |
| [Bias ratio](ch5-13.md) | - | bias_ratio | - |
| [Risk efficiency ratio](ch5-13.md) | - | - | - |
| [Tail ratio](ch5-13.md) | - | tail_ratio | - |
| [Convexity](ch5-13.md) | - | - | - |
| [Bera–Jarque statistic](ch5-13.md) | SkewnessKurtosisRatio.R | bera_jarque_statistic | ❌ |
| [Factor alpha](ch5-13.md) | SFM.coefficients.R | sfm_coefficients | ❌ |
| [M² for VaR](ch5-13.md) | - | - | - |
| [Omega–prospect ratio](ch5-13.md) | - | - | - |
| [Prospect ratio](ch5-13.md) | ProspectRatio.R | prospect_ratio | ❌ |
| [Omega–Sharpe ratio](ch5-13.md) | OmegaSharpeRatio.R | omega_sharpe_ratio | ❌ |
| [Sortino ratio](ch5-13.md) | SortinoRatio.R | sortino_ratio | ✅ |
| [Kappa](ch5-13.md) | Kappa.R | kappa_ratio | ✅ |
| [Sortino–Satchell ratio](ch5-13.md) | - | sortino_satchell_ratio | - |
| [Farnelli–Tibiletti ratio](ch5-13.md) | - | farinelli_tibiletti_ratio | - |
| [Generalised Rachev ratio](ch5-13.md) | RachevRatio.R | - | - |
| [Generalised Z ratio](ch5-13.md) | - | - | - |
| [Ulcer ratio](ch5-13.md) | UlcerIndex.R | ulcer_index | ✅ |
| [Pain ratio](ch5-13.md) | PainRatio.R | pain_ratio | ✅ |
| [Martin ratio](ch5-13.md) | MartinRatio.R | martin_ratio | ✅ |
| [Calmar ratio](ch5-13.md) | CalmarRatio.R | calmar_ratio | ✅ |
| [Sterling ratio](ch5-13.md) | CalmarRatio.R | sterling_ratio | ✅ |
| [Burke ratio](ch5-13.md) | BurkeRatio.R | burke_ratio | ✅ |
| [Reward to conditional drawdown](ch5-13.md) | CDAR.alpha.R | cdar_alpha | ❌ |
| [Reward to conditional drawdown](ch5-13.md) | CDAR.beta.R | cdar_beta | ❌ |
| [M² for VaR](ch5-13.md) | - | - | - |
| [Alpha](ch5-13.md) | CAPM.alpha.R | alpha | - |
| [Skew-adjusted M²](ch5-13.md) | - | - | - |
| [Adjusted M²](ch5-13.md) | - | - | - |
| ch5-08.md | DRatio.R | d_ratio | - |

## Test Data Generation Notes

The file `external/r-tests/test_results_generation.initial.R` generates test data for:

1. **AdjustedSharpeRatio** - rf 0 to 0.3 by 0.05
2. **SharpeRatio** - StdDev, VaR, ES, SemiSD; rf 0-0.3 by 0.05
3. **DownsideSharpeRatio** - Commented out ("could not find function")
4. **BernardoLedoitRatio** - No parameters
5. **BurkeRatio** - modified True/False; rf 0-0.1 by 0.02; also rf 0-0.3 by 0.05 with yearly dates
6. **Return.excess** - rf 0-0.3 by 0.05
7. **Return.cumulative** - geometric True/False
8. **SortinoRatio** - MAR 0-0.3 by 0.05
9. **Omega** - L 0-0.1 by 0.02; rf=0; method="simple", output="point"
10. **Kappa** - l=1,2,3,4; MAR 0-0.3 by 0.05
11. **InformationRatio** - Portfolio vs Benchmark (both directions)
12. **kurtosis** - methods: excess, moment, fisher, sample, sample_excess
13. **UpsidePotentialRatio** - method full/subset; MAR 0-0.1 by 0.02
14. **Drawdowns/maxDrawdown** - geometric=True
15. **CalmarRatio** - scale=1
16. **SterlingRatio** - excess 0-0.1 by 0.02; scale=1
17. **BurkeRatio** (with yearly dates) - modified True/False; rf 0-0.1 by 0.02
18. **PainRatio** - rf 0-0.1 by 0.02
19. **MartinRatio** - rf 0-0.1 by 0.02
20. **PainIndex** - No parameters
21. **UlcerIndex** - No parameters

## Legend

- ✅ = Test data exists in `external/r-tests/test_results_generation.initial.R`
- ❌ = No test data generated
- DROPPED = Feature not implemented (requires portfolio positions)
- - = Not applicable / Not implemented / No reference

## Proposed Test Parameter Values for Missing Test Data (❌)

### TrackingError (TrackingError.R → tracking_error)
**Parameters:** `scale` (auto-detected from data periodicity: monthly=12)
**Data frequency expected:** Monthly (auto-detected by `periodicity()`)

### CAPM.beta (CAPM.beta.R → beta)
**Parameters:** `Rf` (risk-free rate, monthly), `method` ("LS" or "Robust"), `family` ("mopt", "bisquare", "opt")
**Data frequency expected:** Monthly (Rf should be monthly rate)

### CAPM.jensenAlpha (CAPM.jensenAlpha.R → alpha)
**Parameters:** `Rf` (risk-free rate, monthly), `method` ("LS" or "Rob"), `family` ("mopt"), `series` (FALSE for single value)
**Data frequency expected:** Monthly (Rf should be monthly rate)

### TreynorRatio (TreynorRatio.R → treynor_ratio)
**Parameters:** `Rf` (risk-free rate, monthly), `scale` (auto-detected: monthly=12), `modified` (FALSE/TRUE)
**Data frequency expected:** Monthly (auto-detected by `periodicity()`)

### AppraisalRatio (AppraisalRatio.R → appraisal_ratio)
**Parameters:** `Rf` (risk-free rate, monthly), `method` ("appraisal", "modified", "alternative")
**Data frequency expected:** Monthly (Rf should be monthly rate)

### FamaBeta (FamaBeta.R → fama_beta)
**Parameters:** None additional (just Ra, Rb)
**Data frequency expected:** Monthly (uses `Frequency()` internally)

### DownsideDeviation (DownsideDeviation.R → downside_deviation)
**Parameters:** `MAR` (minimum acceptable return, monthly), `method` ("full" or "subset")
**Data frequency expected:** Monthly (MAR should be monthly rate)

### DownsidePotential (DownsideDeviation.R → downside_potential)
**Parameters:** `MAR` (minimum acceptable return, monthly), `method` ("full")
**Data frequency expected:** Monthly (MAR should be monthly rate)

### SemiDeviation (SemiDeviation.R → semi_deviation)
**Parameters:** None additional (uses MAR = mean(R))
**Data frequency expected:** Monthly

### VolatilitySkewness (VolatilitySkewness.R → volatility_skewness / variability)
**Parameters:** `MAR` (minimum acceptable return, monthly), `stat` ("volatility" or "variability")
**Data frequency expected:** Monthly (MAR should be monthly rate)

### DownsideFrequency (DownsideFrequency.R → downside_frequency)
**Parameters:** `MAR` (minimum acceptable return, monthly)
**Data frequency expected:** Monthly (MAR should be monthly rate)

### HurstIndex (HurstIndex.R → hurst_index)
**Parameters:** None additional
**Data frequency expected:** Monthly

### VaR (VaR.R → var)
**Parameters:** `p` (confidence, default 0.95), `method` ("modified", "gaussian", "historical"), `clean` ("none", "boudt", "geltner", "locScaleRob"), `portfolio_method` ("single")
**Data frequency expected:** Monthly (no auto-annualization for single method)

### ES (ES.R → cvar)
**Parameters:** `p` (confidence, default 0.95), `method` ("modified", "gaussian", "historical"), `clean` ("none", "boudt", "geltner", "locScaleRob"), `portfolio_method` ("single"), `invert` (TRUE), `operational` (TRUE)
**Data frequency expected:** Monthly

### Modigliani (Modigliani.R / MSquared.R → modigliani_modigliani)
**Parameters:** `Rf` (risk-free rate, monthly)
**Data frequency expected:** Monthly (Rf should be monthly rate)

### OmegaSharpeRatio (OmegaSharpeRatio.R → omega_sharpe_ratio)
**Parameters:** `MAR` (minimum acceptable return, monthly)
**Data frequency expected:** Monthly (MAR should be monthly rate)

### RachevRatio (RachevRatio.R → rachev_ratio)
**Parameters:** `alpha` (lower tail prob, default 0.1), `beta` (upper tail prob, default 0.1), `rf` (risk-free rate, monthly)
**Data frequency expected:** Monthly (rf should be monthly rate)

### CDAR.alpha (CDAR.alpha.R → cdar_alpha)
**Parameters:** `p` (confidence, default 0.95), `weights` (NULL), `geometric` (TRUE), `type` (NULL, "average", "max")
**Data frequency expected:** Monthly (annualizes with ^12 internally)

### CDAR.beta (CDAR.beta.R → cdar_beta)
**Parameters:** `p` (confidence, default 0.95), `weights` (NULL), `geometric` (TRUE), `type` (NULL, "average", "max")
**Data frequency expected:** Monthly (annualizes with ^12 internally)

### UpDownRatios (UpDownRatios.R → upside_capture, downside_capture, capture_ratio)
**Parameters:** `method` ("Capture", "Number", "Percent"), `side` ("Up", "Down"), `geometric` (TRUE)
**Data frequency expected:** Monthly

### OmegaExcessReturn (OmegaExcessReturn.R → omega_excess_return)
**Parameters:** `MAR` (minimum acceptable return, monthly)
**Data frequency expected:** Monthly (MAR should be monthly rate)

### skewness (skewness.R → skew)
**Parameters:** `method` ("moment", "fisher", "sample")
**Data frequency expected:** Monthly

### SkewnessKurtosisRatio (SkewnessKurtosisRatio.R → skewness_kurtosis_ratio, bera_jarque_statistic)
**Parameters:** None for skewness_kurtosis_ratio (uses moment method); `confidence` (0.95 or 0.99) for is_normal_distribution
**Data frequency expected:** Monthly

### MeanAbsoluteDeviation (MeanAbsoluteDeviation.R → mad_ratio)
**Parameters:** None additional
**Data frequency expected:** Monthly

### KellyRatio (KellyRatio.R → kelly_ratio)
**Parameters:** `Rf` (risk-free rate, monthly), `method` ("half" or "full")
**Data frequency expected:** Monthly (Rf should be monthly rate)

### ProbSharpeRatio (ProbSharpeRatio.R → probabilistic_sharpe_ratio)
**Parameters:** `Rf` (risk-free rate, monthly), `refSR` (reference Sharpe), `p` (confidence, 0.95), `ignore_skewness` (FALSE), `ignore_kurtosis` (TRUE)
**Data frequency expected:** Monthly (Rf should be monthly rate)

### ProspectRatio (ProspectRatio.R → prospect_ratio)
**Parameters:** `MAR` (minimum acceptable return, monthly)
**Data frequency expected:** Monthly (MAR should be monthly rate)

### DRatio (DRatio.R → d_ratio)
**Parameters:** None additional
**Data frequency expected:** Monthly

### LPM/HPM (lpm.R → lpm, hpm)
**Parameters:** `n` (moment order, default 2), `threshold` (default 0), `about_mean` (FALSE)
**Data frequency expected:** Monthly (threshold should be monthly rate)
