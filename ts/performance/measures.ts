import { KleinKbnAccumulator, RawMomentsKleinKbn } from '../streaming-kbn';

import {
    Capture, ContinuousDrawdownRuns, CumulativeReturn, DrawdownEpisodes,
    HighWaterMarkDrawdown, PartialMoments, RawPartialMoments, SFMRegression, WinLoss,
    esCornishFisher, esGaussian, esHistorical, percentile, probabilisticSharpeRatio,
    varCornishFisher, varGaussian, varHistorical,
} from './core';

const SQRT2 = 1.4142135623730950488016887242097;

export const PERIODS_PER_ANNUM_YEAR = 1;
export const PERIODS_PER_ANNUM_QUARTER = 4;
export const PERIODS_PER_ANNUM_MONTH = 12;
export const PERIODS_PER_ANNUM_WEEK = 52;
export const PERIODS_PER_ANNUM_DAY = 252;

/** 390 regular-session minutes/day by 252 trading days/year. */
export const PERIODS_PER_ANNUM_MINUTE_US_EQUITIES = 98280;

/** 1440 minutes/day by 365 days/year. */
export const PERIODS_PER_ANNUM_MINUTE_CRYPTO = 525600;

/**
 * Python's built-in `sum()` of floats (CPython 3.12+): Neumaier-compensated
 * summation starting at 0.0.
 */
