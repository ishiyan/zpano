import math
import unittest
from datetime import datetime

from py.indicators.gerald_appel.percentage_price_oscillator.percentage_price_oscillator import PercentagePriceOscillator
from py.indicators.gerald_appel.percentage_price_oscillator.params import PercentagePriceOscillatorParams, MovingAverageType
from py.indicators.core.identifier import Identifier
from py.entities.scalar import Scalar

from .test_testdata import INPUT

class TestPercentagePriceOscillator(unittest.TestCase):

    def test_sma_2_3(self):
        ppo = PercentagePriceOscillator(PercentagePriceOscillatorParams(
            fast_length=2, slow_length=3))
        results = [ppo.update(v) for v in INPUT]
        for i in range(2):
            self.assertTrue(math.isnan(results[i]), f"[{i}] expected NaN")
        self.assertAlmostEqual(results[2], 1.10264, delta=5e-4)
        self.assertAlmostEqual(results[3], -0.02813, delta=5e-4)
        self.assertAlmostEqual(results[251], -0.21191, delta=5e-4)
        self.assertTrue(ppo.is_primed())

    def test_sma_12_26(self):
        ppo = PercentagePriceOscillator(PercentagePriceOscillatorParams(
            fast_length=12, slow_length=26))
        results = [ppo.update(v) for v in INPUT]
        for i in range(25):
            self.assertTrue(math.isnan(results[i]), f"[{i}] expected NaN")
        self.assertAlmostEqual(results[25], -3.6393, delta=5e-4)
        self.assertAlmostEqual(results[26], -3.9534, delta=5e-4)
        self.assertAlmostEqual(results[251], -0.15281, delta=5e-4)

    def test_ema_12_26(self):
        ppo = PercentagePriceOscillator(PercentagePriceOscillatorParams(
            fast_length=12, slow_length=26,
            moving_average_type=MovingAverageType.EMA,
            first_is_average=False))
        results = [ppo.update(v) for v in INPUT]
        for i in range(25):
            self.assertTrue(math.isnan(results[i]), f"[{i}] expected NaN")
        self.assertAlmostEqual(results[25], -2.7083, delta=5e-3)
        self.assertAlmostEqual(results[26], -2.7390, delta=5e-3)
        self.assertAlmostEqual(results[251], 0.83644, delta=5e-3)

    def test_is_primed(self):
        ppo = PercentagePriceOscillator(PercentagePriceOscillatorParams(
            fast_length=3, slow_length=5))
        self.assertFalse(ppo.is_primed())
        for i in range(1, 5):
            ppo.update(float(i))
            self.assertFalse(ppo.is_primed())
        ppo.update(5.0)
        self.assertTrue(ppo.is_primed())
        for i in range(6, 10):
            ppo.update(float(i))
            self.assertTrue(ppo.is_primed())

    def test_nan_passthrough(self):
        ppo = PercentagePriceOscillator(PercentagePriceOscillatorParams(
            fast_length=2, slow_length=3))
        result = ppo.update(math.nan)
        self.assertTrue(math.isnan(result))

    def test_metadata_sma(self):
        ppo = PercentagePriceOscillator(PercentagePriceOscillatorParams(
            fast_length=12, slow_length=26))
        meta = ppo.metadata()
        self.assertEqual(meta.identifier, Identifier.PERCENTAGE_PRICE_OSCILLATOR)
        self.assertEqual(meta.mnemonic, "ppo(SMA12/SMA26)")
        self.assertEqual(len(meta.outputs), 1)

    def test_metadata_ema(self):
        ppo = PercentagePriceOscillator(PercentagePriceOscillatorParams(
            fast_length=12, slow_length=26,
            moving_average_type=MovingAverageType.EMA))
        meta = ppo.metadata()
        self.assertEqual(meta.mnemonic, "ppo(EMA12/EMA26)")

    def test_invalid_params(self):
        cases = [
            (1, 26), (12, 1), (-8, 12), (26, -7),
        ]
        for fast, slow in cases:
            with self.assertRaises(ValueError):
                PercentagePriceOscillator(PercentagePriceOscillatorParams(
                    fast_length=fast, slow_length=slow))

    def test_update_entity(self):
        t = datetime(2021, 4, 1)
        ppo = PercentagePriceOscillator(PercentagePriceOscillatorParams(
            fast_length=2, slow_length=3))
        for i in range(2):
            out = ppo.update_scalar(Scalar(t, INPUT[i]))
            self.assertTrue(math.isnan(out[0].value))
        out = ppo.update_scalar(Scalar(t, INPUT[2]))
        self.assertAlmostEqual(out[0].value, 1.10264, delta=5e-4)


if __name__ == "__main__":
    unittest.main()
