import math
import unittest
from datetime import datetime

from py.indicators.don_mak.adaptive_exponential_moving_average.adaptive_exponential_moving_average import AdaptiveExponentialMovingAverage
from py.indicators.don_mak.adaptive_exponential_moving_average.params import AdaptiveExponentialMovingAverageParams
from py.indicators.core.identifier import Identifier
from py.entities.scalar import Scalar

from .test_testdata import (
    INPUT_CLOSE,
    EXPECTED_DEFAULT,
    EXPECTED_A0_8_A0_02,
    EXPECTED_W0_5,
    EXPECTED_W1_5,
    EXPECTED_S0,
    EXPECTED_S6,
    EXPECTED_DEFAULT_OMEGA,
    EXPECTED_DEFAULT_ALPHA,
    TEST1_INPUT_SINE,
    TEST1_EXPECTED,
    TEST1_EXPECTED_OMEGA,
    TEST1_EXPECTED_ALPHA,
)

TOLERANCE = 1e-9

# (alpha_max, alpha_min, omega_0, smoothing, expected_value)
VALUE_COMBOS = [
    (0.5, 0.05, 1.0, 3, EXPECTED_DEFAULT),
    (0.8, 0.02, 1.0, 3, EXPECTED_A0_8_A0_02),
    (0.5, 0.05, 0.5, 3, EXPECTED_W0_5),
    (0.5, 0.05, 1.5, 3, EXPECTED_W1_5),
    (0.5, 0.05, 1.0, 0, EXPECTED_S0),
    (0.5, 0.05, 1.0, 6, EXPECTED_S6),
]


def _assert_series(test, actual, expected, label):
    test.assertEqual(len(actual), len(expected), f"{label}: length mismatch")
    for i in range(len(expected)):
        if math.isnan(expected[i]):
            test.assertTrue(math.isnan(actual[i]), f"{label}[{i}]: expected NaN, got {actual[i]}")
        else:
            test.assertAlmostEqual(actual[i], expected[i], delta=TOLERANCE,
                                   msg=f"{label}[{i}]: expected {expected[i]}, got {actual[i]}")


class TestAdaptiveExponentialMovingAverageValue(unittest.TestCase):
    """Test AEMA value output against the reference for all parameter combos."""

    def test_value_combos(self):
        for alpha_max, alpha_min, omega_0, smoothing, expected in VALUE_COMBOS:
            with self.subTest(alpha_max=alpha_max, alpha_min=alpha_min, omega_0=omega_0, smoothing=smoothing):
                ind = AdaptiveExponentialMovingAverage(AdaptiveExponentialMovingAverageParams(
                    alpha_max=alpha_max, alpha_min=alpha_min, omega_0=omega_0, smoothing=smoothing))
                values = [ind.update(c)[0] for c in INPUT_CLOSE]
                _assert_series(self, values, expected, "value")


class TestAdaptiveExponentialMovingAverageOmegaAlpha(unittest.TestCase):
    """Test omega and alpha outputs for default params."""

    def test_default_omega_alpha(self):
        ind = AdaptiveExponentialMovingAverage(AdaptiveExponentialMovingAverageParams())
        omegas = []
        alphas = []
        for c in INPUT_CLOSE:
            _, omega, alpha = ind.update(c)
            omegas.append(omega)
            alphas.append(alpha)
        _assert_series(self, omegas, EXPECTED_DEFAULT_OMEGA, "omega")
        _assert_series(self, alphas, EXPECTED_DEFAULT_ALPHA, "alpha")


class TestAdaptiveExponentialMovingAverageSine(unittest.TestCase):
    """Test all three outputs on the pure sine-wave series (default params)."""

    def test_sine(self):
        ind = AdaptiveExponentialMovingAverage(AdaptiveExponentialMovingAverageParams())
        values, omegas, alphas = [], [], []
        for c in TEST1_INPUT_SINE:
            v, o, a = ind.update(c)
            values.append(v)
            omegas.append(o)
            alphas.append(a)
        _assert_series(self, values, TEST1_EXPECTED, "value")
        _assert_series(self, omegas, TEST1_EXPECTED_OMEGA, "omega")
        _assert_series(self, alphas, TEST1_EXPECTED_ALPHA, "alpha")


class TestAdaptiveExponentialMovingAverageMnemonic(unittest.TestCase):
    def test_default_mnemonic(self):
        ind = AdaptiveExponentialMovingAverage(AdaptiveExponentialMovingAverageParams())
        self.assertEqual(ind.metadata().mnemonic, "aema(0.50,0.05,1.00,3)")

    def test_custom_mnemonic(self):
        ind = AdaptiveExponentialMovingAverage(AdaptiveExponentialMovingAverageParams(
            alpha_max=0.8, alpha_min=0.02, omega_0=1.5, smoothing=6))
        self.assertEqual(ind.metadata().mnemonic, "aema(0.80,0.02,1.50,6)")


class TestAdaptiveExponentialMovingAverageMetadata(unittest.TestCase):
    def test_default_metadata(self):
        ind = AdaptiveExponentialMovingAverage(AdaptiveExponentialMovingAverageParams())
        meta = ind.metadata()
        self.assertEqual(meta.identifier, Identifier.ADAPTIVE_EXPONENTIAL_MOVING_AVERAGE)
        self.assertEqual(meta.mnemonic, "aema(0.50,0.05,1.00,3)")
        self.assertEqual(len(meta.outputs), 3)


class TestAdaptiveExponentialMovingAverageUpdateScalar(unittest.TestCase):
    def test_update_scalar(self):
        ind = AdaptiveExponentialMovingAverage(AdaptiveExponentialMovingAverageParams())
        tm = datetime(2021, 4, 1)
        out = None
        for c in INPUT_CLOSE:
            out = ind.update_scalar(Scalar(time=tm, value=c))
        self.assertEqual(len(out), 3)
        self.assertAlmostEqual(out[0].value, EXPECTED_DEFAULT[-1], delta=TOLERANCE)
        self.assertAlmostEqual(out[2].value, EXPECTED_DEFAULT_ALPHA[-1], delta=TOLERANCE)


class TestAdaptiveExponentialMovingAverageInvalidParams(unittest.TestCase):
    def test_alpha_order(self):
        with self.assertRaises(ValueError):
            AdaptiveExponentialMovingAverage(AdaptiveExponentialMovingAverageParams(alpha_max=0.05, alpha_min=0.5))

    def test_alpha_max_too_large(self):
        with self.assertRaises(ValueError):
            AdaptiveExponentialMovingAverage(AdaptiveExponentialMovingAverageParams(alpha_max=1.5))

    def test_alpha_min_zero(self):
        with self.assertRaises(ValueError):
            AdaptiveExponentialMovingAverage(AdaptiveExponentialMovingAverageParams(alpha_min=0.0))

    def test_omega_0_too_large(self):
        with self.assertRaises(ValueError):
            AdaptiveExponentialMovingAverage(AdaptiveExponentialMovingAverageParams(omega_0=4.0))

    def test_smoothing_negative(self):
        with self.assertRaises(ValueError):
            AdaptiveExponentialMovingAverage(AdaptiveExponentialMovingAverageParams(smoothing=-1))


if __name__ == '__main__':
    unittest.main()
