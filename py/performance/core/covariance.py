import math

from ...streaming_kbn import RawMomentsKleinKBN

class Covariance:
    """
    Streaming covariance using Welford's algorithm.
    """
    def __init__(self, ddof: int, threshold: float) -> None:
        """
        Streaming covariance using Welford's algorithm.

        Args:
            ddof: Delta Degrees of Freedom (1 for sample, 0 for population).
            threshold: Reference return (in same periodicity as returns) used to form excess returns.
                       This can be target return, minimum acceptable return (MAR),
                       or risk-free rate for CAPM/SFM calculations.
        """
        self._ddof:int = ddof
        self._threshold: float = threshold
        self._count: int = 0
        self._mean_a: float = 0.0
        self._mean_b: float = 0.0
        self._m2_ab: float = 0.0  # Sum of (a - mean_a) * (b - mean_b)
        self._m2_bb: float = 0.0  # Sum of (b - mean_b)^2

    def reset(self) -> None:
        self._count = 0
        self._mean_a = 0.0
        self._mean_b = 0.0
        self._m2_ab = 0.0
        self._m2_bb = 0.0

    def revert(self, ret_a: float, ret_b: float) -> None:
        """Remove the contribution of an old pair (ret_a, ret_b) from the statistics."""
        if self._count == 0:
            return
        elif self._count == 1:
            self.reset()
            return

        a = ret_a - self._threshold
        b = ret_b - self._threshold

        # Inverse of Welford's update
        # Calculate old means before this point was added
        # old_mean = (new_mean * n - value) / (n - 1)
        n = self._count
        n1 = n - 1
        old_mean_a = (self._mean_a * n - a) / n1
        old_mean_b = (self._mean_b * n - b) / n1
 
        # Reverse the m2_ab update
        # The forward step was: m2 += (a - old_mean_a) * (b - new_mean_b)
        # We subtract that contribution
        delta_a = a - old_mean_a
        # Note: uses current mean which is 'new_mean' in forward context
        delta_b = b - self._mean_b
        self._m2_ab -= delta_a * delta_b
        self._m2_bb -= (b - old_mean_b) * delta_b

        # Update means
        self._mean_a = old_mean_a
        self._mean_b = old_mean_b
        
        self._count -= 1

    def update(self, ret_a: float, ret_b: float) -> None:
        a = ret_a - self._threshold
        b = ret_b - self._threshold

        # Welford's update step for covariance
        self._count += 1
        delta_a = a - self._mean_a
        delta_b = b - self._mean_b

        n = self._count        
        self._mean_a += delta_a / n
        self._mean_b += delta_b / n

        # Update co-moment: m2_ab += (n-1)/n * delta_a * delta_b
        # Equivalent to: m2_ab += delta_a * (b - new_mean_b)
        delta_b_new = b - self._mean_b
        self._m2_ab += delta_a * delta_b_new
        self._m2_bb += delta_b * delta_b_new

    @property
    def value(self) -> float:
        n = self._count
        if n < 2:
            return 0.0
        return self._m2_ab / (n - self._ddof)

    @property
    def beta(self) -> float:
        # Don't need the covariance normalization.
        v = self._m2_bb

        # If there is no benchmark variance, then the OLS slope
        # and therefore the OLS intercept are not identifiable.
        return self._m2_ab / v if v != 0 else math.nan
