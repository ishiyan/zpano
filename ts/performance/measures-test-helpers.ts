/**
 * Test-only helpers shared by the `Measures` spec files: the Bacon data
 * sets, ports of the Python test assertion helpers and stream runners, and
 * ports of the Python standard-library functions the tests rely on
 * (`random.Random`, `sum`, `math.fsum`, `math.prod`, `math.isclose`,
 * `statistics.fmean/pstdev/stdev`).
 *
 * Not a spec file itself (not matched by `*.spec.js`).
 */

import { Measures } from './measures';

// ---------------------------------------------------------------------------
// Data
// ---------------------------------------------------------------------------

/**
 * 'Portfolio bacon' dataset from the PerformanceAnalytics R package: Bacon,
 * Practical portfolio performance measurement and attribution, 2nd ed., 2008,
 * p. 65 (portfolio) and p. 66 (benchmark).
 */
export const baconPortfolioReturns: readonly number[] = [
    0.003, 0.026, 0.011, -0.010,
    0.015, 0.025, 0.016, 0.067,
    -0.014, 0.040, -0.005, 0.081,
    0.040, -0.037, -0.061, 0.017,
    -0.049, -0.022, 0.070, 0.058,
    -0.065, 0.024, -0.005, -0.009,
];

export const baconBenchmarkReturns: readonly number[] = [
    0.002, 0.025, 0.018, -0.011,
    0.014, 0.018, 0.014, 0.065,
    -0.015, 0.042, -0.006, 0.083,
    0.039, -0.038, -0.062, 0.015,
    -0.048, 0.021, 0.060, 0.056,
    -0.067, 0.019, -0.003, 0.000,
];

export const baconPortfolioLen = baconPortfolioReturns.length;

/** Extended Bacon 2023 (3rd edition) portfolio data. */
export const bacon2023PortfolioReturns: readonly number[] = [
    0.003, 0.026, 0.011, -0.009, 0.014, 0.024, 0.015, 0.066, -0.014, 0.039,
    -0.005, 0.081, 0.040, -0.037, -0.061, 0.014, -0.049, -0.021, 0.062, 0.058,
    -0.064, 0.017, -0.004, -0.002, -0.021, 0.011, 0.047, 0.024, 0.033, -0.007,
    0.047, 0.006, 0.010, -0.002, 0.034, 0.010];

export const bacon2023DrawdownContinuous: readonly number[] = [
    0, 0, 0, -0.0090, 0, 0, 0, 0, -0.0140, 0,
    -0.0050, 0, 0, 0, -0.0960, 0, 0, -0.0690, 0, 0,
    -0.0640, 0, 0, 0, -0.0270, 0, 0, 0, 0, -0.0070,
    0, 0, 0, -0.0020, 0, 0];

export const bacon2023DrawdownContinuousWithoutZeroes: readonly number[] = [
    -0.0090, -0.0140, -0.0050, -0.0960, -0.0690,
    -0.0640, -0.0270, -0.0070, -0.0020];

export const bacon2023DrawdownFromPeak: readonly number[] = [
    0, 0, 0, -0.0090, 0, 0, 0, 0, -0.0140, 0,
    -0.0050, 0, 0, -0.0370, -0.0957, -0.0831, -0.1280, -0.1463, -0.0934, -0.0408,
    -0.1022, -0.0869, -0.0906, -0.0924, -0.1115, -0.1017, -0.0595, -0.0369, -0.0051, -0.0121,
    0, 0, 0, -0.0020, 0, 0];

export const bacon2023PortfolioLen = bacon2023PortfolioReturns.length;

export const SQRT2 = 1.4142135623730950488016887242097;

// ---------------------------------------------------------------------------
// Python standard library ports
// ---------------------------------------------------------------------------

/** Python's built-in `sum()` of floats (Neumaier-compensated since 3.12). */
export function pySum(xs: Iterable<number>): number {
    let s = 0.0;
    let c = 0.0;
    for (const x of xs) {
        const t = s + x;
        if (Math.abs(s) >= Math.abs(x)) {
            c += (s - t) + x;
        } else {
            c += (x - t) + s;
        }
        s = t;
    }
    return (c !== 0 && Number.isFinite(c)) ? s + c : s;
}

