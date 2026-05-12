/// Spinning Top: a one-candle pattern.
///
/// A candle with a small body and shadows longer than the body on both sides.
///
/// The meaning of "short" is specified with self._short_body.
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

pub fn spinningTop(self: *const CandlestickPatterns) f64 {
            if (!self.enough(1, &[_]*const CriterionState{&self.short_body})) return 0.0;

    const b = self.bar(1);

    const rb = realBodyLen(b.o, b.c);

    const mu_short = self.muLessCs(rb, &self.short_body, 1);

    // Shadows > body: positional comparisons.
    const us = upperShadow(b.o, b.h, b.c);
    const ls = lowerShadow(b.o, b.l, b.c);
    const width_us = if (rb > 0.0) self.fuzz_ratio * rb else 0.0;
    const width_ls = if (rb > 0.0) self.fuzz_ratio * rb else 0.0;
    const mu_us_gt_rb = self.muGtRaw(us, rb, width_us);
    const mu_ls_gt_rb = self.muGtRaw(ls, rb, width_ls);

    const confidence = operators.tProductAll(&.{mu_short, mu_us_gt_rb, mu_ls_gt_rb});

    if (isWhite(b.o, b.c)) {
        return confidence * 100.0;
    }
    if (isBlack(b.o, b.c)) {
        return -confidence * 100.0;
    }
    return 0.0;
}
