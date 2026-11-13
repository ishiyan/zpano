"""Ehlers' Roofing Filter indicator."""

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
from .params import RoofingFilterParams, default_params, roofing_filter_mnemonic


def _smoother_coefficients(period: float) -> tuple[float, float, float]:
    """Compute SuperSmoother (equation 3-3) coefficients at full binary64 precision."""
    smoother_angle = math.sqrt(2.0) * math.pi / period
    a1 = math.exp(-smoother_angle)
    c2 = 2.0 * a1 * math.cos(smoother_angle)
    c3 = -a1 * a1
    return 1.0 - c2 - c3, c2, c3


class RoofingFilter(Indicator):
    """Ehlers' Roofing Filter: high-pass detrending followed by two-pole SuperSmoother."""

    def __init__(self, params: Optional[RoofingFilterParams] = None) -> None:
        source = default_params() if params is None else params

        if isinstance(source.shortest_cycle_period, bool) or not isinstance(
            source.shortest_cycle_period, (int, float)
        ):
            raise ValueError("shortest_cycle_period must be finite and at least 2 bars")
        shortest = float(source.shortest_cycle_period)
        if not math.isfinite(shortest) or shortest < 2.0:
            raise ValueError("shortest_cycle_period must be finite and at least 2 bars")

        if isinstance(source.longest_cycle_period, bool) or not isinstance(
            source.longest_cycle_period, (int, float)
        ):
            raise ValueError("longest_cycle_period must be finite and greater than shortest")
        longest = float(source.longest_cycle_period)
        if not math.isfinite(longest) or longest <= shortest:
            raise ValueError("longest_cycle_period must be finite and greater than shortest")

        if (
            isinstance(source.highpass_poles, bool)
            or not isinstance(source.highpass_poles, int)
            or source.highpass_poles not in (1, 2)
        ):
            raise ValueError("highpass_poles must be integer 1 or 2")

        if not isinstance(source.zero_mean, bool) or (
            source.zero_mean and source.highpass_poles != 1
        ):
            raise ValueError("zero_mean must be a bool and requires highpass_poles=1")

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

        mnemonic = roofing_filter_mnemonic(source)
        description = f"Roofing Filter {mnemonic}"
        self._line = LineIndicator(
            mnemonic, description, bar_func, quote_func, trade_func, self.update
        )

        self._poles = source.highpass_poles
        self._zero_mean = source.zero_mean
        self._warmup_bars = 4 if self._poles == 2 or self._zero_mean else 3

        angle = (
            2.0 * math.pi / longest
            if self._poles == 1
            else math.sqrt(2.0) * math.pi / longest
        )
        tangent = math.tan(angle / 2.0)
        alpha = 2.0 * tangent / (1.0 + tangent)
        self._gain = 1.0 - alpha / 2.0
        self._decay = 1.0 - alpha
        self._c1, self._c2, self._c3 = _smoother_coefficients(shortest)

        self._count = 0
        self._input1 = 0.0
        self._input2 = 0.0
        self._hp1 = 0.0
        self._hp2 = 0.0
        self._filt1 = 0.0
        self._filt2 = 0.0
        self._zero1 = 0.0
        self._current_value = math.nan
        self.reset()

    @staticmethod
    def create(params: Optional[RoofingFilterParams] = None) -> "RoofingFilter":
        """Create a new RoofingFilter from parameters."""
        return RoofingFilter(params)

    @staticmethod
    def create_default() -> "RoofingFilter":
        """Create a new RoofingFilter with default parameters."""
        return RoofingFilter()

    def reset(self) -> None:
        """Reset all streaming histories and priming state."""
        self._count = 0
        self._input1 = 0.0
        self._input2 = 0.0
        self._hp1 = 0.0
        self._hp2 = 0.0
        self._filt1 = 0.0
        self._filt2 = 0.0
        self._zero1 = 0.0
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
            Identifier.ROOFING_FILTER,
            self._line.mnemonic,
            self._line.description,
            [OutputText(self._line.mnemonic, self._line.description)],
        )

    def update(self, sample: float) -> float:
        """Consume one scalar observation and return the current RoofingFilter value."""
        if isinstance(sample, bool) or not isinstance(sample, (int, float)):
            raise ValueError("sample must be finite or NaN")
        numeric = float(sample)
        if math.isnan(numeric):
            return math.nan
        if not math.isfinite(numeric):
            raise ValueError("sample must be finite or NaN")

        if self._count < self._poles:
            hp = 0.0
            filt = 0.0
            zero = 0.0
        else:
            if self._poles == 2:
                difference = numeric - 2.0 * self._input1 + self._input2
                hp = (
                    self._gain * self._gain * difference
                    + 2.0 * self._decay * self._hp1
                    - self._decay * self._decay * self._hp2
                )
            else:
                hp = self._gain * (numeric - self._input1) + self._decay * self._hp1

            filt = (
                self._c1 * (hp + self._hp1) / 2.0
                + self._c2 * self._filt1
                + self._c3 * self._filt2
            )
            zero = (
                self._gain * (filt - self._filt1) + self._decay * self._zero1
                if self._zero_mean
                else 0.0
            )

        self._input2 = self._input1
        self._input1 = numeric
        self._hp2 = self._hp1
        self._hp1 = hp
        self._filt2 = self._filt1
        self._filt1 = filt
        self._zero1 = zero
        self._count += 1

        output = zero if self._zero_mean else filt
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
