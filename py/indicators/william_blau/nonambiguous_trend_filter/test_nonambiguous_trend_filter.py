import math
import unittest
from datetime import datetime

from py.indicators.william_blau.nonambiguous_trend_filter.nonambiguous_trend_filter import NonambiguousTrendFilter
from py.indicators.william_blau.nonambiguous_trend_filter.params import (
    NonambiguousTrendFilterBase, NonambiguousTrendFilterParams,
)
from py.indicators.core.identifier import Identifier
from py.entities.bar_component import BarComponent
from py.entities.quote_component import QuoteComponent
from py.entities.trade_component import TradeComponent
from py.entities.bar import Bar
from py.entities.quote import Quote
from py.entities.trade import Trade
from py.entities.scalar import Scalar

from .test_testdata import (
    INPUT_OPEN,
    INPUT_HIGH,
    INPUT_LOW,
    INPUT_CLOSE,
    EXPECTED_TSI_R32_S13_U3,
    EXPECTED_TSI_R20_S5_U3,
    EXPECTED_TSI_R40_S20_U5,
    EXPECTED_SMI_Q32_R64_S7_U1,
    EXPECTED_SMI_Q5_R20_S5_U3,
    EXPECTED_SMI_Q13_R25_S2_U1,
    EXPECTED_DTI_Q2_R28_S28_U5,
    EXPECTED_DTI_Q2_R20_S5_U3,
    EXPECTED_DTI_Q4_R14_S14_U3,
    EXPECTED_MDI_R20_S5_U3,
    EXPECTED_MDI_R40_S5_U3,
    EXPECTED_CMI_R20_S5_U3,
    EXPECTED_CMI_R10_S5_U3,
    EXPECTED_CSI_R32_S32_U1,
    EXPECTED_CSI_R20_S5_U3,
    EXPECTED_CSI_R1_S1_U1,
    EXPECTED_TVI_R32_S32_U5,
    EXPECTED_TVI_R12_S12_U1,
    EXPECTED_TVI_R25_S13_U1,
)

TOLERANCE = 1e-13

B = NonambiguousTrendFilterBase

# (base, q, r, s, u, expected); q=0 where the base does not use it.
COMBOS = [
    (B.TSI, 2, 32, 13, 3, EXPECTED_TSI_R32_S13_U3),
    (B.TSI, 2, 20, 5, 3, EXPECTED_TSI_R20_S5_U3),
    (B.TSI, 2, 40, 20, 5, EXPECTED_TSI_R40_S20_U5),
    (B.SMI, 32, 64, 7, 1, EXPECTED_SMI_Q32_R64_S7_U1),
    (B.SMI, 5, 20, 5, 3, EXPECTED_SMI_Q5_R20_S5_U3),
    (B.SMI, 13, 25, 2, 1, EXPECTED_SMI_Q13_R25_S2_U1),
    (B.DTI, 2, 28, 28, 5, EXPECTED_DTI_Q2_R28_S28_U5),
    (B.DTI, 2, 20, 5, 3, EXPECTED_DTI_Q2_R20_S5_U3),
    (B.DTI, 4, 14, 14, 3, EXPECTED_DTI_Q4_R14_S14_U3),
    (B.MDI, 0, 20, 5, 3, EXPECTED_MDI_R20_S5_U3),
    (B.MDI, 0, 40, 5, 3, EXPECTED_MDI_R40_S5_U3),
    (B.CMI, 0, 20, 5, 3, EXPECTED_CMI_R20_S5_U3),
    (B.CMI, 0, 10, 5, 3, EXPECTED_CMI_R10_S5_U3),
    (B.CSI, 0, 32, 32, 1, EXPECTED_CSI_R32_S32_U1),
    (B.CSI, 0, 20, 5, 3, EXPECTED_CSI_R20_S5_U3),
    (B.CSI, 0, 1, 1, 1, EXPECTED_CSI_R1_S1_U1),
    (B.TVI, 0, 32, 32, 5, EXPECTED_TVI_R32_S32_U5),
    (B.TVI, 0, 12, 12, 1, EXPECTED_TVI_R12_S12_U1),
    (B.TVI, 0, 25, 13, 1, EXPECTED_TVI_R25_S13_U1),
]

TM = datetime(2021, 4, 1)


def _bar(i: int) -> Bar:
    return Bar(TM, INPUT_OPEN[i], INPUT_HIGH[i], INPUT_LOW[i], INPUT_CLOSE[i], 0.0)


def _check(tc: unittest.TestCase, i: int, value: float, expected: float) -> None:
    if math.isnan(expected):
        tc.assertTrue(math.isnan(value), f"[{i}] expected NaN, got {value}")
    else:
        tc.assertAlmostEqual(value, expected, delta=TOLERANCE,
                             msg=f"[{i}] expected {expected}, got {value}")


