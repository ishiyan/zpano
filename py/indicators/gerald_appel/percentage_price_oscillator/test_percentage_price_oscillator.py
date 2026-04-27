import math
import unittest
from datetime import datetime

from py.indicators.gerald_appel.percentage_price_oscillator.percentage_price_oscillator import PercentagePriceOscillator
from py.indicators.gerald_appel.percentage_price_oscillator.params import PercentagePriceOscillatorParams, MovingAverageType
from py.indicators.core.identifier import Identifier
from py.entities.scalar import Scalar

INPUT = [
    91.500000, 94.815000, 94.375000, 95.095000, 93.780000, 94.625000, 92.530000, 92.750000, 90.315000, 92.470000, 96.125000,
    97.250000, 98.500000, 89.875000, 91.000000, 92.815000, 89.155000, 89.345000, 91.625000, 89.875000, 88.375000, 87.625000,
    84.780000, 83.000000, 83.500000, 81.375000, 84.440000, 89.250000, 86.375000, 86.250000, 85.250000, 87.125000, 85.815000,
    88.970000, 88.470000, 86.875000, 86.815000, 84.875000, 84.190000, 83.875000, 83.375000, 85.500000, 89.190000, 89.440000,
    91.095000, 90.750000, 91.440000, 89.000000, 91.000000, 90.500000, 89.030000, 88.815000, 84.280000, 83.500000, 82.690000,
    84.750000, 85.655000, 86.190000, 88.940000, 89.280000, 88.625000, 88.500000, 91.970000, 91.500000, 93.250000, 93.500000,
    93.155000, 91.720000, 90.000000, 89.690000, 88.875000, 85.190000, 83.375000, 84.875000, 85.940000, 97.250000, 99.875000,
    104.940000, 106.000000, 102.500000, 102.405000, 104.595000, 106.125000, 106.000000, 106.065000, 104.625000, 108.625000,
    109.315000, 110.500000, 112.750000, 123.000000, 119.625000, 118.750000, 119.250000, 117.940000, 116.440000, 115.190000,
    111.875000, 110.595000, 118.125000, 116.000000, 116.000000, 112.000000, 113.750000, 112.940000, 116.000000, 120.500000,
    116.620000, 117.000000, 115.250000, 114.310000, 115.500000, 115.870000, 120.690000, 120.190000, 120.750000, 124.750000,
    123.370000, 122.940000, 122.560000, 123.120000, 122.560000, 124.620000, 129.250000, 131.000000, 132.250000, 131.000000,
    132.810000, 134.000000, 137.380000, 137.810000, 137.880000, 137.250000, 136.310000, 136.250000, 134.630000, 128.250000,
    129.000000, 123.870000, 124.810000, 123.000000, 126.250000, 128.380000, 125.370000, 125.690000, 122.250000, 119.370000,
    118.500000, 123.190000, 123.500000, 122.190000, 119.310000, 123.310000, 121.120000, 123.370000, 127.370000, 128.500000,
    123.870000, 122.940000, 121.750000, 124.440000, 122.000000, 122.370000, 122.940000, 124.000000, 123.190000, 124.560000,
    127.250000, 125.870000, 128.860000, 132.000000, 130.750000, 134.750000, 135.000000, 132.380000, 133.310000, 131.940000,
    130.000000, 125.370000, 130.130000, 127.120000, 125.190000, 122.000000, 125.000000, 123.000000, 123.500000, 120.060000,
    121.000000, 117.750000, 119.870000, 122.000000, 119.190000, 116.370000, 113.500000, 114.250000, 110.000000, 105.060000,
    107.000000, 107.870000, 107.000000, 107.120000, 107.000000, 91.000000, 93.940000, 93.870000, 95.500000, 93.000000,
    94.940000, 98.250000, 96.750000, 94.810000, 94.370000, 91.560000, 90.250000, 93.940000, 93.620000, 97.000000, 95.000000,
    95.870000, 94.060000, 94.620000, 93.750000, 98.000000, 103.940000, 107.870000, 106.060000, 104.500000, 105.000000,
    104.190000, 103.060000, 103.420000, 105.270000, 111.870000, 116.000000, 116.620000, 118.280000, 113.370000, 109.000000,
    109.700000, 109.250000, 107.000000, 109.190000, 110.000000, 109.200000, 110.120000, 108.000000, 108.620000, 109.750000,
    109.810000, 109.000000, 108.750000, 107.870000,
]


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
