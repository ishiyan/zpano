"""Tests for the AutoCorrelation Periodogram indicator."""

import math
import unittest
import datetime

from ....entities.bar import Bar
from ....entities.bar_component import BarComponent
from ....entities.quote import Quote
from ....entities.trade import Trade
from ....entities.scalar import Scalar
from ...core.identifier import Identifier
from .auto_correlation_periodogram import AutoCorrelationPeriodogram
from .params import Params

from .test_testdata import (
    _test_time,
    _test_input,
    _TOLERANCE,
    _MINMAX_TOL,
    _SNAPSHOTS,
)


class TestAutoCorrelationPeriodogramUpdate(unittest.TestCase):
    def test_update(self):
        inp = _test_input()
        t0 = _test_time()
        x = AutoCorrelationPeriodogram(Params())

        si = 0
        for i, val in enumerate(inp):
            t = t0 + datetime.timedelta(minutes=i)
            h = x.update(val, t)

            self.assertIsNotNone(h)
            self.assertEqual(h.parameter_first, 10)
            self.assertEqual(h.parameter_last, 48)
            self.assertEqual(h.parameter_resolution, 1)

            if not x.is_primed():
                self.assertTrue(h.is_empty(), f"[{i}] expected empty before priming")
                continue

            self.assertEqual(len(h.values), 39, f"[{i}] values len")

            if si < len(_SNAPSHOTS) and _SNAPSHOTS[si]['i'] == i:
                snap = _SNAPSHOTS[si]
                self.assertAlmostEqual(h.value_min, snap['value_min'], delta=_MINMAX_TOL,
                                       msg=f"[{i}] ValueMin")
                self.assertAlmostEqual(h.value_max, snap['value_max'], delta=_MINMAX_TOL,
                                       msg=f"[{i}] ValueMax")
                for bin_idx, exp_val in snap['spots']:
                    self.assertAlmostEqual(h.values[bin_idx], exp_val, delta=_TOLERANCE,
                                           msg=f"[{i}] Values[{bin_idx}]")
                si += 1

        self.assertEqual(si, len(_SNAPSHOTS), "did not hit all snapshots")


class TestAutoCorrelationPeriodogramSyntheticSine(unittest.TestCase):
    def test_synthetic_sine(self):
        period = 20.0
        bars = 600
        t0 = _test_time()

        x = AutoCorrelationPeriodogram(Params(
            disable_automatic_gain_control=True,
            fixed_normalization=True,
        ))

        last = None
        for i in range(bars):
            sample = 100 + math.sin(2 * math.pi * i / period)
            last = x.update(sample, t0 + datetime.timedelta(minutes=i))

        self.assertIsNotNone(last)
        self.assertFalse(last.is_empty())

        peak_bin = 0
        for i in range(len(last.values)):
            if last.values[i] > last.values[peak_bin]:
                peak_bin = i

        expected_bin = int(period - last.parameter_first)
        self.assertEqual(peak_bin, expected_bin,
                         f"peak bin: expected {expected_bin} (period {period}), "
                         f"got {peak_bin} (period {last.parameter_first + peak_bin})")


class TestAutoCorrelationPeriodogramNaN(unittest.TestCase):
    def test_nan_input(self):
        x = AutoCorrelationPeriodogram(Params())
        h = x.update(math.nan, _test_time())
        self.assertTrue(h.is_empty())
        self.assertFalse(x.is_primed())


class TestAutoCorrelationPeriodogramMetadata(unittest.TestCase):
    def test_metadata(self):
        x = AutoCorrelationPeriodogram(Params())
        md = x.metadata()
        mn = "acp(10, 48, hl/2)"
        self.assertEqual(md.identifier, Identifier.AUTO_CORRELATION_PERIODOGRAM)
        self.assertEqual(md.mnemonic, mn)
        self.assertEqual(md.description, "Autocorrelation periodogram " + mn)
        self.assertEqual(len(md.outputs), 1)
        self.assertEqual(md.outputs[0].mnemonic, mn)


