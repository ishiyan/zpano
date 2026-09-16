import math
import unittest
from datetime import datetime

from py.indicators.william_blau.candlestick_momentum_index.candlestick_momentum_index import CandlestickMomentumIndex
from py.indicators.william_blau.candlestick_momentum_index.params import CandlestickMomentumIndexParams
from py.indicators.core.identifier import Identifier
from py.entities.bar import Bar
from py.entities.quote import Quote
from py.entities.scalar import Scalar
from py.entities.trade import Trade

from .test_testdata import (
    INPUT_OPEN, INPUT_CLOSE,
    EXPECTED_R20_S5_U3, EXPECTED_R20_S5_U3_SIG_UL3,
    EXPECTED_R20_S5_U1, EXPECTED_R20_S5_U1_SIG_UL3,
    EXPECTED_R1_S1_U1, EXPECTED_R1_S1_U1_SIG_UL3,
    EXPECTED_R25_S13_U1, EXPECTED_R25_S13_U1_SIG_UL3,
    EXPECTED_R13_S13_U1, EXPECTED_R13_S13_U1_SIG_UL3,
    EXPECTED_R5_S5_U5, EXPECTED_R5_S5_U5_SIG_UL3,
    EXPECTED_R9_S3_U1, EXPECTED_R9_S3_U1_SIG_UL3,
    EXPECTED_R64_S64_U1, EXPECTED_R64_S64_U1_SIG_UL3,
    EXPECTED_R32_S5_U3, EXPECTED_R32_S5_U3_SIG_UL3,
    EXPECTED_R40_S20_U1, EXPECTED_R40_S20_U1_SIG_UL3,
    EXPECTED_R2_S2_U2, EXPECTED_R2_S2_U2_SIG_UL3,
    EXPECTED_R7_S4_U2, EXPECTED_R7_S4_U2_SIG_UL3,
    EXPECTED_R12_S12_U12, EXPECTED_R12_S12_U12_SIG_UL3,
    EXPECTED_R3_S10_U10, EXPECTED_R3_S10_U10_SIG_UL3,
    EXPECTED_R50_S1_U1, EXPECTED_R50_S1_U1_SIG_UL3,
    EXPECTED_R20_S5_U5, EXPECTED_R20_S5_U5_SIG_UL3,
)

TOLERANCE = 1e-10

# Signal-line EMA period for every expected signal array (Ergodic default).
UL = 3

# (r, s, u, expected_cmi, expected_signal)
COMBOS = [
    (20, 5, 3, EXPECTED_R20_S5_U3, EXPECTED_R20_S5_U3_SIG_UL3),
    (20, 5, 1, EXPECTED_R20_S5_U1, EXPECTED_R20_S5_U1_SIG_UL3),
    (1, 1, 1, EXPECTED_R1_S1_U1, EXPECTED_R1_S1_U1_SIG_UL3),
    (25, 13, 1, EXPECTED_R25_S13_U1, EXPECTED_R25_S13_U1_SIG_UL3),
    (13, 13, 1, EXPECTED_R13_S13_U1, EXPECTED_R13_S13_U1_SIG_UL3),
    (5, 5, 5, EXPECTED_R5_S5_U5, EXPECTED_R5_S5_U5_SIG_UL3),
    (9, 3, 1, EXPECTED_R9_S3_U1, EXPECTED_R9_S3_U1_SIG_UL3),
    (64, 64, 1, EXPECTED_R64_S64_U1, EXPECTED_R64_S64_U1_SIG_UL3),
    (32, 5, 3, EXPECTED_R32_S5_U3, EXPECTED_R32_S5_U3_SIG_UL3),
    (40, 20, 1, EXPECTED_R40_S20_U1, EXPECTED_R40_S20_U1_SIG_UL3),
    (2, 2, 2, EXPECTED_R2_S2_U2, EXPECTED_R2_S2_U2_SIG_UL3),
    (7, 4, 2, EXPECTED_R7_S4_U2, EXPECTED_R7_S4_U2_SIG_UL3),
    (12, 12, 12, EXPECTED_R12_S12_U12, EXPECTED_R12_S12_U12_SIG_UL3),
    (3, 10, 10, EXPECTED_R3_S10_U10, EXPECTED_R3_S10_U10_SIG_UL3),
    (50, 1, 1, EXPECTED_R50_S1_U1, EXPECTED_R50_S1_U1_SIG_UL3),
    (20, 5, 5, EXPECTED_R20_S5_U5, EXPECTED_R20_S5_U5_SIG_UL3),
]


