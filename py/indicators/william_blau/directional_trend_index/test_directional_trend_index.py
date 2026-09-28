import math
import unittest
from datetime import datetime

from py.indicators.william_blau.directional_trend_index.directional_trend_index import DirectionalTrendIndex
from py.indicators.william_blau.directional_trend_index.params import DirectionalTrendIndexParams
from py.indicators.core.identifier import Identifier
from py.entities.bar import Bar
from py.entities.quote import Quote
from py.entities.scalar import Scalar
from py.entities.trade import Trade

from .test_testdata import (
    INPUT_HIGH, INPUT_LOW,
    EXPECTED_Q2_R20_S5_U3, EXPECTED_Q2_R20_S5_U3_SIG_UL3,
    EXPECTED_Q2_R25_S13_U1, EXPECTED_Q2_R25_S13_U1_SIG_UL3,
    EXPECTED_Q2_R20_S5_U1, EXPECTED_Q2_R20_S5_U1_SIG_UL3,
    EXPECTED_Q2_R28_S28_U5, EXPECTED_Q2_R28_S28_U5_SIG_UL3,
    EXPECTED_Q2_R1_S1_U1, EXPECTED_Q2_R1_S1_U1_SIG_UL3,
    EXPECTED_Q3_R20_S5_U3, EXPECTED_Q3_R20_S5_U3_SIG_UL3,
    EXPECTED_Q5_R20_S5_U3, EXPECTED_Q5_R20_S5_U3_SIG_UL3,
    EXPECTED_Q2_R13_S13_U1, EXPECTED_Q2_R13_S13_U1_SIG_UL3,
    EXPECTED_Q2_R40_S20_U1, EXPECTED_Q2_R40_S20_U1_SIG_UL3,
    EXPECTED_Q2_R5_S5_U5, EXPECTED_Q2_R5_S5_U5_SIG_UL3,
    EXPECTED_Q1_R20_S5_U3, EXPECTED_Q1_R20_S5_U3_SIG_UL3,
    EXPECTED_Q10_R20_S5_U1, EXPECTED_Q10_R20_S5_U1_SIG_UL3,
    EXPECTED_Q2_R9_S3_U1, EXPECTED_Q2_R9_S3_U1_SIG_UL3,
    EXPECTED_Q2_R64_S64_U1, EXPECTED_Q2_R64_S64_U1_SIG_UL3,
    EXPECTED_Q4_R28_S28_U5, EXPECTED_Q4_R28_S28_U5_SIG_UL3,
    EXPECTED_Q2_R7_S4_U2, EXPECTED_Q2_R7_S4_U2_SIG_UL3,
)

TOLERANCE = 1e-13

