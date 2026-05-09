import math
import unittest

from py.indicators.welles_wilder.relative_strength_index.relative_strength_index import RelativeStrengthIndex
from py.indicators.welles_wilder.relative_strength_index.params import RelativeStrengthIndexParams
from py.indicators.core.identifier import Identifier
from py.entities.bar import Bar
from py.entities.quote import Quote
from py.entities.trade import Trade
from py.entities.scalar import Scalar
from py.entities.bar_component import BarComponent

from .test_testdata import TEST_INPUT_1, TEST_EXPECTED_1, TEST_TIME


# Test data from TA-Lib reference (length=9, 25 entries).
def _create() -> RelativeStrengthIndex:
    return RelativeStrengthIndex(RelativeStrengthIndexParams(length=9))


class TestRelativeStrengthIndex(unittest.TestCase):
    """Tests for the Relative Strength Index indicator."""

    def test_update(self):
        """Test RSI update with TA-Lib reference data (length=9)."""
        rsi = _create()
        for i, inp in enumerate(TEST_INPUT_1):
            act = rsi.update(inp)
            if i < 9:
                self.assertTrue(math.isnan(act), f"[{i}] expected NaN, got {act}")
            else:
                self.assertAlmostEqual(act, TEST_EXPECTED_1[i], delta=1e-9,
                                       msg=f"[{i}] expected {TEST_EXPECTED_1[i]}, got {act}")

    def test_nan_passthrough(self):
        """NaN input returns NaN."""
        rsi = _create()
        self.assertTrue(math.isnan(rsi.update(math.nan)))

    def test_is_primed(self):
        """Test priming with length=5."""
        rsi = RelativeStrengthIndex(RelativeStrengthIndexParams(length=5))
        self.assertFalse(rsi.is_primed())
        # Feed values 1..5 (5 updates): should NOT be primed.
        for i in range(1, 6):
            rsi.update(float(i))
            self.assertFalse(rsi.is_primed(), f"[{i}] should not be primed")
        # 6th update: should be primed.
        rsi.update(6.0)
        self.assertTrue(rsi.is_primed(), "[6] should be primed")
        # Further updates remain primed.
        for i in range(7, 12):
            rsi.update(float(i))
            self.assertTrue(rsi.is_primed(), f"[{i}] should be primed")

    def test_update_entity(self):
        """Test entity update methods."""
        inp = 100.0
        rsi = _create()
        # Prime: need length+1 = 10 updates.
        for _ in range(10):
            rsi.update(inp)

        def check(output):
            self.assertEqual(len(output), 1)
            s = output[0]
            self.assertIsInstance(s, Scalar)
            self.assertEqual(s.time, TEST_TIME)
            self.assertFalse(math.isnan(s.value))

        # Scalar
        rsi2 = _create()
        for _ in range(10):
            rsi2.update(inp)
        check(rsi2.update_scalar(Scalar(time=TEST_TIME, value=inp)))

        # Bar
        rsi3 = _create()
        for _ in range(10):
            rsi3.update(inp)
        check(rsi3.update_bar(Bar(time=TEST_TIME, open=0, high=0, low=0, close=inp, volume=0)))

        # Quote
        rsi4 = _create()
        for _ in range(10):
            rsi4.update(inp)
        check(rsi4.update_quote(Quote(time=TEST_TIME, bid_price=inp, ask_price=inp, bid_size=0, ask_size=0)))

        # Trade
        rsi5 = _create()
        for _ in range(10):
            rsi5.update(inp)
        check(rsi5.update_trade(Trade(time=TEST_TIME, price=inp, volume=0)))

    def test_metadata(self):
        """Test metadata."""
        rsi = _create()
        meta = rsi.metadata()
        self.assertEqual(meta.identifier, Identifier.RELATIVE_STRENGTH_INDEX)
        self.assertEqual(len(meta.outputs), 1)
        self.assertEqual(meta.outputs[0].mnemonic, "rsi(9)")
        self.assertEqual(meta.outputs[0].description, "Relative Strength Index rsi(9)")

    def test_invalid_length(self):
        """Length < 2 raises ValueError."""
        with self.assertRaises(ValueError):
            RelativeStrengthIndex(RelativeStrengthIndexParams(length=1))

    def test_non_default_bar_component_mnemonic(self):
        """Non-default bar component appears in mnemonic."""
        rsi = RelativeStrengthIndex(RelativeStrengthIndexParams(
            length=14, bar_component=BarComponent.OPEN))
        self.assertEqual(rsi.mnemonic, "rsi(14, o)")

    def test_update2_252(self):
        """Test RSI with length=14 over 252-element repeating data."""
        input2 = [
            44.34, 44.09, 43.61, 44.33, 44.83, 45.10, 45.42, 45.84, 46.08, 45.89,
            46.03, 45.61, 46.28, 46.28, 46.00, 46.03, 46.41, 46.22, 45.64, 46.21,
            46.25, 45.71, 46.45, 45.78, 45.35, 44.03, 44.18, 44.22, 44.57, 43.42,
            42.66, 43.13, 44.94, 43.61, 44.33, 44.83, 45.10, 45.42, 45.84, 46.08,
            45.89, 46.03, 45.61, 46.28, 46.28, 46.00, 46.03, 46.41, 46.22, 45.64,
            46.21, 46.25, 45.71, 46.45, 45.78, 45.35, 44.03, 44.18, 44.22, 44.57,
            43.42, 42.66, 43.13, 44.94, 43.61, 44.33, 44.83, 45.10, 45.42, 45.84,
            46.08, 45.89, 46.03, 45.61, 46.28, 46.28, 46.00, 46.03, 46.41, 46.22,
            45.64, 46.21, 46.25, 45.71, 46.45, 45.78, 45.35, 44.03, 44.18, 44.22,
            44.57, 43.42, 42.66, 43.13, 44.94, 43.61, 44.33, 44.83, 45.10, 45.42,
            45.84, 46.08, 45.89, 46.03, 45.61, 46.28, 46.28, 46.00, 46.03, 46.41,
            46.22, 45.64, 46.21, 46.25, 45.71, 46.45, 45.78, 45.35, 44.03, 44.18,
            44.22, 44.57, 43.42, 42.66, 43.13, 44.94, 43.61, 44.33, 44.83, 45.10,
            45.42, 45.84, 46.08, 45.89, 46.03, 45.61, 46.28, 46.28, 46.00, 46.03,
            46.41, 46.22, 45.64, 46.21, 46.25, 45.71, 46.45, 45.78, 45.35, 44.03,
            44.18, 44.22, 44.57, 43.42, 42.66, 43.13, 44.94, 43.61, 44.33, 44.83,
            45.10, 45.42, 45.84, 46.08, 45.89, 46.03, 45.61, 46.28, 46.28, 46.00,
            46.03, 46.41, 46.22, 45.64, 46.21, 46.25, 45.71, 46.45, 45.78, 45.35,
            44.03, 44.18, 44.22, 44.57, 43.42, 42.66, 43.13, 44.94, 43.61, 44.33,
            44.83, 45.10, 45.42, 45.84, 46.08, 45.89, 46.03, 45.61, 46.28, 46.28,
            46.00, 46.03, 46.41, 46.22, 45.64, 46.21, 46.25, 45.71, 46.45, 45.78,
            45.35, 44.03, 44.18, 44.22, 44.57, 43.42, 42.66, 43.13, 44.94, 43.61,
            44.33, 44.83, 45.10, 45.42, 45.84, 46.08, 45.89, 46.03, 45.61, 46.28,
            46.28, 46.00, 46.03, 46.41, 46.22, 45.64, 46.21, 46.25, 45.71, 46.45,
            45.78, 45.35, 44.03, 44.18, 44.22, 44.57, 43.42, 42.66, 43.13, 44.94,
            43.61, 44.33,
        ]
        rsi = RelativeStrengthIndex(RelativeStrengthIndexParams(length=14))
        act = math.nan
        for i, v in enumerate(input2):
            act = rsi.update(v)
            if i < 14:
                self.assertTrue(math.isnan(act), f"[{i}] expected NaN")
        # Final value should be in [0, 100].
        self.assertGreaterEqual(act, 0)
        self.assertLessEqual(act, 100)


if __name__ == "__main__":
    unittest.main()
