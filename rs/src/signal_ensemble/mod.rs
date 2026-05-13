/// Signal ensemble: weighted blending of multiple signal sources.
///
/// Provides adaptive weighted blending of independent signal sources
/// with delayed feedback and online weight learning.
pub mod method;
pub mod error_metric;
pub mod aggregator;

pub use method::*;
pub use error_metric::*;
pub use aggregator::*;
