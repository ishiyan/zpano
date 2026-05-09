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

from .test_testdata import (
    INPUT,
    EXPECTED_3,
    EXPECTED_5,
    EXPECTED_10,
)

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
