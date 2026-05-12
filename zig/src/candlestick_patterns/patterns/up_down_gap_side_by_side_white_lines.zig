/// Up/Down-Gap Side-By-Side White Lines: a three-candle pattern.
///
/// Must have:
/// - first candle: white (for up gap) or black (for down gap),
/// - gap (up or down) between the first and second candle — both 2nd AND
///   3rd must gap from the 1st,
/// - second and third candles are both white with similar size and
///   approximately the same open.
///
/// Up gap = bullish continuation, down gap = bearish continuation.
///
/// Category C: both branches evaluated, return stronger signal.
///
/// Returns:
///     Continuous float in [-100, +100].

const cp = @import("../candlestick_patterns.zig");
const operators = @import("fuzzy").operators;

const CandlestickPatterns = cp.CandlestickPatterns;
const CriterionState = cp.CriterionState;
const isRealBodyGapDown = cp.isRealBodyGapDown;
const isRealBodyGapUp = cp.isRealBodyGapUp;
const isWhite = cp.isWhite;
const realBodyLen = cp.realBodyLen;

pub fn upDownGapSideBySideWhiteLines(self: *const CandlestickPatterns) f64 {
            if (!self.enough(3, &[_]*const CriterionState{&self.near, &self.equal})) return 0.0;

    const b1 = self.bar(3);
    const b2 = self.bar(2);
    const b3 = self.bar(1);

    // Crisp: both 2nd and 3rd must be white.
    if (!(isWhite(b2.o, b2.c) and isWhite(b3.o, b3.c))) return 0.0;

    // Both 2nd and 3rd must gap from 1st in the same direction -- crisp.
    // Both 2nd and 3rd must gap from 1st in the same direction — crisp.
    const gap_up = isRealBodyGapUp(b1.o, b1.c, b2.o, b2.c) and isRealBodyGapUp(b1.o, b1.c, b3.o, b3.c);
    const gap_down = isRealBodyGapDown(b1.o, b1.c, b2.o, b2.c) and isRealBodyGapDown(b1.o, b1.c, b3.o, b3.c);

    if (!(gap_up or gap_down)) return 0.0;

    const rb2 = realBodyLen(b2.o, b2.c);
    const rb3 = realBodyLen(b3.o, b3.c);

    // Fuzzy: similar size and same open.
    const mu_near_size = self.muLessCs(@abs(rb2 - rb3), &self.near, 2);
    const mu_equal_open = self.muLessCs(@abs(b3.o - b2.o), &self.equal, 2);

    const conf = operators.tProductAll(&.{mu_near_size, mu_equal_open});

    if (gap_up) {
        return conf * 100.0;
    } else {
        return -conf * 100.0;
    }
}
