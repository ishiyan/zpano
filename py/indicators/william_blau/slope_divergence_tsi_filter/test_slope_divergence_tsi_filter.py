import math
import unittest
from datetime import datetime

from py.indicators.william_blau.slope_divergence_tsi_filter.slope_divergence_tsi_filter import SlopeDivergenceTsiFilter
from py.indicators.william_blau.slope_divergence_tsi_filter.params import SlopeDivergenceTsiFilterParams
from py.indicators.core.identifier import Identifier
from py.entities.bar_component import BarComponent
from py.entities.quote_component import QuoteComponent
from py.entities.trade_component import TradeComponent
from py.entities.scalar import Scalar

from .test_testdata import (
    INPUT_CLOSE,
    EXPECTED_R32_S32_U7_X32_Y7,
    EXPECTED_R32_S32_U1_X32_Y1,
    EXPECTED_R32_S32_U7_X32_Y1,
    EXPECTED_R32_S32_U1_X32_Y7,
    EXPECTED_R1_S1_U1_X1_Y1,
    EXPECTED_R20_S5_U3_X20_Y3,
    EXPECTED_R32_S13_U3_X32_Y7,
    EXPECTED_R12_S12_U1_X12_Y1,
    EXPECTED_R25_S13_U1_X25_Y1,
    EXPECTED_R64_S64_U7_X32_Y7,
    EXPECTED_R32_S32_U7_X16_Y3,
    EXPECTED_R5_S5_U5_X5_Y5,
    EXPECTED_R10_S10_U1_X10_Y1,
    EXPECTED_R40_S20_U5_X32_Y7,
    EXPECTED_R32_S5_U1_X32_Y1,
    EXPECTED_R50_S25_U1_X50_Y1,
)

TOLERANCE = 1e-13

# The book momentum look-back used by every expected array.
Q = 2

# (r, s, u, x, y, expected)
COMBOS = [
    (32, 32, 7, 32, 7, EXPECTED_R32_S32_U7_X32_Y7),
    (32, 32, 1, 32, 1, EXPECTED_R32_S32_U1_X32_Y1),
    (32, 32, 7, 32, 1, EXPECTED_R32_S32_U7_X32_Y1),
    (32, 32, 1, 32, 7, EXPECTED_R32_S32_U1_X32_Y7),
    (1, 1, 1, 1, 1, EXPECTED_R1_S1_U1_X1_Y1),
    (20, 5, 3, 20, 3, EXPECTED_R20_S5_U3_X20_Y3),
    (32, 13, 3, 32, 7, EXPECTED_R32_S13_U3_X32_Y7),
    (12, 12, 1, 12, 1, EXPECTED_R12_S12_U1_X12_Y1),
    (25, 13, 1, 25, 1, EXPECTED_R25_S13_U1_X25_Y1),
    (64, 64, 7, 32, 7, EXPECTED_R64_S64_U7_X32_Y7),
    (32, 32, 7, 16, 3, EXPECTED_R32_S32_U7_X16_Y3),
    (5, 5, 5, 5, 5, EXPECTED_R5_S5_U5_X5_Y5),
    (10, 10, 1, 10, 1, EXPECTED_R10_S10_U1_X10_Y1),
    (40, 20, 5, 32, 7, EXPECTED_R40_S20_U5_X32_Y7),
    (32, 5, 1, 32, 1, EXPECTED_R32_S5_U1_X32_Y1),
    (50, 25, 1, 50, 1, EXPECTED_R50_S25_U1_X50_Y1),
]


class TestSlopeDivergenceTsiFilterData(unittest.TestCase):
    """Test SD_TSI against the reference test data for all parameter combinations."""

    def test_all_combos(self):
        for r, s, u, x, y, expected in COMBOS:
            with self.subTest(r=r, s=s, u=u, x=x, y=y):
                ind = SlopeDivergenceTsiFilter(SlopeDivergenceTsiFilterParams(
                    q=Q, r=r, s=s, u=u, x=x, y=y))
                for i in range(len(INPUT_CLOSE)):
                    value = ind.update(INPUT_CLOSE[i])

                    if math.isnan(expected[i]):
                        self.assertTrue(math.isnan(value), f"[{i}] expected NaN, got {value}")
                    else:
                        self.assertAlmostEqual(value, expected[i], delta=TOLERANCE,
                                               msg=f"[{i}] expected {expected[i]}, got {value}")


