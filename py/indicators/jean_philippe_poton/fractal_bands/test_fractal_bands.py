import math
import unittest
from datetime import datetime

from py.indicators.jean_philippe_poton.fractal_bands.fractal_bands import FractalBands
from py.indicators.jean_philippe_poton.fractal_bands.params import FractalBandsParams
from py.indicators.core.identifier import Identifier
from py.entities.bar import Bar
from py.entities.scalar import Scalar

from .test_testdata import (
    INPUT_CLOSE,
    EXPECTED_FRASMA2_P10_NS20_A2, EXPECTED_UPPER_P10_NS20_A2, EXPECTED_LOWER_P10_NS20_A2,
    EXPECTED_FRASMA2_P20_NS20_A2, EXPECTED_UPPER_P20_NS20_A2, EXPECTED_LOWER_P20_NS20_A2,
    EXPECTED_FRASMA2_P30_NS20_A2, EXPECTED_UPPER_P30_NS20_A2, EXPECTED_LOWER_P30_NS20_A2,
    EXPECTED_FRASMA2_P50_NS20_A2, EXPECTED_UPPER_P50_NS20_A2, EXPECTED_LOWER_P50_NS20_A2,
    EXPECTED_FRASMA2_P30_NS10_A2, EXPECTED_UPPER_P30_NS10_A2, EXPECTED_LOWER_P30_NS10_A2,
    EXPECTED_FRASMA2_P30_NS40_A2, EXPECTED_UPPER_P30_NS40_A2, EXPECTED_LOWER_P30_NS40_A2,
    EXPECTED_FRASMA2_P30_NS20_A1, EXPECTED_UPPER_P30_NS20_A1, EXPECTED_LOWER_P30_NS20_A1,
    EXPECTED_FRASMA2_P30_NS20_A3, EXPECTED_UPPER_P30_NS20_A3, EXPECTED_LOWER_P30_NS20_A3,
)


class TestFractalBands(unittest.TestCase):

    def _run_test(self, period, normal_speed, alpha, exp_frasma2, exp_upper, exp_lower):
        ind = FractalBands(FractalBandsParams(period=period, normal_speed=normal_speed, alpha=alpha))
        for i, val in enumerate(INPUT_CLOSE):
            frasma2, upper, lower = ind.update(val)
            if math.isnan(exp_frasma2[i]):
                self.assertTrue(math.isnan(frasma2), f"index {i}: expected NaN for frasma")
            else:
                self.assertAlmostEqual(frasma2, exp_frasma2[i], places=13, msg=f"frasma at {i}")
            if math.isnan(exp_upper[i]):
                self.assertTrue(math.isnan(upper), f"index {i}: expected NaN for upper")
            else:
                self.assertAlmostEqual(upper, exp_upper[i], places=13, msg=f"upper at {i}")
            if math.isnan(exp_lower[i]):
                self.assertTrue(math.isnan(lower), f"index {i}: expected NaN for lower")
            else:
                self.assertAlmostEqual(lower, exp_lower[i], places=13, msg=f"lower at {i}")

    def test_p10_ns20_a2(self):
        self._run_test(10, 20, 2.0, EXPECTED_FRASMA2_P10_NS20_A2, EXPECTED_UPPER_P10_NS20_A2, EXPECTED_LOWER_P10_NS20_A2)

    def test_p20_ns20_a2(self):
        self._run_test(20, 20, 2.0, EXPECTED_FRASMA2_P20_NS20_A2, EXPECTED_UPPER_P20_NS20_A2, EXPECTED_LOWER_P20_NS20_A2)

    def test_p30_ns20_a2(self):
        self._run_test(30, 20, 2.0, EXPECTED_FRASMA2_P30_NS20_A2, EXPECTED_UPPER_P30_NS20_A2, EXPECTED_LOWER_P30_NS20_A2)

    def test_p50_ns20_a2(self):
        self._run_test(50, 20, 2.0, EXPECTED_FRASMA2_P50_NS20_A2, EXPECTED_UPPER_P50_NS20_A2, EXPECTED_LOWER_P50_NS20_A2)

    def test_p30_ns10_a2(self):
        self._run_test(30, 10, 2.0, EXPECTED_FRASMA2_P30_NS10_A2, EXPECTED_UPPER_P30_NS10_A2, EXPECTED_LOWER_P30_NS10_A2)

    def test_p30_ns40_a2(self):
        self._run_test(30, 40, 2.0, EXPECTED_FRASMA2_P30_NS40_A2, EXPECTED_UPPER_P30_NS40_A2, EXPECTED_LOWER_P30_NS40_A2)

    def test_p30_ns20_a1(self):
        self._run_test(30, 20, 1.0, EXPECTED_FRASMA2_P30_NS20_A1, EXPECTED_UPPER_P30_NS20_A1, EXPECTED_LOWER_P30_NS20_A1)

    def test_p30_ns20_a3(self):
        self._run_test(30, 20, 3.0, EXPECTED_FRASMA2_P30_NS20_A3, EXPECTED_UPPER_P30_NS20_A3, EXPECTED_LOWER_P30_NS20_A3)

    def test_is_primed(self):
        ind = FractalBands(FractalBandsParams(period=30))
        for i in range(29):
            ind.update(INPUT_CLOSE[i])
            self.assertFalse(ind.is_primed())
        ind.update(INPUT_CLOSE[29])
        self.assertTrue(ind.is_primed())

    def test_nan_passthrough(self):
        ind = FractalBands(FractalBandsParams(period=5))
        frasma2, upper, lower = ind.update(math.nan)
        self.assertTrue(math.isnan(frasma2))
        self.assertTrue(math.isnan(upper))
        self.assertTrue(math.isnan(lower))

    def test_invalid_period(self):
        with self.assertRaises(ValueError):
            FractalBands(FractalBandsParams(period=1))

    def test_invalid_normal_speed(self):
        with self.assertRaises(ValueError):
            FractalBands(FractalBandsParams(normal_speed=0))

    def test_invalid_alpha(self):
        with self.assertRaises(ValueError):
            FractalBands(FractalBandsParams(alpha=0.0))

    def test_metadata(self):
        ind = FractalBands(FractalBandsParams(period=30))
        meta = ind.metadata()
        self.assertEqual(meta.identifier, Identifier.FRACTAL_BANDS)
        self.assertIn("fban(30", meta.mnemonic)

    def test_update_bar(self):
        ind = FractalBands(FractalBandsParams(period=10, normal_speed=20, alpha=2.0))
        for i, val in enumerate(INPUT_CLOSE):
            bar = Bar(datetime(2020, 1, 1), val + 1, val + 2, val - 1, val, 100.0)
            output = ind.update_bar(bar)
            expected = EXPECTED_FRASMA2_P10_NS20_A2[i]
            if math.isnan(expected):
                self.assertTrue(math.isnan(output[0].value))
            else:
                self.assertAlmostEqual(output[0].value, expected, places=13)

    def test_update_scalar(self):
        ind = FractalBands(FractalBandsParams(period=10, normal_speed=20, alpha=2.0))
        for i, val in enumerate(INPUT_CLOSE):
            scalar = Scalar(datetime(2020, 1, 1), val)
            output = ind.update_scalar(scalar)
            expected = EXPECTED_FRASMA2_P10_NS20_A2[i]
            if math.isnan(expected):
                self.assertTrue(math.isnan(output[0].value))
            else:
                self.assertAlmostEqual(output[0].value, expected, places=13)


if __name__ == '__main__':
    unittest.main()
