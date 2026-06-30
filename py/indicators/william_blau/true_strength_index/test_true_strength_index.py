import math
import unittest
from datetime import datetime

from py.indicators.william_blau.true_strength_index.true_strength_index import TrueStrengthIndex
from py.indicators.william_blau.true_strength_index.params import TrueStrengthIndexParams
from py.indicators.core.identifier import Identifier
from py.entities.scalar import Scalar

from .test_testdata import (
    INPUT_CLOSE,
    EXPECTED_Q2_R20_S5_U3, EXPECTED_Q2_R20_S5_U3_SIG_UL3,
    EXPECTED_Q2_R25_S13_U1, EXPECTED_Q2_R25_S13_U1_SIG_UL3,
    EXPECTED_Q2_R20_S5_U1, EXPECTED_Q2_R20_S5_U1_SIG_UL3,
    EXPECTED_Q2_R32_S5_U1, EXPECTED_Q2_R32_S5_U1_SIG_UL3,
    EXPECTED_Q2_R13_S13_U1, EXPECTED_Q2_R13_S13_U1_SIG_UL3,
    EXPECTED_Q2_R20_S40_U1, EXPECTED_Q2_R20_S40_U1_SIG_UL3,
    EXPECTED_Q2_R40_S20_U1, EXPECTED_Q2_R40_S20_U1_SIG_UL3,
    EXPECTED_Q2_R64_S64_U1, EXPECTED_Q2_R64_S64_U1_SIG_UL3,
    EXPECTED_Q2_R100_S5_U1, EXPECTED_Q2_R100_S5_U1_SIG_UL3,
    EXPECTED_Q2_R1_S1_U1, EXPECTED_Q2_R1_S1_U1_SIG_UL3,
    EXPECTED_Q2_R1_S5_U3, EXPECTED_Q2_R1_S5_U3_SIG_UL3,
    EXPECTED_Q2_R20_S1_U1, EXPECTED_Q2_R20_S1_U1_SIG_UL3,
    EXPECTED_Q2_R5_S5_U5, EXPECTED_Q2_R5_S5_U5_SIG_UL3,
    EXPECTED_Q3_R20_S5_U3, EXPECTED_Q3_R20_S5_U3_SIG_UL3,
    EXPECTED_Q5_R20_S5_U3, EXPECTED_Q5_R20_S5_U3_SIG_UL3,
    EXPECTED_Q10_R20_S5_U1, EXPECTED_Q10_R20_S5_U1_SIG_UL3,
    EXPECTED_Q2_R9_S3_U1, EXPECTED_Q2_R9_S3_U1_SIG_UL3,
    EXPECTED_Q2_R7_S4_U2, EXPECTED_Q2_R7_S4_U2_SIG_UL3,
)

TOLERANCE = 1e-10

# Signal-line EMA period for every expected signal array (Ergodic default).
UL = 3

# (q, r, s, u, expected_tsi, expected_signal)
COMBOS = [
    (2, 20, 5, 3, EXPECTED_Q2_R20_S5_U3, EXPECTED_Q2_R20_S5_U3_SIG_UL3),
    (2, 25, 13, 1, EXPECTED_Q2_R25_S13_U1, EXPECTED_Q2_R25_S13_U1_SIG_UL3),
    (2, 20, 5, 1, EXPECTED_Q2_R20_S5_U1, EXPECTED_Q2_R20_S5_U1_SIG_UL3),
    (2, 32, 5, 1, EXPECTED_Q2_R32_S5_U1, EXPECTED_Q2_R32_S5_U1_SIG_UL3),
    (2, 13, 13, 1, EXPECTED_Q2_R13_S13_U1, EXPECTED_Q2_R13_S13_U1_SIG_UL3),
    (2, 20, 40, 1, EXPECTED_Q2_R20_S40_U1, EXPECTED_Q2_R20_S40_U1_SIG_UL3),
    (2, 40, 20, 1, EXPECTED_Q2_R40_S20_U1, EXPECTED_Q2_R40_S20_U1_SIG_UL3),
    (2, 64, 64, 1, EXPECTED_Q2_R64_S64_U1, EXPECTED_Q2_R64_S64_U1_SIG_UL3),
    (2, 100, 5, 1, EXPECTED_Q2_R100_S5_U1, EXPECTED_Q2_R100_S5_U1_SIG_UL3),
    (2, 1, 1, 1, EXPECTED_Q2_R1_S1_U1, EXPECTED_Q2_R1_S1_U1_SIG_UL3),
    (2, 1, 5, 3, EXPECTED_Q2_R1_S5_U3, EXPECTED_Q2_R1_S5_U3_SIG_UL3),
    (2, 20, 1, 1, EXPECTED_Q2_R20_S1_U1, EXPECTED_Q2_R20_S1_U1_SIG_UL3),
    (2, 5, 5, 5, EXPECTED_Q2_R5_S5_U5, EXPECTED_Q2_R5_S5_U5_SIG_UL3),
    (3, 20, 5, 3, EXPECTED_Q3_R20_S5_U3, EXPECTED_Q3_R20_S5_U3_SIG_UL3),
    (5, 20, 5, 3, EXPECTED_Q5_R20_S5_U3, EXPECTED_Q5_R20_S5_U3_SIG_UL3),
    (10, 20, 5, 1, EXPECTED_Q10_R20_S5_U1, EXPECTED_Q10_R20_S5_U1_SIG_UL3),
    (2, 9, 3, 1, EXPECTED_Q2_R9_S3_U1, EXPECTED_Q2_R9_S3_U1_SIG_UL3),
    (2, 7, 4, 2, EXPECTED_Q2_R7_S4_U2, EXPECTED_Q2_R7_S4_U2_SIG_UL3),
]


