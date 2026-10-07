import { KleinKbnAccumulator } from './klein-kbn-accumulator';
import { RawMomentsKleinKbn } from './raw-moments-klein-kbn';

/**
 * Streaming ordinary least squares (OLS) regression y = a + b·x with
 * Klein KBN (Kahan-Babuška-Neumaier) compensated accumulation.
 *
 * Tracks the means and variances of x and y (as RawMomentsKleinKbn with
 * ddof=0) and the co-moment S_xy = Σ(x − x̄)(y − ȳ), updated in O(1) per
 * sample (n = count before adding the sample):
 *
 *     S_xy += (x − x̄)·(y − ȳ)·n / (n + 1)
 *
 * where x̄ and ȳ are the means before adding the sample.
 *
 * Like RawMomentsKleinKbn, revert() may remove any previously added
 * (x, y) pair, not only the most recent one, so the class works for FIFO
 * rolling windows.
 *
 * Derived quantities, with S_xx = Σ(x − x̄)² and S_yy = Σ(y − ȳ)²:
 *
 *     slope       b = S_xy / S_xx
 *     intercept   a = ȳ − b·x̄
 *     correlation r = S_xy / √(S_xx·S_yy)
 *     covariance      S_xy / n   (population)
 */
export class LinearRegressionKleinKbn {
    private _n = 0;
    private readonly _xMoments = new RawMomentsKleinKbn(0);
    private readonly _yMoments = new RawMomentsKleinKbn(0);
    private readonly _sXy = new KleinKbnAccumulator();

    /** Clears all accumulated state. */
    reset(): void {
        this._n = 0;
        this._xMoments.reset();
        this._yMoments.reset();
        this._sXy.reset();
    }

    /** Adds a sample (x, y). */
    update(x: number, y: number): void {
        const nOld = this._n;
        this._n += 1;
        const term = (this._xMoments.mean - x) * (this._yMoments.mean - y) * nOld / (nOld + 1);
        this._sXy.update(term);
        this._xMoments.update(x);
        this._yMoments.update(y);
    }

    /**
     * Removes a previously added sample (x, y), not necessarily the most
     * recent one.
     *
     * Throws an Error if there are no samples.
     */
    revert(x: number, y: number): void {
        if (this._n === 0) {
            throw new Error('Cannot revert from an empty regression');
        }
        if (this._n === 1) {
            this.reset();
            return;
        }
        this._xMoments.revert(x);
        this._yMoments.revert(y);
        // The means are now those without (x, y), as in update().
        const n = this._n - 1;
        const term = (this._xMoments.mean - x) * (this._yMoments.mean - y) * n / (n + 1);
        this._sXy.revert(term);
        this._n = n;
    }

    /** The number of samples. */
    get n(): number {
        return this._n;
    }

    /** The mean of x (0.0 when empty). */
    get meanX(): number {
        return this._xMoments.mean;
    }

    /** The mean of y (0.0 when empty). */
    get meanY(): number {
        return this._yMoments.mean;
    }

    /** The population variance of x, S_xx / n (NaN when empty). */
    get varianceX(): number {
        return this._xMoments.variance;
    }

    /** The population variance of y, S_yy / n (NaN when empty). */
    get varianceY(): number {
        return this._yMoments.variance;
    }

    /** The co-moment S_xy = Σ(x − x̄)(y − ȳ) (0.0 when empty). */
    get coMoment(): number {
        return this._sXy.value;
    }

    /** The population covariance S_xy / n (NaN when empty). */
    get covariance(): number {
        const n = this._n;
        if (n < 1) {
            return NaN;
        }
        return this._sXy.value / n;
    }

    /**
     * The OLS slope b = S_xy / S_xx.
     *
     * NaN when n < 2 or all x are equal (S_xx = 0).
     */
    get slope(): number {
        const n = this._n;
        if (n < 2) {
            return NaN;
        }
        const sXx = this._xMoments.variance * n;
        return sXx !== 0 ? this._sXy.value / sXx : NaN;
    }

    /** The OLS intercept a = ȳ − b·x̄ (NaN when the slope is NaN). */
    get intercept(): number {
        return this._yMoments.mean - this.slope * this._xMoments.mean;
    }

    /**
     * The Pearson correlation coefficient r = S_xy / √(S_xx·S_yy),
     * clamped to [−1, 1] to absorb rounding.
     *
     * NaN when n < 2 or either x or y is constant.
     */
    get correlation(): number {
        const n = this._n;
        if (n < 2) {
            return NaN;
        }
        const t = this._xMoments.standardDeviation * this._yMoments.standardDeviation;
        if (t === 0) {
            return NaN;
        }
        const r = this._sXy.value / (t * n);
        return Math.max(-1.0, Math.min(1.0, r));
    }
}
