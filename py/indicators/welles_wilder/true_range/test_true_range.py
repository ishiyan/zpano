"""Tests for Welles Wilder's True Range indicator."""

import math
import unittest

from .true_range import TrueRange
from .params import TrueRangeParams
from .output import TrueRangeOutput
from ...core.identifier import Identifier

from .test_testdata import (
    _HIGH,
    _LOW,
    _CLOSE,
    _EXPECTED_TR,
)


# fmt: off
# fmt: on


class TestTrueRange(unittest.TestCase):
    """Tests for the TrueRange indicator."""

    def test_update(self) -> None:
        """Test update with 252 bars of TA-Lib reference data."""
        tr = TrueRange()

        for i in range(len(_CLOSE)):
            act = tr.update(_CLOSE[i], _HIGH[i], _LOW[i])

            if i == 0:
                self.assertTrue(math.isnan(act), f"[{i}] expected NaN, got {act}")
                continue

            exp = _EXPECTED_TR[i]
            self.assertFalse(math.isnan(act), f"[{i}] expected {exp}, got NaN")
            self.assertAlmostEqual(act, exp, delta=1e-3,
                                   msg=f"[{i}] expected {exp}, got {act}")

    def test_nan_passthrough(self) -> None:
        """NaN in any input produces NaN output."""
        tr = TrueRange()

        self.assertTrue(math.isnan(tr.update(math.nan, 1, 1)),
                        "expected NaN passthrough for NaN close")
        self.assertTrue(math.isnan(tr.update(1, math.nan, 1)),
                        "expected NaN passthrough for NaN high")
        self.assertTrue(math.isnan(tr.update(1, 1, math.nan)),
                        "expected NaN passthrough for NaN low")

    def test_is_primed(self) -> None:
        """Priming requires two updates."""
        tr = TrueRange()

        self.assertFalse(tr.is_primed(), "should not be primed before any updates")

        tr.update(_CLOSE[0], _HIGH[0], _LOW[0])
        self.assertFalse(tr.is_primed(), "should not be primed after first update")

        tr.update(_CLOSE[1], _HIGH[1], _LOW[1])
        self.assertTrue(tr.is_primed(), "should be primed after second update")

        tr.update(_CLOSE[2], _HIGH[2], _LOW[2])
        self.assertTrue(tr.is_primed(), "should remain primed")

    def test_update_sample(self) -> None:
        """When H=L=C, TR = abs(current - previous)."""
        tr = TrueRange()

        v = tr.update_sample(100.0)
        self.assertTrue(math.isnan(v), f"expected NaN, got {v}")

        v = tr.update_sample(105.0)
        self.assertAlmostEqual(v, 5.0, delta=1e-10)

        v = tr.update_sample(102.0)
        self.assertAlmostEqual(v, 3.0, delta=1e-10)

    def test_metadata(self) -> None:
        """Metadata fields are correct."""
        tr = TrueRange()
        md = tr.metadata()

        self.assertEqual(md.identifier, Identifier.TRUE_RANGE)
        self.assertEqual(md.mnemonic, "tr")
        self.assertEqual(md.description, "True Range")
        self.assertEqual(len(md.outputs), 1)
        self.assertEqual(md.outputs[0].kind, int(TrueRangeOutput.VALUE))
        self.assertEqual(md.outputs[0].mnemonic, "tr")
        self.assertEqual(md.outputs[0].description, "True Range")


if __name__ == '__main__':
    unittest.main()
