"""Slope Divergence TSI Filter (SD_TSI) -- William Blau.

A trend/congestion prefilter built on the True Strength Index. It keeps the TSI
value only when the slope of the TSI agrees in sign with the slope of a separate
double EMA of price; otherwise it outputs 0 (a "slope divergence" / congestion
zone):

    ind_k = TSI(close, q, r, s, u)_k                     (triple-smoothed TSI)
    ref_k = DEMA(close, x, y)_k = EMA(EMA(close, x), y)_k (price reference)

    SD_TSI_k = ind_k   if  ind_k - ind_(k-1) > 0  and  ref_k - ref_(k-1) > 0
             = ind_k   if  ind_k - ind_(k-1) < 0  and  ref_k - ref_(k-1) < 0
             = 0       otherwise

The gate is strict (book Fig. B-25): a flat slope on either series yields 0.

Priming convention -- BOOK / EasyLanguage (Option B):
    * The TSI is NaN for bars 0..q-2 (momentum look-back), so SD_TSI is NaN there.
    * The price DEMA seeds at bar 0 and advances every bar, including through
      the TSI warm-up.
    * At the first finite TSI bar there is no prior TSI value, hence no slope,
      so the output is 0.0.

Reference:

Blau, William (1995). Momentum, Direction, and Divergence, ch. 12, Appendix B
Fig. B-25. Wiley.
"""

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
from ....entities.bar_component import DEFAULT_BAR_COMPONENT, bar_component_value
from ....entities.quote_component import DEFAULT_QUOTE_COMPONENT, quote_component_value
from ....entities.trade_component import DEFAULT_TRADE_COMPONENT, trade_component_value
from ..true_strength_index.true_strength_index import TrueStrengthIndex
from ..true_strength_index.params import TrueStrengthIndexParams
from .params import SlopeDivergenceTsiFilterParams


class _Ema:
    """Stateful streaming EMA: alpha = 2/(period+1), seeds e_0 = x_0.

    Inlined verbatim from the Blau exponential moving average so the indicator
    is a standalone porting unit. Do NOT change its numerics.

    period == 1 -> alpha == 1 -> pure passthrough (output == input).
    """

    def __init__(self, period: int) -> None:
        self._alpha = 2.0 / (float(period) + 1.0)
        self._prev = 0.0
        self._primed = False

    def update(self, x: float) -> float:
        if not self._primed:
            self._prev = x
            self._primed = True
            return self._prev
        e = self._alpha * x + (1.0 - self._alpha) * self._prev
        self._prev = e
        return e


class SlopeDivergenceTsiFilter(Indicator):
    """William Blau's Slope Divergence TSI Filter (SD_TSI)."""

    def __init__(self, params: SlopeDivergenceTsiFilterParams) -> None:
        q = params.q
        r = params.r
        s = params.s
        u = params.u
        x = params.x
        y = params.y

        if q < 1:
            raise ValueError(
                "invalid slope divergence tsi filter parameters: "
                "q should be greater than 0")
        if r < 1:
            raise ValueError(
                "invalid slope divergence tsi filter parameters: "
                "r should be greater than 0")
        if s < 1:
            raise ValueError(
                "invalid slope divergence tsi filter parameters: "
                "s should be greater than 0")
        if u < 1:
            raise ValueError(
                "invalid slope divergence tsi filter parameters: "
                "u should be greater than 0")
        if x < 1:
            raise ValueError(
                "invalid slope divergence tsi filter parameters: "
                "x should be greater than 0")
        if y < 1:
            raise ValueError(
                "invalid slope divergence tsi filter parameters: "
                "y should be greater than 0")

        bc = params.bar_component if params.bar_component is not None else DEFAULT_BAR_COMPONENT
        qc = params.quote_component if params.quote_component is not None else DEFAULT_QUOTE_COMPONENT
        tc = params.trade_component if params.trade_component is not None else DEFAULT_TRADE_COMPONENT

        bar_func = bar_component_value(bc)
        quote_func = quote_component_value(qc)
        trade_func = trade_component_value(tc)

        mnemonic = f"sdtsi({q},{r},{s},{u},{x},{y}{component_triple_mnemonic(bc, qc, tc)})"
        description = f"Slope Divergence TSI Filter {mnemonic}"

        self._line = LineIndicator(mnemonic, description, bar_func, quote_func, trade_func, self.update)

        # The TSI oscillator (its signal line is unused, so ul=1).
        self._tsi = TrueStrengthIndex(TrueStrengthIndexParams(q=q, r=r, s=s, u=u, ul=1))

        # Price reference: DEMA(close, x, y) = EMA(EMA(close, x), y).
        self._reference_x = _Ema(x)
        self._reference_y = _Ema(y)

        # Slope state: the previous finite TSI and the previous-bar reference.
        self._previous_tsi = 0.0
        self._has_previous_tsi = False
        self._previous_reference = 0.0

        self._primed = False

    def is_primed(self) -> bool:
        return self._primed

    def metadata(self) -> Metadata:
        return build_metadata(
            Identifier.SLOPE_DIVERGENCE_TSI_FILTER,
            self._line.mnemonic,
            self._line.description,
            [OutputText(self._line.mnemonic, self._line.description)],
        )

    def update(self, sample: float) -> float:
        """Feeds one close and returns this bar's SD_TSI (or NaN during warm-up)."""
        # The price reference advances every bar (it has no NaN warm-up).
        reference = self._reference_y.update(self._reference_x.update(sample))

        tsi, _ = self._tsi.update(sample)
        if math.isnan(tsi):
            # TSI momentum warm-up: keep the previous-bar reference current.
            self._previous_reference = reference
            return math.nan

        if not self._has_previous_tsi:
            # First finite TSI: no prior TSI, hence no slope.
            result = 0.0
        else:
            delta_tsi = tsi - self._previous_tsi
            delta_reference = reference - self._previous_reference

            # Keep the TSI only when both slopes are strictly same-signed.
            if (delta_tsi > 0.0 and delta_reference > 0.0) or \
                    (delta_tsi < 0.0 and delta_reference < 0.0):
                result = tsi
            else:
                result = 0.0

        self._previous_tsi = tsi
        self._has_previous_tsi = True
        self._previous_reference = reference
        self._primed = True
        return result

    def update_scalar(self, sample: Scalar) -> Output:
        return self._line.update_scalar(sample)

    def update_bar(self, sample: Bar) -> Output:
        return self._line.update_bar(sample)

    def update_quote(self, sample: Quote) -> Output:
        return self._line.update_quote(sample)

    def update_trade(self, sample: Trade) -> Output:
        return self._line.update_trade(sample)
