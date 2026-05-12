//! Hikkake pattern (3-candle) with optional confirmation.

use crate::candlestick_patterns::CandlestickPatterns;

/// Hikkake: a three-candle pattern with stateful confirmation.
///
/// TA-Lib behavior:
/// - Detection bar: outputs +100.0 (bullish) or -100.0 (bearish)
/// - Confirmation bar (within 3 bars of detection): outputs +200.0 or -200.0
/// - If a new hikkake is detected on the same bar as a confirmation,
/// the new hikkake takes priority.
///
/// Must have:
/// - first and second candle: inside bar (2nd lower high, higher low)
/// - third candle: lower high AND lower low (bull) or higher high AND
/// higher low (bear)
///
/// Confirmation: close > high of 2nd candle (bull) or close < low of
/// 2nd candle (bear) within 3 bars.
///
/// Returns:
/// +100.0/-100.0 for initial detection, +200.0/-200.0 for confirmation,
/// 0.0 for no pattern.
pub fn hikkake(cp: &CandlestickPatterns) -> f64 {
    // Check if there's a newer hikkake at this position
    if !cp.enough(3, &[]) {
        return 0.0;
    }

    // Check for new hikkake pattern at current bar.
    let b1 = cp.bar(3);
    let b2 = cp.bar(2);
    let b3 = cp.bar(1);

    // Inside bar check.
    if b2.h < b1.h && b2.l > b1.l {
        // Bullish: 3rd has lower high AND lower low.
        if b3.h < b2.h && b3.l < b2.l {
            return 100.0;
        }
        // Bearish: 3rd has higher high AND higher low.
        if b3.h > b2.h && b3.l > b2.l {
            return -100.0;
        }
    }

    // No new pattern -- check for confirmation of a recent hikkake.
    // Look back 1-3 bars for a hikkake pattern.
    for lookback in 1..=3 {
        let n = 3 + lookback;
        if !cp.enough(n, &[]) {
            break;
        }

        let p1 = cp.bar(n);
        let p2 = cp.bar(n - 1);
        let p3 = cp.bar(n - 2);

        // Must be a valid hikkake at that position.
        if !(p2.h < p1.h && p2.l > p1.l) {
            continue;
        }

        let pattern_result;
        if p3.h < p2.h && p3.l < p2.l {
            pattern_result = 100.0;
        } else if p3.h > p2.h && p3.l > p2.l {
            pattern_result = -100.0;
        } else {
            continue;
        }

        // Check that no intervening bar already confirmed or re-detected.
        // If there's a newer hikkake between the pattern and current bar,
        // the older one is superseded.
        let mut superseded = false;
        for gap in 1..lookback {
            let gb = n - 2 - gap;
            if gb < 1 {
                break;
            }
            if cp.enough(gb + 2, &[]) {
                let ga = cp.bar(gb + 2);
                let gbo = cp.bar(gb + 1);
                let gc = cp.bar(gb);
                if gbo.h < ga.h && gbo.l > ga.l
                    && ((gc.h < gbo.h && gc.l < gbo.l)
                        || (gc.h > gbo.h && gc.l > gbo.l))
                {
                    superseded = true;
                    break;
                }
            }
            // Check if confirmation already happened.
            if cp.enough(gb, &[]) {
                let cc_gap = cp.bar(gb);
                if pattern_result > 0.0 && cc_gap.c > p2.h {
                    superseded = true;
                    break;
                }
                if pattern_result < 0.0 && cc_gap.c < p2.l {
                    superseded = true;
                    break;
                }
            }
        }

        if superseded {
            continue;
        }

        // Current bar confirms?
        let cc = cp.bar(1);
        if pattern_result > 0.0 && cc.c > p2.h {
            return 200.0;
        }
        if pattern_result < 0.0 && cc.c < p2.l {
            return -200.0;
        }
    }

    0.0
}
