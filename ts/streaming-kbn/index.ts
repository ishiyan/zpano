/**
 * Streaming (one-pass, O(1) per sample) statistics with Klein second-order
 * Kahan-Babuška-Neumaier (KBN) compensated summation.
 *
 * Uses no external dependencies to keep the algorithms portable.
 *
 * Inputs are assumed to be finite numbers. The module does not validate NaN or
 * infinity, or recover when an intermediate product overflows.
 *
 * - KleinKbnAccumulator: compensated sum; supports set() and revert().
 * - KleinKbnSummator: compensated sum plus sample count and mean.
 * - RawMomentsKleinKbn: mean, variance, skewness and kurtosis from raw power
 *   sums; revert() removes any previously added sample (FIFO rolling windows).
 * - CentralMomentsKleinKbn: mean, variance, skewness and kurtosis from Pébay's
 *   central moment updates; more accurate for data with a large mean and
 *   supports removal of any previously added sample.
 * - LinearRegressionKleinKbn: OLS slope, intercept, correlation and covariance;
 *   revert() removes any previously added sample.
 *
 * @module
 */

export { KleinKbnAccumulator } from './klein-kbn-accumulator';
export { KleinKbnSummator } from './klein-kbn-summator';
export { RawMomentsKleinKbn } from './raw-moments-klein-kbn';
export { CentralMomentsKleinKbn } from './central-moments-klein-kbn';
export { LinearRegressionKleinKbn } from './linear-regression-klein-kbn';
