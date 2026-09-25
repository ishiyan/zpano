import math
import unittest
from datetime import datetime

from py.indicators.william_blau.double_smoothed_stochastic.double_smoothed_stochastic import DoubleSmoothedStochastic
from py.indicators.william_blau.double_smoothed_stochastic.params import DoubleSmoothedStochasticParams
from py.indicators.core.identifier import Identifier
from py.entities.bar import Bar
from py.entities.quote import Quote
from py.entities.scalar import Scalar
from py.entities.trade import Trade

from .test_testdata import (
    INPUT_HIGH, INPUT_LOW, INPUT_CLOSE,
    EXPECTED_DS_Q5_R7_S3_G3, EXPECTED_SIG_Q5_R7_S3_G3,
    EXPECTED_DS_Q2_R3_S15_G3, EXPECTED_SIG_Q2_R3_S15_G3,
    EXPECTED_DS_Q5_R20_S5_G3, EXPECTED_SIG_Q5_R20_S5_G3,
    EXPECTED_DS_Q5_R7_S3_G1, EXPECTED_SIG_Q5_R7_S3_G1,
    EXPECTED_DS_Q2_R3_S15_G1, EXPECTED_SIG_Q2_R3_S15_G1,
    EXPECTED_DS_Q1_R1_S1_G1, EXPECTED_SIG_Q1_R1_S1_G1,
    EXPECTED_DS_Q1_R5_S5_G3, EXPECTED_SIG_Q1_R5_S5_G3,
    EXPECTED_DS_Q8_R5_S3_G3, EXPECTED_SIG_Q8_R5_S3_G3,
    EXPECTED_DS_Q21_R13_S4_G3, EXPECTED_SIG_Q21_R13_S4_G3,
    EXPECTED_DS_Q5_R1_S1_G3, EXPECTED_SIG_Q5_R1_S1_G3,
    EXPECTED_DS_Q3_R10_S10_G5, EXPECTED_SIG_Q3_R10_S10_G5,
    EXPECTED_DS_Q34_R5_S5_G3, EXPECTED_SIG_Q34_R5_S5_G3,
    EXPECTED_DS_Q2_R3_S15_G5, EXPECTED_SIG_Q2_R3_S15_G5,
    EXPECTED_DS_Q10_R7_S3_G3, EXPECTED_SIG_Q10_R7_S3_G3,
)

TOLERANCE = 1e-10

# (q, r, s, g, expected_dss, expected_signal)
COMBOS = [
    (5, 7, 3, 3, EXPECTED_DS_Q5_R7_S3_G3, EXPECTED_SIG_Q5_R7_S3_G3),
    (2, 3, 15, 3, EXPECTED_DS_Q2_R3_S15_G3, EXPECTED_SIG_Q2_R3_S15_G3),
    (5, 20, 5, 3, EXPECTED_DS_Q5_R20_S5_G3, EXPECTED_SIG_Q5_R20_S5_G3),
    (5, 7, 3, 1, EXPECTED_DS_Q5_R7_S3_G1, EXPECTED_SIG_Q5_R7_S3_G1),
    (2, 3, 15, 1, EXPECTED_DS_Q2_R3_S15_G1, EXPECTED_SIG_Q2_R3_S15_G1),
    (1, 1, 1, 1, EXPECTED_DS_Q1_R1_S1_G1, EXPECTED_SIG_Q1_R1_S1_G1),
    (1, 5, 5, 3, EXPECTED_DS_Q1_R5_S5_G3, EXPECTED_SIG_Q1_R5_S5_G3),
    (8, 5, 3, 3, EXPECTED_DS_Q8_R5_S3_G3, EXPECTED_SIG_Q8_R5_S3_G3),
    (21, 13, 4, 3, EXPECTED_DS_Q21_R13_S4_G3, EXPECTED_SIG_Q21_R13_S4_G3),
    (5, 1, 1, 3, EXPECTED_DS_Q5_R1_S1_G3, EXPECTED_SIG_Q5_R1_S1_G3),
    (3, 10, 10, 5, EXPECTED_DS_Q3_R10_S10_G5, EXPECTED_SIG_Q3_R10_S10_G5),
    (34, 5, 5, 3, EXPECTED_DS_Q34_R5_S5_G3, EXPECTED_SIG_Q34_R5_S5_G3),
    (2, 3, 15, 5, EXPECTED_DS_Q2_R3_S15_G5, EXPECTED_SIG_Q2_R3_S15_G5),
    (10, 7, 3, 3, EXPECTED_DS_Q10_R7_S3_G3, EXPECTED_SIG_Q10_R7_S3_G3),
]


