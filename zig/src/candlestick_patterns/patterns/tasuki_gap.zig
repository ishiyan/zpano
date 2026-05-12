/// Tasuki Gap: a three-candle continuation pattern.
///
/// Upside Tasuki Gap (bullish):
/// - real-body gap up between 1st and 2nd candles,
/// - 2nd candle: white,
/// - 3rd candle: black, opens within 2nd white body, closes below 2nd
///   open but above 1st candle's real body top (inside the gap),
/// - 2nd and 3rd have near-equal body sizes.
///
/// Downside Tasuki Gap (bearish):
/// - real-body gap down between 1st and 2nd candles,
/// - 2nd candle: black,
/// - 3rd candle: white, opens within 2nd black body, closes above 2nd
///   open but below 1st candle's real body bottom (inside the gap),
/// - 2nd and 3rd have near-equal body sizes.
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

pub fn tasukiGap(self: *const CandlestickPatterns) f64 {
            if (!self.enough(3, &[_]*const CriterionState{&self.near})) return 0.0;

    const b1 = self.bar(3);
    const b2 = self.bar(2);
    const b3 = self.bar(1);

    // Upside Tasuki Gap (bullish).
    var bull_signal: f64 = 0.0;
    if (isRealBodyGapUp(b1.o, b1.c, b2.o, b2.c) and isWhite(b2.o, b2.c) and isBlack(b3.o, b3.c))
    {
        const rb2 = realBodyLen(b2.o, b2.c);
        const rb3 = realBodyLen(b3.o, b3.c);
        const width = if (rb2 > 0.0) self.fuzz_ratio * rb2 else 0.0;
        // o3 within 2nd body: o3 < c2 and o3 > o2
        const mu_o3_lt_c2 = self.muLtRaw(b3.o, b2.c, width);
        const mu_o3_gt_o2 = self.muGtRaw(b3.o, b2.o, width);
        // c3 below o2
        const mu_c3_lt_o2 = self.muLtRaw(b3.c, b2.o, width);
        // c3 above 1st body top (inside gap)
        const body1_top = @max(b1.c, b1.o);
        const mu_c3_gt_top1 = self.muGtRaw(b3.c, body1_top, width);
        // near-equal bodies
        const mu_near = self.muLessCs(@abs(rb2 - rb3), &self.near, 2);
        const conf = operators.tProductAll(&.{mu_o3_lt_c2, mu_o3_gt_o2, mu_c3_lt_o2, mu_c3_gt_top1, mu_near});
        bull_signal = conf * 100.0;
    }

    // Downside Tasuki Gap (bearish).
    var bear_signal: f64 = 0.0;
    if (isRealBodyGapDown(b1.o, b1.c, b2.o, b2.c) and isBlack(b2.o, b2.c) and isWhite(b3.o, b3.c))
    {
        const rb2 = realBodyLen(b2.o, b2.c);
        const rb3 = realBodyLen(b3.o, b3.c);
        const width = if (rb2 > 0.0) self.fuzz_ratio * rb2 else 0.0;
        // o3 within 2nd body: o3 < o2 and o3 > c2
        const mu_o3_lt_o2 = self.muLtRaw(b3.o, b2.o, width);
        const mu_o3_gt_c2 = self.muGtRaw(b3.o, b2.c, width);
        // c3 above o2
        const mu_c3_gt_o2 = self.muGtRaw(b3.c, b2.o, width);
        // c3 below 1st body bottom (inside gap)
        const body1_bot = @min(b1.c, b1.o);
        const mu_c3_lt_bot1 = self.muLtRaw(b3.c, body1_bot, width);
        // near-equal bodies
        const mu_near = self.muLessCs(@abs(rb2 - rb3), &self.near, 2);
        const conf = operators.tProductAll(&.{mu_o3_lt_o2, mu_o3_gt_c2, mu_c3_gt_o2, mu_c3_lt_bot1, mu_near});
        bear_signal = -conf * 100.0;
    }

    return if (@abs(bull_signal) >= @abs(bear_signal)) bull_signal else bear_signal;
}
