import math
import unittest
from datetime import datetime

from py.indicators.patrick_mulloy.double_exponential_moving_average.double_exponential_moving_average import DoubleExponentialMovingAverage
from py.indicators.patrick_mulloy.double_exponential_moving_average.params import (
    DoubleExponentialMovingAverageLengthParams,
    DoubleExponentialMovingAverageSmoothingFactorParams,
)
from py.indicators.core.identifier import Identifier
from py.entities.bar import Bar
from py.entities.quote import Quote
from py.entities.trade import Trade
from py.entities.scalar import Scalar

from .test_testdata import INPUT, TASC_INPUT, DEMA_TASC_EXPECTED

class TestDoubleExponentialMovingAverage(unittest.TestCase):

    def test_update_length_2_first_is_average_true(self):
        dema = DoubleExponentialMovingAverage.from_length(
            DoubleExponentialMovingAverageLengthParams(length=2, first_is_average=True))
        results = [dema.update(v) for v in INPUT]
        self.assertTrue(math.isnan(results[0]))
        self.assertTrue(math.isnan(results[1]))
        self.assertAlmostEqual(results[4], 94.013, delta=1e-2)
        self.assertAlmostEqual(results[5], 94.539, delta=1e-2)
        self.assertAlmostEqual(results[251], 107.94, delta=1e-2)

    def test_update_length_14_first_is_average_true(self):
        dema = DoubleExponentialMovingAverage.from_length(
            DoubleExponentialMovingAverageLengthParams(length=14, first_is_average=True))
        results = [dema.update(v) for v in INPUT]
        lprimed = 2 * 14 - 2
        for i in range(lprimed):
            self.assertTrue(math.isnan(results[i]), f"[{i}] expected NaN")
        self.assertAlmostEqual(results[28], 84.347, delta=1e-2)
        self.assertAlmostEqual(results[29], 84.487, delta=1e-2)
        self.assertAlmostEqual(results[30], 84.374, delta=1e-2)
        self.assertAlmostEqual(results[31], 84.772, delta=1e-2)
        self.assertAlmostEqual(results[48], 89.803, delta=1e-2)
        self.assertAlmostEqual(results[251], 109.4676, delta=1e-2)

    def test_update_length_2_first_is_average_false(self):
        dema = DoubleExponentialMovingAverage.from_length(
            DoubleExponentialMovingAverageLengthParams(length=2, first_is_average=False))
        results = [dema.update(v) for v in INPUT]
        self.assertTrue(math.isnan(results[0]))
        self.assertTrue(math.isnan(results[1]))
        self.assertAlmostEqual(results[4], 93.977, delta=1e-2)
        self.assertAlmostEqual(results[5], 94.522, delta=1e-2)
        self.assertAlmostEqual(results[251], 107.94, delta=1e-2)

    def test_update_length_14_first_is_average_false(self):
        dema = DoubleExponentialMovingAverage.from_length(
            DoubleExponentialMovingAverageLengthParams(length=14, first_is_average=False))
        results = [dema.update(v) for v in INPUT]
        lprimed = 2 * 14 - 2
        for i in range(lprimed):
            self.assertTrue(math.isnan(results[i]), f"[{i}] expected NaN")
        self.assertAlmostEqual(results[28], 84.87, delta=1e-2)
        self.assertAlmostEqual(results[29], 84.94, delta=1e-2)
        self.assertAlmostEqual(results[30], 84.77, delta=1e-2)
        self.assertAlmostEqual(results[31], 85.12, delta=1e-2)
        self.assertAlmostEqual(results[48], 89.83, delta=1e-2)
        self.assertAlmostEqual(results[251], 109.4676, delta=1e-2)

    def test_update_length_26_tasc_first_is_average_false(self):
        dema = DoubleExponentialMovingAverage.from_length(
            DoubleExponentialMovingAverageLengthParams(length=26, first_is_average=False))
        lprimed = 2 * 26 - 2
        first_check = 216

        results = [dema.update(v) for v in TASC_INPUT]

        for i in range(lprimed):
            self.assertTrue(math.isnan(results[i]), f"[{i}] expected NaN")

        for i in range(first_check, len(TASC_INPUT)):
            self.assertAlmostEqual(results[i], DEMA_TASC_EXPECTED[i], delta=1e-2,
                                   msg=f"[{i}]")

    def test_is_primed_length_14(self):
        lprimed = 2 * 14 - 2
        for fia in [True, False]:
            with self.subTest(first_is_average=fia):
                dema = DoubleExponentialMovingAverage.from_length(
                    DoubleExponentialMovingAverageLengthParams(length=14, first_is_average=fia))
                self.assertFalse(dema.is_primed())
                for i in range(lprimed):
                    dema.update(INPUT[i])
                    self.assertFalse(dema.is_primed(), f"[{i}]")
                dema.update(INPUT[lprimed])
                self.assertTrue(dema.is_primed())

    def test_metadata_length(self):
        dema = DoubleExponentialMovingAverage.from_length(
            DoubleExponentialMovingAverageLengthParams(length=10, first_is_average=True))
        meta = dema.metadata()
        self.assertEqual(meta.identifier, Identifier.DOUBLE_EXPONENTIAL_MOVING_AVERAGE)
        self.assertEqual(meta.mnemonic, "dema(10)")
        self.assertEqual(meta.description, "Double exponential moving average dema(10)")
        self.assertEqual(len(meta.outputs), 1)
        self.assertEqual(meta.outputs[0].mnemonic, "dema(10)")

    def test_metadata_alpha(self):
        dema = DoubleExponentialMovingAverage.from_smoothing_factor(
            DoubleExponentialMovingAverageSmoothingFactorParams(smoothing_factor=2/11, first_is_average=False))
        meta = dema.metadata()
        self.assertEqual(meta.identifier, Identifier.DOUBLE_EXPONENTIAL_MOVING_AVERAGE)
        self.assertEqual(meta.mnemonic, "dema(10, 0.18181818)")
        self.assertEqual(meta.description, "Double exponential moving average dema(10, 0.18181818)")

    def test_construction_errors(self):
        with self.assertRaises(ValueError):
            DoubleExponentialMovingAverage.from_length(
                DoubleExponentialMovingAverageLengthParams(length=0))
        with self.assertRaises(ValueError):
            DoubleExponentialMovingAverage.from_length(
                DoubleExponentialMovingAverageLengthParams(length=-1))
        with self.assertRaises(ValueError):
            DoubleExponentialMovingAverage.from_smoothing_factor(
                DoubleExponentialMovingAverageSmoothingFactorParams(smoothing_factor=-1))
        with self.assertRaises(ValueError):
            DoubleExponentialMovingAverage.from_smoothing_factor(
                DoubleExponentialMovingAverageSmoothingFactorParams(smoothing_factor=2))

    def test_update_entity(self):
        t = datetime(2021, 4, 1)
        l = 2
        lprimed = 2 * l - 2
        inp = 3.0
        exp_false = 2.666666666666667

        # update_scalar
        dema = DoubleExponentialMovingAverage.from_length(
            DoubleExponentialMovingAverageLengthParams(length=l, first_is_average=False))
        for _ in range(lprimed):
            dema.update(0.0)
        output = dema.update_scalar(Scalar(t, inp))
        self.assertEqual(len(output), 1)
        self.assertIsInstance(output[0], Scalar)
        self.assertEqual(output[0].time, t)
        self.assertAlmostEqual(output[0].value, exp_false, delta=1e-13)

        # update_bar
        dema = DoubleExponentialMovingAverage.from_length(
            DoubleExponentialMovingAverageLengthParams(length=l, first_is_average=False))
        for _ in range(lprimed):
            dema.update(0.0)
        output = dema.update_bar(Bar(t, inp, inp, inp, inp, 1.0))
        self.assertEqual(len(output), 1)
        self.assertAlmostEqual(output[0].value, exp_false, delta=1e-13)

        # update_quote
        dema = DoubleExponentialMovingAverage.from_length(
            DoubleExponentialMovingAverageLengthParams(length=l, first_is_average=False))
        for _ in range(lprimed):
            dema.update(0.0)
        output = dema.update_quote(Quote(t, inp, inp, 1.0, 1.0))
        self.assertEqual(len(output), 1)
        self.assertAlmostEqual(output[0].value, exp_false, delta=1e-13)

        # update_trade
        dema = DoubleExponentialMovingAverage.from_length(
            DoubleExponentialMovingAverageLengthParams(length=l, first_is_average=False))
        for _ in range(lprimed):
            dema.update(0.0)
        output = dema.update_trade(Trade(t, inp, 1.0))
        self.assertEqual(len(output), 1)
        self.assertAlmostEqual(output[0].value, exp_false, delta=1e-13)

    def test_nan_passthrough(self):
        dema = DoubleExponentialMovingAverage.from_length(
            DoubleExponentialMovingAverageLengthParams(length=2, first_is_average=True))
        for v in INPUT:
            dema.update(v)
        self.assertTrue(math.isnan(dema.update(math.nan)))


if __name__ == "__main__":
    unittest.main()
