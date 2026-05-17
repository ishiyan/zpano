import math
import unittest
from datetime import datetime

from py.indicators.jean_philippe_poton.fractal_graph_dimension_index.fractal_graph_dimension_index import FractalGraphDimensionIndex
from py.indicators.jean_philippe_poton.fractal_graph_dimension_index.params import FractalGraphDimensionIndexParams
from py.indicators.core.identifier import Identifier
from py.entities.bar import Bar
from py.entities.scalar import Scalar

from .test_testdata import (
    INPUT_CLOSE,
    EXPECTED_FGDI_P5, EXPECTED_UPPER_P5, EXPECTED_LOWER_P5, EXPECTED_STDDEV_P5,
    EXPECTED_FGDI_P10, EXPECTED_UPPER_P10, EXPECTED_LOWER_P10, EXPECTED_STDDEV_P10,
    EXPECTED_FGDI_P15, EXPECTED_UPPER_P15, EXPECTED_LOWER_P15, EXPECTED_STDDEV_P15,
    EXPECTED_FGDI_P20, EXPECTED_UPPER_P20, EXPECTED_LOWER_P20, EXPECTED_STDDEV_P20,
    EXPECTED_FGDI_P30, EXPECTED_UPPER_P30, EXPECTED_LOWER_P30, EXPECTED_STDDEV_P30,
    EXPECTED_FGDI_P50, EXPECTED_UPPER_P50, EXPECTED_LOWER_P50, EXPECTED_STDDEV_P50,
    EXPECTED_FGDI_P80, EXPECTED_UPPER_P80, EXPECTED_LOWER_P80, EXPECTED_STDDEV_P80,
    EXPECTED_FGDI_P120, EXPECTED_UPPER_P120, EXPECTED_LOWER_P120, EXPECTED_STDDEV_P120,
)


class TestFractalGraphDimensionIndex(unittest.TestCase):

    def _run_test(self, period, exp_fgdi, exp_upper, exp_lower, exp_stddev):
        ind = FractalGraphDimensionIndex(FractalGraphDimensionIndexParams(period=period))
        for i, val in enumerate(INPUT_CLOSE):
            fgdi, upper, lower, stddev = ind.update(val)
            if math.isnan(exp_fgdi[i]):
                self.assertTrue(math.isnan(fgdi), f"index {i}: expected NaN for fgdi")
            else:
                self.assertAlmostEqual(fgdi, exp_fgdi[i], places=13, msg=f"fgdi at {i}")
            if math.isnan(exp_upper[i]):
                self.assertTrue(math.isnan(upper), f"index {i}: expected NaN for upper")
            else:
                self.assertAlmostEqual(upper, exp_upper[i], places=13, msg=f"upper at {i}")
            if math.isnan(exp_lower[i]):
                self.assertTrue(math.isnan(lower), f"index {i}: expected NaN for lower")
            else:
                self.assertAlmostEqual(lower, exp_lower[i], places=13, msg=f"lower at {i}")
            if math.isnan(exp_stddev[i]):
                self.assertTrue(math.isnan(stddev), f"index {i}: expected NaN for stddev")
            else:
                self.assertAlmostEqual(stddev, exp_stddev[i], places=13, msg=f"stddev at {i}")

    def test_period_5(self):
        self._run_test(5, EXPECTED_FGDI_P5, EXPECTED_UPPER_P5, EXPECTED_LOWER_P5, EXPECTED_STDDEV_P5)

    def test_period_10(self):
        self._run_test(10, EXPECTED_FGDI_P10, EXPECTED_UPPER_P10, EXPECTED_LOWER_P10, EXPECTED_STDDEV_P10)

    def test_period_15(self):
        self._run_test(15, EXPECTED_FGDI_P15, EXPECTED_UPPER_P15, EXPECTED_LOWER_P15, EXPECTED_STDDEV_P15)

    def test_period_20(self):
        self._run_test(20, EXPECTED_FGDI_P20, EXPECTED_UPPER_P20, EXPECTED_LOWER_P20, EXPECTED_STDDEV_P20)

    def test_period_30(self):
        self._run_test(30, EXPECTED_FGDI_P30, EXPECTED_UPPER_P30, EXPECTED_LOWER_P30, EXPECTED_STDDEV_P30)

    def test_period_50(self):
        self._run_test(50, EXPECTED_FGDI_P50, EXPECTED_UPPER_P50, EXPECTED_LOWER_P50, EXPECTED_STDDEV_P50)

    def test_period_80(self):
        self._run_test(80, EXPECTED_FGDI_P80, EXPECTED_UPPER_P80, EXPECTED_LOWER_P80, EXPECTED_STDDEV_P80)

    def test_period_120(self):
        self._run_test(120, EXPECTED_FGDI_P120, EXPECTED_UPPER_P120, EXPECTED_LOWER_P120, EXPECTED_STDDEV_P120)

    def test_is_primed(self):
        ind = FractalGraphDimensionIndex(FractalGraphDimensionIndexParams(period=30))
        for i in range(29):
            ind.update(INPUT_CLOSE[i])
            self.assertFalse(ind.is_primed())
        ind.update(INPUT_CLOSE[29])
        self.assertTrue(ind.is_primed())

    def test_nan_passthrough(self):
        ind = FractalGraphDimensionIndex(FractalGraphDimensionIndexParams(period=5))
        fgdi, upper, lower, stddev = ind.update(math.nan)
        self.assertTrue(math.isnan(fgdi))
        self.assertTrue(math.isnan(upper))
        self.assertTrue(math.isnan(lower))
        self.assertTrue(math.isnan(stddev))

    def test_invalid_period(self):
        with self.assertRaises(ValueError):
            FractalGraphDimensionIndex(FractalGraphDimensionIndexParams(period=1))

    def test_metadata(self):
        ind = FractalGraphDimensionIndex(FractalGraphDimensionIndexParams(period=30))
        meta = ind.metadata()
        self.assertEqual(meta.identifier, Identifier.FRACTAL_GRAPH_DIMENSION_INDEX)
        self.assertIn("fgdi(30)", meta.mnemonic)

    def test_update_bar(self):
        ind = FractalGraphDimensionIndex(FractalGraphDimensionIndexParams(period=5))
        for i, val in enumerate(INPUT_CLOSE):
            bar = Bar(datetime(2020, 1, 1), val + 1, val + 2, val - 1, val, 100.0)
            output = ind.update_bar(bar)
            expected = EXPECTED_FGDI_P5[i]
            if math.isnan(expected):
                self.assertTrue(math.isnan(output[0].value))
            else:
                self.assertAlmostEqual(output[0].value, expected, places=13)

    def test_update_scalar(self):
        ind = FractalGraphDimensionIndex(FractalGraphDimensionIndexParams(period=5))
        for i, val in enumerate(INPUT_CLOSE):
            scalar = Scalar(datetime(2020, 1, 1), val)
            output = ind.update_scalar(scalar)
            expected = EXPECTED_FGDI_P5[i]
            if math.isnan(expected):
                self.assertTrue(math.isnan(output[0].value))
            else:
                self.assertAlmostEqual(output[0].value, expected, places=13)


if __name__ == '__main__':
    unittest.main()
