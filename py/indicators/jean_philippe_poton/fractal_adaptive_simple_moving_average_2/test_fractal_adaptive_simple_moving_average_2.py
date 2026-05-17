import math
import unittest
from datetime import datetime

from py.indicators.jean_philippe_poton.fractal_adaptive_simple_moving_average_2.fractal_adaptive_simple_moving_average_2 import FractalAdaptiveSimpleMovingAverage2
from py.indicators.jean_philippe_poton.fractal_adaptive_simple_moving_average_2.params import FractalAdaptiveSimpleMovingAverage2Params
from py.indicators.core.identifier import Identifier
from py.entities.bar import Bar
from py.entities.scalar import Scalar

from .test_testdata import (
    INPUT_CLOSE,
    EXPECTED_P5,
    EXPECTED_P10,
    EXPECTED_P15,
    EXPECTED_P20,
    EXPECTED_P30,
    EXPECTED_P50,
    EXPECTED_P80,
    EXPECTED_P120,
)


class TestFractalAdaptiveSimpleMovingAverage2(unittest.TestCase):

    def test_update_period_5(self):
        frasma2 = FractalAdaptiveSimpleMovingAverage2(
            FractalAdaptiveSimpleMovingAverage2Params(period=5, normal_speed=20))
        for i, val in enumerate(INPUT_CLOSE):
            result = frasma2.update(val)
            expected = EXPECTED_P5[i]
            if math.isnan(expected):
                self.assertTrue(math.isnan(result), f"[{i}] expected NaN, got {result}")
            else:
                self.assertAlmostEqual(result, expected, places=13,
                                       msg=f"[{i}] expected {expected}, got {result}")

    def test_update_period_10(self):
        frasma2 = FractalAdaptiveSimpleMovingAverage2(
            FractalAdaptiveSimpleMovingAverage2Params(period=10, normal_speed=20))
        for i, val in enumerate(INPUT_CLOSE):
            result = frasma2.update(val)
            expected = EXPECTED_P10[i]
            if math.isnan(expected):
                self.assertTrue(math.isnan(result), f"[{i}] expected NaN, got {result}")
            else:
                self.assertAlmostEqual(result, expected, places=13,
                                       msg=f"[{i}] expected {expected}, got {result}")

    def test_update_period_15(self):
        frasma2 = FractalAdaptiveSimpleMovingAverage2(
            FractalAdaptiveSimpleMovingAverage2Params(period=15, normal_speed=20))
        for i, val in enumerate(INPUT_CLOSE):
            result = frasma2.update(val)
            expected = EXPECTED_P15[i]
            if math.isnan(expected):
                self.assertTrue(math.isnan(result), f"[{i}] expected NaN, got {result}")
            else:
                self.assertAlmostEqual(result, expected, places=13,
                                       msg=f"[{i}] expected {expected}, got {result}")

    def test_update_period_20(self):
        frasma2 = FractalAdaptiveSimpleMovingAverage2(
            FractalAdaptiveSimpleMovingAverage2Params(period=20, normal_speed=20))
        for i, val in enumerate(INPUT_CLOSE):
            result = frasma2.update(val)
            expected = EXPECTED_P20[i]
            if math.isnan(expected):
                self.assertTrue(math.isnan(result), f"[{i}] expected NaN, got {result}")
            else:
                self.assertAlmostEqual(result, expected, places=13,
                                       msg=f"[{i}] expected {expected}, got {result}")

    def test_update_period_30(self):
        frasma2 = FractalAdaptiveSimpleMovingAverage2(
            FractalAdaptiveSimpleMovingAverage2Params(period=30, normal_speed=20))
        for i, val in enumerate(INPUT_CLOSE):
            result = frasma2.update(val)
            expected = EXPECTED_P30[i]
            if math.isnan(expected):
                self.assertTrue(math.isnan(result), f"[{i}] expected NaN, got {result}")
            else:
                self.assertAlmostEqual(result, expected, places=13,
                                       msg=f"[{i}] expected {expected}, got {result}")

    def test_update_period_50(self):
        frasma2 = FractalAdaptiveSimpleMovingAverage2(
            FractalAdaptiveSimpleMovingAverage2Params(period=50, normal_speed=20))
        for i, val in enumerate(INPUT_CLOSE):
            result = frasma2.update(val)
            expected = EXPECTED_P50[i]
            if math.isnan(expected):
                self.assertTrue(math.isnan(result), f"[{i}] expected NaN, got {result}")
            else:
                self.assertAlmostEqual(result, expected, places=13,
                                       msg=f"[{i}] expected {expected}, got {result}")

    def test_update_period_80(self):
        frasma2 = FractalAdaptiveSimpleMovingAverage2(
            FractalAdaptiveSimpleMovingAverage2Params(period=80, normal_speed=20))
        for i, val in enumerate(INPUT_CLOSE):
            result = frasma2.update(val)
            expected = EXPECTED_P80[i]
            if math.isnan(expected):
                self.assertTrue(math.isnan(result), f"[{i}] expected NaN, got {result}")
            else:
                self.assertAlmostEqual(result, expected, places=13,
                                       msg=f"[{i}] expected {expected}, got {result}")

    def test_update_period_120(self):
        frasma2 = FractalAdaptiveSimpleMovingAverage2(
            FractalAdaptiveSimpleMovingAverage2Params(period=120, normal_speed=20))
        for i, val in enumerate(INPUT_CLOSE):
            result = frasma2.update(val)
            expected = EXPECTED_P120[i]
            if math.isnan(expected):
                self.assertTrue(math.isnan(result), f"[{i}] expected NaN, got {result}")
            else:
                self.assertAlmostEqual(result, expected, places=13,
                                       msg=f"[{i}] expected {expected}, got {result}")

    def test_is_primed_period_30(self):
        frasma2 = FractalAdaptiveSimpleMovingAverage2(
            FractalAdaptiveSimpleMovingAverage2Params(period=30, normal_speed=20))
        for i in range(29):
            frasma2.update(INPUT_CLOSE[i])
            self.assertFalse(frasma2.is_primed())
        frasma2.update(INPUT_CLOSE[29])
        self.assertTrue(frasma2.is_primed())

    def test_nan_passthrough(self):
        frasma2 = FractalAdaptiveSimpleMovingAverage2(
            FractalAdaptiveSimpleMovingAverage2Params(period=5, normal_speed=20))
        result = frasma2.update(math.nan)
        self.assertTrue(math.isnan(result))

    def test_invalid_period(self):
        with self.assertRaises(ValueError):
            FractalAdaptiveSimpleMovingAverage2(
                FractalAdaptiveSimpleMovingAverage2Params(period=1, normal_speed=20))

    def test_invalid_normal_speed(self):
        with self.assertRaises(ValueError):
            FractalAdaptiveSimpleMovingAverage2(
                FractalAdaptiveSimpleMovingAverage2Params(period=5, normal_speed=0))

    def test_metadata(self):
        frasma2 = FractalAdaptiveSimpleMovingAverage2(
            FractalAdaptiveSimpleMovingAverage2Params(period=30, normal_speed=20))
        meta = frasma2.metadata()
        self.assertEqual(meta.identifier, Identifier.FRACTAL_ADAPTIVE_SIMPLE_MOVING_AVERAGE_2)
        self.assertIn("frasma2(30,20)", meta.mnemonic)

    def test_update_bar(self):
        frasma2 = FractalAdaptiveSimpleMovingAverage2(
            FractalAdaptiveSimpleMovingAverage2Params(period=5, normal_speed=20))
        for i, val in enumerate(INPUT_CLOSE):
            bar = Bar(datetime(2020, 1, 1), val + 1, val + 2, val - 1, val, 100.0)
            output = frasma2.update_bar(bar)
            expected = EXPECTED_P5[i]
            if math.isnan(expected):
                self.assertTrue(math.isnan(output[0].value))
            else:
                self.assertAlmostEqual(output[0].value, expected, places=13)


if __name__ == '__main__':
    unittest.main()
