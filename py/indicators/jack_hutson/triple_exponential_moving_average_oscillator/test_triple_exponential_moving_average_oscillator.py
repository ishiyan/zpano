"""Tests for Jack Hutson's Triple Exponential Moving Average Oscillator (TRIX)."""

import math
import unittest

from .triple_exponential_moving_average_oscillator import TripleExponentialMovingAverageOscillator
from .params import TripleExponentialMovingAverageOscillatorParams
from .output import TripleExponentialMovingAverageOscillatorOutput
from ...core.identifier import Identifier

from .test_testdata import _CLOSES, _EXPECTED


# fmt: off
# fmt: on


class TestTripleExponentialMovingAverageOscillator(unittest.TestCase):
    """Tests for the TRIX indicator."""

    def test_values(self) -> None:
        """Test TRIX(5) with 252 close prices."""
        tolerance = 1e-10
        trix = TripleExponentialMovingAverageOscillator(
            TripleExponentialMovingAverageOscillatorParams(length=5))

        for i in range(len(_CLOSES)):
            result = trix.update(_CLOSES[i])

            if _EXPECTED[i] is None:
                self.assertTrue(math.isnan(result), f"[{i}] expected NaN, got {result}")
            else:
                self.assertFalse(math.isnan(result), f"[{i}] expected {_EXPECTED[i]}, got NaN")
                self.assertAlmostEqual(result, _EXPECTED[i], delta=tolerance,
                                       msg=f"[{i}] expected {_EXPECTED[i]}, got {result}")

    def test_is_primed(self) -> None:
        """Priming requires 3*(L-1)+1 = 13 samples for L=5."""
        trix = TripleExponentialMovingAverageOscillator(
            TripleExponentialMovingAverageOscillatorParams(length=5))

        for i in range(13):
            trix.update(_CLOSES[i])
            self.assertFalse(trix.is_primed(), f"should not be primed at index {i}")

        trix.update(_CLOSES[13])
        self.assertTrue(trix.is_primed(), "should be primed at index 13")

    def test_metadata(self) -> None:
        """Metadata fields are correct."""
        trix = TripleExponentialMovingAverageOscillator(
            TripleExponentialMovingAverageOscillatorParams(length=30))
        md = trix.metadata()

        self.assertEqual(md.identifier, Identifier.TRIPLE_EXPONENTIAL_MOVING_AVERAGE_OSCILLATOR)
        self.assertEqual(md.mnemonic, "trix(30)")
        self.assertEqual(len(md.outputs), 1)
        self.assertEqual(md.outputs[0].kind, int(TripleExponentialMovingAverageOscillatorOutput.VALUE))

    def test_invalid_params(self) -> None:
        """Invalid length raises ValueError."""
        with self.assertRaises(ValueError):
            TripleExponentialMovingAverageOscillator(
                TripleExponentialMovingAverageOscillatorParams(length=0))

    def test_nan_passthrough(self) -> None:
        """NaN input produces NaN output."""
        trix = TripleExponentialMovingAverageOscillator(
            TripleExponentialMovingAverageOscillatorParams(length=5))
        self.assertTrue(math.isnan(trix.update(math.nan)))


if __name__ == '__main__':
    unittest.main()
