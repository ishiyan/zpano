/// High Wave: a one-candle pattern.
///
/// Must have:
/// - short real body,
/// - very long upper shadow,
/// - very long lower shadow.
///
/// The meaning of "short" is specified with self._short_body.
/// The meaning of "very long" (shadow) is specified with self._very_long_shadow.
///
/// Category C: color determines sign.
///
/// Returns:
///     Continuous float in [-100, +100].

const cp = @import("../candlestick_patterns.zig");
const operators = @import("fuzzy").operators;

const CandlestickPatterns = cp.CandlestickPatterns;
const CriterionState = cp.CriterionState;
const isWhite = cp.isWhite;
const lowerShadow = cp.lowerShadow;
const realBodyLen = cp.realBodyLen;
const upperShadow = cp.upperShadow;

pub fn highWave(self: *const CandlestickPatterns) f64 {
            if (!self.enough(1, &[_]*const CriterionState{&self.short_body, &self.very_long_shadow})) return 0.0;

    const b = self.bar(1);
    const mu_short = self.muLessCs(realBodyLen(b.o, b.c), &self.short_body, 1);
    const mu_long_us = self.muGreaterCs(upperShadow(b.o, b.h, b.c), &self.very_long_shadow, 1);
    const mu_long_ls = self.muGreaterCs(lowerShadow(b.o, b.l, b.c), &self.very_long_shadow, 1);

    const confidence = operators.tProductAll(&.{mu_short, mu_long_us, mu_long_ls});
    if (isWhite(b.o, b.c)) {
        return confidence * 100.0;
    } else {
        return -confidence * 100.0;
    }
}
