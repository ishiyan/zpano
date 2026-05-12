/** Dark Cloud Cover pattern (2-candle bearish). */
import { CandlestickPatternsEngine } from '../core/engine.ts';
import { tProductAll } from '../../fuzzy/index.ts';
import { realBody, isWhite, isBlack } from '../core/primitives.ts';

const DARK_CLOUD_COVER_PENETRATION = 0.5;

/**
 * Dark Cloud Cover: a two-candle bearish reversal pattern.
 *
 * Must have:
 * - first candle: long white candle,
 * - second candle: black candle that opens above the prior high and
 *   closes well within the first candle's real body (below the midpoint).
 *
 * Returns:
 *   Continuous float in [-100, 0].  More negative = stronger signal.
 */
export function darkCloudCover(cp: CandlestickPatternsEngine): number {
    if (!cp.enough(2, cp.longBody)) return 0.0;

    const b1 = cp.bar(2);
    const b2 = cp.bar(1);

    // Color checks stay crisp
    if (!isWhite(b1.o, b1.c) || !isBlack(b2.o, b2.c)) return 0.0;

    const rb1 = realBody(b1.o, b1.c);
    const eqAvg = cp.avgCS(cp.equal, 1);
    let eqWidth = cp.fuzzRatio * eqAvg;
    if (eqAvg <= 0.0) eqWidth = 0.0;

    const muLong = cp.muGreaterCS(rb1, cp.longBody, 2);
    const muOpenAbove = cp.muGtRaw(b2.o, b1.h, eqWidth);
    const penThreshold = b1.c - rb1 * DARK_CLOUD_COVER_PENETRATION;
    const penProduct = rb1 * DARK_CLOUD_COVER_PENETRATION;
    let penWidth = cp.fuzzRatio * penProduct;
    if (penProduct <= 0.0) penWidth = 0.0;
    const muPen = cp.muLtRaw(b2.c, penThreshold, penWidth);
    const muAboveOpen1 = cp.muGtRaw(b2.c, b1.o, eqWidth);

    return -tProductAll(muLong, muOpenAbove, muPen, muAboveOpen1) * 100.0;
}
