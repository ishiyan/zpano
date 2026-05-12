/** Homing Pigeon pattern (2-candle bullish). */
import { CandlestickPatternsEngine } from '../core/engine.ts';
import { tProductAll } from '../../fuzzy/index.ts';
import { realBody, isBlack } from '../core/primitives.ts';

/**
 * Homing Pigeon: a two-candle bullish pattern.
 *
 * Must have:
 * - first candle: long black,
 * - second candle: short black, real body engulfed by first candle's
 *   real body.
 *
 * The meaning of "long" is specified with longBody.
 * The meaning of "short" is specified with shortBody.
 *
 * Category A: always bullish (continuous).
 *
 * Returns:
 *     Continuous number in [0, 100]. Always bullish.
 */
export function homingPigeon(cp: CandlestickPatternsEngine): number {
    if (!cp.enough(2, cp.longBody, cp.shortBody)) return 0.0;

    const b1 = cp.bar(2);
    const b2 = cp.bar(1);

    // Crisp gates: both black.
    if (!(isBlack(b1.o, b1.c) && isBlack(b2.o, b2.c))) return 0.0;

    // Fuzzy conditions.
    const muLong1 = cp.muGreaterCS(realBody(b1.o, b1.c), cp.longBody, 2);
    const muShort2 = cp.muLessCS(realBody(b2.o, b2.c), cp.shortBody, 1);

    // Containment: second body engulfed by first body.
    // For black candles: open > close, so upper = open, lower = close.
    const eqWidth = cp.fuzzRatio * cp.avgCS(cp.equal, 2);
    const muEncUpper = cp.muLtRaw(b2.o, b1.o, eqWidth);
    const muEncLower = cp.muGtRaw(b2.c, b1.c, eqWidth);

    return tProductAll(muLong1, muShort2, muEncUpper, muEncLower) * 100.0;
}
