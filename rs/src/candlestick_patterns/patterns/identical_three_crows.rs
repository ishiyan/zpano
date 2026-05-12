//! Identical Three Crows pattern (3-candle bearish).

use crate::candlestick_patterns::CandlestickPatterns;
use crate::candlestick_patterns::core::{is_black, lower_shadow};
use crate::fuzzy;

/// Identical Three Crows: a three-candle bearish pattern.
///
/// Must have:
/// - three consecutive declining black candles,
/// - each opens very close to the prior candle's close (equal criterion),
/// - very short lower shadows.
///
/// The meaning of "equal" is specified with self._equal.
/// The meaning of "very short" for shadows is specified with
/// self._very_short_shadow.
///
/// Category A: always bearish (continuous).
///
/// Returns:
/// Continuous float in [-100, 0].  Always bearish.
pub fn identical_three_crows(cp: &CandlestickPatterns) -> f64 {
    if !cp.enough(3, &[&cp.equal, &cp.very_short_shadow]) {
        return 0.0;
    }

    let b1 = cp.bar(3);
    let b2 = cp.bar(2);
    let b3 = cp.bar(1);

    // Crisp gates: all black, declining closes.
    if !(is_black(b1.o, b1.c) && is_black(b2.o, b2.c) && is_black(b3.o, b3.c)) {
        return 0.0;
    }
    if !(b1.c > b2.c && b2.c > b3.c) {
        return 0.0;
    }

    // Fuzzy conditions.
    let mu_ls1 = cp.mu_less(lower_shadow(b1.o, b1.l, b1.c), &cp.very_short_shadow, 3);
    let mu_ls2 = cp.mu_less(lower_shadow(b2.o, b2.l, b2.c), &cp.very_short_shadow, 2);
    let mu_ls3 = cp.mu_less(lower_shadow(b3.o, b3.l, b3.c), &cp.very_short_shadow, 1);
    // Opens near prior close (equal criterion, two-sided band).
    let mu_eq2 = cp.mu_less((b2.o - b1.c).abs(), &cp.equal, 3);
    let mu_eq3 = cp.mu_less((b3.o - b2.c).abs(), &cp.equal, 2);

    let confidence = fuzzy::t_product_all(&[mu_ls1, mu_ls2, mu_ls3, mu_eq2, mu_eq3]);
    -confidence * 100.0
}
