import math
import unittest
from datetime import datetime

from py.indicators.doug_schaff.schaff_trend_cycle.schaff_trend_cycle import SchaffTrendCycle
from py.indicators.doug_schaff.schaff_trend_cycle.params import SchaffTrendCycleParams
from py.indicators.core.identifier import Identifier
from py.entities.scalar import Scalar

from .test_testdata import (
    INPUT_CLOSE,
    EXPECTED_STC_F23_S50_T10_C50, EXPECTED_MACD_F23_S50_T10_C50, EXPECTED_PF_F23_S50_T10_C50,
    EXPECTED_STC_F12_S26_T10_C50, EXPECTED_MACD_F12_S26_T10_C50, EXPECTED_PF_F12_S26_T10_C50,
    EXPECTED_STC_F5_S10_T5_C50, EXPECTED_MACD_F5_S10_T5_C50, EXPECTED_PF_F5_S10_T5_C50,
    EXPECTED_STC_F3_S7_T3_C50,
    EXPECTED_STC_F8_S21_T10_C50,
    EXPECTED_STC_F10_S30_T10_C50,
    EXPECTED_STC_F15_S40_T14_C50,
    EXPECTED_STC_F6_S13_T8_C60,
    EXPECTED_STC_F23_S50_T23_C50,
    EXPECTED_STC_F23_S50_T5_C50,
    EXPECTED_STC_F12_S26_T10_C25,
    EXPECTED_STC_F12_S26_T10_C80,
    EXPECTED_STC_F12_S26_T10_C100,
    EXPECTED_STC_F20_S40_T10_C50,
)

TOLERANCE = 1e-9

# (fast, slow, tclen, factor, expected_stc, expected_macd_or_None, expected_pf_or_None)
COMBOS = [
    (23, 50, 10, 0.5, EXPECTED_STC_F23_S50_T10_C50, EXPECTED_MACD_F23_S50_T10_C50, EXPECTED_PF_F23_S50_T10_C50),
    (12, 26, 10, 0.5, EXPECTED_STC_F12_S26_T10_C50, EXPECTED_MACD_F12_S26_T10_C50, EXPECTED_PF_F12_S26_T10_C50),
    (5, 10, 5, 0.5, EXPECTED_STC_F5_S10_T5_C50, EXPECTED_MACD_F5_S10_T5_C50, EXPECTED_PF_F5_S10_T5_C50),
    (3, 7, 3, 0.5, EXPECTED_STC_F3_S7_T3_C50, None, None),
    (8, 21, 10, 0.5, EXPECTED_STC_F8_S21_T10_C50, None, None),
    (10, 30, 10, 0.5, EXPECTED_STC_F10_S30_T10_C50, None, None),
    (15, 40, 14, 0.5, EXPECTED_STC_F15_S40_T14_C50, None, None),
    (6, 13, 8, 0.6, EXPECTED_STC_F6_S13_T8_C60, None, None),
    (23, 50, 23, 0.5, EXPECTED_STC_F23_S50_T23_C50, None, None),
    (23, 50, 5, 0.5, EXPECTED_STC_F23_S50_T5_C50, None, None),
    (12, 26, 10, 0.25, EXPECTED_STC_F12_S26_T10_C25, None, None),
    (12, 26, 10, 0.8, EXPECTED_STC_F12_S26_T10_C80, None, None),
    (12, 26, 10, 1.0, EXPECTED_STC_F12_S26_T10_C100, None, None),
    (20, 40, 10, 0.5, EXPECTED_STC_F20_S40_T10_C50, None, None),
]


