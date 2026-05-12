/// Stick Sandwich: a three-candle bullish pattern.
///
/// Must have:
/// - first candle: black,
/// - second candle: white, trades above the first candle's close
///   (low > first close),
/// - third candle: black, close equals the first candle's close.
///
/// The meaning of "equal" is specified with self._equal.
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

pub fn stickSandwich(self: *const CandlestickPatterns) f64 {
            if (!self.enough(3, &[_]*const CriterionState{&self.equal})) return 0.0;

    const b1 = self.bar(3);
    const b2 = self.bar(2);
    const b3 = self.bar(1);

    // Crisp gates: colors and gap.
    if (!(isBlack(b1.o, b1.c) and isWhite(b2.o, b2.c) and isBlack(b3.o, b3.c) and b2.l > b1.c)) return 0.0;

    // Fuzzy: third close equals first close (two-sided band).
    const mu_eq = self.muLessCs(@abs(b3.c - b1.c), &self.equal, 3);

    const confidence = mu_eq;

    return confidence * 100.0;
}
