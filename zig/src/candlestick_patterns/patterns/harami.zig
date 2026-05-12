/// Harami: a two-candle reversal pattern.
///
/// Must have:
/// - first candle: long real body,
/// - second candle: short real body contained within the first candle's
///   real body.
///
// Direction: opposite of 1st candle color.
/// Category B: direction from 1st candle color (continuous).
/// Containment degree is fuzzy.
///
/// Returns:
///     Continuous float in [-100, +100].

const cp = @import("../candlestick_patterns.zig");
const operators = @import("fuzzy").operators;

const CandlestickPatterns = cp.CandlestickPatterns;
const CriterionState = cp.CriterionState;
const realBodyLen = cp.realBodyLen;

pub fn patternHarami(self: *const CandlestickPatterns) f64 {
            if (!self.enough(2, &[_]*const CriterionState{&self.long_body, &self.short_body})) return 0.0;

    const b1 = self.bar(2);
    const b2 = self.bar(1);

    // Fuzzy size conditions.
    const mu_long1 = self.muGreaterCs(realBodyLen(b1.o, b1.c), &self.long_body, 2);
    const mu_short2 = self.muLessCs(realBodyLen(b2.o, b2.c), &self.short_body, 1);

    // Fuzzy containment: 1st body encloses 2nd body.
    const eq_avg = self.avgCS(&self.equal, 1);
    const eq_width = if (eq_avg > 0.0) self.fuzz_ratio * eq_avg else 0.0;

    const mu_enc_upper = self.muGeRaw(@max(b1.o, b1.c), @max(b2.o, b2.c), eq_width);
    const mu_enc_lower = self.muLtRaw(@min(b1.o, b1.c), @min(b2.o, b2.c), eq_width);

    const confidence = operators.tProductAll(&.{mu_long1, mu_short2, mu_enc_upper, mu_enc_lower});
    const direction: f64 = if (b1.c < b1.o) 1.0 else -1.0;
    return direction * confidence * 100.0;
}
