/** Three Outside Up/Down pattern (3-candle reversal). */
import { CandlestickPatternsEngine } from '../core/engine.ts';
import { tProductAll } from '../../fuzzy/index.ts';
import { realBody, isWhite, isBlack } from '../core/primitives.ts';

/**
 * Three Outside Up/Down: a three-candle reversal pattern.
 *
 * Must have:
 * - first and second candles form an engulfing pattern,
 * - third candle confirms the direction by closing higher (up) or
 *   lower (down).
 *
 * Three Outside Up: first candle is black, second is white engulfing
 * the first, third closes higher than the second.
 *
 * Three Outside Down: first candle is white, second is black engulfing
 * the first, third closes lower than the second.
 *
 * Category C: both branches evaluated, return stronger signal.
 *
 * Returns:
 *     Continuous float in [-100, +100].
 */
export function threeOutside(cp: CandlestickPatternsEngine): number {
    if (!cp.enough(3)) return 0.0;

    const b1 = cp.bar(3);
    const b2 = cp.bar(2);
    const b3 = cp.bar(1);

    // Fuzzy engulfment width.
    const eqAvg = cp.avgCS(cp.equal, 1);
    let eqWidth = 0.0;
    if (eqAvg > 0.0) eqWidth = cp.fuzzRatio * eqAvg;

    // Three Outside Up: black + white engulfing + 3rd closes higher.
    let bullSignal = 0.0;
    if (isBlack(b1.o, b1.c) && isWhite(b2.o, b2.c)) {
        const muEncUpper = cp.muGeRaw(Math.max(b2.o, b2.c), Math.max(b1.o, b1.c), eqWidth);
        const muEncLower = cp.muLtRaw(Math.min(b2.o, b2.c), Math.min(b1.o, b1.c), eqWidth);
        const rb2 = realBody(b2.o, b2.c);
        let width = 0.0;
        if (rb2 > 0.0) width = cp.fuzzRatio * rb2;
        const muCloseHigher = cp.muGtRaw(b3.c, b2.c, width);
        bullSignal = tProductAll(muEncUpper, muEncLower, muCloseHigher) * 100.0;
    }

    // Three Outside Down: white + black engulfing + 3rd closes lower.
    let bearSignal = 0.0;
    if (isWhite(b1.o, b1.c) && isBlack(b2.o, b2.c)) {
        const muEncUpper = cp.muGeRaw(Math.max(b2.o, b2.c), Math.max(b1.o, b1.c), eqWidth);
        const muEncLower = cp.muLtRaw(Math.min(b2.o, b2.c), Math.min(b1.o, b1.c), eqWidth);
        const rb2 = realBody(b2.o, b2.c);
        let width = 0.0;
        if (rb2 > 0.0) width = cp.fuzzRatio * rb2;
        const muCloseLower = cp.muLtRaw(b3.c, b2.c, width);
        bearSignal = -tProductAll(muEncUpper, muEncLower, muCloseLower) * 100.0;
    }

    return Math.abs(bullSignal) >= Math.abs(bearSignal) ? bullSignal : bearSignal;
}
