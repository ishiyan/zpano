# Implementation Plan: Missing Risk Measures from Bacon's Periodic Table

## Overview

This document compares the **Periodic Table of Risk Measures** (Table 5.27, Bacon 2023) against current implementations in:
- **My implementation** (`py/performance/ratios.py`)
- **Reference sources**: Stuart, Ranaroussi (quantstats), Braverock (PerformanceAnalytics R)

The plan is split into **Phase 1** (single return series) and **Phase 2** (multiple series: benchmarks/portfolios).

---

## Phase 1: Single Return Series (No Benchmark Required)

### Already Implemented ✅

| Measure | Location | Notes |
|---------|----------|-------|
| **Sharpe ratio** | `ratios.sharpe_ratio()` | With `ignore_risk_free_rate`, `autocorrelation_penalty` |
| **Sortino ratio** | `ratios.sortino_ratio()` | With `autocorrelation_penalty`, `divide_by_sqrt2` (Schwager) |
| **Omega ratio** | `ratios.omega_ratio()` | |
| **Kappa ratio (orders 1,2,3)** | `ratios.kappa_ratio()`, `kappa3_ratio()` | |
| **Bernardo-Ledoit ratio** | `ratios.bernardo_ledoit_ratio()` | |
| **Upside Potential Ratio** | `ratios.upside_potential_ratio(full=True/False)` | Both methods |
| **Calmar ratio** | `ratios.calmar_ratio()` | |
| **Sterling ratio** | `ratios.sterling_ratio()` | With `annual_excess_rate` |
| **Burke ratio** | `ratios.burke_ratio(modified)` | Standard & modified |
| **Pain index / Pain ratio** | `ratios.pain_index()`, `pain_ratio()` | |
| **Ulcer index / Martin ratio** | `ratios.ulcer_index()`, `martin_ratio()` | |
| **Gain-to-Pain ratio** | `ratios.gain_to_pain_ratio` | Schwager GPR |
| **Risk of Ruin** | `ratios.risk_of_ruin` | |
| **Risk-Return ratio** | `ratios.risk_return_ratio` | Sharpe without RF |
| **Skewness** | `ratios.skew` | scipy.stats.skew |
| **Kurtosis** | `ratios.kurtosis` | scipy.stats.kurtosis (excess) |
| **Cumulative return** | `ratios.cumulative_return` | Geometric |
| **Compound Growth Rate (CAGR)** | `ratios.compound_growth_rate()` | |
| **Drawdowns** | `drawdowns_cumulative`, `drawdowns_peaks`, `drawdowns_continuous` | Multiple methods |
| **Max Drawdown** | `worst_drawdowns_cumulative` | |

---

### Missing from Main Table — Phase 1 Targets

#### **First Moment (Average) — Absolute Column**
| Measure | Status | Source Reference | Priority |
|---------|--------|------------------|----------|
| **MAD Ratio** (Mean Absolute Deviation ratio) | ✅ **DONE** | Braverock: `MeanAbsoluteDeviation` | High |
| **Omega–Sharpe Ratio** | ✅ **DONE** | Braverock: `OmegaSharpeRatio` | Medium |
| **Excess Return** | ⚠️ Partial | `cumulative_return` exists; need annualized excess | Medium |

#### **Second Moment (Variability) — Absolute Column**
| Measure | Status | Source Reference | Priority |
|---------|--------|------------------|----------|
| **Adjusted Sharpe Ratio** (Pezier/White) | ✅ **DONE** | `ratios.adjusted_sharpe_ratio()` | High |
| **Kappa** (generalized downside) | ✅ Done | `kappa_ratio(order)` covers this | — |
| **Sortino–Satchell Ratio** | ✅ **DONE** | `ratios.sortino_satchell_ratio()` | Medium |

#### **Third Moment (Skewness) — Absolute Column**
| Measure | Status | Source Reference | Priority |
|---------|--------|------------------|----------|
| **Skew-adjusted Sharpe Ratio** | ✅ **DONE** | `AdjustedSharpeRatio` covers this | High |
| **SkewnessKurtosisRatio** | ✅ **DONE** | Braverock: `SkewnessKurtosisRatio` | Low |

#### **Fourth Moment (Kurtosis) — Absolute Column**
| Measure | Status | Source Reference | Priority |
|---------|--------|------------------|----------|
| **Adjusted Sharpe Ratio** (kurtosis) | ✅ **DONE** | Same as above — one function covers both | — |

