import math
import unittest
from datetime import datetime

from py.indicators.william_blau.ergodic_oscillator.ergodic_oscillator import ErgodicOscillator
from py.indicators.william_blau.ergodic_oscillator.params import ErgodicOscillatorParams
from py.indicators.core.identifier import Identifier
from py.entities.scalar import Scalar

from .test_testdata import (
    INPUT_CLOSE,
    EXPECTED_ERG_Q2_R20_S5_U3_L3, EXPECTED_SIG_Q2_R20_S5_U3_L3,
    EXPECTED_ERG_Q2_R32_S5_U1_L5, EXPECTED_SIG_Q2_R32_S5_U1_L5,
    EXPECTED_ERG_Q2_R20_S5_U1_L5, EXPECTED_SIG_Q2_R20_S5_U1_L5,
    EXPECTED_ERG_Q2_R32_S5_U1_L7, EXPECTED_SIG_Q2_R32_S5_U1_L7,
    EXPECTED_ERG_Q2_R25_S13_U1_L5, EXPECTED_SIG_Q2_R25_S13_U1_L5,
    EXPECTED_ERG_Q2_R20_S5_U3_L1, EXPECTED_SIG_Q2_R20_S5_U3_L1,
    EXPECTED_ERG_Q2_R1_S1_U1_L1, EXPECTED_SIG_Q2_R1_S1_U1_L1,
    EXPECTED_ERG_Q2_R20_S5_U3_L9, EXPECTED_SIG_Q2_R20_S5_U3_L9,
    EXPECTED_ERG_Q2_R64_S64_U1_L5, EXPECTED_SIG_Q2_R64_S64_U1_L5,
    EXPECTED_ERG_Q2_R9_S3_U1_L3, EXPECTED_SIG_Q2_R9_S3_U1_L3,
    EXPECTED_ERG_Q3_R20_S5_U3_L3, EXPECTED_SIG_Q3_R20_S5_U3_L3,
    EXPECTED_ERG_Q5_R20_S5_U3_L5, EXPECTED_SIG_Q5_R20_S5_U3_L5,
    EXPECTED_ERG_Q2_R13_S7_U1_L7, EXPECTED_SIG_Q2_R13_S7_U1_L7,
    EXPECTED_ERG_Q2_R20_S5_U1_L3, EXPECTED_SIG_Q2_R20_S5_U1_L3,
)

TOLERANCE = 1e-10

# (q, r, s, u, ul, expected_ergodic, expected_signal)
COMBOS = [
    (2, 20, 5, 3, 3, EXPECTED_ERG_Q2_R20_S5_U3_L3, EXPECTED_SIG_Q2_R20_S5_U3_L3),
    (2, 32, 5, 1, 5, EXPECTED_ERG_Q2_R32_S5_U1_L5, EXPECTED_SIG_Q2_R32_S5_U1_L5),
    (2, 20, 5, 1, 5, EXPECTED_ERG_Q2_R20_S5_U1_L5, EXPECTED_SIG_Q2_R20_S5_U1_L5),
    (2, 32, 5, 1, 7, EXPECTED_ERG_Q2_R32_S5_U1_L7, EXPECTED_SIG_Q2_R32_S5_U1_L7),
    (2, 25, 13, 1, 5, EXPECTED_ERG_Q2_R25_S13_U1_L5, EXPECTED_SIG_Q2_R25_S13_U1_L5),
    (2, 20, 5, 3, 1, EXPECTED_ERG_Q2_R20_S5_U3_L1, EXPECTED_SIG_Q2_R20_S5_U3_L1),
    (2, 1, 1, 1, 1, EXPECTED_ERG_Q2_R1_S1_U1_L1, EXPECTED_SIG_Q2_R1_S1_U1_L1),
    (2, 20, 5, 3, 9, EXPECTED_ERG_Q2_R20_S5_U3_L9, EXPECTED_SIG_Q2_R20_S5_U3_L9),
    (2, 64, 64, 1, 5, EXPECTED_ERG_Q2_R64_S64_U1_L5, EXPECTED_SIG_Q2_R64_S64_U1_L5),
    (2, 9, 3, 1, 3, EXPECTED_ERG_Q2_R9_S3_U1_L3, EXPECTED_SIG_Q2_R9_S3_U1_L3),
    (3, 20, 5, 3, 3, EXPECTED_ERG_Q3_R20_S5_U3_L3, EXPECTED_SIG_Q3_R20_S5_U3_L3),
    (5, 20, 5, 3, 5, EXPECTED_ERG_Q5_R20_S5_U3_L5, EXPECTED_SIG_Q5_R20_S5_U3_L5),
    (2, 13, 7, 1, 7, EXPECTED_ERG_Q2_R13_S7_U1_L7, EXPECTED_SIG_Q2_R13_S7_U1_L7),
    (2, 20, 5, 1, 3, EXPECTED_ERG_Q2_R20_S5_U1_L3, EXPECTED_SIG_Q2_R20_S5_U1_L3),
]


