"""Tests for the CoronaSpectrum indicator."""

import math
import unittest
import datetime

from .corona_spectrum import CoronaSpectrum
from .params import Params

from .test_testdata import _test_input, _TOLERANCE


class TestCoronaSpectrumUpdate(unittest.TestCase):
    """Snapshot tests matching Go's coronaspectrum_test.go."""

    def test_snapshots(self):
        inp = _test_input()
        t0 = datetime.datetime(2021, 4, 1)

        snapshots = [
            (11, 17.7604672565, 17.7604672565),
            (12, 6.0000000000, 6.0000000000),
            (50, 15.9989078712, 15.9989078712),
            (100, 14.7455497547, 14.7455497547),
            (150, 17.5000000000, 17.2826036069),
            (200, 19.7557338512, 20.0000000000),
            (251, 6.0000000000, 6.0000000000),
        ]

        x = CoronaSpectrum(Params())
        si = 0

        for i, v in enumerate(inp):
            t = t0 + datetime.timedelta(minutes=i)
            h, dc, dcm = x.update(v, t)

            self.assertIsNotNone(h)
            self.assertEqual(h.parameter_first, 6)
            self.assertEqual(h.parameter_last, 30)
            self.assertEqual(h.parameter_resolution, 2)

            if not x.is_primed():
                self.assertTrue(h.is_empty())
                self.assertTrue(math.isnan(dc))
                self.assertTrue(math.isnan(dcm))
                continue

            self.assertEqual(len(h.values), 49)

            if si < len(snapshots) and snapshots[si][0] == i:
                self.assertAlmostEqual(dc, snapshots[si][1], delta=_TOLERANCE)
                self.assertAlmostEqual(dcm, snapshots[si][2], delta=_TOLERANCE)
                si += 1

        self.assertEqual(si, len(snapshots))


class TestCoronaSpectrumPriming(unittest.TestCase):

    def test_primes_at_bar_11(self):
        x = CoronaSpectrum(Params())
        self.assertFalse(x.is_primed())
        inp = _test_input()
        t0 = datetime.datetime(2021, 4, 1)
        primed_at = -1
        for i, v in enumerate(inp):
            x.update(v, t0 + datetime.timedelta(minutes=i))
            if x.is_primed() and primed_at < 0:
                primed_at = i
        self.assertEqual(primed_at, 11)


class TestCoronaSpectrumNaN(unittest.TestCase):

    def test_nan_input(self):
        x = CoronaSpectrum(Params())
        t = datetime.datetime(2021, 4, 1)
        h, dc, dcm = x.update(float('nan'), t)
        self.assertTrue(h.is_empty())
        self.assertTrue(math.isnan(dc))
        self.assertTrue(math.isnan(dcm))
        self.assertFalse(x.is_primed())


class TestCoronaSpectrumMetadata(unittest.TestCase):

    def test_default_metadata(self):
        x = CoronaSpectrum(Params())
        md = x.metadata()
        self.assertEqual(md.mnemonic, "cspect(6, 20, 6, 30, 30, hl/2)")
        self.assertEqual(md.description, "Corona spectrum cspect(6, 20, 6, 30, 30, hl/2)")
        self.assertEqual(len(md.outputs), 3)
        self.assertEqual(md.outputs[0].mnemonic, "cspect(6, 20, 6, 30, 30, hl/2)")
        self.assertEqual(md.outputs[1].mnemonic, "cspect-dc(30, hl/2)")
        self.assertEqual(md.outputs[2].mnemonic, "cspect-dcm(30, hl/2)")


class TestCoronaSpectrumValidation(unittest.TestCase):

    def test_max_raster_le_min(self):
        with self.assertRaises(ValueError):
            CoronaSpectrum(Params(min_raster_value=10, max_raster_value=10,
                                  min_parameter_value=6, max_parameter_value=30))

    def test_min_param_lt_2(self):
        with self.assertRaises(ValueError):
            CoronaSpectrum(Params(min_parameter_value=1, max_parameter_value=30))

    def test_max_param_le_min(self):
        with self.assertRaises(ValueError):
            CoronaSpectrum(Params(min_parameter_value=20, max_parameter_value=20))

    def test_hp_cutoff_lt_2(self):
        with self.assertRaises(ValueError):
            CoronaSpectrum(Params(high_pass_filter_cutoff=1))


if __name__ == '__main__':
    unittest.main()
