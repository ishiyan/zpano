/// Identical Three Crows: a three-candle bearish pattern.
///
/// Must have:
/// - three consecutive declining black candles,
/// - each opens very close to the prior candle's close (equal criterion),
/// - very short lower shadows.
///
/// The meaning of "equal" is specified with self._equal.
/// The meaning of "very short" for shadows is specified with
/// self._very_short_shadow.
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
const lowerShadow = cp.lowerShadow;

pub fn identicalThreeCrows(self: *const CandlestickPatterns) f64 {
            if (!self.enough(3, &[_]*const CriterionState{&self.equal, &self.very_short_shadow})) return 0.0;

    const b1 = self.bar(3);
    const b2 = self.bar(2);
    const b3 = self.bar(1);

    // Crisp gates: all black, declining closes.
    if (!(isBlack(b1.o, b1.c) and isBlack(b2.o, b2.c) and isBlack(b3.o, b3.c))) return 0.0;
    if (!(b1.c > b2.c and b2.c > b3.c)) return 0.0;

    // Fuzzy conditions.
    const mu_ls1 = self.muLessCs(lowerShadow(b1.o, b1.l, b1.c), &self.very_short_shadow, 3);
    const mu_ls2 = self.muLessCs(lowerShadow(b2.o, b2.l, b2.c), &self.very_short_shadow, 2);
    const mu_ls3 = self.muLessCs(lowerShadow(b3.o, b3.l, b3.c), &self.very_short_shadow, 1);
    // Opens near prior close (equal criterion, two-sided band).
    const mu_eq2 = self.muLessCs(@abs(b2.o - b1.c), &self.equal, 3);
    const mu_eq3 = self.muLessCs(@abs(b3.o - b2.c), &self.equal, 2);

    const confidence = operators.tProductAll(&.{mu_ls1, mu_ls2, mu_ls3, mu_eq2, mu_eq3});
    return -confidence * 100.0;
}