/**
 * Correctly rounded sum of finite numbers, a port of CPython's math.fsum
 * (Shewchuk's partials with the final half-even correction).
 */
export function fsum(values: Iterable<number>): number {
    const partials: number[] = [];
    for (let x of values) {
        let i = 0;
        for (let j = 0; j < partials.length; j++) {
            let y = partials[j];
            if (Math.abs(x) < Math.abs(y)) {
                const t = x;
                x = y;
                y = t;
            }
            const hi = x + y;
            const lo = y - (hi - x);
            if (lo !== 0.0) {
                partials[i++] = lo;
            }
            x = hi;
        }
        partials.length = i;
        partials.push(x);
    }

    let n = partials.length;
    let hi = 0.0;
    if (n > 0) {
        let lo = 0.0;
        hi = partials[--n];
        while (n > 0) {
            const x = hi;
            const y = partials[--n];
            hi = x + y;
            const yr = hi - x;
            lo = y - yr;
            if (lo !== 0.0) {
                break;
            }
        }
        if (n > 0 && ((lo < 0.0 && partials[n - 1] < 0.0) || (lo > 0.0 && partials[n - 1] > 0.0))) {
            const y = lo * 2.0;
            const x = hi + y;
            const yr = x - hi;
            if (y === yr) {
                hi = x;
            }
        }
    }
    return hi;
}

/** Python's `math.prod()`: left-to-right product starting at 1. */
export function prod(xs: Iterable<number>): number {
    let p = 1;
    for (const x of xs) {
        p *= x;
    }
    return p;
}

/** Python's `math.isclose()`. */
export function isClose(a: number, b: number, relTol: number = 1e-9, absTol: number = 0.0): boolean {
    if (a === b) {
        return true;
    }
    if (!Number.isFinite(a) || !Number.isFinite(b)) {
        return false;
    }
    const diff = Math.abs(b - a);
    return diff <= Math.abs(relTol * b) || diff <= Math.abs(relTol * a) || diff <= absTol;
}

/** Python's `statistics.fmean()`: `fsum(data) / n`. */
export function fmean(xs: readonly number[]): number {
    return fsum(xs) / xs.length;
}

/**
 * Sum of squared deviations from the mean, computed with exactly rounded
 * sums. Python's `statistics` computes it with exact rational arithmetic;
 * this agrees to within a few ulps, far inside the tolerances of the tests
 * that use it.
 */
function sumSquaredDeviations(xs: readonly number[]): number {
    const n = xs.length;
    const mean = fsum(xs) / n;
    const dev = xs.map(x => x - mean);
    // Correction term for the rounding error of the mean (two-pass algorithm).
    const sumDev = fsum(dev);
    return fsum(dev.map(d => d * d)) - sumDev * sumDev / n;
}

/** Python's `statistics.pstdev()` (population standard deviation). */
export function pstdev(xs: readonly number[]): number {
    return Math.sqrt(sumSquaredDeviations(xs) / xs.length);
}

/** Python's `statistics.stdev()` (sample standard deviation). */
export function stdev(xs: readonly number[]): number {
    return Math.sqrt(sumSquaredDeviations(xs) / (xs.length - 1));
}

/**
 * Port of Python's random.Random(seed) for integer seeds (MT19937 seeded
 * with init_by_array) with Python's random() and gauss() algorithms, so the
 * generated data match the Python tests exactly.
 */
export class SeededRandom {
    private readonly mt = new Uint32Array(624);
    private mti = 625;
    private gaussNext: number | null = null;

