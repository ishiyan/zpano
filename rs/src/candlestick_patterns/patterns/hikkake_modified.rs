//! Hikkake Modified pattern (4-candle) with stateful confirmation.

use crate::candlestick_patterns::CandlestickPatterns;

/// Hikkake Modified: a four-candle pattern with near criterion.
///
/// Returns:
/// +100.0/-100.0 for detection, +200.0/-200.0 for confirmation, 0.0 otherwise.
pub fn hikkake_modified(cp: &CandlestickPatterns) -> f64 {
    // If pattern was just detected this bar (takes priority over confirmation)
    if cp.count < 4 {
        return 0.0;
    }

    if cp.hikmod_pattern_idx == cp.count && cp.hikmod_pattern_result != 0.0 {
        return cp.hikmod_pattern_result;
    }

    // If just confirmed this bar
    if cp.hikmod_confirmed {
        return cp.hikmod_last_signal;
    }

    0.0
}
