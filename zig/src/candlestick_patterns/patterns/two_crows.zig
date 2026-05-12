/// Two Crows: a three-candle bearish pattern.
///
/// Must have:
/// - first candle: long white,
/// - second candle: black, gaps up (real body gap up from the first),
/// - third candle: black, opens within the second candle's real body,
///   closes within the first candle's real body.
///
/// The meaning of "long" is specified with self._long_body.
///
/// Category A: always bearish (continuous).
///
/// Returns:
///     Continuous float in [-100, 0].  Always bearish.

const cp = @import("../candlestick_patterns.zig");

const CandlestickPatterns = cp.CandlestickPatterns;
const CriterionState = cp.CriterionState;
const isBlack = cp.isBlack;
const isRealBodyGapUp = cp.isRealBodyGapUp;
const isWhite = cp.isWhite;
const realBodyLen = cp.realBodyLen;

pub fn twoCrows(self: *const CandlestickPatterns) f64 {
            if (!self.enough(3, &[_]*const CriterionState{&self.long_body})) return 0.0;

    const b1 = self.bar(3);
    const b2 = self.bar(2);
    const b3 = self.bar(1);

    // Crisp gates: colors.
    // Crisp gates: color checks.
    if (!(isWhite(b1.o, b1.c) and isBlack(b2.o, b2.c) and isBlack(b3.o, b3.c))) return 0.0;

    // Crisp: gap up.
    if (!isRealBodyGapUp(b1.o, b1.c, b2.o, b2.c)) return 0.0;

    // Crisp: third opens within second body (o3 < o2 and o3 > c2).
    if (!(b3.o < b2.o and b3.o > b2.c)) return 0.0;

    // Crisp: third closes within first body (c3 > o1 and c3 < c1).
    if (!(b3.c > b1.o and b3.c < b1.c)) return 0.0;

    // Fuzzy: first candle is long.
    const mu_long1 = self.muGreaterCs(realBodyLen(b1.o, b1.c), &self.long_body, 3);

    const confidence = mu_long1;

    return -confidence * 100.0;
}
