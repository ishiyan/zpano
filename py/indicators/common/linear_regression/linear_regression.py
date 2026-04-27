"""Linear regression indicator."""

import math

from ...core.indicator import Indicator
from ...core.metadata import Metadata
from ...core.build_metadata import build_metadata, OutputText
from ...core.identifier import Identifier
from ...core.component_triple_mnemonic import component_triple_mnemonic
from ...core.output import Output
from ....entities.bar import Bar
from ....entities.quote import Quote
from ....entities.trade import Trade
from ....entities.scalar import Scalar
from ....entities.bar_component import BarComponent, DEFAULT_BAR_COMPONENT, bar_component_value
from ....entities.quote_component import QuoteComponent, DEFAULT_QUOTE_COMPONENT, quote_component_value
from ....entities.trade_component import TradeComponent, DEFAULT_TRADE_COMPONENT, trade_component_value
from .params import LinearRegressionParams


class LinearRegression(Indicator):
    """Computes the least-squares regression line over a rolling window.

    Produces five outputs per sample:
    - Value:     b + m*(period-1)
    - Forecast:  b + m*period
    - Intercept: b
    - SlopeRad:  m
    - SlopeDeg:  atan(m)*180/pi

    The indicator is not primed during the first (period-1) updates.
    """

    def __init__(self, params: LinearRegressionParams) -> None:
        length = params.length
        if length < 2:
            raise ValueError(
                "invalid linear regression parameters: length should be greater than 1")

        bc = params.bar_component if params.bar_component is not None else DEFAULT_BAR_COMPONENT
        qc = params.quote_component if params.quote_component is not None else DEFAULT_QUOTE_COMPONENT
        tc = params.trade_component if params.trade_component is not None else DEFAULT_TRADE_COMPONENT

        self._bar_func = bar_component_value(bc)
        self._quote_func = quote_component_value(qc)
        self._trade_func = trade_component_value(tc)

        self._mnemonic = f"linreg({length}{component_triple_mnemonic(bc, qc, tc)})"
        self._description = f"Linear Regression {self._mnemonic}"

        n = float(length)
        self._length = length
        self._length_f = n
        self._sum_x = n * (n - 1) * 0.5
        sum_x_sqr = n * (n - 1) * (2 * n - 1) / 6
        self._divisor = self._sum_x * self._sum_x - n * sum_x_sqr

        self._window: list[float] = [0.0] * length
        self._window_count: int = 0
        self._primed: bool = False

        self._cur_value: float = math.nan
        self._cur_forecast: float = math.nan
        self._cur_intercept: float = math.nan
        self._cur_slope_rad: float = math.nan
        self._cur_slope_deg: float = math.nan

    def is_primed(self) -> bool:
        return self._primed

    def metadata(self) -> Metadata:
        return build_metadata(
            Identifier.LINEAR_REGRESSION,
            self._mnemonic,
            self._description,
            [
                OutputText(self._mnemonic, self._description + " value"),
                OutputText(self._mnemonic, self._description + " forecast"),
                OutputText(self._mnemonic, self._description + " intercept"),
                OutputText(self._mnemonic, self._description + " slope"),
                OutputText(self._mnemonic, self._description + " angle"),
            ],
        )

    def _compute_from_window(self) -> None:
        rad_to_deg = 180.0 / math.pi
        sum_xy = 0.0
        sum_y = 0.0

        for i in range(self._length, 0, -1):
            v = self._window[self._length - i]
            sum_y += v
            sum_xy += float(i - 1) * v

        m = (self._length_f * sum_xy - self._sum_x * sum_y) / self._divisor
        b = (sum_y - m * self._sum_x) / self._length_f

        self._cur_slope_rad = m
        self._cur_slope_deg = math.atan(m) * rad_to_deg
        self._cur_intercept = b
        self._cur_value = b + m * (self._length_f - 1)
        self._cur_forecast = b + m * self._length_f

    def update(self, sample: float) -> float:
        if math.isnan(sample):
            return sample

        if self._primed:
            for i in range(self._length - 1):
                self._window[i] = self._window[i + 1]
            self._window[self._length - 1] = sample
            self._compute_from_window()
            return self._cur_value

        self._window[self._window_count] = sample
        self._window_count += 1

        if self._window_count == self._length:
            self._primed = True
            self._compute_from_window()
            return self._cur_value

        return math.nan

    def _update_entity(self, t, sample: float) -> Output:
        value = self.update(sample)

        if math.isnan(value):
            nan = math.nan
            return [
                Scalar(t, nan),
                Scalar(t, nan),
                Scalar(t, nan),
                Scalar(t, nan),
                Scalar(t, nan),
            ]

        return [
            Scalar(t, self._cur_value),
            Scalar(t, self._cur_forecast),
            Scalar(t, self._cur_intercept),
            Scalar(t, self._cur_slope_rad),
            Scalar(t, self._cur_slope_deg),
        ]

    def update_scalar(self, sample: Scalar) -> Output:
        return self._update_entity(sample.time, sample.value)

    def update_bar(self, sample: Bar) -> Output:
        return self._update_entity(sample.time, self._bar_func(sample))

    def update_quote(self, sample: Quote) -> Output:
        return self._update_entity(sample.time, self._quote_func(sample))

    def update_trade(self, sample: Trade) -> Output:
        return self._update_entity(sample.time, self._trade_func(sample))
