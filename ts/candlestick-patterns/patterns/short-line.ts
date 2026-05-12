/** Short Line pattern (1-candle). */
import { CandlestickPatternsEngine } from '../core/engine.ts';
import { tProductAll } from '../../fuzzy/index.ts';
import { realBody, upperShadow, lowerShadow, isWhite, isBlack } from '../core/primitives.ts';

/**
 * Short Line: a one-candle pattern.
 *
 * A candle with a short body, short upper shadow, and short lower shadow.
 *
 * The meaning of "short" for body is specified with `shortBody`.
 * The meaning of "short" for shadows is specified with `shortShadow`.
 *
 * Category C: color determines sign.
 *
 * Returns:
 *     Continuous float in [-100, +100].
 */
export function shortLine(cp: CandlestickPatternsEngine): number {
    if (!cp.enough(1, cp.shortBody, cp.shortShadow)) return 0.0;

    const b = cp.bar(1);

    const muShortBody = cp.muLessCS(realBody(b.o, b.c), cp.shortBody, 1);
    const muShortUS = cp.muLessCS(upperShadow(b.o, b.h, b.c), cp.shortShadow, 1);
    const muShortLS = cp.muLessCS(lowerShadow(b.o, b.l, b.c), cp.shortShadow, 1);

    const confidence = tProductAll(muShortBody, muShortUS, muShortLS);

    if (isWhite(b.o, b.c)) return confidence * 100.0;
    if (isBlack(b.o, b.c)) return -confidence * 100.0;
    return 0.0;
}
