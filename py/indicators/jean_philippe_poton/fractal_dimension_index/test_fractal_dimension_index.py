import math
import unittest
from datetime import datetime

from py.indicators.jean_philippe_poton.fractal_dimension_index.fractal_dimension_index import FractalDimensionIndex
from py.indicators.jean_philippe_poton.fractal_dimension_index.params import FractalDimensionIndexParams
from py.indicators.core.identifier import Identifier
from py.entities.bar import Bar
from py.entities.quote import Quote
from py.entities.trade import Trade
from py.entities.scalar import Scalar

from .test_testdata import (
    INPUT_CLOSE,
    EXPECTED_P5,
    EXPECTED_P10,
    EXPECTED_P15,
    EXPECTED_P20,
    EXPECTED_P30,
    EXPECTED_P50,
    EXPECTED_P80,
    EXPECTED_P120,
)


class TestFractalDimensionIndex(unittest.TestCase):

    def test_update_period_5(self):
        fdi = FractalDimensionIndex(FractalDimensionIndexParams(period=5))
        for i, val in enumerate(INPUT_CLOSE):
            result = fdi.update(val)
            expected = EXPECTED_P5[i]
            if math.isnan(expected):
                self.assertTrue(math.isnan(result))
            else:
                self.assertAlmostEqual(result, expected, places=13)

    def test_update_period_10(self):
        fdi = FractalDimensionIndex(FractalDimensionIndexParams(period=10))
        for i, val in enumerate(INPUT_CLOSE):
            result = fdi.update(val)
            expected = EXPECTED_P10[i]
            if math.isnan(expected):
                self.assertTrue(math.isnan(result))
            else:
                self.assertAlmostEqual(result, expected, places=13)

    def test_update_period_15(self):
        fdi = FractalDimensionIndex(FractalDimensionIndexParams(period=15))
        for i, val in enumerate(INPUT_CLOSE):
            result = fdi.update(val)
            expected = EXPECTED_P15[i]
            if math.isnan(expected):
                self.assertTrue(math.isnan(result))
            else:
                self.assertAlmostEqual(result, expected, places=13)

    def test_update_period_20(self):
        fdi = FractalDimensionIndex(FractalDimensionIndexParams(period=20))
        for i, val in enumerate(INPUT_CLOSE):
            result = fdi.update(val)
            expected = EXPECTED_P20[i]
            if math.isnan(expected):
                self.assertTrue(math.isnan(result))
            else:
                self.assertAlmostEqual(result, expected, places=13)

    def test_update_period_30(self):
        fdi = FractalDimensionIndex(FractalDimensionIndexParams(period=30))
        for i, val in enumerate(INPUT_CLOSE):
            result = fdi.update(val)
            expected = EXPECTED_P30[i]
            if math.isnan(expected):
                self.assertTrue(math.isnan(result))
            else:
                self.assertAlmostEqual(result, expected, places=13)

    def test_update_period_50(self):
        fdi = FractalDimensionIndex(FractalDimensionIndexParams(period=50))
        for i, val in enumerate(INPUT_CLOSE):
            result = fdi.update(val)
            expected = EXPECTED_P50[i]
            if math.isnan(expected):
                self.assertTrue(math.isnan(result))
            else:
                self.assertAlmostEqual(result, expected, places=13)

    def test_update_period_80(self):
        fdi = FractalDimensionIndex(FractalDimensionIndexParams(period=80))
        for i, val in enumerate(INPUT_CLOSE):
            result = fdi.update(val)
            expected = EXPECTED_P80[i]
            if math.isnan(expected):
                self.assertTrue(math.isnan(result))
            else:
                self.assertAlmostEqual(result, expected, places=13)

    def test_update_period_120(self):
        fdi = FractalDimensionIndex(FractalDimensionIndexParams(period=120))
        for i, val in enumerate(INPUT_CLOSE):
            result = fdi.update(val)
            expected = EXPECTED_P120[i]
            if math.isnan(expected):
                self.assertTrue(math.isnan(result))
            else:
                self.assertAlmostEqual(result, expected, places=13)

    def test_is_primed_period_30(self):
        fdi = FractalDimensionIndex(FractalDimensionIndexParams(period=30))
        for i in range(30):
            fdi.update(INPUT_CLOSE[i])
            self.assertFalse(fdi.is_primed())
        fdi.update(INPUT_CLOSE[30])
        self.assertTrue(fdi.is_primed())

    def test_nan_passthrough(self):
        fdi = FractalDimensionIndex(FractalDimensionIndexParams(period=5))
        result = fdi.update(math.nan)
        self.assertTrue(math.isnan(result))

    def test_invalid_period(self):
        with self.assertRaises(ValueError):
            FractalDimensionIndex(FractalDimensionIndexParams(period=1))

    def test_metadata(self):
        fdi = FractalDimensionIndex(FractalDimensionIndexParams(period=30))
        meta = fdi.metadata()
        self.assertEqual(meta.identifier, Identifier.FRACTAL_DIMENSION_INDEX)
        self.assertIn("fdi(30)", meta.mnemonic)

    def test_update_bar(self):
        fdi = FractalDimensionIndex(FractalDimensionIndexParams(period=5))
        for i, val in enumerate(INPUT_CLOSE):
            bar = Bar(datetime(2020, 1, 1), val + 1, val + 2, val - 1, val, 100.0)
            output = fdi.update_bar(bar)
            expected = EXPECTED_P5[i]
            if math.isnan(expected):
                self.assertTrue(math.isnan(output[0].value))
            else:
                self.assertAlmostEqual(output[0].value, expected, places=13)

    def test_update_scalar(self):
        fdi = FractalDimensionIndex(FractalDimensionIndexParams(period=5))
        for i, val in enumerate(INPUT_CLOSE):
            scalar = Scalar(datetime(2020, 1, 1), val)
            output = fdi.update_scalar(scalar)
            expected = EXPECTED_P5[i]
            if math.isnan(expected):
                self.assertTrue(math.isnan(output[0].value))
            else:
                self.assertAlmostEqual(output[0].value, expected, places=13)


if __name__ == '__main__':
    unittest.main()
