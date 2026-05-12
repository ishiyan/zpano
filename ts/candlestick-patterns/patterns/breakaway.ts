/** Breakaway pattern (5-candle reversal). */
import { CandlestickPatternsEngine } from '../core/engine.ts';
import { tProductAll } from '../../fuzzy/index.ts';
import { realBody, isWhite, isBlack, isRealBodyGapUp, isRealBodyGapDown } from '../core/primitives.ts';

/**
 * Breakaway: a five-candle reversal pattern.
 *
 * Bullish: first candle is long black, second candle is black gapping down,
 * third and fourth candles have consecutively lower highs and lows, fifth
 * candle is white closing into the gap (between first and second candle's
 * real bodies).
 *
 * Bearish: mirror image with colors reversed and gaps reversed.
 *
 * Category C: both branches evaluated, return stronger signal.
 *
 * Returns:
 *     Continuous float in [-100, +100].
 */
export function breakaway(cp: CandlestickPatternsEngine): number {
    if (!cp.enough(5, cp.longBody)) return 0.0;

    const b1 = cp.bar(5);
    const b2 = cp.bar(4);
    const b3 = cp.bar(3);
    const b4 = cp.bar(2);
    const b5 = cp.bar(1);

    // Fuzzy: 1st candle is long.
    const muLong1 = cp.muGreaterCS(realBody(b1.o, b1.c), cp.longBody, 5);

    // Bullish breakaway.
    let bullSignal = 0.0;
    if (isBlack(b1.o, b1.c) && isBlack(b2.o, b2.c) &&
        isBlack(b4.o, b4.c) && isWhite(b5.o, b5.c) &&
        b3.h < b2.h && b3.l < b2.l &&
        b4.h < b3.h && b4.l < b3.l &&
        isRealBodyGapDown(b1.o, b1.c, b2.o, b2.c)) {
        const rb1 = realBody(b1.o, b1.c);
        let width = cp.fuzzRatio * rb1;
        if (rb1 <= 0.0) width = 0.0;
        // Fuzzy: c5 > o2 and c5 < c1 (closing into the gap).
        const muC5AboveO2 = cp.muGtRaw(b5.c, b2.o, width);
        const muC5BelowC1 = cp.muLtRaw(b5.c, b1.c, width);
        bullSignal = tProductAll(muLong1, muC5AboveO2, muC5BelowC1) * 100.0;
    }

    // Bearish breakaway.
    let bearSignal = 0.0;
    if (isWhite(b1.o, b1.c) && isWhite(b2.o, b2.c) &&
        isWhite(b4.o, b4.c) && isBlack(b5.o, b5.c) &&
        b3.h > b2.h && b3.l > b2.l &&
        b4.h > b3.h && b4.l > b3.l &&
        isRealBodyGapUp(b1.o, b1.c, b2.o, b2.c)) {
        const rb1 = realBody(b1.o, b1.c);
        let width = cp.fuzzRatio * rb1;
        if (rb1 <= 0.0) width = 0.0;
        const muC5BelowO2 = cp.muLtRaw(b5.c, b2.o, width);
        const muC5AboveC1 = cp.muGtRaw(b5.c, b1.c, width);
        bearSignal = -tProductAll(muLong1, muC5BelowO2, muC5AboveC1) * 100.0;
    }

    return Math.abs(bullSignal) >= Math.abs(bearSignal) ? bullSignal : bearSignal;
}
