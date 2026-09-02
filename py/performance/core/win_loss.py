from ...streaming_kbn import KleinKBNSummator

class WinLoss:
    """
    Streaming winning/loosing return averages and counts.
    """
    def __init__(self) -> None:
        self._non_zero_sum: KleinKBNSummator = KleinKBNSummator()
        self._win_sum: KleinKBNSummator = KleinKBNSummator()
        self._loss_sum: KleinKBNSummator = KleinKBNSummator()

    def reset(self) -> None:
        self._non_zero_sum.reset()
        self._win_sum.reset()
        self._loss_sum.reset()

    def revert(self, ret: float) -> None:
        if ret != 0:
            self._non_zero_sum.revert(ret)
        if ret > 0:
            self._win_sum.revert(ret)
        if ret < 0:
            self._loss_sum.revert(ret)

    def update(self, ret: float) -> None:
        if ret != 0:
            self._non_zero_sum.update(ret)
        if ret > 0:
            self._win_sum.update(ret)
        if ret < 0:
            self._loss_sum.update(ret)

    @property
    def non_zero_returns_mean(self) -> float:
        """
        Arithmetic mean (average) of non-zero returns
        """
        return self._non_zero_sum.mean

    @property
    def non_zero_returns_count(self) -> float:
        """
        The number of non-zero returns
        """
        return self._non_zero_sum.n

    @property
    def winning_returns_sum(self) -> float:
        """
        Sum of winning (positive) returns
        """
        return self._win_sum.value

    @property
    def winning_returns_mean(self) -> float:
        """
        Arithmetic mean (average) of winning (positive) returns
        """
        return self._win_sum.mean

    @property
    def winning_returns_count(self) -> float:
        """
        The number of winning (positive) returns
        """
        return self._win_sum.n

    @property
    def losing_returns_sum(self) -> float:
        """
        Sum of losing (negative) returns
        """
        return self._loss_sum.value

    @property
    def losing_returns_mean(self) -> float:
        """
        Arithmetic mean (average) of losing (negative) returns
        """
        return self._loss_sum.mean

    @property
    def losing_returns_count(self) -> float:
        """
        The number of losing (negative) returns
        """
        return self._loss_sum.n
