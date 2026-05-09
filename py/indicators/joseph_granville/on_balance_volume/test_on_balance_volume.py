"""Tests for Joseph Granville's On-Balance Volume indicator."""

import math
import unittest
from datetime import datetime

from .on_balance_volume import OnBalanceVolume
from .params import OnBalanceVolumeParams
from .output import OnBalanceVolumeOutput
from ...core.identifier import Identifier
from ....entities.bar import Bar
from ....entities.scalar import Scalar

from .test_testdata import _PRICES, _VOLUMES, _EXPECTED


def _round_to(v: float, digits: int) -> float:
    p = 10.0 ** digits
    return round(v * p) / p


class TestOnBalanceVolume(unittest.TestCase):
    """Tests for the OnBalanceVolume indicator."""

    def test_update_with_volume(self) -> None:
        """Test with 12 entries of C# reference data."""
        digits = 1
        obv = OnBalanceVolume(OnBalanceVolumeParams())

        for i in range(len(_PRICES)):
            act = obv.update_with_volume(float(_PRICES[i]), float(_VOLUMES[i]))
            self.assertFalse(math.isnan(act), f"[{i}] expected non-NaN")
            self.assertTrue(obv.is_primed(), f"[{i}] expected primed")

            got = _round_to(act, digits)
            exp = _round_to(float(_EXPECTED[i]), digits)
            self.assertEqual(got, exp, f"[{i}] expected {exp}, got {got}")

    def test_is_primed(self) -> None:
        """Not primed initially, primed after first update."""
        obv = OnBalanceVolume(OnBalanceVolumeParams())
        self.assertFalse(obv.is_primed())

        obv.update_with_volume(1.0, 100.0)
        self.assertTrue(obv.is_primed())

        obv.update_with_volume(2.0, 50.0)
        self.assertTrue(obv.is_primed())

    def test_nan_passthrough(self) -> None:
        """NaN inputs produce NaN output."""
        obv = OnBalanceVolume(OnBalanceVolumeParams())
        self.assertTrue(math.isnan(obv.update(math.nan)))
        self.assertTrue(math.isnan(obv.update_with_volume(1.0, math.nan)))
        self.assertTrue(math.isnan(obv.update_with_volume(math.nan, math.nan)))

    def test_metadata(self) -> None:
        """Metadata fields are correct."""
        obv = OnBalanceVolume(OnBalanceVolumeParams())
        md = obv.metadata()

        self.assertEqual(md.identifier, Identifier.ON_BALANCE_VOLUME)
        self.assertEqual(md.mnemonic, "obv")
        self.assertEqual(len(md.outputs), 1)
        self.assertEqual(md.outputs[0].kind, int(OnBalanceVolumeOutput.VALUE))

    def test_update_scalar(self) -> None:
        """UpdateScalar uses volume=1 path."""
        obv = OnBalanceVolume(OnBalanceVolumeParams())
        tm = datetime(2021, 4, 1)

        out = obv.update_scalar(Scalar(time=tm, value=10.0))
        v = out[0].value
        self.assertEqual(v, 1.0, "expected 1.0 (volume=1 on first call)")

    def test_update_bar(self) -> None:
        """Test update_bar with bar volume."""
        digits = 1
        obv = OnBalanceVolume(OnBalanceVolumeParams())
        tm = datetime(2021, 4, 1)

        for i in range(len(_PRICES)):
            bar = Bar(time=tm, open=0.0, high=0.0, low=0.0,
                      close=float(_PRICES[i]), volume=float(_VOLUMES[i]))
            out = obv.update_bar(bar)
            v = out[0].value

            got = _round_to(v, digits)
            exp = _round_to(float(_EXPECTED[i]), digits)
            self.assertEqual(got, exp, f"[{i}] expected {exp}, got {got}")

    def test_equal_prices(self) -> None:
        """When price unchanged, OBV stays the same."""
        obv = OnBalanceVolume(OnBalanceVolumeParams())

        v = obv.update_with_volume(10.0, 100.0)
        self.assertEqual(v, 100.0)

        v = obv.update_with_volume(10.0, 200.0)
        self.assertEqual(v, 100.0, "expected 100.0 (unchanged)")


if __name__ == '__main__':
    unittest.main()
