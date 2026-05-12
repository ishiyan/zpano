/** Ladder Bottom pattern (5-candle bullish). */
import { CandlestickPatternsEngine } from '../core/engine.ts';
import { upperShadow, isBlack, isWhite } from '../core/primitives.ts';

/**
 * Ladder Bottom: a five-candle bullish pattern.
 *
 * Must have:
 * - first three candles: descending black candles (each closes lower),
 * - fourth candle: black with a long upper shadow,
 * - fifth candle: white, opens above the fourth candle's real body,
 *   closes above the fourth candle's high.
 *
 * The meaning of "long" for shadows is specified with `cp.longShadow`.
 *
 * Category A: always bullish (continuous).
 *
 * Returns:
 *     Continuous float in [0, 100].  Always bullish.
 */
export function ladderBottom(cp: CandlestickPatternsEngine): number {
    if (!cp.enough(5, cp.veryShortShadow)) return 0.0;

    const b1 = cp.bar(5);
    const b2 = cp.bar(4);
    const b3 = cp.bar(3);
    const b4 = cp.bar(2);
    const b5 = cp.bar(1);

    // Crisp gates: colors.
    if (!(isBlack(b1.o, b1.c) && isBlack(b2.o, b2.c) &&
        isBlack(b3.o, b3.c) && isBlack(b4.o, b4.c) &&
        isWhite(b5.o, b5.c))) return 0.0;
    // Crisp: three descending opens and closes.
    if (!(b1.o > b2.o && b2.o > b3.o && b1.c > b2.c && b2.c > b3.c)) return 0.0;
    // Crisp: fifth opens above fourth's open, closes above fourth's high.
    if (!(b5.o > b4.o && b5.c > b4.h)) return 0.0;

    // Fuzzy: fourth candle has upper shadow > very short avg.
    const muUS4 = cp.muGreaterCS(upperShadow(b4.o, b4.h, b4.c), cp.veryShortShadow, 2);
    return muUS4 * 100.0;
}
