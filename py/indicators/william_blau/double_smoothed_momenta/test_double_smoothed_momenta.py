import math
import unittest
from datetime import datetime

from py.indicators.william_blau.double_smoothed_momenta.double_smoothed_momenta import DoubleSmoothedMomenta
from py.indicators.william_blau.double_smoothed_momenta.params import DoubleSmoothedMomentaParams
from py.indicators.core.identifier import Identifier
from py.entities.bar_component import BarComponent
from py.entities.quote_component import QuoteComponent
from py.entities.trade_component import TradeComponent
from py.entities.scalar import Scalar

from .test_testdata import (
    INPUT_CLOSE,
    EXPECTED_A2_Y2_Z14,
    EXPECTED_A2_Y1_Z14,
    EXPECTED_A2_Y1_Z9,
    EXPECTED_A2_Y1_Z2,
    EXPECTED_A2_Y3_Z9,
    EXPECTED_A2_Y5_Z5,
    EXPECTED_A2_Y2_Z5,
    EXPECTED_A2_Y1_Z1,
    EXPECTED_A1_Y1_Z1,
    EXPECTED_A5_Y2_Z14,
    EXPECTED_A10_Y3_Z5,
    EXPECTED_A14_Y2_Z9,
    EXPECTED_A20_Y5_Z3,
    EXPECTED_A3_Y3_Z3,
    EXPECTED_A7_Y4_Z2,
    EXPECTED_A32_Y2_Z7,
)

TOLERANCE = 1e-10

# (a, y, z, expected)
COMBOS = [
    (2, 2, 14, EXPECTED_A2_Y2_Z14),
    (2, 1, 14, EXPECTED_A2_Y1_Z14),
    (2, 1, 9, EXPECTED_A2_Y1_Z9),
    (2, 1, 2, EXPECTED_A2_Y1_Z2),
    (2, 3, 9, EXPECTED_A2_Y3_Z9),
    (2, 5, 5, EXPECTED_A2_Y5_Z5),
    (2, 2, 5, EXPECTED_A2_Y2_Z5),
    (2, 1, 1, EXPECTED_A2_Y1_Z1),
    (1, 1, 1, EXPECTED_A1_Y1_Z1),
    (5, 2, 14, EXPECTED_A5_Y2_Z14),
    (10, 3, 5, EXPECTED_A10_Y3_Z5),
    (14, 2, 9, EXPECTED_A14_Y2_Z9),
    (20, 5, 3, EXPECTED_A20_Y5_Z3),
    (3, 3, 3, EXPECTED_A3_Y3_Z3),
    (7, 4, 2, EXPECTED_A7_Y4_Z2),
    (32, 2, 7, EXPECTED_A32_Y2_Z7),
]


class TestDoubleSmoothedMomentaData(unittest.TestCase):
    """Test the DM against the reference test data for all parameter combinations."""

    def test_all_combos(self):
        for a, y, z, expected in COMBOS:
            with self.subTest(a=a, y=y, z=z):
                ind = DoubleSmoothedMomenta(DoubleSmoothedMomentaParams(a=a, y=y, z=z))
                for i in range(len(INPUT_CLOSE)):
                    value = ind.update(INPUT_CLOSE[i])

                    if math.isnan(expected[i]):
                        self.assertTrue(math.isnan(value), f"[{i}] expected NaN, got {value}")
                    else:
                        self.assertAlmostEqual(value, expected[i], delta=TOLERANCE,
                                               msg=f"[{i}] expected {expected[i]}, got {value}")


class TestDoubleSmoothedMomentaWarmUp(unittest.TestCase):
    """Test the a-bar warm-up region: NaN for bars 0..a-2, finite from bar a-1."""

    def test_warm_up_region(self):
        a = 5
        ind = DoubleSmoothedMomenta(DoubleSmoothedMomentaParams(a=a, y=2, z=14))
        for i in range(a - 1):
            self.assertTrue(math.isnan(ind.update(INPUT_CLOSE[i])), f"[{i}] expected NaN")

        for i in range(a - 1, len(INPUT_CLOSE)):
            self.assertFalse(math.isnan(ind.update(INPUT_CLOSE[i])), f"[{i}] unexpected NaN")

    def test_no_warm_up_when_a_is_one(self):
        ind = DoubleSmoothedMomenta(DoubleSmoothedMomentaParams(a=1, y=1, z=1))
        for i in range(len(INPUT_CLOSE)):
            self.assertFalse(math.isnan(ind.update(INPUT_CLOSE[i])), f"[{i}] unexpected NaN")


