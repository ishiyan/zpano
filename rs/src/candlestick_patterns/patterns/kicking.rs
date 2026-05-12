//! Kicking pattern (2-candle).

use crate::candlestick_patterns::CandlestickPatterns;
use crate::candlestick_patterns::core::{is_high_low_gap_down, is_high_low_gap_up, lower_shadow, real_body_len, upper_shadow};
use crate::fuzzy;

/// Kicking: a two-candle pattern with opposite-color marubozus and gap.
///
/// Must have:
/// - first candle: marubozu (long body, very short shadows),
/// - second candle: opposite-color marubozu with a high-low gap,
/// - bullish: black marubozu followed by white marubozu gapping up,
/// - bearish: white marubozu followed by black marubozu gapping down.
///
/// Category B: direction from second candle's color.
///
/// Returns:
/// Continuous float in [-100, +100].
pub fn kicking(cp: &CandlestickPatterns) -> f64 {
    if !cp.enough(2, &[&cp.very_short_shadow, &cp.long_body]) {
        return 0.0;
    }

    let b1 = cp.bar(2);
    let b2 = cp.bar(1);

    // Crisp: opposite colors.
    let color1: i32 = if b1.c < b1.o { -1 } else { 1 };
    let color2: i32 = if b2.c < b2.o { -1 } else { 1 };
    if color1 == color2 { return 0.0; }

    // Crisp: gap check.
    if color1 == -1 && !is_high_low_gap_up(b1.h, b2.l) { return 0.0; }
    if color1 == 1 && !is_high_low_gap_down(b1.l, b2.h) { return 0.0; }

    // Fuzzy: both are marubozu (long body, very short shadows).
    let mu_long1 = cp.mu_greater(real_body_len(b1.o, b1.c), &cp.long_body, 2);
    let mu_vs_us1 = cp.mu_less(upper_shadow(b1.o, b1.h, b1.c), &cp.very_short_shadow, 2);
    let mu_vs_ls1 = cp.mu_less(lower_shadow(b1.o, b1.l, b1.c), &cp.very_short_shadow, 2);
    let mu_long2 = cp.mu_greater(real_body_len(b2.o, b2.c), &cp.long_body, 1);
    let mu_vs_us2 = cp.mu_less(upper_shadow(b2.o, b2.h, b2.c), &cp.very_short_shadow, 1);
    let mu_vs_ls2 = cp.mu_less(lower_shadow(b2.o, b2.l, b2.c), &cp.very_short_shadow, 1);

    let confidence = fuzzy::t_product_all(&[mu_long1, mu_vs_us1, mu_vs_ls1, mu_long2, mu_vs_us2, mu_vs_ls2]);
    (color2 as f64) * confidence * 100.0
}
