/// Engulfing: a two-candle reversal pattern.
///
/// Must have:
/// - first candle and second candle have opposite colors,
/// - second candle's real body engulfs the first (at least one end strictly
///   exceeds, the other can match).
///
// Direction sign from 2nd candle (TA-Lib: c >= o is bullish).
/// Category B: direction from 2nd candle color (continuous).
/// Opposite-color check stays crisp (doji edge case).
///
/// Returns:
///     Continuous float in [-100, +100].  Sign from 2nd candle direction.

const cp = @import("../candlestick_patterns.zig");
const operators = @import("fuzzy").operators;

const CandlestickPatterns = cp.CandlestickPatterns;
const CriterionState = cp.CriterionState;

pub fn patternEngulfing(self: *const CandlestickPatterns) f64 {
            if (!self.enough(2, &[_]*const CriterionState{})) return 0.0;

    const b1 = self.bar(2);
    const b2 = self.bar(1);

    // Opposite colors — crisp gate (TA-Lib convention: c >= o is white).
    const color1: i32 = if (b1.c < b1.o) -1 else 1;
    const color2: i32 = if (b2.c < b2.o) -1 else 1;
    if (color1 == color2) return 0.0;

    // Fuzzy engulfment: 2nd body upper >= 1st body upper AND
    // 2nd body lower <= 1st body lower.
    const upper1 = @max(b1.o, b1.c);
    const lower1 = @min(b1.o, b1.c);
    const upper2 = @max(b2.o, b2.c);
    const lower2 = @min(b2.o, b2.c);

    // Width based on the equal criterion for tight comparisons.
    const eq_avg = self.avgCS(&self.equal, 1);
    const eq_width = if (eq_avg > 0.0) self.fuzz_ratio * eq_avg else 0.0;

    const mu_upper = self.muGeRaw(upper2, upper1, eq_width);
    const mu_lower = self.muLtRaw(lower2, lower1, eq_width);

    const confidence = operators.tProductAll(&.{mu_upper, mu_lower});
    const direction: f64 = if (b2.c < b2.o) -1.0 else 1.0;
    return direction * confidence * 100.0;
}
