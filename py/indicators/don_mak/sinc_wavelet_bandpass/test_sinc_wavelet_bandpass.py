import math
import unittest
from datetime import datetime

from py.indicators.don_mak.sinc_wavelet_bandpass.sinc_wavelet_bandpass import SincWaveletBandpass
from py.indicators.don_mak.sinc_wavelet_bandpass.params import SincWaveletBandpassParams, Band
from py.indicators.core.identifier import Identifier
from py.entities.scalar import Scalar

from .test_testdata import (
    INPUT_CLOSE,
    EXPECTED_HIGH, EXPECTED_MID, EXPECTED_LOW, EXPECTED_FULL,
    EXPECTED_HIGH_V, EXPECTED_MID_V, EXPECTED_LOW_V, EXPECTED_FULL_V,
    TEST1_INPUT_SINE, TEST1_EXPECTED_MID,
    TEST2_INPUT_MIXED, TEST2_EXPECTED_HIGH_V, TEST2_EXPECTED_MID_V, TEST2_EXPECTED_LOW_V,
)

TOLERANCE = 1e-9


def _assert_series(test, actual, expected, label):
    test.assertEqual(len(actual), len(expected), f"{label}: length mismatch")
    for i in range(len(expected)):
        if math.isnan(expected[i]):
            test.assertTrue(math.isnan(actual[i]), f"{label}[{i}]: expected NaN, got {actual[i]}")
        else:
            test.assertAlmostEqual(actual[i], expected[i], delta=TOLERANCE,
                                   msg=f"{label}[{i}]: expected {expected[i]}, got {actual[i]}")


def _run(params, inputs):
    ind = SincWaveletBandpass(params)
    return [ind.update(c) for c in inputs]


class TestSincWaveletBandpassData(unittest.TestCase):
    """Test SWB against the reference test data for all band/velocity combos."""

    def test_bands(self):
        cases = [
            (Band.HIGH, False, EXPECTED_HIGH, "HIGH"),
            (Band.MID, False, EXPECTED_MID, "MID"),
            (Band.LOW, False, EXPECTED_LOW, "LOW"),
            (Band.FULL, False, EXPECTED_FULL, "FULL"),
            (Band.HIGH, True, EXPECTED_HIGH_V, "HIGH_V"),
            (Band.MID, True, EXPECTED_MID_V, "MID_V"),
            (Band.LOW, True, EXPECTED_LOW_V, "LOW_V"),
            (Band.FULL, True, EXPECTED_FULL_V, "FULL_V"),
        ]
        for band, velocity, expected, label in cases:
            with self.subTest(label=label):
                params = SincWaveletBandpassParams(band=band, velocity=velocity)
                _assert_series(self, _run(params, INPUT_CLOSE), expected, label)

    def test1_sine_mid(self):
        params = SincWaveletBandpassParams(band=Band.MID, velocity=False)
        _assert_series(self, _run(params, TEST1_INPUT_SINE), TEST1_EXPECTED_MID, "TEST1_MID")

    def test2_mixed_velocity(self):
        cases = [
            (Band.HIGH, TEST2_EXPECTED_HIGH_V, "TEST2_HIGH_V"),
            (Band.MID, TEST2_EXPECTED_MID_V, "TEST2_MID_V"),
            (Band.LOW, TEST2_EXPECTED_LOW_V, "TEST2_LOW_V"),
        ]
        for band, expected, label in cases:
            with self.subTest(label=label):
                params = SincWaveletBandpassParams(band=band, velocity=True)
                _assert_series(self, _run(params, TEST2_INPUT_MIXED), expected, label)


class TestSincWaveletBandpassMnemonic(unittest.TestCase):
    def test_default_mnemonic(self):
        self.assertEqual(SincWaveletBandpass(SincWaveletBandpassParams()).metadata().mnemonic, "swb(mid)")

    def test_band_mnemonics(self):
        self.assertEqual(SincWaveletBandpass(SincWaveletBandpassParams(band=Band.HIGH)).metadata().mnemonic, "swb(high)")
        self.assertEqual(SincWaveletBandpass(SincWaveletBandpassParams(band=Band.FULL)).metadata().mnemonic, "swb(full)")

    def test_velocity_mnemonics(self):
        self.assertEqual(
            SincWaveletBandpass(SincWaveletBandpassParams(band=Band.MID, velocity=True)).metadata().mnemonic,
            "swb(mid,v)")
        self.assertEqual(
            SincWaveletBandpass(SincWaveletBandpassParams(band=Band.FULL, velocity=True)).metadata().mnemonic,
            "swb(full,v)")


class TestSincWaveletBandpassMetadata(unittest.TestCase):
    def test_default_metadata(self):
        meta = SincWaveletBandpass(SincWaveletBandpassParams()).metadata()
        self.assertEqual(meta.identifier, Identifier.SINC_WAVELET_BANDPASS)
        self.assertEqual(meta.mnemonic, "swb(mid)")
        self.assertEqual(len(meta.outputs), 1)


class TestSincWaveletBandpassUpdateScalar(unittest.TestCase):
    def test_update_scalar(self):
        ind = SincWaveletBandpass(SincWaveletBandpassParams(band=Band.HIGH))
        tm = datetime(2021, 4, 1)
        out = None
        for c in INPUT_CLOSE:
            out = ind.update_scalar(Scalar(time=tm, value=c))
        self.assertEqual(len(out), 1)
        last = len(INPUT_CLOSE) - 1
        self.assertAlmostEqual(out[0].value, EXPECTED_HIGH[last], delta=TOLERANCE)


if __name__ == '__main__':
    unittest.main()
