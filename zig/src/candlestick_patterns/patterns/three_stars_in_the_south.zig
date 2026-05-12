/// Three Stars In The South: a three-candle bullish pattern.
///
/// Must have:
/// - all three candles are black,
/// - first candle: long body with long lower shadow,
/// - second candle: smaller body, opens within or above prior range,
///   trades lower but its low does not go below the first candle's low,
/// - third candle: small marubozu (very short shadows) engulfed by the
///   second candle's range.
///
/// The meaning of "long" is specified with self._long_body.
/// The meaning of "short" is specified with self._short_body.
/// The meaning of "long" for shadows is specified with self._long_shadow.
/// The meaning of "very short" for shadows is specified with
/// self._very_short_shadow.
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
const lowerShadow = cp.lowerShadow;
const realBodyLen = cp.realBodyLen;
const upperShadow = cp.upperShadow;

pub fn threeStarsInTheSouth(self: *const CandlestickPatterns) f64 {
            if (!self.enough(3, &[_]*const CriterionState{&self.long_body, &self.short_body, &self.long_shadow, &self.very_short_shadow})) return 0.0;

    const b1 = self.bar(3);
    const b2 = self.bar(2);
    const b3 = self.bar(1);

    // Crisp gates: all black.
    if (!(isBlack(b1.o, b1.c) and isBlack(b2.o, b2.c) and isBlack(b3.o, b3.c))) return 0.0;

    const rb1 = realBodyLen(b1.o, b1.c);
    const rb2 = realBodyLen(b2.o, b2.c);

    // Crisp: second body smaller than first.
    if (!(rb2 < rb1)) return 0.0;

    // Crisp: second opens within or above prior range, low not below first's low.
    if (!(b2.o <= b1.h and b2.o >= b1.l and b2.l >= b1.l)) return 0.0;

    // Crisp: third engulfed by second's range.
    if (!(b3.h <= b2.h and b3.l >= b2.l)) return 0.0;

    // Fuzzy: first candle long body.
    const mu_long1 = self.muGreaterCs(rb1, &self.long_body, 3);

    // Fuzzy: first candle long lower shadow.
    const mu_ls1 = self.muGreaterCs(lowerShadow(b1.o, b1.l, b1.c), &self.long_shadow, 3);

    // Fuzzy: third candle short body.
    const mu_short3 = self.muLessCs(realBodyLen(b3.o, b3.c), &self.short_body, 1);

    // Fuzzy: third candle very short shadows (marubozu).
    const mu_vs_us3 = self.muLessCs(upperShadow(b3.o, b3.h, b3.c), &self.very_short_shadow, 1);
    const mu_vs_ls3 = self.muLessCs(lowerShadow(b3.o, b3.l, b3.c), &self.very_short_shadow, 1);

    const confidence = operators.tProductAll(&.{mu_long1, mu_ls1, mu_short3, mu_vs_us3, mu_vs_ls3});

    return confidence * 100.0;
}