class TestSlopeDivergenceTsiFilterPassthrough(unittest.TestCase):
    """Test the full-passthrough case: TSI is +/-100, the reference is the close."""

    def test_passthrough(self):
        ind = SlopeDivergenceTsiFilter(SlopeDivergenceTsiFilterParams(
            q=2, r=1, s=1, u=1, x=1, y=1))
        self.assertTrue(math.isnan(ind.update(10.0)))       # momentum undefined
        self.assertEqual(ind.update(12.0), 0.0)             # first finite TSI, no slope
        self.assertEqual(ind.update(11.0), -100.0)          # both falling -> keep
        self.assertEqual(ind.update(13.0), 100.0)           # both rising -> keep
        self.assertEqual(ind.update(14.0), 0.0)             # TSI flat -> zero


class TestSlopeDivergenceTsiFilterWarmUp(unittest.TestCase):
    """Test the warm-up: NaN for bars 0..q-2, then 0.0 at the first finite bar."""

    def test_warm_up_region(self):
        q = 5
        ind = SlopeDivergenceTsiFilter(SlopeDivergenceTsiFilterParams(q=q))
        for i in range(q - 1):
            self.assertTrue(math.isnan(ind.update(INPUT_CLOSE[i])), f"[{i}] expected NaN")

        self.assertEqual(ind.update(INPUT_CLOSE[q - 1]), 0.0)

        for i in range(q, len(INPUT_CLOSE)):
            self.assertFalse(math.isnan(ind.update(INPUT_CLOSE[i])), f"[{i}] unexpected NaN")


class TestSlopeDivergenceTsiFilterBounds(unittest.TestCase):
    """Test that every finite value is bounded to [-100, 100]."""

    def test_bounded(self):
        ind = SlopeDivergenceTsiFilter(SlopeDivergenceTsiFilterParams())
        for i in range(len(INPUT_CLOSE)):
            value = ind.update(INPUT_CLOSE[i])
            if not math.isnan(value):
                self.assertGreaterEqual(value, -100.0, f"[{i}] below -100")
                self.assertLessEqual(value, 100.0, f"[{i}] above 100")


class TestSlopeDivergenceTsiFilterPrimed(unittest.TestCase):
    """Test priming: the filter is primed from the first finite TSI bar."""

    def test_is_primed(self):
        q = 5
        ind = SlopeDivergenceTsiFilter(SlopeDivergenceTsiFilterParams(q=q))
        for i in range(q - 1):
            ind.update(INPUT_CLOSE[i])
            self.assertFalse(ind.is_primed(), f"[{i}] must not be primed")

        for i in range(q - 1, len(INPUT_CLOSE)):
            ind.update(INPUT_CLOSE[i])
            self.assertTrue(ind.is_primed(), f"[{i}] must be primed")