    constructor(seed: number) {
        const mt = this.mt;
        mt[0] = 19650218;
        for (let i = 1; i < 624; i++) {
            const p = mt[i - 1] ^ (mt[i - 1] >>> 30);
            mt[i] = (Math.imul(1812433253, p) + i) >>> 0;
        }
        // init_by_array with key = [seed] (seed < 2**32).
        const key = [seed >>> 0];
        let i = 1;
        let j = 0;
        for (let k = Math.max(624, key.length); k > 0; k--) {
            const p = mt[i - 1] ^ (mt[i - 1] >>> 30);
            mt[i] = ((mt[i] ^ Math.imul(p, 1664525)) + key[j] + j) >>> 0;
            i++;
            j++;
            if (i >= 624) {
                mt[0] = mt[623];
                i = 1;
            }
            if (j >= key.length) {
                j = 0;
            }
        }
        for (let k = 623; k > 0; k--) {
            const p = mt[i - 1] ^ (mt[i - 1] >>> 30);
            mt[i] = ((mt[i] ^ Math.imul(p, 1566083941)) - i) >>> 0;
            i++;
            if (i >= 624) {
                mt[0] = mt[623];
                i = 1;
            }
        }
        mt[0] = 0x80000000;
        this.mti = 624;
    }

    private genrandUint32(): number {
        const mt = this.mt;
        if (this.mti >= 624) {
            let kk = 0;
            let y: number;
            for (; kk < 624 - 397; kk++) {
                y = (mt[kk] & 0x80000000) | (mt[kk + 1] & 0x7fffffff);
                mt[kk] = mt[kk + 397] ^ (y >>> 1) ^ ((y & 1) ? 0x9908b0df : 0);
            }
            for (; kk < 623; kk++) {
                y = (mt[kk] & 0x80000000) | (mt[kk + 1] & 0x7fffffff);
                mt[kk] = mt[kk + (397 - 624)] ^ (y >>> 1) ^ ((y & 1) ? 0x9908b0df : 0);
            }
            y = (mt[623] & 0x80000000) | (mt[0] & 0x7fffffff);
            mt[623] = mt[396] ^ (y >>> 1) ^ ((y & 1) ? 0x9908b0df : 0);
            this.mti = 0;
        }
        let y = mt[this.mti++];
        y ^= y >>> 11;
        y ^= (y << 7) & 0x9d2c5680;
        y ^= (y << 15) & 0xefc60000;
        y ^= y >>> 18;
        return y >>> 0;
    }

    random(): number {
        const a = this.genrandUint32() >>> 5;
        const b = this.genrandUint32() >>> 6;
        return (a * 67108864.0 + b) * (1.0 / 9007199254740992.0);
    }

    gauss(mu: number, sigma: number): number {
        let z = this.gaussNext;
        this.gaussNext = null;
        if (z === null) {
            const x2pi = this.random() * 2 * Math.PI;
            const g2rad = Math.sqrt(-2.0 * Math.log(1.0 - this.random()));
            z = Math.cos(x2pi) * g2rad;
            this.gaussNext = Math.sin(x2pi) * g2rad;
        }
        return mu + z * sigma;
    }
}

// ---------------------------------------------------------------------------
// Assertions (ports of unittest / test_measures.py helpers)
// ---------------------------------------------------------------------------

/** Thrown by the assertion helpers, like Python's AssertionError. */
export class AssertionError extends Error {
    constructor(message: string) {
        super(message);
        this.name = 'AssertionError';
    }
}

function check(condition: boolean, message: string): void {
    if (!condition) {
        throw new AssertionError(message);
    }
}

/**
 * `unittest.TestCase.assertAlmostEqual`: passes if the values are equal, or
 * if `|a - b| <= delta` when `delta` is given, else if
 * `round(|a - b|, places) == 0`.
 */
export function assertAlmostEqual(actual: number, expected: number,
    opts: { places?: number; delta?: number; msg?: string } = {}): void {
    if (actual === expected) {
        return;
    }
    const diff = Math.abs(actual - expected);
    const msg = opts.msg ?? '';
    if (opts.delta !== undefined) {
        check(diff <= opts.delta, `${actual} != ${expected} within ${opts.delta} delta : ${msg}`);
        return;
    }
    const places = opts.places ?? 7;
    // toFixed rounds the exact binary value, like Python's round().
    check(Number(diff.toFixed(places)) === 0, `${actual} != ${expected} within ${places} places : ${msg}`);
}

