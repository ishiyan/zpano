import math
import unittest
from datetime import datetime

from py.indicators.william_blau.candlestick_strength_index.candlestick_strength_index import CandlestickStrengthIndex
from py.indicators.william_blau.candlestick_strength_index.params import CandlestickStrengthIndexParams
from py.indicators.core.identifier import Identifier
from py.entities.bar import Bar
from py.entities.quote import Quote
from py.entities.scalar import Scalar
from py.entities.trade import Trade

from .test_testdata import (
    INPUT_OPEN, INPUT_HIGH, INPUT_LOW, INPUT_CLOSE,
    EXPECTED_R20_S5_U3, EXPECTED_R20_S5_U3_SIG_UL3,
    EXPECTED_R32_S32_U1, EXPECTED_R32_S32_U1_SIG_UL3,
    EXPECTED_R1_S1_U1, EXPECTED_R1_S1_U1_SIG_UL3,
    EXPECTED_R25_S13_U1, EXPECTED_R25_S13_U1_SIG_UL3,
    EXPECTED_R13_S13_U1, EXPECTED_R13_S13_U1_SIG_UL3,
    EXPECTED_R5_S5_U5, EXPECTED_R5_S5_U5_SIG_UL3,
    EXPECTED_R9_S3_U1, EXPECTED_R9_S3_U1_SIG_UL3,
    EXPECTED_R64_S64_U1, EXPECTED_R64_S64_U1_SIG_UL3,
    EXPECTED_R32_S32_U3, EXPECTED_R32_S32_U3_SIG_UL3,
    EXPECTED_R40_S20_U1, EXPECTED_R40_S20_U1_SIG_UL3,
    EXPECTED_R2_S2_U2, EXPECTED_R2_S2_U2_SIG_UL3,
    EXPECTED_R7_S4_U2, EXPECTED_R7_S4_U2_SIG_UL3,
    EXPECTED_R12_S12_U12, EXPECTED_R12_S12_U12_SIG_UL3,
    EXPECTED_R3_S10_U10, EXPECTED_R3_S10_U10_SIG_UL3,
    EXPECTED_R50_S1_U1, EXPECTED_R50_S1_U1_SIG_UL3,
    EXPECTED_R32_S5_U3, EXPECTED_R32_S5_U3_SIG_UL3,
)

TOLERANCE = 1e-10

# Signal-line EMA period for every expected signal array (Ergodic default).
UL = 3

# (r, s, u, expected_csi, expected_signal)
COMBOS = [
    (20, 5, 3, EXPECTED_R20_S5_U3, EXPECTED_R20_S5_U3_SIG_UL3),
    (32, 32, 1, EXPECTED_R32_S32_U1, EXPECTED_R32_S32_U1_SIG_UL3),
    (1, 1, 1, EXPECTED_R1_S1_U1, EXPECTED_R1_S1_U1_SIG_UL3),
    (25, 13, 1, EXPECTED_R25_S13_U1, EXPECTED_R25_S13_U1_SIG_UL3),
    (13, 13, 1, EXPECTED_R13_S13_U1, EXPECTED_R13_S13_U1_SIG_UL3),
    (5, 5, 5, EXPECTED_R5_S5_U5, EXPECTED_R5_S5_U5_SIG_UL3),
    (9, 3, 1, EXPECTED_R9_S3_U1, EXPECTED_R9_S3_U1_SIG_UL3),
    (64, 64, 1, EXPECTED_R64_S64_U1, EXPECTED_R64_S64_U1_SIG_UL3),
    (32, 32, 3, EXPECTED_R32_S32_U3, EXPECTED_R32_S32_U3_SIG_UL3),
    (40, 20, 1, EXPECTED_R40_S20_U1, EXPECTED_R40_S20_U1_SIG_UL3),
    (2, 2, 2, EXPECTED_R2_S2_U2, EXPECTED_R2_S2_U2_SIG_UL3),
    (7, 4, 2, EXPECTED_R7_S4_U2, EXPECTED_R7_S4_U2_SIG_UL3),
    (12, 12, 12, EXPECTED_R12_S12_U12, EXPECTED_R12_S12_U12_SIG_UL3),
    (3, 10, 10, EXPECTED_R3_S10_U10, EXPECTED_R3_S10_U10_SIG_UL3),
    (50, 1, 1, EXPECTED_R50_S1_U1, EXPECTED_R50_S1_U1_SIG_UL3),
    (32, 5, 3, EXPECTED_R32_S5_U3, EXPECTED_R32_S5_U3_SIG_UL3),
]


