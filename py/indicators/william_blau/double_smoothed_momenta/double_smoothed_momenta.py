"""Double-Smoothed Momenta (DM) / Double-Smoothed RSI (DRSI) -- William Blau.

A close-based, double-smoothed momentum oscillator bounded to [0, 100]:

    LCa_k = min(Close over the last a bars)          (lowest close)
    HCa_k = max(Close over the last a bars)          (highest close)
    st_k  = Close_k - LCa_k                          (close above the low close)
    rng_k = HCa_k - LCa_k                            (a-bar close range)

    DM(a, y, z) = 100 * EMA(EMA(st, y), z) / EMA(EMA(rng, y), z)

i.e. each of the numerator (st) and denominator (rng) series is double-smoothed
by an inner EMA of period ``y`` then an outer EMA of period ``z`` (Blau's
Ez(Ey(.)) ), and the ratio is scaled by 100.

This is structurally the Double-Smoothed Stochastic computed on the CLOSE -- it
uses the highest/lowest *close* over ``a`` bars instead of the high/low of the
bar -- and it has NO signal line (a single output per bar).

Named instances:
    * RSI equivalence:     DM(2, 1, z) == RSI(z), the EMA-form RSI.
    * Double-smoothed RSI: DRSI(y, z) = DM(2, y, z).

The equivalence DM(2,1,z) == RSI(z) holds for the EMA-form RSI built from the
Blau EMA (alpha = 2/(z+1)), NOT Wilder's classic RSI (which uses RMA smoothing,
alpha = 1/z).

Priming convention -- BOOK / EasyLanguage (Option B):
    * st/rng are valid once ``a`` closes exist (bar a-1); all four EMA stages
      seed there. DM is NaN for bars 0..a-2 and finite from bar a-1.
    * For a == 1 there is no NaN warm-up, but the a-bar close range is then
      always 0, so DM is 0.0 on every bar via the guard (a degenerate setting).

Division guard: EMA(EMA(rng)) <= 0 -> DM = 0.0.
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
from .params import DoubleSmoothedMomentaParams


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


class DoubleSmoothedMomenta(Indicator):
    """William Blau's Double-Smoothed Momenta (DM) oscillator."""

    def __init__(self, params: DoubleSmoothedMomentaParams) -> None:
        a = params.a
        y = params.y
        z = params.z

        if a < 1:
            raise ValueError(
                "invalid double smoothed momenta parameters: "
                "a should be greater than 0")
        if y < 1:
            raise ValueError(
                "invalid double smoothed momenta parameters: "
                "y should be greater than 0")
        if z < 1:
            raise ValueError(
                "invalid double smoothed momenta parameters: "
                "z should be greater than 0")

        bc = params.bar_component if params.bar_component is not None else DEFAULT_BAR_COMPONENT
        qc = params.quote_component if params.quote_component is not None else DEFAULT_QUOTE_COMPONENT
        tc = params.trade_component if params.trade_component is not None else DEFAULT_TRADE_COMPONENT

        bar_func = bar_component_value(bc)
        quote_func = quote_component_value(qc)
        trade_func = trade_component_value(tc)

        mnemonic = f"dm({a},{y},{z}{component_triple_mnemonic(bc, qc, tc)})"
        description = f"Double-Smoothed Momenta {mnemonic}"

        self._line = LineIndicator(mnemonic, description, bar_func, quote_func, trade_func, self.update)

        # Rolling window of the last a closes (for the highest/lowest close).
        self._window: list[float] = [0.0] * a
        self._window_length: int = a
        self._window_count: int = 0
        self._last_index: int = a - 1

        # Two independent 2-stage EMA cascades (double smoothing), each wired
        # inner(y) -> outer(z): EMA(EMA(x, y), z).
        self._numerator_y = _Ema(y)
        self._numerator_z = _Ema(z)
        self._denominator_y = _Ema(y)
        self._denominator_z = _Ema(z)

        self._primed: bool = False

    def is_primed(self) -> bool:
        return self._primed

    def metadata(self) -> Metadata:
        return build_metadata(
            Identifier.DOUBLE_SMOOTHED_MOMENTA,
            self._line.mnemonic,
            self._line.description,
            [OutputText(self._line.mnemonic, self._line.description)],
        )

    def update(self, sample: float) -> float:
        """Feeds one close and returns this bar's DM (or NaN during warm-up)."""
        if self._primed:
            for i in range(self._last_index):
                self._window[i] = self._window[i + 1]

            self._window[self._last_index] = sample
        else:
            self._window[self._window_count] = sample
            self._window_count += 1

            # Need a closes before the highest/lowest close is defined. While
            # unprimed, the EMA cascades must NOT advance (they seed at bar a-1).
            if self._window_length > self._window_count:
                return math.nan

            self._primed = True

        # Highest/lowest close over the last a bars.
        hc = self._window[0]
        lc = self._window[0]

        for i in range(1, self._window_length):
            v = self._window[i]

            if v > hc:
                hc = v

            if v < lc:
                lc = v

        # Raw close-above-low (>= 0) and a-bar close range (>= 0).
        st = sample - lc
        rng = hc - lc

        # Double-smooth each separately (inner y, then outer z), then divide.
        num = self._numerator_z.update(self._numerator_y.update(st))
        den = self._denominator_z.update(self._denominator_y.update(rng))

        # Division guard: smoothed range <= 0 -> DM = 0.0.
        return 0.0 if den <= 0.0 else 100.0 * num / den

    def update_scalar(self, sample: Scalar) -> Output:
        return self._line.update_scalar(sample)

    def update_bar(self, sample: Bar) -> Output:
        return self._line.update_bar(sample)

    def update_quote(self, sample: Quote) -> Output:
        return self._line.update_quote(sample)

    def update_trade(self, sample: Trade) -> Output:
        return self._line.update_trade(sample)
