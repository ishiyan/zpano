import math
import unittest
from datetime import datetime

from py.indicators.tim_tillson.t3_exponential_moving_average.t3_exponential_moving_average import T3ExponentialMovingAverage
from py.indicators.tim_tillson.t3_exponential_moving_average.params import (
    T3ExponentialMovingAverageLengthParams,
    T3ExponentialMovingAverageSmoothingFactorParams,
)
from py.indicators.core.identifier import Identifier
from py.entities.bar import Bar
from py.entities.quote import Quote
from py.entities.trade import Trade
from py.entities.scalar import Scalar

from .test_testdata import INPUT, T3_EXPECTED

# Expected data from Go reference test: T3(5, 0.7).
# First 24 values are NaN (lprimed = 6*5 - 6 = 24), then 228 floats.
class TestT3ExponentialMovingAverage(unittest.TestCase):

    def test_update_length_5_first_is_average_true(self):
        l = 5
        lprimed = 6 * l - 6
        t3 = T3ExponentialMovingAverage.from_length(
            T3ExponentialMovingAverageLengthParams(length=l, volume_factor=0.7, first_is_average=True))
        results = [t3.update(v) for v in INPUT]
        for i in range(lprimed):
            self.assertTrue(math.isnan(results[i]), f"[{i}] expected NaN")
        for i in range(lprimed, len(INPUT)):
            self.assertAlmostEqual(results[i], T3_EXPECTED[i], delta=1e-3, msg=f"[{i}]")
        # Spot checks
        self.assertAlmostEqual(results[250], 109.03210341275800, delta=1e-3)
        self.assertAlmostEqual(results[251], 108.87915000449300, delta=1e-3)

    def test_update_length_5_first_is_average_false(self):
        l = 5
        lprimed = 6 * l - 6
        first_check = lprimed + 63  # convergence point
        t3 = T3ExponentialMovingAverage.from_length(
            T3ExponentialMovingAverageLengthParams(length=l, volume_factor=0.7, first_is_average=False))
        results = [t3.update(v) for v in INPUT]
        for i in range(lprimed):
            self.assertTrue(math.isnan(results[i]), f"[{i}] expected NaN")
        for i in range(first_check, len(INPUT)):
            self.assertAlmostEqual(results[i], T3_EXPECTED[i], delta=1e-3, msg=f"[{i}]")
        # Spot checks
        self.assertAlmostEqual(results[250], 109.03210341275800, delta=1e-3)
        self.assertAlmostEqual(results[251], 108.87915000449300, delta=1e-3)

    def test_is_primed_length_5(self):
        l = 5
        lprimed = 6 * l - 6
        for fia in [True, False]:
            with self.subTest(first_is_average=fia):
                t3 = T3ExponentialMovingAverage.from_length(
                    T3ExponentialMovingAverageLengthParams(length=l, volume_factor=0.7, first_is_average=fia))
                self.assertFalse(t3.is_primed())
                for i in range(lprimed):
                    t3.update(INPUT[i])
                    self.assertFalse(t3.is_primed(), f"[{i}]")
                t3.update(INPUT[lprimed])
                self.assertTrue(t3.is_primed())

    def test_metadata_length(self):
        t3 = T3ExponentialMovingAverage.from_length(
            T3ExponentialMovingAverageLengthParams(length=10, volume_factor=0.3333, first_is_average=True))
        meta = t3.metadata()
        self.assertEqual(meta.identifier, Identifier.T3_EXPONENTIAL_MOVING_AVERAGE)
        self.assertEqual(meta.mnemonic, "t3(10, 0.33330000)")
        self.assertEqual(meta.description, "T3 exponential moving average t3(10, 0.33330000)")
        self.assertEqual(len(meta.outputs), 1)
        self.assertEqual(meta.outputs[0].mnemonic, "t3(10, 0.33330000)")

    def test_metadata_alpha(self):
        t3 = T3ExponentialMovingAverage.from_smoothing_factor(
            T3ExponentialMovingAverageSmoothingFactorParams(smoothing_factor=2/11, volume_factor=0.3333333, first_is_average=False))
        meta = t3.metadata()
        self.assertEqual(meta.identifier, Identifier.T3_EXPONENTIAL_MOVING_AVERAGE)
        self.assertEqual(meta.mnemonic, "t3(10, 0.18181818, 0.33333330)")
        self.assertEqual(meta.description, "T3 exponential moving average t3(10, 0.18181818, 0.33333330)")

    def test_construction_errors(self):
        with self.assertRaises(ValueError):
            T3ExponentialMovingAverage.from_length(
                T3ExponentialMovingAverageLengthParams(length=1))
        with self.assertRaises(ValueError):
            T3ExponentialMovingAverage.from_length(
                T3ExponentialMovingAverageLengthParams(length=0))
        with self.assertRaises(ValueError):
            T3ExponentialMovingAverage.from_length(
                T3ExponentialMovingAverageLengthParams(length=-1))
        with self.assertRaises(ValueError):
            T3ExponentialMovingAverage.from_smoothing_factor(
                T3ExponentialMovingAverageSmoothingFactorParams(smoothing_factor=-1))
        with self.assertRaises(ValueError):
            T3ExponentialMovingAverage.from_smoothing_factor(
                T3ExponentialMovingAverageSmoothingFactorParams(smoothing_factor=2))

    def test_update_entity(self):
        t = datetime(2021, 4, 1)
        l = 2
        lprimed = 6 * l - 6
        inp = 3.0
        exp_false = 1.6675884773662544
        exp_true = 1.6901728395061721

        # update_scalar (false)
        t3 = T3ExponentialMovingAverage.from_length(
            T3ExponentialMovingAverageLengthParams(length=l, volume_factor=0.7, first_is_average=False))
        for _ in range(lprimed):
            t3.update(0.0)
        output = t3.update_scalar(Scalar(t, inp))
        self.assertEqual(len(output), 1)
        self.assertIsInstance(output[0], Scalar)
        self.assertEqual(output[0].time, t)
        self.assertAlmostEqual(output[0].value, exp_false, delta=1e-13)

        # update_bar (true)
        t3 = T3ExponentialMovingAverage.from_length(
            T3ExponentialMovingAverageLengthParams(length=l, volume_factor=0.7, first_is_average=True))
        for _ in range(lprimed):
            t3.update(0.0)
        output = t3.update_bar(Bar(t, inp, inp, inp, inp, 1.0))
        self.assertEqual(len(output), 1)
        self.assertAlmostEqual(output[0].value, exp_true, delta=1e-13)

        # update_quote (false)
        t3 = T3ExponentialMovingAverage.from_length(
            T3ExponentialMovingAverageLengthParams(length=l, volume_factor=0.7, first_is_average=False))
        for _ in range(lprimed):
            t3.update(0.0)
        output = t3.update_quote(Quote(t, inp, inp, 1.0, 1.0))
        self.assertEqual(len(output), 1)
        self.assertAlmostEqual(output[0].value, exp_false, delta=1e-13)

        # update_trade (true)
        t3 = T3ExponentialMovingAverage.from_length(
            T3ExponentialMovingAverageLengthParams(length=l, volume_factor=0.7, first_is_average=True))
        for _ in range(lprimed):
            t3.update(0.0)
        output = t3.update_trade(Trade(t, inp, 1.0))
        self.assertEqual(len(output), 1)
        self.assertAlmostEqual(output[0].value, exp_true, delta=1e-13)

    def test_nan_passthrough(self):
        t3 = T3ExponentialMovingAverage.from_length(
            T3ExponentialMovingAverageLengthParams(length=5, volume_factor=0.7, first_is_average=True))
        for v in INPUT:
            t3.update(v)
        self.assertTrue(math.isnan(t3.update(math.nan)))


if __name__ == "__main__":
    unittest.main()