class TestNonambiguousTrendFilterData(unittest.TestCase):
    """Test the filter against the reference test data for all bases and parameters."""

    def test_all_combos_from_bars(self):
        for base, q, r, s, u, expected in COMBOS:
            with self.subTest(base=base.name, q=q, r=r, s=s, u=u):
                ind = NonambiguousTrendFilter(NonambiguousTrendFilterParams(base=base, q=q, r=r, s=s, u=u))
                for i in range(len(INPUT_CLOSE)):
                    out = ind.update_bar(_bar(i))
                    self.assertEqual(len(out), 1)
                    self.assertEqual(out[0].time, TM)
                    _check(self, i, out[0].value, expected[i])

    def test_close_bases_from_samples(self):
        for base, q, r, s, u, expected in COMBOS:
            if base not in (B.TSI, B.MDI):
                continue
            with self.subTest(base=base.name, q=q, r=r, s=s, u=u):
                ind = NonambiguousTrendFilter(NonambiguousTrendFilterParams(base=base, q=q, r=r, s=s, u=u))
                for i in range(len(INPUT_CLOSE)):
                    _check(self, i, ind.update(INPUT_CLOSE[i]), expected[i])


class TestNonambiguousTrendFilterRule(unittest.TestCase):
    """Test the keep/zero rule on a passthrough CSI base (CSI == 100*(C-O)/(H-L))."""

    def test_rule(self):
        ind = NonambiguousTrendFilter(NonambiguousTrendFilterParams(base=B.CSI, r=1, s=1, u=1))

        def value(o, h, l, c):
            return ind.update_bar(Bar(TM, o, h, l, c, 0.0))[0].value

        self.assertEqual(value(10.0, 12.0, 8.0, 11.0), 0.0)     # first finite: 25 -> 0
        self.assertEqual(value(10.0, 12.0, 8.0, 12.0), 50.0)    # positive and rising
        self.assertEqual(value(10.0, 12.0, 8.0, 11.0), 0.0)     # positive and falling
        self.assertEqual(value(10.0, 12.0, 8.0, 9.0), -25.0)    # negative and falling
        self.assertEqual(value(10.0, 12.0, 8.0, 9.0), 0.0)      # flat
        self.assertEqual(value(10.0, 12.0, 8.0, 9.5), 0.0)      # negative and rising


class TestNonambiguousTrendFilterWarmUp(unittest.TestCase):
    """Test the NaN warm-up of the base and the 0.0 at the first finite bar."""

    def test_smi_warm_up(self):
        q = 13
        ind = NonambiguousTrendFilter(NonambiguousTrendFilterParams(base=B.SMI, q=q, r=25, s=2, u=1))
        for i in range(q - 1):
            self.assertTrue(math.isnan(ind.update_bar(_bar(i))[0].value), f"[{i}] expected NaN")
            self.assertFalse(ind.is_primed(), f"[{i}] must not be primed")

        self.assertEqual(ind.update_bar(_bar(q - 1))[0].value, 0.0)
        self.assertTrue(ind.is_primed())

    def test_tsi_first_bar_is_zero_after_momentum_warm_up(self):
        ind = NonambiguousTrendFilter(NonambiguousTrendFilterParams())
        self.assertTrue(math.isnan(ind.update(INPUT_CLOSE[0])))
        self.assertEqual(ind.update(INPUT_CLOSE[1]), 0.0)


class TestNonambiguousTrendFilterEntityUpdates(unittest.TestCase):
    """Test that the entities are routed to the base with its own mapping."""

    def test_tsi_entities(self):
        expected = EXPECTED_TSI_R32_S13_U3
        inds = [NonambiguousTrendFilter(NonambiguousTrendFilterParams()) for _ in range(4)]

        for i in range(len(INPUT_CLOSE)):
            v = INPUT_CLOSE[i]
            outs = [
                inds[0].update_scalar(Scalar(time=TM, value=v)),
                inds[1].update_bar(_bar(i)),
                inds[2].update_quote(Quote(TM, v, v, 0.0, 0.0)),
                inds[3].update_trade(Trade(TM, v, 0.0)),
            ]
            for out in outs:
                self.assertEqual(len(out), 1)
                self.assertEqual(out[0].time, TM)
                _check(self, i, out[0].value, expected[i])


