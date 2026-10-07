import { KleinKbnAccumulator } from './klein-kbn-accumulator';

// ########################################################
// Raw moments with Klein KBN (Kahan-Babuška-Neumaier)
// compensated summation for improved numerical stability.
//
// References:
//   https://github.com/kuiperzone/Compensated-Accumulators
//   https://en.wikipedia.org/wiki/Algorithms_for_calculating_variance
// ########################################################

/**
 * Relative threshold below which the population variance μ₂, computed
 * from raw power sums as Σx²/n − (Σx/n)², is considered to be lost
 * in rounding error (i.e. indistinguishable from zero).
 */
const CANCELLATION_EPSILON = 1e-14;

function validateDdof(ddof: number): void {
    if (typeof ddof !== 'number' || !Number.isInteger(ddof) || ddof < 0) {
        throw new Error('ddof must be a nonnegative integer');
    }
}

/**
 * Streaming mean, variance, skewness, kurtosis via raw power sums (x¹..x⁴)
 * with Klein KBN (Kahan-Babuška-Neumaier) compensated accumulation.
 *
 * Accumulates Σx, Σx², Σx³, Σx⁴ using a KleinKbnAccumulator for each,
 * plus a separate Welford mean/variance tracker (also KBN-compensated).
 * Mean and variance come from the Welford tracker; skewness and kurtosis
 * are converted from the raw power sums at query time.
 *
 * Supports removal of any previously added sample with revert(), not
 * only the most recent one, because both the power sums and Welford's
 * mean/M₂ are symmetric functions of the samples.  This makes the class
 * suitable for FIFO rolling windows (update the new sample, revert the
 * oldest one).
 *
 * Accuracy caveat: converting raw power sums to central moments suffers
 * catastrophic cancellation when the mean is large compared to the spread
 * (e.g. prices rather than returns).  For such data skewness and kurtosis
 * lose precision.  Use CentralMomentsKleinKbn for such data; it supports
 * FIFO removal, but repeated removals clear compensation and can accumulate
 * rounding error.
 *
 * Notation:
 *
 * μₖ' = Σxᵏ / n are the raw moments (properties x1..x4) and μₖ are the
 * population central moments derived from them:
 *
 *     μ₂ = μ₂' − μ₁'²
 *     μ₃ = μ₃' − μ₁'³ − 3·μ₁'·μ₂
 *     μ₄ = μ₄' − μ₁'⁴ − 6·μ₁'²·μ₂ − 4·μ₁'·μ₃
 *
 * g₁ = μ₃ / μ₂^1.5 is the population skewness and β₂ = μ₄ / μ₂² is the
 * population (Pearson) kurtosis.  All skewness and kurtosis properties
 * are NaN when n < 2 or μ₂ is zero relative to μ₂' (constant data).
 *
 * Parameters:
 * - `ddof` (nonnegative integer, default 1): delta degrees of freedom for
 *   variance and standardDeviation. variance = Σ(x - x̄)² / (n - ddof).
 *   ddof=0 gives population, ddof=1 gives sample.
 * - `bias` (default true): selects the skewness and kurtosis properties,
 *   see below.
 * - `fisher` (default true): selects the kurtosis property, see below.
 *
 * The `skewness` and `kurtosis` properties dispatch as follows
 * (matching scipy.stats.skew and scipy.stats.kurtosis):
 *
 * | bias  | fisher | skewness            | kurtosis                    |
 * |-------|--------|---------------------|-----------------------------|
 * | true  | true   | skewnessMoment, g₁  | kurtosisExcess, β₂ − 3      |
 * | true  | false  | skewnessMoment, g₁  | kurtosisMoment, β₂          |
 * | false | true   | skewnessFisher, G₁  | kurtosisSampleExcess, G₂    |
 * | false | false  | skewnessFisher, G₁  | kurtosisSample, G₂ + 3      |
 *
 * The remaining variants (skewnessSample, kurtosisSampleCorrected),
 * which match the R PerformanceAnalytics package, are available as
 * separate properties.
 */
