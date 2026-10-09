// Package performance provides streaming calculation of time-series
// performance and risk measures of a portfolio and its benchmark.
//
// [Measures] consumes one (portfolio, benchmark) return pair at a time via
// [Measures.AddReturn] and exposes about 140 measures: moments and
// normality tests, VaR and expected shortfall, upside/downside partial
// moments, the Sharpe, Sortino, Omega and Kappa families, drawdown
// measures and conditional drawdown at risk (CDaR), single-factor-model
// (SFM) and benchmark-relative measures, capture ratios and
// miscellaneous ratios. Most measures are O(1) reads from Klein
// Kahan-Babuška-Neumaier compensated streaming state (package
// zpano/streamingkbn and zpano/performance/core); a documented minority
// recompute over the stored window.
//
// Returns must be regularly spaced periodic returns expressed as
// decimals. No timestamps are used: the annualization convention is the
// explicit periodsPerAnnum argument of [NewMeasures] (see the
// PeriodsPerAnnum* constants). The annual risk-free rate and the annual
// target return (minimum acceptable return, MAR) are converted to
// periodic rates as (1 + r)^(1/P) - 1.
//
// A positive rolling window size restricts every measure to the most
// recent observations; zero means an unbounded running window. In both
// modes the portfolio and benchmark returns of the window are stored.
//
// Conventions:
//   - Measures return NaN when they are undefined (for example before
//     enough observations have been added), 0 where the reference
//     implementation returns 0, and ±Inf where noted.
//   - Python properties are methods without parameters (SharpeRatio());
//     Python methods with default arguments take explicit parameters, the
//     Python defaults being documented on each method.
//   - Only methods that validate their arguments return an error
//     (IsNormalDistribution, FarinelliTibilettiRatio, RachevRatio,
//     CdarAverage, CdarDiscrete, CdarBeta, CdarAlpha, TailRatio,
//     BiasRatio, RewardToConditionalDrawdown and the constructor). Methods that pass a confidence
//     level to the core helpers (VaR, ES, Sharpe VaR/ES, reward-to-VaR/ES)
//     return NaN for an invalid confidence level instead.
//
// This package is a port of the Python reference implementation
// py/performance/measures.py. Ordinary-scale results match to 13+ decimal
// places; large annualized values require relative tolerance because
// transcendental functions can differ by several ulps across platforms.
package performance
