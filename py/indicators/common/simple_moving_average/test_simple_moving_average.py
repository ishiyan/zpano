import math
import unittest
from datetime import datetime

from py.indicators.common.simple_moving_average.simple_moving_average import SimpleMovingAverage
from py.indicators.common.simple_moving_average.params import SimpleMovingAverageParams
from py.indicators.core.identifier import Identifier
from py.entities.bar import Bar
from py.entities.quote import Quote
from py.entities.trade import Trade
from py.entities.scalar import Scalar

INPUT = [
    64.59, 64.23, 65.26, 65.24, 65.07, 65.14, 64.98, 64.76, 65.11, 65.46,
    65.94, 66.10, 66.87, 66.56, 66.71, 66.19, 66.14, 66.64, 67.33, 68.18,
    67.48, 67.19, 66.46, 67.20, 67.62, 67.66, 67.89, 69.19, 69.68, 69.31,
    69.11, 69.27, 68.97, 69.11, 69.50, 69.70, 69.94, 69.11, 67.64, 67.75,
    67.47, 67.50, 68.18, 67.35, 66.74, 67.00, 67.46, 67.36, 67.37, 67.78,
    67.96,
]

EXPECTED_3 = [
    math.nan, math.nan, 64.69, 64.91, 65.19, 65.15, 65.06, 64.96, 64.95, 65.11,
    65.50, 65.83, 66.30, 66.51, 66.71, 66.49, 66.35, 66.32, 66.70, 67.38,
    67.66, 67.62, 67.04, 66.95, 67.09, 67.49, 67.72, 68.25, 68.92, 69.39,
    69.37, 69.23, 69.12, 69.12, 69.19, 69.44, 69.71, 69.58, 68.90, 68.17,
    67.62, 67.57, 67.72, 67.68, 67.42, 67.03, 67.07, 67.27, 67.40, 67.50,
    67.70,
]

EXPECTED_5 = [
    math.nan, math.nan, math.nan, math.nan, 64.88, 64.99, 65.14, 65.04, 65.01, 65.09,
    65.25, 65.47, 65.90, 66.19, 66.44, 66.49, 66.49, 66.45, 66.60, 66.90,
    67.15, 67.36, 67.33, 67.30, 67.19, 67.23, 67.37, 67.91, 68.41, 68.75,
    69.04, 69.31, 69.27, 69.15, 69.19, 69.31, 69.44, 69.47, 69.18, 68.83,
    68.38, 67.89, 67.71, 67.65, 67.45, 67.35, 67.35, 67.18, 67.19, 67.39,
    67.59,
]

EXPECTED_10 = [
    math.nan, math.nan, math.nan, math.nan, math.nan, math.nan, math.nan, math.nan, math.nan, 64.98,
    65.12, 65.31, 65.47, 65.60, 65.76, 65.87, 65.98, 66.17, 66.39, 66.67,
    66.82, 66.93, 66.89, 66.95, 67.04, 67.19, 67.37, 67.62, 67.86, 67.97,
    68.13, 68.34, 68.59, 68.78, 68.97, 69.17, 69.38, 69.37, 69.17, 69.01,
    68.85, 68.67, 68.59, 68.41, 68.14, 67.87, 67.62, 67.45, 67.42, 67.42,
    67.47,
]


