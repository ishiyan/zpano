/// Hanging Man: a two-candle bearish pattern.
///
/// Must have:
/// - small real body,
/// - long lower shadow,
/// - no or very short upper shadow,
/// - body is above or near the highs of the previous candle.
///
/// Returns:
///     Continuous float in [-100, 0].  More negative = stronger signal.

const cp = @import("../candlestick_patterns.zig");
const operators = @import("fuzzy").operators;

const CandlestickPatterns = cp.CandlestickPatterns;
const CriterionState = cp.CriterionState;
const lowerShadow = cp.lowerShadow;
const realBodyLen = cp.realBodyLen;
const upperShadow = cp.upperShadow;

pub fn hangingMan(self: *const CandlestickPatterns) f64 {
            if (!self.enough(2, &[_]*const CriterionState{&self.short_body, &self.long_shadow, &self.very_short_shadow, &self.near})) return 0.0;

    const b1 = self.bar(2);
    const b2 = self.bar(1);

    const near_avg = self.avgCS(&self.near, 2);
    const near_width = if (near_avg > 0.0) self.fuzz_ratio * near_avg else 0.0;

    const mu_short = self.muLessCs(realBodyLen(b2.o, b2.c), &self.short_body, 1);
    const mu_long_ls = self.muGreaterCs(lowerShadow(b2.o, b2.l, b2.c), &self.long_shadow, 1);
    const mu_short_us = self.muLessCs(upperShadow(b2.o, b2.h, b2.c), &self.very_short_shadow, 1);
    const mu_near_high = self.muGeRaw(@min(b2.o, b2.c), b1.h - near_avg, near_width);

    const confidence = operators.tProductAll(&.{mu_short, mu_long_ls, mu_short_us, mu_near_high});
    return -confidence * 100.0;
}
