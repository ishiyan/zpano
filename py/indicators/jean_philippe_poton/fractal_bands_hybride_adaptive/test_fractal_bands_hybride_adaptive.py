import math
import unittest
from datetime import datetime

from py.indicators.jean_philippe_poton.fractal_bands_hybride_adaptive.fractal_bands_hybride_adaptive import FractalBandsHybrideAdaptive
from py.indicators.jean_philippe_poton.fractal_bands_hybride_adaptive.params import FractalBandsHybrideAdaptiveParams
from py.indicators.core.identifier import Identifier
from py.entities.bar import Bar
from py.entities.scalar import Scalar

from .test_testdata import (
    INPUT_CLOSE,
    EXPECTED_FRASMA_P10_NY05_AHP007, EXPECTED_UPPER_P10_NY05_AHP007, EXPECTED_LOWER_P10_NY05_AHP007,
    EXPECTED_FRASMA_P10_NY05_AHP015, EXPECTED_UPPER_P10_NY05_AHP015, EXPECTED_LOWER_P10_NY05_AHP015,
    EXPECTED_FRASMA_P10_NY10_AHP007, EXPECTED_UPPER_P10_NY10_AHP007, EXPECTED_LOWER_P10_NY10_AHP007,
    EXPECTED_FRASMA_P10_NY10_AHP015, EXPECTED_UPPER_P10_NY10_AHP015, EXPECTED_LOWER_P10_NY10_AHP015,
    EXPECTED_FRASMA_P20_NY05_AHP007, EXPECTED_UPPER_P20_NY05_AHP007, EXPECTED_LOWER_P20_NY05_AHP007,
    EXPECTED_FRASMA_P20_NY05_AHP015, EXPECTED_UPPER_P20_NY05_AHP015, EXPECTED_LOWER_P20_NY05_AHP015,
    EXPECTED_FRASMA_P20_NY10_AHP007, EXPECTED_UPPER_P20_NY10_AHP007, EXPECTED_LOWER_P20_NY10_AHP007,
    EXPECTED_FRASMA_P20_NY10_AHP015, EXPECTED_UPPER_P20_NY10_AHP015, EXPECTED_LOWER_P20_NY10_AHP015,
    EXPECTED_FRASMA_P30_NY05_AHP007, EXPECTED_UPPER_P30_NY05_AHP007, EXPECTED_LOWER_P30_NY05_AHP007,
    EXPECTED_FRASMA_P30_NY05_AHP015, EXPECTED_UPPER_P30_NY05_AHP015, EXPECTED_LOWER_P30_NY05_AHP015,
    EXPECTED_FRASMA_P30_NY10_AHP007, EXPECTED_UPPER_P30_NY10_AHP007, EXPECTED_LOWER_P30_NY10_AHP007,
    EXPECTED_FRASMA_P30_NY10_AHP015, EXPECTED_UPPER_P30_NY10_AHP015, EXPECTED_LOWER_P30_NY10_AHP015,
    EXPECTED_FRASMA_P50_NY05_AHP007, EXPECTED_UPPER_P50_NY05_AHP007, EXPECTED_LOWER_P50_NY05_AHP007,
    EXPECTED_FRASMA_P50_NY05_AHP015, EXPECTED_UPPER_P50_NY05_AHP015, EXPECTED_LOWER_P50_NY05_AHP015,
    EXPECTED_FRASMA_P50_NY10_AHP007, EXPECTED_UPPER_P50_NY10_AHP007, EXPECTED_LOWER_P50_NY10_AHP007,
    EXPECTED_FRASMA_P50_NY10_AHP015, EXPECTED_UPPER_P50_NY10_AHP015, EXPECTED_LOWER_P50_NY10_AHP015,
)