#### **Systematic Risk — Absolute Column**
| Measure | Status | Source Reference | Priority |
|---------|--------|------------------|----------|
| **Beta** | ✅ **DONE (Phase 2)** | `ratios.beta` property | Phase 2 |

#### **Extreme Risk — Absolute Column**
| Measure | Status | Source Reference | Priority |
|---------|--------|------------------|----------|
| **VaR** (Value at Risk) | ✅ **DONE** | `ratios.var(confidence, method)` — historical, gaussian, modified | High |
| **CVaR / Expected Shortfall** | ✅ **DONE** | `ratios.cvar(confidence, method)` — historical, gaussian, modified | High |
| **Reward-to-VaR** | ✅ **DONE** | `ratios.reward_to_var()` | Medium |

#### **Downside (Partial Moments) Column**
| Measure | Status | Source Reference | Priority |
|---------|--------|------------------|----------|
| **Lower Partial Moment (LPM)** | ✅ **DONE** | `ratios.lpm(order, threshold)` | High |
| **Downside Deviation / Semi-Deviation** | ✅ **DONE** | `ratios.downside_deviation()`, `ratios.semi_deviation()` | High |
| **Downside Frequency** | ✅ **DONE** | Braverock: `DownsideFrequency` | Low |
| **Downside Potential** | ✅ **DONE** | Braverock: `DownsidePotential` | Low |
| **Downside Sharpe Ratio** | ✅ **DONE** | `ratios.downside_sharpe_ratio()` | Medium |
| **Sortino Ratio** | ✅ Done | | — |
| **Kappa** | ✅ Done | | — |

#### **Gain–Loss (Partial Moments) Column**
| Measure | Status | Source Reference | Priority |
|---------|--------|------------------|----------|
| **Variability Skewness** | ✅ **DONE** | Braverock: `VolatilitySkewness` | Low |
| **Gain-Loss Ratio** | ✅ **DONE** | `ratios.gain_loss_ratio()` | Medium |
| **Farnelli–Tibiletti Ratio** | ❌ Missing | Braverock: — | Low |

#### **Prospect (Partial Moments) Column**
| Measure | Status | Source Reference | Priority |
|---------|--------|------------------|----------|
| **Omega–Prospect Ratio** | ❌ Missing | Braverock: — | Low |
| **Prospect Ratio** | ✅ **DONE** | Braverock: `ProspectRatio` | Low |
| **Skew-adjusted Prospect Ratio** | ❌ Missing | Braverock: — | Low |
| **New Prospect Ratio** | ❌ Missing | Braverock: — | Low |

#### **Drawdown (Partial Moments) Column**
| Measure | Status | Source Reference | Priority |
|---------|--------|------------------|----------|
| **Sterling Ratio** | ✅ Done | | — |
| **Burke Ratio** | ✅ Done | | — |
| **Calmar Ratio** | ✅ Done | | — |
| **Pain Ratio** | ✅ Done | | — |
| **Ulcer Ratio / Martin Ratio** | ✅ Done | | — |
| **Reward to Conditional Drawdown** | ✅ **DONE** | `ratios.reward_to_conditional_drawdown()` | Medium |

