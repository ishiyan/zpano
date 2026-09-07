"""Ergodic Oscillator -- William Blau.

The Ergodic oscillator is the True Strength Index plotted together with a
signal line -- the EMA of the oscillator that Blau introduces as the trading
vehicle for the TSI (ch. 2, Fig. 2-14):

    ergodic_k = TSI(q, r, s, u)_k                    (the oscillator)
    signal_k  = EMA(ergodic, ul)_k                   (ul-period EMA of it)

It is a TWO-output indicator: each update returns ``(ergodic, signal)``.

The numerics are exactly those of the True Strength Index, so this indicator
wraps a :class:`TrueStrengthIndex` instance instead of duplicating the triple
EMA cascade. Only the mnemonic, the identifier, and the output naming differ.

Priming convention -- BOOK / EasyLanguage (Option B), see description.md §2:
    * Both outputs are NaN for bars 0..q-2 and finite from bar q-1 onward.
    * The signal EMA seeds on the first finite oscillator value (bar q-1).
    * ul == 1 -> signal is a passthrough -> signal == ergodic for every bar.

Division guard: denominator == 0 -> oscillator 0.0 (matches Blau_Ergodic.mq5).
"""

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
from ..true_strength_index.true_strength_index import TrueStrengthIndex
from ..true_strength_index.params import TrueStrengthIndexParams
from .params import ErgodicOscillatorParams


class ErgodicOscillator(Indicator):
    """William Blau's Ergodic Oscillator: the TSI with its EMA signal line."""

    def __init__(self, p: ErgodicOscillatorParams) -> None:
        bc = p.bar_component if p.bar_component is not None else DEFAULT_BAR_COMPONENT
        qc = p.quote_component if p.quote_component is not None else DEFAULT_QUOTE_COMPONENT
        tc = p.trade_component if p.trade_component is not None else DEFAULT_TRADE_COMPONENT

        self._bar_func = bar_component_value(bc)
        self._quote_func = quote_component_value(qc)
        self._trade_func = trade_component_value(tc)

        # The oscillator and its signal line are exactly the True Strength Index
        # outputs; wrap an instance rather than duplicating its numerics. The
        # resolved components are passed through so both agree on the mnemonic.
        self._tsi = TrueStrengthIndex(TrueStrengthIndexParams(
            q=p.q, r=p.r, s=p.s, u=p.u, ul=p.ul,
            bar_component=bc, quote_component=qc, trade_component=tc,
        ))

        self._mnemonic = f"ergodic({p.q},{p.r},{p.s},{p.u},{p.ul}" \
                         f"{component_triple_mnemonic(bc, qc, tc)})"

    def is_primed(self) -> bool:
        return self._tsi.is_primed()

    def metadata(self) -> Metadata:
        desc = f"Ergodic Oscillator {self._mnemonic}"
        return build_metadata(
            Identifier.ERGODIC_OSCILLATOR,
            self._mnemonic,
            desc,
            [
                OutputText(f"{self._mnemonic} ergodic", f"{desc} ergodic"),
                OutputText(f"{self._mnemonic} signal", f"{desc} signal"),
            ],
        )

    def update(self, price: float) -> tuple[float, float]:
        """Update with a scalar value. Returns (ergodic, signal)."""
        return self._tsi.update(price)

    def update_scalar(self, sample: Scalar) -> List[Any]:
        ergodic, signal = self.update(sample.value)
        return [
            Scalar(time=sample.time, value=ergodic),
            Scalar(time=sample.time, value=signal),
        ]

    def update_bar(self, sample: Bar) -> List[Any]:
        v = self._bar_func(sample)
        return self.update_scalar(Scalar(time=sample.time, value=v))

    def update_quote(self, sample: Quote) -> List[Any]:
        v = self._quote_func(sample)
        return self.update_scalar(Scalar(time=sample.time, value=v))

    def update_trade(self, sample: Trade) -> List[Any]:
        v = self._trade_func(sample)
        return self.update_scalar(Scalar(time=sample.time, value=v))
