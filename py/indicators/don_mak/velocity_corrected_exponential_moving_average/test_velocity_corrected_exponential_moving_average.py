import math
import unittest
from datetime import datetime

from py.indicators.don_mak.velocity_corrected_exponential_moving_average.velocity_corrected_exponential_moving_average import VelocityCorrectedExponentialMovingAverage
from py.indicators.don_mak.velocity_corrected_exponential_moving_average.params import VelocityCorrectedExponentialMovingAverageParams
from py.indicators.core.identifier import Identifier
from py.entities.scalar import Scalar

from . import test_testdata as td

TOLERANCE = 1e-9

# (period, degree)
COMBOS = [
    (3, 2), (3, 3), (3, 4), (3, 5),
    (6, 2), (6, 3), (6, 4), (6, 5),
    (12, 2), (12, 3), (12, 4), (12, 5),
]


def _assert_series(test, actual, expected, label):
    test.assertEqual(len(actual), len(expected), f"{label}: length mismatch")
    for i in range(len(expected)):
        if math.isnan(expected[i]):
            test.assertTrue(math.isnan(actual[i]), f"{label}[{i}]: expected NaN, got {actual[i]}")
        else:
            test.assertAlmostEqual(actual[i], expected[i], delta=TOLERANCE,
                                   msg=f"{label}[{i}]: expected {expected[i]}, got {actual[i]}")


class TestVelocityCorrectedExponentialMovingAverageData(unittest.TestCase):
    """Test VCEMA against the reference test data for all parameter combos."""

    def test_combos(self):
        for period, degree in COMBOS:
            name = f"EXPECTED_P{period}_D{degree}"
            expected = getattr(td, name)
            with self.subTest(name=name):
                ind = VelocityCorrectedExponentialMovingAverage(
                    VelocityCorrectedExponentialMovingAverageParams(period=period, degree=degree))
                values = [ind.update(c) for c in td.INPUT_CLOSE]
                _assert_series(self, values, expected, name)

    def test1_linear(self):
        ind = VelocityCorrectedExponentialMovingAverage(
            VelocityCorrectedExponentialMovingAverageParams(period=6, degree=3))
        values = [ind.update(c) for c in td.TEST1_INPUT_LINEAR]
        _assert_series(self, values, td.TEST1_EXPECTED_P6_D3, "TEST1")


class TestVelocityCorrectedExponentialMovingAverageMnemonic(unittest.TestCase):
    def test_default_mnemonic(self):
        ind = VelocityCorrectedExponentialMovingAverage(VelocityCorrectedExponentialMovingAverageParams())
        self.assertEqual(ind.metadata().mnemonic, "vcema(6,3)")

    def test_custom_mnemonic(self):
        ind = VelocityCorrectedExponentialMovingAverage(
            VelocityCorrectedExponentialMovingAverageParams(period=12, degree=5))
        self.assertEqual(ind.metadata().mnemonic, "vcema(12,5)")


class TestVelocityCorrectedExponentialMovingAverageMetadata(unittest.TestCase):
    def test_default_metadata(self):
        ind = VelocityCorrectedExponentialMovingAverage(VelocityCorrectedExponentialMovingAverageParams())
        meta = ind.metadata()
        self.assertEqual(meta.identifier, Identifier.VELOCITY_CORRECTED_EXPONENTIAL_MOVING_AVERAGE)
        self.assertEqual(meta.mnemonic, "vcema(6,3)")
        self.assertEqual(len(meta.outputs), 1)


class TestVelocityCorrectedExponentialMovingAverageUpdateScalar(unittest.TestCase):
    def test_update_scalar(self):
        ind = VelocityCorrectedExponentialMovingAverage(VelocityCorrectedExponentialMovingAverageParams())
        tm = datetime(2021, 4, 1)
        out = None
        for c in td.INPUT_CLOSE:
            out = ind.update_scalar(Scalar(time=tm, value=c))
        self.assertEqual(len(out), 1)
        last = len(td.INPUT_CLOSE) - 1
        self.assertAlmostEqual(out[0].value, td.EXPECTED_P6_D3[last], delta=TOLERANCE)


class TestVelocityCorrectedExponentialMovingAverageInvalidParams(unittest.TestCase):
    def test_period_too_small(self):
        with self.assertRaises(ValueError):
            VelocityCorrectedExponentialMovingAverage(VelocityCorrectedExponentialMovingAverageParams(period=1))

    def test_degree_too_small(self):
        with self.assertRaises(ValueError):
            VelocityCorrectedExponentialMovingAverage(VelocityCorrectedExponentialMovingAverageParams(degree=1))


if __name__ == '__main__':
    unittest.main()
