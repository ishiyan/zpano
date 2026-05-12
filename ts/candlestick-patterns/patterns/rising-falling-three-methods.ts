/** Rising/Falling Three Methods pattern (5-candle continuation). */
import { CandlestickPatternsEngine } from '../core/engine.ts';
import { tProductAll } from '../../fuzzy/index.ts';
import { realBody, isWhite } from '../core/primitives.ts';

/**
 * Rising/Falling Three Methods: a five-candle continuation pattern.
 *
 * Uses TA-Lib logic: opposite-color check via color multiplication,
 * real-body overlap (not full candle containment), sequential closes,
 * 5th opens beyond 4th close.
 *
 * Category B: direction from 1st candle color (crisp sign).
 *
 * Returns:
 *     Continuous float in [-100, +100].
 */
export function risingFallingThreeMethods(cp: CandlestickPatternsEngine): number {
    if (!cp.enough(5, cp.longBody, cp.shortBody)) return 0.0;

    const b1 = cp.bar(5);
    const b2 = cp.bar(4);
    const b3 = cp.bar(3);
    const b4 = cp.bar(2);
    const b5 = cp.bar(1);

    // Fuzzy: 1st long, 2nd-4th short, 5th long.
    const muLong1 = cp.muGreaterCS(realBody(b1.o, b1.c), cp.longBody, 5);
    const muShort2 = cp.muLessCS(realBody(b2.o, b2.c), cp.shortBody, 4);
    const muShort3 = cp.muLessCS(realBody(b3.o, b3.c), cp.shortBody, 3);
    const muShort4 = cp.muLessCS(realBody(b4.o, b4.c), cp.shortBody, 2);
    const muLong5 = cp.muGreaterCS(realBody(b5.o, b5.c), cp.longBody, 1);

    // Determine color of 1st candle: +1 white, -1 black — crisp sign.
    const color1 = isWhite(b1.o, b1.c) ? 1.0 : -1.0;

    // Color check: white, 3 black, white  OR  black, 3 white, black — crisp.
    const c2 = isWhite(b2.o, b2.c) ? 1.0 : -1.0;
    const c3 = isWhite(b3.o, b3.c) ? 1.0 : -1.0;
    const c4 = isWhite(b4.o, b4.c) ? 1.0 : -1.0;
    const c5 = isWhite(b5.o, b5.c) ? 1.0 : -1.0;

    if (!(c2 === -color1 && c3 === c2 && c4 === c3 && c5 === -c4)) return 0.0;

    // 2nd to 4th hold within 1st: a part of the real body overlaps 1st range — crisp.
    if (!(Math.min(b2.o, b2.c) < b1.h && Math.max(b2.o, b2.c) > b1.l &&
        Math.min(b3.o, b3.c) < b1.h && Math.max(b3.o, b3.c) > b1.l &&
        Math.min(b4.o, b4.c) < b1.h && Math.max(b4.o, b4.c) > b1.l)) return 0.0;

    // 2nd to 4th are falling (rising) — using color multiply trick — crisp.
    if (!(b3.c * color1 < b2.c * color1 && b4.c * color1 < b3.c * color1)) return 0.0;
    // 5th opens above (below) the prior close — crisp.
    if (!(b5.o * color1 > b4.c * color1)) return 0.0;
    // 5th closes above (below) the 1st close — crisp.
    if (!(b5.c * color1 > b1.c * color1)) return 0.0;

    return color1 * tProductAll(muLong1, muShort2, muShort3, muShort4, muLong5) * 100.0;
}
