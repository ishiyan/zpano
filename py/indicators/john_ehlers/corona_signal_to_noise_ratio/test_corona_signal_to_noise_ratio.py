"""Tests for the CoronaSignalToNoiseRatio indicator."""

import math
import unittest
import datetime

from .corona_signal_to_noise_ratio import CoronaSignalToNoiseRatio
from .params import Params
from ....entities.bar import Bar
from ....entities.quote import Quote
from ....entities.trade import Trade
from ....entities.scalar import Scalar

from .test_testdata import _test_input, _TOLERANCE


def _make_hl(i: int, sample: float) -> tuple[float, float]:
    """Synthetic High/Low around sample, matching Go's makeHL."""
    frac = 0.005 + 0.03 * (1 + math.sin(i * 0.37))
    half = sample * frac
    return sample - half, sample + half


class TestCoronaSignalToNoiseRatioUpdate(unittest.TestCase):
    """Snapshot tests matching Go's coronasignaltonoiseratio_test.go."""

    def test_snapshots(self):
        inp = _test_input()
        t0 = datetime.datetime(2021, 4, 1)

        snapshots = [
            (11, 1.0000000000, 0.0000000000, 20.0000000000),
            (12, 1.0000000000, 0.0000000000, 20.0000000000),
            (50, 1.0000000000, 0.0000000000, 20.0000000000),
            (100, 2.9986583538, 4.2011609652, 20.0000000000),
            (150, 1.0000000000, 0.0000000035, 20.0000000000),
            (200, 1.0000000000, 0.0000000000, 20.0000000000),
            (251, 1.0000000000, 0.0000000026, 20.0000000000),
        ]

        x = CoronaSignalToNoiseRatio(Params())
        si = 0

        for i, v in enumerate(inp):
            t = t0 + datetime.timedelta(minutes=i)
            low, high = _make_hl(i, v)
            h, snr = x.update(v, low, high, t)

            self.assertIsNotNone(h)
            self.assertEqual(h.parameter_first, 1)
            self.assertEqual(h.parameter_last, 11)
            self.assertAlmostEqual(h.parameter_resolution, 4.9, places=9)

            if not x.is_primed():
                self.assertTrue(h.is_empty())
                self.assertTrue(math.isnan(snr))
                continue

            self.assertEqual(len(h.values), 50)

            if si < len(snapshots) and snapshots[si][0] == i:
                self.assertAlmostEqual(snr, snapshots[si][1], delta=_TOLERANCE)
                self.assertAlmostEqual(h.value_min, snapshots[si][2], delta=_TOLERANCE)
                self.assertAlmostEqual(h.value_max, snapshots[si][3], delta=_TOLERANCE)
                si += 1

        self.assertEqual(si, len(snapshots))


class TestCoronaSignalToNoiseRatioPriming(unittest.TestCase):

    def test_primes_at_bar_11(self):
        x = CoronaSignalToNoiseRatio(Params())
        self.assertFalse(x.is_primed())
        inp = _test_input()
        t0 = datetime.datetime(2021, 4, 1)
        primed_at = -1
        for i, v in enumerate(inp):
            low, high = _make_hl(i, v)
            x.update(v, low, high, t0 + datetime.timedelta(minutes=i))
            if x.is_primed() and primed_at < 0:
                primed_at = i
        self.assertEqual(primed_at, 11)


class TestCoronaSignalToNoiseRatioNaN(unittest.TestCase):

    def test_nan_input(self):
        x = CoronaSignalToNoiseRatio(Params())
        t = datetime.datetime(2021, 4, 1)
        h, snr = x.update(float('nan'), float('nan'), float('nan'), t)
        self.assertTrue(h.is_empty())
        self.assertTrue(math.isnan(snr))
        self.assertFalse(x.is_primed())


class TestCoronaSignalToNoiseRatioMetadata(unittest.TestCase):

    def test_default_metadata(self):
        x = CoronaSignalToNoiseRatio(Params())
        md = x.metadata()
        self.assertEqual(md.mnemonic, "csnr(50, 20, 1, 11, 30, hl/2)")
        self.assertEqual(md.description, "Corona signal to noise ratio csnr(50, 20, 1, 11, 30, hl/2)")
        self.assertEqual(len(md.outputs), 2)
        self.assertEqual(md.outputs[0].mnemonic, "csnr(50, 20, 1, 11, 30, hl/2)")
        self.assertEqual(md.outputs[1].mnemonic, "csnr-snr(30, hl/2)")


class TestCoronaSignalToNoiseRatioEntity(unittest.TestCase):

    def _prime(self, x: CoronaSignalToNoiseRatio):
        inp = _test_input()
        t = datetime.datetime(2021, 4, 1)
        for i in range(50):
            low, high = _make_hl(i, inp[i % len(inp)])
            x.update(inp[i % len(inp)], low, high, t)

    def test_update_bar(self):
        t = datetime.datetime(2021, 4, 1)
        x = CoronaSignalToNoiseRatio(Params())
        self._prime(x)
        b = Bar(t, 0, 100 * 1.005, 100 * 0.995, 100, 0)
        out = x.update_bar(b)
        self.assertEqual(len(out), 2)
        self.assertEqual(out[1].time, t)

    def test_update_scalar(self):
        t = datetime.datetime(2021, 4, 1)
        x = CoronaSignalToNoiseRatio(Params())
        self._prime(x)
        s = Scalar(t, 100.0)
        out = x.update_scalar(s)
        self.assertEqual(len(out), 2)
        self.assertEqual(out[1].time, t)

    def test_update_quote(self):
        t = datetime.datetime(2021, 4, 1)
        x = CoronaSignalToNoiseRatio(Params())
        self._prime(x)
        q = Quote(t, 100.0, 100.0, 1.0, 1.0)
        out = x.update_quote(q)
        self.assertEqual(len(out), 2)
        self.assertEqual(out[1].time, t)

    def test_update_trade(self):
        t = datetime.datetime(2021, 4, 1)
        x = CoronaSignalToNoiseRatio(Params())
        self._prime(x)
        r = Trade(t, 100.0, 1.0)
        out = x.update_trade(r)
        self.assertEqual(len(out), 2)
        self.assertEqual(out[1].time, t)


class TestCoronaSignalToNoiseRatioValidation(unittest.TestCase):

    def test_raster_length_lt_2(self):
        with self.assertRaises(ValueError):
            CoronaSignalToNoiseRatio(Params(raster_length=1))

    def test_max_param_le_min(self):
        with self.assertRaises(ValueError):
            CoronaSignalToNoiseRatio(Params(min_parameter_value=5, max_parameter_value=5))

    def test_hp_cutoff_lt_2(self):
        with self.assertRaises(ValueError):
            CoronaSignalToNoiseRatio(Params(high_pass_filter_cutoff=1))

    def test_minimal_period_lt_2(self):
        with self.assertRaises(ValueError):
            CoronaSignalToNoiseRatio(Params(minimal_period=1))

    def test_maximal_period_le_minimal(self):
        with self.assertRaises(ValueError):
            CoronaSignalToNoiseRatio(Params(minimal_period=10, maximal_period=10))


if __name__ == '__main__':
    unittest.main()
