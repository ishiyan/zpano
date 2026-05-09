"""Tests for HomodyneDiscriminatorEstimatorUnrolled."""
import math
import unittest

from py.indicators.john_ehlers.hilbert_transformer.homodyne_discriminator_estimator_unrolled import HomodyneDiscriminatorEstimatorUnrolled
from py.indicators.john_ehlers.hilbert_transformer.cycle_estimator_params import CycleEstimatorParams
from .test_testdata_homodyne_discriminator_estimator_unrolled import *


def _create_default():
    return HomodyneDiscriminatorEstimatorUnrolled(CycleEstimatorParams())


def _create_warmup(warm_up):
    return HomodyneDiscriminatorEstimatorUnrolled(
        CycleEstimatorParams(warm_up_period=warm_up)
    )


class TestHomodyneDiscriminatorEstimatorUnrolled(unittest.TestCase):
    """Tests for HomodyneDiscriminatorEstimatorUnrolled."""

    def test_smoothed(self):
        """Test smoothed values against reference data."""
        hdeu = _create_default()
        inp = INPUT
        exp = EXPECTED_SMOOTHED
        lprimed = 3

        for i in range(lprimed):
            hdeu.update(inp[i])
            self.assertAlmostEqual(0, hdeu.smoothed(), delta=1e-8, msg=f"[{i}]")

        for i in range(lprimed, len(inp)):
            hdeu.update(inp[i])
            self.assertAlmostEqual(exp[i], hdeu.smoothed(), delta=1e-8, msg=f"[{i}]")

        previous = hdeu.smoothed()
        hdeu.update(math.nan)
        self.assertAlmostEqual(previous, hdeu.smoothed(), delta=1e-8)

    def test_period(self):
        """Test period values against reference data."""
        hdeu = _create_default()
        inp = INPUT
        exp = EXPECTED_PERIOD
        lprimed = 3

        for i in range(lprimed):
            hdeu.update(inp[i])
            self.assertAlmostEqual(6.0, hdeu.period(), delta=1e-8, msg=f"[{i}]")

        for i in range(lprimed, len(inp)):
            hdeu.update(inp[i])
            self.assertAlmostEqual(exp[i], hdeu.period(), delta=1e-8, msg=f"[{i}]")

        previous = hdeu.period()
        hdeu.update(math.nan)
        self.assertAlmostEqual(previous, hdeu.period(), delta=1e-8)

    def test_detrended(self):
        """Test detrended values against reference data."""
        hdeu = _create_default()
        inp = INPUT
        exp = EXPECTED_DETRENDED
        lprimed = 3

        for i in range(lprimed):
            hdeu.update(inp[i])
            self.assertAlmostEqual(0, hdeu.detrended(), delta=1e-8, msg=f"[{i}]")

        for i in range(lprimed, len(inp)):
            hdeu.update(inp[i])
            self.assertAlmostEqual(exp[i], hdeu.detrended(), delta=1e-8, msg=f"[{i}]")

        previous = hdeu.detrended()
        hdeu.update(math.nan)
        self.assertAlmostEqual(previous, hdeu.detrended(), delta=1e-8)

    def test_quadrature(self):
        """Test quadrature values against reference data."""
        hdeu = _create_default()
        inp = INPUT
        exp = EXPECTED_QUADRATURE
        lprimed = 3

        for i in range(lprimed):
            hdeu.update(inp[i])
            self.assertAlmostEqual(0, hdeu.quadrature(), delta=1e-8, msg=f"[{i}]")

        for i in range(lprimed, len(inp)):
            hdeu.update(inp[i])
            self.assertAlmostEqual(exp[i], hdeu.quadrature(), delta=1e-8, msg=f"[{i}]")

        previous = hdeu.quadrature()
        hdeu.update(math.nan)
        self.assertAlmostEqual(previous, hdeu.quadrature(), delta=1e-8)

    def test_in_phase(self):
        """Test in-phase values against reference data."""
        hdeu = _create_default()
        inp = INPUT
        exp = EXPECTED_IN_PHASE
        lprimed = 3

        for i in range(lprimed):
            hdeu.update(inp[i])
            self.assertAlmostEqual(0, hdeu.in_phase(), delta=1e-8, msg=f"[{i}]")

        for i in range(lprimed, len(inp)):
            hdeu.update(inp[i])
            self.assertAlmostEqual(exp[i], hdeu.in_phase(), delta=1e-8, msg=f"[{i}]")

        previous = hdeu.in_phase()
        hdeu.update(math.nan)
        self.assertAlmostEqual(previous, hdeu.in_phase(), delta=1e-8)

    def test_primed(self):
        """Test priming behavior."""
        hdeu = _create_default()
        inp = INPUT
        lprimed = 2 + 7 * 3  # 23

        self.assertFalse(hdeu.primed())

        for i in range(lprimed):
            hdeu.update(inp[i])
            self.assertFalse(hdeu.primed(), f"[{i + 1}]")

        for i in range(lprimed, len(inp)):
            hdeu.update(inp[i])
            self.assertTrue(hdeu.primed(), f"[{i + 1}]")

    def test_primed_warmup(self):
        """Test priming with warmup period."""
        lprimed = 50
        hdeu = _create_warmup(lprimed)
        inp = INPUT

        self.assertFalse(hdeu.primed())

        for i in range(lprimed):
            hdeu.update(inp[i])
            self.assertFalse(hdeu.primed(), f"[{i + 1}]")

        for i in range(lprimed, len(inp)):
            hdeu.update(inp[i])
            self.assertTrue(hdeu.primed(), f"[{i + 1}]")

    def test_sin_period(self):
        """Test period estimation with sinusoidal input."""
        period = 30
        omega = 2 * math.pi / period
        hdeu = _create_default()

        for i in range(512):
            hdeu.update(math.sin(omega * i))

        self.assertAlmostEqual(period, hdeu.period(), delta=1e-2)

    def test_sin_min_period(self):
        """Test that very short period input clamps to min period."""
        period = 3
        omega = 2 * math.pi / period
        hdeu = _create_default()

        for i in range(512):
            hdeu.update(math.sin(omega * i))

        self.assertAlmostEqual(float(hdeu.min_period()), hdeu.period(), delta=1e-14)

    def test_sin_max_period(self):
        """Test that very long period input clamps to max period."""
        period = 60
        omega = 2 * math.pi / period
        hdeu = _create_default()

        for i in range(512):
            hdeu.update(math.sin(omega * i))

        self.assertAlmostEqual(float(hdeu.max_period()), hdeu.period(), delta=1e-14)

    def test_param_validation(self):
        """Test parameter validation."""
        with self.assertRaises(ValueError):
            HomodyneDiscriminatorEstimatorUnrolled(
                CycleEstimatorParams(smoothing_length=1))
        with self.assertRaises(ValueError):
            HomodyneDiscriminatorEstimatorUnrolled(
                CycleEstimatorParams(smoothing_length=0))
        with self.assertRaises(ValueError):
            HomodyneDiscriminatorEstimatorUnrolled(
                CycleEstimatorParams(smoothing_length=-1))
        with self.assertRaises(ValueError):
            HomodyneDiscriminatorEstimatorUnrolled(
                CycleEstimatorParams(smoothing_length=5))
        with self.assertRaises(ValueError):
            HomodyneDiscriminatorEstimatorUnrolled(
                CycleEstimatorParams(alpha_ema_quadrature_in_phase=0.0))
        with self.assertRaises(ValueError):
            HomodyneDiscriminatorEstimatorUnrolled(
                CycleEstimatorParams(alpha_ema_period=0.0))
        with self.assertRaises(ValueError):
            HomodyneDiscriminatorEstimatorUnrolled(
                CycleEstimatorParams(alpha_ema_quadrature_in_phase=-0.01))
        with self.assertRaises(ValueError):
            HomodyneDiscriminatorEstimatorUnrolled(
                CycleEstimatorParams(alpha_ema_period=-0.01))
        with self.assertRaises(ValueError):
            HomodyneDiscriminatorEstimatorUnrolled(
                CycleEstimatorParams(alpha_ema_quadrature_in_phase=1.0))
        with self.assertRaises(ValueError):
            HomodyneDiscriminatorEstimatorUnrolled(
                CycleEstimatorParams(alpha_ema_period=1.0))
        with self.assertRaises(ValueError):
            HomodyneDiscriminatorEstimatorUnrolled(
                CycleEstimatorParams(alpha_ema_quadrature_in_phase=1.01))
        with self.assertRaises(ValueError):
            HomodyneDiscriminatorEstimatorUnrolled(
                CycleEstimatorParams(alpha_ema_period=1.01))


if __name__ == '__main__':
    unittest.main()