export class RawMomentsKleinKbn {
    private _n = 0;
    private readonly _x1 = new KleinKbnAccumulator();
    private readonly _x2 = new KleinKbnAccumulator();
    private readonly _x3 = new KleinKbnAccumulator();
    private readonly _x4 = new KleinKbnAccumulator();
    private _ddof = 1;
    /** Selects the skewness and kurtosis properties. */
    bias: boolean;
    /** Selects the kurtosis property. */
    fisher: boolean;
    // Welford's mean and sum of squared deviations Σ(x - x̄)².
    private readonly _mean = new KleinKbnAccumulator();
    private readonly _s = new KleinKbnAccumulator();

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
        this._x1.reset();
        this._x2.reset();
        this._x3.reset();
        this._x4.reset();
        this._mean.reset();
        this._s.reset();
    }

    /** Adds a sample x. */
    update(x: number): void {
        this._n += 1;
        this._x1.update(x);
        const x2 = x * x;
        this._x2.update(x2);
        const x3 = x2 * x;
        this._x3.update(x3);
        const x4 = x3 * x;
        this._x4.update(x4);
        // Welford: mean += (x - mean_old) / n;  S += (x - mean_old)·(x - mean_new)
        const delta = x - this._mean.value;
        this._mean.update(delta / this._n);
        this._s.update(delta * (x - this._mean.value));
    }

    /**
     * Removes a previously added sample x (any sample, not only the most
     * recent one).  Reverting a value that was never added corrupts the
     * state.
     *
     * Throws an Error if there are no samples.
     */
    revert(x: number): void {
        if (this._n <= 0) {
            throw new Error('Cannot revert from an empty accumulator');
        }
        if (this._n === 1) {
            this.reset();
            return;
        }
        this._n -= 1;
        this._x1.revert(x);
        const x2 = x * x;
        this._x2.revert(x2);
        const x3 = x2 * x;
        this._x3.revert(x3);
        const x4 = x3 * x;
        this._x4.revert(x4);
        // Inverse Welford: mean_old = mean_new - (x - mean_new) / (n - 1);
        // S -= (x - mean_new)·(x - mean_old)
        const delta = x - this._mean.value;
        this._mean.revert(delta / this._n);
        this._s.revert(delta * (x - this._mean.value));
    }

    private _variance(ddof: number): number {
        // A slightly negative S caused by rounding after revert() is clamped to zero.
        const d = this._n - ddof;
        if (d <= 0) {
            return NaN;
        }
        return Math.max(this._s.value, 0.0) / d;
    }

    private _standardDeviation(ddof: number): number {
        const v = this._variance(ddof);
        return Number.isNaN(v) ? v : Math.sqrt(v);
    }

    /** The arithmetic mean (0.0 when empty). */
    get mean(): number {
        return this._mean.value;
    }

    /** The variance Σ(x - x̄)² / (n - ddof), NaN when n ≤ ddof. */
    get variance(): number {
        return this._variance(this._ddof);
    }

    /** The population variance Σ(x - x̄)² / n, regardless of ddof. */
    get varianceDdof0(): number {
        return this._variance(0);
    }

    /** The sample variance Σ(x - x̄)² / (n - 1), regardless of ddof. */
    get varianceDdof1(): number {
        return this._variance(1);
    }

    /** The square root of variance. */
    get standardDeviation(): number {
        return this._standardDeviation(this._ddof);
    }

    /** The population standard deviation, regardless of ddof. */
    get standardDeviationDdof0(): number {
        return this._standardDeviation(0);
    }

    /** The sample standard deviation, regardless of ddof. */
    get standardDeviationDdof1(): number {
        return this._standardDeviation(1);
    }

    /**
     * Converts the raw power sums to population central moments
     * (μ₂, μ₃, μ₄), see the class Notation.
     *
     * Returns null when n < 2 or μ₂ is lost in rounding error.
     */
    private _centralMoments(): [number, number, number] | null {
        const n = this._n;
        if (n < 2) {
            return null;
        }
        const mu1 = this._x1.value / n;
        let r = mu1 * mu1;
        const meanX2 = this._x2.value / n;
        const mu2 = meanX2 - r;
        if (mu2 <= CANCELLATION_EPSILON * meanX2) {
            return null;
        }
        r *= mu1;
        const mu3 = this._x3.value / n - r - 3 * mu1 * mu2;
        r *= mu1;
        const mu4 = this._x4.value / n - r - 6 * mu2 * mu1 * mu1 - 4 * mu3 * mu1;
        return [mu2, mu3, mu4];
    }

    /** The population skewness g₁ = μ₃ / μ₂^1.5. */
    private get _g1(): number {
        const cm = this._centralMoments();
        if (cm === null) {
            return NaN;
        }
        const [mu2, mu3] = cm;
        return mu3 / (mu2 * Math.sqrt(mu2));
    }

    /** The population (Pearson) kurtosis β₂ = μ₄ / μ₂². */
    private get _b2(): number {
        const cm = this._centralMoments();
        if (cm === null) {
            return NaN;
        }
        const [mu2, , mu4] = cm;
        return mu4 / (mu2 * mu2);
    }

    /**
     * The 'moment' (biased, population) skewness, requires n ≥ 2:
     *
     *     g₁ = μ₃ / μ₂^1.5
     *
     * Matches scipy.stats.skew(bias=True) and PerformanceAnalytics
     * skewness(method="moment").
     */
    get skewnessMoment(): number {
        return this._g1;
    }

    /**
     * The 'fisher' (bias-adjusted Fisher-Pearson) skewness, requires n ≥ 3:
     *
     *     G₁ = g₁ · √(n(n−1)) / (n−2)
     *
     * Matches scipy.stats.skew(bias=False) and PerformanceAnalytics
     * skewness(method="fisher").
     */
    get skewnessFisher(): number {
        const g1 = this._g1;
        if (Number.isNaN(g1)) {
            return NaN;
        }
        const n = this._n;
        return n < 3 ? NaN : g1 * Math.sqrt(n * (n - 1)) / (n - 2);
    }

    /**
     * The 'sample' skewness, requires n ≥ 3:
     *
     *     g₁ · n² / ((n−1)(n−2))
     *
     * Matches PerformanceAnalytics skewness(method="sample").
     * Doesn't depend on the bias parameter.
     */
    get skewnessSample(): number {
        const g1 = this._g1;
        if (Number.isNaN(g1)) {
            return NaN;
        }
        const n = this._n;
        return n < 3 ? NaN : g1 * (n * n) / ((n - 1) * (n - 2));
    }

    /**
     * The skewness selected by the bias parameter:
     *
     * - bias=true:  skewnessMoment, g₁
     * - bias=false: skewnessFisher, G₁ = g₁ · √(n(n−1)) / (n−2)
     *
     * The third variant, skewnessSample, doesn't depend on bias.
     */
    get skewness(): number {
        return this.bias ? this.skewnessMoment : this.skewnessFisher;
    }

    /**
     * The 'moment' (biased, population) Pearson kurtosis, requires n ≥ 2:
     *
     *     β₂ = μ₄ / μ₂²
     *
     * Matches scipy.stats.kurtosis(bias=True, fisher=False) and
     * PerformanceAnalytics kurtosis(method="moment").
     */
    get kurtosisMoment(): number {
        return this._b2;
    }

    /**
     * The 'excess' (biased, population) excess kurtosis, requires n ≥ 2:
     *
     *     β₂ − 3
     *
     * Matches scipy.stats.kurtosis(bias=True, fisher=True) and
     * PerformanceAnalytics kurtosis(method="excess").
     */
    get kurtosisExcess(): number {
        const b2 = this._b2;
        if (Number.isNaN(b2)) {
            return NaN;
        }
        return b2 - 3;
    }

    /**
     * The 'sample excess' (unbiased) excess kurtosis, requires n ≥ 4:
     *
     *     G₂ = ((n²−1)·β₂ − 3(n−1)²) / ((n−2)(n−3))
     *
     * Matches scipy.stats.kurtosis(bias=False, fisher=True) and
     * PerformanceAnalytics kurtosis(method="sample_excess").
     */
    get kurtosisSampleExcess(): number {
        const b2 = this._b2;
        if (Number.isNaN(b2)) {
            return NaN;
        }
        const n = this._n;
        if (n <= 3) {
            return NaN;
        }
        return ((n * n - 1) * b2 - 3 * (n - 1) ** 2) / ((n - 2) * (n - 3));
    }

    /**
     * The 'sample' (unbiased) Pearson kurtosis, requires n ≥ 4, calculated
     * as the 'sample excess' kurtosis plus 3:
     *
     *     G₂ + 3 = ((n²−1)·β₂ − 3(n−1)²) / ((n−2)(n−3)) + 3
     *
     * Matches scipy.stats.kurtosis(bias=False, fisher=False).
     *
     * The PerformanceAnalytics kurtosis(method="sample") variant is
     * available as kurtosisSampleCorrected; it is larger than this one
     * by (9n−15) / ((n−2)(n−3)), approximately 0.44 for n = 24.
     */
    get kurtosisSample(): number {
        const b2 = this._b2;
        if (Number.isNaN(b2)) {
            return NaN;
        }
        const n = this._n;
        if (n <= 3) {
            return NaN;
        }
        return ((n * n - 1) * b2 - 3 * (n - 1) ** 2) / ((n - 2) * (n - 3)) + 3;
    }

    /**
     * The PerformanceAnalytics 'sample' (unbiased) Pearson kurtosis,
     * requires n ≥ 4:
     *
     *     (n²−1)·β₂ / ((n−2)(n−3))
     *
     * Matches PerformanceAnalytics kurtosis(method="sample").
     * Doesn't depend on the bias and fisher parameters.
     *
     * It differs from kurtosisSample (G₂ + 3) by (9n−15) / ((n−2)(n−3)),
     * approximately 0.44 for n = 24.
     */
    get kurtosisSampleCorrected(): number {
        const b2 = this._b2;
        if (Number.isNaN(b2)) {
            return NaN;
        }
        const n = this._n;
        if (n <= 3) {
            return NaN;
        }
        return b2 * (n * n - 1) / ((n - 2) * (n - 3));
    }

    /**
     * The kurtosis selected by the bias and fisher parameters:
     *
     * - bias=true,  fisher=true:  kurtosisExcess, β₂ − 3
     * - bias=true,  fisher=false: kurtosisMoment, β₂
     * - bias=false, fisher=true:  kurtosisSampleExcess, G₂
     * - bias=false, fisher=false: kurtosisSample, G₂ + 3
     */
    get kurtosis(): number {
        if (this.bias) {
            return this.fisher ? this.kurtosisExcess : this.kurtosisMoment;
        }
        return this.fisher ? this.kurtosisSampleExcess : this.kurtosisSample;
    }

    /** The sum Σx. */
    get x1Sum(): number {
        return this._x1.value;
    }

    /** The sum Σx². */
    get x2Sum(): number {
        return this._x2.value;
    }

    /** The sum Σx³. */
    get x3Sum(): number {
        return this._x3.value;
    }

    /** The sum Σx⁴. */
    get x4Sum(): number {
        return this._x4.value;
    }

    /** The first raw moment Σx / n (NaN when empty). */
    get x1(): number {
        return this._n > 0 ? this._x1.value / this._n : NaN;
    }

    /** The second raw moment Σx² / n (NaN when empty). */
    get x2(): number {
        return this._n > 0 ? this._x2.value / this._n : NaN;
    }

    /** The third raw moment Σx³ / n (NaN when empty). */
    get x3(): number {
        return this._n > 0 ? this._x3.value / this._n : NaN;
    }

    /** The fourth raw moment Σx⁴ / n (NaN when empty). */
    get x4(): number {
        return this._n > 0 ? this._x4.value / this._n : NaN;
    }

    /** The number of samples. */
    get n(): number {
        return this._n;
    }
}
