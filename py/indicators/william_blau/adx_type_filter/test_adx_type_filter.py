import math
import unittest
from datetime import datetime

from py.indicators.william_blau.adx_type_filter.adx_type_filter import AdxTypeFilter
from py.indicators.william_blau.adx_type_filter.params import AdxTypeFilterParams, AdxTypeFilterSource
from py.indicators.core.identifier import Identifier
from py.entities.bar_component import BarComponent
from py.entities.quote_component import QuoteComponent
from py.entities.trade_component import TradeComponent
from py.entities.bar import Bar
from py.entities.quote import Quote
from py.entities.trade import Trade
from py.entities.scalar import Scalar

from .test_testdata import (
    INPUT_CLOSE,
    INPUT_HIGH,
    INPUT_LOW,
    EXPECTED_TSIMTM_Q2_R32_S32,
    EXPECTED_TSIMTM_Q2_R20_S5,
    EXPECTED_TSIMTM_Q2_R13_S1,
    EXPECTED_TSIMTM_Q2_R1_S1,
    EXPECTED_TSIMTM_Q5_R32_S32,
    EXPECTED_SMIRAW_Q32_R32_S32,
    EXPECTED_SMIRAW_Q32_R20_S5,
    EXPECTED_SMIRAW_Q5_R32_S32,
    EXPECTED_DTINUM_Q2_R32_S32,
    EXPECTED_DTINUM_Q2_R28_S28,
    EXPECTED_DTINUM_Q5_R32_S32,
    EXPECTED_TVI_R32_S32,
    EXPECTED_TVI_R12_S12,
    EXPECTED_TVI_R1_S1,
    EXPECTED_TSINORM_R32_S32,
    EXPECTED_TSINORM_R20_S20,
)

TOLERANCE = 1e-13

TSI = AdxTypeFilterSource.TSI_MOMENTUM
SMI = AdxTypeFilterSource.SMI_MOMENTUM
DTI = AdxTypeFilterSource.DTI_MOMENTUM
TVI = AdxTypeFilterSource.TVI_BALANCE
TSIN = AdxTypeFilterSource.TSI_NORMALIZED

# (source, q, r, s, expected); q=0 is used where the source ignores it.
CLOSE_COMBOS = [
    (TSI, 2, 32, 32, EXPECTED_TSIMTM_Q2_R32_S32),
    (TSI, 2, 20, 5, EXPECTED_TSIMTM_Q2_R20_S5),
    (TSI, 2, 13, 1, EXPECTED_TSIMTM_Q2_R13_S1),
    (TSI, 2, 1, 1, EXPECTED_TSIMTM_Q2_R1_S1),
    (TSI, 5, 32, 32, EXPECTED_TSIMTM_Q5_R32_S32),
    (TSIN, 2, 32, 32, EXPECTED_TSINORM_R32_S32),
    (TSIN, 2, 20, 20, EXPECTED_TSINORM_R20_S20),
]

BAR_COMBOS = [
    (SMI, 32, 32, 32, EXPECTED_SMIRAW_Q32_R32_S32),
    (SMI, 32, 20, 5, EXPECTED_SMIRAW_Q32_R20_S5),
    (SMI, 5, 32, 32, EXPECTED_SMIRAW_Q5_R32_S32),
    (DTI, 2, 32, 32, EXPECTED_DTINUM_Q2_R32_S32),
    (DTI, 2, 28, 28, EXPECTED_DTINUM_Q2_R28_S28),
    (DTI, 5, 32, 32, EXPECTED_DTINUM_Q5_R32_S32),
    (TVI, 0, 32, 32, EXPECTED_TVI_R32_S32),
    (TVI, 0, 12, 12, EXPECTED_TVI_R12_S12),
    (TVI, 0, 1, 1, EXPECTED_TVI_R1_S1),
]

TM = datetime(2021, 4, 1)


def _bar(i: int) -> Bar:
    return Bar(TM, INPUT_CLOSE[i], INPUT_HIGH[i], INPUT_LOW[i], INPUT_CLOSE[i], 0.0)


