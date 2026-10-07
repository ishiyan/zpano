// Package streamingkbn implements streaming (one-pass, O(1) per sample)
// statistics with Klein second-order Kahan-Babuška-Neumaier (KBN)
// compensated summation.
//
// Uses only the Go standard library to keep the algorithms portable.
//
// Inputs are assumed to be finite floats. The package does not validate NaN
// or infinity, or recover when an intermediate product overflows.
//
//   - [KleinKBNAccumulator]: compensated sum; supports Set and Revert.
//   - [KleinKBNSummator]: compensated sum plus sample count and mean.
//   - [RawMomentsKleinKBN]: mean, variance, skewness and kurtosis from raw
//     power sums; Revert removes any previously added sample (FIFO rolling
//     windows).
//   - [CentralMomentsKleinKBN]: mean, variance, skewness and kurtosis from
//     Pébay's central moment updates; more accurate for data with a large
//     mean and supports removal of any previously added sample.
//   - [LinearRegressionKleinKBN]: OLS slope, intercept, correlation and
//     covariance; Revert removes any previously added sample.
package streamingkbn
