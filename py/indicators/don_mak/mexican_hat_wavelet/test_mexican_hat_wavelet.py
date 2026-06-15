import math
import unittest
from datetime import datetime

from py.indicators.don_mak.mexican_hat_wavelet.mexican_hat_wavelet import MexicanHatWavelet
from py.indicators.don_mak.mexican_hat_wavelet.params import MexicanHatWaveletParams, Band
from py.indicators.core.identifier import Identifier
from py.entities.scalar import Scalar

from .test_testdata import (
    INPUT_CLOSE,
    EXPECTED_HIGH, EXPECTED_MID, EXPECTED_LOW,
    EXPECTED_P8, EXPECTED_P20, EXPECTED_P32,
    EXPECTED_D2_0, EXPECTED_D8_0,
    TEST1_INPUT_SINE, TEST1_EXPECTED_MID,
    TEST2_INPUT_MIXED, TEST2_EXPECTED_HIGH, TEST2_EXPECTED_MID, TEST2_EXPECTED_LOW,
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


def _run(ind, inputs):
    return [ind.update(c) for c in inputs]


class TestMexicanHatWaveletData(unittest.TestCase):
    """Test MHW against the reference test data for all band/custom combos."""

    def test_bands(self):
        cases = [
            (MexicanHatWaveletParams(band=Band.HIGH), EXPECTED_HIGH, "HIGH"),
            (MexicanHatWaveletParams(band=Band.MID), EXPECTED_MID, "MID"),
            (MexicanHatWaveletParams(band=Band.LOW), EXPECTED_LOW, "LOW"),
            (MexicanHatWaveletParams(band=Band.CUSTOM, period=8.0), EXPECTED_P8, "P8"),
            (MexicanHatWaveletParams(band=Band.CUSTOM, period=20.0), EXPECTED_P20, "P20"),
            (MexicanHatWaveletParams(band=Band.CUSTOM, period=32.0), EXPECTED_P32, "P32"),
            (MexicanHatWaveletParams(band=Band.CUSTOM, dilation=2.0), EXPECTED_D2_0, "D2_0"),
            (MexicanHatWaveletParams(band=Band.CUSTOM, dilation=8.0), EXPECTED_D8_0, "D8_0"),
        ]
        for params, expected, label in cases:
            with self.subTest(label=label):
                _assert_series(self, _run(MexicanHatWavelet(params), INPUT_CLOSE), expected, label)

    def test1_sine_mid(self):
        ind = MexicanHatWavelet(MexicanHatWaveletParams(band=Band.MID))
        _assert_series(self, _run(ind, TEST1_INPUT_SINE), TEST1_EXPECTED_MID, "TEST1_MID")

    def test2_mixed(self):
        cases = [
            (Band.HIGH, TEST2_EXPECTED_HIGH, "TEST2_HIGH"),
            (Band.MID, TEST2_EXPECTED_MID, "TEST2_MID"),
            (Band.LOW, TEST2_EXPECTED_LOW, "TEST2_LOW"),
        ]
        for band, expected, label in cases:
            with self.subTest(label=label):
                ind = MexicanHatWavelet(MexicanHatWaveletParams(band=band))
                _assert_series(self, _run(ind, TEST2_INPUT_MIXED), expected, label)


class TestMexicanHatWaveletMnemonic(unittest.TestCase):
    def test_default_mnemonic(self):
        ind = MexicanHatWavelet(MexicanHatWaveletParams())
        self.assertEqual(ind.metadata().mnemonic, "mhw(mid)")

    def test_band_mnemonics(self):
        self.assertEqual(MexicanHatWavelet(MexicanHatWaveletParams(band=Band.HIGH)).metadata().mnemonic, "mhw(high)")
        self.assertEqual(MexicanHatWavelet(MexicanHatWaveletParams(band=Band.LOW)).metadata().mnemonic, "mhw(low)")

    def test_custom_mnemonics(self):
        self.assertEqual(
            MexicanHatWavelet(MexicanHatWaveletParams(band=Band.CUSTOM, dilation=2.0)).metadata().mnemonic,
            "mhw(d2.00)")
        self.assertEqual(
            MexicanHatWavelet(MexicanHatWaveletParams(band=Band.CUSTOM, period=20.0)).metadata().mnemonic,
            "mhw(p20.00)")


class TestMexicanHatWaveletMetadata(unittest.TestCase):
    def test_default_metadata(self):
        ind = MexicanHatWavelet(MexicanHatWaveletParams())
        meta = ind.metadata()
        self.assertEqual(meta.identifier, Identifier.MEXICAN_HAT_WAVELET)
        self.assertEqual(meta.mnemonic, "mhw(mid)")
        self.assertEqual(len(meta.outputs), 1)


class TestMexicanHatWaveletUpdateScalar(unittest.TestCase):
    def test_update_scalar(self):
        ind = MexicanHatWavelet(MexicanHatWaveletParams(band=Band.HIGH))
        tm = datetime(2021, 4, 1)
        out = None
        for c in INPUT_CLOSE:
            out = ind.update_scalar(Scalar(time=tm, value=c))
        self.assertEqual(len(out), 1)
        last = len(INPUT_CLOSE) - 1
        self.assertAlmostEqual(out[0].value, EXPECTED_HIGH[last], delta=TOLERANCE)


class TestMexicanHatWaveletInvalidParams(unittest.TestCase):
    def test_custom_no_params(self):
        with self.assertRaises(ValueError):
            MexicanHatWavelet(MexicanHatWaveletParams(band=Band.CUSTOM))

    def test_custom_both_params(self):
        with self.assertRaises(ValueError):
            MexicanHatWavelet(MexicanHatWaveletParams(band=Band.CUSTOM, dilation=2.0, period=20.0))

    def test_custom_period_too_small(self):
        with self.assertRaises(ValueError):
            MexicanHatWavelet(MexicanHatWaveletParams(band=Band.CUSTOM, period=2.0))

    def test_custom_dilation_nonpositive(self):
        with self.assertRaises(ValueError):
            MexicanHatWavelet(MexicanHatWaveletParams(band=Band.CUSTOM, dilation=-1.0))


if __name__ == '__main__':
    unittest.main()
