import math
import unittest
from datetime import datetime

from py.indicators.jean_philippe_poton.rescaled_fractal_adaptive_simple_moving_average.rescaled_fractal_adaptive_simple_moving_average import RescaledFractalAdaptiveSimpleMovingAverage
from py.indicators.jean_philippe_poton.rescaled_fractal_adaptive_simple_moving_average.params import RescaledFractalAdaptiveSimpleMovingAverageParams
from py.indicators.core.identifier import Identifier
from py.entities.bar import Bar
from py.entities.scalar import Scalar

from .test_testdata import (
    INPUT_CLOSE,
    EXPECTED_P4_S1,
    EXPECTED_P8_S1,
    EXPECTED_P16_S1,
    EXPECTED_P32_S1,
    EXPECTED_P64_S1,
    EXPECTED_P128_S1,
    EXPECTED_P32_S100,
    EXPECTED_P32_S10000,
)


class TestRescaledFractalAdaptiveSimpleMovingAverage(unittest.TestCase):

    def _run_test(self, period, normal_speed, price_scale, expected):
        ind = RescaledFractalAdaptiveSimpleMovingAverage(
            RescaledFractalAdaptiveSimpleMovingAverageParams(
                period=period, normal_speed=normal_speed, price_scale=price_scale))
        for i, val in enumerate(INPUT_CLOSE):
            result = ind.update(val)
            exp = expected[i]
            if math.isnan(exp):
                self.assertTrue(math.isnan(result), f"[{i}] expected NaN, got {result}")
            else:
                self.assertAlmostEqual(result, exp, places=13,
                                       msg=f"[{i}] expected {exp}, got {result}")

    def test_update_p4_s1(self):
        self._run_test(4, 30, 1.0, EXPECTED_P4_S1)

    def test_update_p8_s1(self):
        self._run_test(8, 30, 1.0, EXPECTED_P8_S1)

    def test_update_p16_s1(self):
        self._run_test(16, 30, 1.0, EXPECTED_P16_S1)

    def test_update_p32_s1(self):
        self._run_test(32, 30, 1.0, EXPECTED_P32_S1)

    def test_update_p64_s1(self):
        self._run_test(64, 30, 1.0, EXPECTED_P64_S1)

    def test_update_p128_s1(self):
        self._run_test(128, 30, 1.0, EXPECTED_P128_S1)

    def test_update_p32_s100(self):
        self._run_test(32, 30, 100.0, EXPECTED_P32_S100)

    def test_update_p32_s10000(self):
        self._run_test(32, 30, 10000.0, EXPECTED_P32_S10000)

    def test_is_primed_p64(self):
        ind = RescaledFractalAdaptiveSimpleMovingAverage(
            RescaledFractalAdaptiveSimpleMovingAverageParams(period=64, normal_speed=30))
        for i in range(64):
            ind.update(INPUT_CLOSE[i])
            self.assertFalse(ind.is_primed())
        ind.update(INPUT_CLOSE[64])
        self.assertTrue(ind.is_primed())

    def test_nan_passthrough(self):
        ind = RescaledFractalAdaptiveSimpleMovingAverage(
            RescaledFractalAdaptiveSimpleMovingAverageParams(period=4, normal_speed=30))
        result = ind.update(math.nan)
        self.assertTrue(math.isnan(result))

    def test_invalid_period_too_small(self):
        with self.assertRaises(ValueError):
            RescaledFractalAdaptiveSimpleMovingAverage(
                RescaledFractalAdaptiveSimpleMovingAverageParams(period=2, normal_speed=30))

    def test_invalid_period_not_power_of_2(self):
        with self.assertRaises(ValueError):
            RescaledFractalAdaptiveSimpleMovingAverage(
                RescaledFractalAdaptiveSimpleMovingAverageParams(period=6, normal_speed=30))

    def test_invalid_normal_speed(self):
        with self.assertRaises(ValueError):
            RescaledFractalAdaptiveSimpleMovingAverage(
                RescaledFractalAdaptiveSimpleMovingAverageParams(period=4, normal_speed=0))

    def test_metadata(self):
        ind = RescaledFractalAdaptiveSimpleMovingAverage(
            RescaledFractalAdaptiveSimpleMovingAverageParams(period=64, normal_speed=30))
        meta = ind.metadata()
        self.assertEqual(meta.identifier, Identifier.RESCALED_FRACTAL_ADAPTIVE_SIMPLE_MOVING_AVERAGE)
        self.assertIn("rsfrasma(64,30,1.0)", meta.mnemonic)

    def test_update_bar(self):
        ind = RescaledFractalAdaptiveSimpleMovingAverage(
            RescaledFractalAdaptiveSimpleMovingAverageParams(period=4, normal_speed=30))
        for i, val in enumerate(INPUT_CLOSE):
            bar = Bar(datetime(2020, 1, 1), val + 1, val + 2, val - 1, val, 100.0)
            output = ind.update_bar(bar)
            expected = EXPECTED_P4_S1[i]
            if math.isnan(expected):
                self.assertTrue(math.isnan(output[0].value))
            else:
                self.assertAlmostEqual(output[0].value, expected, places=13)


if __name__ == '__main__':
    unittest.main()
