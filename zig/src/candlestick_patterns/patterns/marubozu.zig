/// Marubozu: a one-candle pattern.
///
/// Must have:
/// - long real body,
/// - very short upper shadow,
/// - very short lower shadow.
///
/// The meaning of "long" is specified with self._long_body.
/// The meaning of "very short" for shadows is specified with
/// self._very_short_shadow.
///
// Crisp direction from color.
/// Category B: direction from candle color.
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

pub fn patternMarubozu(self: *const CandlestickPatterns) f64 {
            if (!self.enough(1, &[_]*const CriterionState{&self.long_body, &self.very_short_shadow})) return 0.0;

    const b = self.bar(1);
    // Fuzzy: long body, very short shadows.
    const mu_long = self.muGreaterCs(realBodyLen(b.o, b.c), &self.long_body, 1);
    const mu_us = self.muLessCs(upperShadow(b.o, b.h, b.c), &self.very_short_shadow, 1);
    const mu_ls = self.muLessCs(lowerShadow(b.o, b.l, b.c), &self.very_short_shadow, 1);

    const confidence = operators.tProductAll(&.{mu_long, mu_us, mu_ls});
    const direction: i32 = if (!isWhite(b.o, b.c)) -1 else 1;
    return @as(f64, @floatFromInt(direction)) * confidence * 100.0;
}
