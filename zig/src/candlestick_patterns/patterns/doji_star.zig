/// Doji Star: a two-candle reversal pattern.
///
/// Must have:
/// - first candle: long real body,
/// - second candle: doji that gaps away from the first candle.
///
/// - bearish: first candle is long white, doji gaps up,
/// - bullish: first candle is long black, doji gaps down.
///
/// The meaning of "long" is specified with self._long_body.
/// The meaning of "doji" is specified with self._doji_body.
///
// Crisp gates: gap direction must match color.
// Direction: opposite of 1st candle color.
/// Category B: direction from 1st candle color (continuous).
///
/// Returns:
///     Continuous float in [-100, +100].

const cp = @import("../candlestick_patterns.zig");
const operators = @import("fuzzy").operators;

const CandlestickPatterns = cp.CandlestickPatterns;
const CriterionState = cp.CriterionState;
const isRealBodyGapDown = cp.isRealBodyGapDown;
const isRealBodyGapUp = cp.isRealBodyGapUp;
const realBodyLen = cp.realBodyLen;

pub fn dojiStar(self: *const CandlestickPatterns) f64 {
            if (!self.enough(2, &[_]*const CriterionState{&self.long_body, &self.doji_body})) return 0.0;

    const b1 = self.bar(2);
    const b2 = self.bar(1);

    const color1: i32 = if (b1.c < b1.o) -1 else 1;

    if (color1 == 1 and !isRealBodyGapUp(b1.o, b1.c, b2.o, b2.c)) {
        return 0.0;
    }
    if (color1 == -1 and !isRealBodyGapDown(b1.o, b1.c, b2.o, b2.c)) {
        return 0.0;
    }

    // Fuzzy conditions.
    const mu_long1 = self.muGreaterCs(realBodyLen(b1.o, b1.c), &self.long_body, 2);
    const mu_doji2 = self.muLessCs(realBodyLen(b2.o, b2.c), &self.doji_body, 1);

    const confidence = operators.tProductAll(&.{mu_long1, mu_doji2});
    const direction: f64 = if (color1 == -1) 1.0 else -1.0;
    return direction * confidence * 100.0;
}
