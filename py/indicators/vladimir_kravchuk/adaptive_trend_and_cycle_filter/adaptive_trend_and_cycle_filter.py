"""Adaptive Trend and Cycle Filter (ATCF) indicator by Vladimir Kravchuk.

The suite is a bank of five Finite Impulse Response (FIR) filters applied
to the same input series plus three composite outputs derived from them:

  - FATL (Fast Adaptive Trend Line)       — 39-tap FIR.
  - SATL (Slow Adaptive Trend Line)       — 65-tap FIR.
  - RFTL (Reference Fast Trend Line)      — 44-tap FIR.
  - RSTL (Reference Slow Trend Line)      — 91-tap FIR.
  - RBCI (Range Bound Channel Index)      — 56-tap FIR.
  - FTLM (Fast Trend Line Momentum)       = FATL − RFTL.
  - STLM (Slow Trend Line Momentum)       = SATL − RSTL.
  - PCCI (Perfect Commodity Channel Index) = input − FATL.

Each FIR filter emits NaN until its own window fills. Indicator-level
is_primed mirrors RSTL (the longest pole, 91 samples).
"""

import math
from typing import List, Any

from ...core.indicator import Indicator
from ...core.metadata import Metadata
from ...core.build_metadata import build_metadata, OutputText
from ...core.identifier import Identifier
from ...core.component_triple_mnemonic import component_triple_mnemonic
from ....entities.bar import Bar
from ....entities.quote import Quote
from ....entities.trade import Trade
from ....entities.scalar import Scalar
from ....entities.bar_component import BarComponent, DEFAULT_BAR_COMPONENT, bar_component_value
from ....entities.quote_component import QuoteComponent, DEFAULT_QUOTE_COMPONENT, quote_component_value
from ....entities.trade_component import TradeComponent, DEFAULT_TRADE_COMPONENT, trade_component_value
from .params import AdaptiveTrendAndCycleFilterParams
from .coefficients import (
    FATL_COEFFICIENTS, SATL_COEFFICIENTS, RFTL_COEFFICIENTS,
    RSTL_COEFFICIENTS, RBCI_COEFFICIENTS,
)


class _FirFilter:
    """Internal FIR engine shared by all five ATCF lines."""

    def __init__(self, coeffs: list[float]) -> None:
        self._window = [0.0] * len(coeffs)
        self._coeffs = coeffs
        self._count = 0
        self._primed = False
        self._value = math.nan

    def is_primed(self) -> bool:
        return self._primed

    @property
    def value(self) -> float:
        return self._value

    def update(self, sample: float) -> float:
        if self._primed:
            w = self._window
            w[:len(w) - 1] = w[1:]
            w[-1] = sample

            s = 0.0
            coeffs = self._coeffs
            for i in range(len(w)):
                s += w[i] * coeffs[i]

            self._value = s
            return self._value

        self._window[self._count] = sample
        self._count += 1

        if self._count == len(self._window):
            self._primed = True

            s = 0.0
            w = self._window
            coeffs = self._coeffs
            for i in range(len(w)):
                s += w[i] * coeffs[i]

            self._value = s

        return self._value


