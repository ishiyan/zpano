import math
import unittest
from datetime import datetime

from py.indicators.common.rate_of_change_percent.rate_of_change_percent import RateOfChangePercent
from py.indicators.common.rate_of_change_percent.params import RateOfChangePercentParams
from py.indicators.core.identifier import Identifier
from py.entities.bar import Bar
from py.entities.quote import Quote
from py.entities.trade import Trade
from py.entities.scalar import Scalar

from .test_testdata import INPUT

class TestRateOfChangePercent(unittest.TestCase):

    def test_update_length_14(self):
        rocp = RateOfChangePercent(RateOfChangePercentParams(length=14))
        results = [rocp.update(v) for v in INPUT]
        for i in range(14):
            self.assertTrue(math.isnan(results[i]), f"[{i}] expected NaN")
        self.assertAlmostEqual(results[14], -0.00546, delta=1e-4)
        self.assertAlmostEqual(results[15], -0.02109, delta=1e-4)
        self.assertAlmostEqual(results[16], -0.0553, delta=1e-4)
        self.assertAlmostEqual(results[251], -0.010367, delta=1e-4)

    def test_nan_passthrough(self):
        rocp = RateOfChangePercent(RateOfChangePercentParams(length=14))
        for v in INPUT:
            rocp.update(v)
        result = rocp.update(math.nan)
        self.assertTrue(math.isnan(result))

    def test_is_primed_length_1(self):
        rocp = RateOfChangePercent(RateOfChangePercentParams(length=1))
        self.assertFalse(rocp.is_primed())
        rocp.update(INPUT[0])
        self.assertFalse(rocp.is_primed())
        rocp.update(INPUT[1])
        self.assertTrue(rocp.is_primed())

    def test_is_primed_length_2(self):
        rocp = RateOfChangePercent(RateOfChangePercentParams(length=2))
        self.assertFalse(rocp.is_primed())
        for i in range(2):
            rocp.update(INPUT[i])
            self.assertFalse(rocp.is_primed())
        rocp.update(INPUT[2])
        self.assertTrue(rocp.is_primed())

    def test_is_primed_length_5(self):
        rocp = RateOfChangePercent(RateOfChangePercentParams(length=5))
        self.assertFalse(rocp.is_primed())
        for i in range(5):
            rocp.update(INPUT[i])
            self.assertFalse(rocp.is_primed())
        rocp.update(INPUT[5])
        self.assertTrue(rocp.is_primed())

    def test_is_primed_length_10(self):
        rocp = RateOfChangePercent(RateOfChangePercentParams(length=10))
        self.assertFalse(rocp.is_primed())
        for i in range(10):
            rocp.update(INPUT[i])
            self.assertFalse(rocp.is_primed())
        rocp.update(INPUT[10])
        self.assertTrue(rocp.is_primed())

    def test_metadata(self):
        rocp = RateOfChangePercent(RateOfChangePercentParams(length=5))
        meta = rocp.metadata()
        self.assertEqual(meta.identifier, Identifier.RATE_OF_CHANGE_PERCENT)
        self.assertEqual(meta.mnemonic, "rocp(5)")
        self.assertEqual(meta.description, "Rate of Change Percent rocp(5)")
        self.assertEqual(len(meta.outputs), 1)
        self.assertEqual(meta.outputs[0].mnemonic, "rocp(5)")

    def test_construction_errors(self):
        with self.assertRaises(ValueError) as ctx:
            RateOfChangePercent(RateOfChangePercentParams(length=0))
        self.assertIn("length should be positive", str(ctx.exception))
        with self.assertRaises(ValueError) as ctx:
            RateOfChangePercent(RateOfChangePercentParams(length=-1))
        self.assertIn("length should be positive", str(ctx.exception))

    def test_update_entity(self):
        t = datetime(2021, 4, 1)
        inp = 3.0
        exp = 0.0

        # update_bar
        rocp = RateOfChangePercent(RateOfChangePercentParams(length=2))
        rocp.update(inp)
        rocp.update(inp)
        output = rocp.update_bar(Bar(t, inp, inp, inp, inp, 1.0))
        self.assertEqual(len(output), 1)
        self.assertIsInstance(output[0], Scalar)
        self.assertAlmostEqual(output[0].value, exp, delta=1e-13)

        # update_scalar
        rocp = RateOfChangePercent(RateOfChangePercentParams(length=2))
        rocp.update(inp)
        rocp.update(inp)
        output = rocp.update_scalar(Scalar(t, inp))
        self.assertEqual(len(output), 1)
        self.assertAlmostEqual(output[0].value, exp, delta=1e-13)

        # update_quote
        rocp = RateOfChangePercent(RateOfChangePercentParams(length=2))
        rocp.update(inp)
        rocp.update(inp)
        output = rocp.update_quote(Quote(t, inp, inp, 1.0, 1.0))
        self.assertEqual(len(output), 1)
        self.assertAlmostEqual(output[0].value, exp, delta=1e-13)

        # update_trade
        rocp = RateOfChangePercent(RateOfChangePercentParams(length=2))
        rocp.update(inp)
        rocp.update(inp)
        output = rocp.update_trade(Trade(t, inp, 1.0))
        self.assertEqual(len(output), 1)
        self.assertAlmostEqual(output[0].value, exp, delta=1e-13)


if __name__ == "__main__":
    unittest.main()
