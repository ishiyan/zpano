import { KleinKbnAccumulator } from './klein-kbn-accumulator';

/**
 * Streaming mean, variance, skewness, and kurtosis via Pébay's central moment
 * update with KBN double-compensated accumulation.
 *
 * Maintains running sums of central moments m2, m3, m4 (as KleinKbnAccumulators)
 * updated in O(1) per sample. Preferred over RawMomentsKleinKbn for forward-only
 * computation (no revert) because it avoids the numerical cancellation inherent
 * in converting raw power sums to central moments.
 *
 * Only the most recent sample can be reverted (LIFO stack, not FIFO queue).
 */
export class CentralMomentsKleinKbn {
    private _n = 0;
    private readonly _m1 = new KleinKbnAccumulator();
    private readonly _m2 = new KleinKbnAccumulator();
    private readonly _m3 = new KleinKbnAccumulator();
    private readonly _m4 = new KleinKbnAccumulator();
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
        this._m1.reset();
        this._m2.reset();
        this._m3.reset();
        this._m4.reset();
    }

    /** Adds a new sample x using Pébay's central moment update formulas. */
    update(x: number): void {
        const nOld = this._n;
        const nNew = nOld + 1;
        this._n = nNew;
        const delta = x - this._m1.value;
        const deltaN = delta / nNew;
        const deltaN2 = deltaN * deltaN;
        const term = delta * deltaN * nOld;

        this._m1.update(deltaN);
        this._m4.update(
            term * deltaN2 * (nNew * nNew - 3 * nNew + 3)
            + 6 * deltaN2 * this._m2.value
            - 4 * deltaN * this._m3.value,
        );
        this._m3.update(term * deltaN * (nNew - 2) - 3 * deltaN * this._m2.value);
        this._m2.update(term);
    }

    /**
     * Removes the most recently added sample x (LIFO).
     * Uses inverse Pébay formulas. KleinKbnAccumulator.set() resets
     * compensation terms to zero, so subsequent updates rebuild compensation.
     */
    revert(x: number): void {
        const nNew = this._n;
        if (nNew === 0) throw new Error('cannot revert below 0');
        const nOld = nNew - 1;
        if (nOld === 0) {
            this._n = 0;
            this._m1.reset();
            this._m2.reset();
            this._m3.reset();
            this._m4.reset();
            return;
        }

        const m1New = this._m1.value;
        const m2New = this._m2.value;
        const m3New = this._m3.value;
        const m4New = this._m4.value;

        const m1Old = (nNew * m1New - x) / nOld;
        const delta = x - m1Old;
        const deltaN = delta / nNew;
        const deltaN2 = deltaN * deltaN;
        const term = delta * deltaN * nOld;

        const m2Old = m2New - term;
        const m3Old = m3New - (term * deltaN * (nNew - 2) - 3 * deltaN * m2Old);
        const m4Old = m4New
            - (term * deltaN2 * (nNew * nNew - 3 * nNew + 3)
                + 6 * deltaN2 * m2Old
                - 4 * deltaN * m3Old);

        this._n = nOld;
        this._m1.set(m1Old);
        this._m2.set(m2Old);
        this._m3.set(m3Old);
        this._m4.set(m4Old);
    }

    /** Returns the current arithmetic mean. */
    get mean(): number {
        return this._m1.value;
    }

    /** Returns the current sample count. */
    get n(): number {
        return this._n;
    }

    /** Returns the current variance. Returns NaN if n <= ddof. */
    get variance(): number {
        const n = this._n - this._ddof;
        return n > 0 ? this._m2.value / n : NaN;
    }

    /** Returns the current standard deviation. Returns NaN if n <= ddof. */
    get standardDeviation(): number {
        const n = this._n - this._ddof;
        return n > 0 ? Math.sqrt(this._m2.value / n) : NaN;
    }

    /** Returns the current skewness. Returns NaN if n < 3 or m2 <= 0. */
    get skewness(): number {
        const N = this._n;
        if (N < 3 || this._m2.value <= 0) return NaN;
        const g1 = Math.sqrt(N) * this._m3.value / Math.pow(this._m2.value, 1.5);
        if (this._bias) return g1;
        return g1 * Math.sqrt(N * (N - 1)) / (N - 2);
    }

    /** Returns the current kurtosis. Returns NaN if n < 4 or m2 <= 0. */
    get kurtosis(): number {
        const N = this._n;
        if (N < 4 || this._m2.value <= 0) return NaN;
        const raw = N * this._m4.value / (this._m2.value * this._m2.value);
        if (!this._bias) {
            const adj = ((N * N - 1) * raw - 3 * (N - 1) * (N - 1)) / ((N - 2) * (N - 3));
            return this._fisher ? adj : adj + 3.0;
        }
        return this._fisher ? raw - 3.0 : raw;
    }
}
