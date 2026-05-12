/// Morning Doji Star: a three-candle bullish reversal pattern.
///
/// Must have:
/// - first candle: long black real body,
/// - second candle: doji that gaps down (real body gap down from the first),
/// - third candle: white real body that closes well within the first candle's
///   real body.
///
/// The meaning of "long" is specified with self._long_body.
/// The meaning of "doji" is specified with self._doji_body.
/// The meaning of "short" is specified with self._short_body.
///
/// Category A: always bullish (continuous).
///
/// Returns:
///     Continuous float in [0, +100].  Always bullish.

const cp = @import("../candlestick_patterns.zig");
const operators = @import("fuzzy").operators;

const CandlestickPatterns = cp.CandlestickPatterns;
const CriterionState = cp.CriterionState;
const isBlack = cp.isBlack;
const isRealBodyGapDown = cp.isRealBodyGapDown;
const isWhite = cp.isWhite;
const realBodyLen = cp.realBodyLen;

const morning_doji_star_penetration_factor = 0.3;

pub fn morningDojiStar(self: *const CandlestickPatterns) f64 {
            if (!self.enough(3, &[_]*const CriterionState{&self.long_body, &self.doji_body, &self.short_body})) return 0.0;

    const b1 = self.bar(3);
    const b2 = self.bar(2);
    const b3 = self.bar(1);

    // Crisp gates: color checks and gap.
    if (!(isBlack(b1.o, b1.c) and isRealBodyGapDown(b1.o, b1.c, b2.o, b2.c) and isWhite(b3.o, b3.c))) return 0.0;

    // Fuzzy conditions.
    // c3 > c1 + rb1 * penetration
    const mu_long1 = self.muGreaterCs(realBodyLen(b1.o, b1.c), &self.long_body, 3);
    const mu_doji2 = self.muLessCs(realBodyLen(b2.o, b2.c), &self.doji_body, 2);

    const rb1 = realBodyLen(b1.o, b1.c);
    const threshold = b1.c + rb1 * morning_doji_star_penetration_factor;
    const width = self.fuzz_ratio * rb1 * morning_doji_star_penetration_factor;
    const mu_penetration = self.muGtRaw(b3.c, threshold, width);

    const confidence = operators.tProductAll(&.{mu_long1, mu_doji2, mu_penetration});
    return confidence * 100.0;
}
