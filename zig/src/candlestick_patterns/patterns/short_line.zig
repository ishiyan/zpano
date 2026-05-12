/// Short Line: a one-candle pattern.
///
/// A candle with a short body, short upper shadow, and short lower shadow.
///
/// The meaning of "short" for body is specified with self._short_body.
/// The meaning of "short" for shadows is specified with self._short_shadow.
///
/// Category C: color determines sign.
///
/// Returns:
///     Continuous float in [-100, +100].

const cp = @import("../candlestick_patterns.zig");
const operators = @import("fuzzy").operators;

const CandlestickPatterns = cp.CandlestickPatterns;
const CriterionState = cp.CriterionState;
const isBlack = cp.isBlack;
const isWhite = cp.isWhite;
const lowerShadow = cp.lowerShadow;
const realBodyLen = cp.realBodyLen;
const upperShadow = cp.upperShadow;

pub fn shortLine(self: *const CandlestickPatterns) f64 {
            if (!self.enough(1, &[_]*const CriterionState{&self.short_body, &self.short_shadow})) return 0.0;

    const b = self.bar(1);

    const mu_short_body = self.muLessCs(realBodyLen(b.o, b.c), &self.short_body, 1);
    const mu_short_us = self.muLessCs(upperShadow(b.o, b.h, b.c), &self.short_shadow, 1);
    const mu_short_ls = self.muLessCs(lowerShadow(b.o, b.l, b.c), &self.short_shadow, 1);

    const confidence = operators.tProductAll(&.{mu_short_body, mu_short_us, mu_short_ls});

    if (isWhite(b.o, b.c)) {
        return confidence * 100.0;
    }
    if (isBlack(b.o, b.c)) {
        return -confidence * 100.0;
    }
    return 0.0;
}
