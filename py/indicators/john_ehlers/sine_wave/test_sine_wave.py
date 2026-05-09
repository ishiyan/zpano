"""Tests for the SineWave indicator."""

import math
import unittest

from py.indicators.john_ehlers.sine_wave.sine_wave import SineWave
from py.indicators.john_ehlers.sine_wave.params import SineWaveParams
from py.indicators.john_ehlers.sine_wave.output import SineWaveOutput
from py.indicators.core.identifier import Identifier
from py.indicators.core.outputs.band import Band
from py.entities.scalar import Scalar
from py.entities.bar import Bar
from py.entities.quote import Quote
from py.entities.trade import Trade
from py.indicators.john_ehlers.hilbert_transformer.cycle_estimator_type import CycleEstimatorType
from py.indicators.john_ehlers.hilbert_transformer.cycle_estimator_params import CycleEstimatorParams
from py.entities.bar_component import BarComponent
from py.entities.quote_component import QuoteComponent
from py.entities.trade_component import TradeComponent

from .test_testdata import (
    _TIME,
    _INPUT,
    _EXPECTED_PERIOD,
    _EXPECTED_PHASE,
    _EXPECTED_SINE,
    _EXPECTED_SINE_LEAD,
    _TOLERANCE,
    _SKIP,
    _SETTLE_SKIP,
)


# Input data from TA-Lib test_MAMA.xsl, 252 entries.
# Expected period data from test_MAMA.xsl, smoothed, 252 entries.
# Expected phase data from test_HT.xsl, 252 entries.
# Expected sine wave values, 252 entries.
# Expected sine wave lead values, 252 entries.
def _phase_diff(a: float, b: float) -> float:
    """Shortest signed angular difference in (-180, 180]."""
    d = (a - b) % 360
    if d > 180:
        d -= 360
    elif d <= -180:
        d += 360
    return d


def _create_default() -> SineWave:
    return SineWave.create_default()


def _create_cycle_estimator_params() -> CycleEstimatorParams:
    return CycleEstimatorParams(
        smoothing_length=4, alpha_ema_quadrature_in_phase=0.2,
        alpha_ema_period=0.2, warm_up_period=0)


def _create_alpha(alpha: float, estimator_type: CycleEstimatorType) -> SineWave:
    params = SineWaveParams(
        alpha_ema_period_additional=alpha,
        estimator_type=estimator_type,
        estimator_params=_create_cycle_estimator_params(),
    )
    return SineWave.create(params)


class TestSineWaveUpdate(unittest.TestCase):
    def test_reference_sine(self):
        sw = _create_default()
        for i in range(_SKIP, len(_INPUT)):
            value, _, _, _ = sw.update(_INPUT[i])
            if math.isnan(value) or i < _SETTLE_SKIP:
                continue
            if math.isnan(_EXPECTED_SINE[i]):
                continue
            self.assertAlmostEqual(_EXPECTED_SINE[i], value, delta=_TOLERANCE,
                                   msg=f"[{i}] sine")

    def test_reference_sine_lead(self):
        sw = _create_default()
        for i in range(_SKIP, len(_INPUT)):
            _, lead, _, _ = sw.update(_INPUT[i])
            if math.isnan(lead) or i < _SETTLE_SKIP:
                continue
            if math.isnan(_EXPECTED_SINE_LEAD[i]):
                continue
            self.assertAlmostEqual(_EXPECTED_SINE_LEAD[i], lead, delta=_TOLERANCE,
                                   msg=f"[{i}] sine lead")

    def test_reference_period(self):
        sw = _create_default()
        for i in range(_SKIP, len(_INPUT)):
            _, _, period, _ = sw.update(_INPUT[i])
            if math.isnan(period) or i < _SETTLE_SKIP:
                continue
            self.assertAlmostEqual(_EXPECTED_PERIOD[i], period, delta=_TOLERANCE,
                                   msg=f"[{i}] period")

    def test_reference_phase(self):
        sw = _create_default()
        for i in range(_SKIP, len(_INPUT)):
            _, _, _, phase = sw.update(_INPUT[i])
            if math.isnan(phase) or i < _SETTLE_SKIP:
                continue
            if math.isnan(_EXPECTED_PHASE[i]):
                continue
            self.assertAlmostEqual(0.0, abs(_phase_diff(_EXPECTED_PHASE[i], phase)),
                                   delta=_TOLERANCE, msg=f"[{i}] phase")

    def test_nan_input(self):
        sw = _create_default()
        value, lead, period, phase = sw.update(math.nan)
        self.assertTrue(math.isnan(value))
        self.assertTrue(math.isnan(lead))
        self.assertTrue(math.isnan(period))
        self.assertTrue(math.isnan(phase))


class TestSineWaveUpdateEntity(unittest.TestCase):
    _PRIME_COUNT = 200

    def _check_output(self, out):
        self.assertEqual(5, len(out))
        for i in [0, 1, 3, 4]:
            self.assertIsInstance(out[i], Scalar)
            self.assertEqual(_TIME, out[i].time)
        self.assertIsInstance(out[2], Band)
        self.assertEqual(_TIME, out[2].time)

    def test_update_scalar(self):
        sw = _create_default()
        for i in range(self._PRIME_COUNT):
            sw.update(_INPUT[i % len(_INPUT)])
        s = Scalar(time=_TIME, value=100.0)
        self._check_output(sw.update_scalar(s))

    def test_update_bar(self):
        sw = _create_default()
        for i in range(self._PRIME_COUNT):
            sw.update(_INPUT[i % len(_INPUT)])
        b = Bar(time=_TIME, open=0.0, high=100.0, low=100.0, close=100.0, volume=0.0)
        self._check_output(sw.update_bar(b))

    def test_update_quote(self):
        sw = _create_default()
        for i in range(self._PRIME_COUNT):
            sw.update(_INPUT[i % len(_INPUT)])
        q = Quote(time=_TIME, bid_price=100.0, ask_price=100.0, bid_size=0.0, ask_size=0.0)
        self._check_output(sw.update_quote(q))

    def test_update_trade(self):
        sw = _create_default()
        for i in range(self._PRIME_COUNT):
            sw.update(_INPUT[i % len(_INPUT)])
        t = Trade(time=_TIME, price=100.0, volume=0.0)
        self._check_output(sw.update_trade(t))