class TestSchaffTrendCycleData(unittest.TestCase):
    """Test STC against the reference test data for all parameter combinations."""

    def test_all_combos(self):
        for fast, slow, tclen, factor, exp_stc, exp_macd, exp_pf in COMBOS:
            with self.subTest(fast=fast, slow=slow, tclen=tclen, factor=factor):
                ind = SchaffTrendCycle(SchaffTrendCycleParams(
                    fast=fast, slow=slow, tclen=tclen, factor=factor))
                for i in range(len(INPUT_CLOSE)):
                    stc, macd, pf = ind.update(INPUT_CLOSE[i])

                    if math.isnan(exp_stc[i]):
                        self.assertTrue(math.isnan(stc), f"[{i}] stc: expected NaN, got {stc}")
                    else:
                        self.assertAlmostEqual(stc, exp_stc[i], delta=TOLERANCE,
                                               msg=f"[{i}] stc: expected {exp_stc[i]}, got {stc}")

                    if exp_macd is not None:
                        self.assertAlmostEqual(macd, exp_macd[i], delta=TOLERANCE,
                                               msg=f"[{i}] macd: expected {exp_macd[i]}, got {macd}")
                    if exp_pf is not None:
                        self.assertAlmostEqual(pf, exp_pf[i], delta=TOLERANCE,
                                               msg=f"[{i}] pf: expected {exp_pf[i]}, got {pf}")


class TestSchaffTrendCycleMnemonic(unittest.TestCase):
    """Test mnemonic generation."""

    def test_default_mnemonic(self):
        ind = SchaffTrendCycle(SchaffTrendCycleParams())
        self.assertEqual(ind.metadata().mnemonic, "stc(23,50,10,0.50)")

    def test_custom_mnemonic(self):
        ind = SchaffTrendCycle(SchaffTrendCycleParams(fast=12, slow=26, tclen=10, factor=0.25))
        self.assertEqual(ind.metadata().mnemonic, "stc(12,26,10,0.25)")


class TestSchaffTrendCycleMetadata(unittest.TestCase):
    """Test metadata generation."""

    def test_default_metadata(self):
        ind = SchaffTrendCycle(SchaffTrendCycleParams())
        meta = ind.metadata()
        self.assertEqual(meta.identifier, Identifier.SCHAFF_TREND_CYCLE)
        self.assertEqual(meta.mnemonic, "stc(23,50,10,0.50)")
        self.assertEqual(len(meta.outputs), 3)


class TestSchaffTrendCycleUpdateScalar(unittest.TestCase):
    """Test update_scalar output ordering (stc, macd, pf)."""

    def test_update_scalar(self):
        ind = SchaffTrendCycle(SchaffTrendCycleParams())
        tm = datetime(2021, 4, 1)
        out = None
        for i in range(len(INPUT_CLOSE)):
            out = ind.update_scalar(Scalar(time=tm, value=INPUT_CLOSE[i]))
        self.assertEqual(len(out), 3)
        self.assertAlmostEqual(out[0].value, EXPECTED_STC_F23_S50_T10_C50[-1], delta=TOLERANCE)
        self.assertAlmostEqual(out[1].value, EXPECTED_MACD_F23_S50_T10_C50[-1], delta=TOLERANCE)
        self.assertAlmostEqual(out[2].value, EXPECTED_PF_F23_S50_T10_C50[-1], delta=TOLERANCE)


class TestSchaffTrendCycleInvalidParams(unittest.TestCase):
    """Test invalid parameter validation."""

    def test_fast_too_small(self):
        with self.assertRaises(ValueError):
            SchaffTrendCycle(SchaffTrendCycleParams(fast=0))

    def test_slow_too_small(self):
        with self.assertRaises(ValueError):
            SchaffTrendCycle(SchaffTrendCycleParams(slow=0))

    def test_tclen_too_small(self):
        with self.assertRaises(ValueError):
            SchaffTrendCycle(SchaffTrendCycleParams(tclen=0))

    def test_factor_zero(self):
        with self.assertRaises(ValueError):
            SchaffTrendCycle(SchaffTrendCycleParams(factor=0.0))

    def test_factor_too_large(self):
        with self.assertRaises(ValueError):
            SchaffTrendCycle(SchaffTrendCycleParams(factor=1.5))


class TestSchaffTrendCycleNaN(unittest.TestCase):
    """Test NaN input handling."""

    def test_nan(self):
        ind = SchaffTrendCycle(SchaffTrendCycleParams())
        # Warm-up: stc is NaN; macd/pf are 0.0 pre-gate.
        stc, macd, pf = ind.update(math.nan)
        self.assertTrue(math.isnan(stc))


if __name__ == '__main__':
    unittest.main()
