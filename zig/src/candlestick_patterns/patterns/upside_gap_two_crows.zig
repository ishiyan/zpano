/// Upside Gap Two Crows: a three-candle bearish pattern.
///
/// Must have:
/// - first candle: long white,
/// - second candle: small black that gaps up from the first,
/// - third candle: black that engulfs the second candle's body and
///   closes above the first candle's close (gap not filled).
///
/// The meaning of "long" is specified with self._long_body.
/// The meaning of "short" is specified with self._short_body.
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
const isRealBodyGapUp = cp.isRealBodyGapUp;
const isWhite = cp.isWhite;
const realBodyLen = cp.realBodyLen;

pub fn upsideGapTwoCrows(self: *const CandlestickPatterns) f64 {
            if (!self.enough(3, &[_]*const CriterionState{&self.long_body, &self.short_body})) return 0.0;

    const b1 = self.bar(3);
    const b2 = self.bar(2);
    const b3 = self.bar(1);

    // Crisp gates: colors.
    if (!(isWhite(b1.o, b1.c) and isBlack(b2.o, b2.c) and isBlack(b3.o, b3.c))) return 0.0;

    // Crisp: gap up from first to second.
    if (!isRealBodyGapUp(b1.o, b1.c, b2.o, b2.c)) return 0.0;

    // Crisp: third engulfs second (o3 > o2 and c3 < c2) and closes above c1.
    if (!(b3.o > b2.o and b3.c < b2.c and b3.c > b1.c)) return 0.0;

    // Fuzzy: first candle is long.
    const mu_long1 = self.muGreaterCs(realBodyLen(b1.o, b1.c), &self.long_body, 3);

    // Fuzzy: second candle is short.
    const mu_short2 = self.muLessCs(realBodyLen(b2.o, b2.c), &self.short_body, 2);

    const confidence = operators.tProductAll(&.{mu_long1, mu_short2});

    return -confidence * 100.0;
}
