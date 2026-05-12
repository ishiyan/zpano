/// Piercing: a two-candle bullish reversal pattern.
///
/// Must have:
/// - first candle: long black,
/// - second candle: long white that opens below the prior low and closes
///   above the midpoint of the first candle's real body but within the body.
///
/// Returns:
///     Continuous float in [0, 100].  Higher = stronger signal.

const cp = @import("../candlestick_patterns.zig");
const operators = @import("fuzzy").operators;

const CandlestickPatterns = cp.CandlestickPatterns;
const CriterionState = cp.CriterionState;
const isBlack = cp.isBlack;
const isWhite = cp.isWhite;
const realBodyLen = cp.realBodyLen;

pub fn patternPiercing(self: *const CandlestickPatterns) f64 {
            if (!self.enough(2, &[_]*const CriterionState{&self.long_body})) return 0.0;

    const b1 = self.bar(2);
    const b2 = self.bar(1);

    // Color checks stay crisp.
    if (!isBlack(b1.o, b1.c) or !isWhite(b2.o, b2.c)) {
        return 0.0;
    }

    const rb1 = realBodyLen(b1.o, b1.c);
    const eq_avg = self.avgCS(&self.equal, 1);
    const eq_width = if (eq_avg > 0.0) self.fuzz_ratio * eq_avg else 0.0;

    const mu_long1 = self.muGreaterCs(rb1, &self.long_body, 2);
    const mu_long2 = self.muGreaterCs(realBodyLen(b2.o, b2.c), &self.long_body, 1);
    const mu_open_below = self.muLtRaw(b2.o, b1.l, eq_width);
    const pen_threshold = b1.c + rb1 * 0.5;
    const pen_width = if (rb1 > 0.0) self.fuzz_ratio * rb1 * 0.5 else 0.0;
    const mu_pen = self.muGtRaw(b2.c, pen_threshold, pen_width);
    const mu_below_open1 = self.muLtRaw(b2.c, b1.o, eq_width);

    const confidence = operators.tProductAll(&.{mu_long1, mu_long2, mu_open_below, mu_pen, mu_below_open1});
    return confidence * 100.0;
}
