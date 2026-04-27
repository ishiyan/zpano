"""Jurik moving average indicator."""

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
from .params import JurikMovingAverageParams


class JurikMovingAverage(Indicator):
    """Computes the Jurik moving average (JMA).

    The indicator is not primed during the first 30 updates.
    """

    def __init__(self, params: JurikMovingAverageParams) -> None:
        length = params.length
        phase = params.phase

        if length < 1:
            raise ValueError(
                "invalid jurik moving average parameters: length should be positive")
        if phase < -100 or phase > 100:
            raise ValueError(
                "invalid jurik moving average parameters: phase should be in range [-100, 100]")

        bc = params.bar_component if params.bar_component is not None else DEFAULT_BAR_COMPONENT
        qc = params.quote_component if params.quote_component is not None else DEFAULT_QUOTE_COMPONENT
        tc = params.trade_component if params.trade_component is not None else DEFAULT_TRADE_COMPONENT

        bar_func = bar_component_value(bc)
        quote_func = quote_component_value(qc)
        trade_func = trade_component_value(tc)

        mnemonic = f"jma({length}, {phase}{component_triple_mnemonic(bc, qc, tc)})"
        description = f"Jurik moving average {mnemonic}"

        self._line = LineIndicator(mnemonic, description, bar_func, quote_func, trade_func, self.update)

        _EPSILON = 1e-10
        _C_INIT = 1000000.0

        # Initialize arrays.
        self._list = [0.0] * 128
        self._ring = [0.0] * 128
        self._ring2 = [0.0] * 11
        self._buffer = [0.0] * 62

        for i in range(64):
            self._list[i] = -_C_INIT
        for i in range(64, 128):
            self._list[i] = _C_INIT

        f80 = _EPSILON
        if length > 1:
            f80 = (float(length) - 1.0) / 2.0

        self._f10 = float(phase) / 100.0 + 1.5

        v1 = math.log(math.sqrt(f80))
        self._v1 = v1
        self._v2 = v1
        self._v3 = max(v1 / math.log(2.0) + 2.0, 0.0)

        f98 = self._v3
        self._f88 = max(f98 - 2.0, 0.5)
        self._f98 = f98

        f78 = math.sqrt(f80) * f98
        self._f78 = f78
        self._f90 = f78 / (f78 + 1.0)
        f80 *= 0.9
        self._f50 = f80 / (f80 + 2.0)

        # Integer state.
        self._s28 = 63
        self._s30 = 64
        self._s38 = 0
        self._s40 = 0
        self._s48 = 0
        self._s50 = 0
        self._s70 = 0
        self._f0 = 1
        self._fD8 = 0
        self._fF0 = 0
        self._v5 = 0

        # Float state.
        self._s8 = 0.0
        self._s18 = 0.0
        self._f18 = 0.0
        self._f38 = 0.0
        self._f58 = 0.0
        self._fA8 = 0.0
        self._fB8 = 0.0
        self._fC0 = 0.0
        self._fC8 = 0.0
        self._fF8 = 0

        self._primed = False

    def is_primed(self) -> bool:
        return self._primed

    def metadata(self) -> Metadata:
        return build_metadata(
            Identifier.JURIK_MOVING_AVERAGE,
            self._line.mnemonic,
            self._line.description,
            [OutputText(mnemonic=self._line.mnemonic, description=self._line.description)],
        )

    def update(self, sample: float) -> float:
        """Update the indicator with a new sample value."""
        if math.isnan(sample):
            return sample

        if self._fF0 < 61:
            self._fF0 += 1
            self._buffer[self._fF0] = sample

        if self._fF0 <= 30:
            return math.nan

        self._primed = True

        if self._f0 == 0:
            self._fD8 = 0
        else:
            self._f0 = 0
            self._v5 = 0

            for i in range(1, 30):
                if self._buffer[i + 1] != self._buffer[i]:
                    self._v5 = 1

            self._fD8 = self._v5 * 30
            if self._fD8 == 0:
                self._f38 = sample
            else:
                self._f38 = self._buffer[1]

            self._f18 = self._f38
            if self._fD8 > 29:
                self._fD8 = 29

        for i in range(self._fD8, -1, -1):
            f8 = sample
            if i != 0:
                f8 = self._buffer[31 - i]

            f28 = f8 - self._f18
            f48 = f8 - self._f38
            a28 = abs(f28)
            a48 = abs(f48)
            self._v2 = max(a28, a48)

            fA0 = self._v2
            v = fA0 + 1e-10

            if self._s48 <= 1:
                self._s48 = 127
            else:
                self._s48 -= 1

            if self._s50 <= 1:
                self._s50 = 10
            else:
                self._s50 -= 1

            if self._s70 < 128:
                self._s70 += 1

            self._s8 += v - self._ring2[self._s50]
            self._ring2[self._s50] = v
            s20 = self._s8 / float(self._s70)

            if self._s70 > 10:
                s20 = self._s8 / 10.0

            s58 = 0
            s68 = 0

            if self._s70 > 127:
                s10 = self._ring[self._s48]
                self._ring[self._s48] = s20
                s68 = 64
                s58 = s68

                while s68 > 1:
                    if self._list[s58] < s10:
                        s68 //= 2
                        s58 += s68
                    elif self._list[s58] <= s10:
                        s68 = 1
                    else:
                        s68 //= 2
                        s58 -= s68
            else:
                self._ring[self._s48] = s20
                if self._s28 + self._s30 > 127:
                    self._s30 -= 1
                    s58 = self._s30
                else:
                    self._s28 += 1
                    s58 = self._s28

                self._s38 = min(self._s28, 96)
                self._s40 = max(self._s30, 32)

            s68 = 64
            s60 = s68

            while s68 > 1:
                if self._list[s60] >= s20:
                    if self._list[s60 - 1] <= s20:
                        s68 = 1
                    else:
                        s68 //= 2
                        s60 -= s68
                else:
                    s68 //= 2
                    s60 += s68

                if s60 == 127 and s20 > self._list[127]:
                    s60 = 128

            if self._s70 > 127:
                if s58 >= s60:
                    if self._s38 + 1 > s60 and self._s40 - 1 < s60:
                        self._s18 += s20
                    elif self._s40 > s60 and self._s40 - 1 < s58:
                        self._s18 += self._list[self._s40 - 1]
                elif self._s40 >= s60:
                    if self._s38 + 1 < s60 and self._s38 + 1 > s58:
                        self._s18 += self._list[self._s38 + 1]
                elif self._s38 + 2 > s60:
                    self._s18 += s20
                elif self._s38 + 1 < s60 and self._s38 + 1 > s58:
                    self._s18 += self._list[self._s38 + 1]

                if s58 > s60:
                    if self._s40 - 1 < s58 and self._s38 + 1 > s58:
                        self._s18 -= self._list[s58]
                    elif self._s38 < s58 and self._s38 + 1 > s60:
                        self._s18 -= self._list[self._s38]
                else:
                    if self._s38 + 1 > s58 and self._s40 - 1 < s58:
                        self._s18 -= self._list[s58]
                    elif self._s40 > s58 and self._s40 < s60:
                        self._s18 -= self._list[self._s40]

            if s58 <= s60:
                if s58 >= s60:
                    self._list[s60] = s20
                else:
                    for k in range(s58 + 1, s60):
                        self._list[k - 1] = self._list[k]
                    self._list[s60 - 1] = s20
            else:
                for k in range(s58 - 1, s60 - 1, -1):
                    self._list[k + 1] = self._list[k]
                self._list[s60] = s20

            if self._s70 < 128:
                self._s18 = 0.0
                for k in range(self._s40, self._s38 + 1):
                    self._s18 += self._list[k]

            f60 = self._s18 / float(self._s38 - self._s40 + 1)

            if self._fF8 + 1 > 31:
                self._fF8 = 31
            else:
                self._fF8 += 1

            if self._fF8 <= 30:
                if f28 > 0:
                    self._f18 = f8
                else:
                    self._f18 = f8 - f28 * self._f90

                if f48 < 0:
                    self._f38 = f8
                else:
                    self._f38 = f8 - f48 * self._f90

                self._fB8 = sample
                if self._fF8 != 30:
                    continue

                v4 = 1
                self._fC0 = sample

                if math.ceil(self._f78) >= 1:
                    v4 = int(math.ceil(self._f78))

                v2 = 1
                fE8 = v4

                if math.floor(self._f78) >= 1:
                    v2 = int(math.floor(self._f78))

                f68 = 1.0
                fE0 = v2

                if fE8 != fE0:
                    v4 = fE8 - fE0
                    f68 = (self._f78 - float(fE0)) / float(v4)

                v5 = min(fE0, 29)
                v6 = min(fE8, 29)
                self._fA8 = (sample - self._buffer[self._fF0 - v5]) * (1 - f68) / float(fE0) + \
                    (sample - self._buffer[self._fF0 - v6]) * f68 / float(fE8)
            else:
                p = math.pow(fA0 / f60, self._f88)
                self._v1 = min(self._f98, p)

                if self._v1 < 1:
                    self._v2 = 1
                else:
                    self._v3 = min(self._f98, p)
                    self._v2 = self._v3

                self._f58 = self._v2
                f70 = math.pow(self._f90, math.sqrt(self._f58))

                if f28 > 0:
                    self._f18 = f8
                else:
                    self._f18 = f8 - f28 * f70

                if f48 < 0:
                    self._f38 = f8
                else:
                    self._f38 = f8 - f48 * f70

        if self._fF8 > 30:
            f30 = math.pow(self._f50, self._f58)
            self._fC0 = (1.0 - f30) * sample + f30 * self._fC0
            self._fC8 = (sample - self._fC0) * (1.0 - self._f50) + self._f50 * self._fC8
            fD0 = self._f10 * self._fC8 + self._fC0
            f20 = f30 * -2.0
            f40 = f30 * f30
            fB0 = f20 + f40 + 1.0
            self._fA8 = (fD0 - self._fB8) * fB0 + f40 * self._fA8
            self._fB8 += self._fA8

        return self._fB8

    def update_bar(self, bar: Bar) -> Output:
        return self._line.update_bar(bar)

    def update_quote(self, quote: Quote) -> Output:
        return self._line.update_quote(quote)

    def update_trade(self, trade: Trade) -> Output:
        return self._line.update_trade(trade)

    def update_scalar(self, scalar: Scalar) -> Output:
        return self._line.update_scalar(scalar)