class TestSineWaveBandOrdering(unittest.TestCase):
    def test_band_matches_value_lead(self):
        sw = _create_default()
        for i in range(200):
            sw.update(_INPUT[i % len(_INPUT)])
        s = Scalar(time=_TIME, value=_INPUT[0])
        out = sw.update_scalar(s)
        value_scalar = out[int(SineWaveOutput.VALUE)]
        lead_scalar = out[int(SineWaveOutput.LEAD)]
        band = out[int(SineWaveOutput.BAND)]
        self.assertEqual(band.upper, value_scalar.value)
        self.assertEqual(band.lower, lead_scalar.value)


class TestSineWaveIsPrimed(unittest.TestCase):
    def test_primes_within_sequence(self):
        sw = _create_default()
        self.assertFalse(sw.is_primed())
        primed_at = -1
        for i in range(len(_INPUT)):
            sw.update(_INPUT[i])
            if sw.is_primed() and primed_at < 0:
                primed_at = i
        self.assertGreater(primed_at, 0)
        self.assertTrue(sw.is_primed())


class TestSineWaveMetadata(unittest.TestCase):
    def _check_instance(self, meta, mnemonic):
        mnemonic_lead = mnemonic.replace("sw(", "sw-lead(")
        mnemonic_band = mnemonic.replace("sw(", "sw-band(")
        mnemonic_dcp = mnemonic.replace("sw(", "dcp(")
        mnemonic_dcph = mnemonic.replace("sw(", "dcph(")

        self.assertEqual(Identifier.SINE_WAVE, meta.identifier)
        self.assertEqual(mnemonic, meta.mnemonic)
        self.assertEqual(f"Sine wave {mnemonic}", meta.description)
        self.assertEqual(5, len(meta.outputs))

        self.assertEqual(mnemonic, meta.outputs[0].mnemonic)
        self.assertEqual(f"Sine wave {mnemonic}", meta.outputs[0].description)
        self.assertEqual(mnemonic_lead, meta.outputs[1].mnemonic)
        self.assertEqual(f"Sine wave lead {mnemonic_lead}", meta.outputs[1].description)
        self.assertEqual(mnemonic_band, meta.outputs[2].mnemonic)
        self.assertEqual(f"Sine wave band {mnemonic_band}", meta.outputs[2].description)
        self.assertEqual(mnemonic_dcp, meta.outputs[3].mnemonic)
        self.assertEqual(f"Dominant cycle period {mnemonic_dcp}", meta.outputs[3].description)
        self.assertEqual(mnemonic_dcph, meta.outputs[4].mnemonic)
        self.assertEqual(f"Dominant cycle phase {mnemonic_dcph}", meta.outputs[4].description)

    def test_default(self):
        sw = _create_default()
        self._check_instance(sw.metadata(), "sw(0.330, hl/2)")

    def test_phase_accumulator(self):
        sw = _create_alpha(0.5, CycleEstimatorType.PHASE_ACCUMULATOR)
        self._check_instance(sw.metadata(), "sw(0.500, pa(4, 0.200, 0.200), hl/2)")


class TestSineWaveConstruction(unittest.TestCase):
    def test_default(self):
        sw = _create_default()
        self.assertFalse(sw.is_primed())
        self.assertEqual("sw(0.330, hl/2)", sw._mnemonic)

    def test_custom_components(self):
        params = SineWaveParams(
            alpha_ema_period_additional=0.5,
            estimator_type=CycleEstimatorType.HOMODYNE_DISCRIMINATOR,
            estimator_params=_create_cycle_estimator_params(),
            bar_component=BarComponent.CLOSE,
            quote_component=QuoteComponent.MID,
            trade_component=TradeComponent.PRICE,
        )
        sw = SineWave.create(params)
        self.assertEqual("sw(0.500)", sw._mnemonic)

    def test_phase_accumulator(self):
        params = SineWaveParams(
            alpha_ema_period_additional=0.5,
            estimator_type=CycleEstimatorType.PHASE_ACCUMULATOR,
            estimator_params=_create_cycle_estimator_params(),
            bar_component=BarComponent.CLOSE,
            quote_component=QuoteComponent.MID,
            trade_component=TradeComponent.PRICE,
        )
        sw = SineWave.create(params)
        self.assertEqual("sw(0.500, pa(4, 0.200, 0.200))", sw._mnemonic)

    def test_invalid_alpha_zero(self):
        params = SineWaveParams(alpha_ema_period_additional=0.0,
                                estimator_params=_create_cycle_estimator_params())
        with self.assertRaises(ValueError):
            SineWave.create(params)

    def test_invalid_alpha_gt_one(self):
        params = SineWaveParams(alpha_ema_period_additional=1.00000001,
                                estimator_params=_create_cycle_estimator_params())
        with self.assertRaises(ValueError):
            SineWave.create(params)


if __name__ == '__main__':
    unittest.main()
