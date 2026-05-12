/** Stalled pattern (3-candle bearish). */
import { CandlestickPatternsEngine } from '../core/engine.ts';
import { tProductAll } from '../../fuzzy/index.ts';
import { realBody, upperShadow, isWhite } from '../core/primitives.ts';

/**
 * Stalled (Deliberation): a three-candle bearish pattern.
 *
 * Three white candles with progressively higher closes:
 * - first candle: long white body,
 * - second candle: long white body, opens within or near the first
 *   candle's body, very short upper shadow,
 * - third candle: small body that rides on the shoulder of the second
 *   (opens near the second's close, accounting for its own body size).
 *
 * Category A: always bearish (continuous).
 *
 * Returns:
 *     Continuous float in [-100, 0].  Always bearish.
 */
export function stalled(cp: CandlestickPatternsEngine): number {
    if (!cp.enough(3, cp.longBody, cp.shortBody, cp.veryShortShadow, cp.near)) return 0.0;

    const b1 = cp.bar(3);
    const b2 = cp.bar(2);
    const b3 = cp.bar(1);

    // Crisp gates: all white, rising closes.
    if (!(isWhite(b1.o, b1.c) && isWhite(b2.o, b2.c) && isWhite(b3.o, b3.c))) return 0.0;
    if (!(b3.c > b2.c && b2.c > b1.c)) return 0.0;
    // Crisp: o2 > o1 (opens above prior open).
    if (!(b2.o > b1.o)) return 0.0;

    const rb3 = realBody(b3.o, b3.c);

    // Fuzzy conditions.
    const muLong1 = cp.muGreaterCS(realBody(b1.o, b1.c), cp.longBody, 3);
    const muLong2 = cp.muGreaterCS(realBody(b2.o, b2.c), cp.longBody, 2);
    const muUS2 = cp.muLessCS(upperShadow(b2.o, b2.h, b2.c), cp.veryShortShadow, 2);

    // o2 <= c1 + nearAvg (opens within or near prior body).
    const near3 = cp.avgCS(cp.near, 3);
    let near3Width = 0.0;
    if (near3 > 0.0) near3Width = cp.fuzzRatio * near3;
    const muO2Near = cp.muLtRaw(b2.o, b1.c + near3, near3Width);

    // Third candle: short body.
    const muShort3 = cp.muLessCS(rb3, cp.shortBody, 1);

    // o3 >= c2 - rb3 - nearAvg (rides on shoulder).
    const near2 = cp.avgCS(cp.near, 2);
    let near2Width = 0.0;
    if (near2 > 0.0) near2Width = cp.fuzzRatio * near2;
    const muO3Shoulder = cp.muGeRaw(b3.o, b2.c - rb3 - near2, near2Width);

    return -tProductAll(muLong1, muLong2, muUS2, muO2Near, muShort3, muO3Shoulder) * 100.0;
}