function pySum(xs: readonly number[]): number {
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

/** Python's `int(x)` for a float: truncation, throwing on NaN and infinities. */
function pyInt(x: number): number {
    if (Number.isNaN(x)) {
        throw new Error('cannot convert float NaN to integer');
    }
    if (!Number.isFinite(x)) {
        throw new Error('cannot convert float infinity to integer');
    }
    return Math.trunc(x);
}

/**
 * Streaming calculation of time-series performance and risk measures.
 *
 * `Measures` operates on return observations that have a common, explicitly
 * defined observation period. It does not require timestamps and does not
 * perform time-based resampling.
 *
 * `periodsPerAnnum` defines the annualization convention used by annualized
 * measures: the number of return observations assumed to represent one year
 * (daily 252, weekly 52, monthly 12, quarterly 4, annual 1, one-minute US
 * equity session 98 280). It also converts the annual risk-free rate and
 * annual target return to periodic rates: `(1 + r) ** (1 / periodsPerAnnum) - 1`.
 *
 * A port of the Python `performance.Measures` class; properties become
 * getters and methods keep the Python default arguments.
 */
export class Measures {
    /** Number of return periods per annum used for annualization. */
    readonly periodsPerAnnum: number;
    /** Periodic risk-free rate. */
    readonly riskFreeRate: number;
    /** Periodic target return (minimum acceptable return, MAR). */
    readonly targetReturn: number;

    private readonly _sqrtPeriodsPerAnnum: number;
    private readonly _annualRiskFreeRate: number;
    private readonly _rollingWindowSize: number;

    private readonly _returns: number[] = [];
    private readonly _returnsBenchmark: number[] = [];

    private readonly _winLoss = new WinLoss();
    private readonly _capture = new Capture();

    // ddof=1, bias=true, fisher=true matches scipy's default kurtosis behavior.
    private readonly _returnsKbn = new RawMomentsKleinKbn(1, true, true);
    private readonly _excessReturnsKbn = new RawMomentsKleinKbn(1, true, true);
    private readonly _benchmarkReturnsKbn = new RawMomentsKleinKbn(1, true, true);
    private readonly _benchmarkExcessReturnsKbn = new RawMomentsKleinKbn(1, true, true);

    private readonly _sfmRegression: SFMRegression;
    private readonly _activeReturnsKbn = new RawMomentsKleinKbn(1, true, true);

    private readonly _targetReturnsKbn = new RawMomentsKleinKbn(1, true, true);
    private readonly _targetPartialMoments: PartialMoments;
    private readonly _rawPartialMoments = new RawPartialMoments();
    private readonly _benchmarkTargetPartialMoments: PartialMoments;

    private readonly _cumulativeReturn = new CumulativeReturn();
    private readonly _cumulativeExcessReturn = new CumulativeReturn();
    private readonly _benchmarkCumulativeReturn = new CumulativeReturn();

    private readonly _drawdownContinuousRuns = new ContinuousDrawdownRuns();
    private readonly _drawdownHighWatermark: HighWaterMarkDrawdown;
    private readonly _drawdownHighWatermarkBenchmark: HighWaterMarkDrawdown;
    private readonly _drawdownEpisodes = new DrawdownEpisodes();
    private readonly _drawdownEpisodesBenchmark = new DrawdownEpisodes();

    /**
     * @param periodsPerAnnum Number of return periods per annum (must be positive).
     * @param annualRiskFreeRate Annual risk-free rate as a decimal.
     * @param annualTargetReturn Annual target return (MAR) as a decimal.
     * @param rollingWindowSize Number of most recent observations retained;
     *     zero, negative or null specifies an unbounded window.
     * @throws Error if `periodsPerAnnum` is not positive.
     */
    constructor(periodsPerAnnum: number = 252.0, annualRiskFreeRate: number = 0.0,
        annualTargetReturn: number = 0.0, rollingWindowSize: number | null = 0) {
        if (periodsPerAnnum <= 0) {
            throw new Error('periods_per_annum must be positive');
        }
        this.periodsPerAnnum = periodsPerAnnum;
        this._sqrtPeriodsPerAnnum = Math.sqrt(periodsPerAnnum);

        this._annualRiskFreeRate = annualRiskFreeRate;
        this.riskFreeRate = (annualRiskFreeRate === 0 || periodsPerAnnum === 1)
            ? annualRiskFreeRate
            : ((1 + annualRiskFreeRate) ** (1 / periodsPerAnnum) - 1);

        this.targetReturn = (annualTargetReturn === 0 || periodsPerAnnum === 1)
            ? annualTargetReturn
            : ((1 + annualTargetReturn) ** (1 / periodsPerAnnum) - 1);

        this._rollingWindowSize = (rollingWindowSize === null || rollingWindowSize === undefined
            || rollingWindowSize < 0) ? 0 : rollingWindowSize;

        this._sfmRegression = new SFMRegression(this.riskFreeRate);
        this._targetPartialMoments = new PartialMoments(this.targetReturn);
        this._benchmarkTargetPartialMoments = new PartialMoments(this.targetReturn);
        this._drawdownHighWatermark = new HighWaterMarkDrawdown(this._rollingWindowSize);
        this._drawdownHighWatermarkBenchmark = new HighWaterMarkDrawdown(this._rollingWindowSize);
    }

    /**
     * Reset all accumulated return data and derived streaming state.
     * Configuration parameters are preserved.
     */
    reset(): void {
        this._returns.length = 0;
        this._returnsBenchmark.length = 0;

        this._winLoss.reset();
        this._capture.reset();

        this._returnsKbn.reset();
        this._excessReturnsKbn.reset();
        this._benchmarkReturnsKbn.reset();
        this._benchmarkExcessReturnsKbn.reset();

        this._sfmRegression.reset();
        this._activeReturnsKbn.reset();

        this._targetReturnsKbn.reset();
        this._targetPartialMoments.reset();
        this._rawPartialMoments.reset();
        this._benchmarkTargetPartialMoments.reset();

        this._cumulativeReturn.reset();
        this._cumulativeExcessReturn.reset();
        this._benchmarkCumulativeReturn.reset();

        this._drawdownContinuousRuns.reset();
        this._drawdownHighWatermark.reset();
        this._drawdownHighWatermarkBenchmark.reset();
        this._drawdownEpisodes.reset();
        this._drawdownEpisodesBenchmark.reset();
    }

    /**
     * Add one periodic portfolio and benchmark return observation.
     *
     * @param ret Portfolio return for the current period, as a decimal.
     * @param retBench Benchmark return for the same period, as a decimal.
     */
    addReturn(ret: number, retBench: number): void {
        const evicted = this._rollingWindowSize > 0 && this._returns.length === this._rollingWindowSize;
        if (evicted) {
            const retOld = this._returns.shift() as number;
            const retBenchOld = this._returnsBenchmark.shift() as number;
            this._returnsKbn.revert(retOld);
            // Excess returns (returns less risk-free rate)
            this._excessReturnsKbn.revert(retOld - this.riskFreeRate);
            // Target returns (returns less target return)
            this._targetReturnsKbn.revert(retOld - this.targetReturn);
            this._targetPartialMoments.revert(retOld);
            this._rawPartialMoments.revert(retOld);
            this._benchmarkTargetPartialMoments.revert(retBenchOld);
            this._winLoss.revert(retOld);
            this._capture.revert(retOld, retBenchOld);
            this._cumulativeReturn.revert(retOld);
            this._cumulativeExcessReturn.revert(retOld - this.riskFreeRate);
            this._benchmarkCumulativeReturn.revert(retBenchOld);
            // Benchmarks
            this._benchmarkReturnsKbn.revert(retBenchOld);
            this._benchmarkExcessReturnsKbn.revert(retBenchOld - this.riskFreeRate);
            this._activeReturnsKbn.revert(retOld - retBenchOld);
            this._sfmRegression.revert(retOld, retBenchOld);
            // Drawdowns: high watermark drawdown and drawdown episodes have no
            // revert(), they evict the oldest observation themselves.
            this._drawdownContinuousRuns.revert(retOld); // Burke
        }

        this._returnsKbn.update(ret);
        // Excess returns (returns less risk-free rate)
        const retExcess = ret - this.riskFreeRate;
        this._excessReturnsKbn.update(retExcess);
        // Target returns (returns less target return)
        this._targetReturnsKbn.update(ret - this.targetReturn);
        this._targetPartialMoments.update(ret);
        this._rawPartialMoments.update(ret);
        this._benchmarkTargetPartialMoments.update(retBench);
        this._winLoss.update(ret);
        this._capture.update(ret, retBench);
        // Benchmarks
        this._benchmarkReturnsKbn.update(retBench);
        const retBenchExcess = retBench - this.riskFreeRate;
        this._benchmarkExcessReturnsKbn.update(retBenchExcess);
        this._activeReturnsKbn.update(ret - retBench);
        this._sfmRegression.update(ret, retBench);

        this._returns.push(ret);
        this._returnsBenchmark.push(retBench);

        // Cumulative return
        this._cumulativeReturn.update(ret);
        this._cumulativeExcessReturn.update(retExcess);
        this._benchmarkCumulativeReturn.update(retBench);

        // Drawdown calculation used in Burke
        this._drawdownContinuousRuns.update(ret);

        // High-water-mark drawdown and drawdown episodes. Episodes have no
        // revert(): when an observation leaves the rolling window (or the
        // window's drawdowns were recomputed), the episodes are rebuilt from
        // the window's drawdowns, so that episode indices refer to positions
        // in the window.
        let hwm = this._drawdownHighWatermark;
        let recalculated = hwm.update(ret);
        if (recalculated || evicted) {
            this._drawdownEpisodes.recalculate(hwm.drawdowns);
        } else {
            this._drawdownEpisodes.update(hwm.drawdown);
        }
        hwm = this._drawdownHighWatermarkBenchmark;
        recalculated = hwm.update(retBench);
        if (recalculated || evicted) {
            this._drawdownEpisodesBenchmark.recalculate(hwm.drawdowns);
        } else {
            this._drawdownEpisodesBenchmark.update(hwm.drawdown);
        }
    }

    /**
     * Lo (2002) autocorrelation penalty factor
     * `sqrt(1 + 2 * sum_{k=1}^{q-1} (1 - k/q) * rho_k)`, with `q = min(P, n - 1)`.
     * Returns 1.0 with fewer than two observations or zero variance. O(n·q).
     */
    get autocorrelationPenalty(): number {
        const n = this._returnsKbn.n;
        if (n < 2) {
            return 1.0;
        }
        const mean = this._returnsKbn.mean;
        const denom = this._returnsKbn.varianceDdof0 * n;
        if (Number.isNaN(denom) || denom === 0) {
            return 1.0;
        }

        // Lo's recommended aggregation period: daily 252, weekly 52, monthly 12, quarterly 4.
        const ppa = this.periodsPerAnnum;
        const q = pyInt((n - 1) < ppa ? (n - 1) : ppa);

        const w = this._returns;
        let s = 0.0;
        for (let k = 1; k < q; k++) {
            let numer = 0.0;
            for (let t = k; t < n; t++) {
                numer += (w[t] - mean) * (w[t - k] - mean);
            }
            const rho = numer / denom;
            s += (1.0 - k / q) * rho;
        }

        const v = 1.0 + 2.0 * s;
        return Math.sqrt(v > 0.0 ? v : 0.0); // Python max(0.0, v)
    }

    /** Cumulative geometric return (0.0 when empty). */
    get cumulativeGeometricReturn(): number {
        return this._cumulativeReturn.cumulativeGeometricReturn;
    }

    /** Geometric mean return per observation. */
    get geometricMeanReturn(): number {
        return this._cumulativeReturn.geometricMeanReturn;
    }

    /** Compound annual growth rate (annualized geometric mean return). */
    get compoundAnnualGrowthRate(): number {
        return this._cumulativeReturn.annualizedGeometricMeanReturn(this.periodsPerAnnum);
    }

    /** Skewness ('moment', bias=True). */
    get skewness(): number {
        return this._returnsKbn.skewness;
    }

    /** 'Moment' skewness g1 = mu3 / mu2^(3/2). */
    get skewnessMoment(): number {
        return this._returnsKbn.skewnessMoment;
    }

    /** 'Fisher' skewness g1 * sqrt(n(n-1)) / (n-2). */
    get skewnessFisher(): number {
        return this._returnsKbn.skewnessFisher;
    }

    /** 'Sample' skewness g1 * n^2 / ((n-1)(n-2)). */
    get skewnessSample(): number {
        return this._returnsKbn.skewnessSample;
    }

    /** Kurtosis (biased excess, the scipy default). */
    get kurtosis(): number {
        return this._returnsKbn.kurtosis;
    }

    /** Biased excess kurtosis beta2 - 3. */
    get kurtosisExcess(): number {
        return this._returnsKbn.kurtosisExcess;
    }

    /** Biased Pearson kurtosis beta2 = mu4 / mu2^2. */
    get kurtosisMoment(): number {
        return this._returnsKbn.kurtosisMoment;
    }

    /** Unbiased excess kurtosis ((n^2-1) beta2 - 3(n-1)^2) / ((n-2)(n-3)). */
    get kurtosisSampleExcess(): number {
        return this._returnsKbn.kurtosisSampleExcess;
    }

    /** PerformanceAnalytics 'sample' kurtosis (n^2-1) beta2 / ((n-2)(n-3)). */
    get kurtosisSampleCorrected(): number {
        return this._returnsKbn.kurtosisSampleCorrected;
    }

    /** Unbiased Pearson kurtosis (sample excess + 3). */
    get kurtosisSample(): number {
        return this._returnsKbn.kurtosisSample;
    }

    /** Ratio of the population skewness g1 to the population kurtosis beta2. */
    get skewnessKurtosisRatio(): number {
        const s = this._returnsKbn.skewnessMoment;
        const k = this._returnsKbn.kurtosisMoment;
        return k !== 0 ? s / k : NaN;
    }

    /** Jarque-Bera normality test statistic n/6 * (S^2 + K_excess^2 / 4). */
    get jarqueBeraNormalityTestStatistic(): number {
        // Population skewness and excess kurtosis (not sample corrected).
        const s = this._returnsKbn.skewnessMoment;
        const k = this._returnsKbn.kurtosisExcess;
        if (Number.isNaN(s) || Number.isNaN(k)) {
            return NaN;
        }
        // Bera-Jarque formula (Equation 5.17 from Bacon 3rd ed)
        const n = this._returnsKbn.n;
        return n / 6 * (s * s + k * k / 4);
    }

    /**
     * Jarque-Bera test of the normality hypothesis.
     *
     * @param confidence Confidence level in (0, 1).
     * @returns True if normality cannot be rejected; false if it is rejected
     *     or there is insufficient data.
     * @throws Error if confidence is not in (0, 1) (checked only when the
     *     statistic is available).
     */
    isNormalDistribution(confidence: number = 0.95): boolean {
        const bj = this.jarqueBeraNormalityTestStatistic;
        if (Number.isNaN(bj)) {
            return false;
        }
        if (confidence <= 0 || confidence >= 1) {
            throw new Error('confidence must be between 0 and 1');
        }
        // Inverse CDF of chi-squared with 2 degrees of freedom: -2 ln(1 - p).
        const critical = -2.0 * Math.log1p(-confidence);
        return bj <= critical;
    }

    /** Historical VaR of the returns. */
    varHistorical(confidence: number = 0.95): number {
        return varHistorical(this._returns, 0.0, confidence);
    }

    /** Gaussian (parametric) VaR of the returns. */
    varGaussian(confidence: number = 0.95): number {
        return varGaussian(this._returnsKbn, confidence);
    }

    /** Modified Cornish-Fisher VaR of the returns. */
    varCornishFisher(confidence: number = 0.95): number {
        return varCornishFisher(this._returnsKbn, confidence);
    }

    /** Historical expected shortfall of the returns. */
    esHistorical(confidence: number = 0.95): number {
        return esHistorical(this._returns, 0.0, confidence);
    }

    /** Gaussian (parametric) expected shortfall of the returns. */
    esGaussian(confidence: number = 0.95): number {
        return esGaussian(this._returnsKbn, confidence);
    }

    /** Modified Cornish-Fisher expected shortfall of the returns. */
    esCornishFisher(confidence: number = 0.95): number {
        return esCornishFisher(this._returnsKbn, confidence);
    }

    /** Mean excess return divided by the historical VaR of the raw returns. */
    rewardToVarRatioHistorical(confidence: number = 0.95): number {
        const denom = this.varHistorical(confidence);
        return denom !== 0 ? this._excessReturnsKbn.mean / denom : NaN;
    }

    /** Mean excess return divided by the Gaussian VaR of the raw returns. */
    rewardToVarRatioGaussian(confidence: number = 0.95): number {
        const denom = this.varGaussian(confidence);
        return denom !== 0 ? this._excessReturnsKbn.mean / denom : NaN;
    }

    /** Mean excess return divided by the Cornish-Fisher VaR of the raw returns. */
    rewardToVarRatioCornishFisher(confidence: number = 0.95): number {
        const denom = this.varCornishFisher(confidence);
        return denom !== 0 ? this._excessReturnsKbn.mean / denom : NaN;
    }

    /** Mean excess return divided by the historical ES of the raw returns. */
    rewardToEsRatioHistorical(confidence: number = 0.95): number {
        const denom = this.esHistorical(confidence);
        return denom !== 0 ? this._excessReturnsKbn.mean / denom : NaN;
    }

    /** Mean excess return divided by the Gaussian ES of the raw returns. */
    rewardToEsRatioGaussian(confidence: number = 0.95): number {
        const denom = this.esGaussian(confidence);
        return denom !== 0 ? this._excessReturnsKbn.mean / denom : NaN;
    }

    /** Mean excess return divided by the Cornish-Fisher ES of the raw returns. */
    rewardToEsRatioCornishFisher(confidence: number = 0.95): number {
        const denom = this.esCornishFisher(confidence);
        return denom !== 0 ? this._excessReturnsKbn.mean / denom : NaN;
    }

    /** Mean return divided by the mean absolute deviation about the mean. O(n). */
    get meanAbsoluteDeviationRatio(): number {
        const n = this._returnsKbn.n;
        const w = this._returns;
        if (n < 1) {
            return NaN;
        }
        const mean = this._returnsKbn.mean;
        const sum = new KleinKbnAccumulator();
        for (const x of w) {
            sum.update(Math.abs(x - mean));
        }
        const mad = sum.value / n;
        return mad > 0 ? mean / mad : NaN;
    }

    /** Upside potential ratio HPM1 / sqrt(LPM2) about the target return. */
    get upsidePotentialRatio(): number {
        const hpm1 = this._targetPartialMoments.higherPartialMoment1;
        const lpm2 = this._targetPartialMoments.lowerPartialMoment2;
        if (Number.isNaN(hpm1) || Number.isNaN(lpm2) || lpm2 === 0) {
            return NaN;
        }
        return hpm1 / Math.sqrt(lpm2);
    }

    /** Upside potential ratio using subset (above/below target) averages. */
    get upsidePotentialRatioSubset(): number {
        const n1 = this._targetPartialMoments.upperExcessCount;
        const n2 = this._targetPartialMoments.lowerExcessCount;
        if (n1 === 0 || n2 === 0) {
            return NaN;
        }
        const hpm1 = this._targetPartialMoments.upperExcessMoment1Sum / n1;
        const lpm2 = this._targetPartialMoments.lowerExcessMoment2Sum / n2;
        if (Number.isNaN(hpm1) || Number.isNaN(lpm2) || lpm2 === 0) {
            return NaN;
        }
        return hpm1 / Math.sqrt(lpm2);
    }

    /** Fraction of returns above the target return. */
    get upsideFrequency(): number {
        return this._targetPartialMoments.upsideFrequency;
    }

    /** Upside potential HPM1 about the target return. */
    get upsidePotential(): number {
        return this._targetPartialMoments.higherPartialMoment1;
    }

    /** Upside potential averaged over the returns above the target (0 if none). */
    get upsidePotentialSubset(): number {
        const n = this._targetPartialMoments.upperExcessCount;
        return n > 0 ? this._targetPartialMoments.upperExcessMoment1Sum / n : 0;
    }

    /** Upside variance HPM2 about the target return. */
    get upsideVariance(): number {
        return this._targetPartialMoments.higherPartialMoment2;
    }

    /** Upside variance averaged over the returns above the target (0 if none). */
    get upsideVarianceSubset(): number {
        const n = this._targetPartialMoments.upperExcessCount;
        return n > 0 ? this._targetPartialMoments.upperExcessMoment2Sum / n : 0;
    }

    /** Upside risk sqrt(HPM2). */
    get upsideRisk(): number {
        const variance = this.upsideVariance;
        return Number.isNaN(variance) ? NaN : Math.sqrt(variance);
    }

    /** Upside risk over the returns above the target. */
    get upsideRiskSubset(): number {
        const variance = this.upsideVarianceSubset;
        return Number.isNaN(variance) ? NaN : Math.sqrt(variance);
    }

    /** Semi-deviation: sqrt of the mean squared negative deviation from the mean (÷n). O(n). */
    get semiDeviation(): number {
        const n = this._returnsKbn.n;
        const returns = this._returns;
        if (n === 0) {
            return NaN;
        }
        const mean = this._returnsKbn.mean;
        const sumSquared = new KleinKbnAccumulator();
        for (const r of returns) {
            const deviation = r - mean;
            if (deviation < 0) {
                sumSquared.update(deviation * deviation);
            }
        }
        return Math.sqrt(sumSquared.value / n);
    }

    /** Downside deviation sqrt(LPM2) about the target return. */
    get downsideDeviation(): number {
        const denom = this._targetPartialMoments.totalCount;
        if (denom === 0) {
            return NaN;
        }
        return Math.sqrt(this._targetPartialMoments.lowerExcessMoment2Sum / denom);
    }

    /** Downside deviation over the returns below the target (0 if none). */
    get downsideDeviationSubset(): number {
        const denom = this._targetPartialMoments.lowerExcessCount;
        if (denom === 0) {
            return 0;
        }
        return Math.sqrt(this._targetPartialMoments.lowerExcessMoment2Sum / denom);
    }

    /** Fraction of returns below the target return. */
    get downsideFrequency(): number {
        return this._targetPartialMoments.downsideFrequency;
    }

    /** Downside potential LPM1 about the target return. */
    get downsidePotential(): number {
        return this._targetPartialMoments.downsidePotential;
    }

    /** Sharpe ratio mean(e) / stdev_ddof1(e) of the excess returns. */
    get sharpeRatio(): number {
        const std = this._excessReturnsKbn.standardDeviationDdof1;
        if (Number.isNaN(std) || std === 0) {
            return NaN;
        }
        return this._excessReturnsKbn.mean / std;
    }

    /** Mean excess return divided by the historical VaR of the excess returns. */
    sharpeRatioVarHistorical(confidence: number = 0.95): number {
        if (this._excessReturnsKbn.n < 2) {
            return NaN;
        }
        const denom = varHistorical(this._returns, this.riskFreeRate, confidence);
        if (Number.isNaN(denom) || denom === 0) {
            return NaN;
        }
        return this._excessReturnsKbn.mean / denom;
    }

    /** Mean excess return divided by the Gaussian VaR of the excess returns. */
    sharpeRatioVarGaussian(confidence: number = 0.95): number {
        if (this._excessReturnsKbn.n < 2) {
            return NaN;
        }
        const denom = varGaussian(this._excessReturnsKbn, confidence);
        if (Number.isNaN(denom) || denom === 0) {
            return NaN;
        }
        return this._excessReturnsKbn.mean / denom;
    }

    /** Mean excess return divided by the Cornish-Fisher VaR of the excess returns. */
    sharpeRatioVarCornishFisher(confidence: number = 0.95): number {
        if (this._excessReturnsKbn.n < 2) {
            return NaN;
        }
        const denom = varCornishFisher(this._excessReturnsKbn, confidence);
        if (Number.isNaN(denom) || denom === 0) {
            return NaN;
        }
        return this._excessReturnsKbn.mean / denom;
    }

    /** Mean excess return divided by the historical ES of the excess returns. */
    sharpeRatioEsHistorical(confidence: number = 0.95): number {
        if (this._excessReturnsKbn.n < 2) {
            return NaN;
        }
        const denom = esHistorical(this._returns, this.riskFreeRate, confidence);
        if (Number.isNaN(denom) || denom === 0) {
            return NaN;
        }
        return this._excessReturnsKbn.mean / denom;
    }

    /** Mean excess return divided by the Gaussian ES of the excess returns. */
    sharpeRatioEsGaussian(confidence: number = 0.95): number {
        if (this._excessReturnsKbn.n < 2) {
            return NaN;
        }
        const denom = esGaussian(this._excessReturnsKbn, confidence);
        if (Number.isNaN(denom) || denom === 0) {
            return NaN;
        }
        return this._excessReturnsKbn.mean / denom;
    }

    /** Mean excess return divided by the Cornish-Fisher ES of the excess returns. */
    sharpeRatioEsCornishFisher(confidence: number = 0.95): number {
        if (this._excessReturnsKbn.n < 2) {
            return NaN;
        }
        const denom = esCornishFisher(this._excessReturnsKbn, confidence);
        if (Number.isNaN(denom) || denom === 0) {
            return NaN;
        }
        return this._excessReturnsKbn.mean / denom;
    }

    /** Downside Sharpe ratio mean(e) / (sqrt(2) * semi-deviation); ±Infinity if the semi-deviation is 0. */
    get downsideSharpeRatio(): number {
        const semiDev = this.semiDeviation;
        if (Number.isNaN(semiDev)) {
            return NaN;
        }
        if (semiDev === 0) {
            return this._excessReturnsKbn.mean < 0 ? -Infinity : Infinity;
        }
        return this._excessReturnsKbn.mean / (SQRT2 * semiDev);
    }

    /** Adjusted Sharpe ratio SR * (1 + S*SR/6 - K*SR^2/24) (Pezier and White). */
    get adjustedSharpeRatio(): number {
        const skewness = this._returnsKbn.skewnessMoment;
        const kurtosis = this._returnsKbn.kurtosisExcess;
        if (Number.isNaN(skewness) || Number.isNaN(kurtosis)) {
            return NaN;
        }
        const sr = this.sharpeRatio;
        if (Number.isNaN(sr)) {
            return NaN;
        }
        return sr * (1 + skewness * sr / 6 - kurtosis * sr * sr / 24);
    }

    /** Skewness-only adjusted Sharpe ratio SR * (1 + S*SR/6). */
    get adjustedSharpeRatioSkewOnly(): number {
        const s = this._returnsKbn.skewnessMoment;
        if (Number.isNaN(s)) {
            return NaN;
        }
        const sr = this.sharpeRatio;
        if (Number.isNaN(sr)) {
            return NaN;
        }
        return sr * (1 + s * sr / 6);
    }

    /** Probabilistic Sharpe ratio with sample skewness and normal kurtosis. */
    probabilisticSharpeRatio(referenceSr: number = 0.0): number {
        return probabilisticSharpeRatio(this._returnsKbn, this.sharpeRatio, referenceSr, false, true);
    }

    /** Probabilistic Sharpe ratio with sample skewness and sample kurtosis. */
    probabilisticSharpeRatioFull(referenceSr: number = 0.0): number {
        return probabilisticSharpeRatio(this._returnsKbn, this.sharpeRatio, referenceSr, false, false);
    }

    /** Probabilistic Sharpe ratio with zero skewness and sample kurtosis. */
    probabilisticSharpeRatioSymmetric(referenceSr: number = 0.0): number {
        return probabilisticSharpeRatio(this._returnsKbn, this.sharpeRatio, referenceSr, true, false);
    }

    /** Probabilistic Sharpe ratio with zero skewness and normal kurtosis. */
    probabilisticSharpeRatioGaussian(referenceSr: number = 0.0): number {
        return probabilisticSharpeRatio(this._returnsKbn, this.sharpeRatio, referenceSr, true, true);
    }

    /** Sortino ratio mean(r - T) / sqrt(LPM2). */
    get sortinoRatio(): number {
        const lpm2 = this._targetPartialMoments.lowerPartialMoment2;
        if (Number.isNaN(lpm2) || lpm2 === 0) {
            return NaN;
        }
        return this._targetReturnsKbn.mean / Math.sqrt(lpm2);
    }

    /** Sortino ratio divided by sqrt(2) (Jack Schwager's version). */
    get sortinoRatioSqrt2(): number {
        return this.sortinoRatio / SQRT2;
    }

    /** Sortino-Satchell ratio (equals the Sortino ratio and kappa 2). */
    get sortinoSatchellRatio(): number {
        const lpm2 = this._targetPartialMoments.lowerPartialMoment2;
        if (Number.isNaN(lpm2) || lpm2 === 0) {
            return NaN;
        }
        return this._targetReturnsKbn.mean / Math.sqrt(lpm2);
    }

    /** Omega ratio mean(r - T) / LPM1 + 1. */
    get omegaRatio(): number {
        const lpm1 = this._targetPartialMoments.lowerPartialMoment1;
        if (Number.isNaN(lpm1) || lpm1 === 0) {
            return NaN;
        }
        return this._targetReturnsKbn.mean / lpm1 + 1;
    }

    /** Omega-Sharpe ratio (omega ratio - 1). */
    get omegaSharpeRatio(): number {
        return this.omegaRatio - 1;
    }

    /**
     * Omega excess return (annualized): Rp_ann - 3 * sigma_d_ann * sigma_d_ann_bench.
     * Recomputes the benchmark downside deviation over the window, O(n).
     */
    get omegaExcessReturn(): number {
        const benchmarkDownsideDeviation = (): number => {
            const rb = this._returnsBenchmark;
            if (rb.length === 0) {
                return NaN;
            }
            const lowerExcessKbn = new RawMomentsKleinKbn(1, true, true);
            for (const r of rb) {
                const excess = r - this.targetReturn;
                if (excess < 0) {
                    lowerExcessKbn.update(-excess);
                }
            }
            return Math.sqrt(lowerExcessKbn.x2Sum / rb.length);
        };

        const period = this.periodsPerAnnum;
        const rp = this._cumulativeReturn.geometricMeanReturn;
        if (Number.isNaN(rp)) {
            return NaN;
        }
        const rpAnnual = (1 + rp) ** period - 1;
        const sqrtPeriod = Math.sqrt(period);
        const sigmaDAnn = this.downsideDeviation * sqrtPeriod;
        const sigmaDAnnBench = benchmarkDownsideDeviation() * sqrtPeriod;
        return rpAnnual - 3 * sigmaDAnn * sigmaDAnnBench;
    }

    /** Kappa 1 ratio mean(r - T) / LPM1. */
    get kappa1Ratio(): number {
        const lpm = this._targetPartialMoments.lowerPartialMoment1;
        if (Number.isNaN(lpm) || lpm === 0) {
            return NaN;
        }
        return this._targetReturnsKbn.mean / lpm;
    }

    /** Kappa 2 ratio mean(r - T) / LPM2^(1/2). */
    get kappa2Ratio(): number {
        const lpm = this._targetPartialMoments.lowerPartialMoment2;
        if (Number.isNaN(lpm) || lpm === 0) {
            return NaN;
        }
        return this._targetReturnsKbn.mean / Math.sqrt(lpm);
    }

    /** Kappa 3 ratio mean(r - T) / LPM3^(1/3). */
    get kappa3Ratio(): number {
        const lpm = this._targetPartialMoments.lowerPartialMoment3;
        if (Number.isNaN(lpm) || lpm === 0) {
            return NaN;
        }
        return this._targetReturnsKbn.mean / lpm ** (1 / 3);
    }

    /** Kappa 4 ratio mean(r - T) / LPM4^(1/4). */
    get kappa4Ratio(): number {
        const lpm = this._targetPartialMoments.lowerPartialMoment4;
        if (Number.isNaN(lpm) || lpm === 0) {
            return NaN;
        }
        return this._targetReturnsKbn.mean / lpm ** (1 / 4);
    }

    /** Prospect ratio ((Σr⁺ + λ·Σr⁻)/n - T) / downside deviation (reference formula). */
    prospectRatio(lambdaLoss: number = 2.25): number {
        const ddev = this.downsideDeviation;
        if (Number.isNaN(ddev) || ddev === 0) {
            return NaN;
        }
        const pm = this._rawPartialMoments;
        const n = pm.count;
        if (n === 0) {
            return NaN;
        }
        const prospectReturn = (pm.sumPositive + lambdaLoss * pm.sumNegative) / n;
        return (prospectReturn - this.targetReturn) / ddev;
    }

    /** Prospect ratio as computed by PerformanceAnalytics (λ = 2.25). */
    get prospectRatioPerformanceAnalytics(): number {
        const lambdaLoss = 2.25;
        const ddev = this.downsideDeviation;
        if (Number.isNaN(ddev) || ddev === 0) {
            return NaN;
        }
        const pm = this._rawPartialMoments;
        const n = pm.count;
        if (n === 0) {
            return NaN;
        }
        return (pm.sumPositive + lambdaLoss * pm.sumNegative - this.targetReturn) / (ddev * n);
    }

    /** Bernardo-Ledoit ratio Σmax(r, 0) / Σmax(-r, 0). */
    get bernardoLedoitRatio(): number {
        const lpm1 = this._rawPartialMoments.lowerPartialMoment1;
        const hpm1 = this._rawPartialMoments.higherPartialMoment1;
        return lpm1 !== 0 ? hpm1 / lpm1 : NaN;
    }

    /** d ratio -n_down·Σdown / (n_up·Σup); Infinity with no positive returns, 0 with no negative ones. */
    get dRatio(): number {
        const nUp = this._winLoss.winningReturnsCount;
        if (nUp === 0) {
            return Infinity;
        }
        const nDown = this._winLoss.losingReturnsCount;
        if (nDown === 0) {
            return 0.0;
        }
        const sumUp = this._winLoss.winningReturnsSum;
        const sumDown = this._winLoss.losingReturnsSum;
        return -nDown * sumDown / (nUp * sumUp);
    }

    /** Gain-loss ratio Σgains / |Σlosses|. */
    get gainLossRatio(): number {
        const sumLosses = Math.abs(this._winLoss.losingReturnsSum);
        return sumLosses !== 0 ? this._winLoss.winningReturnsSum / sumLosses : NaN;
    }

    /** Mean of the non-zero returns. */
    get meanNonZeroReturn(): number {
        return this._winLoss.nonZeroReturnsMean;
    }

    /** Mean of the positive returns. */
    get meanWinReturn(): number {
        return this._winLoss.winningReturnsMean;
    }

    /** Mean of the negative returns. */
    get meanLossReturn(): number {
        return this._winLoss.losingReturnsMean;
    }

    /** Fraction of non-zero returns that are positive. */
    get winRate(): number {
        const nonZeroCount = this._winLoss.nonZeroReturnsCount;
        if (nonZeroCount <= 0) {
            return NaN;
        }
        return this._winLoss.winningReturnsCount / nonZeroCount;
    }

    /** Fraction of non-zero returns that are negative. */
    get lossRate(): number {
        const nonZeroCount = this._winLoss.nonZeroReturnsCount;
        if (nonZeroCount <= 0) {
            return NaN;
        }
        return this._winLoss.losingReturnsCount / nonZeroCount;
    }

    /** Variability skewness HPM2 / LPM2. */
    get variabilitySkewness(): number {
        const upMoment = this._targetPartialMoments.higherPartialMoment2;
        const downMoment = this._targetPartialMoments.lowerPartialMoment2;
        if (Number.isNaN(upMoment) || Number.isNaN(downMoment) || downMoment === 0) {
            return NaN;
        }
        return upMoment / downMoment;
    }

    /** Volatility skewness sqrt(HPM2 / LPM2). */
    get volatilitySkewness(): number {
        const varSkew = this.variabilitySkewness;
        return !Number.isNaN(varSkew) ? Math.sqrt(varSkew) : NaN;
    }

    /**
     * Farinelli-Tibiletti ratio HPM_p^(1/p) / LPM_q^(1/q).
     *
     * @throws Error if an order is not 1, 2, 3 or 4.
     */
    farinelliTibilettiRatio(upperOrder: number = 2, lowerOrder: number = 2): number {
        if (!(upperOrder === 1 || upperOrder === 2 || upperOrder === 3 || upperOrder === 4)) {
            throw new Error('upper_order must be 1, 2, 3, or 4');
        }
        if (!(lowerOrder === 1 || lowerOrder === 2 || lowerOrder === 3 || lowerOrder === 4)) {
            throw new Error('lower_order must be 1, 2, 3, or 4');
        }
        let denom: number;
        if (lowerOrder === 1) {
            denom = this._targetPartialMoments.lowerPartialMoment1;
        } else if (lowerOrder === 2) {
            denom = this._targetPartialMoments.lowerPartialMoment2;
            denom = Math.sqrt(denom);
        } else if (lowerOrder === 3) {
            denom = this._targetPartialMoments.lowerPartialMoment3;
            denom = denom ** (1.0 / 3);
        } else {
            denom = this._targetPartialMoments.lowerPartialMoment4;
            denom = denom ** (1.0 / 4);
        }
        let num: number;
        if (upperOrder === 1) {
            num = this._targetPartialMoments.higherPartialMoment1;
        } else if (upperOrder === 2) {
            num = this._targetPartialMoments.higherPartialMoment2;
            num = Math.sqrt(num);
        } else if (upperOrder === 3) {
            num = this._targetPartialMoments.higherPartialMoment3;
            num = num ** (1.0 / 3);
        } else {
            num = this._targetPartialMoments.higherPartialMoment4;
            num = num ** (1.0 / 4);
        }
        if (Number.isNaN(num) || Number.isNaN(denom) || denom === 0) {
            return NaN;
        }
        return num / denom;
    }

    /**
     * Rachev ratio (PerformanceAnalytics non-parametric): upper-tail ES at
     * `beta` divided by lower-tail ES at `alpha`. O(n log n).
     *
     * @throws Error if alpha or beta is not in (0, 1) (checked only with n >= 2).
     */
    rachevRatio(alpha: number = 0.1, beta: number = 0.1): number {
        const n = this._returnsKbn.n;
        const returns = this._returns;
        if (n < 2) {
            return NaN;
        }
        if (!(0 < alpha && alpha < 1)) {
            throw new Error('alpha must be between 0 and 1');
        }
        if (!(0 < beta && beta < 1)) {
            throw new Error('beta must be between 0 and 1');
        }
        const lowerVar = percentile(returns, alpha);
        const lowerTail = returns.filter(r => r <= lowerVar);
        if (lowerTail.length === 0) {
            return NaN;
        }
        const esLower = -pySum(lowerTail) / lowerTail.length;
        const sortedReturns = returns.slice().sort((a, b) => a - b);
        let upperPosition = Math.floor((1.0 - beta) * n);
        if (upperPosition < 1) {
            upperPosition = 1;
        } else if (upperPosition > n) {
            upperPosition = n;
        }
        const upperVar = sortedReturns[upperPosition - 1];
        const upperTail = returns.filter(r => r >= upperVar);
        if (upperTail.length === 0 || esLower === 0) {
            return NaN;
        }
        const esUpper = pySum(upperTail) / upperTail.length;
        return esUpper / esLower;
    }

    /** Per-observation high-water-mark drawdowns (PerformanceAnalytics `Drawdowns()`), a new array. */
    get drawdownsCumulative(): number[] {
        return this._drawdownHighWatermark.drawdowns.slice();
    }

    /** Minimum (most negative) cumulative drawdown. O(n). */
    get minDrawdownsCumulative(): number {
        return this._drawdownHighWatermark.maximumDrawdown;
    }

    /** Magnitude of the worst cumulative drawdown. O(n). */
    get worstDrawdownsCumulative(): number {
        return Math.abs(this._drawdownHighWatermark.maximumDrawdown);
    }

    /** Per-observation high-water-mark drawdowns, a new array. */
    get drawdownsHighWatermark(): number[] {
        return this._drawdownHighWatermark.drawdowns.slice();
    }

    /**
     * Continuous drawdowns (runs of consecutive losses, R percent convention).
     * When `maxRuns > 0` they are sorted worst-first and truncated.
     */
    drawdownsContinuousRuns(maxRuns: number | null = null): number[] {
        let drawdowns = this._drawdownContinuousRuns.drawdowns;
        if (drawdowns.length < 1) {
            return [];
        }
        if (maxRuns !== null && maxRuns !== undefined && maxRuns > 0) {
            drawdowns = drawdowns.slice().sort((a, b) => a - b);
            drawdowns = drawdowns.slice(0, maxRuns);
        }
        return drawdowns;
    }

    /** Calmar ratio: geometric mean return / |maximum drawdown| (not annualized). */
    get calmarRatio(): number {
        const wdd = this.worstDrawdownsCumulative;
        if (wdd === 0) {
            return NaN;
        }
        const cagr = this._cumulativeReturn.geometricMeanReturn;
        if (Number.isNaN(cagr)) {
            return NaN;
        }
        return cagr / wdd;
    }

    /** Sterling ratio: geometric mean return / (|maximum drawdown| + excess). */
    sterlingRatio(excess: number = 0.1): number {
        const wdd = this.worstDrawdownsCumulative + excess;
        if (wdd === 0) {
            return NaN;
        }
        const cagr = this._cumulativeReturn.geometricMeanReturn;
        if (Number.isNaN(cagr)) {
            return NaN;
        }
        return cagr / wdd;
    }

    /** Burke ratio (Gm - rf) / sqrt(Σ continuous drawdowns²). */
    get burkeRatio(): number {
        const rate = this._cumulativeReturn.geometricMeanReturn - this.riskFreeRate;
        if (Number.isNaN(rate)) {
            return NaN;
        }
        const sqrtSumDrawdownsSquared = this._drawdownContinuousRuns.sqrtSumDrawdownsSquared;
        if (sqrtSumDrawdownsSquared === 0) {
            return NaN;
        }
        return rate / sqrtSumDrawdownsSquared;
    }

    /** Modified Burke ratio: Burke ratio * sqrt(n). */
    get burkeRatioModified(): number {
        const burke = this.burkeRatio;
        if (Number.isNaN(burke)) {
            return NaN;
        }
        return burke * Math.sqrt(this._returnsKbn.n);
    }

    /** Pain index: mean drawdown magnitude. */
    get painIndex(): number {
        return -this._drawdownHighWatermark.drawdownsMean;
    }

    /** Pain ratio (Gm - rf) / pain index. */
    get painRatio(): number {
        const rate = this._cumulativeReturn.geometricMeanReturn - this.riskFreeRate;
        if (Number.isNaN(rate)) {
            return NaN;
        }
        const painIndex = this.painIndex;
        return painIndex !== 0 ? rate / painIndex : NaN;
    }

    /** Ulcer index: root-mean-square drawdown. */
    get ulcerIndex(): number {
        return Math.sqrt(this._drawdownHighWatermark.drawdownsSquaredMean);
    }

    /** Martin ratio (Gm - rf) / ulcer index. */
    get martinRatio(): number {
        const rate = this._cumulativeReturn.geometricMeanReturn - this.riskFreeRate;
        if (Number.isNaN(rate)) {
            return NaN;
        }
        const ulcerIndex = this.ulcerIndex;
        return ulcerIndex !== 0 ? rate / ulcerIndex : NaN;
    }

    /** Average drawdown episode depth (0.0 when there are none). */
    get drawdownAverage(): number {
        return this._drawdownEpisodes.averageEpisodeDrawdown;
    }

    /** Average drawdown episode length (0.0 when there are none). */
    get drawdownAverageLength(): number {
        return this._drawdownEpisodes.averageEpisodeLength;
    }

    /** Average drawdown episode peak-to-trough length (0.0 when there are none). */
    get drawdownAveragePeakToTrough(): number {
        return this._drawdownEpisodes.averageEpisodePeakToTrough;
    }

    /** Average drawdown episode recovery length (0.0 when there are none). */
    get drawdownAverageRecovery(): number {
        return this._drawdownEpisodes.averageEpisodeRecovery;
    }

    /** Drawdown deviation sqrt(Σdepth² / n). */
    get drawdownDeviation(): number {
        return Math.sqrt(this._drawdownEpisodes.averageEpisodeDrawdownSquared);
    }

    /**
     * Conditional drawdown at risk over the continuous drawdown path.
     *
     * @throws Error if confidence is not in (0, 1).
     */
    cdarAverage(confidence: number = 0.95): number {
        if (!(0.0 < confidence && confidence < 1.0)) {
            throw new Error('confidence must be between 0 and 1');
        }
        const drawdowns = this._drawdownHighWatermark.drawdowns;
        if (drawdowns.length === 0) {
            return 0;
        }
        const q = percentile(drawdowns, 1.0 - confidence);
        if (q >= 0.0) {
            return 0;
        }
        const tailSum = new KleinKbnAccumulator();
        let tailLen = 0;
        for (const dd of drawdowns) {
            if (dd <= q) {
                tailLen += 1;
                tailSum.update(dd);
            }
        }
        return tailLen > 0 ? -tailSum.value / tailLen : 0;
    }

    /**
     * Conditional drawdown at risk over the drawdown episode depths
     * (PerformanceAnalytics default).
     *
     * @throws Error if confidence is not in (0, 1).
     */
    cdarDiscrete(confidence: number = 0.95): number {
        if (!(0.0 < confidence && confidence < 1.0)) {
            throw new Error('confidence must be between 0 and 1');
        }
        const depths = this._drawdownEpisodes.depths;
        if (depths.length === 0) {
            return 0;
        }
        const q = percentile(depths, 1.0 - confidence);
        const tailSum = new KleinKbnAccumulator();
        let tailLen = 0;
        for (const depth of depths) {
            if (depth <= q) {
                tailLen += 1;
                tailSum.update(depth);
            }
        }
        return tailLen > 0 ? -tailSum.value / tailLen : 0;
    }

    /**
     * CDaR beta: portfolio returns over the worst benchmark drawdown
     * episodes, relative to the benchmark's conditional drawdown.
     *
     * @throws Error if confidence is not in (0, 1).
     */
    cdarBeta(confidence: number = 0.95): number {
        if (!(0.0 < confidence && confidence < 1.0)) {
            throw new Error('confidence must be between 0 and 1');
        }
        const w = this._returns;
        const episodes = this._drawdownEpisodesBenchmark.episodes;
        const depths = this._drawdownEpisodesBenchmark.depths;
        if (depths.length === 0) {
            return NaN;
        }
        const tailCount = Math.max(1, Math.ceil(depths.length * (1.0 - confidence)));
        const q = depths.slice().sort((a, b) => a - b)[tailCount - 1];
        if (q === 0.0) {
            return NaN;
        }
        const sumRet = new KleinKbnAccumulator();
        const ret = new KleinKbnAccumulator();
        let tailLen = 0;
        for (const episode of episodes) {
            if (episode.depth <= q) {
                tailLen += 1;
                ret.reset();
                for (let i = episode.fromIdx; i < episode.troughIdx + 1; i++) {
                    ret.update(Math.log1p(w[i]));
                }
                sumRet.update(Math.expm1(ret.value));
            }
        }
        return tailLen !== 0 ? sumRet.value / (tailLen * q) : NaN;
    }

    /**
     * CDaR alpha: annualized mean return less CDaR beta times the annualized
     * mean benchmark return.
     *
     * @throws Error if confidence is not in (0, 1).
     */
    cdarAlpha(confidence: number = 0.95): number {
        const beta = this.cdarBeta(confidence);
        if (Number.isNaN(beta)) {
            return NaN;
        }
        const period = this.periodsPerAnnum;
        const rMean = this._returnsKbn.mean;
        const bMean = this._benchmarkReturnsKbn.mean;
        const rAnnual = (1.0 + rMean) ** period - 1.0;
        const bAnnual = (1.0 + bMean) ** period - 1.0;
        return rAnnual - beta * bAnnual;
    }

    /**
     * Geometric mean return divided by the mean magnitude of the worst
     * `max(1, int(n * (1 - confidence)))` drawdowns. No confidence validation.
     */
    rewardToConditionalDrawdown(confidence: number = 0.95): number {
        const cagr = this._cumulativeReturn.geometricMeanReturn;
        if (Number.isNaN(cagr)) {
            return NaN;
        }
        const dd = this.drawdownsCumulative;
        if (dd.length < 1) {
            return NaN;
        }
        const nTail = Math.max(1, pyInt(dd.length * (1 - confidence)));
        const sortedDd = dd.sort((a, b) => a - b);
        const sortedTail = sortedDd.slice(0, nTail);
        const cdar = -pySum(sortedTail) / sortedTail.length;
        return cdar !== 0 ? cagr / cdar : NaN;
    }

    /** Single-factor-model risk premium: mean excess return. */
    get sfmRiskPremium(): number {
        return this._excessReturnsKbn.mean;
    }

    /** Single-factor-model alpha (periodic). */
    get sfmAlpha(): number {
        return this._sfmRegression.alpha;
    }

    /** Single-factor-model beta. */
    get sfmBeta(): number {
        return this._sfmRegression.beta;
    }

    /** Single-factor-model beta over bull (benchmark excess > 0) periods. */
    get sfmBetaBull(): number {
        return this._sfmRegression.betaBull;
    }

    /** Single-factor-model beta over bear (benchmark excess < 0) periods. */
    get sfmBetaBear(): number {
        return this._sfmRegression.betaBear;
    }

    /** Timing ratio beta_bull / beta_bear. */
    get timingRatio(): number {
        const denom = this._sfmRegression.betaBear;
        return denom !== 0 ? this._sfmRegression.betaBull / denom : NaN;
    }

    /** Single-factor-model R². */
    get sfmR2(): number {
        return this._sfmRegression.r2;
    }

    /** Jensen's alpha (annual): Rp_ann - (beta * Rb_ann + (1 - beta) * rf_annual). */
    get jensenAlpha(): number {
        const rf = this._annualRiskFreeRate;
        const mean = this._cumulativeReturn.annualizedGeometricMeanReturn(this.periodsPerAnnum);
        const meanB = this._benchmarkCumulativeReturn.annualizedGeometricMeanReturn(this.periodsPerAnnum);
        const beta = this.sfmBeta;
        return mean - (beta * meanB + (1.0 - beta) * rf);
    }

    /** Fama beta sigma0(r) / sigma0(b). */
    get famaBeta(): number {
        const sigma = this._returnsKbn.standardDeviationDdof0;
        const sigmaB = this._benchmarkReturnsKbn.standardDeviationDdof0;
        return sigmaB !== 0 ? sigma / sigmaB : NaN;
    }

    /** Modigliani-Modigliani measure (periodic): rf + mean(e) * sigma0(b) / sigma0(e). */
    get modigliani(): number {
        const sigma = this._excessReturnsKbn.standardDeviationDdof0;
        if (sigma === 0) {
            return NaN;
        }
        const sigmaB = this._benchmarkReturnsKbn.standardDeviationDdof0;
        return this.riskFreeRate + this._excessReturnsKbn.mean * sigmaB / sigma;
    }

    /** Annualized tracking error sigma1(r - b) * sqrt(P). */
    get trackingError(): number {
        return this._activeReturnsKbn.standardDeviationDdof1 * this._sqrtPeriodsPerAnnum;
    }

    /** Active premium: annualized portfolio return less annualized benchmark return. */
    get activePremium(): number {
        const mean = this._cumulativeReturn.annualizedGeometricMeanReturn(this.periodsPerAnnum);
        const meanB = this._benchmarkCumulativeReturn.annualizedGeometricMeanReturn(this.periodsPerAnnum);
        return mean - meanB;
    }

    /** Information ratio: active premium / tracking error. */
    get informationRatio(): number {
        const te = this.trackingError;
        return te !== 0 ? this.activePremium / te : NaN;
    }

    /** Information ratio with its sign taken from the arithmetic mean active return. */
    get informationRatioModified(): number {
        const excess = this._activeReturnsKbn.mean;
        const ir = this.informationRatio;
        if (Number.isNaN(excess) || Number.isNaN(ir)) {
            return NaN;
        }
        return excess > 0 ? ir : -ir;
    }

    /** Annualized systematic risk |beta| * sigma1(b - rf) * sqrt(P). */
    get systematicRisk(): number {
        const beta = this.sfmBeta;
        if (Number.isNaN(beta)) {
            return NaN;
        }
        const benchmarkRisk = this._benchmarkExcessReturnsKbn.standardDeviationDdof1;
        if (Number.isNaN(benchmarkRisk)) {
            return NaN;
        }
        return Math.abs(beta) * benchmarkRisk * this._sqrtPeriodsPerAnnum;
    }

    /** Treynor ratio: annualized excess return / beta. */
    get treynorRatio(): number {
        const beta = this.sfmBeta;
        if (beta === 0) {
            return NaN;
        }
        return this._cumulativeExcessReturn.annualizedGeometricMeanReturn(this.periodsPerAnnum) / beta;
    }

    /** Modified Treynor ratio: annualized excess return / systematic risk. */
    get treynorRatioModified(): number {
        const sr = this.systematicRisk;
        if (sr === 0) {
            return NaN;
        }
        return this._cumulativeExcessReturn.annualizedGeometricMeanReturn(this.periodsPerAnnum) / sr;
    }

    /** Annualized specific (idiosyncratic) risk: sigma0 of the SFM residuals * sqrt(P). O(n). */
    get specificRisk(): number {
        const rP = this._returns;
        const rB = this._returnsBenchmark;
        const beta = this.sfmBeta;
        if (Number.isNaN(beta)) {
            return NaN;
        }
        const alpha = this.sfmAlpha;
        if (Number.isNaN(alpha)) {
            return NaN;
        }
        const epsilonKbn = new RawMomentsKleinKbn(0, true, true);
        const rf = this.riskFreeRate;
        for (let i = 0; i < rP.length; i++) {
            epsilonKbn.update(rP[i] - rf - alpha - beta * (rB[i] - rf));
        }
        return epsilonKbn.standardDeviationDdof0 * this._sqrtPeriodsPerAnnum;
    }

    /** Annualized total risk sqrt(systematic² + specific²). */
    get totalRisk(): number {
        const syr = this.systematicRisk;
        if (Number.isNaN(syr)) {
            return NaN;
        }
        const spr = this.specificRisk;
        if (Number.isNaN(spr)) {
            return NaN;
        }
        return Math.sqrt(syr * syr + spr * spr);
    }

    /** Appraisal ratio: Jensen's alpha / specific risk. */
    get appraisalRatio(): number {
        const alpha = this.jensenAlpha;
        if (Number.isNaN(alpha)) {
            return NaN;
        }
        const spr = this.specificRisk;
        return spr !== 0 ? alpha / spr : NaN;
    }

    /** Modified Jensen's alpha: Jensen's alpha / beta. */
    get jensenAlphaModified(): number {
        const alpha = this.jensenAlpha;
        if (Number.isNaN(alpha)) {
            return NaN;
        }
        const beta = this.sfmBeta;
        return beta !== 0 ? alpha / beta : NaN;
    }

    /** Alternative Jensen's alpha: Jensen's alpha / systematic risk. */
    get jensenAlphaAlternative(): number {
        const alpha = this.jensenAlpha;
        if (Number.isNaN(alpha)) {
            return NaN;
        }
        const spr = this.systematicRisk;
        return spr !== 0 ? alpha / spr : NaN;
    }

    /** M² (annualized): Rp_ann * sigma_b / sigma_p + rf_annual * (1 - sigma_b / sigma_p). */
    get mSquared(): number {
        const pRet = this._cumulativeReturn.annualizedGeometricMeanReturn(this.periodsPerAnnum);
        if (Number.isNaN(pRet)) {
            return NaN;
        }
        const pStd = this._returnsKbn.standardDeviationDdof0 * this._sqrtPeriodsPerAnnum;
        if (Number.isNaN(pStd) || pStd === 0) {
            return NaN;
        }
        const bStd = this._benchmarkReturnsKbn.standardDeviationDdof0 * this._sqrtPeriodsPerAnnum;
        if (Number.isNaN(bStd)) {
            return NaN;
        }
        const scale = bStd / pStd;
        return pRet * scale + this._annualRiskFreeRate * (1.0 - scale);
    }

    /** M² excess (geometric): (1 + M²) / (1 + Rb_ann) - 1. */
    get mSquaredExcess(): number {
        const mSq = this.mSquared;
        if (Number.isNaN(mSq)) {
            return NaN;
        }
        const bRet = this._benchmarkCumulativeReturn.annualizedGeometricMeanReturn(this.periodsPerAnnum);
        if (Number.isNaN(bRet)) {
            return NaN;
        }
        return (1.0 + mSq) / (1.0 + bRet) - 1.0;
    }

    /** Sortino M²: Rp_ann + Sortino * sqrt(P) * (DD_b - DD_p). */
    get mSquaredSortino(): number {
        const sortino = this.sortinoRatio;
        if (Number.isNaN(sortino)) {
            return NaN;
        }
        const pRet = this._cumulativeReturn.annualizedGeometricMeanReturn(this.periodsPerAnnum);
        if (Number.isNaN(pRet)) {
            return NaN;
        }
        const pDd = this.downsideDeviation;
        if (Number.isNaN(pDd)) {
            return NaN;
        }
        const bCount = this._benchmarkTargetPartialMoments.totalCount;
        if (bCount === 0) {
            return NaN;
        }
        const bDd = Math.sqrt(this._benchmarkTargetPartialMoments.lowerExcessMoment2Sum / bCount);
        if (Number.isNaN(bDd)) {
            return NaN;
        }
        return pRet + sortino * this._sqrtPeriodsPerAnnum * (bDd - pDd);
    }

    /**
     * Tail ratio: the `cutoff` percentile divided by the magnitude of the
     * `1 - cutoff` percentile.
     *
     * @throws Error unless 0.5 < cutoff < 1.
     */
    tailRatio(cutoff: number = 0.95): number {
        if (!(0.5 < cutoff && cutoff < 1.0)) {
            throw new Error('cutoff must be between 0.5 and 1.0');
        }
        const w = this._returns;
        if (w.length < 2) {
            return NaN;
        }
        const rightTail = percentile(w, cutoff);
        const leftTail = percentile(w, 1 - cutoff);
        return leftTail !== 0 ? rightTail / Math.abs(leftTail) : NaN;
    }

    /** Full Kelly ratio mean(e) / Var1(e). */
    get kellyRatioFull(): number {
        const meanExcess = this._excessReturnsKbn.mean;
        const varExcess = this._excessReturnsKbn.variance;
        return varExcess !== 0 ? meanExcess / varExcess : NaN;
    }

    /** Half Kelly ratio. */
    get kellyRatio(): number {
        return this.kellyRatioFull / 2;
    }

    /** Single-scale rescaled-range Hurst exponent log(R/S) / log(n). O(n). */
    get hurstExponent(): number {
        const n = this._returnsKbn.n;
        const w = this._returns;
        if (n < 2) {
            return NaN;
        }
        const mean = this._returnsKbn.mean;
        const std = this._returnsKbn.standardDeviationDdof1;
        if (std === 0) {
            return NaN;
        }
        const cumSum = new KleinKbnAccumulator();
        let cumMin = Infinity;
        let cumMax = -Infinity;
        for (const x of w) {
            cumSum.update(x - mean);
            const val = cumSum.value;
            if (cumMin > val) {
                cumMin = val;
            }
            if (cumMax < val) {
                cumMax = val;
            }
        }
        const delta = cumMax - cumMin;
        const rescaledRange = delta / std;
        if (rescaledRange <= 0) {
            return NaN;
        }
        return Math.log(rescaledRange) / Math.log(n);
    }

    /**
     * Bias ratio: count of returns in [0, k·σ] over 1 + count in [-k·σ, 0). O(n).
     *
     * @throws Error if the multiplier is not positive.
     */
    biasRatio(stdDevMultiplier: number = 1.0): number {
        if (stdDevMultiplier <= 0) {
            throw new Error('std_dev_multiplier must be positive');
        }
        const w = this._returns;
        const std = this._returnsKbn.standardDeviationDdof1;
        if (Number.isNaN(std) || std === 0) {
            return NaN;
        }
        const threshold = stdDevMultiplier * std;
        let countPositive = 0;
        let countNegative = 0;
        for (const x of w) {
            if (0 <= x && x <= threshold) {
                countPositive += 1;
            } else if (-threshold <= x && x < 0) {
                countNegative += 1;
            }
        }
        return countPositive / (1 + countNegative);
    }

    /** Kestner K-ratio of the log-equity curve (plain sums). O(n). */
    get kRatio(): number {
        const n = this._returnsKbn.n;
        const w = this._returns;
        if (n < 3) {
            return NaN;
        }
        const equity: number[] = [];
        let cumSum = 0.0;
        for (const x of w) {
            cumSum += Math.log1p(x);
            equity.push(cumSum);
        }
        let sumT = 0.0;
        let sumT2 = 0.0;
        let sumEq = 0.0;
        let sumTe = 0.0;
        for (let i = 0; i < equity.length; i++) {
            const tVal = i;
            const eqVal = equity[i];
            sumT += tVal;
            sumT2 += tVal * tVal;
            sumEq += eqVal;
            sumTe += tVal * eqVal;
        }
        const tMean = sumT / n;
        const equityMean = sumEq / n;
        const sTt = sumT2 - n * (tMean * tMean);
        const sTe = sumTe - n * tMean * equityMean;
        if (sTt === 0) {
            return NaN;
        }
        const slope = sTe / sTt;
        const intercept = equityMean - slope * tMean;
        let sumSqResiduals = 0.0;
        for (let i = 0; i < equity.length; i++) {
            const predicted = intercept + slope * i;
            const residual = equity[i] - predicted;
            sumSqResiduals += residual * residual;
        }
        if (n <= 2) {
            return NaN;
        }
        let residualVar = sumSqResiduals / (n - 2);
        if (residualVar < 0) {
            residualVar = 0.0;
        }
        const seSlope = Math.sqrt(residualVar / sTt);
        if (seSlope === 0) {
            return NaN;
        }
        return slope / (seSlope * Math.sqrt(n));
    }

    /**
     * Gain-to-pain ratio: mean return divided by the raw lower partial
     * moment (the sum of losses, as in the Python implementation).
     */
    get gainToPainRatio(): number {
        const lpm1 = this._rawPartialMoments.lowerPartialMoment1;
        if (Number.isNaN(lpm1) || lpm1 === 0) {
            return NaN;
        }
        return this._returnsKbn.mean / lpm1;
    }

    /** Upside capture ratio (geometric or arithmetic). */
    upsideCaptureRatio(geometric: boolean = true): number {
        if (geometric === true) {
            return this._capture.upsideCaptureRatioGeometric;
        }
        return this._capture.upsideCaptureRatioArithmetic;
    }

    /** Downside capture ratio (geometric or arithmetic). */
    downsideCaptureRatio(geometric: boolean = true): number {
        if (geometric === true) {
            return this._capture.downsideCaptureRatioGeometric;
        }
        return this._capture.downsideCaptureRatioArithmetic;
    }

    /** Overall capture ratio: upside capture / downside capture. */
    overallCaptureRatio(geometric: boolean = true): number {
        const up = this.upsideCaptureRatio(geometric);
        const down = this.downsideCaptureRatio(geometric);
        if (Number.isNaN(up) || Number.isNaN(down) || down === 0) {
            return NaN;
        }
        return up / down;
    }

    /** Fraction of up-benchmark periods (b > 0) where the portfolio is also up. */
    get upNumberRatio(): number {
        return this._capture.upNumberRatio;
    }

    /** Fraction of down-benchmark periods (b <= 0) where the portfolio is also down. */
    get downNumberRatio(): number {
        return this._capture.downNumberRatio;
    }

    /** Fraction of up-benchmark periods where the portfolio outperforms. */
    get upPercentageRatio(): number {
        return this._capture.upPercentageRatio;
    }

    /** Fraction of down-benchmark periods (b < 0) where the portfolio outperforms. */
    get downPercentageRatio(): number {
        return this._capture.downPercentageRatio;
    }
}
