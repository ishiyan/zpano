import { RawMomentsKleinKbn } from '../../streaming-kbn';

import { normPpf, normPdf } from './norm';
import { varHistorical } from './var';

/**
 * Historical Expected Shortfall: the negated mean of the excess returns
 * (returns less the risk-free rate) at or below the historical VaR.
 *
 * Returns NaN when `returns` is null, undefined or empty, or when the
 * tail is empty. The input array is not modified.
 */
export function esHistorical(returns: readonly number[] | null | undefined,
    riskFreeRate: number = 0.0, confidence: number = 0.95): number {
    if (returns === null || returns === undefined || returns.length === 0) {
        return NaN;
    }
    const v = varHistorical(returns, riskFreeRate, confidence);
    if (Number.isNaN(v)) {
        return NaN;
    }

    let sumTail = 0.0;
    let count = 0;
    for (const r of returns) {
        const excess = r - riskFreeRate;
        if (excess <= -v) {
            sumTail += excess;
            count += 1;
        }
    }

    return count !== 0 ? -sumTail / count : NaN;
}

/**
 * Gaussian (parametric) Expected Shortfall from the mean and the population
 * (ddof=0) standard deviation of the returns.
 *
 * Returns NaN when the standard deviation is unavailable.
 */
export function esGaussian(returnsKbn: RawMomentsKleinKbn, confidence: number = 0.95): number {
    const mean = returnsKbn.mean;
    const std = returnsKbn.standardDeviationDdof0;
    if (Number.isNaN(std)) {
        return NaN;
    }
    const z = normPpf(confidence);
    const phiZ = normPdf(z);
    return -mean + phiZ * std / (1 - confidence);
}

/**
 * Modified (Cornish-Fisher) Expected Shortfall, using the biased skewness
 * and biased excess kurtosis of the returns.
 *
 * Falls back to Gaussian ES when skewness or kurtosis is unavailable
 * (very small samples).
 */
export function esCornishFisher(returnsKbn: RawMomentsKleinKbn, confidence: number = 0.95): number {
    const alpha = 1.0 - confidence;
    const z = normPpf(alpha);
    const mean = returnsKbn.mean;
    const sigma = returnsKbn.standardDeviationDdof0;
    const skew = returnsKbn.skewnessMoment; // bias=True
    const kurtosis = returnsKbn.kurtosisExcess; // bias=True, fisher=True
    // Skewness and kurtosis are unavailable for very small samples.
    // Fall back to Gaussian ES.
    if (Number.isNaN(skew) || Number.isNaN(kurtosis)) {
        return esGaussian(returnsKbn, confidence);
    }
    const z2 = z * z;
    const z3 = z2 * z;
    const h = (z
        + (z2 - 1) * skew / 6
        + (z3 - 3 * z) * kurtosis / 24
        - (2 * z3 - 5 * z) * skew * skew / 36);
    const h2 = h * h;
    const h4 = h2 * h2;
    const mes = (normPdf(h) * (1
        + h2 * h * skew / 6
        + (h4 * h2 - 9 * h4 + 9 * h2 + 3) * skew * skew / 72
        + (h4 - 2 * h2 - 1) * kurtosis / 24));
    // Python min(a, b) semantics: b is chosen only if b < a.
    const tail = -mes / alpha;
    return -mean - sigma * (h < tail ? h : tail);
}