class TestAdxTypeFilterData(unittest.TestCase):
    """Test the ATF against the reference test data for all parameter combinations."""

    def _check(self, i, value, expected):
        if math.isnan(expected):
            self.assertTrue(math.isnan(value), f"[{i}] expected NaN, got {value}")
        else:
            self.assertAlmostEqual(value, expected, delta=TOLERANCE,
                                   msg=f"[{i}] expected {expected}, got {value}")

    def test_close_sources(self):
        for source, q, r, s, expected in CLOSE_COMBOS:
            with self.subTest(source=source.name, q=q, r=r, s=s):
                ind = AdxTypeFilter(AdxTypeFilterParams(source=source, q=q, r=r, s=s))
                for i in range(len(INPUT_CLOSE)):
                    self._check(i, ind.update(INPUT_CLOSE[i]), expected[i])

    def test_bar_sources(self):
        for source, q, r, s, expected in BAR_COMBOS:
            with self.subTest(source=source.name, q=q, r=r, s=s):
                ind = AdxTypeFilter(AdxTypeFilterParams(source=source, q=q, r=r, s=s))
                for i in range(len(INPUT_CLOSE)):
                    out = ind.update_bar(_bar(i))
                    self.assertEqual(len(out), 1)
                    self._check(i, out[0].value, expected[i])

    def test_bar_sources_high_low_close(self):
        for source, q, r, s, expected in BAR_COMBOS:
            with self.subTest(source=source.name, q=q, r=r, s=s):
                ind = AdxTypeFilter(AdxTypeFilterParams(source=source, q=q, r=r, s=s))
                for i in range(len(INPUT_CLOSE)):
                    value = ind.update_high_low_close(INPUT_HIGH[i], INPUT_LOW[i], INPUT_CLOSE[i])
                    self._check(i, value, expected[i])


class TestAdxTypeFilterPassthrough(unittest.TestCase):
    """Test the r = s = 1 invariant: ATF == |momentum|."""

    def test_tsi_momentum(self):
        ind = AdxTypeFilter(AdxTypeFilterParams(source=TSI, q=2, r=1, s=1))
        self.assertTrue(math.isnan(ind.update(10.0)))
        self.assertEqual(ind.update(12.0), 2.0)
        self.assertEqual(ind.update(11.0), 1.0)

    def test_smi_momentum(self):
        ind = AdxTypeFilter(AdxTypeFilterParams(source=SMI, q=1, r=1, s=1))
        self.assertEqual(ind.update_high_low_close(11.0, 9.0, 10.5), 0.5)

    def test_tvi_tick_rule(self):
        ind = AdxTypeFilter(AdxTypeFilterParams(source=TVI, r=1, s=1))
        self.assertEqual(ind.update(10.0), 0.0)
        self.assertEqual(ind.update(12.0), 2.0)
        self.assertEqual(ind.update(9.0), 3.0)


class TestAdxTypeFilterWarmUp(unittest.TestCase):
    """Test the q-bar warm-up: NaN for bars 0..q-2, finite from bar q-1."""

    def test_smi_default_warm_up(self):
        ind = AdxTypeFilter(AdxTypeFilterParams(source=SMI))
        for i in range(31):
            self.assertTrue(math.isnan(ind.update_bar(_bar(i))[0].value), f"[{i}] expected NaN")
        for i in range(31, len(INPUT_CLOSE)):
            self.assertFalse(math.isnan(ind.update_bar(_bar(i))[0].value), f"[{i}] unexpected NaN")

    def test_tvi_has_no_warm_up(self):
        ind = AdxTypeFilter(AdxTypeFilterParams(source=TVI))
        for i in range(len(INPUT_CLOSE)):
            self.assertFalse(math.isnan(ind.update_bar(_bar(i))[0].value), f"[{i}] unexpected NaN")


class TestAdxTypeFilterNonNegative(unittest.TestCase):
    """Test that every finite value is non-negative."""

    def test_non_negative(self):
        for source in AdxTypeFilterSource:
            with self.subTest(source=source.name):
                ind = AdxTypeFilter(AdxTypeFilterParams(source=source))
                for i in range(len(INPUT_CLOSE)):
                    value = ind.update_bar(_bar(i))[0].value
                    if not math.isnan(value):
                        self.assertGreaterEqual(value, 0.0, f"[{i}] negative")


