//! Up/Down-Gap Side-By-Side White Lines pattern (3-candle).

use crate::candlestick_patterns::CandlestickPatterns;
use crate::candlestick_patterns::core::{is_real_body_gap_down, is_real_body_gap_up, is_white, real_body_len};
use crate::fuzzy;

/// Up/Down-Gap Side-By-Side White Lines: a three-candle pattern.
///
/// Must have:
/// - first candle: white (for up gap) or black (for down gap),
/// - gap (up or down) between the first and second candle — both 2nd AND
/// 3rd must gap from the 1st,
/// - second and third candles are both white with similar size and
/// approximately the same open.
///
/// Up gap = bullish continuation, down gap = bearish continuation.
///
/// Category C: both branches evaluated, return stronger signal.
///
/// Returns:
/// Continuous float in [-100, +100].
pub fn up_down_gap_side_by_side_white_lines(cp: &CandlestickPatterns) -> f64 {
    if !cp.enough(3, &[&cp.near, &cp.equal]) {
        return 0.0;
    }

    let b1 = cp.bar(3);
    let b2 = cp.bar(2);
    let b3 = cp.bar(1);

    // Crisp: both 2nd and 3rd must be white.
    if !(is_white(b2.o, b2.c) && is_white(b3.o, b3.c)) {
        return 0.0;
    }

    // Both 2nd and 3rd must gap from 1st in the same direction -- crisp.
    let gap_up = is_real_body_gap_up(b1.o, b1.c, b2.o, b2.c) && is_real_body_gap_up(b1.o, b1.c, b3.o, b3.c);
    let gap_down = is_real_body_gap_down(b1.o, b1.c, b2.o, b2.c) && is_real_body_gap_down(b1.o, b1.c, b3.o, b3.c);

    if !(gap_up || gap_down) {
        return 0.0;
    }

    let rb2 = real_body_len(b2.o, b2.c);
    let rb3 = real_body_len(b3.o, b3.c);

    // Fuzzy: similar size and same open.
    let mu_near_size = cp.mu_less((rb2 - rb3).abs(), &cp.near, 2);
    let mu_equal_open = cp.mu_less((b3.o - b2.o).abs(), &cp.equal, 2);

    let conf = fuzzy::t_product_all(&[mu_near_size, mu_equal_open]);

    if gap_up {
        conf * 100.0
    } else {
        -conf * 100.0
    }
}
