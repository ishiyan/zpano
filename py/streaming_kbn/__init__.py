"""
Streaming (one-pass, O(1) per sample) statistics with Klein second-order
Kahan-Babuška-Neumaier (KBN) compensated summation.

Uses only the Python standard library, so the code can be ported to other
languages verbatim.

- KleinKBNAccumulator: compensated sum; supports set() and revert().
- KleinKBNSummator: compensated sum plus sample count and mean.
- RawMomentsKleinKBN: mean, variance, skewness and kurtosis from raw power
  sums; revert() removes any previously added sample (FIFO rolling windows).
- CentralMomentsKleinKBN: mean, variance, skewness and kurtosis from Pébay's
  central moment updates; more accurate for data with a large mean, but
  revert() only removes the most recent sample (LIFO).
- LinearRegressionKleinKBN: OLS slope, intercept, correlation and covariance;
  revert() removes any previously added sample.
"""

from .klein_kbn_accumulator import KleinKBNAccumulator
from .klein_kbn_summator import KleinKBNSummator
from .raw_moments_klein_kbn import RawMomentsKleinKBN
from .central_moments_klein_kbn import CentralMomentsKleinKBN
from .linear_regression_klein_kbn import LinearRegressionKleinKBN
