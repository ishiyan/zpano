pub mod core;
mod patterns;
#[path = "candlestick_patterns.rs"]
mod candlestick_patterns_impl;

pub use core::*;
pub use candlestick_patterns_impl::*;

#[cfg(test)]
#[path = "candlestick_patterns_test.rs"]
mod candlestick_patterns_test;
