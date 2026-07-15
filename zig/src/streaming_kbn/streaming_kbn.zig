// Barrel module for the streaming_kbn library.
// Re-exports all accumulator and moment types.

pub const klein_kbn_accumulator = @import("klein_kbn_accumulator");
pub const raw_moments_klein_kbn = @import("raw_moments_klein_kbn");
pub const central_moments_klein_kbn = @import("central_moments_klein_kbn");
pub const linear_regression_klein_kbn = @import("linear_regression_klein_kbn");

pub const KleinKBNAccumulator = klein_kbn_accumulator.KleinKBNAccumulator;
pub const RawMomentsKleinKBN = raw_moments_klein_kbn.RawMomentsKleinKBN;
pub const CentralMomentsKleinKBN = central_moments_klein_kbn.CentralMomentsKleinKBN;
pub const LinearRegressionKleinKBN = linear_regression_klein_kbn.LinearRegressionKleinKBN;
