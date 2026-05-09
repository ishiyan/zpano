import math
import unittest
from datetime import datetime

from py.indicators.tim_tillson.t2_exponential_moving_average.t2_exponential_moving_average import T2ExponentialMovingAverage
from py.indicators.tim_tillson.t2_exponential_moving_average.params import (
    T2ExponentialMovingAverageLengthParams,
    T2ExponentialMovingAverageSmoothingFactorParams,
)
from py.indicators.core.identifier import Identifier
from py.entities.bar import Bar
from py.entities.quote import Quote
from py.entities.trade import Trade
from py.entities.scalar import Scalar

from .test_testdata import INPUT, T2_EXPECTED

# Expected data from modified TA-Lib test_T3.xls: test_T2.xls, T2(5,0.7) column.
# First 16 values are NaN (lprimed = 4*5 - 4 = 16), then 236 floats.
class TestT2ExponentialMovingAverage(unittest.TestCase):

    def test_update_length_5_first_is_average_true(self):
        l = 5
        lprimed = 4 * l - 4
        t2 = T2ExponentialMovingAverage.from_length(
            T2ExponentialMovingAverageLengthParams(length=l, volume_factor=0.7, first_is_average=True))
        results = [t2.update(v) for v in INPUT]
        for i in range(lprimed):
            self.assertTrue(math.isnan(results[i]), f"[{i}] expected NaN")
        for i in range(lprimed, len(INPUT)):
            self.assertAlmostEqual(results[i], T2_EXPECTED[i], delta=1e-8, msg=f"[{i}]")

    def test_update_length_5_first_is_average_false(self):
        l = 5
        lprimed = 4 * l - 4
        first_check = lprimed + 43
        t2 = T2ExponentialMovingAverage.from_length(
            T2ExponentialMovingAverageLengthParams(length=l, volume_factor=0.7, first_is_average=False))
        results = [t2.update(v) for v in INPUT]
        for i in range(lprimed):
            self.assertTrue(math.isnan(results[i]), f"[{i}] expected NaN")
        for i in range(first_check, len(INPUT)):
            self.assertAlmostEqual(results[i], T2_EXPECTED[i], delta=1e-8, msg=f"[{i}]")

    def test_is_primed_length_5(self):
        l = 5
        lprimed = 4 * l - 4
        for fia in [True, False]:
            with self.subTest(first_is_average=fia):
                t2 = T2ExponentialMovingAverage.from_length(
                    T2ExponentialMovingAverageLengthParams(length=l, volume_factor=0.7, first_is_average=fia))
                self.assertFalse(t2.is_primed())
                for i in range(lprimed):
                    t2.update(INPUT[i])
                    self.assertFalse(t2.is_primed(), f"[{i}]")
                t2.update(INPUT[lprimed])
                self.assertTrue(t2.is_primed())

    def test_metadata_length(self):
        t2 = T2ExponentialMovingAverage.from_length(
            T2ExponentialMovingAverageLengthParams(length=10, volume_factor=0.3333, first_is_average=True))
        meta = t2.metadata()
        self.assertEqual(meta.identifier, Identifier.T2_EXPONENTIAL_MOVING_AVERAGE)
        self.assertEqual(meta.mnemonic, "t2(10, 0.33330000)")
        self.assertEqual(meta.description, "T2 exponential moving average t2(10, 0.33330000)")
        self.assertEqual(len(meta.outputs), 1)
        self.assertEqual(meta.outputs[0].mnemonic, "t2(10, 0.33330000)")

    def test_metadata_alpha(self):
        t2 = T2ExponentialMovingAverage.from_smoothing_factor(
            T2ExponentialMovingAverageSmoothingFactorParams(smoothing_factor=2/11, volume_factor=0.3333333, first_is_average=False))
        meta = t2.metadata()
        self.assertEqual(meta.identifier, Identifier.T2_EXPONENTIAL_MOVING_AVERAGE)
        self.assertEqual(meta.mnemonic, "t2(10, 0.18181818, 0.33333330)")
        self.assertEqual(meta.description, "T2 exponential moving average t2(10, 0.18181818, 0.33333330)")

    def test_construction_errors(self):
        with self.assertRaises(ValueError):
            T2ExponentialMovingAverage.from_length(
                T2ExponentialMovingAverageLengthParams(length=1))
        with self.assertRaises(ValueError):
            T2ExponentialMovingAverage.from_length(
                T2ExponentialMovingAverageLengthParams(length=0))
        with self.assertRaises(ValueError):
            T2ExponentialMovingAverage.from_length(
                T2ExponentialMovingAverageLengthParams(length=-1))
        with self.assertRaises(ValueError):
            T2ExponentialMovingAverage.from_smoothing_factor(
                T2ExponentialMovingAverageSmoothingFactorParams(smoothing_factor=-1))
        with self.assertRaises(ValueError):
            T2ExponentialMovingAverage.from_smoothing_factor(
                T2ExponentialMovingAverageSmoothingFactorParams(smoothing_factor=2))

    def test_update_entity(self):
        t = datetime(2021, 4, 1)
        l = 2
        lprimed = 4 * l - 4
        inp = 3.0
        exp_false = 2.0281481481481483
        exp_true = 1.9555555555555555

        # update_scalar (false)
        t2 = T2ExponentialMovingAverage.from_length(
            T2ExponentialMovingAverageLengthParams(length=l, volume_factor=0.7, first_is_average=False))
        for _ in range(lprimed):
            t2.update(0.0)
        output = t2.update_scalar(Scalar(t, inp))
        self.assertEqual(len(output), 1)
        self.assertIsInstance(output[0], Scalar)
        self.assertEqual(output[0].time, t)
        self.assertAlmostEqual(output[0].value, exp_false, delta=1e-13)

        # update_bar (true)
        t2 = T2ExponentialMovingAverage.from_length(
            T2ExponentialMovingAverageLengthParams(length=l, volume_factor=0.7, first_is_average=True))
        for _ in range(lprimed):
            t2.update(0.0)
        output = t2.update_bar(Bar(t, inp, inp, inp, inp, 1.0))
        self.assertEqual(len(output), 1)
        self.assertAlmostEqual(output[0].value, exp_true, delta=1e-13)

        # update_quote (false)
        t2 = T2ExponentialMovingAverage.from_length(
            T2ExponentialMovingAverageLengthParams(length=l, volume_factor=0.7, first_is_average=False))
        for _ in range(lprimed):
            t2.update(0.0)
        output = t2.update_quote(Quote(t, inp, inp, 1.0, 1.0))
        self.assertEqual(len(output), 1)
        self.assertAlmostEqual(output[0].value, exp_false, delta=1e-13)

        # update_trade (true)
        t2 = T2ExponentialMovingAverage.from_length(
            T2ExponentialMovingAverageLengthParams(length=l, volume_factor=0.7, first_is_average=True))
        for _ in range(lprimed):
            t2.update(0.0)
        output = t2.update_trade(Trade(t, inp, 1.0))
        self.assertEqual(len(output), 1)
        self.assertAlmostEqual(output[0].value, exp_true, delta=1e-13)

    def test_nan_passthrough(self):
        t2 = T2ExponentialMovingAverage.from_length(
            T2ExponentialMovingAverageLengthParams(length=5, volume_factor=0.7, first_is_average=True))
        for v in INPUT:
            t2.update(v)
        self.assertTrue(math.isnan(t2.update(math.nan)))


if __name__ == "__main__":
    unittest.main()