class AdaptiveTrendAndCycleFilter(Indicator):
    """Vladimir Kravchuk's combined ATCF suite.

    Exposes eight scalar outputs (five FIR filters + three composites).
    """

    def __init__(self, p: AdaptiveTrendAndCycleFilterParams) -> None:
        bc = p.bar_component if p.bar_component is not None else DEFAULT_BAR_COMPONENT
        qc = p.quote_component if p.quote_component is not None else DEFAULT_QUOTE_COMPONENT
        tc = p.trade_component if p.trade_component is not None else DEFAULT_TRADE_COMPONENT

        self._bar_func = bar_component_value(bc)
        self._quote_func = quote_component_value(qc)
        self._trade_func = trade_component_value(tc)

        cm = component_triple_mnemonic(bc, qc, tc)
        top_arg = cm[2:] if cm else ""
        sub_arg = cm[2:] if cm else ""

        self._mnemonic = f"atcf({top_arg})"

        def mk_sub(name: str, full: str) -> tuple[str, str]:
            m = f"{name}({sub_arg})"
            d = f"{full} {m}"
            return m, d

        self._mnemonic_fatl, self._description_fatl = mk_sub("fatl", "Fast Adaptive Trend Line")
        self._mnemonic_satl, self._description_satl = mk_sub("satl", "Slow Adaptive Trend Line")
        self._mnemonic_rftl, self._description_rftl = mk_sub("rftl", "Reference Fast Trend Line")
        self._mnemonic_rstl, self._description_rstl = mk_sub("rstl", "Reference Slow Trend Line")
        self._mnemonic_rbci, self._description_rbci = mk_sub("rbci", "Range Bound Channel Index")
        self._mnemonic_ftlm, self._description_ftlm = mk_sub("ftlm", "Fast Trend Line Momentum")
        self._mnemonic_stlm, self._description_stlm = mk_sub("stlm", "Slow Trend Line Momentum")
        self._mnemonic_pcci, self._description_pcci = mk_sub("pcci", "Perfect Commodity Channel Index")

        self._description = f"Adaptive trend and cycle filter {self._mnemonic}"

        self._fatl = _FirFilter(FATL_COEFFICIENTS)
        self._satl = _FirFilter(SATL_COEFFICIENTS)
        self._rftl = _FirFilter(RFTL_COEFFICIENTS)
        self._rstl = _FirFilter(RSTL_COEFFICIENTS)
        self._rbci = _FirFilter(RBCI_COEFFICIENTS)

        self._ftlm_value = math.nan
        self._stlm_value = math.nan
        self._pcci_value = math.nan

    def is_primed(self) -> bool:
        return self._rstl.is_primed()

    def metadata(self) -> Metadata:
        return build_metadata(
            Identifier.ADAPTIVE_TREND_AND_CYCLE_FILTER,
            self._mnemonic,
            self._description,
            [
                OutputText(self._mnemonic_fatl, self._description_fatl),
                OutputText(self._mnemonic_satl, self._description_satl),
                OutputText(self._mnemonic_rftl, self._description_rftl),
                OutputText(self._mnemonic_rstl, self._description_rstl),
                OutputText(self._mnemonic_rbci, self._description_rbci),
                OutputText(self._mnemonic_ftlm, self._description_ftlm),
                OutputText(self._mnemonic_stlm, self._description_stlm),
                OutputText(self._mnemonic_pcci, self._description_pcci),
            ],
        )

    def update(self, sample: float) -> tuple[float, float, float, float, float, float, float, float]:
        """Update with a scalar value.

        Returns (fatl, satl, rftl, rstl, rbci, ftlm, stlm, pcci).
        """
        if math.isnan(sample):
            nan = math.nan
            return nan, nan, nan, nan, nan, nan, nan, nan

        fatl = self._fatl.update(sample)
        satl = self._satl.update(sample)
        rftl = self._rftl.update(sample)
        rstl = self._rstl.update(sample)
        rbci = self._rbci.update(sample)

        if self._fatl.is_primed() and self._rftl.is_primed():
            self._ftlm_value = fatl - rftl

        if self._satl.is_primed() and self._rstl.is_primed():
            self._stlm_value = satl - rstl

        if self._fatl.is_primed():
            self._pcci_value = sample - fatl

        return fatl, satl, rftl, rstl, rbci, self._ftlm_value, self._stlm_value, self._pcci_value

    def update_scalar(self, sample: Scalar) -> List[Any]:
        vals = self.update(sample.value)
        return [Scalar(time=sample.time, value=v) for v in vals]

    def update_bar(self, sample: Bar) -> List[Any]:
        v = self._bar_func(sample)
        return self.update_scalar(Scalar(time=sample.time, value=v))

    def update_quote(self, sample: Quote) -> List[Any]:
        v = self._quote_func(sample)
        return self.update_scalar(Scalar(time=sample.time, value=v))

    def update_trade(self, sample: Trade) -> List[Any]:
        v = self._trade_func(sample)
        return self.update_scalar(Scalar(time=sample.time, value=v))
