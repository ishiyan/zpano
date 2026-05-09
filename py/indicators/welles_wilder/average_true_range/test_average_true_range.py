"""Tests for Welles Wilder's Average True Range indicator."""

import math
import unittest

from .average_true_range import AverageTrueRange
from .params import AverageTrueRangeParams
from .output import AverageTrueRangeOutput
from ...core.identifier import Identifier

from .test_testdata import (
    _HIGH,
    _LOW,
    _CLOSE,
    _EXPECTED_TR,
    _EXPECTED_ATR,
)


# fmt: off
# fmt: on


class TestAverageTrueRange(unittest.TestCase):
    """Tests for the AverageTrueRange indicator."""

    def test_constructor(self) -> None:
        """Valid and invalid construction."""
        atr = AverageTrueRange(AverageTrueRangeParams(length=14))
        self.assertFalse(atr.is_primed())

        with self.assertRaises(ValueError):
            AverageTrueRange(AverageTrueRangeParams(length=0))

        with self.assertRaises(ValueError):
            AverageTrueRange(AverageTrueRangeParams(length=-8))

    def test_is_primed(self) -> None:
        """Priming requires length+1 updates (length=5 -> 6 updates)."""
        atr = AverageTrueRange(AverageTrueRangeParams(length=5))

        for i in range(5):
            atr.update(_CLOSE[i], _HIGH[i], _LOW[i])
            self.assertFalse(atr.is_primed(), f"should not be primed after {i + 1} updates")

        atr.update(_CLOSE[5], _HIGH[5], _LOW[5])
        self.assertTrue(atr.is_primed(), "should be primed after 6 updates")

    def test_update(self) -> None:
        """Test update with 252 bars, length=14."""
        atr = AverageTrueRange(AverageTrueRangeParams(length=14))

        for i in range(len(_CLOSE)):
            act = atr.update(_CLOSE[i], _HIGH[i], _LOW[i])
            exp = _EXPECTED_ATR[i]

            if exp is None:
                self.assertTrue(math.isnan(act), f"[{i}] expected NaN, got {act}")
            else:
                self.assertFalse(math.isnan(act), f"[{i}] expected {exp}, got NaN")
                self.assertAlmostEqual(act, exp, delta=1e-12,
                                       msg=f"[{i}] expected {exp}, got {act}")

    def test_length_1(self) -> None:
        """Length=1 ATR should equal TR values."""
        atr = AverageTrueRange(AverageTrueRangeParams(length=1))

        for i in range(len(_CLOSE)):
            act = atr.update(_CLOSE[i], _HIGH[i], _LOW[i])
            exp = _EXPECTED_TR[i]

            if exp is None:
                self.assertTrue(math.isnan(act), f"[{i}] expected NaN, got {act}")
            else:
                self.assertFalse(math.isnan(act), f"[{i}] expected {exp}, got NaN")
                self.assertAlmostEqual(act, exp, delta=1e-3,
                                       msg=f"[{i}] expected {exp}, got {act}")

    def test_nan_passthrough(self) -> None:
        """NaN in any input produces NaN output."""
        atr = AverageTrueRange(AverageTrueRangeParams(length=14))

        self.assertTrue(math.isnan(atr.update(math.nan, 1, 1)),
                        "expected NaN passthrough for NaN close")
        self.assertTrue(math.isnan(atr.update(1, math.nan, 1)),
                        "expected NaN passthrough for NaN high")
        self.assertTrue(math.isnan(atr.update(1, 1, math.nan)),
                        "expected NaN passthrough for NaN low")
        self.assertTrue(math.isnan(atr.update_sample(math.nan)),
                        "expected NaN passthrough for NaN sample")

    def test_metadata(self) -> None:
        """Metadata fields are correct."""
        atr = AverageTrueRange(AverageTrueRangeParams(length=14))
        md = atr.metadata()

        self.assertEqual(md.identifier, Identifier.AVERAGE_TRUE_RANGE)
        self.assertEqual(md.mnemonic, "atr")
        self.assertEqual(md.description, "Average True Range")
        self.assertEqual(len(md.outputs), 1)
        self.assertEqual(md.outputs[0].kind, int(AverageTrueRangeOutput.VALUE))
        self.assertEqual(md.outputs[0].mnemonic, "atr")
        self.assertEqual(md.outputs[0].description, "Average True Range")


if __name__ == '__main__':
    unittest.main()
