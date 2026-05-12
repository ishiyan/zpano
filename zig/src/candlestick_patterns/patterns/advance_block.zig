/// Advance Block: a bearish three-candle pattern.
///
/// Three white candles with consecutively higher closes and opens, but
/// showing signs of weakening (diminishing bodies, growing upper shadows).
///
/// Category A: always bearish (continuous).
///
/// Returns:
///     Continuous float in [-100, 0].  Always bearish.

const cp = @import("../candlestick_patterns.zig");
const operators = @import("fuzzy").operators;

const CandlestickPatterns = cp.CandlestickPatterns;
const CriterionState = cp.CriterionState;
const isWhite = cp.isWhite;
const realBodyLen = cp.realBodyLen;
const upperShadow = cp.upperShadow;

pub fn advanceBlock(self: *const CandlestickPatterns) f64 {
            if (!self.enough(3, &[_]*const CriterionState{&self.long_body, &self.short_shadow, &self.long_shadow, &self.near, &self.far})) return 0.0;

    // Fuzzy: first candle short upper shadow.
    const b1 = self.bar(3);
    const b2 = self.bar(2);
    const b3 = self.bar(1);

    // Crisp gates: all white with rising closes.
    if (!(isWhite(b1.o, b1.c) and isWhite(b2.o, b2.c) and isWhite(b3.o, b3.c) and b3.c > b2.c and b2.c > b1.c)) return 0.0;
    if (!(b2.o > b1.o)) return 0.0;
    if (!(b3.o > b2.o)) return 0.0;

    const rb1 = realBodyLen(b1.o, b1.c);
    const rb2 = realBodyLen(b2.o, b2.c);
    const rb3 = realBodyLen(b3.o, b3.c);

    // Fuzzy: 2nd opens within/near 1st body (upper bound).
    const near3 = self.avgCS(&self.near, 3);
    const near3_width = if (near3 > 0.0) self.fuzz_ratio * near3 else 0.0;
    const mu_o2_near = self.muLtRaw(b2.o, b1.c + near3, near3_width);

    // Fuzzy: 3rd opens within/near 2nd body (upper bound).
    const near2 = self.avgCS(&self.near, 2);
    const near2_width = if (near2 > 0.0) self.fuzz_ratio * near2 else 0.0;
    const mu_o3_near = self.muLtRaw(b3.o, b2.c + near2, near2_width);

    // Fuzzy: first candle long body.
    const mu_long1 = self.muGreaterCs(rb1, &self.long_body, 3);
    const mu_us1 = self.muLessCs(upperShadow(b1.o, b1.h, b1.c), &self.short_shadow, 3);

    // At least one weakness condition must hold (OR -> max).
    // At least one weakness condition must hold (OR → max).
    const far2 = self.avgCS(&self.far, 3);
    const far2_width = if (far2 > 0.0) self.fuzz_ratio * far2 else 0.0;
    const far1 = self.avgCS(&self.far, 2);
    const far1_width = if (far1 > 0.0) self.fuzz_ratio * far1 else 0.0;
    const near1 = self.avgCS(&self.near, 2);
    const near1_width = if (near1 > 0.0) self.fuzz_ratio * near1 else 0.0;

    // Branch 1: 2 far smaller than 1 AND 3 not longer than 2
    const mu_b1a = self.muLtRaw(rb2, rb1 - far2, far2_width);
    const mu_b1b = self.muLtRaw(rb3, rb2 + near1, near1_width);
    const branch1 = operators.tProductAll(&.{mu_b1a, mu_b1b});

    // Branch 2: 3 far smaller than 2
    const branch2 = self.muLtRaw(rb3, rb2 - far1, far1_width);

    // Branch 3: 3 < 2 AND 2 < 1 AND (3 or 2 has non-short upper shadow)
    const rb3_width = if (rb2 > 0.0) self.fuzz_ratio * rb2 else 0.0;
    const rb2_width = if (rb1 > 0.0) self.fuzz_ratio * rb1 else 0.0;
    const mu_b3a = self.muLtRaw(rb3, rb2, rb3_width);
    const mu_b3b = self.muLtRaw(rb2, rb1, rb2_width);
    const mu_b3_us3 = self.muGreaterCs(upperShadow(b3.o, b3.h, b3.c), &self.short_shadow, 1);
    const mu_b3_us2 = self.muGreaterCs(upperShadow(b2.o, b2.h, b2.c), &self.short_shadow, 2);
    const branch3 = operators.tProductAll(&.{mu_b3a, mu_b3b, @max(mu_b3_us3, mu_b3_us2)});

    // Branch 4: 3 < 2 AND 3 has long upper shadow
    const mu_b4a = self.muLtRaw(rb3, rb2, rb3_width);
    const mu_b4b = self.muGreaterCs(upperShadow(b3.o, b3.h, b3.c), &self.long_shadow, 1);
    const branch4 = operators.tProductAll(&.{mu_b4a, mu_b4b});

    const weakness = @max(@max(branch1, branch2), @max(branch3, branch4));

    const confidence = operators.tProductAll(&.{mu_o2_near, mu_o3_near, mu_long1, mu_us1, weakness});

    return -confidence * 100.0;
}
