/** Piercing pattern (2-candle bullish). */
import { CandlestickPatternsEngine } from '../core/engine.ts';
import { tProductAll } from '../../fuzzy/index.ts';
import { realBody, isBlack, isWhite } from '../core/primitives.ts';

/**
 * Piercing: a two-candle bullish reversal pattern.
 *
 * Must have:
 * - first candle: long black,
 * - second candle: long white that opens below the prior low and closes
 *   above the midpoint of the first candle's real body but within the body.
 *
 * Returns:
 *     Continuous float in [0, 100].  Higher = stronger signal.
 */
export function piercing(cp: CandlestickPatternsEngine): number {
    if (!cp.enough(2, cp.longBody)) return 0.0;

    const b1 = cp.bar(2);
    const b2 = cp.bar(1);

    // Color checks stay crisp
    if (!isBlack(b1.o, b1.c) || !isWhite(b2.o, b2.c)) return 0.0;

    const rb1 = realBody(b1.o, b1.c);
    const eqAvg = cp.avgCS(cp.equal, 1);
    let eqWidth = 0.0;
    if (eqAvg > 0.0) eqWidth = cp.fuzzRatio * eqAvg;

    const muLong1 = cp.muGreaterCS(rb1, cp.longBody, 2);
    const muLong2 = cp.muGreaterCS(realBody(b2.o, b2.c), cp.longBody, 1);
    const muOpenBelow = cp.muLtRaw(b2.o, b1.l, eqWidth);
    const penThreshold = b1.c + rb1 * 0.5;
    let penWidth = 0.0;
    if (rb1 > 0.0) penWidth = cp.fuzzRatio * rb1 * 0.5;
    const muPen = cp.muGtRaw(b2.c, penThreshold, penWidth);
    const muBelowOpen1 = cp.muLtRaw(b2.c, b1.o, eqWidth);

    return tProductAll(muLong1, muLong2, muOpenBelow, muPen, muBelowOpen1) * 100.0;
}
