/** Thrusting pattern (2-candle bearish). */
import { CandlestickPatternsEngine } from '../core/engine.ts';
import { tProductAll } from '../../fuzzy/index.ts';
import { realBody, isBlack, isWhite } from '../core/primitives.ts';

/**
 * Thrusting: a two-candle bearish continuation pattern.
 *
 * Must have:
 * - first candle: long black,
 * - second candle: white, opens below the prior candle's low, closes
 *   into the prior candle's real body but below the midpoint, and the
 *   close is not equal to the prior candle's close (to distinguish
 *   from in-neck).
 *
 * The meaning of "long" is specified with `longBody`.
 * The meaning of "equal" is specified with `equal`.
 *
 * Category A: always bearish (continuous).
 *
 * Returns:
 *     Continuous float in [-100, 0].  Always bearish.
 */
export function thrusting(cp: CandlestickPatternsEngine): number {
    if (!cp.enough(2, cp.longBody, cp.equal)) return 0.0;

    const b1 = cp.bar(2);
    const b2 = cp.bar(1);

    const rb1 = realBody(b1.o, b1.c);

    // Crisp gates: color checks and open below prior low.
    if (!(isBlack(b1.o, b1.c) && isWhite(b2.o, b2.c) && b2.o < b1.l)) return 0.0;

    // Fuzzy conditions.
    const muLong1 = cp.muGreaterCS(rb1, cp.longBody, 2);

    // Close above prior close + equal avg (not equal to prior close).
    const eq = cp.avgCS(cp.equal, 2);
    let eqWidth = 0.0;
    if (eq > 0.0) eqWidth = cp.fuzzRatio * eq;
    const muAboveClose = cp.muGtRaw(b2.c, b1.c + eq, eqWidth);

    // Close at or below midpoint of prior body: c2 <= c1 + rb1 * 0.5
    const mid = b1.c + rb1 * 0.5;
    let midWidth = 0.0;
    if (rb1 > 0.0) midWidth = cp.fuzzRatio * rb1 * 0.5;
    const muBelowMid = cp.muLtRaw(b2.c, mid, midWidth);

    return -tProductAll(muLong1, muAboveClose, muBelowMid) * 100.0;
}
