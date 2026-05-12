/// Homing Pigeon: a two-candle bullish pattern.
///
/// Must have:
/// - first candle: long black,
/// - second candle: short black, real body engulfed by first candle's
///   real body.
///
/// The meaning of "long" is specified with self._long_body.
/// The meaning of "short" is specified with self._short_body.
///
/// Category A: always bullish (continuous).
///
/// Returns:
///     Continuous float in [0, 100].  Always bullish.

const cp = @import("../candlestick_patterns.zig");
const operators = @import("fuzzy").operators;

const CandlestickPatterns = cp.CandlestickPatterns;
const CriterionState = cp.CriterionState;
const isBlack = cp.isBlack;
const realBodyLen = cp.realBodyLen;

pub fn homingPigeon(self: *const CandlestickPatterns) f64 {
            if (!self.enough(2, &[_]*const CriterionState{&self.long_body, &self.short_body})) return 0.0;

    const b1 = self.bar(2);
    const b2 = self.bar(1);

    // Crisp gates: both black.
    if (!(isBlack(b1.o, b1.c) and isBlack(b2.o, b2.c))) return 0.0;

    // Fuzzy conditions.
    const mu_long1 = self.muGreaterCs(realBodyLen(b1.o, b1.c), &self.long_body, 2);
    const mu_short2 = self.muLessCs(realBodyLen(b2.o, b2.c), &self.short_body, 1);

    // Containment: second body engulfed by first body.
    // For black candles: open > close, so upper = open, lower = close.
    const eq_width = self.fuzz_ratio * self.avgCS(&self.equal, 2);
    const mu_enc_upper = self.muLtRaw(b2.o, b1.o, eq_width);
    const mu_enc_lower = self.muGtRaw(b2.c, b1.c, eq_width);

    const confidence = operators.tProductAll(&.{mu_long1, mu_short2, mu_enc_upper, mu_enc_lower});
    return confidence * 100.0;
}
