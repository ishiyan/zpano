// Crisp: opposite colors.
// Crisp: gap check.
/// Kicking: a two-candle pattern with opposite-color marubozus and gap.
///
/// Must have:
/// - first candle: marubozu (long body, very short shadows),
/// - second candle: opposite-color marubozu with a high-low gap,
/// - bullish: black marubozu followed by white marubozu gapping up,
/// - bearish: white marubozu followed by black marubozu gapping down.
///
/// Category B: direction from second candle's color.
///
/// Returns:
///     Continuous float in [-100, +100].

const cp = @import("../candlestick_patterns.zig");
const operators = @import("fuzzy").operators;

const CandlestickPatterns = cp.CandlestickPatterns;
const CriterionState = cp.CriterionState;
const isHighLowGapDown = cp.isHighLowGapDown;
const isHighLowGapUp = cp.isHighLowGapUp;
const lowerShadow = cp.lowerShadow;
const realBodyLen = cp.realBodyLen;
const upperShadow = cp.upperShadow;

pub fn patternKicking(self: *const CandlestickPatterns) f64 {
            if (!self.enough(2, &[_]*const CriterionState{&self.very_short_shadow, &self.long_body})) return 0.0;

    const b1 = self.bar(2);
    const b2 = self.bar(1);

    const color1: i32 = if (b1.c < b1.o) -1 else 1;
    const color2: i32 = if (b2.c < b2.o) -1 else 1;
    if (color1 == color2) return 0.0;

    if (color1 == -1 and !isHighLowGapUp(b1.h, b2.l)) return 0.0;
    if (color1 == 1 and !isHighLowGapDown(b1.l, b2.h)) return 0.0;

    // Fuzzy: both are marubozu (long body, very short shadows).
    const mu_long1 = self.muGreaterCs(realBodyLen(b1.o, b1.c), &self.long_body, 2);
    const mu_vs_us1 = self.muLessCs(upperShadow(b1.o, b1.h, b1.c), &self.very_short_shadow, 2);
    const mu_vs_ls1 = self.muLessCs(lowerShadow(b1.o, b1.l, b1.c), &self.very_short_shadow, 2);
    const mu_long2 = self.muGreaterCs(realBodyLen(b2.o, b2.c), &self.long_body, 1);
    const mu_vs_us2 = self.muLessCs(upperShadow(b2.o, b2.h, b2.c), &self.very_short_shadow, 1);
    const mu_vs_ls2 = self.muLessCs(lowerShadow(b2.o, b2.l, b2.c), &self.very_short_shadow, 1);

    const confidence = operators.tProductAll(&.{mu_long1, mu_vs_us1, mu_vs_ls1, mu_long2, mu_vs_us2, mu_vs_ls2});
    return @as(f64, @floatFromInt(color2)) * confidence * 100.0;
}
