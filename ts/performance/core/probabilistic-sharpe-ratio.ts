import { RawMomentsKleinKbn } from '../../streaming-kbn';

import { normCdf } from './norm';

/**
 * Probabilistic Sharpe Ratio: the probability that the true Sharpe ratio
 * exceeds `referenceSr`, given the observed Sharpe ratio `sr` and the
 * sample size, skewness and kurtosis of the returns.
 *
 * @param returnsKbn Raw moments of the returns.
 * @param sr Observed (per-period) Sharpe ratio.
 * @param referenceSr Benchmark Sharpe ratio.
 * @param zeroSkewness Assume zero skewness instead of the biased sample skewness.
 * @param normalKurtosis Assume normal kurtosis (3) instead of the biased sample kurtosis.
 * @returns NaN when `sr`, the required moments, or the denominator is unavailable.
 */
export function probabilisticSharpeRatio(returnsKbn: RawMomentsKleinKbn, sr: number,
    referenceSr: number = 0.0, zeroSkewness: boolean = false, normalKurtosis: boolean = true): number {
    if (Number.isNaN(sr)) {
        return NaN;
    }
    let skewness: number;
    if (zeroSkewness) {
        skewness = 0;
    } else {
        skewness = returnsKbn.skewnessMoment; // or skewnessSample
        if (Number.isNaN(skewness)) {
            return NaN;
        }
    }
    let kurtosis: number;
    if (normalKurtosis) {
        kurtosis = 3; // excess kurtosis = 0, so K = 3
    } else {
        kurtosis = returnsKbn.kurtosisExcess;
        if (Number.isNaN(kurtosis)) {
            return NaN;
        }
        kurtosis += 3; // convert to regular kurtosis
    }

    const denom = Math.sqrt(1 - sr * skewness + (sr * sr) * (kurtosis - 1) / 4);
    if (denom === 0) {
        return NaN;
    }

    const z = (sr - referenceSr) * Math.sqrt(returnsKbn.n - 1) / denom;
    return normCdf(z);
}