class TestDoubleSmoothedStochasticData(unittest.TestCase):
    """Test DSS against the reference test data for all parameter combinations."""

    def test_all_combos(self):
        for q, r, s, g, exp_dss, exp_signal in COMBOS:
            with self.subTest(q=q, r=r, s=s, g=g):
                ind = DoubleSmoothedStochastic(DoubleSmoothedStochasticParams(
                    q=q, r=r, s=s, g=g))
                for i in range(len(INPUT_CLOSE)):
                    dss, signal = ind.update(INPUT_HIGH[i], INPUT_LOW[i], INPUT_CLOSE[i])

                    if math.isnan(exp_dss[i]):
                        self.assertTrue(math.isnan(dss), f"[{i}] dss: expected NaN, got {dss}")
                    else:
                        self.assertAlmostEqual(dss, exp_dss[i], delta=TOLERANCE,
                                               msg=f"[{i}] dss: expected {exp_dss[i]}, got {dss}")

                    if math.isnan(exp_signal[i]):
                        self.assertTrue(math.isnan(signal), f"[{i}] signal: expected NaN, got {signal}")
                    else:
                        self.assertAlmostEqual(signal, exp_signal[i], delta=TOLERANCE,
                                               msg=f"[{i}] signal: expected {exp_signal[i]}, got {signal}")


class TestDoubleSmoothedStochasticPassthrough(unittest.TestCase):
    """Test the one-bar HLC index DSS(1,1,1,1) = 100*(close-low)/(high-low)."""

    def test_passthrough(self):
        ind = DoubleSmoothedStochastic(DoubleSmoothedStochasticParams(q=1, r=1, s=1, g=1))
        r0 = ind.update(12.0, 10.0, 12.0)  # close at high -> 100
        self.assertAlmostEqual(r0[0], 100.0, delta=TOLERANCE)
        self.assertAlmostEqual(r0[1], 100.0, delta=TOLERANCE)
        r1 = ind.update(12.0, 10.0, 10.0)  # close at low -> 0
        self.assertAlmostEqual(r1[0], 0.0, delta=TOLERANCE)
        self.assertAlmostEqual(r1[1], 0.0, delta=TOLERANCE)
        r2 = ind.update(12.0, 10.0, 11.0)  # exact midpoint -> 50
        self.assertAlmostEqual(r2[0], 50.0, delta=TOLERANCE)
        self.assertAlmostEqual(r2[1], 50.0, delta=TOLERANCE)

    def test_division_guard(self):
        ind = DoubleSmoothedStochastic(DoubleSmoothedStochasticParams(q=1, r=1, s=1, g=1))
        dss, signal = ind.update(11.0, 11.0, 11.0)  # flat window -> 0.0
        self.assertEqual(dss, 0.0)
        self.assertEqual(signal, 0.0)

    def test_signal_expanding_window(self):
        # The signal SMA averages the oscillator values seen so far until g of
        # them exist, then rolls over the last g values.
        ind = DoubleSmoothedStochastic(DoubleSmoothedStochasticParams(q=1, r=1, s=1, g=3))
        _, signal = ind.update(12.0, 10.0, 12.0)  # dss 100
        self.assertAlmostEqual(signal, 100.0, delta=TOLERANCE)
        _, signal = ind.update(12.0, 10.0, 10.0)  # dss 0
        self.assertAlmostEqual(signal, 50.0, delta=TOLERANCE)
        _, signal = ind.update(12.0, 10.0, 11.0)  # dss 50
        self.assertAlmostEqual(signal, 50.0, delta=TOLERANCE)
        _, signal = ind.update(12.0, 10.0, 11.0)  # dss 50, 100 dropped
        self.assertAlmostEqual(signal, 100.0 / 3.0, delta=TOLERANCE)

    def test_signal_equals_dss_when_g1(self):
        ind = DoubleSmoothedStochastic(DoubleSmoothedStochasticParams(q=5, r=7, s=3, g=1))
        for i in range(len(INPUT_CLOSE)):
            dss, signal = ind.update(INPUT_HIGH[i], INPUT_LOW[i], INPUT_CLOSE[i])
            if math.isnan(dss):
                self.assertTrue(math.isnan(signal))
            else:
                self.assertEqual(dss, signal, f"[{i}]")


class TestDoubleSmoothedStochasticPrimed(unittest.TestCase):
    """The DSS is NaN and not primed for bars 0..q-2, primed from bar q-1."""

    def test_primed_default(self):
        q = 5
        ind = DoubleSmoothedStochastic(DoubleSmoothedStochasticParams(q=q))
        self.assertFalse(ind.is_primed())
        for i in range(q - 1):
            dss, signal = ind.update(INPUT_HIGH[i], INPUT_LOW[i], INPUT_CLOSE[i])
            self.assertFalse(ind.is_primed(), f"[{i}] primed too early")
            self.assertTrue(math.isnan(dss))
            self.assertTrue(math.isnan(signal))
        for i in range(q - 1, len(INPUT_CLOSE)):
            ind.update(INPUT_HIGH[i], INPUT_LOW[i], INPUT_CLOSE[i])
            self.assertTrue(ind.is_primed(), f"[{i}] not primed")

    def test_primed_q1(self):
        ind = DoubleSmoothedStochastic(DoubleSmoothedStochasticParams(q=1))
        self.assertFalse(ind.is_primed())
        ind.update(INPUT_HIGH[0], INPUT_LOW[0], INPUT_CLOSE[0])
        self.assertTrue(ind.is_primed())


