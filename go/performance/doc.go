// Package performance provides portfolio performance measurement ratios
// for evaluating the risk-adjusted returns of financial strategies.
//
// This package implements various financial ratios commonly used in
// portfolio performance analysis, including Sharpe, Sortino, Omega,
// Kappa, Calmar, Sterling, Burke, Pain, Ulcer, and Martin ratios.
//
// # Key Features
//
//   - Incremental computation via AddReturn for streaming/online analysis
//   - Multiple risk-adjusted return ratios (Sharpe, Sortino, Omega, etc.)
//   - Drawdown analysis (cumulative, peak-based, continuous)
//   - Higher and lower partial moment calculations
//   - Support for multiple periodicities (daily, weekly, monthly, quarterly, annual)
//   - Configurable risk-free rate and target return
//   - Day count convention support via the daycounting package
//
// # Architecture
//
// The central type is [Ratios], which accumulates returns incrementally
// and maintains running statistics for all supported ratios. Each call to
// [Ratios.AddReturn] updates internal state so that any ratio can be
// queried at any point in time.
//
// # Validation
//
// All calculations are validated against:
//
//   - R's PerformanceAnalytics package (Portfolio Bacon dataset)
//   - Python reference implementation (scipy.stats for kurtosis/skewness)
//   - Results must match to 13+ decimal places
//
// # Usage Example
//
//	import (
//	    "time"
//	    "zpano/performance"
//	    "zpano/daycounting/conventions"
//	)
//
//	r := performance.New(
//	    performance.Daily,
//	    0.0,   // annual risk-free rate
//	    0.0,   // annual target return
//	    conventions.RAW,
//	)
//	r.Reset()
//	r.AddReturn(0.003, 0.002, 1.0, date1, date2)
//	sharpe := r.SharpeRatio(false, false)
//
// # Standards Compliance
//
// The package implements methods documented in:
//
//   - Carl R. Bacon, "Practical Portfolio Performance Measurement and Attribution" (2nd & 3rd ed.)
//   - R's PerformanceAnalytics package
//   - ISDA and ISO 20022 day count conventions (via the daycounting package)
package performance
