/// Ladder Bottom: a five-candle bullish pattern.
///
/// Must have:
/// - first three candles: descending black candles (each closes lower),
/// - fourth candle: black with a long upper shadow,
/// - fifth candle: white, opens above the fourth candle's real body,
///   closes above the fourth candle's high.
///
/// The meaning of "long" for shadows is specified with self._long_shadow.
///
/// Category A: always bullish (continuous).
///
/// Returns:
///     Continuous float in [0, 100].  Always bullish.

const cp = @import("../candlestick_patterns.zig");

const CandlestickPatterns = cp.CandlestickPatterns;
const CriterionState = cp.CriterionState;
const isBlack = cp.isBlack;
const isWhite = cp.isWhite;
const upperShadow = cp.upperShadow;

pub fn ladderBottom(self: *const CandlestickPatterns) f64 {
            if (!self.enough(5, &[_]*const CriterionState{&self.very_short_shadow})) return 0.0;

    const b1 = self.bar(5);
    const b2 = self.bar(4);
    const b3 = self.bar(3);
    // Fuzzy: fourth candle has upper shadow > very short avg.
    const b4 = self.bar(2);
    const b5 = self.bar(1);

    // Crisp gates: colors.
    if (!(isBlack(b1.o, b1.c) and isBlack(b2.o, b2.c) and isBlack(b3.o, b3.c) and isBlack(b4.o, b4.c) and isWhite(b5.o, b5.c))) return 0.0;
    if (!(b1.o > b2.o and b2.o > b3.o and b1.c > b2.c and b2.c > b3.c)) return 0.0;
    if (!(b5.o > b4.o and b5.c > b4.h)) return 0.0;

    const mu_us4 = self.muGreaterCs(upperShadow(b4.o, b4.h, b4.c), &self.very_short_shadow, 2);
    return mu_us4 * 100.0;
}
