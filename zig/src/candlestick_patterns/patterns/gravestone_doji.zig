/// Gravestone Doji: a one-candle pattern.
///
/// Must have:
/// - doji body (very small real body relative to high-low range),
/// - no or very short lower shadow,
/// - upper shadow is not very short.
///
/// Returns:
///     Continuous float in [0, 100].  Higher = stronger signal.

const cp = @import("../candlestick_patterns.zig");
const operators = @import("fuzzy").operators;

const CandlestickPatterns = cp.CandlestickPatterns;
const CriterionState = cp.CriterionState;
const lowerShadow = cp.lowerShadow;
const realBodyLen = cp.realBodyLen;
const upperShadow = cp.upperShadow;

pub fn gravestoneDoji(self: *const CandlestickPatterns) f64 {
            if (!self.enough(1, &[_]*const CriterionState{&self.doji_body, &self.very_short_shadow})) return 0.0;

    const b = self.bar(1);
    const mu_doji = self.muLessCs(realBodyLen(b.o, b.c), &self.doji_body, 1);
    const mu_short_ls = self.muLessCs(lowerShadow(b.o, b.l, b.c), &self.very_short_shadow, 1);
    const mu_long_us = self.muGreaterCs(upperShadow(b.o, b.h, b.c), &self.very_short_shadow, 1);

    const confidence = operators.tProductAll(&.{mu_doji, mu_short_ls, mu_long_us});
    return confidence * 100.0;
}
