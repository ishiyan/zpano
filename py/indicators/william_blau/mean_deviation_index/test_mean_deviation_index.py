import math
import unittest
from datetime import datetime

from py.indicators.william_blau.mean_deviation_index.mean_deviation_index import MeanDeviationIndex
from py.indicators.william_blau.mean_deviation_index.params import MeanDeviationIndexParams
from py.indicators.core.identifier import Identifier
from py.entities.scalar import Scalar

from .test_testdata import (
    INPUT_CLOSE,
    EXPECTED_R20_S5_U3, EXPECTED_R20_S5_U3_SIG_UL3,
    EXPECTED_R20_S5_U1, EXPECTED_R20_S5_U1_SIG_UL3,
    EXPECTED_R1_S5_U3, EXPECTED_R1_S5_U3_SIG_UL3,
    EXPECTED_R40_S5_U3, EXPECTED_R40_S5_U3_SIG_UL3,
    EXPECTED_R10_S5_U3, EXPECTED_R10_S5_U3_SIG_UL3,
    EXPECTED_R5_S5_U5, EXPECTED_R5_S5_U5_SIG_UL3,
    EXPECTED_R20_S9_U1, EXPECTED_R20_S9_U1_SIG_UL3,
    EXPECTED_R26_S12_U9, EXPECTED_R26_S12_U9_SIG_UL3,
    EXPECTED_R50_S13_U1, EXPECTED_R50_S13_U1_SIG_UL3,
    EXPECTED_R30_S5_U3, EXPECTED_R30_S5_U3_SIG_UL3,
    EXPECTED_R3_S3_U3, EXPECTED_R3_S3_U3_SIG_UL3,
    EXPECTED_R7_S4_U2, EXPECTED_R7_S4_U2_SIG_UL3,
    EXPECTED_R2_S5_U3, EXPECTED_R2_S5_U3_SIG_UL3,
    EXPECTED_R20_S1_U1, EXPECTED_R20_S1_U1_SIG_UL3,
    EXPECTED_R20_S20_U5, EXPECTED_R20_S20_U5_SIG_UL3,
    EXPECTED_R60_S30_U10, EXPECTED_R60_S30_U10_SIG_UL3,
)

TOLERANCE = 1e-10

# Signal-line EMA period for every expected signal array (Ergodic default).
UL = 3

# (r, s, u, expected_mdi, expected_signal)
COMBOS = [
    (20, 5, 3, EXPECTED_R20_S5_U3, EXPECTED_R20_S5_U3_SIG_UL3),
    (20, 5, 1, EXPECTED_R20_S5_U1, EXPECTED_R20_S5_U1_SIG_UL3),
    (1, 5, 3, EXPECTED_R1_S5_U3, EXPECTED_R1_S5_U3_SIG_UL3),
    (40, 5, 3, EXPECTED_R40_S5_U3, EXPECTED_R40_S5_U3_SIG_UL3),
    (10, 5, 3, EXPECTED_R10_S5_U3, EXPECTED_R10_S5_U3_SIG_UL3),
    (5, 5, 5, EXPECTED_R5_S5_U5, EXPECTED_R5_S5_U5_SIG_UL3),
    (20, 9, 1, EXPECTED_R20_S9_U1, EXPECTED_R20_S9_U1_SIG_UL3),
    (26, 12, 9, EXPECTED_R26_S12_U9, EXPECTED_R26_S12_U9_SIG_UL3),
    (50, 13, 1, EXPECTED_R50_S13_U1, EXPECTED_R50_S13_U1_SIG_UL3),
    (30, 5, 3, EXPECTED_R30_S5_U3, EXPECTED_R30_S5_U3_SIG_UL3),
    (3, 3, 3, EXPECTED_R3_S3_U3, EXPECTED_R3_S3_U3_SIG_UL3),
    (7, 4, 2, EXPECTED_R7_S4_U2, EXPECTED_R7_S4_U2_SIG_UL3),
    (2, 5, 3, EXPECTED_R2_S5_U3, EXPECTED_R2_S5_U3_SIG_UL3),
    (20, 1, 1, EXPECTED_R20_S1_U1, EXPECTED_R20_S1_U1_SIG_UL3),
    (20, 20, 5, EXPECTED_R20_S20_U5, EXPECTED_R20_S20_U5_SIG_UL3),
    (60, 30, 10, EXPECTED_R60_S30_U10, EXPECTED_R60_S30_U10_SIG_UL3),
]


class TestMeanDeviationIndexData(unittest.TestCase):
    """Test the MDI against the reference test data for all parameter combinations."""

    def test_all_combos(self):
        for r, s, u, exp_mdi, exp_signal in COMBOS:
            with self.subTest(r=r, s=s, u=u):
                ind = MeanDeviationIndex(MeanDeviationIndexParams(
                    r=r, s=s, u=u, ul=UL))
                for i in range(len(INPUT_CLOSE)):
                    mdi, signal = ind.update(INPUT_CLOSE[i])

                    self.assertAlmostEqual(mdi, exp_mdi[i], delta=TOLERANCE,
                                           msg=f"[{i}] mdi: expected {exp_mdi[i]}, got {mdi}")
                    self.assertAlmostEqual(signal, exp_signal[i], delta=TOLERANCE,
                                           msg=f"[{i}] signal: expected {exp_signal[i]}, got {signal}")


