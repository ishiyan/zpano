import { KleinKbnAccumulator } from './klein-kbn-accumulator';

// ########################################################
// Central moments with Klein KBN (Kahan-Babuška-Neumaier)
// compensated summation for improved numerical stability.
//
// References:
//   P. Pébay, "Formulas for Robust, One-Pass Parallel Computation
//     of Covariances and Arbitrary-Order Statistical Moments",
//     Sandia Report SAND2008-6212 (2008).
//   https://www.johndcook.com/skewness_kurtosis.html
//   https://github.com/kuiperzone/Compensated-Accumulators
// ########################################################

function validateDdof(ddof: number): void {
    if (typeof ddof !== 'number' || !Number.isInteger(ddof) || ddof < 0) {
        throw new Error('ddof must be a nonnegative integer');
    }
}

/**
 * Streaming mean, variance, skewness, kurtosis via Pébay's central moment
 * update with Klein KBN (Kahan-Babuška-Neumaier) compensated accumulation.
 *
 * Maintains the mean M₁ and the sums of central powers
 *
 *     M₂ = Σ(x - x̄)²,   M₃ = Σ(x - x̄)³,   M₄ = Σ(x - x̄)⁴
 *
 * (each as a KleinKbnAccumulator), updated in O(1) per sample.
 * The population central moments are μₖ = Mₖ / n.
 *
 * Avoids the catastrophic cancellation inherent in converting raw power
 * sums Σxᵏ to central moments.  This matters for data with a large mean
 * relative to its spread.  Inverse updates can remove any previously
 * added sample, including the oldest sample in a FIFO rolling window.
 * Reversion clears the compensation terms, so repeated removals can
 * accumulate rounding error.
 *
 * Parameters:
 * - `ddof` (nonnegative integer, default 1): delta degrees of freedom for
 *   variance. variance = M₂ / (n - ddof).  ddof=0 gives population,
 *   ddof=1 gives sample.
 * - `bias` (default true): if true, return the biased (population) skewness
 *   and kurtosis. If false, apply the bias corrections (see Notes).
 * - `fisher` (default true): if true, return excess kurtosis (subtract 3 so
 *   Gaussian→0). If false, return raw (Pearson) kurtosis (Gaussian→3).
 *   Applied after the bias correction when bias=false.
 *
 * Notes:
 *
 * The results match scipy.stats.skew(bias=...) and
 * scipy.stats.kurtosis(bias=..., fisher=...).
 *
 * Skewness (bias=true), requires n ≥ 2:
 *     g₁ = μ₃ / μ₂^1.5 = √n · M₃ / M₂^1.5
 *
 * Skewness (bias=false), requires n ≥ 3:
 *     G₁ = g₁ · √(n·(n-1)) / (n-2)
 *
 * Kurtosis (bias=true), requires n ≥ 2:
 *     β₂ = μ₄ / μ₂² = n · M₄ / M₂²
 *     fisher=true:  g₂ = β₂ - 3
 *     fisher=false: β₂
 *
 * Kurtosis (bias=false), requires n ≥ 4:
 *     G₂ = ((n²-1) · β₂  -  3·(n-1)²) / ((n-2)·(n-3))
 *     fisher=true:  G₂
 *     fisher=false: G₂ + 3
 *
 * Skewness and kurtosis are NaN when M₂ = 0 (constant data).
 */
export class CentralMomentsKleinKbn {
    private _ddof = 1;
    /** Selects biased (true) or bias-corrected (false) skewness and kurtosis. */
    bias: boolean;
    /** Selects excess (true) or Pearson (false) kurtosis. */
    fisher: boolean;
    private _n = 0;
    private readonly _m1 = new KleinKbnAccumulator();
    private readonly _m2 = new KleinKbnAccumulator();
    private readonly _m3 = new KleinKbnAccumulator();
    private readonly _m4 = new KleinKbnAccumulator();

    constructor(ddof = 1, bias = true, fisher = true) {
        this.ddof = ddof;
        this.bias = bias;
        this.fisher = fisher;
    }

    /** Delta degrees of freedom for variance (nonnegative integer). */
    get ddof(): number {
        return this._ddof;
    }

    set ddof(value: number) {
        validateDdof(value);
        this._ddof = value;
    }

    /** Clears all accumulated state. */
    reset(): void {
        this._n = 0;
        this._m1.reset();
        this._m2.reset();
        this._m3.reset();
        this._m4.reset();
    }

