/// Belt Hold: a one-candle pattern.
///
/// A long candle with a very short shadow on the opening side:
/// - bullish: long white candle with very short lower shadow,
/// - bearish: long black candle with very short upper shadow.
///
/// The meaning of "long" is specified with self._long_body.
/// The meaning of "very short" for shadows is specified with
/// self._very_short_shadow.
///
/// Category C: both branches evaluated, return stronger signal.
///
/// Returns:
///     Continuous float in [-100, +100].

const cp = @import("../candlestick_patterns.zig");
const operators = @import("fuzzy").operators;

const CandlestickPatterns = cp.CandlestickPatterns;
const CriterionState = cp.CriterionState;
const isBlack = cp.isBlack;
const isWhite = cp.isWhite;
const lowerShadow = cp.lowerShadow;
const realBodyLen = cp.realBodyLen;
const upperShadow = cp.upperShadow;

pub fn beltHold(self: *const CandlestickPatterns) f64 {
            if (!self.enough(1, &[_]*const CriterionState{&self.long_body, &self.very_short_shadow})) return 0.0;

    const b = self.bar(1);
    const mu_long = self.muGreaterCs(realBodyLen(b.o, b.c), &self.long_body, 1);

    // Bullish: white + very short lower shadow.
    var bull_signal: f64 = 0.0;
    if (isWhite(b.o, b.c)) {
        const mu_vs = self.muLessCs(lowerShadow(b.o, b.l, b.c), &self.very_short_shadow, 1);
        const conf = operators.tProductAll(&.{mu_long, mu_vs});
        bull_signal = conf * 100.0;
    }

    // Bearish: black + very short upper shadow.
    var bear_signal: f64 = 0.0;
    if (isBlack(b.o, b.c)) {
        const mu_vs = self.muLessCs(upperShadow(b.o, b.h, b.c), &self.very_short_shadow, 1);
        const conf = operators.tProductAll(&.{mu_long, mu_vs});
        bear_signal = -conf * 100.0;
    }

    return if (@abs(bull_signal) >= @abs(bear_signal)) bull_signal else bear_signal;
}
