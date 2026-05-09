"""Tests for the CoronaTrendVigor indicator."""

import math
import unittest
import datetime

from .corona_trend_vigor import CoronaTrendVigor
from .params import Params
from ....entities.bar import Bar
from ....entities.quote import Quote
from ....entities.trade import Trade
from ....entities.scalar import Scalar

from .test_testdata import _test_input, _TOLERANCE


class TestCoronaTrendVigorUpdate(unittest.TestCase):
    """Snapshot tests matching Go's coronatrendvigor_test.go."""

    def test_snapshots(self):
        inp = _test_input()
        t0 = datetime.datetime(2021, 4, 1)

        # (index, tv, vmin, vmax)
        snapshots = [
            (11, 5.6512200755, 20.0000000000, 20.0000000000),
            (12, 6.8379492897, 20.0000000000, 20.0000000000),
            (50, 2.6145116709, 2.3773561485, 20.0000000000),
            (100, 2.7536803664, 2.4892742850, 20.0000000000),
            (150, -6.4606404251, 20.0000000000, 20.0000000000),
            (200, -10.0000000000, 20.0000000000, 20.0000000000),
            (251, -0.1894989954, 0.5847573715, 20.0000000000),
        ]

        x = CoronaTrendVigor(Params())
        si = 0

        for i, v in enumerate(inp):
            t = t0 + datetime.timedelta(minutes=i)
            h, tv = x.update(v, t)

            self.assertIsNotNone(h)
            self.assertEqual(h.parameter_first, -10)
            self.assertEqual(h.parameter_last, 10)
            self.assertAlmostEqual(h.parameter_resolution, 2.45, places=9)

            if not x.is_primed():
                self.assertTrue(h.is_empty())
                self.assertTrue(math.isnan(tv))
                continue

            self.assertEqual(len(h.values), 50)

            if si < len(snapshots) and snapshots[si][0] == i:
                self.assertAlmostEqual(tv, snapshots[si][1], delta=_TOLERANCE)
                self.assertAlmostEqual(h.value_min, snapshots[si][2], delta=_TOLERANCE)
                self.assertAlmostEqual(h.value_max, snapshots[si][3], delta=_TOLERANCE)
                si += 1

        self.assertEqual(si, len(snapshots))


class TestCoronaTrendVigorPriming(unittest.TestCase):

    def test_primes_at_bar_11(self):
        x = CoronaTrendVigor(Params())
        self.assertFalse(x.is_primed())
        inp = _test_input()
        t0 = datetime.datetime(2021, 4, 1)
        primed_at = -1
        for i, v in enumerate(inp):
            x.update(v, t0 + datetime.timedelta(minutes=i))
            if x.is_primed() and primed_at < 0:
                primed_at = i
        self.assertEqual(primed_at, 11)


class TestCoronaTrendVigorNaN(unittest.TestCase):

    def test_nan_input(self):
        x = CoronaTrendVigor(Params())
        t = datetime.datetime(2021, 4, 1)
        h, tv = x.update(float('nan'), t)
        self.assertTrue(h.is_empty())
        self.assertTrue(math.isnan(tv))
        self.assertFalse(x.is_primed())


class TestCoronaTrendVigorMetadata(unittest.TestCase):

    def test_default_metadata(self):
        x = CoronaTrendVigor(Params())
        md = x.metadata()
        self.assertEqual(md.mnemonic, "ctv(50, 20, -10, 10, 30, hl/2)")
        self.assertEqual(md.description, "Corona trend vigor ctv(50, 20, -10, 10, 30, hl/2)")
        self.assertEqual(len(md.outputs), 2)
        self.assertEqual(md.outputs[0].mnemonic, "ctv(50, 20, -10, 10, 30, hl/2)")
        self.assertEqual(md.outputs[1].mnemonic, "ctv-tv(30, hl/2)")


class TestCoronaTrendVigorEntity(unittest.TestCase):

    def _prime(self, x: CoronaTrendVigor):
        inp = _test_input()
        t = datetime.datetime(2021, 4, 1)
        for i in range(50):
            x.update(inp[i % len(inp)], t)

    def test_update_bar(self):
        t = datetime.datetime(2021, 4, 1)
        x = CoronaTrendVigor(Params())
        self._prime(x)
        b = Bar(t, 0, 100 * 1.005, 100 * 0.995, 100, 0)
        out = x.update_bar(b)
        self.assertEqual(len(out), 2)
        self.assertEqual(out[1].time, t)

    def test_update_scalar(self):
        t = datetime.datetime(2021, 4, 1)
        x = CoronaTrendVigor(Params())
        self._prime(x)
        s = Scalar(t, 100.0)
        out = x.update_scalar(s)
        self.assertEqual(len(out), 2)
        self.assertEqual(out[1].time, t)

    def test_update_quote(self):
        t = datetime.datetime(2021, 4, 1)
        x = CoronaTrendVigor(Params())
        self._prime(x)
        q = Quote(t, 100.0, 100.0, 1.0, 1.0)
        out = x.update_quote(q)
        self.assertEqual(len(out), 2)
        self.assertEqual(out[1].time, t)

    def test_update_trade(self):
        t = datetime.datetime(2021, 4, 1)
        x = CoronaTrendVigor(Params())
        self._prime(x)
        r = Trade(t, 100.0, 1.0)
        out = x.update_trade(r)
        self.assertEqual(len(out), 2)
        self.assertEqual(out[1].time, t)


class TestCoronaTrendVigorValidation(unittest.TestCase):

    def test_raster_length_lt_2(self):
        with self.assertRaises(ValueError):
            CoronaTrendVigor(Params(raster_length=1))

    def test_max_param_le_min(self):
        with self.assertRaises(ValueError):
            CoronaTrendVigor(Params(min_parameter_value=5, max_parameter_value=5))

    def test_hp_cutoff_lt_2(self):
        with self.assertRaises(ValueError):
            CoronaTrendVigor(Params(high_pass_filter_cutoff=1))

    def test_minimal_period_lt_2(self):
        with self.assertRaises(ValueError):
            CoronaTrendVigor(Params(minimal_period=1))

    def test_maximal_period_le_minimal(self):
        with self.assertRaises(ValueError):
            CoronaTrendVigor(Params(minimal_period=10, maximal_period=10))


if __name__ == '__main__':
    unittest.main()
