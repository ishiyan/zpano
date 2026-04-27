"""Ehlers' Sine Wave indicator."""

import math
from typing import Optional

from ...core.indicator import Indicator
from ...core.metadata import Metadata
from ...core.build_metadata import build_metadata, OutputText
from ...core.identifier import Identifier
from ...core.component_triple_mnemonic import component_triple_mnemonic
from ...core.output import Output
from ...core.outputs.band import Band
from ....entities.bar import Bar
from ....entities.quote import Quote
from ....entities.trade import Trade
from ....entities.scalar import Scalar
from ....entities.bar_component import BarComponent, DEFAULT_BAR_COMPONENT, bar_component_value
from ....entities.quote_component import QuoteComponent, DEFAULT_QUOTE_COMPONENT, quote_component_value
from ....entities.trade_component import TradeComponent, DEFAULT_TRADE_COMPONENT, trade_component_value
from ..hilbert_transformer.cycle_estimator_type import CycleEstimatorType
from ..hilbert_transformer.cycle_estimator_params import CycleEstimatorParams
from ..hilbert_transformer.estimator import new_cycle_estimator, estimator_moniker
from ..dominant_cycle.dominant_cycle import DominantCycle
from ..dominant_cycle.params import DominantCycleParams
from .params import SineWaveParams

# Degrees to radians.
_DEG2RAD = math.pi / 180.0
_LEAD_OFFSET = 45.0


class SineWave(Indicator):
    """Ehlers' Sine Wave indicator.

    Exposes five outputs:
      - Value: sin(phase * deg2rad)
      - Lead: sin((phase + 45) * deg2rad)
      - Band: upper=Value, lower=Lead
      - DominantCyclePeriod: smoothed dominant cycle period
      - DominantCyclePhase: dominant cycle phase in degrees
    """

    def __init__(self, dc: DominantCycle,
                 mnemonic: str, description: str,
                 mnemonic_lead: str, description_lead: str,
                 mnemonic_band: str, description_band: str,
                 mnemonic_dcp: str, description_dcp: str,
                 mnemonic_dc_phase: str, description_dc_phase: str,
                 bar_func, quote_func, trade_func) -> None:
        self._dc = dc
        self._mnemonic = mnemonic
        self._description = description
        self._mnemonic_lead = mnemonic_lead
        self._description_lead = description_lead
        self._mnemonic_band = mnemonic_band
        self._description_band = description_band
        self._mnemonic_dcp = mnemonic_dcp
        self._description_dcp = description_dcp
        self._mnemonic_dc_phase = mnemonic_dc_phase
        self._description_dc_phase = description_dc_phase
        self._bar_func = bar_func
        self._quote_func = quote_func
        self._trade_func = trade_func
        self._primed = False
        self._value = math.nan
        self._lead = math.nan

    @staticmethod
    def create(params: SineWaveParams) -> 'SineWave':
        """Creates a new SineWave from parameters."""
        return _new_sine_wave(
            params.estimator_type, params.estimator_params,
            params.alpha_ema_period_additional,
            params.bar_component, params.quote_component, params.trade_component)

    @staticmethod
    def create_default() -> 'SineWave':
        """Creates a new SineWave with default parameters."""
        return _new_sine_wave(
            CycleEstimatorType.HOMODYNE_DISCRIMINATOR,
            CycleEstimatorParams(smoothing_length=4, alpha_ema_quadrature_in_phase=0.2,
                                 alpha_ema_period=0.2, warm_up_period=100),
            0.33, None, None, None)

    def is_primed(self) -> bool:
        return self._primed

    def metadata(self) -> Metadata:
        return build_metadata(
            Identifier.SINE_WAVE,
            self._mnemonic,
            self._description,
            [
                OutputText(self._mnemonic, self._description),
                OutputText(self._mnemonic_lead, self._description_lead),
                OutputText(self._mnemonic_band, self._description_band),
                OutputText(self._mnemonic_dcp, self._description_dcp),
                OutputText(self._mnemonic_dc_phase, self._description_dc_phase),
            ],
        )

    def update(self, sample: float) -> tuple[float, float, float, float]:
        """Updates, returning (value, lead, period, phase). NaN if not primed."""
        if math.isnan(sample):
            return (sample, sample, sample, sample)

        _, period, phase = self._dc.update(sample)

        if math.isnan(phase):
            nan = math.nan
            return (nan, nan, nan, nan)

        self._primed = True
        self._value = math.sin(phase * _DEG2RAD)
        self._lead = math.sin((phase + _LEAD_OFFSET) * _DEG2RAD)

        return (self._value, self._lead, period, phase)

    def update_scalar(self, sample: Scalar) -> Output:
        return self._update_entity(sample.time, sample.value)

    def update_bar(self, sample: Bar) -> Output:
        return self._update_entity(sample.time, self._bar_func(sample))

    def update_quote(self, sample: Quote) -> Output:
        return self._update_entity(sample.time, self._quote_func(sample))

    def update_trade(self, sample: Trade) -> Output:
        return self._update_entity(sample.time, self._trade_func(sample))

    def _update_entity(self, time, sample: float) -> Output:
        value, lead, period, phase = self.update(sample)
        # Band: upper=value, lower=lead (no sorting — bypass Band.__init__).
        band = object.__new__(Band)
        band.time = time
        band.upper = value
        band.lower = lead
        return [
            Scalar(time=time, value=value),
            Scalar(time=time, value=lead),
            band,
            Scalar(time=time, value=period),
            Scalar(time=time, value=phase),
        ]


