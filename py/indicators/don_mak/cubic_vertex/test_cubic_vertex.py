import math
import unittest
from datetime import datetime

from py.indicators.don_mak.cubic_vertex.cubic_vertex import CubicVertex
from py.indicators.don_mak.cubic_vertex.params import CubicVertexParams
from py.indicators.core.identifier import Identifier
from py.entities.scalar import Scalar

from . import test_testdata as td

TOLERANCE = 1e-9


def _assert_series(test, actual, expected, label):
    test.assertEqual(len(actual), len(expected), f"{label}: length mismatch")
    for i in range(len(expected)):
        if math.isnan(expected[i]):
            test.assertTrue(math.isnan(actual[i]), f"{label}[{i}]: expected NaN, got {actual[i]}")
        else:
            # Combined absolute + relative tolerance. Near degenerate points the vertex
            # locations are ill-conditioned (3c -> 0, disc -> 0), so the output magnitude
            # can be large; a relative tolerance preserves 13+ significant-digit agreement.
            delta = TOLERANCE * max(1.0, abs(expected[i]))
            test.assertAlmostEqual(actual[i], expected[i], delta=delta,
                                   msg=f"{label}[{i}]: expected {expected[i]}, got {actual[i]}")


class TestCubicVertexData(unittest.TestCase):
    """Test CVTX against the reference test data."""

    def _run(self, inputs):
        ind = CubicVertex(CubicVertexParams())
        near = []
        far = []
        for c in inputs:
            n, f = ind.update(c)
            near.append(n)
            far.append(f)
        return near, far

    def test_raw(self):
        near, far = self._run(td.INPUT_CLOSE)
        _assert_series(self, near, td.EXPECTED_RAW_NEAR, "EXPECTED_RAW_NEAR")
        _assert_series(self, far, td.EXPECTED_RAW_FAR, "EXPECTED_RAW_FAR")

    def test_ema6(self):
        near, far = self._run(td.INPUT_EMA6)
        _assert_series(self, near, td.EXPECTED_EMA6_NEAR, "EXPECTED_EMA6_NEAR")
        _assert_series(self, far, td.EXPECTED_EMA6_FAR, "EXPECTED_EMA6_FAR")

    def test_ema20(self):
        near, far = self._run(td.INPUT_EMA20)
        _assert_series(self, near, td.EXPECTED_EMA20_NEAR, "EXPECTED_EMA20_NEAR")
        _assert_series(self, far, td.EXPECTED_EMA20_FAR, "EXPECTED_EMA20_FAR")

    def test1_cubic(self):
        near, far = self._run(td.TEST1_INPUT_CUBIC)
        _assert_series(self, near, td.TEST1_EXPECTED_NEAR, "TEST1_NEAR")
        _assert_series(self, far, td.TEST1_EXPECTED_FAR, "TEST1_FAR")


class TestCubicVertexMnemonic(unittest.TestCase):
    def test_default_mnemonic(self):
        ind = CubicVertex(CubicVertexParams())
        self.assertEqual(ind.metadata().mnemonic, "cvtx")


class TestCubicVertexMetadata(unittest.TestCase):
    def test_default_metadata(self):
        ind = CubicVertex(CubicVertexParams())
        meta = ind.metadata()
        self.assertEqual(meta.identifier, Identifier.CUBIC_VERTEX)
        self.assertEqual(meta.mnemonic, "cvtx")
        self.assertEqual(len(meta.outputs), 2)


class TestCubicVertexPriming(unittest.TestCase):
    def test_priming(self):
        ind = CubicVertex(CubicVertexParams())
        for _ in range(3):
            n, f = ind.update(1.0)
            self.assertTrue(math.isnan(n) and math.isnan(f))
            self.assertFalse(ind.is_primed())
        ind.update(2.0)
        self.assertTrue(ind.is_primed())

    def test_linear_returns_nan(self):
        ind = CubicVertex(CubicVertexParams())
        # Four collinear points -> c == 0 and d == 0 -> both NaN.
        n = f = None
        for v in (1.0, 2.0, 3.0, 4.0):
            n, f = ind.update(v)
        self.assertTrue(math.isnan(n) and math.isnan(f))
        self.assertTrue(ind.is_primed())


class TestCubicVertexUpdateScalar(unittest.TestCase):
    def test_update_scalar(self):
        ind = CubicVertex(CubicVertexParams())
        tm = datetime(2021, 4, 1)
        out = None
        for c in td.INPUT_CLOSE:
            out = ind.update_scalar(Scalar(time=tm, value=c))
        self.assertEqual(len(out), 2)


if __name__ == '__main__':
    unittest.main()