class TestDoubleSmoothedMomentaBounds(unittest.TestCase):
    """Test that every finite value is bounded to [0, 100]."""

    def test_bounded(self):
        ind = DoubleSmoothedMomenta(DoubleSmoothedMomentaParams())
        for i in range(len(INPUT_CLOSE)):
            value = ind.update(INPUT_CLOSE[i])
            if not math.isnan(value):
                self.assertGreaterEqual(value, 0.0, f"[{i}] below 0")
                self.assertLessEqual(value, 100.0, f"[{i}] above 100")


class TestDoubleSmoothedMomentaDegenerate(unittest.TestCase):
    """Test the a=1 invariant: the 1-bar close range is 0, so the guard yields 0."""

    def test_a_one_is_identically_zero(self):
        ind = DoubleSmoothedMomenta(DoubleSmoothedMomentaParams(a=1, y=1, z=1))
        for i in range(len(INPUT_CLOSE)):
            self.assertEqual(ind.update(INPUT_CLOSE[i]), 0.0, f"[{i}] must be 0 when a=1")


class TestDoubleSmoothedMomentaRsiEquivalence(unittest.TestCase):
    """Test the DM(2,1,z) == EMA-form RSI(z) invariant."""

    @staticmethod
    def _ema_form_rsi(closes, z):
        """Independently-coded EMA-form RSI: 100 * EMA(up, z) / EMA(up + dn, z)."""
        alpha = 2.0 / (float(z) + 1.0)
        numerator = 0.0
        denominator = 0.0
        primed = False
        result = [math.nan]

        for k in range(1, len(closes)):
            diff = closes[k] - closes[k - 1]
            up = diff if diff > 0.0 else 0.0
            dn = -diff if diff < 0.0 else 0.0

            if primed:
                numerator = alpha * up + (1.0 - alpha) * numerator
                denominator = alpha * (up + dn) + (1.0 - alpha) * denominator
            else:
                numerator = up
                denominator = up + dn
                primed = True

            result.append(0.0 if denominator <= 0.0 else 100.0 * numerator / denominator)

        return result

    def test_rsi_equivalence(self):
        for z in (1, 2, 9, 14):
            with self.subTest(z=z):
                expected = self._ema_form_rsi(INPUT_CLOSE, z)
                ind = DoubleSmoothedMomenta(DoubleSmoothedMomentaParams(a=2, y=1, z=z))
                for i in range(len(INPUT_CLOSE)):
                    value = ind.update(INPUT_CLOSE[i])

                    if math.isnan(expected[i]):
                        self.assertTrue(math.isnan(value), f"[{i}] expected NaN")
                    else:
                        self.assertAlmostEqual(value, expected[i], delta=TOLERANCE,
                                               msg=f"[{i}] expected {expected[i]}, got {value}")


class TestDoubleSmoothedMomentaPrimed(unittest.TestCase):
    """Test priming: the DM is primed once a closes have been seen."""

    def test_is_primed(self):
        a = 5
        ind = DoubleSmoothedMomenta(DoubleSmoothedMomentaParams(a=a, y=2, z=14))
        for i in range(a - 1):
            ind.update(INPUT_CLOSE[i])
            self.assertFalse(ind.is_primed(), f"[{i}] must not be primed")

        for i in range(a - 1, len(INPUT_CLOSE)):
            ind.update(INPUT_CLOSE[i])
            self.assertTrue(ind.is_primed(), f"[{i}] must be primed")


