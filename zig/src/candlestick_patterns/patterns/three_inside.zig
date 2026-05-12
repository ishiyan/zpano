/// Three Inside Up/Down: a three-candle reversal pattern.
///
/// Three Inside Up (bullish):
/// - first candle: long black,
/// - second candle: short, engulfed by the first candle's real body,
/// - third candle: white, closes above the first candle's open.
///
/// Three Inside Down (bearish):
/// - first candle: long white,
/// - second candle: short, engulfed by the first candle's real body,
/// - third candle: black, closes below the first candle's open.
///
/// The meaning of "long" is specified with self._long_body.
/// The meaning of "short" is specified with self._short_body.
///
/// Category C: both branches evaluated, return stronger signal.
///
/// Returns:
///     Continuous float in [-100, +100].

const cp = @import("../candlestick_patterns.zig");
const operators = @import("fuzzy").operators;

const CandlestickPatterns = cp.CandlestickPatterns;
const CriterionState = cp.CriterionState;
const isBlack = cp.isBlack;
const isWhite = cp.isWhite;
const realBodyLen = cp.realBodyLen;

pub fn threeInside(self: *const CandlestickPatterns) f64 {
            if (!self.enough(3, &[_]*const CriterionState{&self.long_body, &self.short_body})) return 0.0;

    const b1 = self.bar(3);
    const b2 = self.bar(2);
    const b3 = self.bar(1);

    // Shared fuzzy conditions.
    const mu_long1 = self.muGreaterCs(realBodyLen(b1.o, b1.c), &self.long_body, 3);
    const mu_short2 = self.muLessCs(realBodyLen(b2.o, b2.c), &self.short_body, 2);

    // Fuzzy containment: 1st body encloses 2nd body.
    const eq_avg = self.avgCS(&self.equal, 2);
    const eq_width = if (eq_avg > 0.0) self.fuzz_ratio * eq_avg else 0.0;
    const mu_enc_upper = self.muGeRaw(@max(b1.o, b1.c), @max(b2.o, b2.c), eq_width);
    const mu_enc_lower = self.muLtRaw(@min(b1.o, b1.c), @min(b2.o, b2.c), eq_width);

    // Three Inside Up: long black, short engulfed, white closes above 1st open.
    var bull_signal: f64 = 0.0;
    if (isBlack(b1.o, b1.c) and isWhite(b3.o, b3.c)) {
        const rb1 = realBodyLen(b1.o, b1.c);
        const width = if (rb1 > 0.0) self.fuzz_ratio * rb1 else 0.0;
        const mu_close_above = self.muGtRaw(b3.c, b1.o, width);
        const conf = operators.tProductAll(&.{mu_long1, mu_short2, mu_enc_upper, mu_enc_lower, mu_close_above});
        bull_signal = conf * 100.0;
    }

    // Three Inside Down: long white, short engulfed, black closes below 1st open.
    var bear_signal: f64 = 0.0;
    if (isWhite(b1.o, b1.c) and isBlack(b3.o, b3.c)) {
        const rb1 = realBodyLen(b1.o, b1.c);
        const width = if (rb1 > 0.0) self.fuzz_ratio * rb1 else 0.0;
        const mu_close_below = self.muLtRaw(b3.c, b1.o, width);
        const conf = operators.tProductAll(&.{mu_long1, mu_short2, mu_enc_upper, mu_enc_lower, mu_close_below});
        bear_signal = -conf * 100.0;
    }

    return if (@abs(bull_signal) >= @abs(bear_signal)) bull_signal else bear_signal;
}
