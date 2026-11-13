"""Ehlers' two-pole SuperSmoother indicator."""

from __future__ import annotations

import math
from typing import Optional

from ...core.build_metadata import OutputText, build_metadata
from ...core.identifier import Identifier
from ...core.indicator import Indicator
from ...core.line_indicator import LineIndicator
from ...core.metadata import Metadata
from ...core.output import Output
from ....entities.bar import Bar
from ....entities.bar_component import DEFAULT_BAR_COMPONENT, bar_component_value
from ....entities.quote import Quote
from ....entities.quote_component import DEFAULT_QUOTE_COMPONENT, quote_component_value
from ....entities.scalar import Scalar
from ....entities.trade import Trade
from ....entities.trade_component import DEFAULT_TRADE_COMPONENT, trade_component_value
from .params import SuperSmootherParams, default_params, super_smoother_mnemonic


def _smoother_coefficients(period: float) -> tuple[float, float, float]:
    """Compute SuperSmoother (equation 3-3) coefficients at full binary64 precision."""
    smoother_angle = math.sqrt(2.0) * math.pi / period
    a1 = math.exp(-smoother_angle)
    c2 = 2.0 * a1 * math.cos(smoother_angle)
    c3 = -a1 * a1
    return 1.0 - c2 - c3, c2, c3


class SuperSmoother(Indicator):
    """Ehlers' two-pole SuperSmoother low-pass filter with a zero at the two-bar Nyquist period."""

    def __init__(self, params: Optional[SuperSmootherParams] = None) -> None:
        source = default_params() if params is None else params

        if isinstance(source.shortest_cycle_period, bool) or not isinstance(
            source.shortest_cycle_period, (int, float)
        ):
            raise ValueError("shortest_cycle_period must be finite and at least 2 bars")
        shortest = float(source.shortest_cycle_period)
        if not math.isfinite(shortest) or shortest < 2.0:
            raise ValueError("shortest_cycle_period must be finite and at least 2 bars")

        bar_comp = source.bar_component if source.bar_component is not None else DEFAULT_BAR_COMPONENT
        quote_comp = (
            source.quote_component if source.quote_component is not None else DEFAULT_QUOTE_COMPONENT
        )
        trade_comp = (
            source.trade_component if source.trade_component is not None else DEFAULT_TRADE_COMPONENT
        )

        bar_func = bar_component_value(bar_comp)
        quote_func = quote_component_value(quote_comp)
        trade_func = trade_component_value(trade_comp)

        mnemonic = super_smoother_mnemonic(source)
        description = f"Super Smoother {mnemonic}"
        self._line = LineIndicator(
            mnemonic, description, bar_func, quote_func, trade_func, self.update
        )

        self._warmup_bars = 2
        self._c1, self._c2, self._c3 = _smoother_coefficients(shortest)
        self._count = 0
        self._input1 = 0.0
        self._output1 = 0.0
        self._output2 = 0.0
        self._current_value = math.nan
        self.reset()

    @staticmethod
    def create(params: Optional[SuperSmootherParams] = None) -> "SuperSmoother":
        """Create a new SuperSmoother from parameters."""
        return SuperSmoother(params)

    @staticmethod
    def create_default() -> "SuperSmoother":
        """Create a new SuperSmoother with default parameters."""
        return SuperSmoother()

    def reset(self) -> None:
        """Reset all streaming histories and priming state."""
        self._count = 0
        self._input1 = 0.0
        self._output1 = 0.0
        self._output2 = 0.0
        self._current_value = math.nan

    def warmup_bars(self) -> int:
        """Return the number of initial NaN outputs before priming."""
        return self._warmup_bars

    def value(self) -> float:
        """Return the latest calculated value (NaN before priming)."""
        return self._current_value

    def is_primed(self) -> bool:
        """Return whether the indicator has produced at least one primed sample."""
        return self._count > self._warmup_bars

    def metadata(self) -> Metadata:
        """Return metadata describing this indicator and its scalar output."""
        return build_metadata(
            Identifier.SUPER_SMOOTHER,
            self._line.mnemonic,
            self._line.description,
            [OutputText(self._line.mnemonic, self._line.description)],
        )

    def update(self, sample: float) -> float:
        """Consume one scalar observation and return the current SuperSmoother value."""
        if isinstance(sample, bool) or not isinstance(sample, (int, float)):
            raise ValueError("sample must be finite or NaN")
        numeric = float(sample)
        if math.isnan(numeric):
            return math.nan
        if not math.isfinite(numeric):
            raise ValueError("sample must be finite or NaN")

        if self._count == 0:
            self._input1 = numeric
            self._output1 = numeric
            self._output2 = numeric

        output = (
            self._c1 * (numeric + self._input1) / 2.0
            + self._c2 * self._output1
            + self._c3 * self._output2
        )

        self._input1 = numeric
        self._output2 = self._output1
        self._output1 = output
        self._count += 1

        if self._count > self._warmup_bars:
            self._current_value = output
            return output
        return math.nan

    def update_scalar(self, sample: Scalar) -> Output:
        return self._line.update_scalar(sample)

    def update_bar(self, sample: Bar) -> Output:
        return self._line.update_bar(sample)

    def update_quote(self, sample: Quote) -> Output:
        return self._line.update_quote(sample)

    def update_trade(self, sample: Trade) -> Output:
        return self._line.update_trade(sample)
