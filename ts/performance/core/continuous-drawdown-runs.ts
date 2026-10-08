import { KleinKbnAccumulator } from '../../streaming-kbn';

/** Convert a compounded log return into a percentage drawdown. */
export function ddPercent(logsum: number): number {
    return Math.expm1(logsum) * 100.0;
}

/**
 * Streaming 'continuous' drawdown runs for Burke-type measures.
 *
 * A continuous drawdown is the compounded loss over a maximal run of
 * consecutive negative returns.  Following PerformanceAnalytics
 * `BurkeRatio`, returns are compounded as if they were percentages:
 *
 *     DD = (prod(1 + r_i * 0.01) - 1) * 100
 *
 * For decimal returns this is close to the sum of the run's returns,
 * not their compounded return; the quirk is kept to match R.
 *
 * The Burke denominator is:
 *
 *     sqrt(sum(DD_j^2))
 *
 * where the sum is taken over all continuous losing runs in the current
 * window.
 *
 * This class is a pure accumulator: the caller owns the rolling window
 * and feeds evicted values to `revert()` and new values to `update()`.
 * Within one step, call `revert(old)` BEFORE `update(new)` so run adjacency stays correct.
 *
 * The complexity is O(1) per call (amortized; eviction of a run shifts the run array).
 */
export class ContinuousDrawdownRuns {
    /** Each run: [logsum, count]. */
    private readonly _runs: [number, number][] = [];

    /** Sum of squared continuous drawdowns. */
    private readonly _sumSq = new KleinKbnAccumulator();

    /** Whether the most recent return is negative. */
    private _lastWasNegative = false;
    private _count = 0;

    reset(): void {
        this._count = 0;
        this._runs.length = 0;
        this._sumSq.reset();
        this._lastWasNegative = false;
    }

    /** Remove the oldest return from the left edge of the window. */
    revert(oldRet: number): void {
        if (oldRet < 0) {
            // oldest negative is the front of the left-most run, shrink it
            const run = this._runs[0];
            let d = ddPercent(run[0]);
            this._sumSq.revert(d * d);
            run[0] -= Math.log1p(oldRet * 0.01);
            run[1] -= 1;
            if (run[1] === 0) {
                this._runs.shift(); // run fully evicted
            } else {
                d = ddPercent(run[0]);
                this._sumSq.update(d * d);
            }
        }
        // oldRet >= 0 is a separator, nothing to update
    }

    /** Add a new (most-recent) return at the right edge of the window. */
    update(ret: number): void {
        if (ret < 0) {
            const logr = Math.log1p(ret * 0.01);
            if (this._lastWasNegative && this._runs.length > 0) {
                // Extend the currently-open (right-most) run.
                const run = this._runs[this._runs.length - 1];
                let d = ddPercent(run[0]);
                this._sumSq.revert(d * d);
                run[0] += logr;
                run[1] += 1;
                d = ddPercent(run[0]);
                this._sumSq.update(d * d);
            } else {
                // Start a new run.
                this._runs.push([logr, 1]);
                const d = ddPercent(logr);
                this._sumSq.update(d * d);
            }
            this._lastWasNegative = true;
        } else {
            // Non-negative return closes any open run (already counted) — a separator.
            this._lastWasNegative = false;
        }
    }

    /**
     * Continuous drawdowns (negative percentages), one value for each losing run.
     * Returns a new array on each call.
     */
    get drawdowns(): number[] {
        return this._runs.map(([logsum]) => ddPercent(logsum));
    }

    /** Sum of squared continuous drawdowns. */
    get sumDrawdownsSquared(): number {
        const v = this._sumSq.value;
        return 0.0 > v ? 0.0 : v; // Python max(v, 0.0)
    }

    /**
     * Square root of the sum of squared continuous drawdowns.
     *
     * This is the denominator used by the Burke ratio.
     */
    get sqrtSumDrawdownsSquared(): number {
        const v = this._sumSq.value;
        return Math.sqrt(0.0 > v ? 0.0 : v); // Python max(v, 0.0)
    }

    /** Number of continuous losing runs in the current window. */
    get runCount(): number {
        return this._runs.length;
    }
}
