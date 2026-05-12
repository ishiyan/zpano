/// Stalled (Deliberation): a three-candle bearish pattern.
///
/// Three white candles with progressively higher closes:
/// - first candle: long white body,
/// - second candle: long white body, opens within or near the first
///   candle's body, very short upper shadow,
/// - third candle: small body that rides on the shoulder of the second
///   (opens near the second's close, accounting for its own body size).
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

pub fn stalled(self: *const CandlestickPatterns) f64 {
            if (!self.enough(3, &[_]*const CriterionState{&self.long_body, &self.short_body, &self.very_short_shadow, &self.near})) return 0.0;

    const b1 = self.bar(3);
    const b2 = self.bar(2);
    const b3 = self.bar(1);

    // Crisp gates: all white, rising closes.
    if (!(isWhite(b1.o, b1.c) and isWhite(b2.o, b2.c) and isWhite(b3.o, b3.c))) return 0.0;
    if (!(b3.c > b2.c and b2.c > b1.c)) return 0.0;
    // Crisp: o2 > o1 (opens above prior open).
    if (!(b2.o > b1.o)) return 0.0;

    const rb3 = realBodyLen(b3.o, b3.c);

    // Fuzzy conditions.
    const mu_long1 = self.muGreaterCs(realBodyLen(b1.o, b1.c), &self.long_body, 3);
    const mu_long2 = self.muGreaterCs(realBodyLen(b2.o, b2.c), &self.long_body, 2);
    const mu_us2 = self.muLessCs(upperShadow(b2.o, b2.h, b2.c), &self.very_short_shadow, 2);

    // o2 <= c1 + near_avg (opens within or near prior body).
    const near3 = self.avgCS(&self.near, 3);
    const near3_width = if (near3 > 0.0) self.fuzz_ratio * near3 else 0.0;
    const mu_o2_near = self.muLtRaw(b2.o, b1.c + near3, near3_width);

    // Third candle: short body.
    const mu_short3 = self.muLessCs(rb3, &self.short_body, 1);

    // o3 >= c2 - rb3 - near_avg (rides on shoulder).
    const near2 = self.avgCS(&self.near, 2);
    const near2_width = if (near2 > 0.0) self.fuzz_ratio * near2 else 0.0;
    const mu_o3_shoulder = self.muGeRaw(b3.o, b2.c - rb3 - near2, near2_width);

    const confidence = operators.tProductAll(&.{mu_long1, mu_long2, mu_us2, mu_o2_near, mu_short3, mu_o3_shoulder});

    return -1.0 * confidence * 100.0;
}
