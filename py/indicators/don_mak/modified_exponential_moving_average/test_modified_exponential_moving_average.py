import math
import unittest
from datetime import datetime

from py.indicators.don_mak.modified_exponential_moving_average.modified_exponential_moving_average import ModifiedExponentialMovingAverage
from py.indicators.don_mak.modified_exponential_moving_average.params import ModifiedExponentialMovingAverageParams
from py.indicators.core.identifier import Identifier
from py.entities.scalar import Scalar

from . import test_testdata as td

TOLERANCE = 1e-9

# (period, degree, skip)
COMBOS = [
    (3, 3, 1), (3, 3, 2), (3, 3, 4),
    (3, 4, 1), (3, 4, 2), (3, 4, 4),
    (6, 3, 1), (6, 3, 2), (6, 3, 4),
    (6, 4, 1), (6, 4, 2), (6, 4, 4),
    (12, 3, 1), (12, 3, 2), (12, 3, 4),
    (12, 4, 1), (12, 4, 2), (12, 4, 4),
]


def _assert_series(test, actual, expected, label):
    test.assertEqual(len(actual), len(expected), f"{label}: length mismatch")
    for i in range(len(expected)):
        if math.isnan(expected[i]):
            test.assertTrue(math.isnan(actual[i]), f"{label}[{i}]: expected NaN, got {actual[i]}")
        else:
            test.assertAlmostEqual(actual[i], expected[i], delta=TOLERANCE,
                                   msg=f"{label}[{i}]: expected {expected[i]}, got {actual[i]}")


class TestModifiedExponentialMovingAverageData(unittest.TestCase):
    """Test MEMA against the reference test data for all parameter combos."""

    def test_combos(self):
        for period, degree, skip in COMBOS:
            name = f"EXPECTED_P{period}_D{degree}_SK{skip}"
            expected = getattr(td, name)
            with self.subTest(name=name):
                ind = ModifiedExponentialMovingAverage(ModifiedExponentialMovingAverageParams(
                    period=period, degree=degree, skip=skip))
                values = [ind.update(c) for c in td.INPUT_CLOSE]
                _assert_series(self, values, expected, name)

    def test1_linear(self):
        ind = ModifiedExponentialMovingAverage(ModifiedExponentialMovingAverageParams(period=6, degree=3, skip=1))
        values = [ind.update(c) for c in td.TEST1_INPUT_LINEAR]
        _assert_series(self, values, td.TEST1_EXPECTED_P6_D3_SK1, "TEST1")


class TestModifiedExponentialMovingAverageMnemonic(unittest.TestCase):
    def test_default_mnemonic(self):
        ind = ModifiedExponentialMovingAverage(ModifiedExponentialMovingAverageParams())
        self.assertEqual(ind.metadata().mnemonic, "mema(6,3,1)")

    def test_custom_mnemonic(self):
        ind = ModifiedExponentialMovingAverage(ModifiedExponentialMovingAverageParams(period=12, degree=4, skip=2))
        self.assertEqual(ind.metadata().mnemonic, "mema(12,4,2)")


class TestModifiedExponentialMovingAverageMetadata(unittest.TestCase):
    def test_default_metadata(self):
        ind = ModifiedExponentialMovingAverage(ModifiedExponentialMovingAverageParams())
        meta = ind.metadata()
        self.assertEqual(meta.identifier, Identifier.MODIFIED_EXPONENTIAL_MOVING_AVERAGE)
        self.assertEqual(meta.mnemonic, "mema(6,3,1)")
        self.assertEqual(len(meta.outputs), 1)


class TestModifiedExponentialMovingAverageUpdateScalar(unittest.TestCase):
    def test_update_scalar(self):
        ind = ModifiedExponentialMovingAverage(ModifiedExponentialMovingAverageParams())
        tm = datetime(2021, 4, 1)
        out = None
        for c in td.INPUT_CLOSE:
            out = ind.update_scalar(Scalar(time=tm, value=c))
        self.assertEqual(len(out), 1)
        last = len(td.INPUT_CLOSE) - 1
        self.assertAlmostEqual(out[0].value, td.EXPECTED_P6_D3_SK1[last], delta=TOLERANCE)


class TestModifiedExponentialMovingAverageInvalidParams(unittest.TestCase):
    def test_period_too_small(self):
        with self.assertRaises(ValueError):
            ModifiedExponentialMovingAverage(ModifiedExponentialMovingAverageParams(period=1))

    def test_degree_too_small(self):
        with self.assertRaises(ValueError):
            ModifiedExponentialMovingAverage(ModifiedExponentialMovingAverageParams(degree=1))

    def test_skip_too_small(self):
        with self.assertRaises(ValueError):
            ModifiedExponentialMovingAverage(ModifiedExponentialMovingAverageParams(skip=0))


if __name__ == '__main__':
    unittest.main()
