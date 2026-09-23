import math
import unittest
from datetime import datetime

from py.indicators.william_blau.stochastic_momentum_index.stochastic_momentum_index import StochasticMomentumIndex
from py.indicators.william_blau.stochastic_momentum_index.params import StochasticMomentumIndexParams
from py.indicators.core.identifier import Identifier
from py.entities.bar import Bar
from py.entities.quote import Quote
from py.entities.scalar import Scalar
from py.entities.trade import Trade

from .test_testdata import (
    INPUT_HIGH, INPUT_LOW, INPUT_CLOSE,
    EXPECTED_Q5_R20_S5_U3, EXPECTED_Q5_R20_S5_U3_SIG_UL3,
    EXPECTED_Q13_R25_S2_U1, EXPECTED_Q13_R25_S2_U1_SIG_UL3,
    EXPECTED_Q2_R20_S20_U1, EXPECTED_Q2_R20_S20_U1_SIG_UL3,
    EXPECTED_Q13_R25_S2_U3, EXPECTED_Q13_R25_S2_U3_SIG_UL3,
    EXPECTED_Q5_R20_S5_U1, EXPECTED_Q5_R20_S5_U1_SIG_UL3,
    EXPECTED_Q8_R5_S3_U1, EXPECTED_Q8_R5_S3_U1_SIG_UL3,
    EXPECTED_Q21_R13_S4_U1, EXPECTED_Q21_R13_S4_U1_SIG_UL3,
    EXPECTED_Q1_R20_S5_U3, EXPECTED_Q1_R20_S5_U3_SIG_UL3,
    EXPECTED_Q1_R40_S20_U1, EXPECTED_Q1_R40_S20_U1_SIG_UL3,
    EXPECTED_Q1_R100_S20_U1, EXPECTED_Q1_R100_S20_U1_SIG_UL3,
    EXPECTED_Q1_R1_S1_U1, EXPECTED_Q1_R1_S1_U1_SIG_UL3,
    EXPECTED_Q5_R1_S1_U1, EXPECTED_Q5_R1_S1_U1_SIG_UL3,
    EXPECTED_Q3_R10_S10_U1, EXPECTED_Q3_R10_S10_U1_SIG_UL3,
    EXPECTED_Q34_R5_S5_U1, EXPECTED_Q34_R5_S5_U1_SIG_UL3,
    EXPECTED_Q2_R2_S2_U2, EXPECTED_Q2_R2_S2_U2_SIG_UL3,
    EXPECTED_Q50_R20_S5_U3, EXPECTED_Q50_R20_S5_U3_SIG_UL3,
)

TOLERANCE = 1e-10

# Signal-line EMA period for every expected signal array (Ergodic default).
UL = 3

# (q, r, s, u, expected_smi, expected_signal)
COMBOS = [
    (5, 20, 5, 3, EXPECTED_Q5_R20_S5_U3, EXPECTED_Q5_R20_S5_U3_SIG_UL3),
    (13, 25, 2, 1, EXPECTED_Q13_R25_S2_U1, EXPECTED_Q13_R25_S2_U1_SIG_UL3),
    (2, 20, 20, 1, EXPECTED_Q2_R20_S20_U1, EXPECTED_Q2_R20_S20_U1_SIG_UL3),
    (13, 25, 2, 3, EXPECTED_Q13_R25_S2_U3, EXPECTED_Q13_R25_S2_U3_SIG_UL3),
    (5, 20, 5, 1, EXPECTED_Q5_R20_S5_U1, EXPECTED_Q5_R20_S5_U1_SIG_UL3),
    (8, 5, 3, 1, EXPECTED_Q8_R5_S3_U1, EXPECTED_Q8_R5_S3_U1_SIG_UL3),
    (21, 13, 4, 1, EXPECTED_Q21_R13_S4_U1, EXPECTED_Q21_R13_S4_U1_SIG_UL3),
    (1, 20, 5, 3, EXPECTED_Q1_R20_S5_U3, EXPECTED_Q1_R20_S5_U3_SIG_UL3),
    (1, 40, 20, 1, EXPECTED_Q1_R40_S20_U1, EXPECTED_Q1_R40_S20_U1_SIG_UL3),
    (1, 100, 20, 1, EXPECTED_Q1_R100_S20_U1, EXPECTED_Q1_R100_S20_U1_SIG_UL3),
    (1, 1, 1, 1, EXPECTED_Q1_R1_S1_U1, EXPECTED_Q1_R1_S1_U1_SIG_UL3),
    (5, 1, 1, 1, EXPECTED_Q5_R1_S1_U1, EXPECTED_Q5_R1_S1_U1_SIG_UL3),
    (3, 10, 10, 1, EXPECTED_Q3_R10_S10_U1, EXPECTED_Q3_R10_S10_U1_SIG_UL3),
    (34, 5, 5, 1, EXPECTED_Q34_R5_S5_U1, EXPECTED_Q34_R5_S5_U1_SIG_UL3),
    (2, 2, 2, 2, EXPECTED_Q2_R2_S2_U2, EXPECTED_Q2_R2_S2_U2_SIG_UL3),
    (50, 20, 5, 3, EXPECTED_Q50_R20_S5_U3, EXPECTED_Q50_R20_S5_U3_SIG_UL3),
]


