import math
import unittest
from datetime import datetime

from py.indicators.common.absolute_price_oscillator.absolute_price_oscillator import AbsolutePriceOscillator
from py.indicators.common.absolute_price_oscillator.params import AbsolutePriceOscillatorParams, MovingAverageType
from py.indicators.core.identifier import Identifier
from py.entities.scalar import Scalar

from .test_testdata import INPUT

class TestAbsolutePriceOscillator(unittest.TestCase):

    def test_sma_12_26(self):
        apo = AbsolutePriceOscillator(AbsolutePriceOscillatorParams(
            fast_length=12, slow_length=26))
        results = [apo.update(v) for v in INPUT]
        for i in range(25):
            self.assertTrue(math.isnan(results[i]), f"[{i}] expected NaN")
        self.assertAlmostEqual(results[25], -3.3124, delta=5e-4)
        self.assertAlmostEqual(results[26], -3.5876, delta=5e-4)
        self.assertAlmostEqual(results[251], -0.1667, delta=5e-4)
        self.assertTrue(apo.is_primed())

    def test_ema_12_26(self):
        apo = AbsolutePriceOscillator(AbsolutePriceOscillatorParams(
            fast_length=12, slow_length=26,
            moving_average_type=MovingAverageType.EMA,
            first_is_average=False))
        results = [apo.update(v) for v in INPUT]
        for i in range(25):
            self.assertTrue(math.isnan(results[i]), f"[{i}] expected NaN")
        self.assertAlmostEqual(results[25], -2.4193, delta=5e-4)
        self.assertAlmostEqual(results[26], -2.4367, delta=5e-4)
        self.assertAlmostEqual(results[251], 0.90401, delta=5e-4)

    def test_is_primed(self):
        apo = AbsolutePriceOscillator(AbsolutePriceOscillatorParams(
            fast_length=3, slow_length=5))
        self.assertFalse(apo.is_primed())
        for i in range(1, 5):
            apo.update(float(i))
            self.assertFalse(apo.is_primed())
        apo.update(5.0)
        self.assertTrue(apo.is_primed())
        for i in range(6, 10):
            apo.update(float(i))
            self.assertTrue(apo.is_primed())

    def test_nan_passthrough(self):
        apo = AbsolutePriceOscillator(AbsolutePriceOscillatorParams(
            fast_length=2, slow_length=3))
        result = apo.update(math.nan)
        self.assertTrue(math.isnan(result))

    def test_metadata_sma(self):
        apo = AbsolutePriceOscillator(AbsolutePriceOscillatorParams(
            fast_length=12, slow_length=26))
        meta = apo.metadata()
        self.assertEqual(meta.identifier, Identifier.ABSOLUTE_PRICE_OSCILLATOR)
        self.assertEqual(meta.mnemonic, "apo(SMA12/SMA26)")
        self.assertEqual(len(meta.outputs), 1)

    def test_metadata_ema(self):
        apo = AbsolutePriceOscillator(AbsolutePriceOscillatorParams(
            fast_length=12, slow_length=26,
            moving_average_type=MovingAverageType.EMA))
        meta = apo.metadata()
        self.assertEqual(meta.mnemonic, "apo(EMA12/EMA26)")

    def test_invalid_params(self):
        cases = [
            (1, 26), (12, 1), (-8, 12), (26, -7),
        ]
        for fast, slow in cases:
            with self.assertRaises(ValueError):
                AbsolutePriceOscillator(AbsolutePriceOscillatorParams(
                    fast_length=fast, slow_length=slow))

    def test_update_entity(self):
        t = datetime(2021, 4, 1)
        apo = AbsolutePriceOscillator(AbsolutePriceOscillatorParams(
            fast_length=2, slow_length=3))
        for i in range(2):
            out = apo.update_scalar(Scalar(t, INPUT[i]))
            self.assertTrue(math.isnan(out[0].value))
        out = apo.update_scalar(Scalar(t, INPUT[2]))
        self.assertFalse(math.isnan(out[0].value))


if __name__ == "__main__":
    unittest.main()
