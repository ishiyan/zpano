/** Hikkake Modified pattern (4-candle) with stateful confirmation. */
import { CandlestickPatternsEngine } from '../core/engine.ts';

/**
 * Hikkake Modified: a four-candle pattern with near criterion.
 *
 * Returns:
 *     +100.0/-100.0 for detection, +200.0/-200.0 for confirmation, 0.0 otherwise.
 */
export function hikkakeModified(cp: CandlestickPatternsEngine): number {
    if (cp.count < 4) return 0.0;

    // If pattern was just detected this bar (takes priority over confirmation)
    if (cp.hikmodPatternIdx === cp.count && cp.hikmodPatternResult !== 0) {
        return cp.hikmodPatternResult;
    }

    // If just confirmed this bar
    if (cp.hikmodConfirmed) return cp.hikmodLastSignal;

    return 0.0;
}