# (q, r, s, u, ul, expected_dti, expected_signal)
COMBOS = [
    (2, 20, 5, 3, 3, EXPECTED_Q2_R20_S5_U3, EXPECTED_Q2_R20_S5_U3_SIG_UL3),
    (2, 25, 13, 1, 3, EXPECTED_Q2_R25_S13_U1, EXPECTED_Q2_R25_S13_U1_SIG_UL3),
    (2, 20, 5, 1, 3, EXPECTED_Q2_R20_S5_U1, EXPECTED_Q2_R20_S5_U1_SIG_UL3),
    (2, 28, 28, 5, 3, EXPECTED_Q2_R28_S28_U5, EXPECTED_Q2_R28_S28_U5_SIG_UL3),
    (2, 1, 1, 1, 3, EXPECTED_Q2_R1_S1_U1, EXPECTED_Q2_R1_S1_U1_SIG_UL3),
    (3, 20, 5, 3, 3, EXPECTED_Q3_R20_S5_U3, EXPECTED_Q3_R20_S5_U3_SIG_UL3),
    (5, 20, 5, 3, 3, EXPECTED_Q5_R20_S5_U3, EXPECTED_Q5_R20_S5_U3_SIG_UL3),
    (2, 13, 13, 1, 3, EXPECTED_Q2_R13_S13_U1, EXPECTED_Q2_R13_S13_U1_SIG_UL3),
    (2, 40, 20, 1, 3, EXPECTED_Q2_R40_S20_U1, EXPECTED_Q2_R40_S20_U1_SIG_UL3),
    (2, 5, 5, 5, 3, EXPECTED_Q2_R5_S5_U5, EXPECTED_Q2_R5_S5_U5_SIG_UL3),
    (1, 20, 5, 3, 3, EXPECTED_Q1_R20_S5_U3, EXPECTED_Q1_R20_S5_U3_SIG_UL3),
    (10, 20, 5, 1, 3, EXPECTED_Q10_R20_S5_U1, EXPECTED_Q10_R20_S5_U1_SIG_UL3),
    (2, 9, 3, 1, 3, EXPECTED_Q2_R9_S3_U1, EXPECTED_Q2_R9_S3_U1_SIG_UL3),
    (2, 64, 64, 1, 3, EXPECTED_Q2_R64_S64_U1, EXPECTED_Q2_R64_S64_U1_SIG_UL3),
    (4, 28, 28, 5, 3, EXPECTED_Q4_R28_S28_U5, EXPECTED_Q4_R28_S28_U5_SIG_UL3),
    (2, 7, 4, 2, 3, EXPECTED_Q2_R7_S4_U2, EXPECTED_Q2_R7_S4_U2_SIG_UL3),
]


class TestDirectionalTrendIndexData(unittest.TestCase):
    """Test DTI against the reference test data for all parameter combinations."""

    def test_all_combos(self):
        for q, r, s, u, ul, exp_dti, exp_signal in COMBOS:
            with self.subTest(q=q, r=r, s=s, u=u, ul=ul):
                ind = DirectionalTrendIndex(DirectionalTrendIndexParams(
                    q=q, r=r, s=s, u=u, ul=ul))
                for i in range(len(INPUT_HIGH)):
                    dti, signal = ind.update(INPUT_HIGH[i], INPUT_LOW[i])

                    if math.isnan(exp_dti[i]):
                        self.assertTrue(math.isnan(dti), f"[{i}] dti: expected NaN, got {dti}")
                    else:
                        self.assertAlmostEqual(dti, exp_dti[i], delta=TOLERANCE,
                                               msg=f"[{i}] dti: expected {exp_dti[i]}, got {dti}")

                    if math.isnan(exp_signal[i]):
                        self.assertTrue(math.isnan(signal), f"[{i}] signal: expected NaN, got {signal}")
                    else:
                        self.assertAlmostEqual(signal, exp_signal[i], delta=TOLERANCE,
                                               msg=f"[{i}] signal: expected {exp_signal[i]}, got {signal}")


class TestDirectionalTrendIndexPassthrough(unittest.TestCase):
    """Test DTI(2,1,1,1,1) = 100*sign(HLM), 0 on flat."""

    def test_passthrough(self):
        ind = DirectionalTrendIndex(DirectionalTrendIndexParams(q=2, r=1, s=1, u=1, ul=1))
        r0 = ind.update(10.0, 9.0)
        self.assertTrue(math.isnan(r0[0]))
        self.assertTrue(math.isnan(r0[1]))
        r1 = ind.update(12.0, 11.0)  # HMU=+2, LMD=0 -> +100
        self.assertAlmostEqual(r1[0], 100.0, delta=TOLERANCE)
        self.assertAlmostEqual(r1[1], 100.0, delta=TOLERANCE)
        r2 = ind.update(11.0, 8.0)  # HMU=0, LMD=3 -> -100
        self.assertAlmostEqual(r2[0], -100.0, delta=TOLERANCE)
        self.assertAlmostEqual(r2[1], -100.0, delta=TOLERANCE)

    def test_division_guard(self):
        # q == 1: HLM == 0 on every bar -> the guard yields 0.0 from bar 0.
        ind = DirectionalTrendIndex(DirectionalTrendIndexParams(q=1))
        for i in range(len(INPUT_HIGH)):
            dti, signal = ind.update(INPUT_HIGH[i], INPUT_LOW[i])
            self.assertEqual(dti, 0.0, f"[{i}]")
            self.assertEqual(signal, 0.0, f"[{i}]")

    def test_signal_equals_dti_when_ul1(self):
        ind = DirectionalTrendIndex(DirectionalTrendIndexParams(ul=1))
        for i in range(len(INPUT_HIGH)):
            dti, signal = ind.update(INPUT_HIGH[i], INPUT_LOW[i])
            if math.isnan(dti):
                self.assertTrue(math.isnan(signal))
            else:
                self.assertEqual(dti, signal, f"[{i}]")