class TestSimpleMovingAverage(unittest.TestCase):

    def test_update_length_3(self):
        sma = SimpleMovingAverage(SimpleMovingAverageParams(length=3))
        for i, val in enumerate(INPUT):
            result = sma.update(val)
            expected = EXPECTED_3[i]
            if math.isnan(expected):
                self.assertTrue(math.isnan(result))
            else:
                self.assertAlmostEqual(result, expected, delta=1e-2)
        # NaN input returns NaN
        result = sma.update(math.nan)
        self.assertTrue(math.isnan(result))

    def test_update_length_5(self):
        sma = SimpleMovingAverage(SimpleMovingAverageParams(length=5))
        for i, val in enumerate(INPUT):
            result = sma.update(val)
            expected = EXPECTED_5[i]
            if math.isnan(expected):
                self.assertTrue(math.isnan(result))
            else:
                self.assertAlmostEqual(result, expected, delta=1e-2)

    def test_update_length_10(self):
        sma = SimpleMovingAverage(SimpleMovingAverageParams(length=10))
        for i, val in enumerate(INPUT):
            result = sma.update(val)
            expected = EXPECTED_10[i]
            if math.isnan(expected):
                self.assertTrue(math.isnan(result))
            else:
                self.assertAlmostEqual(result, expected, delta=1e-2)

    def test_is_primed_length_3(self):
        sma = SimpleMovingAverage(SimpleMovingAverageParams(length=3))
        for i in range(2):
            sma.update(INPUT[i])
            self.assertFalse(sma.is_primed())
        sma.update(INPUT[2])
        self.assertTrue(sma.is_primed())

    def test_is_primed_length_5(self):
        sma = SimpleMovingAverage(SimpleMovingAverageParams(length=5))
        for i in range(4):
            sma.update(INPUT[i])
            self.assertFalse(sma.is_primed())
        sma.update(INPUT[4])
        self.assertTrue(sma.is_primed())

    def test_is_primed_length_10(self):
        sma = SimpleMovingAverage(SimpleMovingAverageParams(length=10))
        for i in range(9):
            sma.update(INPUT[i])
            self.assertFalse(sma.is_primed())
        sma.update(INPUT[9])
        self.assertTrue(sma.is_primed())

    def test_metadata(self):
        sma = SimpleMovingAverage(SimpleMovingAverageParams(length=5))
        meta = sma.metadata()
        self.assertEqual(meta.identifier, Identifier.SIMPLE_MOVING_AVERAGE)
        self.assertEqual(meta.mnemonic, "sma(5)")
        self.assertEqual(meta.description, "Simple moving average sma(5)")
        self.assertEqual(len(meta.outputs), 1)
        self.assertEqual(meta.outputs[0].mnemonic, "sma(5)")

    def test_construction_errors(self):
        with self.assertRaises(ValueError) as ctx:
            SimpleMovingAverage(SimpleMovingAverageParams(length=1))
        self.assertIn("length should be greater than 1", str(ctx.exception))
        with self.assertRaises(ValueError) as ctx:
            SimpleMovingAverage(SimpleMovingAverageParams(length=0))
        self.assertIn("length should be greater than 1", str(ctx.exception))
        with self.assertRaises(ValueError) as ctx:
            SimpleMovingAverage(SimpleMovingAverageParams(length=-1))
        self.assertIn("length should be greater than 1", str(ctx.exception))

    def test_update_entity(self):
        t = datetime(2021, 4, 1)

        # update_bar
        sma = SimpleMovingAverage(SimpleMovingAverageParams(length=2))
        sma.update(0.0)
        output = sma.update_bar(Bar(t, 3.0, 3.0, 3.0, 3.0, 1.0))
        self.assertEqual(len(output), 1)
        self.assertIsInstance(output[0], Scalar)
        self.assertEqual(output[0].time, t)
        self.assertAlmostEqual(output[0].value, 1.5, delta=1e-2)

        # update_scalar
        sma = SimpleMovingAverage(SimpleMovingAverageParams(length=2))
        sma.update(0.0)
        output = sma.update_scalar(Scalar(t, 3.0))
        self.assertEqual(len(output), 1)
        self.assertIsInstance(output[0], Scalar)
        self.assertEqual(output[0].time, t)
        self.assertAlmostEqual(output[0].value, 1.5, delta=1e-2)

        # update_quote
        sma = SimpleMovingAverage(SimpleMovingAverageParams(length=2))
        sma.update(0.0)
        output = sma.update_quote(Quote(t, 3.0, 3.0, 1.0, 1.0))
        self.assertEqual(len(output), 1)
        self.assertIsInstance(output[0], Scalar)
        self.assertEqual(output[0].time, t)
        self.assertAlmostEqual(output[0].value, 1.5, delta=1e-2)

        # update_trade
        sma = SimpleMovingAverage(SimpleMovingAverageParams(length=2))
        sma.update(0.0)
        output = sma.update_trade(Trade(t, 3.0, 1.0))
        self.assertEqual(len(output), 1)
        self.assertIsInstance(output[0], Scalar)
        self.assertEqual(output[0].time, t)
        self.assertAlmostEqual(output[0].value, 1.5, delta=1e-2)


if __name__ == "__main__":
    unittest.main()
