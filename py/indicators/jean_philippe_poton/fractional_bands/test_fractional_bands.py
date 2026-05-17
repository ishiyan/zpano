import math
import unittest
from datetime import datetime

from py.indicators.jean_philippe_poton.fractional_bands.fractional_bands import FractionalBands
from py.indicators.jean_philippe_poton.fractional_bands.params import FractionalBandsParams
from py.indicators.core.identifier import Identifier
from py.entities.bar import Bar
from py.entities.scalar import Scalar

from .test_testdata import (
    INPUT_CLOSE,
    EXPECTED_FRASMA2_P5_S1, EXPECTED_UPPER_P5_S1, EXPECTED_LOWER_P5_S1,
    EXPECTED_FRASMA2_P10_S1, EXPECTED_UPPER_P10_S1, EXPECTED_LOWER_P10_S1,
    EXPECTED_FRASMA2_P20_S1, EXPECTED_UPPER_P20_S1, EXPECTED_LOWER_P20_S1,
    EXPECTED_FRASMA2_P30_S1, EXPECTED_UPPER_P30_S1, EXPECTED_LOWER_P30_S1,
    EXPECTED_FRASMA2_P50_S1, EXPECTED_UPPER_P50_S1, EXPECTED_LOWER_P50_S1,
    EXPECTED_FRASMA2_P80_S1, EXPECTED_UPPER_P80_S1, EXPECTED_LOWER_P80_S1,
    EXPECTED_FRASMA2_P30_S100, EXPECTED_UPPER_P30_S100, EXPECTED_LOWER_P30_S100,
    EXPECTED_FRASMA2_P30_S10000, EXPECTED_UPPER_P30_S10000, EXPECTED_LOWER_P30_S10000,
)


class TestFractionalBands(unittest.TestCase):

    def _run_test(self, period, price_scale, exp_frasma2, exp_upper, exp_lower):
        ind = FractionalBands(FractionalBandsParams(period=period, price_scale=price_scale))
        for i, val in enumerate(INPUT_CLOSE):
            frasma2, upper, lower = ind.update(val)
            if math.isnan(exp_frasma2[i]):
                self.assertTrue(math.isnan(frasma2), f"index {i}: expected NaN for frasma2")
            else:
                self.assertAlmostEqual(frasma2, exp_frasma2[i], places=13, msg=f"frasma2 at {i}")
            if math.isnan(exp_upper[i]):
                self.assertTrue(math.isnan(upper), f"index {i}: expected NaN for upper")
            else:
                self.assertAlmostEqual(upper, exp_upper[i], places=13, msg=f"upper at {i}")
            if math.isnan(exp_lower[i]):
                self.assertTrue(math.isnan(lower), f"index {i}: expected NaN for lower")
            else:
                self.assertAlmostEqual(lower, exp_lower[i], places=13, msg=f"lower at {i}")

    def test_p5_s1(self):
        self._run_test(5, 1.0, EXPECTED_FRASMA2_P5_S1, EXPECTED_UPPER_P5_S1, EXPECTED_LOWER_P5_S1)

    def test_p10_s1(self):
        self._run_test(10, 1.0, EXPECTED_FRASMA2_P10_S1, EXPECTED_UPPER_P10_S1, EXPECTED_LOWER_P10_S1)

    def test_p20_s1(self):
        self._run_test(20, 1.0, EXPECTED_FRASMA2_P20_S1, EXPECTED_UPPER_P20_S1, EXPECTED_LOWER_P20_S1)

    def test_p30_s1(self):
        self._run_test(30, 1.0, EXPECTED_FRASMA2_P30_S1, EXPECTED_UPPER_P30_S1, EXPECTED_LOWER_P30_S1)

    def test_p50_s1(self):
        self._run_test(50, 1.0, EXPECTED_FRASMA2_P50_S1, EXPECTED_UPPER_P50_S1, EXPECTED_LOWER_P50_S1)

    def test_p80_s1(self):
        self._run_test(80, 1.0, EXPECTED_FRASMA2_P80_S1, EXPECTED_UPPER_P80_S1, EXPECTED_LOWER_P80_S1)

    def test_p30_s100(self):
        self._run_test(30, 100.0, EXPECTED_FRASMA2_P30_S100, EXPECTED_UPPER_P30_S100, EXPECTED_LOWER_P30_S100)

    def test_p30_s10000(self):
        self._run_test(30, 10000.0, EXPECTED_FRASMA2_P30_S10000, EXPECTED_UPPER_P30_S10000, EXPECTED_LOWER_P30_S10000)

    def test_is_primed(self):
        ind = FractionalBands(FractionalBandsParams(period=30))
        for i in range(30):
            ind.update(INPUT_CLOSE[i])
            self.assertFalse(ind.is_primed())
        ind.update(INPUT_CLOSE[30])
        self.assertTrue(ind.is_primed())

    def test_nan_passthrough(self):
        ind = FractionalBands(FractionalBandsParams(period=5))
        frasma2, upper, lower = ind.update(math.nan)
        self.assertTrue(math.isnan(frasma2))
        self.assertTrue(math.isnan(upper))
        self.assertTrue(math.isnan(lower))

    def test_invalid_period(self):
        with self.assertRaises(ValueError):
            FractionalBands(FractionalBandsParams(period=1))

    def test_invalid_price_scale(self):
        with self.assertRaises(ValueError):
            FractionalBands(FractionalBandsParams(price_scale=0.0))
        with self.assertRaises(ValueError):
            FractionalBands(FractionalBandsParams(price_scale=-1.0))

    def test_metadata(self):
        ind = FractionalBands(FractionalBandsParams(period=30))
        meta = ind.metadata()
        self.assertEqual(meta.identifier, Identifier.FRACTIONAL_BANDS)
        self.assertIn("fctban(30", meta.mnemonic)

    def test_update_bar(self):
        ind = FractionalBands(FractionalBandsParams(period=10, price_scale=1.0))
        for i, val in enumerate(INPUT_CLOSE):
            bar = Bar(datetime(2020, 1, 1), val + 1, val + 2, val - 1, val, 100.0)
            output = ind.update_bar(bar)
            expected = EXPECTED_FRASMA2_P10_S1[i]
            if math.isnan(expected):
                self.assertTrue(math.isnan(output[0].value))
            else:
                self.assertAlmostEqual(output[0].value, expected, places=13)

    def test_update_scalar(self):
        ind = FractionalBands(FractionalBandsParams(period=10, price_scale=1.0))
        for i, val in enumerate(INPUT_CLOSE):
            scalar = Scalar(datetime(2020, 1, 1), val)
            output = ind.update_scalar(scalar)
            expected = EXPECTED_FRASMA2_P10_S1[i]
            if math.isnan(expected):
                self.assertTrue(math.isnan(output[0].value))
            else:
                self.assertAlmostEqual(output[0].value, expected, places=13)


if __name__ == '__main__':
    unittest.main()
