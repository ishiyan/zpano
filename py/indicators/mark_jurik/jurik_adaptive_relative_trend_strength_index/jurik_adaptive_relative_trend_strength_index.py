"""Jurik adaptive relative trend strength index indicator."""

import math

from ...core.indicator import Indicator
from ...core.line_indicator import LineIndicator
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
from .params import JurikAdaptiveRelativeTrendStrengthIndexParams


class JurikAdaptiveRelativeTrendStrengthIndex(Indicator):
    """Computes the Jurik Adaptive RSX indicator (JARSX).

    Combines adaptive length selection (volatility regime detection) with
    the RSX core (triple-cascaded lag-reduced EMA oscillator). Output in [0, 100].
    """

    def __init__(self, params: JurikAdaptiveRelativeTrendStrengthIndexParams) -> None:
        lo_length = params.lo_length
        hi_length = params.hi_length

        if lo_length < 2:
            raise ValueError(
                "invalid jurik adaptive relative trend strength index parameters: "
                "lo_length should be at least 2")
        if hi_length < lo_length:
            raise ValueError(
                "invalid jurik adaptive relative trend strength index parameters: "
                "hi_length should be at least lo_length")

        bc = params.bar_component if params.bar_component is not None else DEFAULT_BAR_COMPONENT
        qc = params.quote_component if params.quote_component is not None else DEFAULT_QUOTE_COMPONENT
        tc = params.trade_component if params.trade_component is not None else DEFAULT_TRADE_COMPONENT

        bar_func = bar_component_value(bc)
        quote_func = quote_component_value(qc)
        trade_func = trade_component_value(tc)

        mnemonic = f"jarsx({lo_length}, {hi_length}{component_triple_mnemonic(bc, qc, tc)})"
        description = f"Jurik adaptive relative trend strength index {mnemonic}"

        self._line = LineIndicator(mnemonic, description, bar_func, quote_func, trade_func, self.update)
        self._primed = False
        self._lo_length = lo_length
        self._hi_length = hi_length
        self._eps = 0.001

        # Price tracking.
        self._bar_count = 0
        self._previous_price = 0.0

        # Adaptive length state: rolling averages of abs(diff).
        # Long window (up to 100), short window (up to 10).
        self._long_buffer = [0.0] * 100
        self._long_index = 0
        self._long_sum = 0.0
        self._long_count = 0
        self._short_buffer = [0.0] * 10
        self._short_index = 0
        self._short_sum = 0.0
        self._short_count = 0

        # RSX core state.
        self._rsx_initialized = False
        self._kg = 0.0
        self._c = 0.0
        self._warmup = 0
        # Signal path accumulators (3 cascaded stages).
        self._sig1_a = 0.0
        self._sig1_b = 0.0
        self._sig2_a = 0.0
        self._sig2_b = 0.0
        self._sig3_a = 0.0
        self._sig3_b = 0.0
        # Denominator path accumulators (3 cascaded stages).
        self._den1_a = 0.0
        self._den1_b = 0.0
        self._den2_a = 0.0
        self._den2_b = 0.0
        self._den3_a = 0.0
        self._den3_b = 0.0

    def is_primed(self) -> bool:
        return self._primed

    def metadata(self) -> Metadata:
        return build_metadata(
            Identifier.JURIK_ADAPTIVE_RELATIVE_TREND_STRENGTH_INDEX,
            self._line.mnemonic,
            self._line.description,
            [OutputText(mnemonic=self._line.mnemonic, description=self._line.description)],
        )

    def update(self, sample: float) -> float:
        """Update the indicator with a new sample value."""
        if math.isnan(sample):
            return sample

        bar = self._bar_count
        self._bar_count += 1

        if bar == 0:
            # First bar: compute adaptive length, initialize RSX, no output.
            self._previous_price = sample
            # value1[0] = 0 (no previous bar to diff against).
            # Add 0 to both rolling buffers.
            self._long_buffer[0] = 0.0
            self._long_sum = 0.0
            self._long_count = 1
            self._short_buffer[0] = 0.0
            self._short_sum = 0.0
            self._short_count = 1

            # Compute adaptive length from bar 0 averages.
            avg1 = 0.0  # mean of [0] = 0
            avg2 = 0.0
            value2 = math.log((self._eps + avg1) / (self._eps + avg2))
            value3 = value2 / (1.0 + abs(value2))
            adaptive_length = self._lo_length + \
                (self._hi_length - self._lo_length) * (1.0 + value3) / 2.0
            length = max(int(adaptive_length), 2)

            self._kg = 3.0 / (length + 2)
            self._c = 1.0 - self._kg
            self._warmup = max(length - 1, 5)
            self._rsx_initialized = True
            return math.nan

        # Bars 1+: compute RSX.
        old_price = self._previous_price
        self._previous_price = sample
        value1 = abs(sample - old_price)

        # Update long rolling buffer.
        if self._long_count < 100:
            self._long_buffer[self._long_count] = value1
            self._long_sum += value1
            self._long_count += 1
        else:
            self._long_sum -= self._long_buffer[self._long_index]
            self._long_buffer[self._long_index] = value1
            self._long_sum += value1
            self._long_index = (self._long_index + 1) % 100

        # Update short rolling buffer.
        if self._short_count < 10:
            self._short_buffer[self._short_count] = value1
            self._short_sum += value1
            self._short_count += 1
        else:
            self._short_sum -= self._short_buffer[self._short_index]
            self._short_buffer[self._short_index] = value1
            self._short_sum += value1
            self._short_index = (self._short_index + 1) % 10

        # RSX core computation.
        mom = 100.0 * (sample - old_price)
        abs_mom = abs(mom)

        kg = self._kg
        c = self._c

        # Signal path — Stage 1.
        self._sig1_a = c * self._sig1_a + kg * mom
        self._sig1_b = kg * self._sig1_a + c * self._sig1_b
        s1 = 1.5 * self._sig1_a - 0.5 * self._sig1_b

        # Signal path — Stage 2.
        self._sig2_a = c * self._sig2_a + kg * s1
        self._sig2_b = kg * self._sig2_a + c * self._sig2_b
        s2 = 1.5 * self._sig2_a - 0.5 * self._sig2_b

        # Signal path — Stage 3.
        self._sig3_a = c * self._sig3_a + kg * s2
        self._sig3_b = kg * self._sig3_a + c * self._sig3_b
        numerator = 1.5 * self._sig3_a - 0.5 * self._sig3_b

        # Denominator path — Stage 1.
        self._den1_a = c * self._den1_a + kg * abs_mom
        self._den1_b = kg * self._den1_a + c * self._den1_b
        d1 = 1.5 * self._den1_a - 0.5 * self._den1_b

        # Denominator path — Stage 2.
        self._den2_a = c * self._den2_a + kg * d1
        self._den2_b = kg * self._den2_a + c * self._den2_b
        d2 = 1.5 * self._den2_a - 0.5 * self._den2_b

        # Denominator path — Stage 3.
        self._den3_a = c * self._den3_a + kg * d2
        self._den3_b = kg * self._den3_a + c * self._den3_b
        denominator = 1.5 * self._den3_a - 0.5 * self._den3_b

        # Output after warmup (bar is 1-based RSX bar index).
        if bar >= self._warmup:
            self._primed = True
            if denominator != 0.0:
                value = (numerator / denominator + 1.0) * 50.0
            else:
                value = 50.0
            return max(0.0, min(100.0, value))

        return math.nan

    def update_bar(self, sample: Bar) -> Output:
        return self._line.update_bar(sample)

    def update_quote(self, sample: Quote) -> Output:
        return self._line.update_quote(sample)

    def update_trade(self, sample: Trade) -> Output:
        return self._line.update_trade(sample)

    def update_scalar(self, sample: Scalar) -> Output:
        return self._line.update_scalar(sample)
