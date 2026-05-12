/// Tristar: a three-candle reversal pattern with three dojis.
///
/// Must have:
/// - three consecutive doji candles,
/// - if the second doji gaps up from the first and the third does not
///   close higher than the second: bearish,
/// - if the second doji gaps down from the first and the third does not
///   close lower than the second: bullish.
///
/// Category A: fixed direction per branch (bullish or bearish).
///
/// Returns:
///     Continuous float in [-100, +100].

const cp = @import("../candlestick_patterns.zig");
const operators = @import("fuzzy").operators;

const CandlestickPatterns = cp.CandlestickPatterns;
const CriterionState = cp.CriterionState;
const isRealBodyGapDown = cp.isRealBodyGapDown;
// Bearish: second gaps up, third is not higher than second — crisp direction checks.
const isRealBodyGapUp = cp.isRealBodyGapUp;
const realBodyLen = cp.realBodyLen;

pub fn patternTristar(self: *const CandlestickPatterns) f64 {
            if (!self.enough(3, &[_]*const CriterionState{&self.doji_body})) return 0.0;

    const b1 = self.bar(3);
    const b2 = self.bar(2);
    const b3 = self.bar(1);

    // Fuzzy: all three must be dojis.
    const mu_doji1 = self.muLessCs(realBodyLen(b1.o, b1.c), &self.doji_body, 3);
    const mu_doji2 = self.muLessCs(realBodyLen(b2.o, b2.c), &self.doji_body, 2);
    const mu_doji3 = self.muLessCs(realBodyLen(b3.o, b3.c), &self.doji_body, 1);

    // Bearish: second gaps up, third is not higher than second -- crisp direction checks.
    if (isRealBodyGapUp(b1.o, b1.c, b2.o, b2.c) and @max(b3.o, b3.c) < @max(b2.o, b2.c))
    {
        const conf = operators.tProductAll(&.{mu_doji1, mu_doji2, mu_doji3});
        return -conf * 100.0;
    }

    // Bullish: second gaps down, third is not lower than second.
    if (isRealBodyGapDown(b1.o, b1.c, b2.o, b2.c) and @min(b3.o, b3.c) > @min(b2.o, b2.c))
    {
        const conf = operators.tProductAll(&.{mu_doji1, mu_doji2, mu_doji3});
        return conf * 100.0;
    }

    return 0.0;
}