class TestAdxTypeFilterPrimed(unittest.TestCase):
    """Test priming: the filter is primed from the first finite momentum."""

    def test_is_primed(self):
        q = 5
        ind = AdxTypeFilter(AdxTypeFilterParams(source=TSI, q=q))
        for i in range(q - 1):
            ind.update(INPUT_CLOSE[i])
            self.assertFalse(ind.is_primed(), f"[{i}] must not be primed")

        for i in range(q - 1, len(INPUT_CLOSE)):
            ind.update(INPUT_CLOSE[i])
            self.assertTrue(ind.is_primed(), f"[{i}] must be primed")


class TestAdxTypeFilterEntityUpdates(unittest.TestCase):
    """Test the entity mapping of single-valued samples."""

    def test_tsi_momentum_entities(self):
        expected = EXPECTED_TSIMTM_Q2_R32_S32
        ind_scalar = AdxTypeFilter(AdxTypeFilterParams())
        ind_bar = AdxTypeFilter(AdxTypeFilterParams())
        ind_quote = AdxTypeFilter(AdxTypeFilterParams())
        ind_trade = AdxTypeFilter(AdxTypeFilterParams())

        for i in range(len(INPUT_CLOSE)):
            v = INPUT_CLOSE[i]
            outs = [
                ind_scalar.update_scalar(Scalar(time=TM, value=v)),
                ind_bar.update_bar(_bar(i)),
                ind_quote.update_quote(Quote(TM, v, v, 0.0, 0.0)),
                ind_trade.update_trade(Trade(TM, v, 0.0)),
            ]
            for out in outs:
                self.assertEqual(len(out), 1)
                self.assertEqual(out[0].time, TM)
                if math.isnan(expected[i]):
                    self.assertTrue(math.isnan(out[0].value), f"[{i}] expected NaN")
                else:
                    self.assertAlmostEqual(out[0].value, expected[i], delta=TOLERANCE, msg=f"[{i}]")

    def test_smi_single_values_use_value_as_high_low_close(self):
        ind_scalar = AdxTypeFilter(AdxTypeFilterParams(source=SMI, q=5))
        ind_quote = AdxTypeFilter(AdxTypeFilterParams(source=SMI, q=5))
        ind_trade = AdxTypeFilter(AdxTypeFilterParams(source=SMI, q=5))
        ind_hlc = AdxTypeFilter(AdxTypeFilterParams(source=SMI, q=5))

        for i in range(len(INPUT_CLOSE)):
            v = INPUT_CLOSE[i]
            expected = ind_hlc.update_high_low_close(v, v, v)
            for value in (
                ind_scalar.update_scalar(Scalar(time=TM, value=v))[0].value,
                ind_quote.update_quote(Quote(TM, v, v, 0.0, 0.0))[0].value,
                ind_trade.update_trade(Trade(TM, v, 0.0))[0].value,
            ):
                if math.isnan(expected):
                    self.assertTrue(math.isnan(value), f"[{i}] expected NaN")
                else:
                    self.assertEqual(value, expected, f"[{i}]")


