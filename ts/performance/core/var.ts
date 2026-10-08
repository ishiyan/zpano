import { RawMomentsKleinKbn } from '../../streaming-kbn';

import { percentile } from './percentile';
import { normPpf } from './norm';

/**
 * Historical Value-at-Risk: the negated (1 - confidence) percentile of
 * the excess returns (returns less the risk-free rate).
 *
 * Returns NaN when `returns` is null, undefined or empty.
 * The input array is not modified.
 */
export function varHistorical(returns: readonly number[] | null | undefined,
    riskFreeRate: number = 0.0, confidence: number = 0.95): number {
    const w = returns;
    if (w === null || w === undefined || w.length < 1) {
        return NaN;
    }
    const q = 1 - confidence;
    if (riskFreeRate === 0) {
        return -percentile(w, q);
    }
    return -percentile(w.map(r => r - riskFreeRate), q);
}

/**
 * Gaussian (parametric) Value-at-Risk from the mean and the population
 * (ddof=0) standard deviation of the returns.
 *
 * Returns NaN when the standard deviation is unavailable.
 */
export function varGaussian(returnsKbn: RawMomentsKleinKbn, confidence: number = 0.95): number {
    const mean = returnsKbn.mean;
    const std = returnsKbn.standardDeviationDdof0;
    if (Number.isNaN(std)) {
        return NaN;
    }
    const z = normPpf(1 - confidence);
    return -(mean + z * std);
}

/**
 * Modified (Cornish-Fisher) Value-at-Risk: the Gaussian z-score is adjusted
 * with the biased skewness and biased excess kurtosis of the returns.
 *
 * Falls back to Gaussian VaR when skewness or kurtosis is unavailable
 * (very small samples). Returns NaN when the standard deviation is unavailable.
 */
export function varCornishFisher(returnsKbn: RawMomentsKleinKbn, confidence: number = 0.95): number {
    const mean = returnsKbn.mean;
    const std = returnsKbn.standardDeviationDdof0;
    if (Number.isNaN(std)) {
        return NaN;
    }
    // Cornish-Fisher expansion for z-score adjustment
    let z = normPpf(1 - confidence);
    const skew = returnsKbn.skewnessMoment; // bias=True
    const kurtosis = returnsKbn.kurtosisExcess; // bias=True, fisher=True
    // Skewness and kurtosis are unavailable for very small samples.
    // Fall back to Gaussian VaR.
    if (Number.isNaN(skew) || Number.isNaN(kurtosis)) {
        return -(mean + z * std);
    }
    // Cornish-Fisher expansion
    const z2 = z * z;
    const z3 = z2 * z;
    z = (z
        + (z2 - 1) * skew / 6
        + (z3 - 3 * z) * kurtosis / 24
        - (2 * z3 - 5 * z) * skew * skew / 36);
    return -(mean + z * std);
}