class TestCandlestickStrengthIndexData(unittest.TestCase):
    """Test CSI against the reference test data for all parameter combinations."""

    def test_all_combos(self):
        for r, s, u, exp_csi, exp_signal in COMBOS:
            with self.subTest(r=r, s=s, u=u):
                ind = CandlestickStrengthIndex(CandlestickStrengthIndexParams(
                    r=r, s=s, u=u, ul=UL))
                for i in range(len(INPUT_CLOSE)):
                    csi, signal = ind.update(INPUT_OPEN[i], INPUT_HIGH[i],
                                             INPUT_LOW[i], INPUT_CLOSE[i])

                    if math.isnan(exp_csi[i]):
                        self.assertTrue(math.isnan(csi), f"[{i}] csi: expected NaN, got {csi}")
                    else:
                        self.assertAlmostEqual(csi, exp_csi[i], delta=TOLERANCE,
                                               msg=f"[{i}] csi: expected {exp_csi[i]}, got {csi}")

                    if math.isnan(exp_signal[i]):
                        self.assertTrue(math.isnan(signal), f"[{i}] signal: expected NaN, got {signal}")
                    else:
                        self.assertAlmostEqual(signal, exp_signal[i], delta=TOLERANCE,
                                               msg=f"[{i}] signal: expected {exp_signal[i]}, got {signal}")


class TestCandlestickStrengthIndexPassthrough(unittest.TestCase):
    """Test the all-passthrough invariant CSI(1,1,1) = 100*(close-open)/(high-low)."""

    def test_passthrough(self):
        ind = CandlestickStrengthIndex(CandlestickStrengthIndexParams(r=1, s=1, u=1, ul=1))
        r0 = ind.update(10.0, 12.0, 10.0, 12.0)  # body == range -> +100
        self.assertAlmostEqual(r0[0], 100.0, delta=TOLERANCE)
        self.assertAlmostEqual(r0[1], 100.0, delta=TOLERANCE)
        r1 = ind.update(12.0, 12.0, 10.0, 10.0)  # full bearish body -> -100
        self.assertAlmostEqual(r1[0], -100.0, delta=TOLERANCE)
        self.assertAlmostEqual(r1[1], -100.0, delta=TOLERANCE)
        r2 = ind.update(11.0, 11.0, 11.0, 11.0)  # zero range -> division guard 0.0
        self.assertAlmostEqual(r2[0], 0.0, delta=TOLERANCE)
        self.assertAlmostEqual(r2[1], 0.0, delta=TOLERANCE)


class TestCandlestickStrengthIndexPrimed(unittest.TestCase):
    """The CSI has no NaN warm-up region: it is primed after the first bar."""

    def test_primed(self):
        ind = CandlestickStrengthIndex(CandlestickStrengthIndexParams())
        self.assertFalse(ind.is_primed())
        ind.update(INPUT_OPEN[0], INPUT_HIGH[0], INPUT_LOW[0], INPUT_CLOSE[0])
        self.assertTrue(ind.is_primed())


class TestCandlestickStrengthIndexMnemonic(unittest.TestCase):
    """Test mnemonic generation."""

    def test_default_mnemonic(self):
        ind = CandlestickStrengthIndex(CandlestickStrengthIndexParams())
        self.assertEqual(ind.metadata().mnemonic, "csi(20,5,3,3)")

    def test_custom_mnemonic(self):
        ind = CandlestickStrengthIndex(CandlestickStrengthIndexParams(r=25, s=13, u=1, ul=7))
        self.assertEqual(ind.metadata().mnemonic, "csi(25,13,1,7)")