class TestMeanDeviationIndexNoWarmUp(unittest.TestCase):
    """Test that both outputs are finite from bar 0 and bar 0 is exactly 0.0."""

    def test_bar_zero_is_zero(self):
        ind = MeanDeviationIndex(MeanDeviationIndexParams())
        mdi, signal = ind.update(INPUT_CLOSE[0])
        self.assertEqual(mdi, 0.0)
        self.assertEqual(signal, 0.0)

    def test_no_nan_anywhere(self):
        ind = MeanDeviationIndex(MeanDeviationIndexParams())
        for i in range(len(INPUT_CLOSE)):
            mdi, signal = ind.update(INPUT_CLOSE[i])
            self.assertFalse(math.isnan(mdi), f"[{i}] mdi: unexpected NaN")
            self.assertFalse(math.isnan(signal), f"[{i}] signal: unexpected NaN")


class TestMeanDeviationIndexDegenerate(unittest.TestCase):
    """Test the r=1 invariant: the detrend is a passthrough, so the index is 0."""

    def test_r_one_is_identically_zero(self):
        ind = MeanDeviationIndex(MeanDeviationIndexParams(r=1, s=5, u=3, ul=3))
        for i in range(len(INPUT_CLOSE)):
            mdi, signal = ind.update(INPUT_CLOSE[i])
            self.assertEqual(mdi, 0.0, f"[{i}] mdi must be 0 when r=1")
            self.assertEqual(signal, 0.0, f"[{i}] signal must be 0 when r=1")


class TestMeanDeviationIndexPassthrough(unittest.TestCase):
    """Test the all-passthrough form MDI(2,1,1) with ul=1: mdi == price - EMA(price, 2)."""

    def test_passthrough(self):
        ind = MeanDeviationIndex(MeanDeviationIndexParams(r=2, s=1, u=1, ul=1))
        r0 = ind.update(10.0)  # bar 0 seeds the baseline -> deviation 0
        self.assertAlmostEqual(r0[0], 0.0, delta=TOLERANCE)
        self.assertAlmostEqual(r0[1], 0.0, delta=TOLERANCE)
        r1 = ind.update(13.0)  # EMA(2) = (2/3)*13 + (1/3)*10 = 12 -> 13-12 = 1
        self.assertAlmostEqual(r1[0], 1.0, delta=TOLERANCE)
        self.assertAlmostEqual(r1[1], 1.0, delta=TOLERANCE)


class TestMeanDeviationIndexPassthroughSignal(unittest.TestCase):
    """Test the ul=1 invariant: the signal line equals the index exactly."""

    def test_signal_equals_mdi(self):
        ind = MeanDeviationIndex(MeanDeviationIndexParams(r=20, s=5, u=3, ul=1))
        for i in range(len(INPUT_CLOSE)):
            mdi, signal = ind.update(INPUT_CLOSE[i])
            self.assertEqual(mdi, signal, f"[{i}] signal must equal mdi when ul=1")


class TestMeanDeviationIndexPrimed(unittest.TestCase):
    """Test priming: the MDI has no warm-up, so it is primed after the first update."""

    def test_is_primed(self):
        ind = MeanDeviationIndex(MeanDeviationIndexParams())
        self.assertFalse(ind.is_primed())
        ind.update(INPUT_CLOSE[0])
        self.assertTrue(ind.is_primed())


class TestMeanDeviationIndexMnemonic(unittest.TestCase):
    """Test mnemonic generation (ul is excluded from the mnemonic)."""

    def test_default_mnemonic(self):
        ind = MeanDeviationIndex(MeanDeviationIndexParams())
        self.assertEqual(ind.metadata().mnemonic, "mdi(20,5,3)")

    def test_custom_mnemonic(self):
        ind = MeanDeviationIndex(MeanDeviationIndexParams(r=26, s=12, u=9, ul=7))
        self.assertEqual(ind.metadata().mnemonic, "mdi(26,12,9)")


class TestMeanDeviationIndexMetadata(unittest.TestCase):
    """Test metadata generation."""

    def test_default_metadata(self):
        ind = MeanDeviationIndex(MeanDeviationIndexParams())
        meta = ind.metadata()
        self.assertEqual(meta.identifier, Identifier.MEAN_DEVIATION_INDEX)
        self.assertEqual(meta.mnemonic, "mdi(20,5,3)")
        self.assertEqual(len(meta.outputs), 2)


class TestMeanDeviationIndexUpdateScalar(unittest.TestCase):
    """Test update_scalar output ordering (mdi, signal)."""

    def test_update_scalar(self):
        ind = MeanDeviationIndex(MeanDeviationIndexParams(r=20, s=5, u=3, ul=UL))
        tm = datetime(2021, 4, 1)
        out = None
        for i in range(len(INPUT_CLOSE)):
            out = ind.update_scalar(Scalar(time=tm, value=INPUT_CLOSE[i]))
        self.assertEqual(len(out), 2)
        self.assertAlmostEqual(out[0].value, EXPECTED_R20_S5_U3[-1], delta=TOLERANCE)
        self.assertAlmostEqual(out[1].value, EXPECTED_R20_S5_U3_SIG_UL3[-1], delta=TOLERANCE)


class TestMeanDeviationIndexInvalidParams(unittest.TestCase):
    """Test invalid parameter validation."""

    def test_r_too_small(self):
        with self.assertRaises(ValueError):
            MeanDeviationIndex(MeanDeviationIndexParams(r=0))

    def test_s_too_small(self):
        with self.assertRaises(ValueError):
            MeanDeviationIndex(MeanDeviationIndexParams(s=0))

    def test_u_too_small(self):
        with self.assertRaises(ValueError):
            MeanDeviationIndex(MeanDeviationIndexParams(u=0))

    def test_ul_too_small(self):
        with self.assertRaises(ValueError):
            MeanDeviationIndex(MeanDeviationIndexParams(ul=0))


if __name__ == '__main__':
    unittest.main()
