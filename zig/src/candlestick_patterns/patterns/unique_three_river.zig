/// Unique Three River: a three-candle bullish pattern.
///
/// Must have:
/// - first candle: long black,
/// - second candle: black harami (body within first body) with a lower
///   low than the first candle,
/// - third candle: small white, opens not lower than the second candle's
///   low.
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
const isWhite = cp.isWhite;
const realBodyLen = cp.realBodyLen;

pub fn uniqueThreeRiver(self: *const CandlestickPatterns) f64 {
            if (!self.enough(3, &[_]*const CriterionState{&self.long_body, &self.short_body})) return 0.0;

    const b1 = self.bar(3);
    const b2 = self.bar(2);
    const b3 = self.bar(1);

    // Crisp gates: colors.
    if (!(isBlack(b1.o, b1.c) and isBlack(b2.o, b2.c) and isWhite(b3.o, b3.c))) return 0.0;

    // Crisp: harami body containment and lower low.
    if (!(b2.c > b1.c and b2.o <= b1.o and b2.l < b1.l)) return 0.0;

    // Crisp: third opens not lower than second's low.
    if (!(b3.o >= b2.l)) return 0.0;

    // Fuzzy: first candle is long.
    const mu_long1 = self.muGreaterCs(realBodyLen(b1.o, b1.c), &self.long_body, 3);

    // Fuzzy: third candle is short.
    const mu_short3 = self.muLessCs(realBodyLen(b3.o, b3.c), &self.short_body, 1);

    const confidence = operators.tProductAll(&.{mu_long1, mu_short3});

    return confidence * 100.0;
}