class TestCandlestickMomentumIndexData(unittest.TestCase):
    """Test CMI against the reference test data for all parameter combinations."""

    def test_all_combos(self):
        for r, s, u, exp_cmi, exp_signal in COMBOS:
            with self.subTest(r=r, s=s, u=u):
                ind = CandlestickMomentumIndex(CandlestickMomentumIndexParams(
                    r=r, s=s, u=u, ul=UL))
                for i in range(len(INPUT_CLOSE)):
                    cmi, signal = ind.update(INPUT_OPEN[i], INPUT_CLOSE[i])

                    if math.isnan(exp_cmi[i]):
                        self.assertTrue(math.isnan(cmi), f"[{i}] cmi: expected NaN, got {cmi}")
                    else:
                        self.assertAlmostEqual(cmi, exp_cmi[i], delta=TOLERANCE,
                                               msg=f"[{i}] cmi: expected {exp_cmi[i]}, got {cmi}")

                    if math.isnan(exp_signal[i]):
                        self.assertTrue(math.isnan(signal), f"[{i}] signal: expected NaN, got {signal}")
                    else:
                        self.assertAlmostEqual(signal, exp_signal[i], delta=TOLERANCE,
                                               msg=f"[{i}] signal: expected {exp_signal[i]}, got {signal}")


class TestCandlestickMomentumIndexPassthrough(unittest.TestCase):
    """Test the all-passthrough invariant CMI(1,1,1) = sign(close-open)*100."""

    def test_passthrough(self):
        ind = CandlestickMomentumIndex(CandlestickMomentumIndexParams(r=1, s=1, u=1, ul=1))
        r0 = ind.update(10.0, 12.0)  # up candle +2 -> +100, ul=1 passthrough
        self.assertAlmostEqual(r0[0], 100.0, delta=TOLERANCE)
        self.assertAlmostEqual(r0[1], 100.0, delta=TOLERANCE)
        r1 = ind.update(12.0, 11.0)  # down candle -1 -> -100
        self.assertAlmostEqual(r1[0], -100.0, delta=TOLERANCE)
        self.assertAlmostEqual(r1[1], -100.0, delta=TOLERANCE)
        r2 = ind.update(11.0, 11.0)  # doji 0 -> division guard 0.0
        self.assertAlmostEqual(r2[0], 0.0, delta=TOLERANCE)


class TestCandlestickMomentumIndexPrimed(unittest.TestCase):
    """The CMI has no NaN warm-up region: it is primed after the first bar."""

    def test_primed(self):
        ind = CandlestickMomentumIndex(CandlestickMomentumIndexParams())
        self.assertFalse(ind.is_primed())
        ind.update(INPUT_OPEN[0], INPUT_CLOSE[0])
        self.assertTrue(ind.is_primed())


class TestCandlestickMomentumIndexMnemonic(unittest.TestCase):
    """Test mnemonic generation."""

    def test_default_mnemonic(self):
        ind = CandlestickMomentumIndex(CandlestickMomentumIndexParams())
        self.assertEqual(ind.metadata().mnemonic, "cmi(20,5,3,3)")

    def test_custom_mnemonic(self):
        ind = CandlestickMomentumIndex(CandlestickMomentumIndexParams(r=25, s=13, u=1, ul=7))
        self.assertEqual(ind.metadata().mnemonic, "cmi(25,13,1,7)")


