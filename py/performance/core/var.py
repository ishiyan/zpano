import math
import collections

from ...streaming_kbn import RawMomentsKleinKBN

from .percentile import percentile
from .norm import norm_ppf

def var_historical(returns: collections.deque, risk_free_rate:float = 0.0, confidence: float = 0.95) -> float:
    w = returns
    if w is None or len(w) < 1:
        return math.nan
    q = 1 - confidence
    if risk_free_rate == 0:
        return -percentile(w, q)
    return -percentile((r - risk_free_rate for r in w), q)

def var_gaussian(returns_kbn: RawMomentsKleinKBN, confidence: float = 0.95) -> float:
    mean = returns_kbn.mean
    std = returns_kbn.standard_deviation_ddof_0
    if math.isnan(std):
        return math.nan
    z = norm_ppf(1 - confidence)
    return -(mean + z * std)

def var_cornish_fisher(returns_kbn: RawMomentsKleinKBN, confidence: float = 0.95) -> float:
    mean = returns_kbn.mean
    std = returns_kbn.standard_deviation_ddof_0
    if math.isnan(std):
        return math.nan
    # Cornish-Fisher expansion for z-score adjustment
    z = norm_ppf(1 - confidence)
    skew = returns_kbn.skewness_moment # bias=True
    kurtosis = returns_kbn.kurtosis_excess # bias=True, fisher=True
    # Skewness and kurtosis are unavailable for very small samples.
    # Fall back to Gaussian VaR.
    if math.isnan(skew) or math.isnan(kurtosis):
        return -(mean + z * std)
    # Cornish-Fisher expansion
    z2 = z * z
    z3 = z2 * z
    z = (z 
            + (z2 - 1) * skew / 6 
            + (z3 - 3*z) * kurtosis / 24 
            - (2*z3 - 5*z) * skew*skew / 36)
    return -(mean + z * std)
