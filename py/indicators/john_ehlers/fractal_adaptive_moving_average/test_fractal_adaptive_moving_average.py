"""Tests for FractalAdaptiveMovingAverage indicator."""

import math
import unittest

from py.indicators.john_ehlers.fractal_adaptive_moving_average.fractal_adaptive_moving_average import FractalAdaptiveMovingAverage
from py.indicators.john_ehlers.fractal_adaptive_moving_average.params import FractalAdaptiveMovingAverageParams
from py.indicators.john_ehlers.fractal_adaptive_moving_average.output import FractalAdaptiveMovingAverageOutput
from py.indicators.core.identifier import Identifier
from py.entities.scalar import Scalar
from py.entities.bar import Bar
from py.entities.quote import Quote
from py.entities.trade import Trade

from .test_testdata import (
    _time,
    _input_mid,
    _input_high,
    _input_low,
    _expected_frama,
    _expected_fdim,
)


def _create(length, alpha):
    params = FractalAdaptiveMovingAverageParams(length=length, slowest_smoothing_factor=alpha)
    return FractalAdaptiveMovingAverage.create(params)


class TestFractalAdaptiveMovingAverageUpdate(unittest.TestCase):

    def test_reference_frama(self):
        """Reference implementation: FRAMA from test_frama.xls."""
        mid = _input_mid()
        high = _input_high()
        low = _input_low()
        exp = _expected_frama()
        frama = _create(16, 0.01)

        for i in range(15):
            v = frama.update(mid[i], high[i], low[i])
            self.assertTrue(math.isnan(v), f"[{i}] expected NaN, got {v}")

        for i in range(15, len(mid)):
            v = frama.update(mid[i], high[i], low[i])
            self.assertAlmostEqual(exp[i], v, delta=1e-9, msg=f"[{i}]")

        self.assertTrue(math.isnan(frama.update(math.nan, math.nan, math.nan)))

    def test_reference_fdim(self):
        """Reference implementation: Fdim from test_frama.xls."""
        mid = _input_mid()
        high = _input_high()
        low = _input_low()
        exp = _expected_fdim()
        frama = _create(16, 0.01)

        for i in range(15):
            frama.update(mid[i], high[i], low[i])
            self.assertTrue(math.isnan(frama.fractal_dimension), f"[{i}] expected NaN")

        for i in range(15, len(mid)):
            frama.update(mid[i], high[i], low[i])
            self.assertAlmostEqual(exp[i], frama.fractal_dimension, delta=1e-9, msg=f"[{i}]")


class TestFractalAdaptiveMovingAverageUpdateEntity(unittest.TestCase):

    def _prime(self, frama):
        for _ in range(15):
            frama.update(0.0, 0.0, 0.0)

    def test_update_scalar(self):
        frama = _create(16, 0.01)
        self._prime(frama)
        s = Scalar(time=_time(), value=3.0)
        out = frama.update_scalar(s)
        self.assertEqual(2, len(out))
        self.assertAlmostEqual(2.999999999999997, out[0].value, delta=1e-14)
        self.assertAlmostEqual(1.0000000000000002, out[1].value, delta=1e-14)

    def test_update_bar(self):
        frama = _create(16, 0.01)
        self._prime(frama)
        b = Bar(time=_time(), open=0, high=3.0, low=3.0, close=3.0, volume=0)
        out = frama.update_bar(b)
        self.assertEqual(2, len(out))
        self.assertAlmostEqual(2.999999999999997, out[0].value, delta=1e-14)
        self.assertAlmostEqual(1.0000000000000002, out[1].value, delta=1e-14)

    def test_update_quote(self):
        frama = _create(16, 0.01)
        self._prime(frama)
        q = Quote(time=_time(), bid_price=3.0, ask_price=3.0, bid_size=0, ask_size=0)
        out = frama.update_quote(q)
        self.assertEqual(2, len(out))
        self.assertAlmostEqual(2.999999999999997, out[0].value, delta=1e-14)
        self.assertAlmostEqual(1.0000000000000002, out[1].value, delta=1e-14)

    def test_update_trade(self):
        frama = _create(16, 0.01)
        self._prime(frama)
        t = Trade(time=_time(), price=3.0, volume=0)
        out = frama.update_trade(t)
        self.assertEqual(2, len(out))
        self.assertAlmostEqual(2.999999999999997, out[0].value, delta=1e-14)
        self.assertAlmostEqual(1.0000000000000002, out[1].value, delta=1e-14)


class TestFractalAdaptiveMovingAverageIsPrimed(unittest.TestCase):

    def test_primed_at_length(self):
        mid = _input_mid()
        high = _input_high()
        low = _input_low()
        frama = _create(16, 0.01)

        self.assertFalse(frama.is_primed())

        for i in range(15):
            frama.update(mid[i], high[i], low[i])
            self.assertFalse(frama.is_primed(), f"[{i+1}]")

        for i in range(15, len(mid)):
            frama.update(mid[i], high[i], low[i])
            self.assertTrue(frama.is_primed(), f"[{i+1}]")


class TestFractalAdaptiveMovingAverageMetadata(unittest.TestCase):

    def test_default_components(self):
        frama = _create(16, 0.01)
        m = frama.metadata()

        self.assertEqual(Identifier.FRACTAL_ADAPTIVE_MOVING_AVERAGE, m.identifier)
        self.assertEqual("frama(16, 0.010)", m.mnemonic)
        self.assertEqual("Fractal adaptive moving average frama(16, 0.010)", m.description)
        self.assertEqual(2, len(m.outputs))
        self.assertEqual("frama(16, 0.010)", m.outputs[0].mnemonic)
        self.assertEqual("framaDim(16, 0.010)", m.outputs[1].mnemonic)

    def test_non_default_bar_component(self):
        from py.entities.bar_component import BarComponent
        params = FractalAdaptiveMovingAverageParams(
            length=16, slowest_smoothing_factor=0.01,
            bar_component=BarComponent.MEDIAN)
        frama = FractalAdaptiveMovingAverage.create(params)
        m = frama.metadata()

        self.assertEqual("frama(16, 0.010, hl/2)", m.mnemonic)
        self.assertEqual("Fractal adaptive moving average frama(16, 0.010, hl/2)", m.description)


class TestNewFractalAdaptiveMovingAverage(unittest.TestCase):

    def test_length_too_small(self):
        for length in [1, 0, -1]:
            with self.assertRaises(ValueError):
                _create(length, 0.01)

    def test_alpha_out_of_range(self):
        with self.assertRaises(ValueError):
            _create(16, -0.01)
        with self.assertRaises(ValueError):
            _create(16, 1.01)

    def test_odd_length_rounds_up(self):
        frama = _create(17, 0.01)
        self.assertEqual("frama(18, 0.010)", frama.metadata().mnemonic)


if __name__ == '__main__':
    unittest.main()
