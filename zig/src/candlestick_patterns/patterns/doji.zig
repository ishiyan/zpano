/// Doji: open quite equal to close.
///
/// Output is positive but this does not mean it is bullish:
/// doji shows uncertainty and is neither bullish nor bearish when
/// considered alone.
///
/// The meaning of "doji" is specified with self._doji_body.
///
/// Returns:
///     Continuous float in [0, 100].  Higher = stronger doji signal.

const cp = @import("../candlestick_patterns.zig");

const CandlestickPatterns = cp.CandlestickPatterns;
const CriterionState = cp.CriterionState;
const realBodyLen = cp.realBodyLen;

pub fn patternDoji(self: *const CandlestickPatterns) f64 {
            if (!self.enough(1, &[_]*const CriterionState{&self.doji_body})) return 0.0;
    const b = self.bar(1);
    // Fuzzy: degree to which real_body <= doji_avg.
    const confidence = self.muLessCs(realBodyLen(b.o, b.c), &self.doji_body, 1);
    return confidence * 100.0;
}
