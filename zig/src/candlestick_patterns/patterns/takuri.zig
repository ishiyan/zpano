/// Takuri (Dragonfly Doji with very long lower shadow): a one-candle pattern.
///
/// A doji body with a very short upper shadow and a very long lower shadow.
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

pub fn patternTakuri(self: *const CandlestickPatterns) f64 {
            if (!self.enough(1, &[_]*const CriterionState{&self.doji_body, &self.very_short_shadow, &self.very_long_shadow})) return 0.0;

    const b = self.bar(1);

    const mu_doji = self.muLessCs(realBodyLen(b.o, b.c), &self.doji_body, 1);
    const mu_short_us = self.muLessCs(upperShadow(b.o, b.h, b.c), &self.very_short_shadow, 1);
    const mu_long_ls = self.muGreaterCs(lowerShadow(b.o, b.l, b.c), &self.very_long_shadow, 1);

    const confidence = operators.tProductAll(&.{mu_doji, mu_short_us, mu_long_ls});
    return confidence * 100.0;
}
