/** Engulfing pattern (2-candle reversal). */
import { CandlestickPatternsEngine } from '../core/engine.ts';
import { tProductAll } from '../../fuzzy/index.ts';

/**
 * Engulfing: a two-candle reversal pattern.
 *
 * Must have:
 * - first candle and second candle have opposite colors,
 * - second candle's real body engulfs the first (at least one end strictly
 *   exceeds, the other can match).
 *
 * Category B: direction from 2nd candle color (continuous).
 * Opposite-color check stays crisp (doji edge case).
 *
 * Returns:
 *   Continuous float in [-100, +100].  Sign from 2nd candle direction.
 */
export function engulfing(cp: CandlestickPatternsEngine): number {
    if (!cp.enough(2)) return 0.0;

    const b1 = cp.bar(2);
    const b2 = cp.bar(1);

    // Opposite colors — crisp gate (TA-Lib convention: c >= o is white).
    const color1 = b1.c < b1.o ? -1 : 1;
    const color2 = b2.c < b2.o ? -1 : 1;
    if (color1 === color2) return 0.0;

    // Fuzzy engulfment: 2nd body upper >= 1st body upper AND
    //                    2nd body lower <= 1st body lower.
    const upper1 = Math.max(b1.o, b1.c);
    const lower1 = Math.min(b1.o, b1.c);
    const upper2 = Math.max(b2.o, b2.c);
    const lower2 = Math.min(b2.o, b2.c);

    // Width based on the equal criterion for tight comparisons.
    const eqAvg = cp.avgCS(cp.equal, 1);
    let eqWidth = cp.fuzzRatio * eqAvg;
    if (eqAvg <= 0.0) eqWidth = 0.0;

    const muUpper = cp.muGeRaw(upper2, upper1, eqWidth);
    const muLower = cp.muLtRaw(lower2, lower1, eqWidth);

    const confidence = tProductAll(muUpper, muLower);
    // Direction sign from 2nd candle (TA-Lib: c >= o is bullish).
    const direction = b2.c < b2.o ? -1.0 : 1.0;
    return direction * confidence * 100.0;
}