class TestStochasticMomentumIndexData(unittest.TestCase):
    """Test SMI against the reference test data for all parameter combinations."""

    def test_all_combos(self):
        for q, r, s, u, exp_smi, exp_signal in COMBOS:
            with self.subTest(q=q, r=r, s=s, u=u):
                ind = StochasticMomentumIndex(StochasticMomentumIndexParams(
                    q=q, r=r, s=s, u=u, ul=UL))
                for i in range(len(INPUT_CLOSE)):
                    smi, signal = ind.update(INPUT_HIGH[i], INPUT_LOW[i], INPUT_CLOSE[i])

                    if math.isnan(exp_smi[i]):
                        self.assertTrue(math.isnan(smi), f"[{i}] smi: expected NaN, got {smi}")
                    else:
                        self.assertAlmostEqual(smi, exp_smi[i], delta=TOLERANCE,
                                               msg=f"[{i}] smi: expected {exp_smi[i]}, got {smi}")

                    if math.isnan(exp_signal[i]):
                        self.assertTrue(math.isnan(signal), f"[{i}] signal: expected NaN, got {signal}")
                    else:
                        self.assertAlmostEqual(signal, exp_signal[i], delta=TOLERANCE,
                                               msg=f"[{i}] signal: expected {exp_signal[i]}, got {signal}")


class TestStochasticMomentumIndexPassthrough(unittest.TestCase):
    """Test the one-day raw stochastic SMI(1,1,1,1) = 100*(close-mid)/half-range."""

    def test_passthrough(self):
        ind = StochasticMomentumIndex(StochasticMomentumIndexParams(q=1, r=1, s=1, u=1, ul=1))
        r0 = ind.update(12.0, 10.0, 12.0)  # close at high -> +100
        self.assertAlmostEqual(r0[0], 100.0, delta=TOLERANCE)
        self.assertAlmostEqual(r0[1], 100.0, delta=TOLERANCE)
        r1 = ind.update(12.0, 10.0, 10.0)  # close at low -> -100
        self.assertAlmostEqual(r1[0], -100.0, delta=TOLERANCE)
        self.assertAlmostEqual(r1[1], -100.0, delta=TOLERANCE)
        r2 = ind.update(12.0, 10.0, 11.0)  # exact midpoint -> 0
        self.assertAlmostEqual(r2[0], 0.0, delta=TOLERANCE)
        self.assertAlmostEqual(r2[1], 0.0, delta=TOLERANCE)

    def test_division_guard(self):
        ind = StochasticMomentumIndex(StochasticMomentumIndexParams(q=1, r=1, s=1, u=1, ul=1))
        smi, signal = ind.update(11.0, 11.0, 11.0)  # flat window -> 0.0
        self.assertEqual(smi, 0.0)
        self.assertEqual(signal, 0.0)


class TestStochasticMomentumIndexPrimed(unittest.TestCase):
    """The SMI is NaN and not primed for bars 0..q-2, primed from bar q-1."""

    def test_primed_default(self):
        q = 5
        ind = StochasticMomentumIndex(StochasticMomentumIndexParams(q=q))
        self.assertFalse(ind.is_primed())
        for i in range(q - 1):
            smi, signal = ind.update(INPUT_HIGH[i], INPUT_LOW[i], INPUT_CLOSE[i])
            self.assertFalse(ind.is_primed(), f"[{i}] primed too early")
            self.assertTrue(math.isnan(smi))
            self.assertTrue(math.isnan(signal))
        for i in range(q - 1, len(INPUT_CLOSE)):
            ind.update(INPUT_HIGH[i], INPUT_LOW[i], INPUT_CLOSE[i])
            self.assertTrue(ind.is_primed(), f"[{i}] not primed")

    def test_primed_q1(self):
        ind = StochasticMomentumIndex(StochasticMomentumIndexParams(q=1))
        self.assertFalse(ind.is_primed())
        ind.update(INPUT_HIGH[0], INPUT_LOW[0], INPUT_CLOSE[0])
        self.assertTrue(ind.is_primed())


class TestStochasticMomentumIndexMnemonic(unittest.TestCase):
    """Test mnemonic generation."""

    def test_default_mnemonic(self):
        ind = StochasticMomentumIndex(StochasticMomentumIndexParams())
        self.assertEqual(ind.metadata().mnemonic, "smi(5,20,5,3,3)")

    def test_custom_mnemonic(self):
        ind = StochasticMomentumIndex(StochasticMomentumIndexParams(q=13, r=25, s=2, u=1, ul=7))
        self.assertEqual(ind.metadata().mnemonic, "smi(13,25,2,1,7)")


