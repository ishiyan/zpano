import math
import unittest
from datetime import datetime

from py.indicators.john_bollinger.bollinger_bands.bollinger_bands import BollingerBands
from py.indicators.john_bollinger.bollinger_bands.params import BollingerBandsParams
from py.indicators.john_bollinger.bollinger_bands.output import BollingerBandsOutput
from py.indicators.core.identifier import Identifier
from py.indicators.core.outputs.shape import Shape
from py.entities.scalar import Scalar

from .test_testdata import (
    NaN,
    TEST_CLOSING_PRICE,
    TEST_SMA20,
    TEST_SAMPLE_LOWER,
    TEST_SAMPLE_UPPER,
    TEST_SAMPLE_BW,
    TEST_SAMPLE_PCTB,
    TEST_POP_LOWER,
    TEST_POP_UPPER,
    TEST_POP_BW,
    TEST_POP_PCTB,
)

class TestBollingerBandsSampleStdDev(unittest.TestCase):
    """Test BB with sample (unbiased) standard deviation, length=20."""

    def test_full_data(self):
        tolerance = 1e-8
        ind = BollingerBands(BollingerBandsParams(length=20, is_unbiased=True))

        for i in range(252):
            result = ind.update(TEST_CLOSING_PRICE[i])
            lower, middle, upper, bw, pctb = result

            if math.isnan(TEST_SMA20[i]):
                self.assertTrue(math.isnan(lower), f"[{i}] lower expected NaN")
                self.assertTrue(math.isnan(middle), f"[{i}] middle expected NaN")
                self.assertTrue(math.isnan(upper), f"[{i}] upper expected NaN")
                continue

            self.assertAlmostEqual(middle, TEST_SMA20[i], delta=tolerance, msg=f"[{i}] middle")
            self.assertAlmostEqual(lower, TEST_SAMPLE_LOWER[i], delta=tolerance, msg=f"[{i}] lower")
            self.assertAlmostEqual(upper, TEST_SAMPLE_UPPER[i], delta=tolerance, msg=f"[{i}] upper")
            self.assertAlmostEqual(bw, TEST_SAMPLE_BW[i], delta=tolerance, msg=f"[{i}] bandWidth")
            self.assertAlmostEqual(pctb, TEST_SAMPLE_PCTB[i], delta=tolerance, msg=f"[{i}] percentBand")


class TestBollingerBandsPopulationStdDev(unittest.TestCase):
    """Test BB with population (biased) standard deviation, length=20."""

    def test_full_data(self):
        tolerance = 1e-8
        ind = BollingerBands(BollingerBandsParams(length=20, is_unbiased=False))

        for i in range(252):
            result = ind.update(TEST_CLOSING_PRICE[i])
            lower, middle, upper, bw, pctb = result

            if math.isnan(TEST_SMA20[i]):
                self.assertTrue(math.isnan(lower), f"[{i}] lower expected NaN")
                self.assertTrue(math.isnan(middle), f"[{i}] middle expected NaN")
                self.assertTrue(math.isnan(upper), f"[{i}] upper expected NaN")
                continue

            self.assertAlmostEqual(middle, TEST_SMA20[i], delta=tolerance, msg=f"[{i}] middle")
            self.assertAlmostEqual(lower, TEST_POP_LOWER[i], delta=tolerance, msg=f"[{i}] lower")
            self.assertAlmostEqual(upper, TEST_POP_UPPER[i], delta=tolerance, msg=f"[{i}] upper")
            self.assertAlmostEqual(bw, TEST_POP_BW[i], delta=tolerance, msg=f"[{i}] bandWidth")
            self.assertAlmostEqual(pctb, TEST_POP_PCTB[i], delta=tolerance, msg=f"[{i}] percentBand")


class TestBollingerBandsIsPrimed(unittest.TestCase):
    def test_is_primed(self):
        ind = BollingerBands(BollingerBandsParams(length=20, is_unbiased=True))

        self.assertFalse(ind.is_primed())

        for i in range(19):
            ind.update(TEST_CLOSING_PRICE[i])
            self.assertFalse(ind.is_primed(), f"[{i}] expected not primed")

        ind.update(TEST_CLOSING_PRICE[19])
        self.assertTrue(ind.is_primed())


class TestBollingerBandsNaN(unittest.TestCase):
    def test_nan_input(self):
        ind = BollingerBands(BollingerBandsParams(length=20))
        lower, middle, upper, bw, pctb = ind.update(math.nan)

        self.assertTrue(math.isnan(lower))
        self.assertTrue(math.isnan(middle))
        self.assertTrue(math.isnan(upper))
        self.assertTrue(math.isnan(bw))
        self.assertTrue(math.isnan(pctb))


class TestBollingerBandsMetadata(unittest.TestCase):
    def test_metadata(self):
        ind = BollingerBands(BollingerBandsParams(length=20))
        meta = ind.metadata()

        self.assertEqual(meta.identifier, Identifier.BOLLINGER_BANDS)
        self.assertEqual(len(meta.outputs), 6)
        self.assertEqual(meta.outputs[0].kind, int(BollingerBandsOutput.LOWER))
        self.assertEqual(meta.outputs[0].shape, Shape.SCALAR)
        self.assertEqual(meta.outputs[1].kind, int(BollingerBandsOutput.MIDDLE))
        self.assertEqual(meta.outputs[5].kind, int(BollingerBandsOutput.BAND))
        self.assertEqual(meta.outputs[5].shape, Shape.BAND)


class TestBollingerBandsUpdateScalar(unittest.TestCase):
    def test_update_scalar(self):
        tolerance = 1e-8
        ind = BollingerBands(BollingerBandsParams(length=20, is_unbiased=True))
        tm = datetime(2021, 4, 1)

        # First 19 samples — all NaN, empty band.
        for i in range(19):
            out = ind.update_scalar(Scalar(time=tm, value=TEST_CLOSING_PRICE[i]))
            self.assertTrue(math.isnan(out[0].value), f"[{i}] expected NaN lower scalar")
            self.assertTrue(out[5].is_empty(), f"[{i}] expected empty band")

        # Index 19 — first primed value.
        out = ind.update_scalar(Scalar(time=tm, value=TEST_CLOSING_PRICE[19]))
        self.assertAlmostEqual(out[1].value, TEST_SMA20[19], delta=tolerance, msg="[19] middle")
        self.assertAlmostEqual(out[0].value, TEST_SAMPLE_LOWER[19], delta=tolerance, msg="[19] lower")
        self.assertAlmostEqual(out[2].value, TEST_SAMPLE_UPPER[19], delta=tolerance, msg="[19] upper")

        self.assertFalse(out[5].is_empty(), "[19] expected non-empty band")
        self.assertAlmostEqual(out[5].lower, TEST_SAMPLE_LOWER[19], delta=tolerance, msg="[19] band.lower")
        self.assertAlmostEqual(out[5].upper, TEST_SAMPLE_UPPER[19], delta=tolerance, msg="[19] band.upper")


class TestBollingerBandsInvalidParams(unittest.TestCase):
    def test_length_too_small(self):
        with self.assertRaises(ValueError):
            BollingerBands(BollingerBandsParams(length=1))

    def test_length_negative(self):
        with self.assertRaises(ValueError):
            BollingerBands(BollingerBandsParams(length=-1))


if __name__ == '__main__':
    unittest.main()