class TestNonambiguousTrendFilterMnemonic(unittest.TestCase):
    """Test mnemonic generation for all bases and component combinations."""

    def test_default_mnemonic(self):
        ind = NonambiguousTrendFilter(NonambiguousTrendFilterParams())
        self.assertEqual(ind.metadata().mnemonic, "ntf.tsi(2,32,13,3)")
        self.assertEqual(ind.metadata().description, "Nonambiguous Trend Filter ntf.tsi(2,32,13,3)")

    def test_base_mnemonics(self):
        cases = [
            (B.TSI, "ntf.tsi(2,32,13,3)"),
            (B.SMI, "ntf.smi(32,64,7,1)"),
            (B.DTI, "ntf.dti(2,28,28,5)"),
            (B.TVI, "ntf.tvi(32,32,5)"),
            (B.MDI, "ntf.mdi(20,5,3)"),
            (B.CMI, "ntf.cmi(20,5,3)"),
            (B.CSI, "ntf.csi(32,32,1)"),
        ]
        for base, mnemonic in cases:
            with self.subTest(base=base.name):
                ind = NonambiguousTrendFilter(NonambiguousTrendFilterParams(base=base))
                self.assertEqual(ind.metadata().mnemonic, mnemonic)

    def test_custom_periods(self):
        ind = NonambiguousTrendFilter(NonambiguousTrendFilterParams(base=B.SMI, q=5, r=20, s=5, u=3))
        self.assertEqual(ind.metadata().mnemonic, "ntf.smi(5,20,5,3)")

    def test_only_bar_component_set(self):
        ind = NonambiguousTrendFilter(NonambiguousTrendFilterParams(bar_component=BarComponent.MEDIAN))
        self.assertEqual(ind.metadata().mnemonic, "ntf.tsi(2,32,13,3, hl/2)")

    def test_only_quote_component_set(self):
        ind = NonambiguousTrendFilter(NonambiguousTrendFilterParams(quote_component=QuoteComponent.BID))
        self.assertEqual(ind.metadata().mnemonic, "ntf.tsi(2,32,13,3, b)")

    def test_only_trade_component_set(self):
        ind = NonambiguousTrendFilter(NonambiguousTrendFilterParams(trade_component=TradeComponent.VOLUME))
        self.assertEqual(ind.metadata().mnemonic, "ntf.tsi(2,32,13,3, v)")

    def test_bar_and_quote_components_set(self):
        ind = NonambiguousTrendFilter(NonambiguousTrendFilterParams(
            bar_component=BarComponent.OPEN, quote_component=QuoteComponent.BID))
        self.assertEqual(ind.metadata().mnemonic, "ntf.tsi(2,32,13,3, o, b)")

    def test_bar_and_trade_components_set(self):
        ind = NonambiguousTrendFilter(NonambiguousTrendFilterParams(
            bar_component=BarComponent.HIGH, trade_component=TradeComponent.VOLUME))
        self.assertEqual(ind.metadata().mnemonic, "ntf.tsi(2,32,13,3, h, v)")

    def test_quote_and_trade_components_set(self):
        ind = NonambiguousTrendFilter(NonambiguousTrendFilterParams(
            quote_component=QuoteComponent.ASK, trade_component=TradeComponent.VOLUME))
        self.assertEqual(ind.metadata().mnemonic, "ntf.tsi(2,32,13,3, a, v)")

    def test_mdi_component_set(self):
        ind = NonambiguousTrendFilter(NonambiguousTrendFilterParams(base=B.MDI, bar_component=BarComponent.MEDIAN))
        self.assertEqual(ind.metadata().mnemonic, "ntf.mdi(20,5,3, hl/2)")

    def test_components_ignored_by_bar_bases(self):
        ind = NonambiguousTrendFilter(NonambiguousTrendFilterParams(base=B.CSI, bar_component=BarComponent.MEDIAN))
        self.assertEqual(ind.metadata().mnemonic, "ntf.csi(32,32,1)")


class TestNonambiguousTrendFilterMetadata(unittest.TestCase):
    """Test metadata generation."""

    def test_default_metadata(self):
        meta = NonambiguousTrendFilter(NonambiguousTrendFilterParams()).metadata()
        self.assertEqual(meta.identifier, Identifier.NONAMBIGUOUS_TREND_FILTER)
        self.assertEqual(meta.mnemonic, "ntf.tsi(2,32,13,3)")
        self.assertEqual(meta.description, "Nonambiguous Trend Filter ntf.tsi(2,32,13,3)")
        self.assertEqual(len(meta.outputs), 1)
        self.assertEqual(meta.outputs[0].mnemonic, "ntf.tsi(2,32,13,3)")


class TestNonambiguousTrendFilterInvalidParams(unittest.TestCase):
    """Test invalid parameter validation."""

    def test_q_too_small(self):
        with self.assertRaises(ValueError):
            NonambiguousTrendFilter(NonambiguousTrendFilterParams(q=-1))

    def test_r_too_small(self):
        with self.assertRaises(ValueError):
            NonambiguousTrendFilter(NonambiguousTrendFilterParams(r=-1))

    def test_s_too_small(self):
        with self.assertRaises(ValueError):
            NonambiguousTrendFilter(NonambiguousTrendFilterParams(s=-1))

    def test_u_too_small(self):
        with self.assertRaises(ValueError):
            NonambiguousTrendFilter(NonambiguousTrendFilterParams(u=-1))

    def test_unknown_base(self):
        with self.assertRaises(ValueError):
            NonambiguousTrendFilter(NonambiguousTrendFilterParams(base=99))


if __name__ == '__main__':
    unittest.main()
