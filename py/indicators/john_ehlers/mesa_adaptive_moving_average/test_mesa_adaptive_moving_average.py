"""Tests for the Mesa Adaptive Moving Average indicator."""

import math
import unittest

from py.indicators.john_ehlers.mesa_adaptive_moving_average.mesa_adaptive_moving_average import MesaAdaptiveMovingAverage
from py.indicators.john_ehlers.mesa_adaptive_moving_average.params import (
    MesaAdaptiveMovingAverageLengthParams, MesaAdaptiveMovingAverageSmoothingFactorParams)
from py.indicators.john_ehlers.mesa_adaptive_moving_average.output import MesaAdaptiveMovingAverageOutput
from py.indicators.core.identifier import Identifier
from py.indicators.core.outputs.band import Band
from py.entities.scalar import Scalar
from py.entities.bar import Bar
from py.entities.quote import Quote
from py.entities.trade import Trade
from py.entities.bar_component import BarComponent
from py.entities.quote_component import QuoteComponent
from py.entities.trade_component import TradeComponent
from py.indicators.john_ehlers.hilbert_transformer.cycle_estimator_type import CycleEstimatorType
from py.indicators.john_ehlers.hilbert_transformer.cycle_estimator_params import CycleEstimatorParams

from .test_testdata import (
    _TIME,
    _INPUT,
    _EXPECTED_MAMA,
    _EXPECTED_FAMA,
)


