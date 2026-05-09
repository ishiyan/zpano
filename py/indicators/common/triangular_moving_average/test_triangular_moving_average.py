import math
import unittest
from datetime import datetime

from py.indicators.common.triangular_moving_average.triangular_moving_average import TriangularMovingAverage
from py.indicators.common.triangular_moving_average.params import TriangularMovingAverageParams
from py.indicators.core.identifier import Identifier
from py.entities.bar import Bar
from py.entities.quote import Quote
from py.entities.trade import Trade
from py.entities.scalar import Scalar

from .test_testdata import INPUT, EXPECTED_XLS_12

class TestTriangularMovingAverage(unittest.TestCase):

    def test_update_length_9(self):
        trima = TriangularMovingAverage(TriangularMovingAverageParams(length=9))
        results = [trima.update(v) for v in INPUT]
        for i in range(8):
            self.assertTrue(math.isnan(results[i]))
        self.assertAlmostEqual(results[8], 93.8176, delta=1e-4)
        self.assertAlmostEqual(results[251], 109.1312, delta=1e-4)

    def test_update_length_10(self):
        trima = TriangularMovingAverage(TriangularMovingAverageParams(length=10))
        results = [trima.update(v) for v in INPUT]
        for i in range(9):
            self.assertTrue(math.isnan(results[i]))
        self.assertAlmostEqual(results[9], 93.6043, delta=1e-4)
        self.assertAlmostEqual(results[10], 93.4252, delta=1e-4)
        self.assertAlmostEqual(results[250], 109.1850, delta=1e-4)
        self.assertAlmostEqual(results[251], 109.1407, delta=1e-4)

    def test_update_length_12(self):
        trima = TriangularMovingAverage(TriangularMovingAverageParams(length=12))
        results = [trima.update(v) for v in INPUT]
        for i in range(11):
            self.assertTrue(math.isnan(results[i]))
        self.assertAlmostEqual(results[11], 93.5329, delta=1e-4)
        self.assertAlmostEqual(results[251], 109.1157, delta=1e-4)

    def test_update_xls_length_12(self):
        trima = TriangularMovingAverage(TriangularMovingAverageParams(length=12))
        for i, val in enumerate(INPUT):
            result = trima.update(val)
            expected = EXPECTED_XLS_12[i]
            if math.isnan(expected):
                self.assertTrue(math.isnan(result))
            else:
                self.assertAlmostEqual(result, expected, delta=1e-12)

    def test_is_primed_length_9(self):
        trima = TriangularMovingAverage(TriangularMovingAverageParams(length=9))
        for i in range(8):
            trima.update(INPUT[i])
            self.assertFalse(trima.is_primed())
        trima.update(INPUT[8])
        self.assertTrue(trima.is_primed())

    def test_is_primed_length_12(self):
        trima = TriangularMovingAverage(TriangularMovingAverageParams(length=12))
        for i in range(11):
            trima.update(INPUT[i])
            self.assertFalse(trima.is_primed())
        trima.update(INPUT[11])
        self.assertTrue(trima.is_primed())

    def test_metadata(self):
        trima = TriangularMovingAverage(TriangularMovingAverageParams(length=5))
        meta = trima.metadata()
        self.assertEqual(meta.identifier, Identifier.TRIANGULAR_MOVING_AVERAGE)
        self.assertEqual(meta.mnemonic, "trima(5)")
        self.assertEqual(meta.description, "Triangular moving average trima(5)")
        self.assertEqual(len(meta.outputs), 1)
        self.assertEqual(meta.outputs[0].mnemonic, "trima(5)")

    def test_construction_errors(self):
        with self.assertRaises(ValueError) as ctx:
            TriangularMovingAverage(TriangularMovingAverageParams(length=1))
        self.assertIn("length should be greater than 1", str(ctx.exception))
        with self.assertRaises(ValueError) as ctx:
            TriangularMovingAverage(TriangularMovingAverageParams(length=0))
        self.assertIn("length should be greater than 1", str(ctx.exception))
        with self.assertRaises(ValueError) as ctx:
            TriangularMovingAverage(TriangularMovingAverageParams(length=-1))
        self.assertIn("length should be greater than 1", str(ctx.exception))

    def test_update_entity(self):
        t = datetime(2021, 4, 1)
        trima = TriangularMovingAverage(TriangularMovingAverageParams(length=12))
        for i in range(11):
            trima.update(INPUT[i])
        output = trima.update_bar(Bar(t, INPUT[11], INPUT[11], INPUT[11], INPUT[11], 1.0))
        self.assertEqual(len(output), 1)
        self.assertIsInstance(output[0], Scalar)
        self.assertEqual(output[0].time, t)
        self.assertAlmostEqual(output[0].value, 93.5329761904762, delta=1e-12)

        # update_scalar
        trima = TriangularMovingAverage(TriangularMovingAverageParams(length=12))
        for i in range(11):
            trima.update(INPUT[i])
        output = trima.update_scalar(Scalar(t, INPUT[11]))
        self.assertEqual(len(output), 1)
        self.assertIsInstance(output[0], Scalar)
        self.assertEqual(output[0].time, t)
        self.assertAlmostEqual(output[0].value, 93.5329761904762, delta=1e-12)

        # update_quote
        trima = TriangularMovingAverage(TriangularMovingAverageParams(length=12))
        for i in range(11):
            trima.update(INPUT[i])
        output = trima.update_quote(Quote(t, INPUT[11], INPUT[11], 1.0, 1.0))
        self.assertEqual(len(output), 1)
        self.assertIsInstance(output[0], Scalar)
        self.assertEqual(output[0].time, t)
        self.assertAlmostEqual(output[0].value, 93.5329761904762, delta=1e-12)

        # update_trade
        trima = TriangularMovingAverage(TriangularMovingAverageParams(length=12))
        for i in range(11):
            trima.update(INPUT[i])
        output = trima.update_trade(Trade(t, INPUT[11], 1.0))
        self.assertEqual(len(output), 1)
        self.assertIsInstance(output[0], Scalar)
        self.assertEqual(output[0].time, t)
        self.assertAlmostEqual(output[0].value, 93.5329761904762, delta=1e-12)


if __name__ == "__main__":
    unittest.main()
