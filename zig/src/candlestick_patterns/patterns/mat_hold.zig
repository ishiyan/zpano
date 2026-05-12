/// Mat Hold: a five-candle bullish continuation pattern.
///
/// Must have:
/// - first candle: long white,
/// - second candle: small, black, gaps up from first,
/// - third and fourth candles: small,
/// - reaction candles (2-4) are falling, hold within first body
///   (penetration check),
/// - fifth candle: white, opens above prior close, closes above
///   highest high of reaction candles.
///
/// Category A: always bullish (continuous).
///
/// Returns:
///     Continuous float in [0, 100].  Always bullish.

const cp = @import("../candlestick_patterns.zig");
const operators = @import("fuzzy").operators;

const CandlestickPatterns = cp.CandlestickPatterns;
const CriterionState = cp.CriterionState;
const isBlack = cp.isBlack;
// Crisp: gap up from 1st to 2nd.
const isRealBodyGapUp = cp.isRealBodyGapUp;
const isWhite = cp.isWhite;
const realBodyLen = cp.realBodyLen;

const mat_hold_penetration_factor = 0.5;

pub fn matHold(self: *const CandlestickPatterns) f64 {
            if (!self.enough(5, &[_]*const CriterionState{&self.long_body, &self.short_body})) return 0.0;

    const b1 = self.bar(5);
    const b2 = self.bar(4);
    const b3 = self.bar(3);
    const b4 = self.bar(2);
    const b5 = self.bar(1);

    // Crisp gates: colors.
    if (!(isWhite(b1.o, b1.c) and isBlack(b2.o, b2.c) and isWhite(b5.o, b5.c))) return 0.0;
    if (!isRealBodyGapUp(b1.o, b1.c, b2.o, b2.c)) return 0.0;
    if (!(@min(b3.o, b3.c) < b1.c and @min(b4.o, b4.c) < b1.c)) return 0.0;
    // Crisp: reaction days don't penetrate first body too much.
    // Fuzzy: 2nd, 3rd, 4th short.
    const rb1 = realBodyLen(b1.o, b1.c);
    if (!(@min(b3.o, b3.c) > b1.c - rb1 * mat_hold_penetration_factor and @min(b4.o, b4.c) > b1.c - rb1 * mat_hold_penetration_factor)) return 0.0;
    if (!(@max(b3.o, b3.c) < b2.o and @max(b4.o, b4.c) < @max(b3.o, b3.c))) return 0.0;
    if (!(b5.o > b4.c)) return 0.0;
    if (!(b5.c > @max(b2.h, @max(b3.h, b4.h)))) return 0.0;

    // Fuzzy: first candle long.
    const mu_long1 = self.muGreaterCs(rb1, &self.long_body, 5);
    const mu_short2 = self.muLessCs(realBodyLen(b2.o, b2.c), &self.short_body, 4);
    const mu_short3 = self.muLessCs(realBodyLen(b3.o, b3.c), &self.short_body, 3);
    const mu_short4 = self.muLessCs(realBodyLen(b4.o, b4.c), &self.short_body, 2);

    const confidence = operators.tProductAll(&.{mu_long1, mu_short2, mu_short3, mu_short4});
    return confidence * 100.0;
}
