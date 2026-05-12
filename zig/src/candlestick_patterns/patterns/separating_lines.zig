/// Separating Lines: a two-candle continuation pattern.
///
/// Opposite colors with the same open. The second candle is a belt hold
/// (long body with no shadow on the opening side).
///
/// - bullish: first candle is black, second is white with same open,
///   long body, very short lower shadow,
/// - bearish: first candle is white, second is black with same open,
///   long body, very short upper shadow.
///
/// The meaning of "long" is specified with self._long_body.
/// The meaning of "very short" for shadows is specified with
/// self._very_short_shadow.
/// The meaning of "equal" is specified with self._equal.
///
/// Category C: both branches evaluated, return stronger signal.
///
/// Returns:
///     Continuous float in [-100, +100].

const cp = @import("../candlestick_patterns.zig");
const operators = @import("fuzzy").operators;

const CandlestickPatterns = cp.CandlestickPatterns;
const CriterionState = cp.CriterionState;
const lowerShadow = cp.lowerShadow;
const realBodyLen = cp.realBodyLen;
const upperShadow = cp.upperShadow;

pub fn separatingLines(self: *const CandlestickPatterns) f64 {
            if (!self.enough(2, &[_]*const CriterionState{&self.long_body, &self.very_short_shadow, &self.equal})) return 0.0;

    const b1 = self.bar(2);
    const b2 = self.bar(1);

    // Opposite colors -- crisp gate.
    // Opposite colors — crisp gate.
    const color1: i32 = if (b1.c < b1.o) -1 else 1;
    const color2: i32 = if (b2.c < b2.o) -1 else 1;
    if (color1 == color2) return 0.0;

    // Opens near equal -- fuzzy (crisp was abs(o2-o1) <= eq).
    // Opens near equal — fuzzy (crisp was abs(o2-o1) <= eq).
    const mu_eq = self.muLessCs(@abs(b2.o - b1.o), &self.equal, 2);

    // Long body on 2nd candle -- fuzzy.
    // Long body on 2nd candle — fuzzy.
    const mu_long = self.muGreaterCs(realBodyLen(b2.o, b2.c), &self.long_body, 1);

    // Bullish: white belt hold (very short lower shadow).
    var bull_signal: f64 = 0.0;
    if (color2 == 1) {
        const mu_vs = self.muLessCs(lowerShadow(b2.o, b2.l, b2.c), &self.very_short_shadow, 1);
        const conf = operators.tProductAll(&.{mu_eq, mu_long, mu_vs});
        bull_signal = conf * 100.0;
    }

    // Bearish: black belt hold (very short upper shadow).
    var bear_signal: f64 = 0.0;
    if (color2 == -1) {
        const mu_vs = self.muLessCs(upperShadow(b2.o, b2.h, b2.c), &self.very_short_shadow, 1);
        const conf = operators.tProductAll(&.{mu_eq, mu_long, mu_vs});
        bear_signal = -conf * 100.0;
    }

    return if (@abs(bull_signal) >= @abs(bear_signal)) bull_signal else bear_signal;
}
