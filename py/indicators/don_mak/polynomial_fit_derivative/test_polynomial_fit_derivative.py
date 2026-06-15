import math
import unittest
from datetime import datetime

from py.indicators.don_mak.polynomial_fit_derivative.polynomial_fit_derivative import PolynomialFitDerivative
from py.indicators.don_mak.polynomial_fit_derivative.params import PolynomialFitDerivativeParams
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
    (6, 1, 0), (6, 1, 3), (6, 1, 6), (6, 2, 0), (6, 2, 3), (6, 2, 6),
    (4, 3, 6), (5, 3, 6), (6, 3, 6), (6, 5, 6),
]


class TestPolynomialFitDerivativeData(unittest.TestCase):
    """Test PFD against the reference test data for all parameter combos."""

    def test_combos(self):
        for degree, order, smoothing in COMBOS:
            name = f"EXPECTED_D{degree}_O{order}_S{smoothing}"
            expected = getattr(td, name)
            with self.subTest(name=name):
                ind = PolynomialFitDerivative(PolynomialFitDerivativeParams(
                    degree=degree, order=order, smoothing=smoothing))
                self.assertEqual(len(td.INPUT_CLOSE), len(expected))
                for i, c in enumerate(td.INPUT_CLOSE):
                    value = ind.update(c)
                    if math.isnan(expected[i]):
                        self.assertTrue(math.isnan(value), f"{name}[{i}]: expected NaN, got {value}")
                    else:
                        self.assertAlmostEqual(value, expected[i], delta=TOLERANCE,
                                               msg=f"{name}[{i}]: expected {expected[i]}, got {value}")


class TestPolynomialFitDerivativeMnemonic(unittest.TestCase):
    def test_default_mnemonic(self):
        ind = PolynomialFitDerivative(PolynomialFitDerivativeParams())
        self.assertEqual(ind.metadata().mnemonic, "pfd(3,1,6)")

    def test_custom_mnemonic(self):
        ind = PolynomialFitDerivative(PolynomialFitDerivativeParams(degree=4, order=2, smoothing=3))
        self.assertEqual(ind.metadata().mnemonic, "pfd(4,2,3)")


class TestPolynomialFitDerivativeMetadata(unittest.TestCase):
    def test_default_metadata(self):
        ind = PolynomialFitDerivative(PolynomialFitDerivativeParams())
        meta = ind.metadata()
        self.assertEqual(meta.identifier, Identifier.POLYNOMIAL_FIT_DERIVATIVE)
        self.assertEqual(meta.mnemonic, "pfd(3,1,6)")
        self.assertEqual(len(meta.outputs), 1)


class TestPolynomialFitDerivativeUpdateScalar(unittest.TestCase):
    def test_update_scalar(self):
        ind = PolynomialFitDerivative(PolynomialFitDerivativeParams())
        tm = datetime(2021, 4, 1)
        out = None
        for c in td.INPUT_CLOSE:
            out = ind.update_scalar(Scalar(time=tm, value=c))
        self.assertEqual(len(out), 1)
        last = len(td.INPUT_CLOSE) - 1
        expected = td.EXPECTED_D3_O1_S6[last]
        self.assertAlmostEqual(out[0].value, expected, delta=TOLERANCE)


class TestPolynomialFitDerivativeInvalidParams(unittest.TestCase):
    def test_degree_too_small(self):
        with self.assertRaises(ValueError):
            PolynomialFitDerivative(PolynomialFitDerivativeParams(degree=1))

    def test_order_too_small(self):
        with self.assertRaises(ValueError):
            PolynomialFitDerivative(PolynomialFitDerivativeParams(order=0))

    def test_order_gt_degree(self):
        with self.assertRaises(ValueError):
            PolynomialFitDerivative(PolynomialFitDerivativeParams(degree=3, order=4))

    def test_smoothing_negative(self):
        with self.assertRaises(ValueError):
            PolynomialFitDerivative(PolynomialFitDerivativeParams(smoothing=-1))


if __name__ == '__main__':
    unittest.main()
