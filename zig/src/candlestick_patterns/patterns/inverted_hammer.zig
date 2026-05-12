/// Inverted Hammer: a two-candle bullish pattern.
///
/// Must have:
/// - small real body,
/// - long upper shadow,
/// - very short lower shadow,
/// - gap down from the previous candle's real body.
///
/// Returns:
///     Continuous float in [0, 100].  Higher = stronger signal.

const cp = @import("../candlestick_patterns.zig");
const operators = @import("fuzzy").operators;

const CandlestickPatterns = cp.CandlestickPatterns;
const CriterionState = cp.CriterionState;
const isRealBodyGapDown = cp.isRealBodyGapDown;
const lowerShadow = cp.lowerShadow;
const realBodyLen = cp.realBodyLen;
const upperShadow = cp.upperShadow;

pub fn invertedHammer(self: *const CandlestickPatterns) f64 {
            if (!self.enough(2, &[_]*const CriterionState{&self.short_body, &self.long_shadow, &self.very_short_shadow})) return 0.0;

    const b1 = self.bar(2);
    const b2 = self.bar(1);

    if (!isRealBodyGapDown(b1.o, b1.c, b2.o, b2.c)) {
        return 0.0;
    }

    const mu_short = self.muLessCs(realBodyLen(b2.o, b2.c), &self.short_body, 1);
    const mu_long_us = self.muGreaterCs(upperShadow(b2.o, b2.h, b2.c), &self.long_shadow, 1);
    const mu_short_ls = self.muLessCs(lowerShadow(b2.o, b2.l, b2.c), &self.very_short_shadow, 1);

    const confidence = operators.tProductAll(&.{mu_short, mu_long_us, mu_short_ls});
    return confidence * 100.0;
}
