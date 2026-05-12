/// Thrusting: a two-candle bearish continuation pattern.
///
/// Must have:
/// - first candle: long black,
/// - second candle: white, opens below the prior candle's low, closes
///   into the prior candle's real body but below the midpoint, and the
///   close is not equal to the prior candle's close (to distinguish
///   from in-neck).
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

pub fn patternThrusting(self: *const CandlestickPatterns) f64 {
            if (!self.enough(2, &[_]*const CriterionState{&self.long_body, &self.equal})) return 0.0;

    const b1 = self.bar(2);
    const b2 = self.bar(1);

    const rb1 = realBodyLen(b1.o, b1.c);

    // Crisp gates: color checks and open below prior low.
    if (!(isBlack(b1.o, b1.c) and isWhite(b2.o, b2.c) and b2.o < b1.l)) return 0.0;

    // Fuzzy conditions.
    const mu_long1 = self.muGreaterCs(rb1, &self.long_body, 2);

    // Close above prior close + equal avg (not equal to prior close).
    const eq = self.avgCS(&self.equal, 2);
    const eq_width = if (eq > 0.0) self.fuzz_ratio * eq else 0.0;
    const mu_above_close = self.muGtRaw(b2.c, b1.c + eq, eq_width);

    // Close at or below midpoint of prior body: c2 <= c1 + rb1 * 0.5
    const mid = b1.c + rb1 * 0.5;
    const mid_width = if (rb1 > 0.0) self.fuzz_ratio * rb1 * 0.5 else 0.0;
    const mu_below_mid = self.muLtRaw(b2.c, mid, mid_width);

    const confidence = operators.tProductAll(&.{mu_long1, mu_above_close, mu_below_mid});

    return -1.0 * confidence * 100.0;
}
