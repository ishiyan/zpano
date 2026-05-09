"""Tests for Marc Chaikin's Advance-Decline indicator."""

import math
import unittest
from datetime import datetime

from .advance_decline import AdvanceDecline
from .params import AdvanceDeclineParams
from ...core.identifier import Identifier
from ....entities.bar import Bar
from ....entities.scalar import Scalar

from .test_testdata import (
    _HIGHS,
    _LOWS,
    _CLOSES,
    _VOLUMES,
    _EXPECTED,
)


class TestAdvanceDecline(unittest.TestCase):
    """Tests for the AdvanceDecline indicator."""

    def test_full_data(self) -> None:
        """Feed all 252 bars via update_hlcv, compare at 2 decimal places."""
        ad = AdvanceDecline(AdvanceDeclineParams())

        for i in range(len(_HIGHS)):
            v = ad.update_hlcv(_HIGHS[i], _LOWS[i], _CLOSES[i], _VOLUMES[i])
            self.assertFalse(math.isnan(v), f"[{i}] expected non-NaN")
            self.assertTrue(ad.is_primed(), f"[{i}] expected primed")
            self.assertAlmostEqual(v, _EXPECTED[i], places=2,
                                   msg=f"[{i}] expected {_EXPECTED[i]}, got {v}")

    def test_is_primed(self) -> None:
        """Not primed initially, primed after first update."""
        ad = AdvanceDecline(AdvanceDeclineParams())
        self.assertFalse(ad.is_primed())

        ad.update_hlcv(_HIGHS[0], _LOWS[0], _CLOSES[0], _VOLUMES[0])
        self.assertTrue(ad.is_primed())

    def test_metadata(self) -> None:
        """Metadata fields are correct."""
        ad = AdvanceDecline(AdvanceDeclineParams())
        md = ad.metadata()

        self.assertEqual(md.identifier, Identifier.ADVANCE_DECLINE)
        self.assertEqual(md.mnemonic, "ad")
        self.assertEqual(md.description, "Advance-Decline")
        self.assertEqual(len(md.outputs), 1)

    def test_update_bar(self) -> None:
        """Test update_bar extracts HLCV and returns Scalar list."""
        ad = AdvanceDecline(AdvanceDeclineParams())
        tm = datetime(2021, 4, 1)

        for i in range(10):
            bar = Bar(time=tm, open=_HIGHS[i], high=_HIGHS[i],
                      low=_LOWS[i], close=_CLOSES[i], volume=_VOLUMES[i])
            output = ad.update_bar(bar)
            self.assertIsInstance(output, list)
            self.assertEqual(len(output), 1)

            scalar = output[0]
            self.assertIsInstance(scalar, Scalar)
            self.assertAlmostEqual(scalar.value, _EXPECTED[i], places=2,
                                   msg=f"[{i}] expected {_EXPECTED[i]}, got {scalar.value}")

    def test_scalar_update(self) -> None:
        """Scalar update: H=L=C, volume=1, range=0, AD stays 0."""
        ad = AdvanceDecline(AdvanceDeclineParams())
        v = ad.update(100.0)
        self.assertEqual(v, 0.0)
        self.assertTrue(ad.is_primed())

    def test_nan_passthrough(self) -> None:
        """NaN inputs produce NaN output."""
        ad = AdvanceDecline(AdvanceDeclineParams())
        self.assertTrue(math.isnan(ad.update(math.nan)))
        self.assertTrue(math.isnan(ad.update_hlcv(math.nan, 1, 2, 3)))
        self.assertTrue(math.isnan(ad.update_hlcv(1, math.nan, 2, 3)))
        self.assertTrue(math.isnan(ad.update_hlcv(1, 2, math.nan, 3)))
        self.assertTrue(math.isnan(ad.update_hlcv(1, 2, 3, math.nan)))

    def test_not_primed_after_nan(self) -> None:
        """Not primed after NaN-only updates."""
        ad = AdvanceDecline(AdvanceDeclineParams())
        self.assertFalse(ad.is_primed())
        ad.update(math.nan)
        self.assertFalse(ad.is_primed())


if __name__ == '__main__':
    unittest.main()