class TestDoubleSmoothedStochasticMnemonic(unittest.TestCase):
    """Test mnemonic generation."""

    def test_default_mnemonic(self):
        ind = DoubleSmoothedStochastic(DoubleSmoothedStochasticParams())
        self.assertEqual(ind.metadata().mnemonic, "dss(5,7,3,3)")

    def test_custom_mnemonic(self):
        ind = DoubleSmoothedStochastic(DoubleSmoothedStochasticParams(q=2, r=3, s=15, g=5))
        self.assertEqual(ind.metadata().mnemonic, "dss(2,3,15,5)")


class TestDoubleSmoothedStochasticMetadata(unittest.TestCase):
    """Test metadata generation."""

    def test_default_metadata(self):
        ind = DoubleSmoothedStochastic(DoubleSmoothedStochasticParams())
        meta = ind.metadata()
        self.assertEqual(meta.identifier, Identifier.DOUBLE_SMOOTHED_STOCHASTIC)
        self.assertEqual(meta.mnemonic, "dss(5,7,3,3)")
        self.assertEqual(meta.description, "Double Smoothed Stochastic dss(5,7,3,3)")
        self.assertEqual(len(meta.outputs), 2)
        self.assertEqual(meta.outputs[0].mnemonic, "dss(5,7,3,3) dss")
        self.assertEqual(meta.outputs[0].description, "Double Smoothed Stochastic dss(5,7,3,3) DSS")
        self.assertEqual(meta.outputs[1].mnemonic, "dss(5,7,3,3) signal")
        self.assertEqual(meta.outputs[1].description, "Double Smoothed Stochastic dss(5,7,3,3) signal")


class TestDoubleSmoothedStochasticUpdateEntities(unittest.TestCase):
    """Test the entity update methods and their HLC mapping."""

    def test_update_bar(self):
        ind = DoubleSmoothedStochastic(DoubleSmoothedStochasticParams(q=5, r=7, s=3, g=3))
        tm = datetime(2021, 4, 1)
        out = None
        for i in range(len(INPUT_CLOSE)):
            out = ind.update_bar(Bar(time=tm, open=0.0, high=INPUT_HIGH[i],
                                     low=INPUT_LOW[i], close=INPUT_CLOSE[i], volume=0.0))
        self.assertEqual(len(out), 2)
        self.assertEqual(out[0].time, tm)
        self.assertAlmostEqual(out[0].value, EXPECTED_DS_Q5_R7_S3_G3[-1], delta=TOLERANCE)
        self.assertAlmostEqual(out[1].value, EXPECTED_SIG_Q5_R7_S3_G3[-1], delta=TOLERANCE)

    def test_update_scalar(self):
        # A scalar value is used as the high, the low and the close, so the
        # q-bar range spans the last q values.
        ind = DoubleSmoothedStochastic(DoubleSmoothedStochasticParams(q=2, r=1, s=1, g=1))
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
        ind = DoubleSmoothedStochastic(DoubleSmoothedStochasticParams(q=2, r=1, s=1, g=1))
        tm = datetime(2021, 4, 1)
        ind.update_quote(Quote(time=tm, bid_price=12.0, ask_price=14.0,
                               bid_size=1.0, ask_size=1.0))  # mid 13
        out = ind.update_quote(Quote(time=tm, bid_price=10.0, ask_price=12.0,
                                     bid_size=1.0, ask_size=1.0))  # mid 11, at the 2-bar low
        self.assertEqual(len(out), 2)
        self.assertAlmostEqual(out[0].value, 0.0, delta=TOLERANCE)
        self.assertAlmostEqual(out[1].value, 0.0, delta=TOLERANCE)

    def test_update_trade(self):
        # A trade price is used as the high, the low and the close.
        ind = DoubleSmoothedStochastic(DoubleSmoothedStochasticParams(q=2, r=1, s=1, g=1))
        tm = datetime(2021, 4, 1)
        ind.update_trade(Trade(time=tm, price=10.0, volume=1.0))
        out = ind.update_trade(Trade(time=tm, price=12.0, volume=1.0))
        self.assertEqual(len(out), 2)
        self.assertAlmostEqual(out[0].value, 100.0, delta=TOLERANCE)
        self.assertAlmostEqual(out[1].value, 100.0, delta=TOLERANCE)


class TestDoubleSmoothedStochasticInvalidParams(unittest.TestCase):
    """Test invalid parameter validation."""

    def test_q_too_small(self):
        with self.assertRaises(ValueError):
            DoubleSmoothedStochastic(DoubleSmoothedStochasticParams(q=0))

    def test_r_too_small(self):
        with self.assertRaises(ValueError):
            DoubleSmoothedStochastic(DoubleSmoothedStochasticParams(r=0))

    def test_s_too_small(self):
        with self.assertRaises(ValueError):
            DoubleSmoothedStochastic(DoubleSmoothedStochasticParams(s=0))

    def test_g_too_small(self):
        with self.assertRaises(ValueError):
            DoubleSmoothedStochastic(DoubleSmoothedStochasticParams(g=0))


if __name__ == '__main__':
    unittest.main()
