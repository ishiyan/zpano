import math

from ...streaming_kbn import KleinKBNAccumulator

class RawPartialMoments:
    """
    Streaming raw partial moments.
    """
    def __init__(self) -> None:
        """
        Streaming raw low/high partial moments.
        """
        self._count: int = 0
        self._count_pos: int = 0
        self._count_neg: int = 0
        self._lpm_kbn: KleinKBNAccumulator = KleinKBNAccumulator()
        self._hpm_kbn: KleinKBNAccumulator = KleinKBNAccumulator()
        self._pos_kbn: KleinKBNAccumulator = KleinKBNAccumulator()
        self._neg_kbn: KleinKBNAccumulator = KleinKBNAccumulator()

    def reset(self) -> None:
        self._count = 0
        self._count_pos = 0
        self._count_neg = 0
        self._lpm_kbn.reset()
        self._hpm_kbn.reset()
        self._pos_kbn.reset()
        self._neg_kbn.reset()

    def revert(self, ret: float) -> None:
        self._count -= 1
        # Lower partial moment
        pm = - ret
        if pm < 0:
            pm = 0
        self._lpm_kbn.revert(pm)

        # Higher partial moment
        pm = ret
        if pm < 0:
            pm = 0
        self._hpm_kbn.revert(pm)

        if ret > 0:
            self._count_pos -= 1
            self._pos_kbn.revert(ret)
        elif ret < 0:
            self._count_neg -= 1
            self._neg_kbn.revert(ret)

    def update(self, ret: float) -> None:
        self._count += 1
        # Lower partial moment
        pm = - ret
        if pm < 0:
            pm = 0
        self._lpm_kbn.update(pm)

        # Higher partial moment
        pm = ret
        if pm < 0:
            pm = 0
        self._hpm_kbn.update(pm)

        if ret > 0:
            self._count_pos += 1
            self._pos_kbn.update(ret)
        elif ret < 0:
            self._count_neg += 1
            self._neg_kbn.update(ret)

    @property
    def count(self) -> int:
        return self._count

    @property
    def lower_partial_moment_1(self) -> float:
        return self._lpm_kbn.value

    @property
    def higher_partial_moment_1(self) -> float:
        return self._hpm_kbn.value

    @property
    def count_negative(self) -> int:
        return self._count_neg

    @property
    def sum_negative(self) -> float:
        return self._neg_kbn.value

    @property
    def count_positive(self) -> int:
        return self._count_pos

    @property
    def sum_positive(self) -> float:
        return self._pos_kbn.value
