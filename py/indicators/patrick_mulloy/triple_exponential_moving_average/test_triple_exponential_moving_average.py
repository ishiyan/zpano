import math
import unittest
from datetime import datetime

from py.indicators.patrick_mulloy.triple_exponential_moving_average.triple_exponential_moving_average import TripleExponentialMovingAverage
from py.indicators.patrick_mulloy.triple_exponential_moving_average.params import (
    TripleExponentialMovingAverageLengthParams,
    TripleExponentialMovingAverageSmoothingFactorParams,
)
from py.indicators.core.identifier import Identifier
from py.entities.bar import Bar
from py.entities.quote import Quote
from py.entities.trade import Trade
from py.entities.scalar import Scalar

from .test_testdata import INPUT, TASC_INPUT, TEMA_TASC_EXPECTED

class TestTripleExponentialMovingAverage(unittest.TestCase):

    def test_update_length_14_first_is_average_true(self):
        tema = TripleExponentialMovingAverage.from_length(
            TripleExponentialMovingAverageLengthParams(length=14, first_is_average=True))
        lprimed = 3 * 14 - 3
        results = [tema.update(v) for v in INPUT]
        for i in range(lprimed):
            self.assertTrue(math.isnan(results[i]), f"[{i}] expected NaN")
        self.assertAlmostEqual(results[39], 84.8629, delta=1e-3)
        self.assertAlmostEqual(results[40], 84.2246, delta=1e-3)
        self.assertAlmostEqual(results[251], 108.418, delta=1e-3)

    def test_update_length_14_first_is_average_false(self):
        tema = TripleExponentialMovingAverage.from_length(
            TripleExponentialMovingAverageLengthParams(length=14, first_is_average=False))
        lprimed = 3 * 14 - 3
        results = [tema.update(v) for v in INPUT]
        for i in range(lprimed):
            self.assertTrue(math.isnan(results[i]), f"[{i}] expected NaN")
        self.assertAlmostEqual(results[39], 84.721, delta=1e-3)
        self.assertAlmostEqual(results[40], 84.089, delta=1e-3)
        self.assertAlmostEqual(results[251], 108.418, delta=1e-3)

    def test_update_length_26_tasc_first_is_average_false(self):
        tema = TripleExponentialMovingAverage.from_length(
            TripleExponentialMovingAverageLengthParams(length=26, first_is_average=False))
        lprimed = 3 * 26 - 3
        first_check = 216

        results = [tema.update(v) for v in TASC_INPUT]

        for i in range(lprimed):
            self.assertTrue(math.isnan(results[i]), f"[{i}] expected NaN")

        for i in range(first_check, len(TASC_INPUT)):
            self.assertAlmostEqual(results[i], TEMA_TASC_EXPECTED[i], delta=1e-3,
                                   msg=f"[{i}]")

    def test_is_primed_length_14(self):
        lprimed = 3 * 14 - 3
        for fia in [True, False]:
            with self.subTest(first_is_average=fia):
                tema = TripleExponentialMovingAverage.from_length(
                    TripleExponentialMovingAverageLengthParams(length=14, first_is_average=fia))
                self.assertFalse(tema.is_primed())
                for i in range(lprimed):
                    tema.update(INPUT[i])
                    self.assertFalse(tema.is_primed(), f"[{i}]")
                tema.update(INPUT[lprimed])
                self.assertTrue(tema.is_primed())

    def test_metadata_length(self):
        tema = TripleExponentialMovingAverage.from_length(
            TripleExponentialMovingAverageLengthParams(length=10, first_is_average=True))
        meta = tema.metadata()
        self.assertEqual(meta.identifier, Identifier.TRIPLE_EXPONENTIAL_MOVING_AVERAGE)
        self.assertEqual(meta.mnemonic, "tema(10)")
        self.assertEqual(meta.description, "Triple exponential moving average tema(10)")
        self.assertEqual(len(meta.outputs), 1)
        self.assertEqual(meta.outputs[0].mnemonic, "tema(10)")

    def test_metadata_alpha(self):
        tema = TripleExponentialMovingAverage.from_smoothing_factor(
            TripleExponentialMovingAverageSmoothingFactorParams(smoothing_factor=2/11, first_is_average=False))
        meta = tema.metadata()
        self.assertEqual(meta.identifier, Identifier.TRIPLE_EXPONENTIAL_MOVING_AVERAGE)
        self.assertEqual(meta.mnemonic, "tema(10, 0.18181818)")
        self.assertEqual(meta.description, "Triple exponential moving average tema(10, 0.18181818)")

    def test_construction_errors(self):
        with self.assertRaises(ValueError):
            TripleExponentialMovingAverage.from_length(
                TripleExponentialMovingAverageLengthParams(length=1))
        with self.assertRaises(ValueError):
            TripleExponentialMovingAverage.from_length(
                TripleExponentialMovingAverageLengthParams(length=0))
        with self.assertRaises(ValueError):
            TripleExponentialMovingAverage.from_length(
                TripleExponentialMovingAverageLengthParams(length=-1))
        with self.assertRaises(ValueError):
            TripleExponentialMovingAverage.from_smoothing_factor(
                TripleExponentialMovingAverageSmoothingFactorParams(smoothing_factor=-1))
        with self.assertRaises(ValueError):
            TripleExponentialMovingAverage.from_smoothing_factor(
                TripleExponentialMovingAverageSmoothingFactorParams(smoothing_factor=2))

    def test_update_entity(self):
        t = datetime(2021, 4, 1)
        l = 2
        lprimed = 3 * l - 3
        inp = 3.0
        exp_false = 2.888888888888889
        exp_true = 2.6666666666666665

        # update_scalar (false)
        tema = TripleExponentialMovingAverage.from_length(
            TripleExponentialMovingAverageLengthParams(length=l, first_is_average=False))
        for _ in range(lprimed):
            tema.update(0.0)
        output = tema.update_scalar(Scalar(t, inp))
        self.assertEqual(len(output), 1)
        self.assertIsInstance(output[0], Scalar)
        self.assertEqual(output[0].time, t)
        self.assertAlmostEqual(output[0].value, exp_false, delta=1e-13)

        # update_bar (true)
        tema = TripleExponentialMovingAverage.from_length(
            TripleExponentialMovingAverageLengthParams(length=l, first_is_average=True))
        for _ in range(lprimed):
            tema.update(0.0)
        output = tema.update_bar(Bar(t, inp, inp, inp, inp, 1.0))
        self.assertEqual(len(output), 1)
        self.assertAlmostEqual(output[0].value, exp_true, delta=1e-13)

        # update_quote (false)
        tema = TripleExponentialMovingAverage.from_length(
            TripleExponentialMovingAverageLengthParams(length=l, first_is_average=False))
        for _ in range(lprimed):
            tema.update(0.0)
        output = tema.update_quote(Quote(t, inp, inp, 1.0, 1.0))
        self.assertEqual(len(output), 1)
        self.assertAlmostEqual(output[0].value, exp_false, delta=1e-13)

        # update_trade (true)
        tema = TripleExponentialMovingAverage.from_length(
            TripleExponentialMovingAverageLengthParams(length=l, first_is_average=True))
        for _ in range(lprimed):
            tema.update(0.0)
        output = tema.update_trade(Trade(t, inp, 1.0))
        self.assertEqual(len(output), 1)
        self.assertAlmostEqual(output[0].value, exp_true, delta=1e-13)

    def test_nan_passthrough(self):
        tema = TripleExponentialMovingAverage.from_length(
            TripleExponentialMovingAverageLengthParams(length=2, first_is_average=True))
        for v in INPUT:
            tema.update(v)
        self.assertTrue(math.isnan(tema.update(math.nan)))


if __name__ == "__main__":
    unittest.main()