    /**
     * Adds a sample x using Pébay's update (n = count after adding x):
     *
     *     δ    = x − M₁
     *     δₙ   = δ / n
     *     term = δ · δₙ · (n − 1)
     *
     *     M₁ += δₙ
     *     M₄ += term·δₙ²·(n²−3n+3) + 6·δₙ²·M₂ − 4·δₙ·M₃
     *     M₃ += term·δₙ·(n−2) − 3·δₙ·M₂
     *     M₂ += term
     *
     * M₄ and M₃ are updated before M₂ and M₃ respectively, because
     * they use the values from before x was added.
     */
    update(x: number): void {
        const nOld = this._n;
        const nNew = nOld + 1;
        this._n = nNew;
        const delta = x - this._m1.value;
        const deltaN = delta / nNew;
        const deltaN2 = deltaN * deltaN;
        const term = delta * deltaN * nOld;
        const m2 = this._m2.value;
        const m3 = this._m3.value;
        this._m1.update(deltaN);
        this._m4.update(term * deltaN2 * (nNew * nNew - 3 * nNew + 3) + 6 * deltaN2 * m2 - 4 * deltaN * m3);
        this._m3.update(term * deltaN * (nNew - 2) - 3 * deltaN * m2);
        this._m2.update(term);
    }

    /**
     * Removes a previously added sample x, regardless of insertion order.
     * Reverting a value that was never added corrupts the state.
     *
     * The restored M₁–M₄ are written with KleinKbnAccumulator.set(), which
     * clears their compensation terms.  Subsequent updates rebuild the
     * compensation from the restored values.  Repeated reverts can
     * accumulate rounding error, especially for large-offset data.
     *
     * Inverse formulas (where nₙ = count before revert, nₒ = nₙ − 1):
     *
     *     M₁_old = (nₙ · M₁_new − x) / nₒ            [mean undo]
     *     δ      = x − M₁_old
     *     δₙ     = δ / nₙ
     *     term   = δ · δₙ · nₒ
     *
     *     M₂_old = M₂_new − term
     *     M₃_old = M₃_new − (term·δₙ·(nₙ−2) − 3·δₙ·M₂_old)
     *     M₄_old = M₄_new − (term·δₙ²·(nₙ²−3nₙ+3)
     *                         + 6·δₙ²·M₂_old − 4·δₙ·M₃_old)
     *
     * Throws an Error if there are no samples.
     */
    revert(x: number): void {
        const nNew = this._n;
        if (nNew === 0) {
            throw new Error('Cannot revert from an empty accumulator');
        }
        const nOld = nNew - 1;
        if (nOld === 0) {
            this.reset();
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
        const m4Old = m4New - (term * deltaN2 * (nNew * nNew - 3 * nNew + 3) + 6 * deltaN2 * m2Old - 4 * deltaN * m3Old);

        this._n = nOld;
        this._m1.set(m1Old);
        this._m2.set(m2Old);
        this._m3.set(m3Old);
        this._m4.set(m4Old);
    }

    /** The number of samples. */
    get n(): number {
        return this._n;
    }

    /** The arithmetic mean (0.0 when empty). */
    get mean(): number {
        return this._m1.value;
    }

    /**
     * The variance M₂ / (n - ddof), NaN when n ≤ ddof.
     *
     * A slightly negative M₂ caused by rounding after revert()
     * is clamped to zero.
     */
    get variance(): number {
        const d = this._n - this._ddof;
        if (d <= 0) {
            return NaN;
        }
        return Math.max(this._m2.value, 0.0) / d;
    }

    /** The square root of the variance, NaN when n ≤ ddof. */
    get standardDeviation(): number {
        const v = this.variance;
        return Number.isNaN(v) ? v : Math.sqrt(v);
    }

    /** The skewness g₁ (bias=true) or G₁ (bias=false); see the class Notes. */
    get skewness(): number {
        const n = this._n;
        const m2 = this._m2.value;
        if (n < 2 || m2 <= 0) {
            return NaN;
        }
        const g1 = Math.sqrt(n) * this._m3.value / (m2 * Math.sqrt(m2));
        if (this.bias) {
            return g1;
        }
        if (n < 3) {
            return NaN;
        }
        return g1 * Math.sqrt(n * (n - 1)) / (n - 2);
    }

    /** The kurtosis selected by bias and fisher; see the class Notes. */
    get kurtosis(): number {
        const n = this._n;
        const m2 = this._m2.value;
        if (n < 2 || m2 <= 0) {
            return NaN;
        }
        const b2 = n * this._m4.value / (m2 * m2);
        if (this.bias) {
            return this.fisher ? b2 - 3.0 : b2;
        }
        if (n < 4) {
            return NaN;
        }
        const g2 = ((n * n - 1) * b2 - 3 * (n - 1) ** 2) / ((n - 2) * (n - 3));
        return this.fisher ? g2 : g2 + 3.0;
    }
}
