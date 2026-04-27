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

INPUT = [
    91.500000, 94.815000, 94.375000, 95.095000, 93.780000, 94.625000, 92.530000, 92.750000, 90.315000, 92.470000,
    96.125000, 97.250000, 98.500000, 89.875000, 91.000000, 92.815000, 89.155000, 89.345000, 91.625000, 89.875000,
    88.375000, 87.625000, 84.780000, 83.000000, 83.500000, 81.375000, 84.440000, 89.250000, 86.375000, 86.250000,
    85.250000, 87.125000, 85.815000, 88.970000, 88.470000, 86.875000, 86.815000, 84.875000, 84.190000, 83.875000,
    83.375000, 85.500000, 89.190000, 89.440000, 91.095000, 90.750000, 91.440000, 89.000000, 91.000000, 90.500000,
    89.030000, 88.815000, 84.280000, 83.500000, 82.690000, 84.750000, 85.655000, 86.190000, 88.940000, 89.280000,
    88.625000, 88.500000, 91.970000, 91.500000, 93.250000, 93.500000, 93.155000, 91.720000, 90.000000, 89.690000,
    88.875000, 85.190000, 83.375000, 84.875000, 85.940000, 97.250000, 99.875000, 104.940000, 106.000000, 102.500000,
    102.405000, 104.595000, 106.125000, 106.000000, 106.065000, 104.625000, 108.625000, 109.315000, 110.500000,
    112.750000, 123.000000, 119.625000, 118.750000, 119.250000, 117.940000, 116.440000, 115.190000, 111.875000,
    110.595000, 118.125000, 116.000000, 116.000000, 112.000000, 113.750000, 112.940000, 116.000000, 120.500000,
    116.620000, 117.000000, 115.250000, 114.310000, 115.500000, 115.870000, 120.690000, 120.190000, 120.750000,
    124.750000, 123.370000, 122.940000, 122.560000, 123.120000, 122.560000, 124.620000, 129.250000, 131.000000,
    132.250000, 131.000000, 132.810000, 134.000000, 137.380000, 137.810000, 137.880000, 137.250000, 136.310000,
    136.250000, 134.630000, 128.250000, 129.000000, 123.870000, 124.810000, 123.000000, 126.250000, 128.380000,
    125.370000, 125.690000, 122.250000, 119.370000, 118.500000, 123.190000, 123.500000, 122.190000, 119.310000,
    123.310000, 121.120000, 123.370000, 127.370000, 128.500000, 123.870000, 122.940000, 121.750000, 124.440000,
    122.000000, 122.370000, 122.940000, 124.000000, 123.190000, 124.560000, 127.250000, 125.870000, 128.860000,
    132.000000, 130.750000, 134.750000, 135.000000, 132.380000, 133.310000, 131.940000, 130.000000, 125.370000,
    130.130000, 127.120000, 125.190000, 122.000000, 125.000000, 123.000000, 123.500000, 120.060000, 121.000000,
    117.750000, 119.870000, 122.000000, 119.190000, 116.370000, 113.500000, 114.250000, 110.000000, 105.060000,
    107.000000, 107.870000, 107.000000, 107.120000, 107.000000, 91.000000, 93.940000, 93.870000, 95.500000, 93.000000,
    94.940000, 98.250000, 96.750000, 94.810000, 94.370000, 91.560000, 90.250000, 93.940000, 93.620000, 97.000000,
    95.000000, 95.870000, 94.060000, 94.620000, 93.750000, 98.000000, 103.940000, 107.870000, 106.060000, 104.500000,
    105.000000, 104.190000, 103.060000, 103.420000, 105.270000, 111.870000, 116.000000, 116.620000, 118.280000,
    113.370000, 109.000000, 109.700000, 109.250000, 107.000000, 109.190000, 110.000000, 109.200000, 110.120000,
    108.000000, 108.620000, 109.750000, 109.810000, 109.000000, 108.750000, 107.870000,
]


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
