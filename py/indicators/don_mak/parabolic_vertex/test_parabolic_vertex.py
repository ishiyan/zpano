import math
import unittest
from datetime import datetime

from py.indicators.don_mak.parabolic_vertex.parabolic_vertex import ParabolicVertex
from py.indicators.don_mak.parabolic_vertex.params import ParabolicVertexParams
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
            # Combined absolute + relative tolerance. Near collinear points the vertex
            # location is ill-conditioned (denom -> 0), so the output magnitude can be
            # large; a relative tolerance preserves 13+ significant-digit agreement.
            delta = TOLERANCE * max(1.0, abs(expected[i]))
            test.assertAlmostEqual(actual[i], expected[i], delta=delta,
                                   msg=f"{label}[{i}]: expected {expected[i]}, got {actual[i]}")


class TestParabolicVertexData(unittest.TestCase):
    """Test PVTX against the reference test data."""

    def _run(self, inputs):
        ind = ParabolicVertex(ParabolicVertexParams())
        return [ind.update(c) for c in inputs]

    def test_raw(self):
        _assert_series(self, self._run(td.INPUT_CLOSE), td.EXPECTED_RAW, "EXPECTED_RAW")

    def test_ema6(self):
        _assert_series(self, self._run(td.INPUT_EMA6), td.EXPECTED_EMA6, "EXPECTED_EMA6")

    def test_ema20(self):
        _assert_series(self, self._run(td.INPUT_EMA20), td.EXPECTED_EMA20, "EXPECTED_EMA20")

    def test1_parabola(self):
        _assert_series(self, self._run(td.TEST1_INPUT_PARABOLA), td.TEST1_EXPECTED, "TEST1")


class TestParabolicVertexMnemonic(unittest.TestCase):
    def test_default_mnemonic(self):
        ind = ParabolicVertex(ParabolicVertexParams())
        self.assertEqual(ind.metadata().mnemonic, "pvtx")


class TestParabolicVertexMetadata(unittest.TestCase):
    def test_default_metadata(self):
        ind = ParabolicVertex(ParabolicVertexParams())
        meta = ind.metadata()
        self.assertEqual(meta.identifier, Identifier.PARABOLIC_VERTEX)
        self.assertEqual(meta.mnemonic, "pvtx")
        self.assertEqual(len(meta.outputs), 1)


class TestParabolicVertexPriming(unittest.TestCase):
    def test_priming(self):
        ind = ParabolicVertex(ParabolicVertexParams())
        self.assertTrue(math.isnan(ind.update(1.0)))
        self.assertFalse(ind.is_primed())
        self.assertTrue(math.isnan(ind.update(2.0)))
        self.assertFalse(ind.is_primed())
        ind.update(3.0)
        self.assertTrue(ind.is_primed())

    def test_collinear_returns_nan(self):
        ind = ParabolicVertex(ParabolicVertexParams())
        ind.update(1.0)
        ind.update(2.0)
        # Three collinear points -> zero curvature -> NaN.
        self.assertTrue(math.isnan(ind.update(3.0)))
        self.assertTrue(ind.is_primed())


class TestParabolicVertexUpdateScalar(unittest.TestCase):
    def test_update_scalar(self):
        ind = ParabolicVertex(ParabolicVertexParams())
        tm = datetime(2021, 4, 1)
        out = None
        for c in td.INPUT_CLOSE:
            out = ind.update_scalar(Scalar(time=tm, value=c))
        self.assertEqual(len(out), 1)
        last = len(td.INPUT_CLOSE) - 1
        if math.isnan(td.EXPECTED_RAW[last]):
            self.assertTrue(math.isnan(out[0].value))
        else:
            self.assertAlmostEqual(out[0].value, td.EXPECTED_RAW[last], delta=TOLERANCE)


if __name__ == '__main__':
    unittest.main()
