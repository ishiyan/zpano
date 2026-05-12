/// Long Legged Doji: a one-candle pattern.
///
/// Must have:
/// - doji body (very small real body),
/// - one or both shadows are long.
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

pub fn longLeggedDoji(self: *const CandlestickPatterns) f64 {
            if (!self.enough(1, &[_]*const CriterionState{&self.doji_body, &self.long_shadow})) return 0.0;

    const b = self.bar(1);
    const mu_doji = self.muLessCs(realBodyLen(b.o, b.c), &self.doji_body, 1);
    const mu_long_us = self.muGreaterCs(upperShadow(b.o, b.h, b.c), &self.long_shadow, 1);
    const mu_long_ls = self.muGreaterCs(lowerShadow(b.o, b.l, b.c), &self.long_shadow, 1);
    const mu_any_long = operators.sMax(mu_long_us, mu_long_ls);

    const confidence = operators.tProductAll(&.{mu_doji, mu_any_long});
    return confidence * 100.0;
}