class TestDoubleSmoothedMomentaMnemonic(unittest.TestCase):
    """Test mnemonic generation for all component combinations."""

    def test_all_components_default(self):
        ind = DoubleSmoothedMomenta(DoubleSmoothedMomentaParams())
        self.assertEqual(ind.metadata().mnemonic, "dm(2,2,14)")
        self.assertEqual(ind.metadata().description, "Double-Smoothed Momenta dm(2,2,14)")

    def test_custom_periods(self):
        ind = DoubleSmoothedMomenta(DoubleSmoothedMomentaParams(a=10, y=3, z=5))
        self.assertEqual(ind.metadata().mnemonic, "dm(10,3,5)")

    def test_only_bar_component_set(self):
        ind = DoubleSmoothedMomenta(DoubleSmoothedMomentaParams(bar_component=BarComponent.MEDIAN))
        self.assertEqual(ind.metadata().mnemonic, "dm(2,2,14, hl/2)")

    def test_only_quote_component_set(self):
        ind = DoubleSmoothedMomenta(DoubleSmoothedMomentaParams(quote_component=QuoteComponent.BID))
        self.assertEqual(ind.metadata().mnemonic, "dm(2,2,14, b)")

    def test_only_trade_component_set(self):
        ind = DoubleSmoothedMomenta(DoubleSmoothedMomentaParams(trade_component=TradeComponent.VOLUME))
        self.assertEqual(ind.metadata().mnemonic, "dm(2,2,14, v)")

    def test_bar_and_quote_components_set(self):
        ind = DoubleSmoothedMomenta(DoubleSmoothedMomentaParams(
            bar_component=BarComponent.OPEN, quote_component=QuoteComponent.BID))
        self.assertEqual(ind.metadata().mnemonic, "dm(2,2,14, o, b)")

    def test_bar_and_trade_components_set(self):
        ind = DoubleSmoothedMomenta(DoubleSmoothedMomentaParams(
            bar_component=BarComponent.HIGH, trade_component=TradeComponent.VOLUME))
        self.assertEqual(ind.metadata().mnemonic, "dm(2,2,14, h, v)")

    def test_quote_and_trade_components_set(self):
        ind = DoubleSmoothedMomenta(DoubleSmoothedMomentaParams(
            quote_component=QuoteComponent.ASK, trade_component=TradeComponent.VOLUME))
        self.assertEqual(ind.metadata().mnemonic, "dm(2,2,14, a, v)")


class TestDoubleSmoothedMomentaMetadata(unittest.TestCase):
    """Test metadata generation."""

    def test_default_metadata(self):
        ind = DoubleSmoothedMomenta(DoubleSmoothedMomentaParams())
        meta = ind.metadata()
        self.assertEqual(meta.identifier, Identifier.DOUBLE_SMOOTHED_MOMENTA)
        self.assertEqual(meta.mnemonic, "dm(2,2,14)")
        self.assertEqual(len(meta.outputs), 1)


class TestDoubleSmoothedMomentaUpdateScalar(unittest.TestCase):
    """Test the update_scalar output."""

    def test_update_scalar(self):
        ind = DoubleSmoothedMomenta(DoubleSmoothedMomentaParams())
        tm = datetime(2021, 4, 1)
        out = None
        for i in range(len(INPUT_CLOSE)):
            out = ind.update_scalar(Scalar(time=tm, value=INPUT_CLOSE[i]))
        self.assertEqual(len(out), 1)
        self.assertAlmostEqual(out[0].value, EXPECTED_A2_Y2_Z14[-1], delta=TOLERANCE)


class TestDoubleSmoothedMomentaInvalidParams(unittest.TestCase):
    """Test invalid parameter validation."""

    def test_a_too_small(self):
        with self.assertRaises(ValueError):
            DoubleSmoothedMomenta(DoubleSmoothedMomentaParams(a=0))

    def test_y_too_small(self):
        with self.assertRaises(ValueError):
            DoubleSmoothedMomenta(DoubleSmoothedMomentaParams(y=0))

    def test_z_too_small(self):
        with self.assertRaises(ValueError):
            DoubleSmoothedMomenta(DoubleSmoothedMomentaParams(z=0))


if __name__ == '__main__':
    unittest.main()