def _new_sine_wave(
    estimator_type: CycleEstimatorType,
    estimator_params: CycleEstimatorParams,
    alpha_ema_period_additional: float,
    bc: Optional[BarComponent],
    qc: Optional[QuoteComponent],
    tc: Optional[TradeComponent],
) -> 'SineWave':
    invalid = "invalid sine wave parameters"

    if alpha_ema_period_additional <= 0.0 or alpha_ema_period_additional > 1.0:
        raise ValueError(f"{invalid}: \u03b1 for additional smoothing should be in range (0, 1]")

    # SineWave defaults to BarMedianPrice (not framework default).
    bc_resolved = bc if bc is not None else BarComponent.MEDIAN
    qc_resolved = qc if qc is not None else DEFAULT_QUOTE_COMPONENT
    tc_resolved = tc if tc is not None else DEFAULT_TRADE_COMPONENT

    # Build inner DominantCycle with explicit components.
    dc_params = DominantCycleParams(
        estimator_type=estimator_type,
        estimator_params=estimator_params,
        alpha_ema_period_additional=alpha_ema_period_additional,
        bar_component=bc_resolved,
        quote_component=qc_resolved,
        trade_component=tc_resolved,
    )

    try:
        dc = DominantCycle.create(dc_params)
    except ValueError as e:
        raise ValueError(f"{invalid}: {e}") from e

    # Compose the estimator moniker (same logic as DominantCycle).
    est_moniker = ""
    if estimator_type != CycleEstimatorType.HOMODYNE_DISCRIMINATOR or \
       estimator_params.smoothing_length != 4 or \
       estimator_params.alpha_ema_quadrature_in_phase != 0.2 or \
       estimator_params.alpha_ema_period != 0.2:
        estimator = new_cycle_estimator(estimator_type, estimator_params)
        est_moniker = estimator_moniker(estimator_type, estimator)
        if est_moniker:
            est_moniker = ", " + est_moniker

    comp = component_triple_mnemonic(bc_resolved, qc_resolved, tc_resolved)

    mn_val = f"sw({alpha_ema_period_additional:.3f}{est_moniker}{comp})"
    mn_lead = f"sw-lead({alpha_ema_period_additional:.3f}{est_moniker}{comp})"
    mn_band = f"sw-band({alpha_ema_period_additional:.3f}{est_moniker}{comp})"
    mn_dcp = f"dcp({alpha_ema_period_additional:.3f}{est_moniker}{comp})"
    mn_dcph = f"dcph({alpha_ema_period_additional:.3f}{est_moniker}{comp})"

    bar_func = bar_component_value(bc_resolved)
    quote_func = quote_component_value(qc_resolved)
    trade_func = trade_component_value(tc_resolved)

    return SineWave(
        dc, mn_val, f"Sine wave {mn_val}",
        mn_lead, f"Sine wave lead {mn_lead}",
        mn_band, f"Sine wave band {mn_band}",
        mn_dcp, f"Dominant cycle period {mn_dcp}",
        mn_dcph, f"Dominant cycle phase {mn_dcph}",
        bar_func, quote_func, trade_func,
    )