class TestErgodicOscillatorData(unittest.TestCase):
    """Test the Ergodic against the reference test data for all parameter combinations."""

    def test_all_combos(self):
        for q, r, s, u, ul, exp_ergodic, exp_signal in COMBOS:
            with self.subTest(q=q, r=r, s=s, u=u, ul=ul):
                ind = ErgodicOscillator(ErgodicOscillatorParams(
                    q=q, r=r, s=s, u=u, ul=ul))
                for i in range(len(INPUT_CLOSE)):
                    ergodic, signal = ind.update(INPUT_CLOSE[i])

                    if math.isnan(exp_ergodic[i]):
                        self.assertTrue(math.isnan(ergodic), f"[{i}] ergodic: expected NaN, got {ergodic}")
                    else:
                        self.assertAlmostEqual(ergodic, exp_ergodic[i], delta=TOLERANCE,
                                               msg=f"[{i}] ergodic: expected {exp_ergodic[i]}, got {ergodic}")

                    if math.isnan(exp_signal[i]):
                        self.assertTrue(math.isnan(signal), f"[{i}] signal: expected NaN, got {signal}")
                    else:
                        self.assertAlmostEqual(signal, exp_signal[i], delta=TOLERANCE,
                                               msg=f"[{i}] signal: expected {exp_signal[i]}, got {signal}")


class TestErgodicOscillatorPassthroughSignal(unittest.TestCase):
    """Test the ul=1 invariant: the signal line equals the oscillator exactly."""

    def test_signal_equals_ergodic(self):
        ind = ErgodicOscillator(ErgodicOscillatorParams(q=2, r=20, s=5, u=3, ul=1))
        for i in range(len(INPUT_CLOSE)):
            ergodic, signal = ind.update(INPUT_CLOSE[i])
            if math.isnan(ergodic):
                self.assertTrue(math.isnan(signal), f"[{i}] signal: expected NaN, got {signal}")
            else:
                self.assertEqual(ergodic, signal, f"[{i}] signal must equal ergodic when ul=1")


class TestErgodicOscillatorPassthrough(unittest.TestCase):
    """Test the all-passthrough invariant Ergodic(2,1,1,1,1) = sign(mtm)*100."""

    def test_passthrough(self):
        ind = ErgodicOscillator(ErgodicOscillatorParams(q=2, r=1, s=1, u=1, ul=1))
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


class TestErgodicOscillatorPrimed(unittest.TestCase):
    """Test priming: not primed until the first finite oscillator value (bar q-1)."""

    def test_is_primed(self):
        ind = ErgodicOscillator(ErgodicOscillatorParams(q=3, r=20, s=5, u=3, ul=3))
        self.assertFalse(ind.is_primed())
        ind.update(INPUT_CLOSE[0])
        self.assertFalse(ind.is_primed())
        ind.update(INPUT_CLOSE[1])
        self.assertFalse(ind.is_primed())
        ind.update(INPUT_CLOSE[2])
        self.assertTrue(ind.is_primed())


class TestErgodicOscillatorMnemonic(unittest.TestCase):
    """Test mnemonic generation."""

    def test_default_mnemonic(self):
        ind = ErgodicOscillator(ErgodicOscillatorParams())
        self.assertEqual(ind.metadata().mnemonic, "ergodic(2,20,5,3,3)")

    def test_custom_mnemonic(self):
        ind = ErgodicOscillator(ErgodicOscillatorParams(q=2, r=25, s=13, u=1, ul=7))
        self.assertEqual(ind.metadata().mnemonic, "ergodic(2,25,13,1,7)")


class TestErgodicOscillatorMetadata(unittest.TestCase):
    """Test metadata generation."""

    def test_default_metadata(self):
        ind = ErgodicOscillator(ErgodicOscillatorParams())
        meta = ind.metadata()
        self.assertEqual(meta.identifier, Identifier.ERGODIC_OSCILLATOR)
        self.assertEqual(meta.mnemonic, "ergodic(2,20,5,3,3)")
        self.assertEqual(len(meta.outputs), 2)


class TestErgodicOscillatorUpdateScalar(unittest.TestCase):
    """Test update_scalar output ordering (ergodic, signal)."""

    def test_update_scalar(self):
        ind = ErgodicOscillator(ErgodicOscillatorParams(q=2, r=20, s=5, u=3, ul=3))
        tm = datetime(2021, 4, 1)
        out = None
        for i in range(len(INPUT_CLOSE)):
            out = ind.update_scalar(Scalar(time=tm, value=INPUT_CLOSE[i]))
        self.assertEqual(len(out), 2)
        self.assertAlmostEqual(out[0].value, EXPECTED_ERG_Q2_R20_S5_U3_L3[-1], delta=TOLERANCE)
        self.assertAlmostEqual(out[1].value, EXPECTED_SIG_Q2_R20_S5_U3_L3[-1], delta=TOLERANCE)


class TestErgodicOscillatorInvalidParams(unittest.TestCase):
    """Test invalid parameter validation."""

    def test_q_too_small(self):
        with self.assertRaises(ValueError):
            ErgodicOscillator(ErgodicOscillatorParams(q=0))

    def test_r_too_small(self):
        with self.assertRaises(ValueError):
            ErgodicOscillator(ErgodicOscillatorParams(r=0))

    def test_s_too_small(self):
        with self.assertRaises(ValueError):
            ErgodicOscillator(ErgodicOscillatorParams(s=0))

    def test_u_too_small(self):
        with self.assertRaises(ValueError):
            ErgodicOscillator(ErgodicOscillatorParams(u=0))

    def test_ul_too_small(self):
        with self.assertRaises(ValueError):
            ErgodicOscillator(ErgodicOscillatorParams(ul=0))


if __name__ == '__main__':
    unittest.main()