class TestDirectionalTrendIndexPrimed(unittest.TestCase):
    """The DTI is NaN and not primed for bars 0..q-2, primed from bar q-1."""

    def test_primed(self):
        for q in (2, 3, 5, 10):
            with self.subTest(q=q):
                ind = DirectionalTrendIndex(DirectionalTrendIndexParams(q=q))
                self.assertFalse(ind.is_primed())
                for i in range(q - 1):
                    dti, signal = ind.update(INPUT_HIGH[i], INPUT_LOW[i])
                    self.assertFalse(ind.is_primed(), f"[{i}] primed too early")
                    self.assertTrue(math.isnan(dti))
                    self.assertTrue(math.isnan(signal))
                for i in range(q - 1, len(INPUT_HIGH)):
                    ind.update(INPUT_HIGH[i], INPUT_LOW[i])
                    self.assertTrue(ind.is_primed(), f"[{i}] not primed")

    def test_primed_q1(self):
        ind = DirectionalTrendIndex(DirectionalTrendIndexParams(q=1))
        self.assertFalse(ind.is_primed())
        ind.update(INPUT_HIGH[0], INPUT_LOW[0])
        self.assertTrue(ind.is_primed())


class TestDirectionalTrendIndexMnemonic(unittest.TestCase):
    """Test mnemonic generation."""

    def test_default_mnemonic(self):
        ind = DirectionalTrendIndex(DirectionalTrendIndexParams())
        self.assertEqual(ind.metadata().mnemonic, "dti(2,20,5,3,3)")

    def test_custom_mnemonic(self):
        ind = DirectionalTrendIndex(DirectionalTrendIndexParams(q=4, r=28, s=28, u=5, ul=1))
        self.assertEqual(ind.metadata().mnemonic, "dti(4,28,28,5,1)")


class TestDirectionalTrendIndexMetadata(unittest.TestCase):
    """Test metadata generation."""

    def test_default_metadata(self):
        ind = DirectionalTrendIndex(DirectionalTrendIndexParams())
        meta = ind.metadata()
        self.assertEqual(meta.identifier, Identifier.DIRECTIONAL_TREND_INDEX)
        self.assertEqual(meta.mnemonic, "dti(2,20,5,3,3)")
        self.assertEqual(meta.description, "Directional Trend Index dti(2,20,5,3,3)")
        self.assertEqual(len(meta.outputs), 2)
        self.assertEqual(meta.outputs[0].mnemonic, "dti(2,20,5,3,3) dti")
        self.assertEqual(meta.outputs[0].description, "Directional Trend Index dti(2,20,5,3,3) DTI")
        self.assertEqual(meta.outputs[1].mnemonic, "dti(2,20,5,3,3) signal")
        self.assertEqual(meta.outputs[1].description, "Directional Trend Index dti(2,20,5,3,3) signal")