class TestCandlestickStrengthIndexMetadata(unittest.TestCase):
    """Test metadata generation."""

    def test_default_metadata(self):
        ind = CandlestickStrengthIndex(CandlestickStrengthIndexParams())
        meta = ind.metadata()
        self.assertEqual(meta.identifier, Identifier.CANDLESTICK_STRENGTH_INDEX)
        self.assertEqual(meta.mnemonic, "csi(20,5,3,3)")
        self.assertEqual(meta.description, "Candlestick Strength Index csi(20,5,3,3)")
        self.assertEqual(len(meta.outputs), 2)


class TestCandlestickStrengthIndexUpdateEntities(unittest.TestCase):
    """Test the entity update methods and their OHLC mapping."""

    def test_update_bar(self):
        ind = CandlestickStrengthIndex(CandlestickStrengthIndexParams(r=20, s=5, u=3, ul=UL))
        tm = datetime(2021, 4, 1)
        out = None
        for i in range(len(INPUT_CLOSE)):
            out = ind.update_bar(Bar(time=tm, open=INPUT_OPEN[i], high=INPUT_HIGH[i],
                                     low=INPUT_LOW[i], close=INPUT_CLOSE[i], volume=0.0))
        self.assertEqual(len(out), 2)
        self.assertAlmostEqual(out[0].value, EXPECTED_R20_S5_U3[-1], delta=TOLERANCE)
        self.assertAlmostEqual(out[1].value, EXPECTED_R20_S5_U3_SIG_UL3[-1], delta=TOLERANCE)

    def test_update_quote(self):
        # A quote maps bid -> open and low, ask -> close and high, so the body
        # and the range are both the spread.
        ind = CandlestickStrengthIndex(CandlestickStrengthIndexParams(r=1, s=1, u=1, ul=1))
        tm = datetime(2021, 4, 1)
        out = ind.update_quote(Quote(time=tm, bid_price=10.0, ask_price=12.0,
                                     bid_size=1.0, ask_size=1.0))
        self.assertEqual(len(out), 2)
        self.assertAlmostEqual(out[0].value, 100.0, delta=TOLERANCE)
        self.assertAlmostEqual(out[1].value, 100.0, delta=TOLERANCE)

    def test_update_scalar(self):
        # A scalar carries one value, so the body and the range are both zero.
        ind = CandlestickStrengthIndex(CandlestickStrengthIndexParams(r=1, s=1, u=1, ul=1))
        tm = datetime(2021, 4, 1)
        out = ind.update_scalar(Scalar(time=tm, value=10.0))
        self.assertEqual(len(out), 2)
        self.assertAlmostEqual(out[0].value, 0.0, delta=TOLERANCE)
        self.assertAlmostEqual(out[1].value, 0.0, delta=TOLERANCE)

    def test_update_trade(self):
        # A trade carries one price, so the body and the range are both zero.
        ind = CandlestickStrengthIndex(CandlestickStrengthIndexParams(r=1, s=1, u=1, ul=1))
        tm = datetime(2021, 4, 1)
        out = ind.update_trade(Trade(time=tm, price=10.0, volume=1.0))
        self.assertEqual(len(out), 2)
        self.assertAlmostEqual(out[0].value, 0.0, delta=TOLERANCE)
        self.assertAlmostEqual(out[1].value, 0.0, delta=TOLERANCE)


class TestCandlestickStrengthIndexInvalidParams(unittest.TestCase):
    """Test invalid parameter validation."""

    def test_r_too_small(self):
        with self.assertRaises(ValueError):
            CandlestickStrengthIndex(CandlestickStrengthIndexParams(r=0))

    def test_s_too_small(self):
        with self.assertRaises(ValueError):
            CandlestickStrengthIndex(CandlestickStrengthIndexParams(s=0))

    def test_u_too_small(self):
        with self.assertRaises(ValueError):
            CandlestickStrengthIndex(CandlestickStrengthIndexParams(u=0))

    def test_ul_too_small(self):
        with self.assertRaises(ValueError):
            CandlestickStrengthIndex(CandlestickStrengthIndexParams(ul=0))


if __name__ == '__main__':
    unittest.main()