export interface FloatAssertOptions {
    places?: number;
    delta?: number;
    prefix?: string;
    relTol?: number;
}

/** Port of `assertFloatEqual` from test_measures.py (NaN and ±Infinity aware). */
export function assertFloatEqual(actual: number, expected: number, opts: FloatAssertOptions = {}): void {
    const places = opts.places ?? 15;
    const prefix = opts.prefix ?? '';
    if (Number.isNaN(expected)) {
        check(Number.isNaN(actual), `${prefix}: expected NaN, actual ${actual}`);
    } else if (expected === Infinity) {
        check(actual === Infinity, `${prefix}: expected +Inf, actual ${actual}`);
    } else if (expected === -Infinity) {
        check(actual === -Infinity, `${prefix}: expected -Inf, actual ${actual}`);
    } else {
        check(!Number.isNaN(actual), `${prefix}: expected ${expected}, got ${actual}`);
        check(Number.isFinite(actual), `${prefix}: expected ${expected}, got ${actual}`);
        if (opts.relTol !== undefined) {
            check(isClose(actual, expected, opts.relTol, opts.delta || 0.0),
                `${prefix}: expected ${expected}, got ${actual}`);
        } else if (opts.delta === undefined) {
            assertAlmostEqual(actual, expected, { places, msg: `${prefix}: expected ${expected}, got ${actual}` });
        } else {
            assertAlmostEqual(actual, expected, { delta: opts.delta, msg: `${prefix}: expected ${expected}, got ${actual}` });
        }
    }
}

export interface SeriesAssertOptions extends FloatAssertOptions {
    skip?: number;
}

/** Port of `assertSeriesEqual` from test_measures.py (lengths must match). */
export function assertSeriesEqual(actual: Iterable<number>, expected: Iterable<number>,
    opts: SeriesAssertOptions = {}): void {
    const a = Array.from(actual);
    const e = Array.from(expected);
    const prefix = opts.prefix ?? '';
    const skip = opts.skip ?? 0;
    check(a.length === e.length, `${prefix}: series length: ${a.length} != ${e.length}`);
    for (let i = 0; i < a.length; i++) {
        if (i >= skip) {
            assertFloatEqual(a[i], e[i], { ...opts, prefix: `${prefix} step ${i}` });
        }
    }
}

/** `assertTrue`. */
export function assertTrue(condition: boolean, msg: string = ''): void {
    check(condition === true, `false is not true : ${msg}`);
}

/** `assertFalse`. */
export function assertFalse(condition: boolean, msg: string = ''): void {
    check(condition === false, `true is not false : ${msg}`);
}

/** `assertEqual` for primitives. */
export function assertEqual<T>(actual: T, expected: T, msg: string = ''): void {
    check(actual === expected, `${actual} != ${expected} : ${msg}`);
}

/** `assertRaises(ValueError)`: the callback must throw an Error. */
export function assertThrows(fn: () => unknown, msg: string = ''): void {
    let thrown = false;
    try {
        fn();
    } catch (err) {
        if (err instanceof AssertionError) {
            throw err;
        }
        thrown = err instanceof Error;
    }
    check(thrown, `Error not raised : ${msg}`);
}

/** Fail the spec if `fn` does not throw an AssertionError. */
export function assertAssertionFails(fn: () => void): void {
    let thrown = false;
    try {
        fn();
    } catch (err) {
        thrown = err instanceof AssertionError;
        if (!thrown) {
            throw err;
        }
    }
    check(thrown, 'AssertionError not raised');
}

// ---------------------------------------------------------------------------
// Stream runners
// ---------------------------------------------------------------------------

/** Python `periods_per_annum(daily, monthly)`. */
export function periodsPerAnnum(daily: boolean = false, monthly: boolean = false): number {
    if (daily && monthly) {
        throw new Error('Only one of daily or monthly can be True');
    }
    return daily ? 252 : (monthly ? 12 : 1);
}

export interface StreamOptions {
    daily?: boolean;
    monthly?: boolean;
    annualRiskFreeRate?: number;
    annualTargetReturn?: number;
    returns?: readonly number[];
    benchmarkReturns?: readonly number[];
    rollingWindowSize?: number;
    start?: number;
}

