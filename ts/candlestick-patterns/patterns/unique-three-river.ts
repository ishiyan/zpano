/** Unique Three River pattern (3-candle bullish). */
import { CandlestickPatternsEngine } from '../core/engine.ts';
import { tProductAll } from '../../fuzzy/index.ts';
import { realBody, isBlack, isWhite } from '../core/primitives.ts';

/**
 * Unique Three River: a three-candle bullish pattern.
 *
 * Must have:
 * - first candle: long black,
 * - second candle: black harami (body within first body) with a lower
 *   low than the first candle,
 * - third candle: small white, opens not lower than the second candle's
 *   low.
 *
 * The meaning of "long" is specified with `longBody`.
 * The meaning of "short" is specified with `shortBody`.
 *
 * Category A: always bullish (continuous).
 *
 * Returns:
 *     Continuous float in [0, 100].  Always bullish.
 */
export function uniqueThreeRiver(cp: CandlestickPatternsEngine): number {
    if (!cp.enough(3, cp.longBody, cp.shortBody)) return 0.0;

    const b1 = cp.bar(3);
    const b2 = cp.bar(2);
    const b3 = cp.bar(1);

    // Crisp gates: colors.
    if (!(isBlack(b1.o, b1.c) && isBlack(b2.o, b2.c) && isWhite(b3.o, b3.c))) return 0.0;
    // Crisp: harami body containment and lower low.
    if (!(b2.c > b1.c && b2.o <= b1.o && b2.l < b1.l)) return 0.0;
    // Crisp: third opens not lower than second's low.
    if (!(b3.o >= b2.l)) return 0.0;

    // Fuzzy: first candle is long.
    const muLong1 = cp.muGreaterCS(realBody(b1.o, b1.c), cp.longBody, 3);
    // Fuzzy: third candle is short.
    const muShort3 = cp.muLessCS(realBody(b3.o, b3.c), cp.shortBody, 1);

    return tProductAll(muLong1, muShort3) * 100.0;
}
