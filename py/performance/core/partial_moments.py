import math

from ...streaming_kbn import RawMomentsKleinKBN

class PartialMoments:
    """
    Streaming partial moments.
    """
    def __init__(self, threshold: float) -> None:
        """
        Streaming low/high partial moments.

        Args:
            threshold: Target return or minimum acceptable return (MAR) in same periodicity as returns
        """
        self.threshold = threshold
        self._count_total: int = 0

        # This matches scipy's default behavior for kurtosis
        self._upper_excess_kbn: RawMomentsKleinKBN = RawMomentsKleinKBN(ddof=1, bias=True, fisher=True)
        self._lower_excess_kbn: RawMomentsKleinKBN = RawMomentsKleinKBN(ddof=1, bias=True, fisher=True)
        self._lpm_kbn: RawMomentsKleinKBN = RawMomentsKleinKBN(ddof=1, bias=True, fisher=True)
        self._hpm_kbn: RawMomentsKleinKBN = RawMomentsKleinKBN(ddof=1, bias=True, fisher=True)

    def reset(self) -> None:
        self._count_total = 0
        self._upper_excess_kbn.reset()
        self._lower_excess_kbn.reset()
        self._lpm_kbn.reset()
        self._hpm_kbn.reset()

    def revert(self, ret: float) -> None:
        self._count_total -= 1
        # Lower partial moments for the raw returns less target return
        pm = self.threshold - ret
        if pm < 0:
            self._upper_excess_kbn.revert(-pm)
            pm = 0
        self._lpm_kbn.revert(pm)

        # Higher partial moments for the raw returns less required return
        pm = ret - self.threshold
        if pm < 0:
            self._lower_excess_kbn.revert(-pm)
            pm = 0
        self._hpm_kbn.revert(pm)

    def update(self, ret: float) -> None:
        self._count_total += 1
        # Lower partial moments for the raw returns less target return
        pm = self.threshold - ret
        if pm < 0:
            self._upper_excess_kbn.update(-pm)
            pm = 0
        self._lpm_kbn.update(pm)

        # Higher partial moments for the raw returns less required return
        pm = ret - self.threshold
        if pm < 0:
            self._lower_excess_kbn.update(-pm)
            pm = 0
        self._hpm_kbn.update(pm)

    @property
    def lower_partial_moment_1(self) -> float:
        return self._lpm_kbn.x1

    @property
    def lower_partial_moment_2(self) -> float:
        return self._lpm_kbn.x2

    @property
    def lower_partial_moment_3(self) -> float:
        return self._lpm_kbn.x3

    @property
    def lower_partial_moment_4(self) -> float:
        return self._lpm_kbn.x4

    @property
    def higher_partial_moment_1(self) -> float:
        return self._hpm_kbn.x1

    @property
    def higher_partial_moment_2(self) -> float:
        return self._hpm_kbn.x2

    @property
    def higher_partial_moment_3(self) -> float:
        return self._hpm_kbn.x3

    @property
    def higher_partial_moment_4(self) -> float:
        return self._hpm_kbn.x4

    @property
    def downside_frequency(self) -> float:
        """
        Proportion of returns below threshold
        """
        total = self._count_total
        if total == 0:
            return math.nan
        return float(self._lower_excess_kbn.n) / total

    @property
    def upside_frequency(self) -> float:
        """
        Proportion of returns above threshold
        """
        total = self._count_total
        if total == 0:
            return math.nan
        return float(self._upper_excess_kbn.n) / total

    @property
    def downside_potential(self) -> float:
        """
        Mean of lower partial moments (also called shortfall)
        """
        return self._lpm_kbn.mean
        
    @property
    def total_count(self) -> float:
        return self._count_total

    @property
    def upper_excess_count(self) -> float:
        return self._upper_excess_kbn.n
    
    @property
    def lower_excess_count(self) -> float:
        return self._lower_excess_kbn.n
    
    @property
    def upper_excess_moment_1(self) -> float:
        return self._upper_excess_kbn.x1
    
    @property
    def upper_excess_moment_1_sum(self) -> float:
        return self._upper_excess_kbn.x1_sum
    
    @property
    def upper_excess_moment_2(self) -> float:
        return self._upper_excess_kbn.x2
    
    @property
    def upper_excess_moment_2_sum(self) -> float:
        return self._upper_excess_kbn.x2_sum
    
    @property
    def upper_excess_moment_3(self) -> float:
        return self._upper_excess_kbn.x3
    
    @property
    def upper_excess_moment_4(self) -> float:
        return self._upper_excess_kbn.x4

    @property
    def lower_excess_moment_1(self) -> float:
        return self._lower_excess_kbn.x1
    
    @property
    def lower_excess_moment_2(self) -> float:
        return self._lower_excess_kbn.x2
    
    @property
    def lower_excess_moment_2_sum(self) -> float:
        return self._lower_excess_kbn.x2_sum
    
    @property
    def lower_excess_moment_3(self) -> float:
        return self._lower_excess_kbn.x3
    
    @property
    def lower_excess_moment_4(self) -> float:
        return self._lower_excess_kbn.x4
