//! Kicking By Length pattern (2-candle).

use crate::candlestick_patterns::CandlestickPatterns;
use crate::candlestick_patterns::core::{is_high_low_gap_down, is_high_low_gap_up, lower_shadow, real_body_len, upper_shadow};
use crate::fuzzy;

/// Kicking By Length: like Kicking but direction determined by the longer marubozu.
///
/// Must have:
/// - first candle: marubozu (long body, very short shadows),
/// - second candle: opposite-color marubozu with a high-low gap,
/// - bull/bear determined by which marubozu has the longer real body.
///
/// Category B: direction from longer marubozu's color.
///
/// Returns:
/// Continuous float in [-100, +100].
pub fn kicking_by_length(cp: &CandlestickPatterns) -> f64 {
    if !cp.enough(2, &[&cp.very_short_shadow, &cp.long_body]) {
        return 0.0;
    }

    let b1 = cp.bar(2);
    let b2 = cp.bar(1);

    // Crisp: opposite colors.
    let color1: i32 = if b1.c < b1.o { -1 } else { 1 };
    let color2: i32 = if b2.c < b2.o { -1 } else { 1 };
    // Direction determined by the longer marubozu's color.
    if color1 == color2 { return 0.0; }

    // Crisp: gap check.
    let has_gap = if color1 == -1 {
        is_high_low_gap_up(b1.h, b2.l)
    } else {
        is_high_low_gap_down(b1.l, b2.h)
    };
    if !has_gap { return 0.0; }

    let rb1 = real_body_len(b1.o, b1.c);
    let rb2 = real_body_len(b2.o, b2.c);

    // Fuzzy: both are marubozu (long body, very short shadows).
    let mu_long1 = cp.mu_greater(rb1, &cp.long_body, 2);
    let mu_vs_us1 = cp.mu_less(upper_shadow(b1.o, b1.h, b1.c), &cp.very_short_shadow, 2);
    let mu_vs_ls1 = cp.mu_less(lower_shadow(b1.o, b1.l, b1.c), &cp.very_short_shadow, 2);
    let mu_long2 = cp.mu_greater(rb2, &cp.long_body, 1);
    let mu_vs_us2 = cp.mu_less(upper_shadow(b2.o, b2.h, b2.c), &cp.very_short_shadow, 1);
    let mu_vs_ls2 = cp.mu_less(lower_shadow(b2.o, b2.l, b2.c), &cp.very_short_shadow, 1);

    let confidence = fuzzy::t_product_all(&[mu_long1, mu_vs_us1, mu_vs_ls1, mu_long2, mu_vs_us2, mu_vs_ls2]);

    let direction = if rb2 > rb1 { color2 } else { color1 };
    (direction as f64) * confidence * 100.0
}
