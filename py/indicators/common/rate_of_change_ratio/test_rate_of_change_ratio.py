import math
import unittest
from datetime import datetime

from py.indicators.common.rate_of_change_ratio.rate_of_change_ratio import RateOfChangeRatio
from py.indicators.common.rate_of_change_ratio.params import RateOfChangeRatioParams
from py.indicators.core.identifier import Identifier
from py.entities.bar import Bar
from py.entities.quote import Quote
from py.entities.trade import Trade
from py.entities.scalar import Scalar

from .test_testdata import INPUT

class TestRateOfChangeRatio(unittest.TestCase):

    def test_rocr_length_14(self):
        rocr = RateOfChangeRatio(RateOfChangeRatioParams(length=14, hundred_scale=False))
        results = [rocr.update(v) for v in INPUT]
        for i in range(14):
            self.assertTrue(math.isnan(results[i]), f"[{i}] expected NaN")
        self.assertAlmostEqual(results[14], 0.994536, delta=1e-4)
        self.assertAlmostEqual(results[15], 0.978906, delta=1e-4)
        self.assertAlmostEqual(results[16], 0.944689, delta=1e-4)
        self.assertAlmostEqual(results[251], 0.989633, delta=1e-4)

    def test_rocr100_length_14(self):
        rocr100 = RateOfChangeRatio(RateOfChangeRatioParams(length=14, hundred_scale=True))
        results = [rocr100.update(v) for v in INPUT]
        for i in range(14):
            self.assertTrue(math.isnan(results[i]), f"[{i}] expected NaN")
        self.assertAlmostEqual(results[14], 99.4536, delta=1e-4)
        self.assertAlmostEqual(results[15], 97.8906, delta=1e-4)
        self.assertAlmostEqual(results[16], 94.4689, delta=1e-4)
        self.assertAlmostEqual(results[251], 98.9633, delta=1e-4)

    def test_rocr_middle_of_data(self):
        rocr = RateOfChangeRatio(RateOfChangeRatioParams(length=14, hundred_scale=False))
        results = [rocr.update(v) for v in INPUT]
        self.assertAlmostEqual(results[20], 0.955096, delta=1e-4)
        self.assertAlmostEqual(results[21], 0.944744, delta=1e-4)

    def test_rocr100_middle_of_data(self):
        rocr100 = RateOfChangeRatio(RateOfChangeRatioParams(length=14, hundred_scale=True))
        results = [rocr100.update(v) for v in INPUT]
        self.assertAlmostEqual(results[20], 95.5096, delta=1e-4)
        self.assertAlmostEqual(results[21], 94.4744, delta=1e-4)

    def test_nan_passthrough(self):
        rocr = RateOfChangeRatio(RateOfChangeRatioParams(length=14))
        for v in INPUT:
            rocr.update(v)
        self.assertTrue(math.isnan(rocr.update(math.nan)))

    def test_is_primed_length_1(self):
        rocr = RateOfChangeRatio(RateOfChangeRatioParams(length=1))
        self.assertFalse(rocr.is_primed())
        rocr.update(INPUT[0])
        self.assertFalse(rocr.is_primed())
        rocr.update(INPUT[1])
        self.assertTrue(rocr.is_primed())

    def test_is_primed_length_2(self):
        rocr = RateOfChangeRatio(RateOfChangeRatioParams(length=2))
        self.assertFalse(rocr.is_primed())
        for i in range(2):
            rocr.update(INPUT[i])
            self.assertFalse(rocr.is_primed())
        rocr.update(INPUT[2])
        self.assertTrue(rocr.is_primed())

    def test_is_primed_length_5(self):
        rocr = RateOfChangeRatio(RateOfChangeRatioParams(length=5))
        self.assertFalse(rocr.is_primed())
        for i in range(5):
            rocr.update(INPUT[i])
            self.assertFalse(rocr.is_primed())
        rocr.update(INPUT[5])
        self.assertTrue(rocr.is_primed())

    def test_is_primed_length_10(self):
        rocr = RateOfChangeRatio(RateOfChangeRatioParams(length=10))
        self.assertFalse(rocr.is_primed())
        for i in range(10):
            rocr.update(INPUT[i])
            self.assertFalse(rocr.is_primed())
        rocr.update(INPUT[10])
        self.assertTrue(rocr.is_primed())

    def test_metadata_rocr(self):
        rocr = RateOfChangeRatio(RateOfChangeRatioParams(length=5))
        meta = rocr.metadata()
        self.assertEqual(meta.identifier, Identifier.RATE_OF_CHANGE_RATIO)
        self.assertEqual(meta.mnemonic, "rocr(5)")
        self.assertEqual(meta.description, "Rate of Change Ratio rocr(5)")
        self.assertEqual(len(meta.outputs), 1)
        self.assertEqual(meta.outputs[0].mnemonic, "rocr(5)")

    def test_metadata_rocr100(self):
        rocr100 = RateOfChangeRatio(RateOfChangeRatioParams(length=5, hundred_scale=True))
        meta = rocr100.metadata()
        self.assertEqual(meta.identifier, Identifier.RATE_OF_CHANGE_RATIO)
        self.assertEqual(meta.mnemonic, "rocr100(5)")
        self.assertEqual(meta.description, "Rate of Change Ratio 100 Scale rocr100(5)")

    def test_construction_errors(self):
        with self.assertRaises(ValueError) as ctx:
            RateOfChangeRatio(RateOfChangeRatioParams(length=0))
        self.assertIn("length should be positive", str(ctx.exception))
        with self.assertRaises(ValueError) as ctx:
            RateOfChangeRatio(RateOfChangeRatioParams(length=-1))
        self.assertIn("length should be positive", str(ctx.exception))

    def test_update_entity(self):
        t = datetime(2021, 4, 1)
        inp = 3.0
        exp = 1.0  # ROCR: 3/3 = 1

        # update_bar
        rocr = RateOfChangeRatio(RateOfChangeRatioParams(length=2))
        rocr.update(inp)
        rocr.update(inp)
        output = rocr.update_bar(Bar(t, inp, inp, inp, inp, 1.0))
        self.assertEqual(len(output), 1)
        self.assertIsInstance(output[0], Scalar)
        self.assertAlmostEqual(output[0].value, exp, delta=1e-13)

        # update_scalar
        rocr = RateOfChangeRatio(RateOfChangeRatioParams(length=2))
        rocr.update(inp)
        rocr.update(inp)
        output = rocr.update_scalar(Scalar(t, inp))
        self.assertEqual(len(output), 1)
        self.assertAlmostEqual(output[0].value, exp, delta=1e-13)

        # update_quote
        rocr = RateOfChangeRatio(RateOfChangeRatioParams(length=2))
        rocr.update(inp)
        rocr.update(inp)
        output = rocr.update_quote(Quote(t, inp, inp, 1.0, 1.0))
        self.assertEqual(len(output), 1)
        self.assertAlmostEqual(output[0].value, exp, delta=1e-13)

        # update_trade
        rocr = RateOfChangeRatio(RateOfChangeRatioParams(length=2))
        rocr.update(inp)
        rocr.update(inp)
        output = rocr.update_trade(Trade(t, inp, 1.0))
        self.assertEqual(len(output), 1)
        self.assertAlmostEqual(output[0].value, exp, delta=1e-13)


if __name__ == "__main__":
    unittest.main()
