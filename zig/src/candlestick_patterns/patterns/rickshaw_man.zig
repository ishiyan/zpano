/// Rickshaw Man: a one-candle doji pattern.
///
/// Must have:
/// - doji body (very small real body),
/// - two long shadows,
/// - body near the midpoint of the high-low range.
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

pub fn rickshawMan(self: *const CandlestickPatterns) f64 {
            if (!self.enough(1, &[_]*const CriterionState{&self.doji_body, &self.long_shadow, &self.near})) return 0.0;

    const b = self.bar(1);

    const hl_range = b.h - b.l;
    const near_avg = self.avgCS(&self.near, 1);
    const near_width = if (near_avg > 0.0) self.fuzz_ratio * near_avg else 0.0;

    const mu_doji = self.muLessCs(realBodyLen(b.o, b.c), &self.doji_body, 1);
    const mu_long_us = self.muGreaterCs(upperShadow(b.o, b.h, b.c), &self.long_shadow, 1);
    const mu_long_ls = self.muGreaterCs(lowerShadow(b.o, b.l, b.c), &self.long_shadow, 1);
    const midpoint = b.l + hl_range / 2.0;
    const mu_near_mid_lo = self.muLtRaw(@min(b.o, b.c), midpoint + near_avg, near_width);
    const mu_near_mid_hi = self.muGeRaw(@max(b.o, b.c), midpoint - near_avg, near_width);

    const confidence = operators.tProductAll(&.{mu_doji, mu_long_us, mu_long_ls, mu_near_mid_lo, mu_near_mid_hi});
    return confidence * 100.0;
}
