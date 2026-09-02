import math

from ...streaming_kbn import KleinKBNAccumulator
from .min_max import MinMax

class Capture:
    def __init__(self) -> None:
        self._logret_a_sum_up: KleinKBNAccumulator = KleinKBNAccumulator()
        self._logret_b_sum_up: KleinKBNAccumulator = KleinKBNAccumulator()
        self._logret_a_sum_dn: KleinKBNAccumulator = KleinKBNAccumulator()
        self._logret_b_sum_dn: KleinKBNAccumulator = KleinKBNAccumulator()
        self._a_sum_up: KleinKBNAccumulator = KleinKBNAccumulator()
        self._b_sum_up: KleinKBNAccumulator = KleinKBNAccumulator()
        self._a_sum_dn: KleinKBNAccumulator = KleinKBNAccumulator()
        self._b_sum_dn: KleinKBNAccumulator = KleinKBNAccumulator()
        self._a_num_up: int = 0
        self._b_num_up: int = 0
        self._a_num_dn: int = 0
        self._b_num_dn: int = 0
        self._a_perc_up: int = 0
        self._b_perc_up: int = 0
        self._a_perc_dn: int = 0
        self._b_perc_dn: int = 0

    def reset(self) -> None:
        self._logret_a_sum_up.reset()
        self._logret_b_sum_up.reset()
        self._logret_a_sum_dn.reset()
        self._logret_b_sum_dn.reset()
        self._a_sum_up.reset()
        self._b_sum_up.reset()
        self._a_sum_dn.reset()
        self._b_sum_dn.reset()
        self._a_num_up = 0
        self._b_num_up = 0
        self._a_num_dn = 0
        self._b_num_dn = 0
        self._a_perc_up = 0
        self._b_perc_up = 0
        self._a_perc_dn = 0
        self._b_perc_dn = 0

    @staticmethod
    def _logret(ret: float) -> float:
        return math.log1p(ret) if ret != 0 else 0

    def revert(self, ret_asset: float, ret_benchmark: float) -> None:
        if ret_benchmark > 0:  # Upside
            # Geometric
            self._logret_a_sum_up.revert(Capture._logret(ret_asset))
            self._logret_b_sum_up.revert(Capture._logret(ret_benchmark))
            # Arithmetic
            self._a_sum_up.revert(ret_asset)
            self._b_sum_up.revert(ret_benchmark)
            # Number
            self._b_num_up -= 1
            if ret_asset > 0:
                self._a_num_up -= 1
            # Percentage
            self._b_perc_up -= 1
            if ret_asset > ret_benchmark:
                self._a_perc_up -= 1
        else:  # Downside
            # Geometric
            self._logret_a_sum_dn.revert(Capture._logret(ret_asset))
            self._logret_b_sum_dn.revert(Capture._logret(ret_benchmark))
            # Arithmetic
            self._a_sum_dn.revert(ret_asset)
            self._b_sum_dn.revert(ret_benchmark)
            # Number
            self._b_num_dn -= 1
            if ret_asset < 0:
                self._a_num_dn -= 1
            # Percentage
            if ret_benchmark < 0:
                self._b_perc_dn -= 1
                if ret_asset > ret_benchmark:
                    self._a_perc_dn -= 1

    def update(self, ret_asset: float, ret_benchmark: float) -> None:        
        if ret_benchmark > 0: # Upside
            # Geometric
            self._logret_a_sum_up.update(Capture._logret(ret_asset))
            self._logret_b_sum_up.update(Capture._logret(ret_benchmark))
            # Arithmetic
            self._a_sum_up.update(ret_asset)
            self._b_sum_up.update(ret_benchmark)
            # Counts, Perc
            self._b_num_up += 1
            if ret_asset > 0:
                self._a_num_up += 1
            # Perc
            self._b_perc_up += 1
            if ret_asset > ret_benchmark:
                self._a_perc_up += 1
        else: # Downside
            # Geometric
            self._logret_a_sum_dn.update(Capture._logret(ret_asset))
            self._logret_b_sum_dn.update(Capture._logret(ret_benchmark))
            # Arithmetic
            self._a_sum_dn.update(ret_asset)
            self._b_sum_dn.update(ret_benchmark)
            # Counts
            self._b_num_dn += 1
            if ret_asset < 0:
                self._a_num_dn += 1
            # Perc
            if ret_benchmark < 0:
                self._b_perc_dn += 1
                if ret_asset > ret_benchmark:
                    self._a_perc_dn += 1

    @property
    def upside_capture_ratio_geometric(self) -> float:
        a_cum = math.expm1(self._logret_a_sum_up.value)
        b_cum = math.expm1(self._logret_b_sum_up.value)
        return a_cum / b_cum if b_cum != 0 else math.nan

    @property
    def upside_capture_ratio_arithmetic(self) -> float:
        b_sum = self._b_sum_up.value
        return self._a_sum_up.value / b_sum if b_sum != 0 else math.nan

    @property
    def downside_capture_ratio_geometric(self) -> float:
        a_cum = math.expm1(self._logret_a_sum_dn.value)
        b_cum = math.expm1(self._logret_b_sum_dn.value)
        return a_cum / b_cum if b_cum != 0 else math.nan

    @property
    def downside_capture_ratio_arithmetic(self) -> float:
        b_sum = self._b_sum_dn.value
        return self._a_sum_dn.value / b_sum if b_sum != 0 else math.nan

    @property
    def up_number_ratio(self) -> float:
        b_num = self._b_num_up
        return self._a_num_up / b_num if b_num != 0 else math.nan

    @property
    def down_number_ratio(self) -> float:
        b_num = self._b_num_dn
        return self._a_num_dn / b_num if b_num != 0 else math.nan

    @property
    def up_percentage_ratio(self) -> float:
        b_perc = self._b_perc_up
        return self._a_perc_up / b_perc if b_perc != 0 else math.nan

    @property
    def down_percentage_ratio(self) -> float:
        b_perc = self._b_perc_dn
        return self._a_perc_dn / b_perc if b_perc != 0 else math.nan
