import math
import unittest
from datetime import datetime

from py.indicators.common.rate_of_change.rate_of_change import RateOfChange
from py.indicators.common.rate_of_change.params import RateOfChangeParams
from py.indicators.core.identifier import Identifier
from py.entities.bar import Bar
from py.entities.quote import Quote
from py.entities.trade import Trade
from py.entities.scalar import Scalar

from .test_testdata import INPUT

class TestRateOfChange(unittest.TestCase):

    def test_update_length_14(self):
        roc = RateOfChange(RateOfChangeParams(length=14))
        results = [roc.update(v) for v in INPUT]
        for i in range(14):
            self.assertTrue(math.isnan(results[i]), f"[{i}] expected NaN")
        self.assertAlmostEqual(results[14], -0.546, delta=1e-2)
        self.assertAlmostEqual(results[15], -2.109, delta=1e-2)
        self.assertAlmostEqual(results[16], -5.53, delta=1e-2)
        self.assertAlmostEqual(results[251], -1.0367, delta=1e-2)

    def test_update_middle_values(self):
        roc = RateOfChange(RateOfChangeParams(length=14))
        results = [roc.update(v) for v in INPUT]
        self.assertAlmostEqual(results[20], -4.49, delta=1e-2)
        self.assertAlmostEqual(results[21], -5.5256, delta=1e-2)

    def test_nan_passthrough(self):
        roc = RateOfChange(RateOfChangeParams(length=14))
        for v in INPUT:
            roc.update(v)
        result = roc.update(math.nan)
        self.assertTrue(math.isnan(result))

    def test_is_primed_length_1(self):
        roc = RateOfChange(RateOfChangeParams(length=1))
        self.assertFalse(roc.is_primed())
        roc.update(INPUT[0])
        self.assertFalse(roc.is_primed())
        roc.update(INPUT[1])
        self.assertTrue(roc.is_primed())

    def test_is_primed_length_2(self):
        roc = RateOfChange(RateOfChangeParams(length=2))
        self.assertFalse(roc.is_primed())
        for i in range(2):
            roc.update(INPUT[i])
            self.assertFalse(roc.is_primed())
        roc.update(INPUT[2])
        self.assertTrue(roc.is_primed())

    def test_is_primed_length_5(self):
        roc = RateOfChange(RateOfChangeParams(length=5))
        self.assertFalse(roc.is_primed())
        for i in range(5):
            roc.update(INPUT[i])
            self.assertFalse(roc.is_primed())
        roc.update(INPUT[5])
        self.assertTrue(roc.is_primed())

    def test_is_primed_length_10(self):
        roc = RateOfChange(RateOfChangeParams(length=10))
        self.assertFalse(roc.is_primed())
        for i in range(10):
            roc.update(INPUT[i])
            self.assertFalse(roc.is_primed())
        roc.update(INPUT[10])
        self.assertTrue(roc.is_primed())

    def test_metadata(self):
        roc = RateOfChange(RateOfChangeParams(length=5))
        meta = roc.metadata()
        self.assertEqual(meta.identifier, Identifier.RATE_OF_CHANGE)
        self.assertEqual(meta.mnemonic, "roc(5)")
        self.assertEqual(meta.description, "Rate of Change roc(5)")
        self.assertEqual(len(meta.outputs), 1)
        self.assertEqual(meta.outputs[0].mnemonic, "roc(5)")

    def test_construction_errors(self):
        with self.assertRaises(ValueError) as ctx:
            RateOfChange(RateOfChangeParams(length=0))
        self.assertIn("length should be positive", str(ctx.exception))
        with self.assertRaises(ValueError) as ctx:
            RateOfChange(RateOfChangeParams(length=-1))
        self.assertIn("length should be positive", str(ctx.exception))

    def test_update_entity(self):
        t = datetime(2021, 4, 1)
        inp = 3.0
        exp = 0.0

        # update_bar
        roc = RateOfChange(RateOfChangeParams(length=2))
        roc.update(inp)
        roc.update(inp)
        output = roc.update_bar(Bar(t, inp, inp, inp, inp, 1.0))
        self.assertEqual(len(output), 1)
        self.assertIsInstance(output[0], Scalar)
        self.assertEqual(output[0].time, t)
        self.assertAlmostEqual(output[0].value, exp, delta=1e-13)

        # update_scalar
        roc = RateOfChange(RateOfChangeParams(length=2))
        roc.update(inp)
        roc.update(inp)
        output = roc.update_scalar(Scalar(t, inp))
        self.assertEqual(len(output), 1)
        self.assertIsInstance(output[0], Scalar)
        self.assertAlmostEqual(output[0].value, exp, delta=1e-13)

        # update_quote
        roc = RateOfChange(RateOfChangeParams(length=2))
        roc.update(inp)
        roc.update(inp)
        output = roc.update_quote(Quote(t, inp, inp, 1.0, 1.0))
        self.assertEqual(len(output), 1)
        self.assertIsInstance(output[0], Scalar)
        self.assertAlmostEqual(output[0].value, exp, delta=1e-13)

        # update_trade
        roc = RateOfChange(RateOfChangeParams(length=2))
        roc.update(inp)
        roc.update(inp)
        output = roc.update_trade(Trade(t, inp, 1.0))
        self.assertEqual(len(output), 1)
        self.assertIsInstance(output[0], Scalar)
        self.assertAlmostEqual(output[0].value, exp, delta=1e-13)


if __name__ == "__main__":
    unittest.main()
