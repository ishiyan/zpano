"""Tests for the CoronaSwingPosition indicator."""

import math
import unittest
import datetime

from .corona_swing_position import CoronaSwingPosition
from .params import Params
from ....entities.bar import Bar
from ....entities.quote import Quote
from ....entities.trade import Trade
from ....entities.scalar import Scalar

from .test_testdata import _test_input, _TOLERANCE


class TestCoronaSwingPositionUpdate(unittest.TestCase):
    """Snapshot tests matching Go's coronaswingposition_test.go."""

    def test_snapshots(self):
        inp = _test_input()
        t0 = datetime.datetime(2021, 4, 1)

        # (index, sp, vmin, vmax)
        snapshots = [
            (11, 5.0000000000, 20.0000000000, 20.0000000000),
            (12, 5.0000000000, 20.0000000000, 20.0000000000),
            (50, 4.5384908349, 20.0000000000, 20.0000000000),
            (100, -3.8183742675, 3.4957777081, 20.0000000000),
            (150, -1.8516194371, 5.3792287864, 20.0000000000),
            (200, -3.6944428668, 4.2580825738, 20.0000000000),
            (251, -0.8524812061, 4.4822539784, 20.0000000000),
        ]

        x = CoronaSwingPosition(Params())
        si = 0

        for i, v in enumerate(inp):
            t = t0 + datetime.timedelta(minutes=i)
            h, sp = x.update(v, t)

            self.assertIsNotNone(h)
            self.assertEqual(h.parameter_first, -5)
            self.assertEqual(h.parameter_last, 5)
            self.assertAlmostEqual(h.parameter_resolution, 4.9, places=9)

            if not x.is_primed():
                self.assertTrue(h.is_empty())
                self.assertTrue(math.isnan(sp))
                continue

            self.assertEqual(len(h.values), 50)

            if si < len(snapshots) and snapshots[si][0] == i:
                self.assertAlmostEqual(sp, snapshots[si][1], delta=_TOLERANCE)
                self.assertAlmostEqual(h.value_min, snapshots[si][2], delta=_TOLERANCE)
                self.assertAlmostEqual(h.value_max, snapshots[si][3], delta=_TOLERANCE)
                si += 1

        self.assertEqual(si, len(snapshots))


class TestCoronaSwingPositionPriming(unittest.TestCase):

    def test_primes_at_bar_11(self):
        x = CoronaSwingPosition(Params())
        self.assertFalse(x.is_primed())
        inp = _test_input()
        t0 = datetime.datetime(2021, 4, 1)
        primed_at = -1
        for i, v in enumerate(inp):
            x.update(v, t0 + datetime.timedelta(minutes=i))
            if x.is_primed() and primed_at < 0:
                primed_at = i
        self.assertEqual(primed_at, 11)


class TestCoronaSwingPositionNaN(unittest.TestCase):

    def test_nan_input(self):
        x = CoronaSwingPosition(Params())
        t = datetime.datetime(2021, 4, 1)
        h, sp = x.update(float('nan'), t)
        self.assertTrue(h.is_empty())
        self.assertTrue(math.isnan(sp))
        self.assertFalse(x.is_primed())


class TestCoronaSwingPositionMetadata(unittest.TestCase):

    def test_default_metadata(self):
        x = CoronaSwingPosition(Params())
        md = x.metadata()
        self.assertEqual(md.mnemonic, "cswing(50, 20, -5, 5, 30, hl/2)")
        self.assertEqual(md.description, "Corona swing position cswing(50, 20, -5, 5, 30, hl/2)")
        self.assertEqual(len(md.outputs), 2)
        self.assertEqual(md.outputs[0].mnemonic, "cswing(50, 20, -5, 5, 30, hl/2)")
        self.assertEqual(md.outputs[1].mnemonic, "cswing-sp(30, hl/2)")


class TestCoronaSwingPositionEntity(unittest.TestCase):

    def _prime(self, x: CoronaSwingPosition):
        inp = _test_input()
        t = datetime.datetime(2021, 4, 1)
        for i in range(50):
            x.update(inp[i % len(inp)], t)

    def test_update_bar(self):
        t = datetime.datetime(2021, 4, 1)
        x = CoronaSwingPosition(Params())
        self._prime(x)
        b = Bar(t, 0, 100 * 1.005, 100 * 0.995, 100, 0)
        out = x.update_bar(b)
        self.assertEqual(len(out), 2)
        self.assertEqual(out[1].time, t)

    def test_update_scalar(self):
        t = datetime.datetime(2021, 4, 1)
        x = CoronaSwingPosition(Params())
        self._prime(x)
        s = Scalar(t, 100.0)
        out = x.update_scalar(s)
        self.assertEqual(len(out), 2)
        self.assertEqual(out[1].time, t)

    def test_update_quote(self):
        t = datetime.datetime(2021, 4, 1)
        x = CoronaSwingPosition(Params())
        self._prime(x)
        q = Quote(t, 100.0, 100.0, 1.0, 1.0)
        out = x.update_quote(q)
        self.assertEqual(len(out), 2)
        self.assertEqual(out[1].time, t)

    def test_update_trade(self):
        t = datetime.datetime(2021, 4, 1)
        x = CoronaSwingPosition(Params())
        self._prime(x)
        r = Trade(t, 100.0, 1.0)
        out = x.update_trade(r)
        self.assertEqual(len(out), 2)
        self.assertEqual(out[1].time, t)


class TestCoronaSwingPositionValidation(unittest.TestCase):

    def test_raster_length_lt_2(self):
        with self.assertRaises(ValueError):
            CoronaSwingPosition(Params(raster_length=1))

    def test_max_param_le_min(self):
        with self.assertRaises(ValueError):
            CoronaSwingPosition(Params(min_parameter_value=5, max_parameter_value=5))

    def test_hp_cutoff_lt_2(self):
        with self.assertRaises(ValueError):
            CoronaSwingPosition(Params(high_pass_filter_cutoff=1))

    def test_minimal_period_lt_2(self):
        with self.assertRaises(ValueError):
            CoronaSwingPosition(Params(minimal_period=1))

    def test_maximal_period_le_minimal(self):
        with self.assertRaises(ValueError):
            CoronaSwingPosition(Params(minimal_period=10, maximal_period=10))


if __name__ == '__main__':
    unittest.main()
