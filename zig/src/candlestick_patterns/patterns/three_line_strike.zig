/// Three-Line Strike: a four-candle pattern.
///
/// Bullish: three white candles with rising closes, each opening within/near
/// the prior body, 4th black opens above 3rd close and closes below 1st open.
///
/// Bearish: three black candles with falling closes, each opening within/near
/// the prior body, 4th white opens below 3rd close and closes above 1st open.
///
/// Category C: both branches evaluated, return stronger signal.
///
/// Returns:
///     Continuous float in [-100, +100].

const cp = @import("../candlestick_patterns.zig");
const operators = @import("fuzzy").operators;

const CandlestickPatterns = cp.CandlestickPatterns;
const CriterionState = cp.CriterionState;
const isWhite = cp.isWhite;

pub fn threeLineStrike(self: *const CandlestickPatterns) f64 {
            if (!self.enough(4, &[_]*const CriterionState{&self.near})) return 0.0;

    const b1 = self.bar(4);
    const b2 = self.bar(3);
    const b3 = self.bar(2);
    const b4 = self.bar(1);

    // Three same color -- crisp gate.
    // Three same color — crisp gate.
    const color1: i32 = if (!isWhite(b1.o, b1.c)) -1 else 1;
    const color2: i32 = if (!isWhite(b2.o, b2.c)) -1 else 1;
    const color3: i32 = if (!isWhite(b3.o, b3.c)) -1 else 1;
    const color4: i32 = if (!isWhite(b4.o, b4.c)) -1 else 1;

    if (!(color1 == color2 and color2 == color3 and color4 == -color3)) return 0.0;

    // 2nd opens within/near 1st real body -- fuzzy.
    // 2nd opens within/near 1st real body — fuzzy.
    const near4 = self.avgCS(&self.near, 4);
    const near3 = self.avgCS(&self.near, 3);
    const near_width4 = if (near4 > 0.0) self.fuzz_ratio * near4 else 0.0;
    const near_width3 = if (near3 > 0.0) self.fuzz_ratio * near3 else 0.0;

    const mu_o2_ge = self.muGeRaw(b2.o, @min(b1.o, b1.c) - near4, near_width4);
    const mu_o2_le = self.muLtRaw(b2.o, @max(b1.o, b1.c) + near4, near_width4);

    // 3rd opens within/near 2nd real body -- fuzzy.
    // 3rd opens within/near 2nd real body — fuzzy.
    const mu_o3_ge = self.muGeRaw(b3.o, @min(b2.o, b2.c) - near3, near_width3);
    const mu_o3_le = self.muLtRaw(b3.o, @max(b2.o, b2.c) + near3, near_width3);

    // Bullish: three white, rising closes, 4th opens above 3rd close, closes below 1st open.
    var bull_signal: f64 = 0.0;
    if (color3 == 1 and b3.c > b2.c and b2.c > b1.c) {
        const rb1 = @abs(b1.c - b1.o);
        const width = if (rb1 > 0.0) self.fuzz_ratio * rb1 else 0.0;
        const mu_o4_above = self.muGtRaw(b4.o, b3.c, width);
        const mu_c4_below = self.muLtRaw(b4.c, b1.o, width);
        const conf = operators.tProductAll(&.{mu_o2_ge, mu_o2_le, mu_o3_ge, mu_o3_le, mu_o4_above, mu_c4_below});
        bull_signal = conf * 100.0;
    }

    // Bearish: three black, falling closes, 4th opens below 3rd close, closes above 1st open.
    var bear_signal: f64 = 0.0;
    if (color3 == -1 and b3.c < b2.c and b2.c < b1.c) {
        const rb1 = @abs(b1.c - b1.o);
        const width = if (rb1 > 0.0) self.fuzz_ratio * rb1 else 0.0;
        const mu_o4_below = self.muLtRaw(b4.o, b3.c, width);
        const mu_c4_above = self.muGtRaw(b4.c, b1.o, width);
        const conf = operators.tProductAll(&.{mu_o2_ge, mu_o2_le, mu_o3_ge, mu_o3_le, mu_o4_below, mu_c4_above});
        bear_signal = -conf * 100.0;
    }

    return if (@abs(bull_signal) >= @abs(bear_signal)) bull_signal else bear_signal;
}
