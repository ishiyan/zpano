import math
import unittest
from datetime import datetime

from py.indicators.john_ehlers.zero_lag_error_correcting_exponential_moving_average.zero_lag_error_correcting_exponential_moving_average import ZeroLagErrorCorrectingExponentialMovingAverage
from py.indicators.john_ehlers.zero_lag_error_correcting_exponential_moving_average.params import ZeroLagErrorCorrectingExponentialMovingAverageParams
from py.indicators.core.identifier import Identifier
from py.entities.bar import Bar
from py.entities.quote import Quote
from py.entities.trade import Trade
from py.entities.scalar import Scalar


class TestZeroLagErrorCorrectingExponentialMovingAverage(unittest.TestCase):

    def _create(self, sf=0.095, gl=5.0, gs=0.1):
        return ZeroLagErrorCorrectingExponentialMovingAverage.create(
            ZeroLagErrorCorrectingExponentialMovingAverageParams(
                smoothing_factor=sf, gain_limit=gl, gain_step=gs))

    def test_is_primed(self):
        z = self._create()
        self.assertFalse(z.is_primed())
        # First 2 updates should not prime.
        for i in range(2):
            z.update(100.0)
            self.assertFalse(z.is_primed(), f"[{i}]")
        # 3rd update should prime.
        z.update(100.0)
        self.assertTrue(z.is_primed())

    def test_update_nan(self):
        z = self._create()
        self.assertTrue(math.isnan(z.update(math.nan)))
        self.assertFalse(z.is_primed())

    def test_update_constant(self):
        value = 42.0
        z = self._create()
        # First 2 updates should be NaN.
        for i in range(2):
            act = z.update(value)
            self.assertTrue(math.isnan(act), f"[{i}] expected NaN")
        # 3rd update: primed, constant input → output ≈ input.
        act = z.update(value)
        self.assertAlmostEqual(act, value, delta=1e-6)
        # Further updates with same constant.
        for i in range(10):
            act = z.update(value)
            self.assertAlmostEqual(act, value, delta=1e-6, msg=f"[{i}]")

    def test_update_entity(self):
        t = datetime(2021, 4, 1)
        inp = 100.0
        z = self._create()
        # Prime: 3 updates.
        for _ in range(3):
            z.update(inp)

        # Scalar
        z2 = self._create()
        for _ in range(3):
            z2.update(inp)
        output = z2.update_scalar(Scalar(t, inp))
        self.assertEqual(len(output), 1)
        self.assertIsInstance(output[0], Scalar)
        self.assertEqual(output[0].time, t)
        self.assertFalse(math.isnan(output[0].value))

        # Bar
        z3 = self._create()
        for _ in range(3):
            z3.update(inp)
        output = z3.update_bar(Bar(t, inp, inp, inp, inp, 1.0))
        self.assertEqual(len(output), 1)
        self.assertFalse(math.isnan(output[0].value))

        # Quote
        z4 = self._create()
        for _ in range(3):
            z4.update(inp)
        output = z4.update_quote(Quote(t, inp, inp, 1.0, 1.0))
        self.assertEqual(len(output), 1)
        self.assertFalse(math.isnan(output[0].value))

        # Trade
        z5 = self._create()
        for _ in range(3):
            z5.update(inp)
        output = z5.update_trade(Trade(t, inp, 1.0))
        self.assertEqual(len(output), 1)
        self.assertFalse(math.isnan(output[0].value))

    def test_metadata(self):
        z = self._create()
        meta = z.metadata()
        self.assertEqual(meta.identifier, Identifier.ZERO_LAG_ERROR_CORRECTING_EXPONENTIAL_MOVING_AVERAGE)
        self.assertEqual(meta.mnemonic, "zecema(0.095, 5, 0.1)")
        self.assertEqual(len(meta.outputs), 1)
        self.assertEqual(meta.outputs[0].mnemonic, "zecema(0.095, 5, 0.1)")

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
        # gl = 0
        with self.assertRaises(ValueError):
            self._create(gl=0)
        # gl < 0
        with self.assertRaises(ValueError):
            self._create(gl=-1)
        # gs = 0
        with self.assertRaises(ValueError):
            self._create(gs=0)
        # gs < 0
        with self.assertRaises(ValueError):
            self._create(gs=-0.1)


if __name__ == "__main__":
    unittest.main()
