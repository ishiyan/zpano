/** Three Inside Up/Down pattern (3-candle reversal). */
import { CandlestickPatternsEngine } from '../core/engine.ts';
import { tProductAll } from '../../fuzzy/index.ts';
import { realBody, isWhite, isBlack } from '../core/primitives.ts';

/**
 * Three Inside Up/Down: a three-candle reversal pattern.
 *
 * Three Inside Up (bullish):
 * - first candle: long black,
 * - second candle: short, engulfed by the first candle's real body,
 * - third candle: white, closes above the first candle's open.
 *
 * Three Inside Down (bearish):
 * - first candle: long white,
 * - second candle: short, engulfed by the first candle's real body,
 * - third candle: black, closes below the first candle's open.
 *
 * The meaning of "long" is specified with `longBody`.
 * The meaning of "short" is specified with `shortBody`.
 *
 * Category C: both branches evaluated, return stronger signal.
 *
 * Returns:
 *     Continuous float in [-100, +100].
 */
export function threeInside(cp: CandlestickPatternsEngine): number {
    if (!cp.enough(3, cp.longBody, cp.shortBody)) return 0.0;

    const b1 = cp.bar(3);
    const b2 = cp.bar(2);
    const b3 = cp.bar(1);

    // Shared fuzzy conditions.
    const muLong1 = cp.muGreaterCS(realBody(b1.o, b1.c), cp.longBody, 3);
    const muShort2 = cp.muLessCS(realBody(b2.o, b2.c), cp.shortBody, 2);

    // Fuzzy containment: 1st body encloses 2nd body.
    const eqAvg = cp.avgCS(cp.equal, 2);
    let eqWidth = 0.0;
    if (eqAvg > 0.0) eqWidth = cp.fuzzRatio * eqAvg;
    const muEncUpper = cp.muGeRaw(Math.max(b1.o, b1.c), Math.max(b2.o, b2.c), eqWidth);
    const muEncLower = cp.muLtRaw(Math.min(b1.o, b1.c), Math.min(b2.o, b2.c), eqWidth);

    // Three Inside Up: long black, short engulfed, white closes above 1st open.
    let bullSignal = 0.0;
    if (isBlack(b1.o, b1.c) && isWhite(b3.o, b3.c)) {
        const rb1 = realBody(b1.o, b1.c);
        let width = 0.0;
        if (rb1 > 0.0) width = cp.fuzzRatio * rb1;
        const muCloseAbove = cp.muGtRaw(b3.c, b1.o, width);
        bullSignal = tProductAll(muLong1, muShort2, muEncUpper, muEncLower, muCloseAbove) * 100.0;
    }

    // Three Inside Down: long white, short engulfed, black closes below 1st open.
    let bearSignal = 0.0;
    if (isWhite(b1.o, b1.c) && isBlack(b3.o, b3.c)) {
        const rb1 = realBody(b1.o, b1.c);
        let width = 0.0;
        if (rb1 > 0.0) width = cp.fuzzRatio * rb1;
        const muCloseBelow = cp.muLtRaw(b3.c, b1.o, width);
        bearSignal = -tProductAll(muLong1, muShort2, muEncUpper, muEncLower, muCloseBelow) * 100.0;
    }

    return Math.abs(bullSignal) >= Math.abs(bearSignal) ? bullSignal : bearSignal;
}
