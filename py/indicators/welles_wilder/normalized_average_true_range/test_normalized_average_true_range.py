"""Tests for Welles Wilder's Normalized Average True Range indicator."""

import math
import unittest

from .normalized_average_true_range import NormalizedAverageTrueRange
from .params import NormalizedAverageTrueRangeParams
from .output import NormalizedAverageTrueRangeOutput
from ...core.identifier import Identifier

from .test_testdata import (
    _HIGH,
    _LOW,
    _CLOSE,
    _EXPECTED_NATR_14,
    _EXPECTED_NATR_1,
)


# fmt: off
# fmt: on


class TestNormalizedAverageTrueRange(unittest.TestCase):
    """Tests for the NormalizedAverageTrueRange indicator."""

    def test_constructor(self) -> None:
        """Valid and invalid constructor parameters."""
        natr = NormalizedAverageTrueRange(NormalizedAverageTrueRangeParams(length=14))
        self.assertFalse(natr.is_primed())

        with self.assertRaises(ValueError):
            NormalizedAverageTrueRange(NormalizedAverageTrueRangeParams(length=0))

        with self.assertRaises(ValueError):
            NormalizedAverageTrueRange(NormalizedAverageTrueRangeParams(length=-8))

    def test_is_primed(self) -> None:
        """Priming with length=5 requires 6 updates (1 TR warmup + 5 ATR window)."""
        natr = NormalizedAverageTrueRange(NormalizedAverageTrueRangeParams(length=5))

        for i in range(5):
            natr.update(_CLOSE[i], _HIGH[i], _LOW[i])
            self.assertFalse(natr.is_primed(),
                             f"should not be primed after {i + 1} updates")

        natr.update(_CLOSE[5], _HIGH[5], _LOW[5])
        self.assertTrue(natr.is_primed(), "should be primed after 6 updates")

        natr.update(_CLOSE[6], _HIGH[6], _LOW[6])
        self.assertTrue(natr.is_primed(), "should remain primed")

    def test_update(self) -> None:
        """Test NATR with length=14 against 252 bars of reference data."""
        natr = NormalizedAverageTrueRange(NormalizedAverageTrueRangeParams(length=14))

        for i in range(len(_CLOSE)):
            act = natr.update(_CLOSE[i], _HIGH[i], _LOW[i])
            exp = _EXPECTED_NATR_14[i]

            if exp is None:
                self.assertTrue(math.isnan(act),
                                f"[{i}] expected NaN, got {act}")
            else:
                self.assertFalse(math.isnan(act),
                                 f"[{i}] expected {exp}, got NaN")
                self.assertAlmostEqual(act, exp, delta=1e-11,
                                       msg=f"[{i}] expected {exp}, got {act}")

    def test_length_1(self) -> None:
        """Test NATR with length=1 against 252 bars of reference data."""
        natr = NormalizedAverageTrueRange(NormalizedAverageTrueRangeParams(length=1))

        for i in range(len(_CLOSE)):
            act = natr.update(_CLOSE[i], _HIGH[i], _LOW[i])
            exp = _EXPECTED_NATR_1[i]

            if exp is None:
                self.assertTrue(math.isnan(act),
                                f"[{i}] expected NaN, got {act}")
            else:
                self.assertFalse(math.isnan(act),
                                 f"[{i}] expected {exp}, got NaN")
                self.assertAlmostEqual(act, exp, delta=1e-11,
                                       msg=f"[{i}] expected {exp}, got {act}")

    def test_close_zero(self) -> None:
        """When close == 0, NATR returns 0 instead of dividing by zero."""
        natr = NormalizedAverageTrueRange(NormalizedAverageTrueRangeParams(length=14))

        # Prime with 15 bars
        for i in range(15):
            natr.update(_CLOSE[i], _HIGH[i], _LOW[i])
        self.assertTrue(natr.is_primed())

        result = natr.update(0, 3.3, 2.2)
        self.assertEqual(result, 0)

    def test_nan_passthrough(self) -> None:
        """NaN in any input produces NaN output."""
        natr = NormalizedAverageTrueRange(NormalizedAverageTrueRangeParams(length=14))

        self.assertTrue(math.isnan(natr.update(math.nan, 1, 1)),
                        "expected NaN passthrough for NaN close")
        self.assertTrue(math.isnan(natr.update(1, math.nan, 1)),
                        "expected NaN passthrough for NaN high")
        self.assertTrue(math.isnan(natr.update(1, 1, math.nan)),
                        "expected NaN passthrough for NaN low")

        # Also test update_sample
        self.assertTrue(math.isnan(natr.update_sample(math.nan)),
                        "expected NaN passthrough for NaN sample")

    def test_metadata(self) -> None:
        """Metadata fields are correct."""
        natr = NormalizedAverageTrueRange(NormalizedAverageTrueRangeParams())
        md = natr.metadata()

        self.assertEqual(md.identifier, Identifier.NORMALIZED_AVERAGE_TRUE_RANGE)
        self.assertEqual(md.mnemonic, "natr")
        self.assertEqual(md.description, "Normalized Average True Range")
        self.assertEqual(len(md.outputs), 1)
        self.assertEqual(md.outputs[0].kind, int(NormalizedAverageTrueRangeOutput.VALUE))
        self.assertEqual(md.outputs[0].mnemonic, "natr")
        self.assertEqual(md.outputs[0].description, "Normalized Average True Range")


if __name__ == '__main__':
    unittest.main()