class TestFractalBandsHybrideAdaptive(unittest.TestCase):

    def _run_test(self, period, normal_speed_fallback, alpha, nyquist, alpha_hp,
                   exp_frasma2, exp_upper, exp_lower):
        ind = FractalBandsHybrideAdaptive(FractalBandsHybrideAdaptiveParams(
            period=period, normal_speed_fallback=normal_speed_fallback,
            alpha=alpha, nyquist=nyquist, alpha_hp=alpha_hp))
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

    def test_p10_ny05_ahp007(self):
        self._run_test(10, 30, 2.0, 0.5, 0.07,
                       EXPECTED_FRASMA_P10_NY05_AHP007, EXPECTED_UPPER_P10_NY05_AHP007, EXPECTED_LOWER_P10_NY05_AHP007)

    def test_p10_ny05_ahp015(self):
        self._run_test(10, 30, 2.0, 0.5, 0.15,
                       EXPECTED_FRASMA_P10_NY05_AHP015, EXPECTED_UPPER_P10_NY05_AHP015, EXPECTED_LOWER_P10_NY05_AHP015)

    def test_p10_ny10_ahp007(self):
        self._run_test(10, 30, 2.0, 1.0, 0.07,
                       EXPECTED_FRASMA_P10_NY10_AHP007, EXPECTED_UPPER_P10_NY10_AHP007, EXPECTED_LOWER_P10_NY10_AHP007)

    def test_p10_ny10_ahp015(self):
        self._run_test(10, 30, 2.0, 1.0, 0.15,
                       EXPECTED_FRASMA_P10_NY10_AHP015, EXPECTED_UPPER_P10_NY10_AHP015, EXPECTED_LOWER_P10_NY10_AHP015)

    def test_p20_ny05_ahp007(self):
        self._run_test(20, 30, 2.0, 0.5, 0.07,
                       EXPECTED_FRASMA_P20_NY05_AHP007, EXPECTED_UPPER_P20_NY05_AHP007, EXPECTED_LOWER_P20_NY05_AHP007)

    def test_p20_ny05_ahp015(self):
        self._run_test(20, 30, 2.0, 0.5, 0.15,
                       EXPECTED_FRASMA_P20_NY05_AHP015, EXPECTED_UPPER_P20_NY05_AHP015, EXPECTED_LOWER_P20_NY05_AHP015)

    def test_p20_ny10_ahp007(self):
        self._run_test(20, 30, 2.0, 1.0, 0.07,
                       EXPECTED_FRASMA_P20_NY10_AHP007, EXPECTED_UPPER_P20_NY10_AHP007, EXPECTED_LOWER_P20_NY10_AHP007)

    def test_p20_ny10_ahp015(self):
        self._run_test(20, 30, 2.0, 1.0, 0.15,
                       EXPECTED_FRASMA_P20_NY10_AHP015, EXPECTED_UPPER_P20_NY10_AHP015, EXPECTED_LOWER_P20_NY10_AHP015)

    def test_p30_ny05_ahp007(self):
        self._run_test(30, 30, 2.0, 0.5, 0.07,
                       EXPECTED_FRASMA_P30_NY05_AHP007, EXPECTED_UPPER_P30_NY05_AHP007, EXPECTED_LOWER_P30_NY05_AHP007)

    def test_p30_ny05_ahp015(self):
        self._run_test(30, 30, 2.0, 0.5, 0.15,
                       EXPECTED_FRASMA_P30_NY05_AHP015, EXPECTED_UPPER_P30_NY05_AHP015, EXPECTED_LOWER_P30_NY05_AHP015)

    def test_p30_ny10_ahp007(self):
        self._run_test(30, 30, 2.0, 1.0, 0.07,
                       EXPECTED_FRASMA_P30_NY10_AHP007, EXPECTED_UPPER_P30_NY10_AHP007, EXPECTED_LOWER_P30_NY10_AHP007)

    def test_p30_ny10_ahp015(self):
        self._run_test(30, 30, 2.0, 1.0, 0.15,
                       EXPECTED_FRASMA_P30_NY10_AHP015, EXPECTED_UPPER_P30_NY10_AHP015, EXPECTED_LOWER_P30_NY10_AHP015)

    def test_p50_ny05_ahp007(self):
        self._run_test(50, 30, 2.0, 0.5, 0.07,
                       EXPECTED_FRASMA_P50_NY05_AHP007, EXPECTED_UPPER_P50_NY05_AHP007, EXPECTED_LOWER_P50_NY05_AHP007)

    def test_p50_ny05_ahp015(self):
        self._run_test(50, 30, 2.0, 0.5, 0.15,
                       EXPECTED_FRASMA_P50_NY05_AHP015, EXPECTED_UPPER_P50_NY05_AHP015, EXPECTED_LOWER_P50_NY05_AHP015)

    def test_p50_ny10_ahp007(self):
        self._run_test(50, 30, 2.0, 1.0, 0.07,
                       EXPECTED_FRASMA_P50_NY10_AHP007, EXPECTED_UPPER_P50_NY10_AHP007, EXPECTED_LOWER_P50_NY10_AHP007)

    def test_p50_ny10_ahp015(self):
        self._run_test(50, 30, 2.0, 1.0, 0.15,
                       EXPECTED_FRASMA_P50_NY10_AHP015, EXPECTED_UPPER_P50_NY10_AHP015, EXPECTED_LOWER_P50_NY10_AHP015)

    def test_is_primed(self):
        ind = FractalBandsHybrideAdaptive(FractalBandsHybrideAdaptiveParams(period=30))
        for i in range(30):
            ind.update(INPUT_CLOSE[i])
            self.assertFalse(ind.is_primed())
        ind.update(INPUT_CLOSE[30])
        self.assertTrue(ind.is_primed())

    def test_nan_passthrough(self):
        ind = FractalBandsHybrideAdaptive(FractalBandsHybrideAdaptiveParams(period=5))
        frasma2, upper, lower = ind.update(math.nan)
        self.assertTrue(math.isnan(frasma2))
        self.assertTrue(math.isnan(upper))
        self.assertTrue(math.isnan(lower))

    def test_invalid_period(self):
        with self.assertRaises(ValueError):
            FractalBandsHybrideAdaptive(FractalBandsHybrideAdaptiveParams(period=1))

    def test_invalid_normal_speed_fallback(self):
        with self.assertRaises(ValueError):
            FractalBandsHybrideAdaptive(FractalBandsHybrideAdaptiveParams(normal_speed_fallback=0))

    def test_invalid_alpha(self):
        with self.assertRaises(ValueError):
            FractalBandsHybrideAdaptive(FractalBandsHybrideAdaptiveParams(alpha=0.0))

    def test_invalid_nyquist(self):
        with self.assertRaises(ValueError):
            FractalBandsHybrideAdaptive(FractalBandsHybrideAdaptiveParams(nyquist=0.0))

    def test_invalid_alpha_hp(self):
        with self.assertRaises(ValueError):
            FractalBandsHybrideAdaptive(FractalBandsHybrideAdaptiveParams(alpha_hp=0.0))
        with self.assertRaises(ValueError):
            FractalBandsHybrideAdaptive(FractalBandsHybrideAdaptiveParams(alpha_hp=1.0))

    def test_metadata(self):
        ind = FractalBandsHybrideAdaptive(FractalBandsHybrideAdaptiveParams(period=30))
        meta = ind.metadata()
        self.assertEqual(meta.identifier, Identifier.FRACTAL_BANDS_HYBRIDE_ADAPTIVE)
        self.assertIn("fbanha(30", meta.mnemonic)


if __name__ == '__main__':
    unittest.main()
