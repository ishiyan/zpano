/// Fuzzy signal functions for technical indicator interpretation.
///
/// Standalone functions that take raw indicator values and return fuzzy
/// membership degrees ∈ [0, 1]. Zero dependency on the indicators module.
pub mod threshold;
pub mod crossover;
pub mod band;
pub mod histogram;
pub mod compose;

pub use threshold::*;
pub use crossover::*;
pub use band::*;
pub use histogram::*;
pub use compose::*;
