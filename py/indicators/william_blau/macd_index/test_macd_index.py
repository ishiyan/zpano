import math
import unittest
from datetime import datetime

from py.indicators.william_blau.macd_index.macd_index import MacdIndex
from py.indicators.william_blau.macd_index.params import MacdIndexParams
from py.indicators.core.identifier import Identifier
from py.entities.scalar import Scalar

from .test_testdata import (
    INPUT_CLOSE,
    EXPECTED_R20_S5_U3, EXPECTED_R20_S5_U3_SIG_UL3,
    EXPECTED_R20_S5_U1, EXPECTED_R20_S5_U1_SIG_UL3,
    EXPECTED_R26_S12_U3, EXPECTED_R26_S12_U3_SIG_UL3,
    EXPECTED_R26_S12_U1, EXPECTED_R26_S12_U1_SIG_UL3,
    EXPECTED_R35_S5_U3, EXPECTED_R35_S5_U3_SIG_UL3,
    EXPECTED_R10_S3_U5, EXPECTED_R10_S3_U5_SIG_UL3,
    EXPECTED_R32_S12_U5, EXPECTED_R32_S12_U5_SIG_UL3,
    EXPECTED_R17_S8_U1, EXPECTED_R17_S8_U1_SIG_UL3,
    EXPECTED_R20_S10_U3, EXPECTED_R20_S10_U3_SIG_UL3,
    EXPECTED_R8_S4_U2, EXPECTED_R8_S4_U2_SIG_UL3,
    EXPECTED_R30_S15_U1, EXPECTED_R30_S15_U1_SIG_UL3,
    EXPECTED_R3_S2_U3, EXPECTED_R3_S2_U3_SIG_UL3,
    EXPECTED_R50_S12_U1, EXPECTED_R50_S12_U1_SIG_UL3,
    EXPECTED_R19_S6_U3, EXPECTED_R19_S6_U3_SIG_UL3,
    EXPECTED_R20_S5_U5, EXPECTED_R20_S5_U5_SIG_UL3,
    EXPECTED_R60_S30_U10, EXPECTED_R60_S30_U10_SIG_UL3,
)

TOLERANCE = 1e-10

# Signal-line EMA period for every expected signal array (Ergodic default).
UL = 3

# (r, s, u, expected_macdi, expected_signal)
COMBOS = [
    (20, 5, 3, EXPECTED_R20_S5_U3, EXPECTED_R20_S5_U3_SIG_UL3),
    (20, 5, 1, EXPECTED_R20_S5_U1, EXPECTED_R20_S5_U1_SIG_UL3),
    (26, 12, 3, EXPECTED_R26_S12_U3, EXPECTED_R26_S12_U3_SIG_UL3),
    (26, 12, 1, EXPECTED_R26_S12_U1, EXPECTED_R26_S12_U1_SIG_UL3),
    (35, 5, 3, EXPECTED_R35_S5_U3, EXPECTED_R35_S5_U3_SIG_UL3),
    (10, 3, 5, EXPECTED_R10_S3_U5, EXPECTED_R10_S3_U5_SIG_UL3),
    (32, 12, 5, EXPECTED_R32_S12_U5, EXPECTED_R32_S12_U5_SIG_UL3),
    (17, 8, 1, EXPECTED_R17_S8_U1, EXPECTED_R17_S8_U1_SIG_UL3),
    (20, 10, 3, EXPECTED_R20_S10_U3, EXPECTED_R20_S10_U3_SIG_UL3),
    (8, 4, 2, EXPECTED_R8_S4_U2, EXPECTED_R8_S4_U2_SIG_UL3),
    (30, 15, 1, EXPECTED_R30_S15_U1, EXPECTED_R30_S15_U1_SIG_UL3),
    (3, 2, 3, EXPECTED_R3_S2_U3, EXPECTED_R3_S2_U3_SIG_UL3),
    (50, 12, 1, EXPECTED_R50_S12_U1, EXPECTED_R50_S12_U1_SIG_UL3),
    (19, 6, 3, EXPECTED_R19_S6_U3, EXPECTED_R19_S6_U3_SIG_UL3),
    (20, 5, 5, EXPECTED_R20_S5_U5, EXPECTED_R20_S5_U5_SIG_UL3),
    (60, 30, 10, EXPECTED_R60_S30_U10, EXPECTED_R60_S30_U10_SIG_UL3),
]


class TestMacdIndexData(unittest.TestCase):
    """Test the MACD Index against the reference test data for all parameter combinations."""

    def test_all_combos(self):
        for r, s, u, exp_macdi, exp_signal in COMBOS:
            with self.subTest(r=r, s=s, u=u):
                ind = MacdIndex(MacdIndexParams(r=r, s=s, u=u, ul=UL))
                for i in range(len(INPUT_CLOSE)):
                    macdi, signal = ind.update(INPUT_CLOSE[i])

                    self.assertAlmostEqual(macdi, exp_macdi[i], delta=TOLERANCE,
                                           msg=f"[{i}] macdi: expected {exp_macdi[i]}, got {macdi}")
                    self.assertAlmostEqual(signal, exp_signal[i], delta=TOLERANCE,
                                           msg=f"[{i}] signal: expected {exp_signal[i]}, got {signal}")


