import math
import unittest
from datetime import datetime

from py.indicators.common.weighted_moving_average.weighted_moving_average import WeightedMovingAverage
from py.indicators.common.weighted_moving_average.params import WeightedMovingAverageParams
from py.indicators.core.identifier import Identifier
from py.entities.bar import Bar
from py.entities.quote import Quote
from py.entities.trade import Trade
from py.entities.scalar import Scalar

from .test_testdata import INPUT

class TestWeightedMovingAverage(unittest.TestCase):

    def test_update_length_2(self):
        wma = WeightedMovingAverage(WeightedMovingAverageParams(length=2))
        results = [wma.update(v) for v in INPUT]
        self.assertTrue(math.isnan(results[0]))
        self.assertAlmostEqual(results[1], 93.71, delta=1e-2)
        self.assertAlmostEqual(results[2], 94.52, delta=1e-2)
        self.assertAlmostEqual(results[3], 94.855, delta=1e-2)
        self.assertAlmostEqual(results[251], 108.16, delta=1e-2)

    def test_update_length_30(self):
        wma = WeightedMovingAverage(WeightedMovingAverageParams(length=30))
        results = [wma.update(v) for v in INPUT]
        for i in range(29):
            self.assertTrue(math.isnan(results[i]))
        self.assertAlmostEqual(results[29], 88.5677, delta=1e-2)
        self.assertAlmostEqual(results[30], 88.2337, delta=1e-2)
        self.assertAlmostEqual(results[31], 88.034, delta=1e-2)
        self.assertAlmostEqual(results[58], 87.191, delta=1e-2)
        self.assertAlmostEqual(results[250], 109.3466, delta=1e-2)
        self.assertAlmostEqual(results[251], 109.3413, delta=1e-2)

    def test_is_primed_length_2(self):
        wma = WeightedMovingAverage(WeightedMovingAverageParams(length=2))
        wma.update(INPUT[0])
        self.assertFalse(wma.is_primed())
        wma.update(INPUT[1])
        self.assertTrue(wma.is_primed())

    def test_is_primed_length_30(self):
        wma = WeightedMovingAverage(WeightedMovingAverageParams(length=30))
        for i in range(29):
            wma.update(INPUT[i])
            self.assertFalse(wma.is_primed())
        wma.update(INPUT[29])
        self.assertTrue(wma.is_primed())

    def test_metadata(self):
        wma = WeightedMovingAverage(WeightedMovingAverageParams(length=5))
        meta = wma.metadata()
        self.assertEqual(meta.identifier, Identifier.WEIGHTED_MOVING_AVERAGE)
        self.assertEqual(meta.mnemonic, "wma(5)")
        self.assertEqual(meta.description, "Weighted moving average wma(5)")
        self.assertEqual(len(meta.outputs), 1)
        self.assertEqual(meta.outputs[0].mnemonic, "wma(5)")

    def test_construction_errors(self):
        with self.assertRaises(ValueError) as ctx:
            WeightedMovingAverage(WeightedMovingAverageParams(length=1))
        self.assertIn("length should be greater than 1", str(ctx.exception))
        with self.assertRaises(ValueError) as ctx:
            WeightedMovingAverage(WeightedMovingAverageParams(length=0))
        self.assertIn("length should be greater than 1", str(ctx.exception))
        with self.assertRaises(ValueError) as ctx:
            WeightedMovingAverage(WeightedMovingAverageParams(length=-1))
        self.assertIn("length should be greater than 1", str(ctx.exception))

    def test_update_entity(self):
        t = datetime(2021, 4, 1)

        # update_bar
        wma = WeightedMovingAverage(WeightedMovingAverageParams(length=2))
        wma.update(INPUT[0])
        output = wma.update_bar(Bar(t, INPUT[1], INPUT[1], INPUT[1], INPUT[1], 1.0))
        self.assertEqual(len(output), 1)
        self.assertIsInstance(output[0], Scalar)
        self.assertEqual(output[0].time, t)
        self.assertAlmostEqual(output[0].value, 93.71, delta=1e-2)

        # update_scalar
        wma = WeightedMovingAverage(WeightedMovingAverageParams(length=2))
        wma.update(INPUT[0])
        output = wma.update_scalar(Scalar(t, INPUT[1]))
        self.assertEqual(len(output), 1)
        self.assertIsInstance(output[0], Scalar)
        self.assertEqual(output[0].time, t)
        self.assertAlmostEqual(output[0].value, 93.71, delta=1e-2)

        # update_quote
        wma = WeightedMovingAverage(WeightedMovingAverageParams(length=2))
        wma.update(INPUT[0])
        output = wma.update_quote(Quote(t, INPUT[1], INPUT[1], 1.0, 1.0))
        self.assertEqual(len(output), 1)
        self.assertIsInstance(output[0], Scalar)
        self.assertEqual(output[0].time, t)
        self.assertAlmostEqual(output[0].value, 93.71, delta=1e-2)

        # update_trade
        wma = WeightedMovingAverage(WeightedMovingAverageParams(length=2))
        wma.update(INPUT[0])
        output = wma.update_trade(Trade(t, INPUT[1], 1.0))
        self.assertEqual(len(output), 1)
        self.assertIsInstance(output[0], Scalar)
        self.assertEqual(output[0].time, t)
        self.assertAlmostEqual(output[0].value, 93.71, delta=1e-2)


if __name__ == "__main__":
    unittest.main()
