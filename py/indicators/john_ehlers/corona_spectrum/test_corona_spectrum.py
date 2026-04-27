"""Tests for the CoronaSpectrum indicator."""

import math
import unittest
import datetime

from .corona_spectrum import CoronaSpectrum
from .params import Params


def _test_input() -> list[float]:
    """252-entry TA-Lib MAMA reference series."""
    return [
        92.0000, 93.1725, 95.3125, 94.8450, 94.4075, 94.1100, 93.5000, 91.7350, 90.9550, 91.6875,
        94.5000, 97.9700, 97.5775, 90.7825, 89.0325, 92.0950, 91.1550, 89.7175, 90.6100, 91.0000,
        88.9225, 87.5150, 86.4375, 83.8900, 83.0025, 82.8125, 82.8450, 86.7350, 86.8600, 87.5475,
        85.7800, 86.1725, 86.4375, 87.2500, 88.9375, 88.2050, 85.8125, 84.5950, 83.6575, 84.4550,
        83.5000, 86.7825, 88.1725, 89.2650, 90.8600, 90.7825, 91.8600, 90.3600, 89.8600, 90.9225,
        89.5000, 87.6725, 86.5000, 84.2825, 82.9075, 84.2500, 85.6875, 86.6100, 88.2825, 89.5325,
        89.5000, 88.0950, 90.6250, 92.2350, 91.6725, 92.5925, 93.0150, 91.1725, 90.9850, 90.3775,
        88.2500, 86.9075, 84.0925, 83.1875, 84.2525, 97.8600, 99.8750, 103.2650, 105.9375, 103.5000,
        103.1100, 103.6100, 104.6400, 106.8150, 104.9525, 105.5000, 107.1400, 109.7350, 109.8450, 110.9850,
        120.0000, 119.8750, 117.9075, 119.4075, 117.9525, 117.2200, 115.6425, 113.1100, 111.7500, 114.5175,
        114.7450, 115.4700, 112.5300, 112.0300, 113.4350, 114.2200, 119.5950, 117.9650, 118.7150, 115.0300,
        114.5300, 115.0000, 116.5300, 120.1850, 120.5000, 120.5950, 124.1850, 125.3750, 122.9700, 123.0000,
        124.4350, 123.4400, 124.0300, 128.1850, 129.6550, 130.8750, 132.3450, 132.0650, 133.8150, 135.6600,
        137.0350, 137.4700, 137.3450, 136.3150, 136.4400, 136.2850, 129.0950, 128.3100, 126.0000, 124.0300,
        123.9350, 125.0300, 127.2500, 125.6200, 125.5300, 123.9050, 120.6550, 119.9650, 120.7800, 124.0000,
        122.7800, 120.7200, 121.7800, 122.4050, 123.2500, 126.1850, 127.5600, 126.5650, 123.0600, 122.7150,
        123.5900, 122.3100, 122.4650, 123.9650, 123.9700, 124.1550, 124.4350, 127.0000, 125.5000, 128.8750,
        130.5350, 132.3150, 134.0650, 136.0350, 133.7800, 132.7500, 133.4700, 130.9700, 127.5950, 128.4400,
        127.9400, 125.8100, 124.6250, 122.7200, 124.0900, 123.2200, 121.4050, 120.9350, 118.2800, 118.3750,
        121.1550, 120.9050, 117.1250, 113.0600, 114.9050, 112.4350, 107.9350, 105.9700, 106.3700, 106.8450,
        106.9700, 110.0300, 91.0000, 93.5600, 93.6200, 95.3100, 94.1850, 94.7800, 97.6250, 97.5900,
        95.2500, 94.7200, 92.2200, 91.5650, 92.2200, 93.8100, 95.5900, 96.1850, 94.6250, 95.1200,
        94.0000, 93.7450, 95.9050, 101.7450, 106.4400, 107.9350, 103.4050, 105.0600, 104.1550, 103.3100,
        103.3450, 104.8400, 110.4050, 114.5000, 117.3150, 118.2500, 117.1850, 109.7500, 109.6550, 108.5300,
        106.2200, 107.7200, 109.8400, 109.0950, 109.0900, 109.1550, 109.3150, 109.0600, 109.9050, 109.6250,
        109.5300, 108.0600,
    ]


_TOLERANCE = 1e-4


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
