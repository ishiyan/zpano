/// Up/Down-side Gap Three Methods: a three-candle continuation pattern.
///
/// Must have:
/// - first and second candles are the same color with a gap between them,
/// - third candle is opposite color, opens within the second candle's
///   real body and closes within the first candle's real body (fills the
///   gap).
///
/// Upside gap: two white candles with gap up, third is black = bullish.
/// Downside gap: two black candles with gap down, third is white = bearish.
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

pub fn xSideGapThreeMethods(self: *const CandlestickPatterns) f64 {
            if (!self.enough(3, &[_]*const CriterionState{})) return 0.0;

    const b1 = self.bar(3);
    const b2 = self.bar(2);
    const b3 = self.bar(1);

    // Upside gap: two whites gap up, third black fills.
    var bull_signal: f64 = 0.0;
    if (isWhite(b1.o, b1.c) and isWhite(b2.o, b2.c) and isBlack(b3.o, b3.c) and isRealBodyGapUp(b1.o, b1.c, b2.o, b2.c))
    {
        const rb2 = realBodyLen(b2.o, b2.c);
        const width = if (rb2 > 0.0) self.fuzz_ratio * rb2 else 0.0;
        // o3 within 2nd body: o3 < c2 and o3 > o2
        const mu_o3_lt_c2 = self.muLtRaw(b3.o, b2.c, width);
        const mu_o3_gt_o2 = self.muGtRaw(b3.o, b2.o, width);
        // c3 within 1st body: c3 > o1 and c3 < c1
        const rb1 = realBodyLen(b1.o, b1.c);
        const width1 = if (rb1 > 0.0) self.fuzz_ratio * rb1 else 0.0;
        const mu_c3_gt_o1 = self.muGtRaw(b3.c, b1.o, width1);
        const mu_c3_lt_c1 = self.muLtRaw(b3.c, b1.c, width1);
        const conf = operators.tProductAll(&.{mu_o3_lt_c2, mu_o3_gt_o2, mu_c3_gt_o1, mu_c3_lt_c1});
        bull_signal = conf * 100.0;
    }

    // Downside gap: two blacks gap down, third white fills.
    var bear_signal: f64 = 0.0;
    if (isBlack(b1.o, b1.c) and isBlack(b2.o, b2.c) and isWhite(b3.o, b3.c) and isRealBodyGapDown(b1.o, b1.c, b2.o, b2.c))
    {
        const rb2 = realBodyLen(b2.o, b2.c);
        const width = if (rb2 > 0.0) self.fuzz_ratio * rb2 else 0.0;
        // o3 within 2nd body: o3 > c2 and o3 < o2
        const mu_o3_gt_c2 = self.muGtRaw(b3.o, b2.c, width);
        const mu_o3_lt_o2 = self.muLtRaw(b3.o, b2.o, width);
        // c3 within 1st body: c3 < o1 and c3 > c1
        const rb1 = realBodyLen(b1.o, b1.c);
        const width1 = if (rb1 > 0.0) self.fuzz_ratio * rb1 else 0.0;
        const mu_c3_lt_o1 = self.muLtRaw(b3.c, b1.o, width1);
        const mu_c3_gt_c1 = self.muGtRaw(b3.c, b1.c, width1);
        const conf = operators.tProductAll(&.{mu_o3_gt_c2, mu_o3_lt_o2, mu_c3_lt_o1, mu_c3_gt_c1});
        bear_signal = -conf * 100.0;
    }

    return if (@abs(bull_signal) >= @abs(bear_signal)) bull_signal else bear_signal;
}
