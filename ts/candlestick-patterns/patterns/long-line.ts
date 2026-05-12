/** Long Line pattern (1-candle). */
import { CandlestickPatternsEngine } from '../core/engine.ts';
import { tProductAll } from '../../fuzzy/index.ts';
import { realBody, upperShadow, lowerShadow, isWhite } from '../core/primitives.ts';

/**
 * Long Line: a one-candle pattern.
 *
 * Must have:
 * - long real body,
 * - short upper shadow,
 * - short lower shadow.
 *
 * The meaning of "long" is specified with `cp.longBody`.
 * The meaning of "short" for shadows is specified with `cp.shortShadow`.
 *
 * Category B: direction from candle color.
 *
 * Returns:
 *     Continuous float in [-100, +100].
 */
export function longLine(cp: CandlestickPatternsEngine): number {
    if (!cp.enough(1, cp.longBody, cp.shortShadow)) return 0.0;

    const b = cp.bar(1);
    // Fuzzy: long body, short shadows.
    const muLong = cp.muGreaterCS(realBody(b.o, b.c), cp.longBody, 1);
    const muUS = cp.muLessCS(upperShadow(b.o, b.h, b.c), cp.shortShadow, 1);
    const muLS = cp.muLessCS(lowerShadow(b.o, b.l, b.c), cp.shortShadow, 1);

    const confidence = tProductAll(muLong, muUS, muLS);
    // Crisp direction from color.
    const direction = isWhite(b.o, b.c) ? 1 : -1;
    return direction * confidence * 100.0;
}
