import math

from ...streaming_kbn import RawMomentsKleinKBN

from .norm import norm_cdf

def probabilistic_sharpe_ratio(returns_kbn: RawMomentsKleinKBN, sr: float, reference_sr: float = 0.0, 
        zero_skewness: bool = False, normal_kurtosis: bool = True) -> float:
    if math.isnan(sr):
        return math.nan
    if zero_skewness:
        skewness = 0
    else:
        skewness = returns_kbn.skewness_moment # or _sample 
        if math.isnan(skewness):
            return math.nan
    if normal_kurtosis:
        kurtosis = 3  # excess kurtosis = 0, so K = 3
    else:
        kurtosis = returns_kbn.kurtosis_excess
        if math.isnan(kurtosis):
            return math.nan
        kurtosis += 3  # convert to regular kurtosis
    
    denom = math.sqrt(1 - sr * skewness + (sr*sr) * (kurtosis - 1) / 4)
    if denom == 0:
        return math.nan
    
    z = (sr - reference_sr) * math.sqrt(returns_kbn.n - 1) / denom
    return norm_cdf(z)
