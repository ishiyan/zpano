import math
import unittest
from datetime import datetime

from py.indicators.common.exponential_moving_average.exponential_moving_average import ExponentialMovingAverage
from py.indicators.common.exponential_moving_average.params import ExponentialMovingAverageLengthParams, ExponentialMovingAverageSmoothingFactorParams
from py.indicators.core.identifier import Identifier
from py.entities.bar import Bar
from py.entities.quote import Quote
from py.entities.trade import Trade
from py.entities.scalar import Scalar

from .test_testdata import INPUT

class TestExponentialMovingAverage(unittest.TestCase):

    def test_update_length_2_first_is_average_true(self):
        ema = ExponentialMovingAverage.from_length(ExponentialMovingAverageLengthParams(length=2, first_is_average=True))
        results = [ema.update(v) for v in INPUT]
        self.assertTrue(math.isnan(results[0]))
        self.assertAlmostEqual(results[1], 93.15, delta=1e-2)
        self.assertAlmostEqual(results[2], 93.96, delta=1e-2)
        self.assertAlmostEqual(results[3], 94.71, delta=1e-2)
        self.assertAlmostEqual(results[251], 108.21, delta=1e-2)

    def test_update_length_10_first_is_average_true(self):
        ema = ExponentialMovingAverage.from_length(ExponentialMovingAverageLengthParams(length=10, first_is_average=True))
        results = [ema.update(v) for v in INPUT]
        for i in range(9):
            self.assertTrue(math.isnan(results[i]))
        self.assertAlmostEqual(results[9], 93.22, delta=1e-2)
        self.assertAlmostEqual(results[10], 93.75, delta=1e-2)
        self.assertAlmostEqual(results[29], 86.46, delta=1e-2)
        self.assertAlmostEqual(results[251], 108.97, delta=1e-2)

    def test_update_length_2_first_is_average_false(self):
        ema = ExponentialMovingAverage.from_length(ExponentialMovingAverageLengthParams(length=2, first_is_average=False))
        results = [ema.update(v) for v in INPUT]
        self.assertTrue(math.isnan(results[0]))
        self.assertAlmostEqual(results[1], 93.71, delta=1e-2)
        self.assertAlmostEqual(results[2], 94.15, delta=1e-2)
        self.assertAlmostEqual(results[3], 94.78, delta=1e-2)
        self.assertAlmostEqual(results[251], 108.21, delta=1e-2)

    def test_update_length_10_first_is_average_false(self):
        ema = ExponentialMovingAverage.from_length(ExponentialMovingAverageLengthParams(length=10, first_is_average=False))
        results = [ema.update(v) for v in INPUT]
        for i in range(9):
            self.assertTrue(math.isnan(results[i]))
        self.assertAlmostEqual(results[9], 92.60, delta=1e-2)
        self.assertAlmostEqual(results[10], 93.24, delta=1e-2)
        self.assertAlmostEqual(results[11], 93.97, delta=1e-2)
        self.assertAlmostEqual(results[30], 86.23, delta=1e-2)
        self.assertAlmostEqual(results[251], 108.97, delta=1e-2)

    def test_is_primed_length_10_first_is_average_true(self):
        ema = ExponentialMovingAverage.from_length(ExponentialMovingAverageLengthParams(length=10, first_is_average=True))
        for i in range(9):
            ema.update(INPUT[i])
            self.assertFalse(ema.is_primed())
        ema.update(INPUT[9])
        self.assertTrue(ema.is_primed())

    def test_is_primed_length_10_first_is_average_false(self):
        ema = ExponentialMovingAverage.from_length(ExponentialMovingAverageLengthParams(length=10, first_is_average=False))
        for i in range(9):
            ema.update(INPUT[i])
            self.assertFalse(ema.is_primed())
        ema.update(INPUT[9])
        self.assertTrue(ema.is_primed())

    def test_metadata_length(self):
        ema = ExponentialMovingAverage.from_length(ExponentialMovingAverageLengthParams(length=10, first_is_average=True))
        meta = ema.metadata()
        self.assertEqual(meta.identifier, Identifier.EXPONENTIAL_MOVING_AVERAGE)
        self.assertEqual(meta.mnemonic, "ema(10)")
        self.assertEqual(meta.description, "Exponential moving average ema(10)")
        self.assertEqual(len(meta.outputs), 1)
        self.assertEqual(meta.outputs[0].mnemonic, "ema(10)")

    def test_metadata_alpha(self):
        ema = ExponentialMovingAverage.from_smoothing_factor(ExponentialMovingAverageSmoothingFactorParams(smoothing_factor=2/11, first_is_average=False))
        meta = ema.metadata()
        self.assertEqual(meta.identifier, Identifier.EXPONENTIAL_MOVING_AVERAGE)
        self.assertEqual(meta.mnemonic, "ema(10, 0.18181818)")
        self.assertEqual(meta.description, "Exponential moving average ema(10, 0.18181818)")

    def test_construction_errors(self):
        with self.assertRaises(ValueError) as ctx:
            ExponentialMovingAverage.from_length(ExponentialMovingAverageLengthParams(length=0, first_is_average=True))
        self.assertIn("length should be positive", str(ctx.exception))
        with self.assertRaises(ValueError) as ctx:
            ExponentialMovingAverage.from_length(ExponentialMovingAverageLengthParams(length=-1, first_is_average=True))
        self.assertIn("length should be positive", str(ctx.exception))
        with self.assertRaises(ValueError) as ctx:
            ExponentialMovingAverage.from_smoothing_factor(ExponentialMovingAverageSmoothingFactorParams(smoothing_factor=-1, first_is_average=True))
        self.assertIn("smoothing factor should be in range [0, 1]", str(ctx.exception))
        with self.assertRaises(ValueError) as ctx:
            ExponentialMovingAverage.from_smoothing_factor(ExponentialMovingAverageSmoothingFactorParams(smoothing_factor=2, first_is_average=True))
        self.assertIn("smoothing factor should be in range [0, 1]", str(ctx.exception))

    def test_update_entity(self):
        t = datetime(2021, 4, 1)

        # update_bar
        ema = ExponentialMovingAverage.from_length(ExponentialMovingAverageLengthParams(length=2, first_is_average=False))
        ema.update(0.0)
        ema.update(0.0)
        output = ema.update_bar(Bar(t, 3.0, 3.0, 3.0, 3.0, 1.0))
        self.assertEqual(len(output), 1)
        self.assertIsInstance(output[0], Scalar)
        self.assertEqual(output[0].time, t)
        self.assertAlmostEqual(output[0].value, 2.0, delta=1e-2)

        # update_scalar
        ema = ExponentialMovingAverage.from_length(ExponentialMovingAverageLengthParams(length=2, first_is_average=False))
        ema.update(0.0)
        ema.update(0.0)
        output = ema.update_scalar(Scalar(t, 3.0))
        self.assertEqual(len(output), 1)
        self.assertIsInstance(output[0], Scalar)
        self.assertEqual(output[0].time, t)
        self.assertAlmostEqual(output[0].value, 2.0, delta=1e-2)

        # update_quote
        ema = ExponentialMovingAverage.from_length(ExponentialMovingAverageLengthParams(length=2, first_is_average=False))
        ema.update(0.0)
        ema.update(0.0)
        output = ema.update_quote(Quote(t, 3.0, 3.0, 1.0, 1.0))
        self.assertEqual(len(output), 1)
        self.assertIsInstance(output[0], Scalar)
        self.assertEqual(output[0].time, t)
        self.assertAlmostEqual(output[0].value, 2.0, delta=1e-2)

        # update_trade
        ema = ExponentialMovingAverage.from_length(ExponentialMovingAverageLengthParams(length=2, first_is_average=False))
        ema.update(0.0)
        ema.update(0.0)
        output = ema.update_trade(Trade(t, 3.0, 1.0))
        self.assertEqual(len(output), 1)
        self.assertIsInstance(output[0], Scalar)
        self.assertEqual(output[0].time, t)
        self.assertAlmostEqual(output[0].value, 2.0, delta=1e-2)


if __name__ == "__main__":
    unittest.main()
