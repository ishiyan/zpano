"""Tests for Dual Differentiator Estimator."""
import math
import unittest

from py.indicators.john_ehlers.hilbert_transformer.dual_differentiator_estimator import DualDifferentiatorEstimator
from py.indicators.john_ehlers.hilbert_transformer.cycle_estimator_params import CycleEstimatorParams
from .test_testdata_dual_differentiator_estimator import *


def _create_default():
    return DualDifferentiatorEstimator(CycleEstimatorParams(
        smoothing_length=4,
        alpha_ema_quadrature_in_phase=0.15,
        alpha_ema_period=0.15,
    ))


def _create_with_warmup(warm_up):
    return DualDifferentiatorEstimator(CycleEstimatorParams(
        smoothing_length=4,
        alpha_ema_quadrature_in_phase=0.15,
        alpha_ema_period=0.15,
        warm_up_period=warm_up,
    ))


class TestDualDifferentiatorEstimator(unittest.TestCase):
    """Tests for DualDifferentiatorEstimator."""

    def test_smoothed(self):
        """Reference implementation: smoothed (test_ht_hd.xsl)."""
        inp = INPUT
        exp = EXPECTED_SMOOTHED
        dde = _create_default()

        lprimed = 3

        for i in range(lprimed):
            dde.update(inp[i])
            self.assertEqual(0, dde.smoothed(), msg=f"[{i}] before primed")

        for i in range(lprimed, len(inp)):
            dde.update(inp[i])
            self.assertAlmostEqual(exp[i], dde.smoothed(), delta=1e-8,
                                   msg=f"[{i}]")

        previous = dde.smoothed()
        dde.update(math.nan)
        self.assertEqual(previous, dde.smoothed(), msg="NaN input")

    def test_period(self):
        """Reference implementation: period (test_ht_hd.xsl)."""
        inp = INPUT
        exp = EXPECTED_PERIOD
        dde = _create_default()

        lprimed = 18
        last = 23

        for i in range(lprimed):
            dde.update(inp[i])
            self.assertEqual(6, dde.period(), msg=f"[{i}] before primed")

        for i in range(lprimed, last):
            dde.update(inp[i])
            if exp[i] is not None:
                self.assertAlmostEqual(exp[i], dde.period(), delta=1e-8,
                                       msg=f"[{i}]")

        previous = dde.period()
        dde.update(math.nan)
        self.assertEqual(previous, dde.period(), msg="NaN input")

    def test_detrended(self):
        """Reference implementation: detrended (test_ht_hd.xsl)."""
        inp = INPUT
        exp = EXPECTED_DETRENDED
        dde = _create_default()

        lprimed = 9
        last = 23

        for i in range(lprimed):
            dde.update(inp[i])
            self.assertEqual(0, dde.detrended(), msg=f"[{i}] before primed")

        for i in range(lprimed, last):
            dde.update(inp[i])
            self.assertAlmostEqual(exp[i], dde.detrended(), delta=1e-8,
                                   msg=f"[{i}]")

        previous = dde.detrended()
        dde.update(math.nan)
        self.assertEqual(previous, dde.detrended(), msg="NaN input")

    def test_quadrature(self):
        """Reference implementation: quadrature (test_ht_hd.xsl)."""
        inp = INPUT
        exp = EXPECTED_QUADRATURE
        dde = _create_default()

        lprimed = 15
        last = 23

        for i in range(lprimed):
            dde.update(inp[i])
            self.assertEqual(0, dde.quadrature(), msg=f"[{i}] before primed")

        for i in range(lprimed, last):
            dde.update(inp[i])
            self.assertAlmostEqual(exp[i], dde.quadrature(), delta=1e-8,
                                   msg=f"[{i}]")

        previous = dde.quadrature()
        dde.update(math.nan)
        self.assertEqual(previous, dde.quadrature(), msg="NaN input")

    def test_in_phase(self):
        """Reference implementation: in-phase (test_ht_hd.xsl)."""
        inp = INPUT
        exp = EXPECTED_IN_PHASE
        dde = _create_default()

        lprimed = 15
        last = 23

        for i in range(lprimed):
            dde.update(inp[i])
            self.assertEqual(0, dde.in_phase(), msg=f"[{i}] before primed")

        for i in range(lprimed, last):
            dde.update(inp[i])
            self.assertAlmostEqual(exp[i], dde.in_phase(), delta=1e-8,
                                   msg=f"[{i}]")

        previous = dde.in_phase()
        dde.update(math.nan)
        self.assertEqual(previous, dde.in_phase(), msg="NaN input")

    def test_primed(self):
        """Reference implementation: primed (test_ht_hd.xsl)."""
        inp = INPUT
        dde = _create_default()

        lprimed = 3 + 7 * 3  # 24

        self.assertFalse(dde.primed())

        for i in range(lprimed):
            dde.update(inp[i])
            self.assertFalse(dde.primed(), msg=f"[{i + 1}] should not be primed")

        for i in range(lprimed, len(inp)):
            dde.update(inp[i])
            self.assertTrue(dde.primed(), msg=f"[{i + 1}] should be primed")

    def test_primed_warmup(self):
        """Reference implementation: primed with warmup."""
        inp = INPUT
        lprimed = 50
        dde = _create_with_warmup(lprimed)

        self.assertFalse(dde.primed())

        for i in range(lprimed):
            dde.update(inp[i])
            self.assertFalse(dde.primed(), msg=f"[{i + 1}] should not be primed")

        for i in range(lprimed, len(inp)):
            dde.update(inp[i])
            self.assertTrue(dde.primed(), msg=f"[{i + 1}] should be primed")

    def test_sin_period(self):
        """Period of sin input with period=30."""
        period = 30
        omega = 2 * math.pi / period
        dde = _create_default()

        for i in range(512):
            dde.update(math.sin(omega * i))

        self.assertAlmostEqual(period, dde.period(), delta=1.0)

    def test_sin_min_period(self):
        """Period=3 sin input clamps to min_period."""
        period = 3
        omega = 2 * math.pi / period
        dde = _create_default()

        for i in range(512):
            dde.update(math.sin(omega * i))

        self.assertAlmostEqual(float(dde.min_period()), dde.period(), delta=1.5)

    def test_sin_max_period(self):
        """Period=60 sin input clamps to max_period."""
        period = 60
        omega = 2 * math.pi / period
        dde = _create_default()

        for i in range(512):
            dde.update(math.sin(omega * i))

        self.assertAlmostEqual(float(dde.max_period()), dde.period(), delta=1.0)

    def test_param_validation(self):
        """Various invalid params raise ValueError."""
        with self.assertRaises(ValueError):
            DualDifferentiatorEstimator(CycleEstimatorParams(smoothing_length=1))
        with self.assertRaises(ValueError):
            DualDifferentiatorEstimator(CycleEstimatorParams(smoothing_length=0))
        with self.assertRaises(ValueError):
            DualDifferentiatorEstimator(CycleEstimatorParams(smoothing_length=-1))
        with self.assertRaises(ValueError):
            DualDifferentiatorEstimator(CycleEstimatorParams(smoothing_length=5))

        with self.assertRaises(ValueError):
            DualDifferentiatorEstimator(CycleEstimatorParams(
                alpha_ema_quadrature_in_phase=0.0))
        with self.assertRaises(ValueError):
            DualDifferentiatorEstimator(CycleEstimatorParams(
                alpha_ema_quadrature_in_phase=-0.01))
        with self.assertRaises(ValueError):
            DualDifferentiatorEstimator(CycleEstimatorParams(
                alpha_ema_quadrature_in_phase=1.0))
        with self.assertRaises(ValueError):
            DualDifferentiatorEstimator(CycleEstimatorParams(
                alpha_ema_quadrature_in_phase=1.01))

        with self.assertRaises(ValueError):
            DualDifferentiatorEstimator(CycleEstimatorParams(
                alpha_ema_period=0.0))
        with self.assertRaises(ValueError):
            DualDifferentiatorEstimator(CycleEstimatorParams(
                alpha_ema_period=-0.01))
        with self.assertRaises(ValueError):
            DualDifferentiatorEstimator(CycleEstimatorParams(
                alpha_ema_period=1.0))
        with self.assertRaises(ValueError):
            DualDifferentiatorEstimator(CycleEstimatorParams(
                alpha_ema_period=1.01))


if __name__ == '__main__':
    unittest.main()