class TestAdxTypeFilterMnemonic(unittest.TestCase):
    """Test mnemonic generation for all sources and component combinations."""

    def test_default_mnemonic(self):
        ind = AdxTypeFilter(AdxTypeFilterParams())
        self.assertEqual(ind.metadata().mnemonic, "atf.tsi(2,32,32)")
        self.assertEqual(ind.metadata().description, "ADX-Type Filter atf.tsi(2,32,32)")

    def test_source_mnemonics(self):
        cases = [
            (TSI, "atf.tsi(2,32,32)"),
            (SMI, "atf.smi(32,32,32)"),
            (DTI, "atf.dti(2,32,32)"),
            (TVI, "atf.tvi(32,32)"),
            (TSIN, "atf.tsin(2,32,32)"),
        ]
        for source, mnemonic in cases:
            with self.subTest(source=source.name):
                ind = AdxTypeFilter(AdxTypeFilterParams(source=source))
                self.assertEqual(ind.metadata().mnemonic, mnemonic)

    def test_custom_periods(self):
        ind = AdxTypeFilter(AdxTypeFilterParams(source=SMI, q=5, r=20, s=5))
        self.assertEqual(ind.metadata().mnemonic, "atf.smi(5,20,5)")

    def test_only_bar_component_set(self):
        ind = AdxTypeFilter(AdxTypeFilterParams(bar_component=BarComponent.MEDIAN))
        self.assertEqual(ind.metadata().mnemonic, "atf.tsi(2,32,32, hl/2)")

    def test_only_quote_component_set(self):
        ind = AdxTypeFilter(AdxTypeFilterParams(quote_component=QuoteComponent.BID))
        self.assertEqual(ind.metadata().mnemonic, "atf.tsi(2,32,32, b)")

    def test_only_trade_component_set(self):
        ind = AdxTypeFilter(AdxTypeFilterParams(trade_component=TradeComponent.VOLUME))
        self.assertEqual(ind.metadata().mnemonic, "atf.tsi(2,32,32, v)")

    def test_bar_and_quote_components_set(self):
        ind = AdxTypeFilter(AdxTypeFilterParams(
            bar_component=BarComponent.OPEN, quote_component=QuoteComponent.BID))
        self.assertEqual(ind.metadata().mnemonic, "atf.tsi(2,32,32, o, b)")

    def test_bar_and_trade_components_set(self):
        ind = AdxTypeFilter(AdxTypeFilterParams(
            bar_component=BarComponent.HIGH, trade_component=TradeComponent.VOLUME))
        self.assertEqual(ind.metadata().mnemonic, "atf.tsi(2,32,32, h, v)")

    def test_quote_and_trade_components_set(self):
        ind = AdxTypeFilter(AdxTypeFilterParams(
            quote_component=QuoteComponent.ASK, trade_component=TradeComponent.VOLUME))
        self.assertEqual(ind.metadata().mnemonic, "atf.tsi(2,32,32, a, v)")

    def test_normalized_component_set(self):
        ind = AdxTypeFilter(AdxTypeFilterParams(source=TSIN, bar_component=BarComponent.MEDIAN))
        self.assertEqual(ind.metadata().mnemonic, "atf.tsin(2,32,32, hl/2)")

    def test_components_ignored_by_bar_sources(self):
        ind = AdxTypeFilter(AdxTypeFilterParams(source=SMI, bar_component=BarComponent.MEDIAN))
        self.assertEqual(ind.metadata().mnemonic, "atf.smi(32,32,32)")


class TestAdxTypeFilterMetadata(unittest.TestCase):
    """Test metadata generation."""

    def test_default_metadata(self):
        ind = AdxTypeFilter(AdxTypeFilterParams())
        meta = ind.metadata()
        self.assertEqual(meta.identifier, Identifier.ADX_TYPE_FILTER)
        self.assertEqual(meta.mnemonic, "atf.tsi(2,32,32)")
        self.assertEqual(meta.description, "ADX-Type Filter atf.tsi(2,32,32)")
        self.assertEqual(len(meta.outputs), 1)
        self.assertEqual(meta.outputs[0].mnemonic, "atf.tsi(2,32,32)")


class TestAdxTypeFilterInvalidParams(unittest.TestCase):
    """Test invalid parameter validation."""

    def test_q_too_small(self):
        with self.assertRaises(ValueError):
            AdxTypeFilter(AdxTypeFilterParams(q=-1))

    def test_r_too_small(self):
        with self.assertRaises(ValueError):
            AdxTypeFilter(AdxTypeFilterParams(r=0))

    def test_s_too_small(self):
        with self.assertRaises(ValueError):
            AdxTypeFilter(AdxTypeFilterParams(s=0))

    def test_unknown_source(self):
        with self.assertRaises(ValueError):
            AdxTypeFilter(AdxTypeFilterParams(source=99))


if __name__ == '__main__':
    unittest.main()
