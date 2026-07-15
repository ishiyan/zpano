import { KleinKbnAccumulator } from './klein-kbn-accumulator';

/**
 * Streaming mean, variance, skewness, and kurtosis via raw power sums (x¹..x⁴)
 * with KBN double-compensated accumulation.
 *
 * Accumulates Σx, Σx², Σx³, Σx⁴ using KleinKbnAccumulator for each,
 * plus a separate Welford-style variance tracker (also KBN-compensated).
 * Raw sums are converted to central moments at query time.
 *
 * Supports both LIFO revert and FIFO rolling window (via revert/update cycle).
 */
export class RawMomentsKleinKbn {
    private _n = 0;
    private readonly _x1 = new KleinKbnAccumulator();
    private readonly _x2 = new KleinKbnAccumulator();
    private readonly _x3 = new KleinKbnAccumulator();
    private readonly _x4 = new KleinKbnAccumulator();
    private readonly _mean = new KleinKbnAccumulator();
    private readonly _s = new KleinKbnAccumulator();
    private readonly _ddof: number;
    private readonly _bias: boolean;
    private readonly _fisher: boolean;

    constructor(ddof = 1, bias = true, fisher = true) {
        this._ddof = ddof;
        this._bias = bias;
        this._fisher = fisher;
    }

    /** Clears all accumulated state. */
    reset(): void {
        this._n = 0;
        this._x1.reset();
        this._x2.reset();
        this._x3.reset();
        this._x4.reset();
        this._mean.reset();
        this._s.reset();
    }

    /** Adds a new sample x to the accumulator. */
    update(x: number): void {
        this._n++;
        this._x1.update(x);
        const x2 = x * x;
        this._x2.update(x2);
        const x3 = x2 * x;
        this._x3.update(x3);
        const x4 = x3 * x;
        this._x4.update(x4);

        const n = this._n;
        const delta = x - this._mean.value;
        this._mean.update(delta / n);
        this._s.update(delta * (x - this._mean.value));
    }

    /** Removes a previously added sample x from the accumulator. */
    revert(x: number): void {
        this._n--;
        this._x1.revert(x);
        const x2 = x * x;
        this._x2.revert(x2);
        const x3 = x2 * x;
        this._x3.revert(x3);
        const x4 = x3 * x;
        this._x4.revert(x4);

        const delta = x - this._mean.value;
        const n = this._n;
        this._mean.revert(delta / n);
        this._s.revert(delta * (x - this._mean.value));
    }

    /** Returns the current arithmetic mean. */
    get mean(): number {
        return this._mean.value;
    }

    /** Returns the current sample count. */
    get n(): number {
        return this._n;
    }

    /** Returns the current variance. Returns NaN if n <= ddof. */
    get variance(): number {
        const n = this._n - this._ddof;
        if (n <= 0) return NaN;
        const sv = this._s.value;
        if (sv < 0) {
            this._s.reset();
            return NaN;
        }
        return sv / n;
    }

    /** Returns the current standard deviation. Returns NaN if n <= ddof. */
    get standardDeviation(): number {
        const n = this._n - this._ddof;
        if (n <= 0) return NaN;
        return Math.sqrt(this._s.value / n);
    }

    /** Returns the current skewness. Returns NaN if n < 3. */
    get skewness(): number {
        const N = this._n;
        if (N < 3) return NaN;
        const a = this._x1.value / N;
        const b = this._x2.value / N - a * a;
        if (b <= 1e-14) return NaN;
        const r = Math.sqrt(b);
        const c = this._x3.value / N - a * a * a - 3 * a * b;
        const g1 = c / (r * r * r);
        if (this._bias) return g1;
        return g1 * Math.sqrt(N * (N - 1)) / (N - 2);
    }

    /** Returns the current kurtosis. Returns NaN if n < 4. */
    get kurtosis(): number {
        const N = this._n;
        if (N < 4) return NaN;
        const a = this._x1.value / N;
        let r = a * a;
        const b = this._x2.value / N - r;
        if (b <= 1e-14) return NaN;
        r *= a;
        const c = this._x3.value / N - r - 3 * a * b;
        r *= a;
        const d = this._x4.value / N - r - 6 * b * a * a - 4 * c * a;
        const raw = d / (b * b);
        if (!this._bias) {
            const adj = ((N * N - 1) * raw - 3 * (N - 1) * (N - 1)) / ((N - 2) * (N - 3));
            return this._fisher ? adj : adj + 3.0;
        }
        return this._fisher ? raw - 3.0 : raw;
    }
}
