"""Jurik relative trend strength index indicator."""

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
from .params import JurikRelativeTrendStrengthIndexParams


class JurikRelativeTrendStrengthIndex(Indicator):
    """Computes the Jurik RSX indicator.

    RSX is a noise-free version of RSI based on triple-smoothed EMA of
    momentum and absolute momentum.
    """

    def __init__(self, params: JurikRelativeTrendStrengthIndexParams) -> None:
        length = params.length

        if length < 2:
            raise ValueError(
                "invalid jurik relative trend strength index parameters: "
                "length should be at least 2")

        bc = params.bar_component if params.bar_component is not None else DEFAULT_BAR_COMPONENT
        qc = params.quote_component if params.quote_component is not None else DEFAULT_QUOTE_COMPONENT
        tc = params.trade_component if params.trade_component is not None else DEFAULT_TRADE_COMPONENT

        bar_func = bar_component_value(bc)
        quote_func = quote_component_value(qc)
        trade_func = trade_component_value(tc)

        mnemonic = f"rsx({length}{component_triple_mnemonic(bc, qc, tc)})"
        description = f"Jurik relative trend strength index {mnemonic}"

        self._line = LineIndicator(mnemonic, description, bar_func, quote_func, trade_func, self.update)
        self._primed = False
        self._param_len = length

        # State variables.
        self._f0 = 0
        self._f88 = 0
        self._f90 = 0

        self._f8 = 0.0
        self._f10 = 0.0
        self._f18 = 0.0
        self._f20 = 0.0
        self._f28 = 0.0
        self._f30 = 0.0
        self._f38 = 0.0
        self._f40 = 0.0
        self._f48 = 0.0
        self._f50 = 0.0
        self._f58 = 0.0
        self._f60 = 0.0
        self._f68 = 0.0
        self._f70 = 0.0
        self._f78 = 0.0
        self._f80 = 0.0

    def is_primed(self) -> bool:
        return self._primed

    def metadata(self) -> Metadata:
        return build_metadata(
            Identifier.JURIK_RELATIVE_TREND_STRENGTH_INDEX,
            self._line.mnemonic,
            self._line.description,
            [OutputText(mnemonic=self._line.mnemonic, description=self._line.description)],
        )

    def update(self, sample: float) -> float:
        """Update the indicator with a new sample value."""
        if math.isnan(sample):
            return sample

        hundred = 100.0
        fifty = 50.0
        one_five = 1.5
        half = 0.5
        min_len = 5
        eps = 1e-10

        length = self._param_len

        if self._f90 == 0:
            # First call: initialize.
            self._f90 = 1
            self._f0 = 0

            if length - 1 >= min_len:
                self._f88 = length - 1
            else:
                self._f88 = min_len

            self._f8 = hundred * sample
            self._f18 = 3.0 / float(length + 2)
            self._f20 = 1 - self._f18
        else:
            if self._f88 <= self._f90:
                self._f90 = self._f88 + 1
            else:
                self._f90 += 1

            self._f10 = self._f8
            self._f8 = hundred * sample
            v8 = self._f8 - self._f10

            self._f28 = self._f20 * self._f28 + self._f18 * v8
            self._f30 = self._f18 * self._f28 + self._f20 * self._f30
            v_c = self._f28 * one_five - self._f30 * half

            self._f38 = self._f20 * self._f38 + self._f18 * v_c
            self._f40 = self._f18 * self._f38 + self._f20 * self._f40
            v10 = self._f38 * one_five - self._f40 * half

            self._f48 = self._f20 * self._f48 + self._f18 * v10
            self._f50 = self._f18 * self._f48 + self._f20 * self._f50
            v14 = self._f48 * one_five - self._f50 * half

            self._f58 = self._f20 * self._f58 + self._f18 * abs(v8)
            self._f60 = self._f18 * self._f58 + self._f20 * self._f60
            v18 = self._f58 * one_five - self._f60 * half

            self._f68 = self._f20 * self._f68 + self._f18 * v18
            self._f70 = self._f18 * self._f68 + self._f20 * self._f70
            v1c = self._f68 * one_five - self._f70 * half

            self._f78 = self._f20 * self._f78 + self._f18 * v1c
            self._f80 = self._f18 * self._f78 + self._f20 * self._f80
            v20 = self._f78 * one_five - self._f80 * half

            if self._f88 >= self._f90 and self._f8 != self._f10:
                self._f0 = 1

            if self._f88 == self._f90 and self._f0 == 0:
                self._f90 = 0

            if self._f88 < self._f90 and v20 > eps:
                v4 = (v14 / v20 + 1) * fifty
                if v4 > hundred:
                    v4 = hundred
                if v4 < 0:
                    v4 = 0

                self._primed = True
                return v4

        # During warmup or when denominator is too small.
        if self._f88 < self._f90:
            self._primed = True

        if not self._primed:
            return math.nan

        return fifty

    def update_bar(self, sample: Bar) -> Output:
        return self._line.update_bar(sample)

    def update_quote(self, sample: Quote) -> Output:
        return self._line.update_quote(sample)

    def update_trade(self, sample: Trade) -> Output:
        return self._line.update_trade(sample)

    def update_scalar(self, sample: Scalar) -> Output:
        return self._line.update_scalar(sample)
