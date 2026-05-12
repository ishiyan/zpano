/// Concealing Baby Swallow: a four-candle bullish pattern.
///
/// Must have:
/// - first candle: black marubozu (very short shadows),
/// - second candle: black marubozu (very short shadows),
/// - third candle: black, opens gapping down, upper shadow extends into
///   the prior real body (upper shadow > very-short avg),
/// - fourth candle: black, completely engulfs the third candle including
///   shadows (strict > / <).
///
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
const isRealBodyGapDown = cp.isRealBodyGapDown;
const lowerShadow = cp.lowerShadow;
const upperShadow = cp.upperShadow;

pub fn concealingBabySwallow(self: *const CandlestickPatterns) f64 {
            if (!self.enough(4, &[_]*const CriterionState{&self.very_short_shadow})) return 0.0;

    const b1 = self.bar(4);
    const b2 = self.bar(3);
    // Fuzzy: third candle upper shadow > very-short avg.
    const b3 = self.bar(2);
    const b4 = self.bar(1);

    // Crisp gates: all black.
    if (!(isBlack(b1.o, b1.c) and isBlack(b2.o, b2.c) and isBlack(b3.o, b3.c) and isBlack(b4.o, b4.c))) return 0.0;
    // Crisp: gap down and upper shadow extends into prior body.
    if (!(isRealBodyGapDown(b2.o, b2.c, b3.o, b3.c) and b3.h > b2.c)) return 0.0;
    // Crisp: fourth engulfs third including shadows (strict).
    if (!(b4.h > b3.h and b4.l < b3.l)) return 0.0;

    // Fuzzy: first and second are marubozu (very short shadows).
    const mu_ls1 = self.muLessCs(lowerShadow(b1.o, b1.l, b1.c), &self.very_short_shadow, 4);
    const mu_us1 = self.muLessCs(upperShadow(b1.o, b1.h, b1.c), &self.very_short_shadow, 4);
    const mu_ls2 = self.muLessCs(lowerShadow(b2.o, b2.l, b2.c), &self.very_short_shadow, 3);
    const mu_us2 = self.muLessCs(upperShadow(b2.o, b2.h, b2.c), &self.very_short_shadow, 3);
    const mu_us3_long = self.muGreaterCs(upperShadow(b3.o, b3.h, b3.c), &self.very_short_shadow, 2);

    const confidence = operators.tProductAll(&.{mu_ls1, mu_us1, mu_ls2, mu_us2, mu_us3_long});
    return confidence * 100.0;
}
