import math
import collections

from ...streaming_kbn import RawMomentsKleinKBN

from .norm import norm_ppf, norm_pdf
from .var import var_historical

def es_historical(returns: collections.deque, risk_free_rate:float = 0.0, confidence: float = 0.95) -> float:
    if returns is None or len(returns) == 0:
        return math.nan
    var = var_historical(returns=returns, risk_free_rate=risk_free_rate, confidence=confidence)
    if math.isnan(var):
        return math.nan

    sum_tail = 0.0
    count = 0
    for r in returns:
        excess = r - risk_free_rate
        if excess <= -var:
            sum_tail += excess
            count += 1

    return -sum_tail / count if count != 0 else math.nan

def es_gaussian(returns_kbn: RawMomentsKleinKBN, confidence: float = 0.95) -> float:
    mean = returns_kbn.mean
    std = returns_kbn.standard_deviation_ddof_0
    if math.isnan(std):
        return math.nan
    z = norm_ppf(confidence)
    phi_z = norm_pdf(z)
    return -mean + phi_z * std / (1 - confidence)

def es_cornish_fisher(returns_kbn: RawMomentsKleinKBN, confidence: float = 0.95) -> float:
    alpha = 1.0 - confidence
    z = norm_ppf(alpha)
    mean = returns_kbn.mean
    sigma = returns_kbn.standard_deviation_ddof_0
    skew = returns_kbn.skewness_moment # bias=True
    kurtosis = returns_kbn.kurtosis_excess # bias=True, fisher=True
    # Skewness and kurtosis are unavailable for very small samples.
    # Fall back to Gaussian ES.
    if math.isnan(skew) or math.isnan(kurtosis):
        return es_gaussian(returns_kbn=returns_kbn, confidence=confidence)
    z2 = z * z
    z3 = z2 * z
    h = (z
        + (z2 - 1) * skew / 6
        + (z3 - 3*z) * kurtosis / 24
        - (2*z3 - 5*z) * skew * skew / 36)
    h2 = h * h
    h4 = h2 * h2
    mes = (norm_pdf(h) * (1
        + h2*h * skew / 6
        + (h4*h2 - 9*h4 + 9*h2 + 3) * skew * skew / 72
        + (h4 - 2*h2 - 1) * kurtosis / 24))
    return -mean - sigma * min(-mes / alpha, h)
