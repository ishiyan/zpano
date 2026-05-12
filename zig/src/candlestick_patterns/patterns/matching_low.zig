/// Matching Low: a two-candle bullish pattern.
///
/// Must have:
/// - first candle: black,
/// - second candle: black with close equal to the first candle's close.
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

pub fn matchingLow(self: *const CandlestickPatterns) f64 {
            if (!self.enough(2, &[_]*const CriterionState{&self.equal})) return 0.0;

    const b1 = self.bar(2);
    const b2 = self.bar(1);

    // Crisp gates: both black.
    if (!(isBlack(b1.o, b1.c) and isBlack(b2.o, b2.c))) return 0.0;

    // Fuzzy: close equal to prior close (two-sided band).
    const mu_eq = self.muLessCs(@abs(b2.c - b1.c), &self.equal, 2);
    return mu_eq * 100.0;
}
