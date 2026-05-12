/** Spinning Top pattern (1-candle). */
import { CandlestickPatternsEngine } from '../core/engine.ts';
import { tProductAll } from '../../fuzzy/index.ts';
import { realBody, upperShadow, lowerShadow, isWhite, isBlack } from '../core/primitives.ts';

/**
 * Spinning Top: a one-candle pattern.
 *
 * A candle with a small body and shadows longer than the body on both sides.
 *
 * The meaning of "short" is specified with `shortBody`.
 *
 * Category C: color determines sign.
 *
 * Returns:
 *     Continuous float in [-100, +100].
 */
export function spinningTop(cp: CandlestickPatternsEngine): number {
    if (!cp.enough(1, cp.shortBody)) return 0.0;

    const b = cp.bar(1);
    const rb = realBody(b.o, b.c);

    const muShort = cp.muLessCS(rb, cp.shortBody, 1);

    // Shadows > body: positional comparisons.
    const us = upperShadow(b.o, b.h, b.c);
    const ls = lowerShadow(b.o, b.l, b.c);
    let widthUS = 0.0;
    if (rb > 0.0) widthUS = cp.fuzzRatio * rb;
    let widthLS = 0.0;
    if (rb > 0.0) widthLS = cp.fuzzRatio * rb;
    const muUSGtRB = cp.muGtRaw(us, rb, widthUS);
    const muLSGtRB = cp.muGtRaw(ls, rb, widthLS);

    const confidence = tProductAll(muShort, muUSGtRB, muLSGtRB);

    if (isWhite(b.o, b.c)) return confidence * 100.0;
    if (isBlack(b.o, b.c)) return -confidence * 100.0;
    return 0.0;
}