class TestAutoCorrelationPeriodogramMnemonicFlags(unittest.TestCase):
    def test_default(self):
        x = AutoCorrelationPeriodogram(Params())
        self.assertEqual(x.mnemonic, "acp(10, 48, hl/2)")

    def test_average_override(self):
        x = AutoCorrelationPeriodogram(Params(averaging_length=5))
        self.assertEqual(x.mnemonic, "acp(10, 48, average=5, hl/2)")

    def test_no_sqr(self):
        x = AutoCorrelationPeriodogram(Params(disable_spectral_squaring=True))
        self.assertEqual(x.mnemonic, "acp(10, 48, no-sqr, hl/2)")

    def test_no_smooth(self):
        x = AutoCorrelationPeriodogram(Params(disable_smoothing=True))
        self.assertEqual(x.mnemonic, "acp(10, 48, no-smooth, hl/2)")

    def test_no_agc(self):
        x = AutoCorrelationPeriodogram(Params(disable_automatic_gain_control=True))
        self.assertEqual(x.mnemonic, "acp(10, 48, no-agc, hl/2)")

    def test_agc_override(self):
        x = AutoCorrelationPeriodogram(Params(automatic_gain_control_decay_factor=0.8))
        self.assertEqual(x.mnemonic, "acp(10, 48, agc=0.8, hl/2)")

    def test_no_fn(self):
        x = AutoCorrelationPeriodogram(Params(fixed_normalization=True))
        self.assertEqual(x.mnemonic, "acp(10, 48, no-fn, hl/2)")

    def test_all_flags(self):
        x = AutoCorrelationPeriodogram(Params(
            averaging_length=5,
            disable_spectral_squaring=True,
            disable_smoothing=True,
            disable_automatic_gain_control=True,
            fixed_normalization=True,
        ))
        self.assertEqual(x.mnemonic, "acp(10, 48, average=5, no-sqr, no-smooth, no-agc, no-fn, hl/2)")


class TestAutoCorrelationPeriodogramValidation(unittest.TestCase):
    def test_min_period_lt_2(self):
        with self.assertRaises(ValueError):
            AutoCorrelationPeriodogram(Params(min_period=1, max_period=48, averaging_length=3))

    def test_max_period_le_min(self):
        with self.assertRaises(ValueError):
            AutoCorrelationPeriodogram(Params(min_period=10, max_period=10, averaging_length=3))

    def test_averaging_length_lt_1(self):
        with self.assertRaises(ValueError):
            AutoCorrelationPeriodogram(Params(averaging_length=-1))

    def test_agc_decay_le_0(self):
        with self.assertRaises(ValueError):
            AutoCorrelationPeriodogram(Params(automatic_gain_control_decay_factor=-0.1))

    def test_agc_decay_ge_1(self):
        with self.assertRaises(ValueError):
            AutoCorrelationPeriodogram(Params(automatic_gain_control_decay_factor=1.0))


class TestAutoCorrelationPeriodogramUpdateEntity(unittest.TestCase):
    def _prime(self, x):
        inp = _test_input()
        t = _test_time()
        for i in range(100):
            x.update(inp[i % len(inp)], t)

    def test_update_scalar(self):
        t = _test_time()
        x = AutoCorrelationPeriodogram(Params())
        self._prime(x)
        s = Scalar(time=t, value=100.0)
        out = x.update_scalar(s)
        self.assertEqual(len(out), 1)
        self.assertEqual(out[0].time, t)

    def test_update_bar(self):
        t = _test_time()
        x = AutoCorrelationPeriodogram(Params())
        self._prime(x)
        b = Bar(time=t, open=100, high=100, low=100, close=100, volume=0)
        out = x.update_bar(b)
        self.assertEqual(len(out), 1)
        self.assertEqual(out[0].time, t)

    def test_update_quote(self):
        t = _test_time()
        x = AutoCorrelationPeriodogram(Params())
        self._prime(x)
        q = Quote(time=t, bid_price=100, ask_price=100, bid_size=1, ask_size=1)
        out = x.update_quote(q)
        self.assertEqual(len(out), 1)
        self.assertEqual(out[0].time, t)

    def test_update_trade(self):
        t = _test_time()
        x = AutoCorrelationPeriodogram(Params())
        self._prime(x)
        r = Trade(time=t, price=100, volume=1)
        out = x.update_trade(r)
        self.assertEqual(len(out), 1)
        self.assertEqual(out[0].time, t)


if __name__ == '__main__':
    unittest.main()
