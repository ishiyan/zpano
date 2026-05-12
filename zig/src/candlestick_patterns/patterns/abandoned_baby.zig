/// Abandoned Baby: a three-candle reversal pattern.
///
/// Must have:
/// - first candle: long real body,
/// - second candle: doji,
/// - third candle: real body longer than short, opposite color to 1st,
///   closes well within 1st body,
/// - upside/downside gap between 1st and doji (shadows don't touch),
/// - downside/upside gap between doji and 3rd (shadows don't touch).
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
const isHighLowGapDown = cp.isHighLowGapDown;
const isHighLowGapUp = cp.isHighLowGapUp;
const isWhite = cp.isWhite;
const realBodyLen = cp.realBodyLen;

const abandoned_baby_penetration_factor = 0.3;

pub fn abandonedBaby(self: *const CandlestickPatterns) f64 {
            if (!self.enough(3, &[_]*const CriterionState{&self.long_body, &self.doji_body, &self.short_body})) return 0.0;

    const b1 = self.bar(3);
    const b2 = self.bar(2);
    const b3 = self.bar(1);

    // Shared fuzzy conditions: 1st long, 2nd doji, 3rd > short.
    const mu_long1 = self.muGreaterCs(realBodyLen(b1.o, b1.c), &self.long_body, 3);
    const mu_doji2 = self.muLessCs(realBodyLen(b2.o, b2.c), &self.doji_body, 2);
    const mu_short3 = self.muGreaterCs(realBodyLen(b3.o, b3.c), &self.short_body, 1);

    const penetration = abandoned_baby_penetration_factor;

    // Bearish: white-doji-black, gap up then gap down.
    var bear_signal: f64 = 0.0;
    if (isWhite(b1.o, b1.c) and isBlack(b3.o, b3.c)) {
        if (isHighLowGapUp(b1.h, b2.l) and isHighLowGapDown(b2.l, b3.h)) {
            const rb1 = realBodyLen(b1.o, b1.c);
            const pen_threshold = b1.c - rb1 * penetration;
            const pen_width = if (rb1 > 0.0) self.fuzz_ratio * rb1 else 0.0;
            const mu_pen = self.muLtRaw(b3.c, pen_threshold, pen_width);
            const conf_bear = operators.tProductAll(&.{mu_long1, mu_doji2, mu_short3, mu_pen});
            bear_signal = -conf_bear * 100.0;
        }
    }

    // Bullish: black-doji-white, gap down then gap up.
    var bull_signal: f64 = 0.0;
    if (isBlack(b1.o, b1.c) and isWhite(b3.o, b3.c)) {
        if (isHighLowGapDown(b1.l, b2.h) and isHighLowGapUp(b2.h, b3.l)) {
            const rb1 = realBodyLen(b1.o, b1.c);
            const pen_threshold = b1.c + rb1 * penetration;
            const pen_width = if (rb1 > 0.0) self.fuzz_ratio * rb1 else 0.0;
            const mu_pen = self.muGtRaw(b3.c, pen_threshold, pen_width);
            const conf_bull = operators.tProductAll(&.{mu_long1, mu_doji2, mu_short3, mu_pen});
            bull_signal = conf_bull * 100.0;
        }
    }

    return if (@abs(bull_signal) >= @abs(bear_signal)) bull_signal else bear_signal;
}
