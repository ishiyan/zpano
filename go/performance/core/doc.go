// Package core provides the streaming building blocks used by the
// performance measures in package performance.
//
// In the Python reference implementation these are private helpers
// (performance.core); in Go they are exported so that package
// performance can import them.
//
// All accumulators are O(1) per sample unless stated otherwise and use
// Klein second-order Kahan-Babuška-Neumaier compensated summation from
// package streamingkbn. Accumulators exposing Revert are pure: the caller
// owns the rolling window and feeds evicted values to Revert and new
// values to Update.
//
//   - [Capture]: upside/downside capture, number and percentage ratios.
//   - [ContinuousDrawdownRuns]: continuous losing-run drawdowns (Burke).
//   - [CumulativeReturn]: cumulative and geometric mean returns.
//   - [DrawdownEpisodes]: high-water-mark drawdown episodes.
//   - [HighWaterMarkDrawdown]: rolling high-water-mark drawdowns.
//   - [PartialMoments], [RawPartialMoments]: lower/higher partial moments.
//   - [SFMRegression]: single factor model (CAPM) regression.
//   - [WinLoss]: winning/losing return sums, means and counts.
//   - [NormCDF], [NormPDF], [NormPPF]: standard normal distribution.
//   - [Percentile]: NumPy "linear" percentile.
//   - [VarHistorical], [VarGaussian], [VarCornishFisher]: value at risk.
//   - [EsHistorical], [EsGaussian], [EsCornishFisher]: expected shortfall.
//   - [ProbabilisticSharpeRatio]: probabilistic Sharpe ratio.
package core
