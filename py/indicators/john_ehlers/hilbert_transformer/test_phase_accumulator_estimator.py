"""Tests for Phase Accumulator Estimator."""
import math
import unittest

from py.indicators.john_ehlers.hilbert_transformer.phase_accumulator_estimator import PhaseAccumulatorEstimator
from py.indicators.john_ehlers.hilbert_transformer.cycle_estimator_params import CycleEstimatorParams
from .test_testdata_phase_accumulator_estimator import *


def _create_default():
    return PhaseAccumulatorEstimator(CycleEstimatorParams(
        smoothing_length=4,
        alpha_ema_quadrature_in_phase=0.15,
        alpha_ema_period=0.25,
    ))


def _create_with_warmup(warm_up):
    return PhaseAccumulatorEstimator(CycleEstimatorParams(
        smoothing_length=4,
        alpha_ema_quadrature_in_phase=0.15,
        alpha_ema_period=0.25,
        warm_up_period=warm_up,
    ))


class TestPhaseAccumulatorEstimator(unittest.TestCase):
    """Tests for PhaseAccumulatorEstimator."""

    def test_smoothed(self):
        """Reference implementation: smoothed (test_ht_hd.xsl)."""
        inp = INPUT
        exp = EXPECTED_SMOOTHED
        pae = _create_default()

        lprimed = 3

        for i in range(lprimed):
            pae.update(inp[i])
            self.assertEqual(0, pae.smoothed(), msg=f"[{i}] before primed")

        for i in range(lprimed, len(inp)):
            pae.update(inp[i])
            self.assertAlmostEqual(exp[i], pae.smoothed(), delta=1e-8,
                                   msg=f"[{i}]")

        previous = pae.smoothed()
        pae.update(math.nan)
        self.assertEqual(previous, pae.smoothed(), msg="NaN input")

    def test_period(self):
        """Reference implementation: period (test_ht_hd.xsl)."""
        inp = INPUT
        exp = EXPECTED_PERIOD
        pae = _create_default()

        lprimed = 18
        last = 18

        for i in range(lprimed):
            pae.update(inp[i])
            self.assertEqual(6, pae.period(), msg=f"[{i}] before primed")

        for i in range(lprimed, last):
            pae.update(inp[i])
            self.assertAlmostEqual(exp[i], pae.period(), delta=1e-8,
                                   msg=f"[{i}]")

        previous = pae.period()
        pae.update(math.nan)
        self.assertEqual(previous, pae.period(), msg="NaN input")

    def test_detrended(self):
        """Reference implementation: detrended (test_ht_hd.xsl)."""
        inp = INPUT
        exp = EXPECTED_DETRENDED
        pae = _create_default()

        lprimed = 9
        last = 24

        for i in range(lprimed):
            pae.update(inp[i])
            self.assertEqual(0, pae.detrended(), msg=f"[{i}] before primed")

        for i in range(lprimed, last):
            pae.update(inp[i])
            self.assertAlmostEqual(exp[i], pae.detrended(), delta=1e-8,
                                   msg=f"[{i}]")

        previous = pae.detrended()
        pae.update(math.nan)
        self.assertEqual(previous, pae.detrended(), msg="NaN input")

    def test_quadrature(self):
        """Reference implementation: quadrature (test_ht_hd.xsl)."""
        inp = INPUT
        exp = EXPECTED_QUADRATURE
        pae = _create_default()

        lprimed = 15
        last = 24

        for i in range(lprimed):
            pae.update(inp[i])
            self.assertEqual(0, pae.quadrature(), msg=f"[{i}] before primed")

        for i in range(lprimed, last):
            pae.update(inp[i])
            self.assertAlmostEqual(exp[i], pae.quadrature(), delta=1e-8,
                                   msg=f"[{i}]")

        previous = pae.quadrature()
        pae.update(math.nan)
        self.assertEqual(previous, pae.quadrature(), msg="NaN input")

    def test_in_phase(self):
        """Reference implementation: in-phase (test_ht_hd.xsl)."""
        inp = INPUT
        exp = EXPECTED_IN_PHASE
        pae = _create_default()

        lprimed = 15
        last = 24

        for i in range(lprimed):
            pae.update(inp[i])
            self.assertEqual(0, pae.in_phase(), msg=f"[{i}] before primed")

        for i in range(lprimed, last):
            pae.update(inp[i])
            self.assertAlmostEqual(exp[i], pae.in_phase(), delta=1e-8,
                                   msg=f"[{i}]")

        previous = pae.in_phase()
        pae.update(math.nan)
        self.assertEqual(previous, pae.in_phase(), msg="NaN input")

    def test_primed(self):
        """Reference implementation: primed (test_ht_hd.xsl)."""
        inp = INPUT
        pae = _create_default()

        lprimed = 4 + 7 * 2  # 18

        self.assertFalse(pae.primed())

        for i in range(lprimed):
            pae.update(inp[i])
            self.assertFalse(pae.primed(), msg=f"[{i + 1}] should not be primed")

        for i in range(lprimed, len(inp)):
            pae.update(inp[i])
            self.assertTrue(pae.primed(), msg=f"[{i + 1}] should be primed")

    def test_primed_warmup(self):
        """Reference implementation: primed with warmup."""
        inp = INPUT
        lprimed = 50
        pae = _create_with_warmup(lprimed)

        self.assertFalse(pae.primed())

        for i in range(lprimed):
            pae.update(inp[i])
            self.assertFalse(pae.primed(), msg=f"[{i + 1}] should not be primed")

        for i in range(lprimed, len(inp)):
            pae.update(inp[i])
            self.assertTrue(pae.primed(), msg=f"[{i + 1}] should be primed")

    def test_sin_period(self):
        """Period of sin input with period=30."""
        period = 30
        omega = 2 * math.pi / period
        pae = _create_default()

        for i in range(512):
            pae.update(math.sin(omega * i))

        self.assertAlmostEqual(period, pae.period(), delta=1.0)

    def test_sin_min_period(self):
        """Period=3 sin input clamps to min_period."""
        period = 3
        omega = 2 * math.pi / period
        pae = _create_default()

        for i in range(512):
            pae.update(math.sin(omega * i))

        self.assertAlmostEqual(float(pae.min_period()), pae.period(), delta=1.5)

    def test_sin_max_period(self):
        """Period=60 sin input clamps to max_period."""
        period = 60
        omega = 2 * math.pi / period
        pae = _create_default()

        for i in range(512):
            pae.update(math.sin(omega * i))

        self.assertAlmostEqual(float(pae.max_period()), pae.period(), delta=12.5)

    def test_param_validation(self):
        """Various invalid params raise ValueError."""
        with self.assertRaises(ValueError):
            PhaseAccumulatorEstimator(CycleEstimatorParams(smoothing_length=1))
        with self.assertRaises(ValueError):
            PhaseAccumulatorEstimator(CycleEstimatorParams(smoothing_length=0))
        with self.assertRaises(ValueError):
            PhaseAccumulatorEstimator(CycleEstimatorParams(smoothing_length=-1))
        with self.assertRaises(ValueError):
            PhaseAccumulatorEstimator(CycleEstimatorParams(smoothing_length=5))

        with self.assertRaises(ValueError):
            PhaseAccumulatorEstimator(CycleEstimatorParams(
                alpha_ema_quadrature_in_phase=0.0))
        with self.assertRaises(ValueError):
            PhaseAccumulatorEstimator(CycleEstimatorParams(
                alpha_ema_quadrature_in_phase=-0.01))
        with self.assertRaises(ValueError):
            PhaseAccumulatorEstimator(CycleEstimatorParams(
                alpha_ema_quadrature_in_phase=1.0))
        with self.assertRaises(ValueError):
            PhaseAccumulatorEstimator(CycleEstimatorParams(
                alpha_ema_quadrature_in_phase=1.01))

        with self.assertRaises(ValueError):
            PhaseAccumulatorEstimator(CycleEstimatorParams(
                alpha_ema_period=0.0))
        with self.assertRaises(ValueError):
            PhaseAccumulatorEstimator(CycleEstimatorParams(
                alpha_ema_period=-0.01))
        with self.assertRaises(ValueError):
            PhaseAccumulatorEstimator(CycleEstimatorParams(
                alpha_ema_period=1.0))
        with self.assertRaises(ValueError):
            PhaseAccumulatorEstimator(CycleEstimatorParams(
                alpha_ema_period=1.01))


if __name__ == '__main__':
    unittest.main()
