//! Doji pattern.

use crate::candlestick_patterns::CandlestickPatterns;
use crate::candlestick_patterns::core::{real_body_len};

/// Doji: open quite equal to close.
///
/// Output is positive but this does not mean it is bullish:
/// doji shows uncertainty and is neither bullish nor bearish when
/// considered alone.
///
/// The meaning of "doji" is specified with self._doji_body.
///
/// Returns:
/// Continuous float in [0, 100].  Higher = stronger doji signal.
pub fn doji(cp: &CandlestickPatterns) -> f64 {
    if !cp.enough(1, &[&cp.doji_body]) {
        return 0.0;
    }
    let b = cp.bar(1);
    // Fuzzy: degree to which real_body <= doji_avg.
    let confidence = cp.mu_less(real_body_len(b.o, b.c), &cp.doji_body, 1);
    confidence * 100.0
}
