import math
import unittest
from datetime import datetime

from py.indicators.common.momentum.momentum import Momentum
from py.indicators.common.momentum.params import MomentumParams
from py.indicators.core.identifier import Identifier
from py.entities.bar import Bar
from py.entities.quote import Quote
from py.entities.trade import Trade
from py.entities.scalar import Scalar

from .test_testdata import INPUT

class TestMomentum(unittest.TestCase):

    def test_update_length_14(self):
        mom = Momentum(MomentumParams(length=14))
        results = [mom.update(v) for v in INPUT]
        for i in range(14):
            self.assertTrue(math.isnan(results[i]))
        self.assertAlmostEqual(results[14], -0.50, places=13)
        self.assertAlmostEqual(results[15], -2.00, places=13)
        self.assertAlmostEqual(results[16], -5.22, places=13)
        self.assertAlmostEqual(results[251], -1.13, places=13)

    def test_is_primed_length_1(self):
        mom = Momentum(MomentumParams(length=1))
        mom.update(INPUT[0])
        self.assertFalse(mom.is_primed())
        mom.update(INPUT[1])
        self.assertTrue(mom.is_primed())

    def test_is_primed_length_2(self):
        mom = Momentum(MomentumParams(length=2))
        for i in range(2):
            mom.update(INPUT[i])
            self.assertFalse(mom.is_primed())
        mom.update(INPUT[2])
        self.assertTrue(mom.is_primed())

    def test_is_primed_length_5(self):
        mom = Momentum(MomentumParams(length=5))
        for i in range(5):
            mom.update(INPUT[i])
            self.assertFalse(mom.is_primed())
        mom.update(INPUT[5])
        self.assertTrue(mom.is_primed())

    def test_is_primed_length_10(self):
        mom = Momentum(MomentumParams(length=10))
        for i in range(10):
            mom.update(INPUT[i])
            self.assertFalse(mom.is_primed())
        mom.update(INPUT[10])
        self.assertTrue(mom.is_primed())

    def test_metadata(self):
        mom = Momentum(MomentumParams(length=5))
        meta = mom.metadata()
        self.assertEqual(meta.identifier, Identifier.MOMENTUM)
        self.assertEqual(meta.mnemonic, "mom(5)")
        self.assertEqual(meta.description, "Momentum mom(5)")
        self.assertEqual(len(meta.outputs), 1)
        self.assertEqual(meta.outputs[0].mnemonic, "mom(5)")

    def test_construction_errors(self):
        with self.assertRaises(ValueError) as ctx:
            Momentum(MomentumParams(length=0))
        self.assertIn("length should be positive", str(ctx.exception))
        with self.assertRaises(ValueError) as ctx:
            Momentum(MomentumParams(length=-1))
        self.assertIn("length should be positive", str(ctx.exception))

    def test_update_entity(self):
        t = datetime(2021, 4, 1)

        # update_bar
        mom = Momentum(MomentumParams(length=2))
        mom.update(0.0)
        mom.update(0.0)
        output = mom.update_bar(Bar(t, 3.0, 3.0, 3.0, 3.0, 1.0))
        self.assertEqual(len(output), 1)
        self.assertIsInstance(output[0], Scalar)
        self.assertEqual(output[0].time, t)
        self.assertAlmostEqual(output[0].value, 3.0, places=13)

        # update_scalar
        mom = Momentum(MomentumParams(length=2))
        mom.update(0.0)
        mom.update(0.0)
        output = mom.update_scalar(Scalar(t, 3.0))
        self.assertEqual(len(output), 1)
        self.assertIsInstance(output[0], Scalar)
        self.assertEqual(output[0].time, t)
        self.assertAlmostEqual(output[0].value, 3.0, places=13)

        # update_quote
        mom = Momentum(MomentumParams(length=2))
        mom.update(0.0)
        mom.update(0.0)
        output = mom.update_quote(Quote(t, 3.0, 3.0, 1.0, 1.0))
        self.assertEqual(len(output), 1)
        self.assertIsInstance(output[0], Scalar)
        self.assertEqual(output[0].time, t)
        self.assertAlmostEqual(output[0].value, 3.0, places=13)

        # update_trade
        mom = Momentum(MomentumParams(length=2))
        mom.update(0.0)
        mom.update(0.0)
        output = mom.update_trade(Trade(t, 3.0, 1.0))
        self.assertEqual(len(output), 1)
        self.assertIsInstance(output[0], Scalar)
        self.assertEqual(output[0].time, t)
        self.assertAlmostEqual(output[0].value, 3.0, places=13)


if __name__ == "__main__":
    unittest.main()
