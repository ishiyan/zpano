/// Three Black Crows: a four-candle bearish reversal pattern.
///
/// Must have:
/// - preceding candle (oldest) is white,
/// - three consecutive black candles with declining closes,
/// - each opens within the prior black candle's real body,
/// - each has a very short lower shadow,
/// - 1st black closes under the prior white candle's high.
///
/// Category A: always bearish (continuous).
///
/// Returns:
///     Continuous float in [-100, 0].  Always bearish.

const cp = @import("../candlestick_patterns.zig");
const operators = @import("fuzzy").operators;

const CandlestickPatterns = cp.CandlestickPatterns;
const CriterionState = cp.CriterionState;
const isBlack = cp.isBlack;
const isWhite = cp.isWhite;
const lowerShadow = cp.lowerShadow;

pub fn threeBlackCrows(self: *const CandlestickPatterns) f64 {
            if (!self.enough(4, &[_]*const CriterionState{&self.very_short_shadow})) return 0.0;

    const b0 = self.bar(4); // prior white;
    const b1 = self.bar(3); // 1st black;
    const b2 = self.bar(2); // 2nd black;
    const b3 = self.bar(1); // 3rd black;

    // Crisp gates: colors, declining closes, opens within prior body.
    if (!isWhite(b0.o, b0.c)) return 0.0;
    if (!(isBlack(b1.o, b1.c) and isBlack(b2.o, b2.c) and isBlack(b3.o, b3.c))) return 0.0;
    if (!(b1.c > b2.c and b2.c > b3.c)) return 0.0;
    // Opens within prior black body (crisp containment for strict ordering).
    if (!(b2.o < b1.o and b2.o > b1.c and b3.o < b2.o and b3.o > b2.c)) return 0.0;
    // Prior white's high > 1st black's close (crisp).
    if (!(b0.h > b1.c)) return 0.0;

    // Fuzzy: very short lower shadows.
    const mu_ls1 = self.muLessCs(lowerShadow(b1.o, b1.l, b1.c), &self.very_short_shadow, 3);
    const mu_ls2 = self.muLessCs(lowerShadow(b2.o, b2.l, b2.c), &self.very_short_shadow, 2);
    const mu_ls3 = self.muLessCs(lowerShadow(b3.o, b3.l, b3.c), &self.very_short_shadow, 1);

    const confidence = operators.tProductAll(&.{mu_ls1, mu_ls2, mu_ls3});

    return -1.0 * confidence * 100.0;
}
