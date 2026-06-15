import math
import unittest
from datetime import datetime

from py.indicators.don_mak.polynomial_forecast.polynomial_forecast import PolynomialForecast
from py.indicators.don_mak.polynomial_forecast.params import PolynomialForecastParams
from py.indicators.core.identifier import Identifier
from py.entities.scalar import Scalar

from . import test_testdata as td

TOLERANCE = 1e-9

# (degree, order, smoothing)
COMBOS = [
    (2, 1, 0), (2, 1, 3), (2, 1, 6), (2, 2, 0), (2, 2, 3), (2, 2, 6),
    (3, 1, 0), (3, 1, 3), (3, 1, 6), (3, 2, 0), (3, 2, 3), (3, 2, 6),
    (4, 1, 0), (4, 1, 3), (4, 1, 6), (4, 2, 0), (4, 2, 3), (4, 2, 6),
    (5, 1, 0), (5, 1, 3), (5, 1, 6), (5, 2, 0), (5, 2, 3), (5, 2, 6),
]


def _assert_series(test, actual, expected, label):
    test.assertEqual(len(actual), len(expected), f"{label}: length mismatch")
    for i in range(len(expected)):
        if math.isnan(expected[i]):
            test.assertTrue(math.isnan(actual[i]), f"{label}[{i}]: expected NaN, got {actual[i]}")
        else:
            test.assertAlmostEqual(actual[i], expected[i], delta=TOLERANCE,
                                   msg=f"{label}[{i}]: expected {expected[i]}, got {actual[i]}")


class TestPolynomialForecastData(unittest.TestCase):
    """Test POF against the reference test data for all parameter combos."""

    def test_combos(self):
        for degree, order, smoothing in COMBOS:
            name = f"EXPECTED_D{degree}_O{order}_S{smoothing}"
            expected = getattr(td, name)
            with self.subTest(name=name):
                ind = PolynomialForecast(
                    PolynomialForecastParams(degree=degree, order=order, smoothing=smoothing))
                values = [ind.update(c) for c in td.INPUT_CLOSE]
                _assert_series(self, values, expected, name)

    def test1_linear_o1(self):
        ind = PolynomialForecast(PolynomialForecastParams(degree=3, order=1, smoothing=0))
        values = [ind.update(c) for c in td.TEST1_INPUT_LINEAR]
        _assert_series(self, values, td.TEST1_EXPECTED_D3_O1_S0, "TEST1_O1")

    def test1_linear_o2(self):
        ind = PolynomialForecast(PolynomialForecastParams(degree=3, order=2, smoothing=0))
        values = [ind.update(c) for c in td.TEST1_INPUT_LINEAR]
        _assert_series(self, values, td.TEST1_EXPECTED_D3_O2_S0, "TEST1_O2")


class TestPolynomialForecastMnemonic(unittest.TestCase):
    def test_default_mnemonic(self):
        ind = PolynomialForecast(PolynomialForecastParams())
        self.assertEqual(ind.metadata().mnemonic, "pof(3,1,0)")

    def test_custom_mnemonic(self):
        ind = PolynomialForecast(PolynomialForecastParams(degree=5, order=2, smoothing=6))
        self.assertEqual(ind.metadata().mnemonic, "pof(5,2,6)")


class TestPolynomialForecastMetadata(unittest.TestCase):
    def test_default_metadata(self):
        ind = PolynomialForecast(PolynomialForecastParams())
        meta = ind.metadata()
        self.assertEqual(meta.identifier, Identifier.POLYNOMIAL_FORECAST)
        self.assertEqual(meta.mnemonic, "pof(3,1,0)")
        self.assertEqual(len(meta.outputs), 1)


class TestPolynomialForecastUpdateScalar(unittest.TestCase):
    def test_update_scalar(self):
        ind = PolynomialForecast(PolynomialForecastParams())
        tm = datetime(2021, 4, 1)
        out = None
        for c in td.INPUT_CLOSE:
            out = ind.update_scalar(Scalar(time=tm, value=c))
        self.assertEqual(len(out), 1)
        last = len(td.INPUT_CLOSE) - 1
        self.assertAlmostEqual(out[0].value, td.EXPECTED_D3_O1_S0[last], delta=TOLERANCE)


class TestPolynomialForecastInvalidParams(unittest.TestCase):
    def test_degree_too_small(self):
        with self.assertRaises(ValueError):
            PolynomialForecast(PolynomialForecastParams(degree=1))

    def test_order_too_small(self):
        with self.assertRaises(ValueError):
            PolynomialForecast(PolynomialForecastParams(order=0))

    def test_order_too_large(self):
        with self.assertRaises(ValueError):
            PolynomialForecast(PolynomialForecastParams(order=3))

    def test_smoothing_negative(self):
        with self.assertRaises(ValueError):
            PolynomialForecast(PolynomialForecastParams(smoothing=-1))


if __name__ == '__main__':
    unittest.main()