class TestSlopeDivergenceTsiFilterMnemonic(unittest.TestCase):
    """Test mnemonic generation for all component combinations."""

    def test_all_components_default(self):
        ind = SlopeDivergenceTsiFilter(SlopeDivergenceTsiFilterParams())
        self.assertEqual(ind.metadata().mnemonic, "sdtsi(2,32,32,7,32,7)")
        self.assertEqual(ind.metadata().description,
                         "Slope Divergence TSI Filter sdtsi(2,32,32,7,32,7)")

    def test_custom_periods(self):
        ind = SlopeDivergenceTsiFilter(SlopeDivergenceTsiFilterParams(
            q=3, r=20, s=5, u=3, x=20, y=3))
        self.assertEqual(ind.metadata().mnemonic, "sdtsi(3,20,5,3,20,3)")

    def test_only_bar_component_set(self):
        ind = SlopeDivergenceTsiFilter(SlopeDivergenceTsiFilterParams(bar_component=BarComponent.MEDIAN))
        self.assertEqual(ind.metadata().mnemonic, "sdtsi(2,32,32,7,32,7, hl/2)")

    def test_only_quote_component_set(self):
        ind = SlopeDivergenceTsiFilter(SlopeDivergenceTsiFilterParams(quote_component=QuoteComponent.BID))
        self.assertEqual(ind.metadata().mnemonic, "sdtsi(2,32,32,7,32,7, b)")

    def test_only_trade_component_set(self):
        ind = SlopeDivergenceTsiFilter(SlopeDivergenceTsiFilterParams(trade_component=TradeComponent.VOLUME))
        self.assertEqual(ind.metadata().mnemonic, "sdtsi(2,32,32,7,32,7, v)")

    def test_bar_and_quote_components_set(self):
        ind = SlopeDivergenceTsiFilter(SlopeDivergenceTsiFilterParams(
            bar_component=BarComponent.OPEN, quote_component=QuoteComponent.BID))
        self.assertEqual(ind.metadata().mnemonic, "sdtsi(2,32,32,7,32,7, o, b)")

    def test_bar_and_trade_components_set(self):
        ind = SlopeDivergenceTsiFilter(SlopeDivergenceTsiFilterParams(
            bar_component=BarComponent.HIGH, trade_component=TradeComponent.VOLUME))
        self.assertEqual(ind.metadata().mnemonic, "sdtsi(2,32,32,7,32,7, h, v)")

    def test_quote_and_trade_components_set(self):
        ind = SlopeDivergenceTsiFilter(SlopeDivergenceTsiFilterParams(
            quote_component=QuoteComponent.ASK, trade_component=TradeComponent.VOLUME))
        self.assertEqual(ind.metadata().mnemonic, "sdtsi(2,32,32,7,32,7, a, v)")


class TestSlopeDivergenceTsiFilterMetadata(unittest.TestCase):
    """Test metadata generation."""

    def test_default_metadata(self):
        ind = SlopeDivergenceTsiFilter(SlopeDivergenceTsiFilterParams())
        meta = ind.metadata()
        self.assertEqual(meta.identifier, Identifier.SLOPE_DIVERGENCE_TSI_FILTER)
        self.assertEqual(meta.mnemonic, "sdtsi(2,32,32,7,32,7)")
        self.assertEqual(meta.description, "Slope Divergence TSI Filter sdtsi(2,32,32,7,32,7)")
        self.assertEqual(len(meta.outputs), 1)
        self.assertEqual(meta.outputs[0].mnemonic, "sdtsi(2,32,32,7,32,7)")


class TestSlopeDivergenceTsiFilterUpdateScalar(unittest.TestCase):
    """Test the update_scalar output."""

    def test_update_scalar(self):
        ind = SlopeDivergenceTsiFilter(SlopeDivergenceTsiFilterParams())
        tm = datetime(2021, 4, 1)
        out = None
        for i in range(len(INPUT_CLOSE)):
            out = ind.update_scalar(Scalar(time=tm, value=INPUT_CLOSE[i]))
            self.assertEqual(len(out), 1)
            self.assertEqual(out[0].time, tm)
            expected = EXPECTED_R32_S32_U7_X32_Y7[i]
            if math.isnan(expected):
                self.assertTrue(math.isnan(out[0].value), f"[{i}] expected NaN")
            else:
                self.assertAlmostEqual(out[0].value, expected, delta=TOLERANCE, msg=f"[{i}]")


class TestSlopeDivergenceTsiFilterInvalidParams(unittest.TestCase):
    """Test invalid parameter validation."""

    def test_q_too_small(self):
        with self.assertRaises(ValueError):
            SlopeDivergenceTsiFilter(SlopeDivergenceTsiFilterParams(q=0))

    def test_r_too_small(self):
        with self.assertRaises(ValueError):
            SlopeDivergenceTsiFilter(SlopeDivergenceTsiFilterParams(r=0))

    def test_s_too_small(self):
        with self.assertRaises(ValueError):
            SlopeDivergenceTsiFilter(SlopeDivergenceTsiFilterParams(s=0))

    def test_u_too_small(self):
        with self.assertRaises(ValueError):
            SlopeDivergenceTsiFilter(SlopeDivergenceTsiFilterParams(u=0))

    def test_x_too_small(self):
        with self.assertRaises(ValueError):
            SlopeDivergenceTsiFilter(SlopeDivergenceTsiFilterParams(x=0))

    def test_y_too_small(self):
        with self.assertRaises(ValueError):
            SlopeDivergenceTsiFilter(SlopeDivergenceTsiFilterParams(y=0))


if __name__ == '__main__':
    unittest.main()
