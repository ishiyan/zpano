import math

from ...streaming_kbn import LinearRegressionKleinKBN

class SFMRegression:
    def __init__(self, risk_free_rate: float) -> None:
        self._risk_free_rate = risk_free_rate
        self._full = LinearRegressionKleinKBN()
        self._bull = LinearRegressionKleinKBN()
        self._bear = LinearRegressionKleinKBN()

    def reset(self) -> None:
        self._full.reset()
        self._bull.reset()
        self._bear.reset()

    def revert(self, ret: float, benchmark: float) -> None:
        x = benchmark - self._risk_free_rate
        y = ret - self._risk_free_rate

        self._full.revert(x, y)

        if x > 0:
            self._bull.revert(x, y)
        elif x < 0:
            self._bear.revert(x, y)

    def update(self, ret: float, benchmark: float) -> None:
        x = benchmark - self._risk_free_rate
        y = ret - self._risk_free_rate

        self._full.update(x, y)

        if x > 0:
            self._bull.update(x, y)
        elif x < 0:
            self._bear.update(x, y)

    @property
    def alpha(self) -> float:
        return self._full.intercept

    @property
    def beta(self) -> float:
        return self._full.slope

    @property
    def beta_bull(self) -> float:
        return self._bull.slope

    @property
    def beta_bear(self) -> float:
        return self._bear.slope

    @property
    def r2(self) -> float:
        corr = self._full.correlation
        return corr * corr if not math.isnan(corr) else math.nan
