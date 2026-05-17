import math
import unittest
from datetime import datetime

from py.indicators.jean_philippe_poton.hurst_difference.hurst_difference import HurstDifference
from py.indicators.jean_philippe_poton.hurst_difference.params import HurstDifferenceParams
from py.indicators.core.identifier import Identifier
from py.entities.bar import Bar
from py.entities.quote import Quote
from py.entities.trade import Trade
from py.entities.scalar import Scalar

from .test_testdata import (
    INPUT_CLOSE,
    EXPECTED_FGDI_P5,
    EXPECTED_HDIFF_P5,
    EXPECTED_FGDI_P10,
    EXPECTED_HDIFF_P10,
    EXPECTED_FGDI_P15,
    EXPECTED_HDIFF_P15,
    EXPECTED_FGDI_P20,
    EXPECTED_HDIFF_P20,
    EXPECTED_FGDI_P30,
    EXPECTED_HDIFF_P30,
    EXPECTED_FGDI_P50,
    EXPECTED_HDIFF_P50,
    EXPECTED_FGDI_P80,
    EXPECTED_HDIFF_P80,
    EXPECTED_FGDI_P120,
    EXPECTED_HDIFF_P120,
)


class TestHurstDifference(unittest.TestCase):

    def test_update_period_5(self):
        ind = HurstDifference(HurstDifferenceParams(period=5))
        for i, val in enumerate(INPUT_CLOSE):
            hurst_diff, fgdi = ind.update(val)
            exp_fgdi = EXPECTED_FGDI_P5[i]
            exp_hdiff = EXPECTED_HDIFF_P5[i]
            if math.isnan(exp_fgdi):
                self.assertTrue(math.isnan(fgdi), f"[{i}] fgdi expected NaN, got {fgdi}")
            else:
                self.assertAlmostEqual(fgdi, exp_fgdi, places=13, msg=f"[{i}] fgdi")
            if math.isnan(exp_hdiff):
                self.assertTrue(math.isnan(hurst_diff), f"[{i}] hdiff expected NaN, got {hurst_diff}")
            else:
                self.assertAlmostEqual(hurst_diff, exp_hdiff, places=13, msg=f"[{i}] hdiff")

    def test_update_period_30(self):
        ind = HurstDifference(HurstDifferenceParams(period=30))
        for i, val in enumerate(INPUT_CLOSE):
            hurst_diff, fgdi = ind.update(val)
            exp_fgdi = EXPECTED_FGDI_P30[i]
            exp_hdiff = EXPECTED_HDIFF_P30[i]
            if math.isnan(exp_fgdi):
                self.assertTrue(math.isnan(fgdi), f"[{i}] fgdi expected NaN, got {fgdi}")
            else:
                self.assertAlmostEqual(fgdi, exp_fgdi, places=13, msg=f"[{i}] fgdi")
            if math.isnan(exp_hdiff):
                self.assertTrue(math.isnan(hurst_diff), f"[{i}] hdiff expected NaN, got {hurst_diff}")
            else:
                self.assertAlmostEqual(hurst_diff, exp_hdiff, places=13, msg=f"[{i}] hdiff")

    def test_update_period_10(self):
        ind = HurstDifference(HurstDifferenceParams(period=10))
        for i, val in enumerate(INPUT_CLOSE):
            hurst_diff, fgdi = ind.update(val)
            exp_fgdi = EXPECTED_FGDI_P10[i]
            exp_hdiff = EXPECTED_HDIFF_P10[i]
            if math.isnan(exp_fgdi):
                self.assertTrue(math.isnan(fgdi), f"[{i}] fgdi expected NaN, got {fgdi}")
            else:
                self.assertAlmostEqual(fgdi, exp_fgdi, places=13, msg=f"[{i}] fgdi")
            if math.isnan(exp_hdiff):
                self.assertTrue(math.isnan(hurst_diff), f"[{i}] hdiff expected NaN, got {hurst_diff}")
            else:
                self.assertAlmostEqual(hurst_diff, exp_hdiff, places=13, msg=f"[{i}] hdiff")

    def test_update_period_15(self):
        ind = HurstDifference(HurstDifferenceParams(period=15))
        for i, val in enumerate(INPUT_CLOSE):
            hurst_diff, fgdi = ind.update(val)
            exp_fgdi = EXPECTED_FGDI_P15[i]
            exp_hdiff = EXPECTED_HDIFF_P15[i]
            if math.isnan(exp_fgdi):
                self.assertTrue(math.isnan(fgdi), f"[{i}] fgdi expected NaN, got {fgdi}")
            else:
                self.assertAlmostEqual(fgdi, exp_fgdi, places=13, msg=f"[{i}] fgdi")
            if math.isnan(exp_hdiff):
                self.assertTrue(math.isnan(hurst_diff), f"[{i}] hdiff expected NaN, got {hurst_diff}")
            else:
                self.assertAlmostEqual(hurst_diff, exp_hdiff, places=13, msg=f"[{i}] hdiff")

    def test_update_period_20(self):
        ind = HurstDifference(HurstDifferenceParams(period=20))
        for i, val in enumerate(INPUT_CLOSE):
            hurst_diff, fgdi = ind.update(val)
            exp_fgdi = EXPECTED_FGDI_P20[i]
            exp_hdiff = EXPECTED_HDIFF_P20[i]
            if math.isnan(exp_fgdi):
                self.assertTrue(math.isnan(fgdi), f"[{i}] fgdi expected NaN, got {fgdi}")
            else:
                self.assertAlmostEqual(fgdi, exp_fgdi, places=13, msg=f"[{i}] fgdi")
            if math.isnan(exp_hdiff):
                self.assertTrue(math.isnan(hurst_diff), f"[{i}] hdiff expected NaN, got {hurst_diff}")
            else:
                self.assertAlmostEqual(hurst_diff, exp_hdiff, places=13, msg=f"[{i}] hdiff")

    def test_update_period_50(self):
        ind = HurstDifference(HurstDifferenceParams(period=50))
        for i, val in enumerate(INPUT_CLOSE):
            hurst_diff, fgdi = ind.update(val)
            exp_fgdi = EXPECTED_FGDI_P50[i]
            exp_hdiff = EXPECTED_HDIFF_P50[i]
            if math.isnan(exp_fgdi):
                self.assertTrue(math.isnan(fgdi), f"[{i}] fgdi expected NaN, got {fgdi}")
            else:
                self.assertAlmostEqual(fgdi, exp_fgdi, places=13, msg=f"[{i}] fgdi")
            if math.isnan(exp_hdiff):
                self.assertTrue(math.isnan(hurst_diff), f"[{i}] hdiff expected NaN, got {hurst_diff}")
            else:
                self.assertAlmostEqual(hurst_diff, exp_hdiff, places=13, msg=f"[{i}] hdiff")

    def test_update_period_80(self):
        ind = HurstDifference(HurstDifferenceParams(period=80))
        for i, val in enumerate(INPUT_CLOSE):
            hurst_diff, fgdi = ind.update(val)
            exp_fgdi = EXPECTED_FGDI_P80[i]
            exp_hdiff = EXPECTED_HDIFF_P80[i]
            if math.isnan(exp_fgdi):
                self.assertTrue(math.isnan(fgdi), f"[{i}] fgdi expected NaN, got {fgdi}")
            else:
                self.assertAlmostEqual(fgdi, exp_fgdi, places=13, msg=f"[{i}] fgdi")
            if math.isnan(exp_hdiff):
                self.assertTrue(math.isnan(hurst_diff), f"[{i}] hdiff expected NaN, got {hurst_diff}")
            else:
                self.assertAlmostEqual(hurst_diff, exp_hdiff, places=13, msg=f"[{i}] hdiff")

    def test_update_period_120(self):
        ind = HurstDifference(HurstDifferenceParams(period=120))
        for i, val in enumerate(INPUT_CLOSE):
            hurst_diff, fgdi = ind.update(val)
            exp_fgdi = EXPECTED_FGDI_P120[i]
            exp_hdiff = EXPECTED_HDIFF_P120[i]
            if math.isnan(exp_fgdi):
                self.assertTrue(math.isnan(fgdi), f"[{i}] fgdi expected NaN, got {fgdi}")
            else:
                self.assertAlmostEqual(fgdi, exp_fgdi, places=13, msg=f"[{i}] fgdi")
            if math.isnan(exp_hdiff):
                self.assertTrue(math.isnan(hurst_diff), f"[{i}] hdiff expected NaN, got {hurst_diff}")
            else:
                self.assertAlmostEqual(hurst_diff, exp_hdiff, places=13, msg=f"[{i}] hdiff")

    def test_is_primed_period_30(self):
        ind = HurstDifference(HurstDifferenceParams(period=30))
        for i in range(30):
            ind.update(INPUT_CLOSE[i])
            self.assertFalse(ind.is_primed())
        ind.update(INPUT_CLOSE[30])
        self.assertTrue(ind.is_primed())

    def test_nan_passthrough(self):
        ind = HurstDifference(HurstDifferenceParams(period=5))
        hurst_diff, fgdi = ind.update(math.nan)
        self.assertTrue(math.isnan(hurst_diff))
        self.assertTrue(math.isnan(fgdi))

    def test_invalid_period(self):
        with self.assertRaises(ValueError):
            HurstDifference(HurstDifferenceParams(period=1))

    def test_metadata(self):
        ind = HurstDifference(HurstDifferenceParams(period=30))
        meta = ind.metadata()
        self.assertEqual(meta.identifier, Identifier.HURST_DIFFERENCE)
        self.assertIn("hurdif(30)", meta.mnemonic)

    def test_update_bar(self):
        ind = HurstDifference(HurstDifferenceParams(period=5))
        for i, val in enumerate(INPUT_CLOSE):
            bar = Bar(datetime(2020, 1, 1), val + 1, val + 2, val - 1, val, 100.0)
            output = ind.update_bar(bar)
            exp_fgdi = EXPECTED_FGDI_P5[i]
            exp_hdiff = EXPECTED_HDIFF_P5[i]
            if math.isnan(exp_hdiff):
                self.assertTrue(math.isnan(output[0].value))
            else:
                self.assertAlmostEqual(output[0].value, exp_hdiff, places=13)
            if math.isnan(exp_fgdi):
                self.assertTrue(math.isnan(output[1].value))
            else:
                self.assertAlmostEqual(output[1].value, exp_fgdi, places=13)

    def test_update_scalar(self):
        ind = HurstDifference(HurstDifferenceParams(period=5))
        for i, val in enumerate(INPUT_CLOSE):
            scalar = Scalar(datetime(2020, 1, 1), val)
            output = ind.update_scalar(scalar)
            exp_fgdi = EXPECTED_FGDI_P5[i]
            exp_hdiff = EXPECTED_HDIFF_P5[i]
            if math.isnan(exp_hdiff):
                self.assertTrue(math.isnan(output[0].value))
            else:
                self.assertAlmostEqual(output[0].value, exp_hdiff, places=13)
            if math.isnan(exp_fgdi):
                self.assertTrue(math.isnan(output[1].value))
            else:
                self.assertAlmostEqual(output[1].value, exp_fgdi, places=13)


if __name__ == '__main__':
    unittest.main()
