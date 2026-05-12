/** Advance Block pattern (3-candle bearish). */
import { CandlestickPatternsEngine } from '../core/engine.ts';
import { tProductAll } from '../../fuzzy/index.ts';
import { realBody, upperShadow, isWhite } from '../core/primitives.ts';

/**
 * Advance Block: a bearish three-candle pattern.
 *
 * Three white candles with consecutively higher closes and opens, but
 * showing signs of weakening (diminishing bodies, growing upper shadows).
 *
 * Category A: always bearish (continuous).
 *
 * Returns:
 *     Continuous float in [-100, 0]. Always bearish.
 */
export function advanceBlock(cp: CandlestickPatternsEngine): number {
    if (!cp.enough(3, cp.longBody, cp.shortShadow, cp.longShadow, cp.near, cp.far)) return 0.0;

    const b1 = cp.bar(3);
    const b2 = cp.bar(2);
    const b3 = cp.bar(1);

    // Crisp gates: all white with rising closes.
    if (!(isWhite(b1.o, b1.c) && isWhite(b2.o, b2.c) && isWhite(b3.o, b3.c) &&
        b3.c > b2.c && b2.c > b1.c)) return 0.0;
    // Crisp: 2nd opens above 1st open.
    if (!(b2.o > b1.o)) return 0.0;
    // Crisp: 3rd opens above 2nd open.
    if (!(b3.o > b2.o)) return 0.0;

    const rb1 = realBody(b1.o, b1.c);
    const rb2 = realBody(b2.o, b2.c);
    const rb3 = realBody(b3.o, b3.c);

    // Fuzzy: 2nd opens within/near 1st body (upper bound).
    const near3 = cp.avgCS(cp.near, 3);
    let near3Width = cp.fuzzRatio * near3;
    if (near3 <= 0.0) near3Width = 0.0;
    const muO2Near = cp.muLtRaw(b2.o, b1.c + near3, near3Width);

    // Fuzzy: 3rd opens within/near 2nd body (upper bound).
    const near2 = cp.avgCS(cp.near, 2);
    let near2Width = cp.fuzzRatio * near2;
    if (near2 <= 0.0) near2Width = 0.0;
    const muO3Near = cp.muLtRaw(b3.o, b2.c + near2, near2Width);

    // Fuzzy: first candle long body.
    const muLong1 = cp.muGreaterCS(rb1, cp.longBody, 3);
    // Fuzzy: first candle short upper shadow.
    const muUS1 = cp.muLessCS(upperShadow(b1.o, b1.h, b1.c), cp.shortShadow, 3);

    // At least one weakness condition must hold (OR → max).
    const far2 = cp.avgCS(cp.far, 3);
    let far2Width = cp.fuzzRatio * far2;
    if (far2 <= 0.0) far2Width = 0.0;
    const far1 = cp.avgCS(cp.far, 2);
    let far1Width = cp.fuzzRatio * far1;
    if (far1 <= 0.0) far1Width = 0.0;
    const near1 = cp.avgCS(cp.near, 2);
    let near1Width = cp.fuzzRatio * near1;
    if (near1 <= 0.0) near1Width = 0.0;

    // Branch 1: 2 far smaller than 1 AND 3 not longer than 2
    const muB1A = cp.muLtRaw(rb2, rb1 - far2, far2Width);
    const muB1B = cp.muLtRaw(rb3, rb2 + near1, near1Width);
    const branch1 = tProductAll(muB1A, muB1B);

    // Branch 2: 3 far smaller than 2
    const branch2 = cp.muLtRaw(rb3, rb2 - far1, far1Width);

    // Branch 3: 3 < 2 AND 2 < 1 AND (3 or 2 has non-short upper shadow)
    let rb3Width = cp.fuzzRatio * rb2;
    if (rb2 <= 0.0) rb3Width = 0.0;
    let rb2Width = cp.fuzzRatio * rb1;
    if (rb1 <= 0.0) rb2Width = 0.0;
    const muB3A = cp.muLtRaw(rb3, rb2, rb3Width);
    const muB3B = cp.muLtRaw(rb2, rb1, rb2Width);
    const muB3US3 = cp.muGreaterCS(upperShadow(b3.o, b3.h, b3.c), cp.shortShadow, 1);
    const muB3US2 = cp.muGreaterCS(upperShadow(b2.o, b2.h, b2.c), cp.shortShadow, 2);
    const branch3 = tProductAll(muB3A, muB3B, Math.max(muB3US3, muB3US2));

    // Branch 4: 3 < 2 AND 3 has long upper shadow
    const muB4A = cp.muLtRaw(rb3, rb2, rb3Width);
    const muB4B = cp.muGreaterCS(upperShadow(b3.o, b3.h, b3.c), cp.longShadow, 1);
    const branch4 = tProductAll(muB4A, muB4B);

    const weakness = Math.max(branch1, branch2, branch3, branch4);
    const confidence = tProductAll(muO2Near, muO3Near, muLong1, muUS1, weakness);
    return -confidence * 100.0;
}
