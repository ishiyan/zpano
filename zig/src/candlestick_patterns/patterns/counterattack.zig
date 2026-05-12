/// Counterattack: a two-candle reversal pattern.
///
/// Two long candles of opposite color with closes that are equal
/// (or very near equal).
///
/// - bullish: first candle is long black, second is long white,
///   closes are equal,
/// - bearish: first candle is long white, second is long black,
///   closes are equal.
///
/// The meaning of "long" is specified with self._long_body.
/// The meaning of "equal" is specified with self._equal.
///
// Direction from 2nd candle color.
/// Category B: direction from 2nd candle color (continuous).
///
/// Returns:
///     Continuous float in [-100, +100].

const cp = @import("../candlestick_patterns.zig");
const operators = @import("fuzzy").operators;

const CandlestickPatterns = cp.CandlestickPatterns;
const CriterionState = cp.CriterionState;
const realBodyLen = cp.realBodyLen;

pub fn counterattack(self: *const CandlestickPatterns) f64 {
            if (!self.enough(2, &[_]*const CriterionState{&self.long_body, &self.equal})) return 0.0;

    const b1 = self.bar(2);
    const b2 = self.bar(1);

    // Opposite colors — crisp gate.
    const color1: i32 = if (b1.c < b1.o) -1 else 1;
    const color2: i32 = if (b2.c < b2.o) -1 else 1;
    if (color1 == color2) return 0.0;

    // Fuzzy conditions.
    const mu_long1 = self.muGreaterCs(realBodyLen(b1.o, b1.c), &self.long_body, 2);
    const mu_long2 = self.muGreaterCs(realBodyLen(b2.o, b2.c), &self.long_body, 1);
    // Closes near equal: crisp was abs(c2-c1) <= eq.
    // Model as mu_less(abs_diff, eq_avg) — crossover at eq boundary.
    const mu_eq = self.muLessCs(@abs(b2.c - b1.c), &self.equal, 2);

    const confidence = operators.tProductAll(&.{mu_long1, mu_long2, mu_eq});
    const direction: f64 = if (b2.c < b2.o) -1.0 else 1.0;
    return direction * confidence * 100.0;
}