class TestStochasticMomentumIndexMetadata(unittest.TestCase):
    """Test metadata generation."""

    def test_default_metadata(self):
        ind = StochasticMomentumIndex(StochasticMomentumIndexParams())
        meta = ind.metadata()
        self.assertEqual(meta.identifier, Identifier.STOCHASTIC_MOMENTUM_INDEX)
        self.assertEqual(meta.mnemonic, "smi(5,20,5,3,3)")
        self.assertEqual(meta.description, "Stochastic Momentum Index smi(5,20,5,3,3)")
        self.assertEqual(len(meta.outputs), 2)
        self.assertEqual(meta.outputs[0].mnemonic, "smi(5,20,5,3,3) smi")
        self.assertEqual(meta.outputs[0].description, "Stochastic Momentum Index smi(5,20,5,3,3) SMI")
        self.assertEqual(meta.outputs[1].mnemonic, "smi(5,20,5,3,3) signal")
        self.assertEqual(meta.outputs[1].description, "Stochastic Momentum Index smi(5,20,5,3,3) signal")


class TestStochasticMomentumIndexUpdateEntities(unittest.TestCase):
    """Test the entity update methods and their HLC mapping."""

    def test_update_bar(self):
        ind = StochasticMomentumIndex(StochasticMomentumIndexParams(q=5, r=20, s=5, u=3, ul=UL))
        tm = datetime(2021, 4, 1)
        out = None
        for i in range(len(INPUT_CLOSE)):
            out = ind.update_bar(Bar(time=tm, open=0.0, high=INPUT_HIGH[i],
                                     low=INPUT_LOW[i], close=INPUT_CLOSE[i], volume=0.0))
        self.assertEqual(len(out), 2)
        self.assertEqual(out[0].time, tm)
        self.assertAlmostEqual(out[0].value, EXPECTED_Q5_R20_S5_U3[-1], delta=TOLERANCE)
        self.assertAlmostEqual(out[1].value, EXPECTED_Q5_R20_S5_U3_SIG_UL3[-1], delta=TOLERANCE)

    def test_update_scalar(self):
        # A scalar value is used as the high, the low and the close, so the
        # q-bar range spans the last q values.
        ind = StochasticMomentumIndex(StochasticMomentumIndexParams(q=2, r=1, s=1, u=1, ul=1))
        tm = datetime(2021, 4, 1)
        out = ind.update_scalar(Scalar(time=tm, value=10.0))
        self.assertEqual(len(out), 2)
        self.assertTrue(math.isnan(out[0].value))
        self.assertTrue(math.isnan(out[1].value))
        out = ind.update_scalar(Scalar(time=tm, value=12.0))  # close at the 2-bar high
        self.assertAlmostEqual(out[0].value, 100.0, delta=TOLERANCE)
        self.assertAlmostEqual(out[1].value, 100.0, delta=TOLERANCE)

    def test_update_quote(self):
        # A quote maps the mid price to the high, the low and the close.
        ind = StochasticMomentumIndex(StochasticMomentumIndexParams(q=2, r=1, s=1, u=1, ul=1))
        tm = datetime(2021, 4, 1)
        ind.update_quote(Quote(time=tm, bid_price=12.0, ask_price=14.0,
                               bid_size=1.0, ask_size=1.0))  # mid 13
        out = ind.update_quote(Quote(time=tm, bid_price=10.0, ask_price=12.0,
                                     bid_size=1.0, ask_size=1.0))  # mid 11, at the 2-bar low
        self.assertEqual(len(out), 2)
        self.assertAlmostEqual(out[0].value, -100.0, delta=TOLERANCE)
        self.assertAlmostEqual(out[1].value, -100.0, delta=TOLERANCE)

    def test_update_trade(self):
        # A trade price is used as the high, the low and the close.
        ind = StochasticMomentumIndex(StochasticMomentumIndexParams(q=2, r=1, s=1, u=1, ul=1))
        tm = datetime(2021, 4, 1)
        ind.update_trade(Trade(time=tm, price=10.0, volume=1.0))
        out = ind.update_trade(Trade(time=tm, price=12.0, volume=1.0))
        self.assertEqual(len(out), 2)
        self.assertAlmostEqual(out[0].value, 100.0, delta=TOLERANCE)
        self.assertAlmostEqual(out[1].value, 100.0, delta=TOLERANCE)


class TestStochasticMomentumIndexInvalidParams(unittest.TestCase):
    """Test invalid parameter validation."""

    def test_q_too_small(self):
        with self.assertRaises(ValueError):
            StochasticMomentumIndex(StochasticMomentumIndexParams(q=0))

    def test_r_too_small(self):
        with self.assertRaises(ValueError):
            StochasticMomentumIndex(StochasticMomentumIndexParams(r=0))

    def test_s_too_small(self):
        with self.assertRaises(ValueError):
            StochasticMomentumIndex(StochasticMomentumIndexParams(s=0))

    def test_u_too_small(self):
        with self.assertRaises(ValueError):
            StochasticMomentumIndex(StochasticMomentumIndexParams(u=0))

    def test_ul_too_small(self):
        with self.assertRaises(ValueError):
            StochasticMomentumIndex(StochasticMomentumIndexParams(ul=0))


if __name__ == '__main__':
    unittest.main()
