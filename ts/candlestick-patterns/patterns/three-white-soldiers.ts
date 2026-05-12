/** Three White Soldiers pattern (3-candle bullish). */
import { CandlestickPatternsEngine } from '../core/engine.ts';
import { tProductAll } from '../../fuzzy/index.ts';
import { realBody, upperShadow, isWhite } from '../core/primitives.ts';

/**
 * Three White Soldiers: a three-candle bullish pattern.
 *
 * Must have:
 * - three consecutive white candles with consecutively higher closes,
 * - all three have very short upper shadows,
 * - each opens within or near the prior candle's real body,
 * - none is far shorter than the prior candle,
 * - third candle is not short.
 *
 * Category A: always bullish (continuous).
 *
 * Returns:
 *     Continuous float in [0, 100].  Always bullish.
 */
export function threeWhiteSoldiers(cp: CandlestickPatternsEngine): number {
    if (!cp.enough(3, cp.shortBody, cp.veryShortShadow, cp.near, cp.far)) return 0.0;

    const b1 = cp.bar(3);
    const b2 = cp.bar(2);
    const b3 = cp.bar(1);

    // Crisp gates: all white with consecutively higher closes.
    if (!(isWhite(b1.o, b1.c) && isWhite(b2.o, b2.c) && isWhite(b3.o, b3.c) &&
        b3.c > b2.c && b2.c > b1.c)) return 0.0;

    const rb1 = realBody(b1.o, b1.c);
    const rb2 = realBody(b2.o, b2.c);
    const rb3 = realBody(b3.o, b3.c);

    // Crisp: each opens above the prior open (ordering).
    if (!(b2.o > b1.o && b3.o > b2.o)) return 0.0;

    // Fuzzy: very short upper shadows (all three).
    const muUS1 = cp.muLessCS(upperShadow(b1.o, b1.h, b1.c), cp.veryShortShadow, 3);
    const muUS2 = cp.muLessCS(upperShadow(b2.o, b2.h, b2.c), cp.veryShortShadow, 2);
    const muUS3 = cp.muLessCS(upperShadow(b3.o, b3.h, b3.c), cp.veryShortShadow, 1);

    // Fuzzy: each opens within or near the prior body (upper bound).
    const near3 = cp.avgCS(cp.near, 3);
    let near3Width = 0.0;
    if (near3 > 0.0) near3Width = cp.fuzzRatio * near3;
    const muO2Near = cp.muLtRaw(b2.o, b1.c + near3, near3Width);

    const near2 = cp.avgCS(cp.near, 2);
    let near2Width = 0.0;
    if (near2 > 0.0) near2Width = cp.fuzzRatio * near2;
    const muO3Near = cp.muLtRaw(b3.o, b2.c + near2, near2Width);

    // Fuzzy: not far shorter than prior candle.
    const far3 = cp.avgCS(cp.far, 3);
    let far3Width = 0.0;
    if (far3 > 0.0) far3Width = cp.fuzzRatio * far3;
    const muNotFar2 = cp.muGtRaw(rb2, rb1 - far3, far3Width);

    const far2 = cp.avgCS(cp.far, 2);
    let far2Width = 0.0;
    if (far2 > 0.0) far2Width = cp.fuzzRatio * far2;
    const muNotFar3 = cp.muGtRaw(rb3, rb2 - far2, far2Width);

    // Fuzzy: third candle is not short.
    const muNotShort3 = cp.muGreaterCS(rb3, cp.shortBody, 1);

    return tProductAll(muUS1, muUS2, muUS3, muO2Near, muO3Near, muNotFar2, muNotFar3, muNotShort3) * 100.0;
}