class TestMesaAdaptiveMovingAverage(unittest.TestCase):

    def _create_length(self, fast=3, slow=39):
        params = MesaAdaptiveMovingAverageLengthParams(
            fast_limit_length=fast, slow_limit_length=slow,
            estimator_type=CycleEstimatorType.HOMODYNE_DISCRIMINATOR,
            estimator_params=CycleEstimatorParams(
                smoothing_length=4, alpha_ema_quadrature_in_phase=0.2,
                alpha_ema_period=0.2, warm_up_period=0),
        )
        return MesaAdaptiveMovingAverage.from_length(params)

    def _create_alpha(self, fast, slow):
        params = MesaAdaptiveMovingAverageSmoothingFactorParams(
            fast_limit_smoothing_factor=fast, slow_limit_smoothing_factor=slow,
            estimator_type=CycleEstimatorType.HOMODYNE_DISCRIMINATOR,
            estimator_params=CycleEstimatorParams(
                smoothing_length=4, alpha_ema_quadrature_in_phase=0.2,
                alpha_ema_period=0.2, warm_up_period=0),
        )
        return MesaAdaptiveMovingAverage.from_smoothing_factor(params)

    def test_mama_reference(self):
        """Reference implementation: MAMA from test_mama_new.xls."""
        mama = self._create_length()
        lprimed = 26

        for i in range(lprimed):
            act = mama.update(_INPUT[i])
            self.assertTrue(math.isnan(act), f"[{i}] expected NaN, got {act}")

        for i in range(lprimed, len(_INPUT)):
            act = mama.update(_INPUT[i])
            self.assertAlmostEqual(_EXPECTED_MAMA[i], act, delta=1e-9,
                                   msg=f"[{i}] MAMA mismatch")

        self.assertTrue(math.isnan(mama.update(math.nan)))

    def test_fama_reference(self):
        """Reference implementation: FAMA from test_mama_new.xls."""
        mama = self._create_length()
        lprimed = 26

        for i in range(lprimed):
            mama.update(_INPUT[i])

        for i in range(lprimed, len(_INPUT)):
            mama.update(_INPUT[i])
            self.assertAlmostEqual(_EXPECTED_FAMA[i], mama._fama, delta=1e-9,
                                   msg=f"[{i}] FAMA mismatch")

    def test_is_primed(self):
        """Primed at index 26."""
        mama = self._create_length()
        lprimed = 26

        self.assertFalse(mama.is_primed())

        for i in range(lprimed):
            mama.update(_INPUT[i])
            self.assertFalse(mama.is_primed(), f"[{i}] should not be primed")

        for i in range(lprimed, len(_INPUT)):
            mama.update(_INPUT[i])
            self.assertTrue(mama.is_primed(), f"[{i}] should be primed")

    def test_update_scalar(self):
        """Entity update: scalar."""
        mama = self._create_length()
        for _ in range(26):
            mama.update(0.0)

        s = Scalar(time=_TIME, value=3.0)
        out = mama.update_scalar(s)
        self.assertEqual(3, len(out))
        self.assertAlmostEqual(1.5, out[0].value, delta=1e-9)
        self.assertAlmostEqual(0.375, out[1].value, delta=1e-9)
        self.assertIsInstance(out[2], Band)
        self.assertAlmostEqual(1.5, out[2].upper, delta=1e-9)
        self.assertAlmostEqual(0.375, out[2].lower, delta=1e-9)

    def test_update_bar(self):
        """Entity update: bar (close component by default)."""
        mama = self._create_length()
        for _ in range(26):
            mama.update(0.0)

        b = Bar(time=_TIME, open=0, high=0, low=0, close=3.0, volume=0)
        out = mama.update_bar(b)
        self.assertEqual(3, len(out))
        self.assertAlmostEqual(1.5, out[0].value, delta=1e-9)
        self.assertAlmostEqual(0.375, out[1].value, delta=1e-9)

    def test_update_quote(self):
        """Entity update: quote (mid price by default)."""
        mama = self._create_length()
        for _ in range(26):
            mama.update(0.0)

        q = Quote(time=_TIME, bid_price=3.0, ask_price=3.0, bid_size=0, ask_size=0)
        out = mama.update_quote(q)
        self.assertEqual(3, len(out))
        self.assertAlmostEqual(1.5, out[0].value, delta=1e-9)
        self.assertAlmostEqual(0.375, out[1].value, delta=1e-9)

    def test_update_trade(self):
        """Entity update: trade."""
        mama = self._create_length()
        for _ in range(26):
            mama.update(0.0)

        t = Trade(time=_TIME, price=3.0, volume=0)
        out = mama.update_trade(t)
        self.assertEqual(3, len(out))
        self.assertAlmostEqual(1.5, out[0].value, delta=1e-9)
        self.assertAlmostEqual(0.375, out[1].value, delta=1e-9)

    def test_metadata_length(self):
        """Metadata for length-based construction."""
        mama = self._create_length(fast=2, slow=40)
        m = mama.metadata()
        self.assertEqual(Identifier.MESA_ADAPTIVE_MOVING_AVERAGE, m.identifier)
        self.assertEqual("mama(2, 40)", m.mnemonic)
        self.assertEqual("Mesa adaptive moving average mama(2, 40)", m.description)
        self.assertEqual(3, len(m.outputs))
        self.assertEqual("mama(2, 40)", m.outputs[0].mnemonic)
        self.assertEqual("fama(2, 40)", m.outputs[1].mnemonic)
        self.assertEqual("mama-fama(2, 40)", m.outputs[2].mnemonic)

    def test_metadata_alpha(self):
        """Metadata for smoothing-factor-based construction."""
        mama = self._create_alpha(0.666666666, 0.064516129)
        m = mama.metadata()
        self.assertEqual("mama(0.667, 0.065)", m.mnemonic)

    def test_default_constructor(self):
        """Default constructor."""
        mama = MesaAdaptiveMovingAverage.create_default()
        m = mama.metadata()
        self.assertEqual("mama(3, 39)", m.mnemonic)

    def test_error_fast_limit_length_1(self):
        with self.assertRaises(ValueError) as ctx:
            MesaAdaptiveMovingAverage.from_length(
                MesaAdaptiveMovingAverageLengthParams(fast_limit_length=1, slow_limit_length=39))
        self.assertIn("fast limit length should be larger than 1", str(ctx.exception))

    def test_error_fast_limit_length_0(self):
        with self.assertRaises(ValueError):
            MesaAdaptiveMovingAverage.from_length(
                MesaAdaptiveMovingAverageLengthParams(fast_limit_length=0, slow_limit_length=39))

    def test_error_slow_limit_length_1(self):
        with self.assertRaises(ValueError) as ctx:
            MesaAdaptiveMovingAverage.from_length(
                MesaAdaptiveMovingAverageLengthParams(fast_limit_length=3, slow_limit_length=1))
        self.assertIn("slow limit length should be larger than 1", str(ctx.exception))

    def test_error_slow_limit_length_0(self):
        with self.assertRaises(ValueError):
            MesaAdaptiveMovingAverage.from_length(
                MesaAdaptiveMovingAverageLengthParams(fast_limit_length=3, slow_limit_length=0))

    def test_error_fast_sf_negative(self):
        with self.assertRaises(ValueError) as ctx:
            MesaAdaptiveMovingAverage.from_smoothing_factor(
                MesaAdaptiveMovingAverageSmoothingFactorParams(
                    fast_limit_smoothing_factor=-0.00000001, slow_limit_smoothing_factor=0.33))
        self.assertIn("fast limit smoothing factor should be in range [0, 1]", str(ctx.exception))

    def test_error_fast_sf_over_1(self):
        with self.assertRaises(ValueError):
            MesaAdaptiveMovingAverage.from_smoothing_factor(
                MesaAdaptiveMovingAverageSmoothingFactorParams(
                    fast_limit_smoothing_factor=1.00000001, slow_limit_smoothing_factor=0.33))

    def test_error_slow_sf_negative(self):
        with self.assertRaises(ValueError) as ctx:
            MesaAdaptiveMovingAverage.from_smoothing_factor(
                MesaAdaptiveMovingAverageSmoothingFactorParams(
                    fast_limit_smoothing_factor=0.66, slow_limit_smoothing_factor=-0.00000001))
        self.assertIn("slow limit smoothing factor should be in range [0, 1]", str(ctx.exception))

    def test_error_slow_sf_over_1(self):
        with self.assertRaises(ValueError):
            MesaAdaptiveMovingAverage.from_smoothing_factor(
                MesaAdaptiveMovingAverageSmoothingFactorParams(
                    fast_limit_smoothing_factor=0.66, slow_limit_smoothing_factor=1.00000001))

    def test_metadata_with_estimator_moniker(self):
        """Non-default estimator shows moniker."""
        params = MesaAdaptiveMovingAverageLengthParams(
            fast_limit_length=2, slow_limit_length=40,
            estimator_type=CycleEstimatorType.HOMODYNE_DISCRIMINATOR,
            estimator_params=CycleEstimatorParams(
                smoothing_length=3, alpha_ema_quadrature_in_phase=0.2,
                alpha_ema_period=0.2, warm_up_period=0),
            bar_component=BarComponent.MEDIAN,
            quote_component=QuoteComponent.MID,
            trade_component=TradeComponent.PRICE,
        )
        mama = MesaAdaptiveMovingAverage.from_length(params)
        m = mama.metadata()
        self.assertEqual("mama(2, 40, hd(3, 0.200, 0.200), hl/2)", m.mnemonic)

    def test_nan_output_when_not_primed(self):
        """Before priming, fama in entity output is NaN."""
        mama = self._create_length()
        out = mama.update_scalar(Scalar(time=_TIME, value=50.0))
        self.assertTrue(math.isnan(out[0].value))
        self.assertTrue(math.isnan(out[1].value))
        self.assertTrue(math.isnan(out[2].upper))
        self.assertTrue(math.isnan(out[2].lower))


if __name__ == '__main__':
    unittest.main()
