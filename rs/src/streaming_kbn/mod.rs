/// Streaming O(1) statistical accumulators using Klein second-order
/// Kahan-Babuška-Neumaier (KBN) compensated summation.
pub mod klein_kbn_accumulator;
pub mod raw_moments_klein_kbn;
pub mod central_moments_klein_kbn;
pub mod linear_regression_klein_kbn;

pub use klein_kbn_accumulator::KleinKbnAccumulator;
pub use raw_moments_klein_kbn::RawMomentsKleinKbn;
pub use central_moments_klein_kbn::CentralMomentsKleinKbn;
pub use linear_regression_klein_kbn::LinearRegressionKleinKbn;