/** Python `run_stream_callback`. */
export function runStreamCallback<T>(callback: (m: Measures) => T, opts: StreamOptions = {}): T[] {
    const returns = opts.returns ?? baconPortfolioReturns;
    const benchmarkReturns = opts.benchmarkReturns ?? baconBenchmarkReturns;
    const measures = new Measures(
        periodsPerAnnum(opts.daily ?? false, opts.monthly ?? false),
        opts.annualRiskFreeRate ?? 0,
        opts.annualTargetReturn ?? 0,
        opts.rollingWindowSize ?? 0);
    measures.reset();
    const results: T[] = [];
    for (let i = opts.start ?? 0; i < returns.length; i++) {
        measures.addReturn(returns[i], benchmarkReturns[i]);
        results.push(callback(measures));
    }
    return results;
}

type Indexable = Record<string, unknown>;

/** Python `run_stream_property`: a getter evaluated after each observation. */
export function runStreamProperty<T = number>(propertyName: string, opts: StreamOptions = {}): T[] {
    return runStreamCallback(m => (m as unknown as Indexable)[propertyName] as T, opts);
}

/** Python `run_stream_method`: a method called with `args` after each observation. */
export function runStreamMethod<T = number>(methodName: string, opts: StreamOptions = {},
    ...args: unknown[]): T[] {
    return runStreamCallback(m => {
        const f = (m as unknown as Indexable)[methodName] as (...a: unknown[]) => T;
        return f.apply(m, args);
    }, opts);
}

/** Python `make_measures`. */
export function makeMeasures(rollingWindowSize: number = 0, annualRf: number = 0.0,
    annualMar: number = 0.0, daily: boolean = false, monthly: boolean = false): Measures {
    const measures = new Measures(periodsPerAnnum(daily, monthly), annualRf, annualMar, rollingWindowSize);
    measures.reset();
    return measures;
}

/** Python `add_bacon`. */
export function addBacon(measures: Measures, start: number = 0, count: number = baconPortfolioLen,
    returns: readonly number[] = baconPortfolioReturns,
    benchmarkReturns: readonly number[] = baconBenchmarkReturns): void {
    for (let i = start; i < count; i++) {
        measures.addReturn(returns[i], benchmarkReturns[i]);
    }
}

/** Python `run_stream_*(..., annual_risk_free_rate=...)` shorthand. */
export function rf(annualRiskFreeRate: number, opts: StreamOptions = {}): StreamOptions {
    return { ...opts, annualRiskFreeRate };
}

/** Python `run_stream_*(..., annual_target_return=...)` shorthand. */
export function mar(annualTargetReturn: number, opts: StreamOptions = {}): StreamOptions {
    return { ...opts, annualTargetReturn };
}

// ---------------------------------------------------------------------------
// Public measures (explicit replacement for the Python inspect-based list)
// ---------------------------------------------------------------------------

