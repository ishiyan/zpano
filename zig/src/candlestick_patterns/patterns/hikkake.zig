/// Hikkake: a three-candle pattern with stateful confirmation.
///
/// TA-Lib behavior:
/// - Detection bar: outputs +100.0 (bullish) or -100.0 (bearish)
/// - Confirmation bar (within 3 bars of detection): outputs +200.0 or -200.0
/// - If a new hikkake is detected on the same bar as a confirmation,
///   the new hikkake takes priority.
///
/// Must have:
/// - first and second candle: inside bar (2nd lower high, higher low)
/// - third candle: lower high AND lower low (bull) or higher high AND
///   higher low (bear)
///
/// Confirmation: close > high of 2nd candle (bull) or close < low of
/// 2nd candle (bear) within 3 bars.
///
/// Returns:
///     +100.0/-100.0 for initial detection, +200.0/-200.0 for confirmation,
///     0.0 for no pattern.

const cp = @import("../candlestick_patterns.zig");

const CandlestickPatterns = cp.CandlestickPatterns;
const CriterionState = cp.CriterionState;

pub fn patternHikkake(self: *const CandlestickPatterns) f64 {
            // Check if there's a newer hikkake at this position
            // Check if confirmation already happened
            if (!self.enough(3, &[_]*const CriterionState{})) return 0.0;

    // Check for new hikkake pattern at current bar.
    const b1 = self.bar(3);
    const b2 = self.bar(2);
    const b3 = self.bar(1);

    // Inside bar check.
    if (b2.h < b1.h and b2.l > b1.l) {
        // Bullish: 3rd has lower high AND lower low.
        if (b3.h < b2.h and b3.l < b2.l) {
            return 100.0;
        }
        // Bearish: 3rd has higher high AND higher low.
        if (b3.h > b2.h and b3.l > b2.l) {
            return -100.0;
        }
    }

    // No new pattern -- check for confirmation of a recent hikkake.
    // No new pattern — check for confirmation of a recent hikkake.
    // Look back 1-3 bars for a hikkake pattern.
    var lookback: usize = 1;
    while (lookback <= 3) : (lookback += 1) {
        const n = 3 + lookback;
        if (!self.enough(n, &[_]*const CriterionState{})) {
            break;
        }

        const p1 = self.bar(n);
        const p2 = self.bar(n - 1);
        const p3 = self.bar(n - 2);

        if (!(p2.h < p1.h and p2.l > p1.l)) {
            continue;
        }

        var pattern_result: f64 = undefined;
        if (p3.h < p2.h and p3.l < p2.l) {
            pattern_result = 100.0;
        } else if (p3.h > p2.h and p3.l > p2.l) {
            pattern_result = -100.0;
        } else {
            continue;
        }

        // Check that no intervening bar already confirmed or re-detected.
        // If there's a newer hikkake between the pattern and current bar,
        // the older one is superseded.
        var superseded: bool = false;
        var gap: usize = 1;
        while (gap < lookback) : (gap += 1) {
            const gb = n - 2 - gap;
            if (gb < 1) {
                break;
            }
            if (self.enough(gb + 2, &[_]*const CriterionState{})) {
                const ga = self.bar(gb + 2);
                const gbo = self.bar(gb + 1);
                const gc = self.bar(gb);
                if (gbo.h < ga.h and gbo.l > ga.l and ((gc.h < gbo.h and gc.l < gbo.l) or (gc.h > gbo.h and gc.l > gbo.l)))
                {
                    superseded = true;
                    break;
                }
            }
            if (self.enough(gb, &[_]*const CriterionState{})) {
                const cc_gap = self.bar(gb);
                if (pattern_result > 0.0 and cc_gap.c > p2.h) {
                    superseded = true;
                    break;
                }
                if (pattern_result < 0.0 and cc_gap.c < p2.l) {
                    superseded = true;
                    break;
                }
            }
        }

        if (superseded) {
            continue;
        }

        const cc = self.bar(1);
        if (pattern_result > 0.0 and cc.c > p2.h) {
            return 200.0;
        }
        if (pattern_result < 0.0 and cc.c < p2.l) {
            return -200.0;
        }
    }

    return 0.0;
}
