import math
import unittest

from py.indicators.marc_chaikin.advance_decline_oscillator.advance_decline_oscillator import AdvanceDeclineOscillator
from py.indicators.marc_chaikin.advance_decline_oscillator.params import AdvanceDeclineOscillatorParams, MovingAverageType
from py.entities.scalar import Scalar
from py.indicators.core.identifier import Identifier

from .test_testdata import (
    HIGHS,
    LOWS,
    CLOSES,
    VOLUMES,
    EXPECTED_EMA,
    EXPECTED_SMA,
)


# High test data, 252 entries. From TA-Lib excel-sma3-sma10-chaikin.csv.
# Low test data, 252 entries.
# Close test data, 252 entries.
# Volume test data, 252 entries.
# Expected EMA ADOSC output, 252 entries. First 9 are NaN (lookback=9), then valid.
# Expected SMA ADOSC output, 252 entries. First 9 are NaN, then 243 valid.
class TestAdvanceDeclineOscillator(unittest.TestCase):

    def test_ema_full_data(self):
        """Test ADOSC with EMA(3,10) against all 252 bars."""
        params = AdvanceDeclineOscillatorParams(
            fast_length=3,
            slow_length=10,
            moving_average_type=MovingAverageType.EMA,
            first_is_average=False,
        )
        ind = AdvanceDeclineOscillator(params)

        for i in range(252):
            v = ind.update_hlcv(HIGHS[i], LOWS[i], CLOSES[i], VOLUMES[i])

            if i < 9:
                self.assertTrue(math.isnan(v), f"[{i}] expected NaN, got {v}")
                self.assertFalse(ind.is_primed(), f"[{i}] expected not primed")
                continue

            self.assertFalse(math.isnan(v), f"[{i}] expected non-NaN, got NaN")
            self.assertTrue(ind.is_primed(), f"[{i}] expected primed")
            self.assertAlmostEqual(v, EXPECTED_EMA[i], places=2,
                                   msg=f"[{i}] EMA mismatch")

    def test_sma_full_data(self):
        """Test ADOSC with SMA(3,10) against all 252 bars."""
        params = AdvanceDeclineOscillatorParams(
            fast_length=3,
            slow_length=10,
            moving_average_type=MovingAverageType.SMA,
        )
        ind = AdvanceDeclineOscillator(params)

        for i in range(252):
            v = ind.update_hlcv(HIGHS[i], LOWS[i], CLOSES[i], VOLUMES[i])

            if i < 9:
                self.assertTrue(math.isnan(v), f"[{i}] expected NaN, got {v}")
                self.assertFalse(ind.is_primed(), f"[{i}] expected not primed")
                continue

            self.assertFalse(math.isnan(v), f"[{i}] expected non-NaN, got NaN")
            self.assertTrue(ind.is_primed(), f"[{i}] expected primed")
            self.assertAlmostEqual(v, EXPECTED_SMA[i], places=2,
                                   msg=f"[{i}] SMA mismatch")

    def test_is_primed(self):
        """Test that the indicator primes at exactly the 10th bar (index 9)."""
        params = AdvanceDeclineOscillatorParams(
            fast_length=3,
            slow_length=10,
            moving_average_type=MovingAverageType.EMA,
            first_is_average=False,
        )
        ind = AdvanceDeclineOscillator(params)

        # Feed first 9 bars — should not be primed.
        for i in range(9):
            ind.update_hlcv(HIGHS[i], LOWS[i], CLOSES[i], VOLUMES[i])
            self.assertFalse(ind.is_primed(), f"[{i}] expected not primed")

        # Feed 10th bar — should become primed.
        ind.update_hlcv(HIGHS[9], LOWS[9], CLOSES[9], VOLUMES[9])
        self.assertTrue(ind.is_primed(), "[9] expected primed")

    def test_metadata(self):
        """Test metadata identifier and mnemonic."""
        params = AdvanceDeclineOscillatorParams(
            fast_length=3,
            slow_length=10,
            moving_average_type=MovingAverageType.EMA,
            first_is_average=False,
        )
        ind = AdvanceDeclineOscillator(params)
        meta = ind.metadata()

        self.assertEqual(meta.identifier, Identifier.ADVANCE_DECLINE_OSCILLATOR)
        self.assertIn("adosc(", meta.mnemonic)

    def test_invalid_params(self):
        """Test that fast_length < 2 raises ValueError."""
        with self.assertRaises(ValueError):
            AdvanceDeclineOscillator(AdvanceDeclineOscillatorParams(
                fast_length=1,
                slow_length=10,
                moving_average_type=MovingAverageType.EMA,
            ))

        with self.assertRaises(ValueError):
            AdvanceDeclineOscillator(AdvanceDeclineOscillatorParams(
                fast_length=3,
                slow_length=1,
                moving_average_type=MovingAverageType.EMA,
            ))


if __name__ == '__main__':
    unittest.main()
