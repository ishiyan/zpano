//! Tristar pattern (3-candle reversal).

use crate::candlestick_patterns::CandlestickPatterns;
use crate::candlestick_patterns::core::{is_real_body_gap_down, is_real_body_gap_up, real_body_len};
use crate::fuzzy;

/// Tristar: a three-candle reversal pattern with three dojis.
///
/// Must have:
/// - three consecutive doji candles,
/// - if the second doji gaps up from the first and the third does not
/// close higher than the second: bearish,
/// - if the second doji gaps down from the first and the third does not
/// close lower than the second: bullish.
///
/// Category A: fixed direction per branch (bullish or bearish).
///
/// Returns:
/// Continuous float in [-100, +100].
pub fn tristar(cp: &CandlestickPatterns) -> f64 {
    if !cp.enough(3, &[&cp.doji_body]) {
        return 0.0;
    }

    let b1 = cp.bar(3);
    let b2 = cp.bar(2);
    let b3 = cp.bar(1);

    // Fuzzy: all three must be dojis.
    let mu_doji1 = cp.mu_less(real_body_len(b1.o, b1.c), &cp.doji_body, 3);
    let mu_doji2 = cp.mu_less(real_body_len(b2.o, b2.c), &cp.doji_body, 2);
    let mu_doji3 = cp.mu_less(real_body_len(b3.o, b3.c), &cp.doji_body, 1);

    // Bearish: second gaps up, third is not higher than second -- crisp direction checks.
    if is_real_body_gap_up(b1.o, b1.c, b2.o, b2.c)
        && f64::max(b3.o, b3.c) < f64::max(b2.o, b2.c)
    {
        let conf = fuzzy::t_product_all(&[mu_doji1, mu_doji2, mu_doji3]);
        return -conf * 100.0;
    }

    // Bullish: second gaps down, third is not lower than second.
    if is_real_body_gap_down(b1.o, b1.c, b2.o, b2.c)
        && f64::min(b3.o, b3.c) > f64::min(b2.o, b2.c)
    {
        let conf = fuzzy::t_product_all(&[mu_doji1, mu_doji2, mu_doji3]);
        return conf * 100.0;
    }

    0.0
}
