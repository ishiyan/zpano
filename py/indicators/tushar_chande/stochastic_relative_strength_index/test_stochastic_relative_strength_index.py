import math
import unittest
from datetime import datetime

from py.indicators.tushar_chande.stochastic_relative_strength_index.stochastic_relative_strength_index import \
    StochasticRelativeStrengthIndex
from py.indicators.tushar_chande.stochastic_relative_strength_index.params import \
    StochasticRelativeStrengthIndexParams, MovingAverageType
from py.indicators.core.identifier import Identifier
from py.entities.scalar import Scalar

from .test_testdata import TEST_INPUT


# Test data from TA-Lib (252 entries).
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
