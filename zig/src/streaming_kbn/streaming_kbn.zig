//! Streaming (one-pass, O(1) per sample) statistics with Klein second-order
//! Kahan-Babuška-Neumaier (KBN) compensated summation.
//!
//! Uses only the Zig standard library to keep the algorithms portable.
//!
//! Inputs are assumed to be finite floats. The package does not validate NaN or
//! infinity, or recover when an intermediate product overflows.
//!
//! - KleinKBNAccumulator: compensated sum; supports set() and revert().
//! - KleinKBNSummator: compensated sum plus sample count and mean.
//! - RawMomentsKleinKBN: mean, variance, skewness and kurtosis from raw power
//!   sums; revert() removes any previously added sample (FIFO rolling windows).
//! - CentralMomentsKleinKBN: mean, variance, skewness and kurtosis from Pébay's
//!   central moment updates; more accurate for data with a large mean and
//!   supports removal of any previously added sample.
//! - LinearRegressionKleinKBN: OLS slope, intercept, correlation and covariance;
//!   revert() removes any previously added sample.

pub const klein_kbn_accumulator = @import("klein_kbn_accumulator");
pub const klein_kbn_summator = @import("klein_kbn_summator");
pub const raw_moments_klein_kbn = @import("raw_moments_klein_kbn");
pub const central_moments_klein_kbn = @import("central_moments_klein_kbn");
pub const linear_regression_klein_kbn = @import("linear_regression_klein_kbn");

pub const KleinKBNAccumulator = klein_kbn_accumulator.KleinKBNAccumulator;
pub const KleinKBNSummator = klein_kbn_summator.KleinKBNSummator;
pub const RawMomentsKleinKBN = raw_moments_klein_kbn.RawMomentsKleinKBN;
pub const CentralMomentsKleinKBN = central_moments_klein_kbn.CentralMomentsKleinKBN;
pub const LinearRegressionKleinKBN = linear_regression_klein_kbn.LinearRegressionKleinKBN;