class TestCandlestickMomentumIndexMetadata(unittest.TestCase):
    """Test metadata generation."""

    def test_default_metadata(self):
        ind = CandlestickMomentumIndex(CandlestickMomentumIndexParams())
        meta = ind.metadata()
        self.assertEqual(meta.identifier, Identifier.CANDLESTICK_MOMENTUM_INDEX)
        self.assertEqual(meta.mnemonic, "cmi(20,5,3,3)")
        self.assertEqual(meta.description, "Candlestick Momentum Index cmi(20,5,3,3)")
        self.assertEqual(len(meta.outputs), 2)


class TestCandlestickMomentumIndexUpdateEntities(unittest.TestCase):
    """Test the entity update methods and their open/close mapping."""

    def test_update_bar(self):
        ind = CandlestickMomentumIndex(CandlestickMomentumIndexParams(r=20, s=5, u=3, ul=UL))
        tm = datetime(2021, 4, 1)
        out = None
        for i in range(len(INPUT_CLOSE)):
            out = ind.update_bar(Bar(time=tm, open=INPUT_OPEN[i], high=INPUT_CLOSE[i],
                                     low=INPUT_OPEN[i], close=INPUT_CLOSE[i], volume=0.0))
        self.assertEqual(len(out), 2)
        self.assertAlmostEqual(out[0].value, EXPECTED_R20_S5_U3[-1], delta=TOLERANCE)
        self.assertAlmostEqual(out[1].value, EXPECTED_R20_S5_U3_SIG_UL3[-1], delta=TOLERANCE)

    def test_update_quote(self):
        # A quote maps bid -> open, ask -> close, so the body is the spread.
        ind = CandlestickMomentumIndex(CandlestickMomentumIndexParams(r=1, s=1, u=1, ul=1))
        tm = datetime(2021, 4, 1)
        out = ind.update_quote(Quote(time=tm, bid_price=10.0, ask_price=12.0,
                                     bid_size=1.0, ask_size=1.0))
        self.assertEqual(len(out), 2)
        self.assertAlmostEqual(out[0].value, 100.0, delta=TOLERANCE)
        self.assertAlmostEqual(out[1].value, 100.0, delta=TOLERANCE)

    def test_update_scalar(self):
        # A scalar carries one value, so open == close and the body is zero.
        ind = CandlestickMomentumIndex(CandlestickMomentumIndexParams(r=1, s=1, u=1, ul=1))
        tm = datetime(2021, 4, 1)
        out = ind.update_scalar(Scalar(time=tm, value=10.0))
        self.assertEqual(len(out), 2)
        self.assertAlmostEqual(out[0].value, 0.0, delta=TOLERANCE)
        self.assertAlmostEqual(out[1].value, 0.0, delta=TOLERANCE)

    def test_update_trade(self):
        # A trade carries one price, so open == close and the body is zero.
        ind = CandlestickMomentumIndex(CandlestickMomentumIndexParams(r=1, s=1, u=1, ul=1))
        tm = datetime(2021, 4, 1)
        out = ind.update_trade(Trade(time=tm, price=10.0, volume=1.0))
        self.assertEqual(len(out), 2)
        self.assertAlmostEqual(out[0].value, 0.0, delta=TOLERANCE)
        self.assertAlmostEqual(out[1].value, 0.0, delta=TOLERANCE)


class TestCandlestickMomentumIndexInvalidParams(unittest.TestCase):
    """Test invalid parameter validation."""

    def test_r_too_small(self):
        with self.assertRaises(ValueError):
            CandlestickMomentumIndex(CandlestickMomentumIndexParams(r=0))

    def test_s_too_small(self):
        with self.assertRaises(ValueError):
            CandlestickMomentumIndex(CandlestickMomentumIndexParams(s=0))

    def test_u_too_small(self):
        with self.assertRaises(ValueError):
            CandlestickMomentumIndex(CandlestickMomentumIndexParams(u=0))

    def test_ul_too_small(self):
        with self.assertRaises(ValueError):
            CandlestickMomentumIndex(CandlestickMomentumIndexParams(ul=0))


if __name__ == '__main__':
    unittest.main()
