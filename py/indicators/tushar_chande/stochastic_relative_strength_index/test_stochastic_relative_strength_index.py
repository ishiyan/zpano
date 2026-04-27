import math
import unittest
from datetime import datetime

from py.indicators.tushar_chande.stochastic_relative_strength_index.stochastic_relative_strength_index import \
    StochasticRelativeStrengthIndex
from py.indicators.tushar_chande.stochastic_relative_strength_index.params import \
    StochasticRelativeStrengthIndexParams, MovingAverageType
from py.indicators.core.identifier import Identifier
from py.entities.scalar import Scalar


# Test data from TA-Lib (252 entries).
TEST_INPUT = [
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


class TestStochRSI_14_14_1_SMA(unittest.TestCase):
    """Test case 1: period=14, fastK=14, fastD=1, SMA."""

    def test_values(self):
        ind = StochasticRelativeStrengthIndex(
            StochasticRelativeStrengthIndexParams(
                length=14, fast_k_length=14, fast_d_length=1))
        tolerance = 1e-4

        # First 27 values should produce NaN for FastK.
        for i in range(27):
            fast_k, _ = ind.update(TEST_INPUT[i])
            self.assertTrue(math.isnan(fast_k), f"[{i}] expected NaN FastK")

        # Index 27: first value.
        fast_k, fast_d = ind.update(TEST_INPUT[27])
        self.assertFalse(math.isnan(fast_k), "[27] expected non-NaN FastK")
        self.assertAlmostEqual(fast_k, 94.156709, delta=tolerance)
        self.assertAlmostEqual(fast_d, 94.156709, delta=tolerance)

        # Feed remaining and check last value.
        for i in range(28, 251):
            ind.update(TEST_INPUT[i])

        fast_k, fast_d = ind.update(TEST_INPUT[251])
        self.assertAlmostEqual(fast_k, 0.0, delta=tolerance)
        self.assertAlmostEqual(fast_d, 0.0, delta=tolerance)


class TestStochRSI_14_45_1_SMA(unittest.TestCase):
    """Test case 2: period=14, fastK=45, fastD=1, SMA."""

    def test_values(self):
        ind = StochasticRelativeStrengthIndex(
            StochasticRelativeStrengthIndexParams(
                length=14, fast_k_length=45, fast_d_length=1))
        tolerance = 1e-4

        # First 58 values should produce NaN for FastK.
        for i in range(58):
            fast_k, _ = ind.update(TEST_INPUT[i])
            self.assertTrue(math.isnan(fast_k), f"[{i}] expected NaN FastK")

        # Index 58: first value.
        fast_k, fast_d = ind.update(TEST_INPUT[58])
        self.assertFalse(math.isnan(fast_k), "[58] expected non-NaN FastK")
        self.assertAlmostEqual(fast_k, 79.729186, delta=tolerance)
        self.assertAlmostEqual(fast_d, 79.729186, delta=tolerance)

        # Feed remaining and check last value.
        for i in range(59, 251):
            ind.update(TEST_INPUT[i])

        fast_k, fast_d = ind.update(TEST_INPUT[251])
        self.assertAlmostEqual(fast_k, 48.1550743, delta=tolerance)
        self.assertAlmostEqual(fast_d, 48.1550743, delta=tolerance)


class TestStochRSI_11_13_16_SMA(unittest.TestCase):
    """Test case 3: period=11, fastK=13, fastD=16, SMA."""

    def test_values(self):
        ind = StochasticRelativeStrengthIndex(
            StochasticRelativeStrengthIndexParams(
                length=11, fast_k_length=13, fast_d_length=16))
        tolerance = 1e-3

        # Feed first 38 values.
        for i in range(38):
            ind.update(TEST_INPUT[i])

        # Index 38: first primed value.
        fast_k, fast_d = ind.update(TEST_INPUT[38])
        self.assertAlmostEqual(fast_k, 5.25947, delta=tolerance)
        self.assertAlmostEqual(fast_d, 57.1711, delta=tolerance)
        self.assertTrue(ind.is_primed())

        # Feed remaining and check last value.
        for i in range(39, 251):
            ind.update(TEST_INPUT[i])

        fast_k, fast_d = ind.update(TEST_INPUT[251])
        self.assertAlmostEqual(fast_k, 0.0, delta=tolerance)
        self.assertAlmostEqual(fast_d, 15.7303, delta=tolerance)


class TestStochRSIIsPrimed(unittest.TestCase):
    """Test priming behavior."""

    def test_priming(self):
        ind = StochasticRelativeStrengthIndex(
            StochasticRelativeStrengthIndexParams(
                length=14, fast_k_length=14, fast_d_length=1))

        self.assertFalse(ind.is_primed())

        for i in range(27):
            ind.update(TEST_INPUT[i])
            self.assertFalse(ind.is_primed(), f"[{i}] should not be primed")

        ind.update(TEST_INPUT[27])
        self.assertTrue(ind.is_primed(), "should be primed after index 27")


class TestStochRSINaN(unittest.TestCase):
    """Test NaN passthrough."""

    def test_nan(self):
        ind = StochasticRelativeStrengthIndex(
            StochasticRelativeStrengthIndexParams(
                length=14, fast_k_length=14, fast_d_length=1))

        fast_k, fast_d = ind.update(math.nan)
        self.assertTrue(math.isnan(fast_k))
        self.assertTrue(math.isnan(fast_d))


class TestStochRSIMetadata(unittest.TestCase):
    """Test metadata output."""

    def test_metadata(self):
        ind = StochasticRelativeStrengthIndex(
            StochasticRelativeStrengthIndexParams(
                length=14, fast_k_length=14, fast_d_length=3))

        m = ind.metadata()
        self.assertEqual(m.identifier, Identifier.STOCHASTIC_RELATIVE_STRENGTH_INDEX)
        self.assertEqual(m.mnemonic, "stochrsi(14/14/SMA3)")
        self.assertEqual(len(m.outputs), 2)


class TestStochRSIEntities(unittest.TestCase):
    """Test entity update methods."""

    def test_update_scalar(self):
        ind = StochasticRelativeStrengthIndex(
            StochasticRelativeStrengthIndexParams(
                length=14, fast_k_length=14, fast_d_length=1))
        tolerance = 1e-4
        tm = datetime(2021, 4, 1)

        for i in range(27):
            out = ind.update_scalar(Scalar(time=tm, value=TEST_INPUT[i]))
            self.assertTrue(math.isnan(out[0].value), f"[{i}] expected NaN")

        out = ind.update_scalar(Scalar(time=tm, value=TEST_INPUT[27]))
        self.assertAlmostEqual(out[0].value, 94.156709, delta=tolerance)
        self.assertAlmostEqual(out[1].value, 94.156709, delta=tolerance)


class TestStochRSIInvalidParams(unittest.TestCase):
    """Test invalid parameter validation."""

    def test_length_too_small(self):
        with self.assertRaises(ValueError):
            StochasticRelativeStrengthIndex(
                StochasticRelativeStrengthIndexParams(length=1, fast_k_length=14, fast_d_length=3))

    def test_fast_k_too_small(self):
        with self.assertRaises(ValueError):
            StochasticRelativeStrengthIndex(
                StochasticRelativeStrengthIndexParams(length=14, fast_k_length=0, fast_d_length=3))

    def test_fast_d_too_small(self):
        with self.assertRaises(ValueError):
            StochasticRelativeStrengthIndex(
                StochasticRelativeStrengthIndexParams(length=14, fast_k_length=14, fast_d_length=0))

    def test_length_negative(self):
        with self.assertRaises(ValueError):
            StochasticRelativeStrengthIndex(
                StochasticRelativeStrengthIndexParams(length=-1, fast_k_length=14, fast_d_length=3))


if __name__ == '__main__':
    unittest.main()
