/** Mat Hold pattern (5-candle bullish). */
import { CandlestickPatternsEngine } from '../core/engine.ts';
import { tProductAll } from '../../fuzzy/index.ts';
import { realBody, isWhite, isBlack, isRealBodyGapUp } from '../core/primitives.ts';

const MAT_HOLD_PENETRATION = 0.5;

/**
 * Mat Hold: a five-candle bullish continuation pattern.
 *
 * Must have:
 * - first candle: long white,
 * - second candle: small, black, gaps up from first,
 * - third and fourth candles: small,
 * - reaction candles (2-4) are falling, hold within first body
 *   (penetration check),
 * - fifth candle: white, opens above prior close, closes above
 *   highest high of reaction candles.
 *
 * Category A: always bullish (continuous).
 *
 * Returns:
 *     Continuous float in [0, 100].  Always bullish.
 */
export function matHold(cp: CandlestickPatternsEngine): number {
    if (!cp.enough(5, cp.longBody, cp.shortBody)) return 0.0;

    const b1 = cp.bar(5);
    const b2 = cp.bar(4);
    const b3 = cp.bar(3);
    const b4 = cp.bar(2);
    const b5 = cp.bar(1);

    // Crisp gates: colors.
    if (!(isWhite(b1.o, b1.c) && isBlack(b2.o, b2.c) && isWhite(b5.o, b5.c))) return 0.0;
    // Crisp: gap up from 1st to 2nd.
    if (!isRealBodyGapUp(b1.o, b1.c, b2.o, b2.c)) return 0.0;
    // Crisp: 3rd to 4th hold within 1st range.
    if (!(Math.min(b3.o, b3.c) < b1.c && Math.min(b4.o, b4.c) < b1.c)) return 0.0;
    // Crisp: reaction days don't penetrate first body too much.
    const rb1 = realBody(b1.o, b1.c);
    if (!(Math.min(b3.o, b3.c) > b1.c - rb1 * MAT_HOLD_PENETRATION &&
        Math.min(b4.o, b4.c) > b1.c - rb1 * MAT_HOLD_PENETRATION)) return 0.0;
    // Crisp: 2nd to 4th are falling.
    if (!(Math.max(b3.o, b3.c) < b2.o && Math.max(b4.o, b4.c) < Math.max(b3.o, b3.c))) return 0.0;
    // Crisp: 5th opens above prior close.
    if (!(b5.o > b4.c)) return 0.0;
    // Crisp: 5th closes above highest high of reaction candles.
    if (!(b5.c > Math.max(b2.h, Math.max(b3.h, b4.h)))) return 0.0;

    // Fuzzy: first candle long.
    const muLong1 = cp.muGreaterCS(rb1, cp.longBody, 5);
    // Fuzzy: 2nd, 3rd, 4th short.
    const muShort2 = cp.muLessCS(realBody(b2.o, b2.c), cp.shortBody, 4);
    const muShort3 = cp.muLessCS(realBody(b3.o, b3.c), cp.shortBody, 3);
    const muShort4 = cp.muLessCS(realBody(b4.o, b4.c), cp.shortBody, 2);

    return tProductAll(muLong1, muShort2, muShort3, muShort4) * 100.0;
}
