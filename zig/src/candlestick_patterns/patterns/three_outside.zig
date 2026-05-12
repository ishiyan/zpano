/// Three Outside Up/Down: a three-candle reversal pattern.
///
/// Must have:
/// - first and second candles form an engulfing pattern,
/// - third candle confirms the direction by closing higher (up) or
///   lower (down).
///
/// Three Outside Up: first candle is black, second is white engulfing
/// the first, third closes higher than the second.
///
/// Three Outside Down: first candle is white, second is black engulfing
/// the first, third closes lower than the second.
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

pub fn threeOutside(self: *const CandlestickPatterns) f64 {
            if (!self.enough(3, &[_]*const CriterionState{})) return 0.0;

    const b1 = self.bar(3);
    const b2 = self.bar(2);
    const b3 = self.bar(1);

    // Fuzzy engulfment width.
    const eq_avg = self.avgCS(&self.equal, 1);
    const eq_width = if (eq_avg > 0.0) self.fuzz_ratio * eq_avg else 0.0;

    // Three Outside Up: black + white engulfing + 3rd closes higher.
    var bull_signal: f64 = 0.0;
    if (isBlack(b1.o, b1.c) and isWhite(b2.o, b2.c)) {
        const mu_enc_upper = self.muGeRaw(@max(b2.o, b2.c), @max(b1.o, b1.c), eq_width);
        const mu_enc_lower = self.muLtRaw(@min(b2.o, b2.c), @min(b1.o, b1.c), eq_width);
        const rb2 = realBodyLen(b2.o, b2.c);
        const width = if (rb2 > 0.0) self.fuzz_ratio * rb2 else 0.0;
        const mu_close_higher = self.muGtRaw(b3.c, b2.c, width);
        const conf = operators.tProductAll(&.{mu_enc_upper, mu_enc_lower, mu_close_higher});
        bull_signal = conf * 100.0;
    }

    // Three Outside Down: white + black engulfing + 3rd closes lower.
    var bear_signal: f64 = 0.0;
    if (isWhite(b1.o, b1.c) and isBlack(b2.o, b2.c)) {
        const mu_enc_upper = self.muGeRaw(@max(b2.o, b2.c), @max(b1.o, b1.c), eq_width);
        const mu_enc_lower = self.muLtRaw(@min(b2.o, b2.c), @min(b1.o, b1.c), eq_width);
        const rb2 = realBodyLen(b2.o, b2.c);
        const width = if (rb2 > 0.0) self.fuzz_ratio * rb2 else 0.0;
        const mu_close_lower = self.muLtRaw(b3.c, b2.c, width);
        const conf = operators.tProductAll(&.{mu_enc_upper, mu_enc_lower, mu_close_lower});
        bear_signal = -conf * 100.0;
    }

    return if (@abs(bull_signal) >= @abs(bear_signal)) bull_signal else bear_signal;
}
