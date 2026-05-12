/// Breakaway: a five-candle reversal pattern.
///
/// Bullish: first candle is long black, second candle is black gapping down,
/// third and fourth candles have consecutively lower highs and lows, fifth
/// candle is white closing into the gap (between first and second candle's
/// real bodies).
///
/// Bearish: mirror image with colors reversed and gaps reversed.
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
const isRealBodyGapDown = cp.isRealBodyGapDown;
const isRealBodyGapUp = cp.isRealBodyGapUp;
const isWhite = cp.isWhite;
const realBodyLen = cp.realBodyLen;

pub fn breakaway(self: *const CandlestickPatterns) f64 {
            if (!self.enough(5, &[_]*const CriterionState{&self.long_body})) return 0.0;

    const b1 = self.bar(5);
    const b2 = self.bar(4);
    const b3 = self.bar(3);
    const b4 = self.bar(2);
    const b5 = self.bar(1);

    // Fuzzy: 1st candle is long.
    // Fuzzy: c5 > o2 and c5 < c1 (closing into the gap).
    const mu_long1 = self.muGreaterCs(realBodyLen(b1.o, b1.c), &self.long_body, 5);

    // Bullish breakaway.
    var bull_signal: f64 = 0.0;
    if (isBlack(b1.o, b1.c) and isBlack(b2.o, b2.c) and isBlack(b4.o, b4.c) and isWhite(b5.o, b5.c) and b3.h < b2.h and b3.l < b2.l and b4.h < b3.h and b4.l < b3.l and isRealBodyGapDown(b1.o, b1.c, b2.o, b2.c))
    {
        const rb1 = realBodyLen(b1.o, b1.c);
        const width = if (rb1 > 0.0) self.fuzz_ratio * rb1 else 0.0;
        const mu_c5_above_o2 = self.muGtRaw(b5.c, b2.o, width);
        const mu_c5_below_c1 = self.muLtRaw(b5.c, b1.c, width);
        const conf = operators.tProductAll(&.{mu_long1, mu_c5_above_o2, mu_c5_below_c1});
        bull_signal = conf * 100.0;
    }

    // Bearish breakaway.
    var bear_signal: f64 = 0.0;
    if (isWhite(b1.o, b1.c) and isWhite(b2.o, b2.c) and isWhite(b4.o, b4.c) and isBlack(b5.o, b5.c) and b3.h > b2.h and b3.l > b2.l and b4.h > b3.h and b4.l > b3.l and isRealBodyGapUp(b1.o, b1.c, b2.o, b2.c))
    {
        const rb1 = realBodyLen(b1.o, b1.c);
        const width = if (rb1 > 0.0) self.fuzz_ratio * rb1 else 0.0;
        const mu_c5_below_o2 = self.muLtRaw(b5.c, b2.o, width);
        const mu_c5_above_c1 = self.muGtRaw(b5.c, b1.c, width);
        const conf = operators.tProductAll(&.{mu_long1, mu_c5_below_o2, mu_c5_above_c1});
        bear_signal = -conf * 100.0;
    }

    return if (@abs(bull_signal) >= @abs(bear_signal)) bull_signal else bear_signal;
}
