/// Dark Cloud Cover: a two-candle bearish reversal pattern.
///
/// Must have:
/// - first candle: long white candle,
/// - second candle: black candle that opens above the prior high and
///   closes well within the first candle's real body (below the midpoint).
///
/// Returns:
///     Continuous float in [-100, 0].  More negative = stronger signal.

const cp = @import("../candlestick_patterns.zig");
const operators = @import("fuzzy").operators;

const CandlestickPatterns = cp.CandlestickPatterns;
const CriterionState = cp.CriterionState;
const isBlack = cp.isBlack;
const isWhite = cp.isWhite;
const realBodyLen = cp.realBodyLen;

const dark_cloud_cover_penetration_factor = 0.5;

pub fn darkCloudCover(self: *const CandlestickPatterns) f64 {
            if (!self.enough(2, &[_]*const CriterionState{&self.long_body})) return 0.0;

    const b1 = self.bar(2);
    const b2 = self.bar(1);

    // Color checks stay crisp
    if (!isWhite(b1.o, b1.c) or !isBlack(b2.o, b2.c)) {
        return 0.0;
    }

    const rb1 = realBodyLen(b1.o, b1.c);
    const eq_avg = self.avgCS(&self.equal, 1);
    const eq_width = if (eq_avg > 0.0) self.fuzz_ratio * eq_avg else 0.0;

    const mu_long = self.muGreaterCs(rb1, &self.long_body, 2);
    const mu_open_above = self.muGtRaw(b2.o, b1.h, eq_width);
    const pen_threshold = b1.c - rb1 * dark_cloud_cover_penetration_factor;
    const pen_product = rb1 * dark_cloud_cover_penetration_factor;
    const pen_width = if (pen_product > 0.0) self.fuzz_ratio * pen_product else 0.0;
    const mu_pen = self.muLtRaw(b2.c, pen_threshold, pen_width);
    const mu_above_open1 = self.muGtRaw(b2.c, b1.o, eq_width);

    const confidence = operators.tProductAll(&.{mu_long, mu_open_above, mu_pen, mu_above_open1});
    return -confidence * 100.0;
}
