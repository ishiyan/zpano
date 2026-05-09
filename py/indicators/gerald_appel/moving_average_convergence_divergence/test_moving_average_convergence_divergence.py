import math
import unittest
from datetime import datetime

from py.indicators.gerald_appel.moving_average_convergence_divergence.moving_average_convergence_divergence import MovingAverageConvergenceDivergence
from py.indicators.gerald_appel.moving_average_convergence_divergence.params import MovingAverageConvergenceDivergenceParams, MovingAverageType
from py.indicators.core.identifier import Identifier
from py.entities.scalar import Scalar

from .test_testdata import (
    TEST_INPUT,
    NaN,
    TEST_MACD,
    TEST_SIGNAL,
    TEST_HISTOGRAM,
)


class TestMovingAverageConvergenceDivergenceDefaultParams(unittest.TestCase):
    """Test MACD with default parameters against full 252-element dataset."""

    def test_full_data(self):
        ind = MovingAverageConvergenceDivergence(MovingAverageConvergenceDivergenceParams())
        tolerance = 1e-8

        for i in range(252):
            macd, signal, histogram = ind.update(TEST_INPUT[i])

            if math.isnan(TEST_MACD[i]):
                self.assertTrue(math.isnan(macd), f"[{i}] macd: expected NaN, got {macd}")
                self.assertTrue(math.isnan(signal), f"[{i}] signal: expected NaN, got {signal}")
                self.assertTrue(math.isnan(histogram), f"[{i}] histogram: expected NaN, got {histogram}")
                continue

            self.assertAlmostEqual(macd, TEST_MACD[i], delta=tolerance,
                                   msg=f"[{i}] macd: expected {TEST_MACD[i]}, got {macd}")

            if math.isnan(TEST_SIGNAL[i]):
                self.assertTrue(math.isnan(signal), f"[{i}] signal: expected NaN, got {signal}")
                self.assertTrue(math.isnan(histogram), f"[{i}] histogram: expected NaN, got {histogram}")
                continue

            self.assertAlmostEqual(signal, TEST_SIGNAL[i], delta=tolerance,
                                   msg=f"[{i}] signal: expected {TEST_SIGNAL[i]}, got {signal}")
            self.assertAlmostEqual(histogram, TEST_HISTOGRAM[i], delta=tolerance,
                                   msg=f"[{i}] histogram: expected {TEST_HISTOGRAM[i]}, got {histogram}")


class TestMovingAverageConvergenceDivergenceSpotCheck(unittest.TestCase):
    """TaLib spot check at index 33."""

    def test_spot_check(self):
        tolerance = 5e-4
        ind = MovingAverageConvergenceDivergence(MovingAverageConvergenceDivergenceParams())

        for i in range(34):
            macd, signal, histogram = ind.update(TEST_INPUT[i])

        self.assertAlmostEqual(macd, -1.9738, delta=tolerance)
        self.assertAlmostEqual(signal, -2.7071, delta=tolerance)
        expected_histogram = (-1.9738) - (-2.7071)
        self.assertAlmostEqual(histogram, expected_histogram, delta=tolerance)


class TestMovingAverageConvergenceDivergencePeriodInversion(unittest.TestCase):
    """Test auto-swap of fast/slow when slow < fast."""

    def test_period_inversion(self):
        tolerance = 5e-4
        ind = MovingAverageConvergenceDivergence(
            MovingAverageConvergenceDivergenceParams(fast_length=26, slow_length=12))

        for i in range(34):
            macd, signal, _ = ind.update(TEST_INPUT[i])

        self.assertAlmostEqual(macd, -1.9738, delta=tolerance)
        self.assertAlmostEqual(signal, -2.7071, delta=tolerance)


class TestMovingAverageConvergenceDivergenceIsPrimed(unittest.TestCase):
    """Test priming behavior."""

    def test_primed(self):
        ind = MovingAverageConvergenceDivergence(
            MovingAverageConvergenceDivergenceParams(fast_length=3, slow_length=5, signal_length=2))

        self.assertFalse(ind.is_primed())

        for i in range(6):
            ind.update(float(i + 1))
            if i < 5:
                self.assertFalse(ind.is_primed(), f"[{i}] expected not primed")

        self.assertTrue(ind.is_primed())


class TestMovingAverageConvergenceDivergenceNaN(unittest.TestCase):
    """Test NaN input handling."""

    def test_nan(self):
        ind = MovingAverageConvergenceDivergence(MovingAverageConvergenceDivergenceParams())
        macd, signal, histogram = ind.update(math.nan)
        self.assertTrue(math.isnan(macd))
        self.assertTrue(math.isnan(signal))
        self.assertTrue(math.isnan(histogram))


class TestMovingAverageConvergenceDivergenceMetadata(unittest.TestCase):
    """Test metadata generation."""

    def test_default_metadata(self):
        ind = MovingAverageConvergenceDivergence(MovingAverageConvergenceDivergenceParams())
        meta = ind.metadata()
        self.assertEqual(meta.identifier, Identifier.MOVING_AVERAGE_CONVERGENCE_DIVERGENCE)
        self.assertEqual(meta.mnemonic, "macd(12,26,9)")
        self.assertEqual(len(meta.outputs), 3)

    def test_sma_metadata(self):
        ind = MovingAverageConvergenceDivergence(
            MovingAverageConvergenceDivergenceParams(moving_average_type=MovingAverageType.SMA))
        meta = ind.metadata()
        self.assertEqual(meta.mnemonic, "macd(12,26,9,SMA,EMA)")


class TestMovingAverageConvergenceDivergenceUpdateScalar(unittest.TestCase):
    """Test update_scalar output."""

    def test_update_scalar(self):
        tolerance = 5e-4
        ind = MovingAverageConvergenceDivergence(MovingAverageConvergenceDivergenceParams())
        tm = datetime(2021, 4, 1)

        for i in range(34):
            out = ind.update_scalar(Scalar(time=tm, value=TEST_INPUT[i]))

        self.assertAlmostEqual(out[0].value, -1.9738, delta=tolerance)
        self.assertAlmostEqual(out[1].value, -2.7071, delta=tolerance)


class TestMovingAverageConvergenceDivergenceInvalidParams(unittest.TestCase):
    """Test invalid parameter validation."""

    def test_fast_too_small(self):
        with self.assertRaises(ValueError):
            MovingAverageConvergenceDivergence(
                MovingAverageConvergenceDivergenceParams(fast_length=1))

    def test_slow_too_small(self):
        with self.assertRaises(ValueError):
            MovingAverageConvergenceDivergence(
                MovingAverageConvergenceDivergenceParams(slow_length=1))

    def test_signal_negative(self):
        with self.assertRaises(ValueError):
            MovingAverageConvergenceDivergence(
                MovingAverageConvergenceDivergenceParams(signal_length=-1))

    def test_fast_negative(self):
        with self.assertRaises(ValueError):
            MovingAverageConvergenceDivergence(
                MovingAverageConvergenceDivergenceParams(fast_length=-8, slow_length=12))

    def test_slow_negative(self):
        with self.assertRaises(ValueError):
            MovingAverageConvergenceDivergence(
                MovingAverageConvergenceDivergenceParams(fast_length=26, slow_length=-7))


if __name__ == '__main__':
    unittest.main()