class TestTrueStrengthIndexData(unittest.TestCase):
    """Test TSI against the reference test data for all parameter combinations."""

    def test_all_combos(self):
        for q, r, s, u, exp_tsi, exp_signal in COMBOS:
            with self.subTest(q=q, r=r, s=s, u=u):
                ind = TrueStrengthIndex(TrueStrengthIndexParams(
                    q=q, r=r, s=s, u=u, ul=UL))
                for i in range(len(INPUT_CLOSE)):
                    tsi, signal = ind.update(INPUT_CLOSE[i])

                    if math.isnan(exp_tsi[i]):
                        self.assertTrue(math.isnan(tsi), f"[{i}] tsi: expected NaN, got {tsi}")
                    else:
                        self.assertAlmostEqual(tsi, exp_tsi[i], delta=TOLERANCE,
                                               msg=f"[{i}] tsi: expected {exp_tsi[i]}, got {tsi}")

                    if math.isnan(exp_signal[i]):
                        self.assertTrue(math.isnan(signal), f"[{i}] signal: expected NaN, got {signal}")
                    else:
                        self.assertAlmostEqual(signal, exp_signal[i], delta=TOLERANCE,
                                               msg=f"[{i}] signal: expected {exp_signal[i]}, got {signal}")


class TestTrueStrengthIndexPassthrough(unittest.TestCase):
    """Test the all-passthrough invariant TSI(2,1,1,1) = sign(mtm)*100."""

    def test_passthrough(self):
        ind = TrueStrengthIndex(TrueStrengthIndexParams(q=2, r=1, s=1, u=1, ul=1))
        prices = [10.0, 12.0, 11.0, 11.0, 13.0]
        r0 = ind.update(prices[0])
        self.assertTrue(math.isnan(r0[0]))
        self.assertTrue(math.isnan(r0[1]))
        r1 = ind.update(prices[1])  # mtm=+2 -> +100, ul=1 passthrough
        self.assertAlmostEqual(r1[0], 100.0, delta=TOLERANCE)
        self.assertAlmostEqual(r1[1], 100.0, delta=TOLERANCE)
        r2 = ind.update(prices[2])  # mtm=-1 -> -100
        self.assertAlmostEqual(r2[0], -100.0, delta=TOLERANCE)
        self.assertAlmostEqual(r2[1], -100.0, delta=TOLERANCE)
        r3 = ind.update(prices[3])  # mtm=0 -> division guard 0.0
        self.assertAlmostEqual(r3[0], 0.0, delta=TOLERANCE)


class TestTrueStrengthIndexMnemonic(unittest.TestCase):
    """Test mnemonic generation (ul is excluded from the mnemonic)."""

    def test_default_mnemonic(self):
        ind = TrueStrengthIndex(TrueStrengthIndexParams())
        self.assertEqual(ind.metadata().mnemonic, "tsi(2,20,5,3)")

    def test_custom_mnemonic(self):
        ind = TrueStrengthIndex(TrueStrengthIndexParams(q=2, r=25, s=13, u=1, ul=7))
        self.assertEqual(ind.metadata().mnemonic, "tsi(2,25,13,1)")


class TestTrueStrengthIndexMetadata(unittest.TestCase):
    """Test metadata generation."""

    def test_default_metadata(self):
        ind = TrueStrengthIndex(TrueStrengthIndexParams())
        meta = ind.metadata()
        self.assertEqual(meta.identifier, Identifier.TRUE_STRENGTH_INDEX)
        self.assertEqual(meta.mnemonic, "tsi(2,20,5,3)")
        self.assertEqual(len(meta.outputs), 2)


class TestTrueStrengthIndexUpdateScalar(unittest.TestCase):
    """Test update_scalar output ordering (tsi, signal)."""

    def test_update_scalar(self):
        ind = TrueStrengthIndex(TrueStrengthIndexParams(q=2, r=20, s=5, u=3, ul=UL))
        tm = datetime(2021, 4, 1)
        out = None
        for i in range(len(INPUT_CLOSE)):
            out = ind.update_scalar(Scalar(time=tm, value=INPUT_CLOSE[i]))
        self.assertEqual(len(out), 2)
        self.assertAlmostEqual(out[0].value, EXPECTED_Q2_R20_S5_U3[-1], delta=TOLERANCE)
        self.assertAlmostEqual(out[1].value, EXPECTED_Q2_R20_S5_U3_SIG_UL3[-1], delta=TOLERANCE)


class TestTrueStrengthIndexInvalidParams(unittest.TestCase):
    """Test invalid parameter validation."""

    def test_q_too_small(self):
        with self.assertRaises(ValueError):
            TrueStrengthIndex(TrueStrengthIndexParams(q=0))

    def test_r_too_small(self):
        with self.assertRaises(ValueError):
            TrueStrengthIndex(TrueStrengthIndexParams(r=0))

    def test_s_too_small(self):
        with self.assertRaises(ValueError):
            TrueStrengthIndex(TrueStrengthIndexParams(s=0))

    def test_u_too_small(self):
        with self.assertRaises(ValueError):
            TrueStrengthIndex(TrueStrengthIndexParams(u=0))

    def test_ul_too_small(self):
        with self.assertRaises(ValueError):
            TrueStrengthIndex(TrueStrengthIndexParams(ul=0))


if __name__ == '__main__':
    unittest.main()
