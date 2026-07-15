import { KleinKbnAccumulator } from './klein-kbn-accumulator';
import { RawMomentsKleinKbn } from './raw-moments-klein-kbn';

/**
 * Streaming simple linear regression (y = slope * x + intercept) with
 * KBN-compensated accumulation.
 *
 * Internally uses two RawMomentsKleinKbn (ddof=0) for x and y moments,
 * and a KleinKbnAccumulator for the cross-product S_xy.
 *
 * Supports both LIFO revert and FIFO rolling window via the revert/update cycle.
 */
export class LinearRegressionKleinKbn {
    private _n = 0;
    private readonly _xMoments = new RawMomentsKleinKbn(0);
    private readonly _yMoments = new RawMomentsKleinKbn(0);
    private readonly _sXY = new KleinKbnAccumulator();

    /** Clears all accumulated state. */
    reset(): void {
        this._n = 0;
        this._xMoments.reset();
        this._yMoments.reset();
        this._sXY.reset();
    }

    /** Adds a new (x, y) observation. */
    update(x: number, y: number): void {
        const nOld = this._n;
        this._n++;
        const term = (this._xMoments.mean - x) * (this._yMoments.mean - y) * nOld / (nOld + 1);
        this._sXY.update(term);
        this._xMoments.update(x);
        this._yMoments.update(y);
    }

    /** Removes a previously added (x, y) observation. */
    revert(x: number, y: number): void {
        if (this._n === 0) return;
        if (this._n === 1) {
            this.reset();
            return;
        }
        this._xMoments.revert(x);
        this._yMoments.revert(y);
        const n = this._n - 1;
        const term = (this._xMoments.mean - x) * (this._yMoments.mean - y) * n / (n + 1);
        this._sXY.revert(term);
        this._n = n;
    }

    /** Returns the current sample count. */
    get n(): number {
        return this._n;
    }

    /** Returns the current slope coefficient. Returns NaN if n < 2 or S_xx == 0. */
    get slope(): number {
        const n = this._n;
        if (n < 2) return NaN;
        const sxx = this._xMoments.variance * n;
        if (sxx === 0) return NaN;
        return this._sXY.value / sxx;
    }

    /** Returns the current intercept coefficient. Returns NaN if n < 2. */
    get intercept(): number {
        return this._yMoments.mean - this.slope * this._xMoments.mean;
    }

    /**
     * Returns the current Pearson correlation coefficient.
     * Returns NaN if n < 2 or either standard deviation is zero.
     */
    get correlation(): number {
        const n = this._n;
        if (n < 2) return NaN;
        const t = this._xMoments.standardDeviation * this._yMoments.standardDeviation;
        if (t === 0) return NaN;
        return this._sXY.value / (t * n);
    }
}
