/// Three White Soldiers: a three-candle bullish pattern.
///
/// Must have:
/// - three consecutive white candles with consecutively higher closes,
/// - all three have very short upper shadows,
/// - each opens within or near the prior candle's real body,
/// - none is far shorter than the prior candle,
/// - third candle is not short.
///
/// Category A: always bullish (continuous).
///
/// Returns:
///     Continuous float in [0, 100].  Always bullish.

const cp = @import("../candlestick_patterns.zig");
const operators = @import("fuzzy").operators;

const CandlestickPatterns = cp.CandlestickPatterns;
const CriterionState = cp.CriterionState;
const isWhite = cp.isWhite;
const realBodyLen = cp.realBodyLen;
const upperShadow = cp.upperShadow;

pub fn threeWhiteSoldiers(self: *const CandlestickPatterns) f64 {
            if (!self.enough(3, &[_]*const CriterionState{&self.short_body, &self.very_short_shadow, &self.near, &self.far})) return 0.0;

    const b1 = self.bar(3);
    const b2 = self.bar(2);
    const b3 = self.bar(1);

    // Crisp gates: all white with consecutively higher closes.
    if (!(isWhite(b1.o, b1.c) and isWhite(b2.o, b2.c) and isWhite(b3.o, b3.c) and b3.c > b2.c and b2.c > b1.c)) return 0.0;

    const rb1 = realBodyLen(b1.o, b1.c);
    const rb2 = realBodyLen(b2.o, b2.c);
    const rb3 = realBodyLen(b3.o, b3.c);

    // Crisp: each opens above the prior open (ordering).
    if (!(b2.o > b1.o and b3.o > b2.o)) return 0.0;

    // Fuzzy: very short upper shadows (all three).
    const mu_us1 = self.muLessCs(upperShadow(b1.o, b1.h, b1.c), &self.very_short_shadow, 3);
    const mu_us2 = self.muLessCs(upperShadow(b2.o, b2.h, b2.c), &self.very_short_shadow, 2);
    const mu_us3 = self.muLessCs(upperShadow(b3.o, b3.h, b3.c), &self.very_short_shadow, 1);

    // Fuzzy: each opens within or near the prior body (upper bound).
    const near3 = self.avgCS(&self.near, 3);
    const near3_width = if (near3 > 0.0) self.fuzz_ratio * near3 else 0.0;
    const mu_o2_near = self.muLtRaw(b2.o, b1.c + near3, near3_width);

    const near2 = self.avgCS(&self.near, 2);
    const near2_width = if (near2 > 0.0) self.fuzz_ratio * near2 else 0.0;
    const mu_o3_near = self.muLtRaw(b3.o, b2.c + near2, near2_width);

    // Fuzzy: not far shorter than prior candle.
    const far3 = self.avgCS(&self.far, 3);
    const far3_width = if (far3 > 0.0) self.fuzz_ratio * far3 else 0.0;
    const mu_not_far2 = self.muGtRaw(rb2, rb1 - far3, far3_width);

    const far2 = self.avgCS(&self.far, 2);
    const far2_width = if (far2 > 0.0) self.fuzz_ratio * far2 else 0.0;
    const mu_not_far3 = self.muGtRaw(rb3, rb2 - far2, far2_width);

    // Fuzzy: third candle is not short.
    const mu_not_short3 = self.muGreaterCs(rb3, &self.short_body, 1);

    const confidence = operators.tProductAll(&.{ mu_us1, mu_us2, mu_us3, mu_o2_near, mu_o3_near, mu_not_far2, mu_not_far3, mu_not_short3, });

    return confidence * 100.0;
}
