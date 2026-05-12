/// Rising/Falling Three Methods: a five-candle continuation pattern.
///
// 2nd to 4th are falling (rising) — using color multiply trick — crisp.
// 5th opens above (below) the prior close — crisp.
// 5th closes above (below) the 1st close — crisp.
/// Uses TA-Lib logic: opposite-color check via color multiplication,
/// real-body overlap (not full candle containment), sequential closes,
/// 5th opens beyond 4th close.
///
/// Category B: direction from 1st candle color (crisp sign).
///
/// Returns:
///     Continuous float in [-100, +100].

const cp = @import("../candlestick_patterns.zig");
const operators = @import("fuzzy").operators;

const CandlestickPatterns = cp.CandlestickPatterns;
const CriterionState = cp.CriterionState;
const isWhite = cp.isWhite;
const realBodyLen = cp.realBodyLen;

pub fn risingFallingThreeMethods(self: *const CandlestickPatterns) f64 {
            if (!self.enough(5, &[_]*const CriterionState{&self.long_body, &self.short_body})) return 0.0;

    const b1 = self.bar(5);
    const b2 = self.bar(4);
    const b3 = self.bar(3);
    const b4 = self.bar(2);
    const b5 = self.bar(1);

    // Fuzzy: 1st long, 2nd-4th short, 5th long.
    const mu_long1 = self.muGreaterCs(realBodyLen(b1.o, b1.c), &self.long_body, 5);
    const mu_short2 = self.muLessCs(realBodyLen(b2.o, b2.c), &self.short_body, 4);
    const mu_short3 = self.muLessCs(realBodyLen(b3.o, b3.c), &self.short_body, 3);
    const mu_short4 = self.muLessCs(realBodyLen(b4.o, b4.c), &self.short_body, 2);
    const mu_long5 = self.muGreaterCs(realBodyLen(b5.o, b5.c), &self.long_body, 1);

    // Determine color of 1st candle: +1 white, -1 black -- crisp sign.
    // Determine color of 1st candle: +1 white, -1 black — crisp sign.
    const color1: f64 = if (!isWhite(b1.o, b1.c)) -1.0 else 1.0;

    // Color check: white, 3 black, white OR black, 3 white, black -- crisp.
    // Color check: white, 3 black, white  OR  black, 3 white, black — crisp.
    const c2: f64 = if (!isWhite(b2.o, b2.c)) -1.0 else 1.0;
    const c3: f64 = if (!isWhite(b3.o, b3.c)) -1.0 else 1.0;
    const c4: f64 = if (!isWhite(b4.o, b4.c)) -1.0 else 1.0;
    const c5: f64 = if (!isWhite(b5.o, b5.c)) -1.0 else 1.0;

    if (!(c2 == -color1 and c3 == c2 and c4 == c3 and c5 == -c4)) return 0.0;

    // 2nd to 4th hold within 1st: a part of the real body overlaps 1st range -- crisp.
    if (!(@min(b2.o, b2.c) < b1.h and @max(b2.o, b2.c) > b1.l and @min(b3.o, b3.c) < b1.h and @max(b3.o, b3.c) > b1.l and @min(b4.o, b4.c) < b1.h and @max(b4.o, b4.c) > b1.l)) return 0.0;

    // 2nd to 4th are falling (rising) -- using color multiply trick -- crisp.
    if (!(b3.c * color1 < b2.c * color1 and b4.c * color1 < b3.c * color1)) return 0.0;

    // 5th opens above (below) the prior close -- crisp.
    if (!(b5.o * color1 > b4.c * color1)) return 0.0;

    // 5th closes above (below) the 1st close -- crisp.
    if (!(b5.c * color1 > b1.c * color1)) return 0.0;

    const conf = operators.tProductAll(&.{mu_long1, mu_short2, mu_short3, mu_short4, mu_long5});
    return color1 * conf * 100.0;
}
