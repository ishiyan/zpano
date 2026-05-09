"""Tests for Homodyne Discriminator Estimator."""
import math
import unittest

from py.indicators.john_ehlers.hilbert_transformer.homodyne_discriminator_estimator import HomodyneDiscriminatorEstimator
from py.indicators.john_ehlers.hilbert_transformer.cycle_estimator_params import CycleEstimatorParams
from .test_testdata_homodyne_discriminator_estimator import *


def _create_default():
    return HomodyneDiscriminatorEstimator(CycleEstimatorParams())


def _create_with_warmup(warm_up):
    return HomodyneDiscriminatorEstimator(
        CycleEstimatorParams(warm_up_period=warm_up)
    )


class TestHomodyneDiscriminatorEstimator(unittest.TestCase):
    """Tests for HomodyneDiscriminatorEstimator."""

    def test_smoothed(self):
        """Reference implementation: smoothed (test_ht_hd.xsl)."""
        inp = INPUT
        exp = EXPECTED_SMOOTHED
        hde = _create_default()

        lprimed = 3

        for i in range(lprimed):
            hde.update(inp[i])
            self.assertEqual(0, hde.smoothed(), msg=f"[{i}] before primed")

        for i in range(lprimed, len(inp)):
            hde.update(inp[i])
            self.assertAlmostEqual(exp[i], hde.smoothed(), delta=1e-8,
                                   msg=f"[{i}]")

        previous = hde.smoothed()
        hde.update(math.nan)
        self.assertEqual(previous, hde.smoothed(), msg="NaN input")

    def test_period(self):
        """Reference implementation: period (test_ht_hd.xsl)."""
        inp = INPUT
        exp = EXPECTED_PERIOD
        hde = _create_default()

        lprimed = 23

        for i in range(lprimed):
            hde.update(inp[i])
            self.assertEqual(6, hde.period(), msg=f"[{i}] before primed")

        for i in range(lprimed, len(inp)):
            hde.update(inp[i])
            self.assertAlmostEqual(exp[i], hde.period(), delta=1e-8,
                                   msg=f"[{i}]")

        previous = hde.period()
        hde.update(math.nan)
        self.assertEqual(previous, hde.period(), msg="NaN input")

    def test_detrended(self):
        """Reference implementation: detrended (test_ht_hd.xsl)."""
        inp = INPUT
        exp = EXPECTED_DETRENDED
        hde = _create_default()

        lprimed = 9

        for i in range(lprimed):
            hde.update(inp[i])
            self.assertEqual(0, hde.detrended(), msg=f"[{i}] before primed")

        for i in range(lprimed, len(inp)):
            hde.update(inp[i])
            self.assertAlmostEqual(exp[i], hde.detrended(), delta=1e-8,
                                   msg=f"[{i}]")

        previous = hde.detrended()
        hde.update(math.nan)
        self.assertEqual(previous, hde.detrended(), msg="NaN input")

    def test_quadrature(self):
        """Reference implementation: quadrature (test_ht_hd.xsl)."""
        inp = INPUT
        exp = EXPECTED_QUADRATURE
        hde = _create_default()

        lprimed = 15

        for i in range(lprimed):
            hde.update(inp[i])
            self.assertEqual(0, hde.quadrature(), msg=f"[{i}] before primed")

        for i in range(lprimed, len(inp)):
            hde.update(inp[i])
            self.assertAlmostEqual(exp[i], hde.quadrature(), delta=1e-8,
                                   msg=f"[{i}]")

        previous = hde.quadrature()
        hde.update(math.nan)
        self.assertEqual(previous, hde.quadrature(), msg="NaN input")

    def test_in_phase(self):
        """Reference implementation: in-phase (test_ht_hd.xsl)."""
        inp = INPUT
        exp = EXPECTED_IN_PHASE
        hde = _create_default()

        lprimed = 15

        for i in range(lprimed):
            hde.update(inp[i])
            self.assertEqual(0, hde.in_phase(), msg=f"[{i}] before primed")

        for i in range(lprimed, len(inp)):
            hde.update(inp[i])
            self.assertAlmostEqual(exp[i], hde.in_phase(), delta=1e-8,
                                   msg=f"[{i}]")

        previous = hde.in_phase()
        hde.update(math.nan)
        self.assertEqual(previous, hde.in_phase(), msg="NaN input")

    def test_primed(self):
        """Reference implementation: primed (test_ht_hd.xsl)."""
        inp = INPUT
        hde = _create_default()

        lprimed = 4 + 7 * 3  # 25

        self.assertFalse(hde.primed())

        for i in range(lprimed):
            hde.update(inp[i])
            self.assertFalse(hde.primed(), msg=f"[{i + 1}] should not be primed")

        for i in range(lprimed, len(inp)):
            hde.update(inp[i])
            self.assertTrue(hde.primed(), msg=f"[{i + 1}] should be primed")

    def test_primed_warmup(self):
        """Reference implementation: primed with warmup."""
        inp = INPUT
        lprimed = 50
        hde = _create_with_warmup(lprimed)

        self.assertFalse(hde.primed())

        for i in range(lprimed):
            hde.update(inp[i])
            self.assertFalse(hde.primed(), msg=f"[{i + 1}] should not be primed")

        for i in range(lprimed, len(inp)):
            hde.update(inp[i])
            self.assertTrue(hde.primed(), msg=f"[{i + 1}] should be primed")

    def test_sin_period(self):
        """Period of sin input with period=30."""
        period = 30
        omega = 2 * math.pi / period
        hde = _create_default()

        for i in range(512):
            hde.update(math.sin(omega * i))

        self.assertAlmostEqual(period, hde.period(), delta=0.01)

    def test_sin_min_period(self):
        """Period=3 sin input clamps to min_period=6."""
        period = 3
        omega = 2 * math.pi / period
        hde = _create_default()

        for i in range(512):
            hde.update(math.sin(omega * i))

        self.assertAlmostEqual(float(hde.min_period()), hde.period(), delta=1e-14)

    def test_sin_max_period(self):
        """Period=60 sin input clamps to max_period=50."""
        period = 60
        omega = 2 * math.pi / period
        hde = _create_default()

        for i in range(512):
            hde.update(math.sin(omega * i))

        self.assertAlmostEqual(float(hde.max_period()), hde.period(), delta=1e-14)

    def test_param_validation(self):
        """Various invalid params raise ValueError."""
        # smoothing_length out of range
        with self.assertRaises(ValueError):
            HomodyneDiscriminatorEstimator(CycleEstimatorParams(smoothing_length=1))
        with self.assertRaises(ValueError):
            HomodyneDiscriminatorEstimator(CycleEstimatorParams(smoothing_length=0))
        with self.assertRaises(ValueError):
            HomodyneDiscriminatorEstimator(CycleEstimatorParams(smoothing_length=-1))
        with self.assertRaises(ValueError):
            HomodyneDiscriminatorEstimator(CycleEstimatorParams(smoothing_length=5))

        # alpha_ema_quadrature_in_phase out of range
        with self.assertRaises(ValueError):
            HomodyneDiscriminatorEstimator(CycleEstimatorParams(
                alpha_ema_quadrature_in_phase=0.0))
        with self.assertRaises(ValueError):
            HomodyneDiscriminatorEstimator(CycleEstimatorParams(
                alpha_ema_quadrature_in_phase=-0.01))
        with self.assertRaises(ValueError):
            HomodyneDiscriminatorEstimator(CycleEstimatorParams(
                alpha_ema_quadrature_in_phase=1.0))
        with self.assertRaises(ValueError):
            HomodyneDiscriminatorEstimator(CycleEstimatorParams(
                alpha_ema_quadrature_in_phase=1.01))

        # alpha_ema_period out of range
        with self.assertRaises(ValueError):
            HomodyneDiscriminatorEstimator(CycleEstimatorParams(
                alpha_ema_period=0.0))
        with self.assertRaises(ValueError):
            HomodyneDiscriminatorEstimator(CycleEstimatorParams(
                alpha_ema_period=-0.01))
        with self.assertRaises(ValueError):
            HomodyneDiscriminatorEstimator(CycleEstimatorParams(
                alpha_ema_period=1.0))
        with self.assertRaises(ValueError):
            HomodyneDiscriminatorEstimator(CycleEstimatorParams(
                alpha_ema_period=1.01))


if __name__ == '__main__':
    unittest.main()
