import math

from ...streaming_kbn import RawMomentsKleinKBN

class CovarianceBullBear:
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

        # Full
        self._count: int = 0
        self._mean_a: float = 0.0
        self._mean_b: float = 0.0
        self._m2_ab: float = 0.0  # Sum of (a - mean_a) * (b - mean_b)
        self._m2_bb: float = 0.0  # Sum of (b - mean_b)^2

        # Bull subset
        self._count_bull: int = 0
        self._mean_a_bull: float = 0.0
        self._mean_b_bull: float = 0.0
        self._m2_ab_bull: float = 0.0
        self._m2_bb_bull: float = 0.0

        # Bear subset
        self._count_bear: int = 0
        self._mean_a_bear: float = 0.0
        self._mean_b_bear: float = 0.0
        self._m2_ab_bear: float = 0.0
        self._m2_bb_bear: float = 0.0

    def reset(self) -> None:
        self._count = 0
        self._mean_a = 0.0
        self._mean_b = 0.0
        self._m2_ab = 0.0
        self._m2_bb = 0.0
        self._count_bull = 0
        self._mean_a_bull = 0.0
        self._mean_b_bull = 0.0
        self._m2_ab_bull = 0.0
        self._m2_bb_bull = 0.0
        self._count_bear = 0
        self._mean_a_bear = 0.0
        self._mean_b_bear = 0.0
        self._m2_ab_bear = 0.0
        self._m2_bb_bear = 0.0

    def revert(self, ret_a: float, ret_b: float) -> None:
        """Remove the contribution of an old pair (ret_a, ret_b) from the statistics."""
        if self._count == 0:
            return

        if self._count == 1:
            self._count = 0
            self._mean_a = 0.0
            self._mean_b = 0.0
            self._m2_ab = 0.0
            self._m2_bb = 0.0
        else:
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

        n = self._count_bull
        if n > 0  and b > 0:
            # Bull subset
            if self._count_bull == 1:
                self._count_bull = 0
                self._mean_a_bull = 0.0
                self._mean_b_bull = 0.0
                self._m2_ab_bull = 0.0
                self._m2_bb_bull = 0.0
            else:
                n1 = n - 1
                old_mean_a = (self._mean_a_bull * n - a) / n1
                old_mean_b = (self._mean_b_bull * n - b) / n1
                delta_a = a - old_mean_a
                delta_b = b - self._mean_b_bull
                self._m2_ab_bull -= delta_a * delta_b
                self._m2_bb_bull -= (b - old_mean_b) * delta_b
                self._mean_a_bull = old_mean_a
                self._mean_b_bull = old_mean_b        
                self._count_bull -= 1

        n = self._count_bear
        if n > 0  and b < 0:
            # Bear subset
            if self._count_bear == 1:
                self._count_bear = 0
                self._mean_a_bear = 0.0
                self._mean_b_bear = 0.0
                self._m2_ab_bear = 0.0
                self._m2_bb_bear = 0.0
            else:
                n1 = n - 1
                old_mean_a = (self._mean_a_bear * n - a) / n1
                old_mean_b = (self._mean_b_bear * n - b) / n1
                delta_a = a - old_mean_a
                delta_b = b - self._mean_b_bear
                self._m2_ab_bear -= delta_a * delta_b
                self._m2_bb_bear -= (b - old_mean_b) * delta_b
                self._mean_a_bear = old_mean_a
                self._mean_b_bear = old_mean_b        
                self._count_bear -= 1

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

        if b > 0:
            # Bull subset
            self._count_bull += 1
            delta_a = a - self._mean_a_bull
            delta_b = b - self._mean_b_bull
            n = self._count_bull
            self._mean_a_bull += delta_a / n
            self._mean_b_bull += delta_b / n
            delta_b_new = b - self._mean_b_bull
            self._m2_ab_bull += delta_a * delta_b_new
            self._m2_bb_bull += delta_b * delta_b_new
        elif b < 0:
            # Bear subset
            self._count_bear += 1
            delta_a = a - self._mean_a_bear
            delta_b = b - self._mean_b_bear
            n = self._count_bear
            self._mean_a_bear += delta_a / n
            self._mean_b_bear += delta_b / n
            delta_b_new = b - self._mean_b_bear
            self._m2_ab_bear += delta_a * delta_b_new
            self._m2_bb_bear += delta_b * delta_b_new

    @property
    def value(self) -> float:
        n = self._count
        if n < 2:
            return 0.0
        return self._m2_ab / (n - self._ddof)

    @property
    def value_bull(self) -> float:
        n = self._count_bull
        if n < 2:
            return 0.0
        return self._m2_ab_bull / (n - self._ddof)

    @property
    def value_bear(self) -> float:
        n = self._count_bear
        if n < 2:
            return 0.0
        return self._m2_ab_bear / (n - self._ddof)

    @property
    def beta(self) -> float:
        # Don't need the covariance normalization.
        v = self._m2_bb

        # If there is no benchmark variance, then the OLS slope
        # and therefore the OLS intercept are not identifiable.
        return self._m2_ab / v if v != 0 else math.nan

    @property
    def beta_bull(self) -> float:
        # Don't need the covariance normalization.
        v = self._m2_bb_bull
        return self._m2_ab_bull / v if v != 0 else math.nan

    @property
    def beta_bear(self) -> float:
        # Don't need the covariance normalization.
        v = self._m2_bb_bear
        return self._m2_ab_bear / v if v != 0 else math.nan

    @property
    def alpha(self) -> float:
        beta = self.beta
        if math.isnan(beta):
            return math.nan
        return self._mean_a - beta * self._mean_b

    @property
    def alpha_bull(self) -> float:
        beta = self.beta_bull
        if math.isnan(beta):
            return math.nan
        return self._mean_a_bull - beta * self._mean_b_bull

    @property
    def alpha_bear(self) -> float:
        beta = self.beta_bear
        if math.isnan(beta):
            return math.nan
        return self._mean_a_bear - beta * self._mean_b_bear
