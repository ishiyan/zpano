"""Tests for the AutoCorrelation Indicator."""

import math
import unittest
import datetime

from py.indicators.john_ehlers.auto_correlation_indicator.auto_correlation_indicator import AutoCorrelationIndicator
from py.indicators.john_ehlers.auto_correlation_indicator.params import Params
from py.indicators.john_ehlers.auto_correlation_indicator.output import Output
from py.indicators.core.identifier import Identifier
from py.indicators.core.outputs.shape import Shape
from py.entities.scalar import Scalar
from py.entities.bar import Bar
from py.entities.quote import Quote
from py.entities.trade import Trade
from py.entities.bar_component import BarComponent

from .test_testdata import (
    TEST_TIME,
    TEST_INPUT,
    TOLERANCE,
    MIN_MAX_TOL,
    SNAPSHOTS,
)


class TestAutoCorrelationIndicatorUpdate(unittest.TestCase):
    def test_reference_snapshots(self):
        x = AutoCorrelationIndicator(Params())
        si = 0
        for i, v in enumerate(TEST_INPUT):
            t = TEST_TIME + datetime.timedelta(minutes=i)
            h = x.update(v, t)
            self.assertIsNotNone(h)
            self.assertEqual(h.parameter_first, 3)
            self.assertEqual(h.parameter_last, 48)
            self.assertEqual(h.parameter_resolution, 1)

            if not x.is_primed():
                self.assertTrue(h.is_empty())
                continue

            self.assertEqual(len(h.values), 46)

            if si < len(SNAPSHOTS) and SNAPSHOTS[si]['i'] == i:
                snap = SNAPSHOTS[si]
                self.assertAlmostEqual(h.value_min, snap['value_min'], delta=MIN_MAX_TOL)
                self.assertAlmostEqual(h.value_max, snap['value_max'], delta=MIN_MAX_TOL)
                for bin_idx, expected in snap['spots']:
                    self.assertAlmostEqual(h.values[bin_idx], expected, delta=TOLERANCE)
                si += 1

        self.assertEqual(si, len(SNAPSHOTS))

    def test_synthetic_sine(self):
        period = 35.0
        bars = 600
        x = AutoCorrelationIndicator(Params())

        last = None
        for i in range(bars):
            sample = 100 + math.sin(2 * math.pi * i / period)
            t = TEST_TIME + datetime.timedelta(minutes=i)
            last = x.update(sample, t)

        self.assertIsNotNone(last)
        self.assertFalse(last.is_empty())

        peak_bin = 0
        for i in range(len(last.values)):
            if last.values[i] > last.values[peak_bin]:
                peak_bin = i

        expected_bin = int(period - last.parameter_first)
        self.assertEqual(peak_bin, expected_bin)

    def test_nan_input(self):
        x = AutoCorrelationIndicator(Params())
        h = x.update(math.nan, TEST_TIME)
        self.assertTrue(h.is_empty())
        self.assertFalse(x.is_primed())


class TestAutoCorrelationIndicatorMetadata(unittest.TestCase):
    def test_metadata(self):
        x = AutoCorrelationIndicator(Params())
        md = x.metadata()
        mn = "aci(3, 48, 10, hl/2)"
        self.assertEqual(md.identifier, Identifier.AUTO_CORRELATION_INDICATOR)
        self.assertEqual(md.mnemonic, mn)
        self.assertEqual(md.description, "Autocorrelation indicator " + mn)
        self.assertEqual(len(md.outputs), 1)
        self.assertEqual(md.outputs[0].kind, int(Output.VALUE))
        self.assertEqual(md.outputs[0].shape, Shape.HEATMAP)
        self.assertEqual(md.outputs[0].mnemonic, mn)


class TestAutoCorrelationIndicatorMnemonicFlags(unittest.TestCase):
    def test_default(self):
        x = AutoCorrelationIndicator(Params())
        self.assertEqual(x.mnemonic, "aci(3, 48, 10, hl/2)")

    def test_average_override(self):
        x = AutoCorrelationIndicator(Params(averaging_length=5))
        self.assertEqual(x.mnemonic, "aci(3, 48, 10, average=5, hl/2)")

    def test_custom_range(self):
        x = AutoCorrelationIndicator(Params(min_lag=5, max_lag=30, smoothing_period=8))
        self.assertEqual(x.mnemonic, "aci(5, 30, 8, hl/2)")


class TestAutoCorrelationIndicatorValidation(unittest.TestCase):
    def test_min_lag_lt_1(self):
        with self.assertRaises(ValueError):
            AutoCorrelationIndicator(Params(min_lag=-1, max_lag=48, smoothing_period=10))

    def test_max_lag_le_min_lag(self):
        with self.assertRaises(ValueError):
            AutoCorrelationIndicator(Params(min_lag=10, max_lag=10, smoothing_period=10))

    def test_smoothing_lt_2(self):
        with self.assertRaises(ValueError):
            AutoCorrelationIndicator(Params(min_lag=3, max_lag=48, smoothing_period=1))

    def test_averaging_lt_0(self):
        with self.assertRaises(ValueError):
            AutoCorrelationIndicator(Params(averaging_length=-1))

    def test_invalid_bar_component(self):
        with self.assertRaises(Exception):
            AutoCorrelationIndicator(Params(bar_component=BarComponent(9999)))


class TestAutoCorrelationIndicatorUpdateEntity(unittest.TestCase):
    def _prime(self, x):
        for i in range(200):
            x.update(TEST_INPUT[i % len(TEST_INPUT)], TEST_TIME)

    def test_update_scalar(self):
        x = AutoCorrelationIndicator(Params())
        self._prime(x)
        s = Scalar(time=TEST_TIME, value=100.0)
        out = x.update_scalar(s)
        self.assertEqual(len(out), 1)
        self.assertEqual(out[0].time, TEST_TIME)

    def test_update_bar(self):
        x = AutoCorrelationIndicator(Params())
        self._prime(x)
        b = Bar(time=TEST_TIME, open=100.0, high=100.0, low=100.0, close=100.0, volume=0.0)
        out = x.update_bar(b)
        self.assertEqual(len(out), 1)
        self.assertEqual(out[0].time, TEST_TIME)

    def test_update_quote(self):
        x = AutoCorrelationIndicator(Params())
        self._prime(x)
        q = Quote(time=TEST_TIME, bid_price=100.0, ask_price=100.0, bid_size=1.0, ask_size=1.0)
        out = x.update_quote(q)
        self.assertEqual(len(out), 1)
        self.assertEqual(out[0].time, TEST_TIME)

    def test_update_trade(self):
        x = AutoCorrelationIndicator(Params())
        self._prime(x)
        tr = Trade(time=TEST_TIME, price=100.0, volume=1.0)
        out = x.update_trade(tr)
        self.assertEqual(len(out), 1)
        self.assertEqual(out[0].time, TEST_TIME)


if __name__ == '__main__':
    unittest.main()