/** Public getters of Measures (Python properties), alphabetical by Python name. */
export const PUBLIC_PROPERTIES: readonly string[] = [
    'activePremium', 'adjustedSharpeRatio', 'adjustedSharpeRatioSkewOnly', 'appraisalRatio',
    'autocorrelationPenalty', 'bernardoLedoitRatio', 'burkeRatio', 'burkeRatioModified',
    'calmarRatio', 'compoundAnnualGrowthRate', 'cumulativeGeometricReturn', 'dRatio',
    'downNumberRatio', 'downPercentageRatio', 'downsideDeviation', 'downsideDeviationSubset',
    'downsideFrequency', 'downsidePotential', 'downsideSharpeRatio', 'drawdownAverage',
    'drawdownAverageLength', 'drawdownAveragePeakToTrough', 'drawdownAverageRecovery',
    'drawdownDeviation', 'drawdownsCumulative', 'drawdownsHighWatermark', 'famaBeta',
    'gainLossRatio', 'gainToPainRatio', 'geometricMeanReturn', 'hurstExponent',
    'informationRatio', 'informationRatioModified', 'jarqueBeraNormalityTestStatistic',
    'jensenAlpha', 'jensenAlphaAlternative', 'jensenAlphaModified', 'kRatio', 'kappa1Ratio',
    'kappa2Ratio', 'kappa3Ratio', 'kappa4Ratio', 'kellyRatio', 'kellyRatioFull', 'kurtosis',
    'kurtosisExcess', 'kurtosisMoment', 'kurtosisSample', 'kurtosisSampleCorrected',
    'kurtosisSampleExcess', 'lossRate', 'mSquared', 'mSquaredExcess', 'mSquaredSortino',
    'martinRatio', 'meanAbsoluteDeviationRatio', 'meanLossReturn', 'meanNonZeroReturn',
    'meanWinReturn', 'minDrawdownsCumulative', 'modigliani', 'omegaExcessReturn', 'omegaRatio',
    'omegaSharpeRatio', 'painIndex', 'painRatio', 'prospectRatioPerformanceAnalytics',
    'semiDeviation', 'sfmAlpha', 'sfmBeta', 'sfmBetaBear', 'sfmBetaBull', 'sfmR2',
    'sfmRiskPremium', 'sharpeRatio', 'skewness', 'skewnessFisher', 'skewnessKurtosisRatio',
    'skewnessMoment', 'skewnessSample', 'sortinoRatio', 'sortinoRatioSqrt2',
    'sortinoSatchellRatio', 'specificRisk', 'systematicRisk', 'timingRatio', 'totalRisk',
    'trackingError', 'treynorRatio', 'treynorRatioModified', 'ulcerIndex', 'upNumberRatio',
    'upPercentageRatio', 'upsideFrequency', 'upsidePotential', 'upsidePotentialRatio',
    'upsidePotentialRatioSubset', 'upsidePotentialSubset', 'upsideRisk', 'upsideRiskSubset',
    'upsideVariance', 'upsideVarianceSubset', 'variabilitySkewness', 'volatilitySkewness',
    'winRate', 'worstDrawdownsCumulative',
];

/** Public methods of Measures except the mutators, alphabetical by Python name. */
export const PUBLIC_METHODS: readonly string[] = [
    'biasRatio', 'cdarAlpha', 'cdarAverage', 'cdarBeta', 'cdarDiscrete', 'downsideCaptureRatio',
    'drawdownsContinuousRuns', 'esCornishFisher', 'esGaussian', 'esHistorical',
    'farinelliTibilettiRatio', 'isNormalDistribution', 'overallCaptureRatio',
    'probabilisticSharpeRatio', 'probabilisticSharpeRatioFull', 'probabilisticSharpeRatioGaussian',
    'probabilisticSharpeRatioSymmetric', 'prospectRatio', 'rachevRatio',
    'rewardToConditionalDrawdown', 'rewardToEsRatioCornishFisher', 'rewardToEsRatioGaussian',
    'rewardToEsRatioHistorical', 'rewardToVarRatioCornishFisher', 'rewardToVarRatioGaussian',
    'rewardToVarRatioHistorical', 'sharpeRatioEsCornishFisher', 'sharpeRatioEsGaussian',
    'sharpeRatioEsHistorical', 'sharpeRatioVarCornishFisher', 'sharpeRatioVarGaussian',
    'sharpeRatioVarHistorical', 'sterlingRatio', 'tailRatio', 'upsideCaptureRatio',
    'varCornishFisher', 'varGaussian', 'varHistorical',
];

const PROPERTY_SET = new Set(PUBLIC_PROPERTIES);

/** Python `public_measures()`: names of all public getters and methods, except mutators. */
export function publicMeasures(): string[] {
    return [...PUBLIC_PROPERTIES, ...PUBLIC_METHODS].sort();
}

/** Python `evaluate()`: a getter value, or a method called with its default arguments. */
export function evaluate(measures: Measures, name: string): unknown {
    const value = (measures as unknown as Indexable)[name];
    if (PROPERTY_SET.has(name)) {
        return value;
    }
    return (value as () => unknown).call(measures);
}
