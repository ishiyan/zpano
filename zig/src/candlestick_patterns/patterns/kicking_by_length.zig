// Direction determined by the longer marubozu's color.
/// Kicking By Length: like Kicking but direction determined by the longer marubozu.
///
/// Must have:
/// - first candle: marubozu (long body, very short shadows),
// Crisp: opposite colors.
/// - second candle: opposite-color marubozu with a high-low gap,
/// - bull/bear determined by which marubozu has the longer real body.
///
/// Category B: direction from longer marubozu's color.
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

pub fn patternKickingByLength(self: *const CandlestickPatterns) f64 {
            if (!self.enough(2, &[_]*const CriterionState{&self.very_short_shadow, &self.long_body})) return 0.0;

    const b1 = self.bar(2);
    const b2 = self.bar(1);

    const color1: i32 = if (b1.c < b1.o) -1 else 1;
    const color2: i32 = if (b2.c < b2.o) -1 else 1;
    if (color1 == color2) return 0.0;

    // Crisp: gap check.
    const has_gap = if (color1 == -1) isHighLowGapUp(b1.h, b2.l) else isHighLowGapDown(b1.l, b2.h);
    if (!has_gap) return 0.0;

    const rb1 = realBodyLen(b1.o, b1.c);
    const rb2 = realBodyLen(b2.o, b2.c);

    // Fuzzy: both are marubozu (long body, very short shadows).
    const mu_long1 = self.muGreaterCs(rb1, &self.long_body, 2);
    const mu_vs_us1 = self.muLessCs(upperShadow(b1.o, b1.h, b1.c), &self.very_short_shadow, 2);
    const mu_vs_ls1 = self.muLessCs(lowerShadow(b1.o, b1.l, b1.c), &self.very_short_shadow, 2);
    const mu_long2 = self.muGreaterCs(rb2, &self.long_body, 1);
    const mu_vs_us2 = self.muLessCs(upperShadow(b2.o, b2.h, b2.c), &self.very_short_shadow, 1);
    const mu_vs_ls2 = self.muLessCs(lowerShadow(b2.o, b2.l, b2.c), &self.very_short_shadow, 1);

    const confidence = operators.tProductAll(&.{mu_long1, mu_vs_us1, mu_vs_ls1, mu_long2, mu_vs_us2, mu_vs_ls2});

    const direction = if (rb2 > rb1) color2 else color1;
    return @as(f64, @floatFromInt(direction)) * confidence * 100.0;
}
