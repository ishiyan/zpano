//! Streaming time-series performance and risk measures.
//!
//! [`Measures`] computes ~135 performance and risk measures (moments,
//! VaR/ES, partial moments, Sharpe/Sortino/Omega/Kappa families,
//! drawdowns and CDaR, single-factor model and benchmark-relative
//! measures, capture ratios) one return observation at a time, over an
//! unbounded or rolling window. Return observations share a common,
//! explicitly defined period; no timestamps are used and the
//! annualization convention is the explicit `periods_per_annum`.
//!
//! The building blocks live in [`core`].

pub mod core;
pub mod measures;

pub use measures::*;

#[cfg(test)]
pub(crate) mod reference_data;

#[cfg(test)]
mod measures_tests;