class TestDirectionalTrendIndexUpdateEntities(unittest.TestCase):
    """Test the entity update methods and their high/low mapping."""

    def test_update_bar(self):
        ind = DirectionalTrendIndex(DirectionalTrendIndexParams())
        tm = datetime(2021, 4, 1)
        out = None
        for i in range(len(INPUT_HIGH)):
            out = ind.update_bar(Bar(time=tm, open=0.0, high=INPUT_HIGH[i],
                                     low=INPUT_LOW[i], close=0.0, volume=0.0))
        self.assertEqual(len(out), 2)
        self.assertEqual(out[0].time, tm)
        self.assertAlmostEqual(out[0].value, EXPECTED_Q2_R20_S5_U3[-1], delta=TOLERANCE)
        self.assertAlmostEqual(out[1].value, EXPECTED_Q2_R20_S5_U3_SIG_UL3[-1], delta=TOLERANCE)

    def test_update_scalar(self):
        # A scalar value is used as both the high and the low, so HLM is the
        # plain (q-1)-bar momentum of the value.
        ind = DirectionalTrendIndex(DirectionalTrendIndexParams(q=2, r=1, s=1, u=1, ul=1))
        tm = datetime(2021, 4, 1)
        out = ind.update_scalar(Scalar(time=tm, value=10.0))
        self.assertEqual(len(out), 2)
        self.assertTrue(math.isnan(out[0].value))
        self.assertTrue(math.isnan(out[1].value))
        out = ind.update_scalar(Scalar(time=tm, value=12.0))
        self.assertAlmostEqual(out[0].value, 100.0, delta=TOLERANCE)
        self.assertAlmostEqual(out[1].value, 100.0, delta=TOLERANCE)

    def test_update_quote(self):
        # A quote maps the mid price to both the high and the low.
        ind = DirectionalTrendIndex(DirectionalTrendIndexParams(q=2, r=1, s=1, u=1, ul=1))
        tm = datetime(2021, 4, 1)
        ind.update_quote(Quote(time=tm, bid_price=12.0, ask_price=14.0,
                               bid_size=1.0, ask_size=1.0))  # mid 13
        out = ind.update_quote(Quote(time=tm, bid_price=10.0, ask_price=12.0,
                                     bid_size=1.0, ask_size=1.0))  # mid 11, falling
        self.assertEqual(len(out), 2)
        self.assertAlmostEqual(out[0].value, -100.0, delta=TOLERANCE)
        self.assertAlmostEqual(out[1].value, -100.0, delta=TOLERANCE)

    def test_update_trade(self):
        # A trade price is used as both the high and the low.
        ind = DirectionalTrendIndex(DirectionalTrendIndexParams(q=2, r=1, s=1, u=1, ul=1))
        tm = datetime(2021, 4, 1)
        ind.update_trade(Trade(time=tm, price=10.0, volume=1.0))
        out = ind.update_trade(Trade(time=tm, price=12.0, volume=1.0))
        self.assertEqual(len(out), 2)
        self.assertAlmostEqual(out[0].value, 100.0, delta=TOLERANCE)
        self.assertAlmostEqual(out[1].value, 100.0, delta=TOLERANCE)


class TestDirectionalTrendIndexInvalidParams(unittest.TestCase):
    """Test invalid parameter validation."""

    def test_q_too_small(self):
        with self.assertRaises(ValueError):
            DirectionalTrendIndex(DirectionalTrendIndexParams(q=0))

    def test_r_too_small(self):
        with self.assertRaises(ValueError):
            DirectionalTrendIndex(DirectionalTrendIndexParams(r=0))

    def test_s_too_small(self):
        with self.assertRaises(ValueError):
            DirectionalTrendIndex(DirectionalTrendIndexParams(s=0))

    def test_u_too_small(self):
        with self.assertRaises(ValueError):
            DirectionalTrendIndex(DirectionalTrendIndexParams(u=0))

    def test_ul_too_small(self):
        with self.assertRaises(ValueError):
            DirectionalTrendIndex(DirectionalTrendIndexParams(ul=0))


if __name__ == '__main__':
    unittest.main()