#### **Risk-Adjusted Returns Column**
| Measure | Status | Source Reference | Priority |
|---------|--------|------------------|----------|
| **M² (Modigliani-Modigliani)** | ✅ **DONE** | `ratios.modigliani_modigliani()` | High |
| **Skew-adjusted M²** | ❌ Missing | Braverock: — | Low |
| **Adjusted M²** | ❌ Missing | Braverock: — | Low |
| **Alpha (Jensen's)** | ✅ **DONE (Phase 2)** | `ratios.alpha` property | Phase 2 |

#### **Miscellaneous Measures**
| Measure | Status | Source Reference | Priority |
|---------|--------|------------------|----------|
| **K-Ratio** | ✅ **DONE** | `ratios.k_ratio()` | Medium |
| **Upside Capture / Downside Capture / Capture Ratio** | ✅ **DONE (Phase 2)** | `ratios.upside_capture`, `downside_capture`, `capture_ratio` | Phase 2 |
| **R²** | ✅ **DONE (Phase 2)** | `ratios.r_squared` property | Phase 2 |
| **Tail Ratio** | ✅ **DONE** | `ratios.tail_ratio()` | Medium |
| **Convexity** | ❌ Missing | Braverock: — | Low |
| **Bera–Jarque Statistic** | ❌ Missing | Braverock: — (normality test) | Low |
| **Active Share** | ❌ Missing | Braverock: — | Phase 2 |
| **Omega Excess** | ✅ DONE | Braverock: `OmegaExcessReturn` | Phase 2 |
| **Kelly Ratio / Kelly Criterion** | ✅ **DONE** | `ratios.kelly_ratio()` | Medium |
| **Probabilistic Sharpe Ratio (PSR)** | ✅ **DONE** | `ratios.probabilistic_sharpe_ratio()` | Medium |
| **Information Ratio** | ✅ **DONE (Phase 2)** | `ratios.information_ratio` property | Phase 2 |
| **Treynor Ratio** | ✅ **DONE (Phase 2)** | `ratios.treynor_ratio` property | Phase 2 |
| **Appraisal Ratio** | ✅ **DONE (Phase 2)** | `ratios.appraisal_ratio` property | Phase 2 |
| **Rachev Ratio** | ✅ DONE | Braverock: `RachevRatio` | Low |
| **Timing Ratio** | ✅ DONE | Braverock: `TimingRatio`, `MarketTiming` | Phase 2 |
| **DRatio** | ✅ DONE | Braverock: `DRatio` | Low |
| **Hurst Index** | ❌ Missing | Braverock: `HurstIndex` | Low |
| **Smoothing Index (Getmansky)** | ✅ DONE | Braverock: `SmoothingIndex` | Low |
| **Bernardo-Ledoit Ratio** | ✅ Done | | — |
| **Gain-to-Pain Ratio** | ✅ Done | | — |

---

## Phase 2: Multiple Return Series (Benchmark / Portfolio Required)

### Required Infrastructure ✅ **DONE**

| Component | Status | Notes |
|-----------|--------|-------|
| **Benchmark returns** in `add_return()` | ✅ Already accepted | `return_benchmark` parameter |
| **Multiple series tracking** | ✅ **DONE** | `benchmark_returns` array + cached stats |
| **Rolling regression (OLS)** | ✅ **DONE** | Incremental covariance, correlation, beta, alpha |
| **Covariance/Correlation tracking** | ✅ **DONE** | Incremental in `add_return()` |

### Phase 2 Targets (require benchmark) — **ALL COMPLETED**

| Measure | Category | Status | Source Reference |
|---------|----------|--------|------------------|
| **Beta (CAPM)** | Systematic | ✅ `ratios.beta` | Braverock: `CAPM.beta` |
| **Alpha (Jensen's)** | Risk-adjusted | ✅ `ratios.alpha` | Braverock: `CAPM.alpha` |
| **Information Ratio** | Relative | ✅ `ratios.information_ratio` | Braverock: `InformationRatio` |
| **Treynor Ratio** | Relative | ✅ `ratios.treynor_ratio` | Braverock: `TreynorRatio` |
| **Appraisal Ratio** | Relative | ✅ `ratios.appraisal_ratio` | Braverock: `AppraisalRatio` |
| **Tracking Error** | Relative | ✅ `ratios.tracking_error` | Braverock: `TrackingError` |
| **Active Premium** | Relative | ✅ `ratios.active_premium` | Braverock: `ActivePremium` |
| **Upside/Downside Capture** | Relative | ✅ `ratios.upside_capture`, `downside_capture` | Braverock: `UpDownRatios` |
| **R²** | Relative | ✅ `ratios.r_squared` | Braverock: — (from regression) |
| **Specific Risk / Systematic Risk / Total Risk** | Relative | ⚠️ Partial (specific via appraisal) | Braverock: `SpecificRisk`, etc. |
| **Modigliani / M²** | Risk-adjusted | ✅ `ratios.modigliani_modigliani()` | Braverock: `Modigliani`, `Modigliani` |
| **Omega Excess Return** | Relative | ✅ DONE | Braverock: `OmegaExcessReturn` |
| **Timing Ratio** | Relative | ✅ DONE | Braverock: `TimingRatio`, `MarketTiming` |
| **Fama Beta** | Relative | ❌ Missing | Braverock: `FamaBeta` |
| **CDaR Alpha/Beta** | Drawdown | ❌ Missing | Braverock: `CDaR.alpha`, `CDaR.beta` |
| **Factor Alpha** | Relative | ❌ Missing | Braverock: `SFM.coefficients` |
| **Bias Ratio** | Relative | ❌ Missing | Braverock: — |
| **Active Share** | Relative | ❌ Missing | Braverock: — |

---

## Implementation Status Summary

### Phase 1 — **ALL HIGH/MEDIUM PRIORITY COMPLETED** ✅
- VaR / CVaR (Historical, Gaussian, Cornish-Fisher)
- LPM/HPM generic with caching
- Downside Deviation / Semi-Deviation
- Adjusted Sharpe Ratio (Pezier/White)
- Tail Ratio
- Kelly Criterion (Full/Half)
- Probabilistic Sharpe Ratio (PSR)
- K-Ratio (Kestner)
- Sortino-Satchell Ratio
- Gain-Loss Ratio
- Reward-to-VaR
- Downside Sharpe Ratio
- Reward to Conditional Drawdown
- M² (Modigliani-Modigliani)

### Phase 2 — **ALL HIGH PRIORITY COMPLETED** ✅
- Rolling OLS infrastructure (beta, alpha, R², tracking error)
- CAPM Beta
- Jensen's Alpha
- Information Ratio
- Treynor Ratio
- Tracking Error
- Active Premium
- Upside/Downside Capture
- Capture Ratio
- Appraisal Ratio
- Modigliani/M² (full with internal benchmark)

---
 
## Remaining Low Priority / Nice-to-Have
 
| Measure | Phase | Status | Notes |
|---------|-------|--------|-------|
| Downside Frequency | 1 | ✅ DONE | `ratios.downside_frequency(mar)` |
| Downside Potential | 1 | ✅ DONE | `ratios.downside_potential(mar)` via LPM |
| Volatility Skewness | 1 | ✅ DONE | `ratios.volatility_skewness(mar, stat)` |
| Prospect Ratio | 1 | ✅ DONE | `ratios.prospect_ratio(mar)` |
| Skewness-Kurtosis Ratio | 1 | ✅ DONE | `ratios.skewness_kurtosis_ratio()` |
| MAD Ratio | 1 | ✅ DONE | `ratios.mad_ratio()` |
| Omega-Sharpe Ratio | 1 | ✅ DONE | `ratios.omega_sharpe_ratio()` |
| Farnelli–Tibiletti Ratio | 1 | ✅ DONE | `ratios.farinelli_tibiletti_ratio(u, l, mar)` |
| Bera-Jarque Normality Test | 1 | ✅ DONE | `ratios.bera_jarque_statistic()`, `ratios.is_normal_distribution()` |
| Hurst Index | 1 | ✅ DONE | `ratios.hurst_index()` |
| DRatio | 1 | ✅ DONE | `ratios.d_ratio()` |
| Rachev Ratio | 1 | ✅ DONE | `ratios.rachev_ratio(alpha, beta, rf)` |
| Timing Ratio | 2 | ✅ DONE | `ratios.timing_ratio()` (beta+/beta-) |
| Fama Beta | 2 | ✅ DONE | `ratios.fama_beta()` |
| CDaR Alpha/Beta | 2 | ✅ DONE | `ratios.cdar_alpha()`, `ratios.cdar_beta()` |
| Factor Models (SFM) | 2 | ✅ DONE | `ratios.sfm_coefficients()` |
| Bias Ratio | 1 | ✅ DONE | `ratios.bias_ratio(std_dev_multiple)` |
| Active Share | 2 | DROPPED | Requires portfolio positions |
| Omega Excess Return | 2 | ✅ DONE | `ratios.omega_excess_return(mar)` |
| Smoothing Index (Getmansky) | 1 | ✅ DONE | `ratios.smoothing_index(neg_thetas, ma_order)` |
| Percentile Rank (5 methods) | 1 | ✅ DONE | `ratios.percentile_rank(method)` |
| Modified Information Ratio (Israelson) | 2 | ✅ DONE | `ratios.modified_information_ratio()` |
| Geometric Information Ratio | 2 | ✅ DONE | `ratios.information_ratio_geometric()` |
| Farnelli–Tibiletti Ratio | 1 | ✅ DONE | `ratios.farinelli_tibiletti_ratio(u, l, mar)` |
| Bera-Jarque Normality Test | 1 | ✅ DONE | `ratios.bera_jarque_statistic()`, `ratios.is_normal_distribution()` |

---

## Design Notes

- **Single file**: All measures in `py/performance/ratios.py`
- **Incremental computation**: All metrics computed in `add_return()` O(1) per call
- **Rolling window support**: All metrics work with `rolling_window` and `min_periods`
- **Test coverage**: 84 unit tests passing, conformance to R PerformanceAnalytics
- **Tolerance**: 13 decimal places matching reference implementations
