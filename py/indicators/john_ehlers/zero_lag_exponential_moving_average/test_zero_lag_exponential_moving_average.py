import math
import unittest
from datetime import datetime

from py.indicators.john_ehlers.zero_lag_exponential_moving_average.zero_lag_exponential_moving_average import ZeroLagExponentialMovingAverage
from py.indicators.john_ehlers.zero_lag_exponential_moving_average.params import ZeroLagExponentialMovingAverageParams
from py.indicators.core.identifier import Identifier
from py.entities.bar import Bar
from py.entities.quote import Quote
from py.entities.trade import Trade
from py.entities.scalar import Scalar


class TestZeroLagExponentialMovingAverage(unittest.TestCase):

    def _create(self, sf=0.25, gf=0.5, ml=3):
        return ZeroLagExponentialMovingAverage.create(
            ZeroLagExponentialMovingAverageParams(
                smoothing_factor=sf, velocity_gain_factor=gf, velocity_momentum_length=ml))

    def test_is_primed(self):
        z = self._create()
        self.assertFalse(z.is_primed())
        # First 3 updates (ml=3) should not prime.
        for i in range(3):
            z.update(100.0)
            self.assertFalse(z.is_primed(), f"[{i}]")
        # 4th update should prime.
        z.update(100.0)
        self.assertTrue(z.is_primed())

    def test_update_nan(self):
        z = self._create()
        self.assertTrue(math.isnan(z.update(math.nan)))
        self.assertFalse(z.is_primed())

    def test_update_constant(self):
        value = 42.0
        z = self._create()
        # First 3 updates should be NaN.
        for i in range(3):
            act = z.update(value)
            self.assertTrue(math.isnan(act), f"[{i}] expected NaN")
        # 4th update: primed, constant input → output ≈ input.
        act = z.update(value)
        self.assertAlmostEqual(act, value, delta=1e-10)
        # Further updates with same constant.
        for i in range(10):
            act = z.update(value)
            self.assertAlmostEqual(act, value, delta=1e-10, msg=f"[{i}]")

    def test_update_entity(self):
        t = datetime(2021, 4, 1)
        inp = 100.0
        z = self._create()
        # Prime: 4 updates.
        for _ in range(4):
            z.update(inp)

        # Scalar
        z2 = self._create()
        for _ in range(4):
            z2.update(inp)
        output = z2.update_scalar(Scalar(t, inp))
        self.assertEqual(len(output), 1)
        self.assertIsInstance(output[0], Scalar)
        self.assertEqual(output[0].time, t)
        self.assertFalse(math.isnan(output[0].value))

        # Bar
        z3 = self._create()
        for _ in range(4):
            z3.update(inp)
        output = z3.update_bar(Bar(t, inp, inp, inp, inp, 1.0))
        self.assertEqual(len(output), 1)
        self.assertFalse(math.isnan(output[0].value))

        # Quote
        z4 = self._create()
        for _ in range(4):
            z4.update(inp)
        output = z4.update_quote(Quote(t, inp, inp, 1.0, 1.0))
        self.assertEqual(len(output), 1)
        self.assertFalse(math.isnan(output[0].value))

        # Trade
        z5 = self._create()
        for _ in range(4):
            z5.update(inp)
        output = z5.update_trade(Trade(t, inp, 1.0))
        self.assertEqual(len(output), 1)
        self.assertFalse(math.isnan(output[0].value))

    def test_metadata(self):
        z = self._create()
        meta = z.metadata()
        self.assertEqual(meta.identifier, Identifier.ZERO_LAG_EXPONENTIAL_MOVING_AVERAGE)
        self.assertEqual(meta.mnemonic, "zema(0.25, 0.5, 3)")
        self.assertEqual(len(meta.outputs), 1)
        self.assertEqual(meta.outputs[0].mnemonic, "zema(0.25, 0.5, 3)")

    def test_construction_errors(self):
        # sf = 0
        with self.assertRaises(ValueError):
            self._create(sf=0)
        # sf < 0
        with self.assertRaises(ValueError):
            self._create(sf=-0.1)
        # sf > 1
        with self.assertRaises(ValueError):
            self._create(sf=1.1)
        # sf = 1 should be valid
        z = self._create(sf=1.0)
        self.assertIsNotNone(z)
        # ml = 0
        with self.assertRaises(ValueError):
            self._create(ml=0)
        # ml < 0
        with self.assertRaises(ValueError):
            self._create(ml=-1)


if __name__ == "__main__":
    unittest.main()