class TestMacdIndexNoWarmUp(unittest.TestCase):
    """Test that both outputs are finite from bar 0 and bar 0 is exactly 0.0."""

    def test_bar_zero_is_zero(self):
        ind = MacdIndex(MacdIndexParams())
        macdi, signal = ind.update(INPUT_CLOSE[0])
        self.assertEqual(macdi, 0.0)
        self.assertEqual(signal, 0.0)

    def test_no_nan_anywhere(self):
        ind = MacdIndex(MacdIndexParams())
        for i in range(len(INPUT_CLOSE)):
            macdi, signal = ind.update(INPUT_CLOSE[i])
            self.assertFalse(math.isnan(macdi), f"[{i}] macdi: unexpected NaN")
            self.assertFalse(math.isnan(signal), f"[{i}] signal: unexpected NaN")


class TestMacdIndexPassthrough(unittest.TestCase):
    """Test the pure two-EMA form MacdIndex(2,1,1) with ul=1: macdi == EMA(c,1) - EMA(c,2)."""

    def test_passthrough(self):
        ind = MacdIndex(MacdIndexParams(r=2, s=1, u=1, ul=1))
        r0 = ind.update(10.0)  # bar 0 seeds both EMAs -> macd 0
        self.assertAlmostEqual(r0[0], 0.0, delta=TOLERANCE)
        self.assertAlmostEqual(r0[1], 0.0, delta=TOLERANCE)
        # EMA(1) is a passthrough -> fast = 12. EMA(2) = (2/3)*12 + (1/3)*10 = 11.3333.
        r1 = ind.update(12.0)
        self.assertAlmostEqual(r1[0], 12.0 - (2.0 / 3.0 * 12.0 + 1.0 / 3.0 * 10.0), delta=TOLERANCE)
        self.assertAlmostEqual(r1[1], r1[0], delta=TOLERANCE)


class TestMacdIndexPassthroughSignal(unittest.TestCase):
    """Test the ul=1 invariant: the signal line equals the index exactly."""

    def test_signal_equals_macdi(self):
        ind = MacdIndex(MacdIndexParams(r=20, s=5, u=3, ul=1))
        for i in range(len(INPUT_CLOSE)):
            macdi, signal = ind.update(INPUT_CLOSE[i])
            self.assertEqual(macdi, signal, f"[{i}] signal must equal macdi when ul=1")


class TestMacdIndexPrimed(unittest.TestCase):
    """Test priming: the MACD Index has no warm-up, so it is primed after the first update."""

    def test_is_primed(self):
        ind = MacdIndex(MacdIndexParams())
        self.assertFalse(ind.is_primed())
        ind.update(INPUT_CLOSE[0])
        self.assertTrue(ind.is_primed())


class TestMacdIndexMnemonic(unittest.TestCase):
    """Test mnemonic generation (ul is excluded from the mnemonic)."""

    def test_default_mnemonic(self):
        ind = MacdIndex(MacdIndexParams())
        self.assertEqual(ind.metadata().mnemonic, "macdi(20,5,3)")

    def test_custom_mnemonic(self):
        ind = MacdIndex(MacdIndexParams(r=26, s=12, u=9, ul=7))
        self.assertEqual(ind.metadata().mnemonic, "macdi(26,12,9)")


class TestMacdIndexMetadata(unittest.TestCase):
    """Test metadata generation."""

    def test_default_metadata(self):
        ind = MacdIndex(MacdIndexParams())
        meta = ind.metadata()
        self.assertEqual(meta.identifier, Identifier.MACD_INDEX)
        self.assertEqual(meta.mnemonic, "macdi(20,5,3)")
        self.assertEqual(len(meta.outputs), 2)


class TestMacdIndexUpdateScalar(unittest.TestCase):
    """Test update_scalar output ordering (macdi, signal)."""

    def test_update_scalar(self):
        ind = MacdIndex(MacdIndexParams(r=20, s=5, u=3, ul=UL))
        tm = datetime(2021, 4, 1)
        out = None
        for i in range(len(INPUT_CLOSE)):
            out = ind.update_scalar(Scalar(time=tm, value=INPUT_CLOSE[i]))
        self.assertEqual(len(out), 2)
        self.assertAlmostEqual(out[0].value, EXPECTED_R20_S5_U3[-1], delta=TOLERANCE)
        self.assertAlmostEqual(out[1].value, EXPECTED_R20_S5_U3_SIG_UL3[-1], delta=TOLERANCE)


class TestMacdIndexInvalidParams(unittest.TestCase):
    """Test invalid parameter validation."""

    def test_r_too_small(self):
        with self.assertRaises(ValueError):
            MacdIndex(MacdIndexParams(r=0))

    def test_s_too_small(self):
        with self.assertRaises(ValueError):
            MacdIndex(MacdIndexParams(s=0))

    def test_u_too_small(self):
        with self.assertRaises(ValueError):
            MacdIndex(MacdIndexParams(u=0))

    def test_ul_too_small(self):
        with self.assertRaises(ValueError):
            MacdIndex(MacdIndexParams(ul=0))

    def test_s_not_less_than_r(self):
        with self.assertRaises(ValueError):
            MacdIndex(MacdIndexParams(r=5, s=5))

    def test_s_greater_than_r(self):
        with self.assertRaises(ValueError):
            MacdIndex(MacdIndexParams(r=5, s=6))


if __name__ == '__main__':
    unittest.main()
