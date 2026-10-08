import { KleinKbnAccumulator } from '../../streaming-kbn';

/**
 * Rolling high-water-mark drawdown.
 *
 * Drawdown at each observation is measured from the high-water mark,
 * the highest equity value reached up to that observation within the
 * current rolling window, including the equity at the start of the window:
 *
 *     drawdown_t = equity_t / max(equity_start, equity_1, ..., equity_t) - 1
 *
 * This matches R PerformanceAnalytics `Drawdowns()`, which uses
 * `cummax(c(1, cumprod(1 + R)))`: a first negative return already
 * produces a drawdown.  For a rolling window, `equity_start` is the
 * equity just before the first observation in the window, so the result
 * equals a fresh calculation over the window's returns.
 *
 * Drawdowns are expressed as decimals and are non-positive.
 *
 * Cumulative log-equity is maintained internally so that returns can be
 * accumulated accurately. Running sums of drawdowns and squared drawdowns
 * are maintained using compensated floating-point accumulation.
 *
 * When an observation leaves the window, the window's starting equity
 * changes.  The drawdowns in the window are recomputed only when this
 * changes their high-water marks; otherwise, the update is O(1)
 * (plus an array shift on eviction).
 *
 * A window size of zero (or negative) means an expanding (unbounded) window.
 */
export class HighWaterMarkDrawdown {
    private readonly _windowSize: number;

    /** Cumulative log-equity at each observation. */
    private readonly _cumlog: number[] = [];

    /** Drawdown at each observation, as a decimal (<= 0). */
    private readonly _dd: number[] = [];

    /** Cumulative log return. */
    private readonly _c = new KleinKbnAccumulator();

    /** Log-equity just before the first observation in the window. */
    private _base = 0.0;

    /** Current high-water mark in log-equity space. */
    private _peak = 0.0;

    // Running drawdown aggregates.
    private readonly _sumDd = new KleinKbnAccumulator();
    private readonly _sumDd2 = new KleinKbnAccumulator();

    constructor(windowSize: number) {
        this._windowSize = windowSize && windowSize > 0 ? windowSize : 0;
    }

    /** Reset the accumulator to its initial empty state. */
    reset(): void {
        this._cumlog.length = 0;
        this._dd.length = 0;
        this._sumDd.reset();
        this._sumDd2.reset();
        this._c.reset();
        this._base = 0.0;
        this._peak = 0.0;
    }

    /**
     * Recompute all drawdowns from the cumulative log-equity values.
     *
     * This is required when an observation leaving the rolling window
     * changes the high-water marks of the remaining observations.
     */
    private recompute(): void {
        this._dd.length = 0;
        this._sumDd.reset();
        this._sumDd2.reset();
        let peak = this._base;
        for (const c of this._cumlog) {
            let dd: number;
            if (c >= peak) {
                peak = c;
                dd = 0.0;
            } else {
                dd = Math.expm1(c - peak);
            }
            this._dd.push(dd);
            this._sumDd.update(dd);
            this._sumDd2.update(dd * dd);
        }
        this._peak = peak;
    }

    /**
     * Add a return observation.
     *
     * If the rolling window is full, the oldest observation is removed
     * before the new observation is added.
     *
     * @param ret Period return expressed as a decimal. For example, `0.02`
     *     represents a 2% return and `-0.015` represents a -1.5% return.
     * @returns True if the rolling window required a drawdown recomputation,
     *     otherwise false.
     */
    update(ret: number): boolean {
        let oldBase: number | null = null;
        if (this._windowSize && this._cumlog.length === this._windowSize) {
            const oldC = this._cumlog.shift() as number;
            const oldDd = this._dd.shift() as number;

            this._sumDd.revert(oldDd);
            this._sumDd2.revert(oldDd * oldDd);

            // The evicted observation's equity is the new starting equity.
            // High-water marks of the remaining observations can only
            // change if the old starting equity was above the evicted one.
            if (oldC < this._base) {
                oldBase = this._base;
            }
            this._base = oldC;
        }

        // Global cumulative log-equity.
        this._c.update(Math.log1p(ret));
        const c = this._c.value;
        this._cumlog.push(c);

        // Peaks of all remaining observations were max(oldBase, c0, ..., cj);
        // without oldBase they are max(c0, c1, ..., cj).  They differ only
        // if the new first observation is also below oldBase.
        if (oldBase !== null && this._cumlog[0] < oldBase) {
            this.recompute();
            return true;
        }

        let dd: number;
        if (c >= this._peak) {
            this._peak = c;
            dd = 0.0;
        } else {
            dd = Math.expm1(c - this._peak);
        }

        this._dd.push(dd);
        this._sumDd.update(dd);
        this._sumDd2.update(dd * dd);
        return false;
    }

    /**
     * Drawdowns for observations currently in the window.
     *
     * Returns the internal (live) array without a defensive copy, because
     * this class is private to the package; callers must not modify it.
     */
    get drawdowns(): readonly number[] {
        return this._dd;
    }

    /** The most recent drawdown in the current window (NaN when empty). */
    get drawdown(): number {
        return this._dd.length > 0 ? this._dd[this._dd.length - 1] : NaN;
    }

    /**
     * Maximum drawdown in the current window (NaN when empty).
     *
     * Drawdowns are non-positive, the largest loss
     * is the minimum drawdown value.
     */
    get maximumDrawdown(): number {
        const n = this._dd.length;
        if (n === 0) {
            return NaN;
        }
        // Python min() semantics: a value replaces the current minimum only if smaller.
        let m = this._dd[0];
        for (let i = 1; i < n; i++) {
            if (this._dd[i] < m) {
                m = this._dd[i];
            }
        }
        return m;
    }

    /** Arithmetic mean of drawdowns in the current window (NaN when empty). */
    get drawdownsMean(): number {
        const n = this._dd.length;
        return n ? this._sumDd.value / n : NaN;
    }

    /** Mean squared drawdown in the current window (NaN when empty). */
    get drawdownsSquaredMean(): number {
        const n = this._dd.length;
        return n ? this._sumDd2.value / n : NaN;
    }

    /** Number of observations in the current window. */
    get drawdownsCount(): number {
        return this._dd.length;
    }
}
