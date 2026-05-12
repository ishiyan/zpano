/// On Neck: a two-candle bearish continuation pattern.
///
/// Must have:
/// - first candle: long black,
/// - second candle: white that opens below the prior low and closes
///   equal to the prior candle's low.
///
/// The meaning of "long" is specified with self._long_body.
/// The meaning of "equal" is specified with self._equal.
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
const realBodyLen = cp.realBodyLen;

pub fn onNeck(self: *const CandlestickPatterns) f64 {
            if (!self.enough(2, &[_]*const CriterionState{&self.long_body, &self.equal})) return 0.0;

    const b1 = self.bar(2);
    const b2 = self.bar(1);

    // Crisp gates: color checks and open below prior low.
    if (!(isBlack(b1.o, b1.c) and isWhite(b2.o, b2.c) and b2.o < b1.l)) return 0.0;

    // Fuzzy conditions.
    const mu_long1 = self.muGreaterCs(realBodyLen(b1.o, b1.c), &self.long_body, 2);

    // Close equal to prior low: crisp was abs(c2-l1) <= eq.
    // Model as mu_less(abs_diff, eq_avg) -- crossover at eq boundary.
    // Model as mu_less(abs_diff, eq_avg) — crossover at eq boundary.
    const mu_near_low = self.muLessCs(@abs(b2.c - b1.l), &self.equal, 2);

    const confidence = operators.tProductAll(&.{mu_long1, mu_near_low});

    return -1.0 * confidence * 100.0;
}
