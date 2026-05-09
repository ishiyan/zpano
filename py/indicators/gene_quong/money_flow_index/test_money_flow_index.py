"""Tests for Gene Quong's Money Flow Index indicator."""

import math
import unittest

from .money_flow_index import MoneyFlowIndex
from .params import MoneyFlowIndexParams
from .output import MoneyFlowIndexOutput
from ...core.identifier import Identifier
from ....entities.bar import Bar
from ....entities.scalar import Scalar

from .test_testdata import (
    _TYPICAL_PRICES,
    _VOLUMES,
    _EXPECTED_MFI,
    _EXPECTED_MFI_VOL1,
)


def _round_to(v: float, digits: int) -> float:
    p = 10.0 ** digits
    return round(v * p) / p


# fmt: off
# fmt: on


class TestMoneyFlowIndex(unittest.TestCase):
    """Tests for the MoneyFlowIndex indicator."""

    def test_with_volume(self) -> None:
        """Test MFI(14) with real volume data."""
        digits = 9
        mfi = MoneyFlowIndex(MoneyFlowIndexParams(length=14))

        for i in range(14):
            v = mfi.update_with_volume(_TYPICAL_PRICES[i], _VOLUMES[i])
            self.assertTrue(math.isnan(v), f"[{i}] expected NaN")
            self.assertFalse(mfi.is_primed(), f"[{i}] expected not primed")

        for i in range(14, 252):
            v = mfi.update_with_volume(_TYPICAL_PRICES[i], _VOLUMES[i])
            self.assertFalse(math.isnan(v), f"[{i}] expected non-NaN")
            self.assertTrue(mfi.is_primed(), f"[{i}] expected primed")

            got = _round_to(v, digits)
            exp = _round_to(_EXPECTED_MFI[i], digits)
            self.assertEqual(got, exp, f"[{i}] expected {exp}, got {got}")

    def test_volume_1(self) -> None:
        """Test MFI(14) with volume=1 (Update path)."""
        digits = 9
        mfi = MoneyFlowIndex(MoneyFlowIndexParams(length=14))

        for i in range(14):
            v = mfi.update(_TYPICAL_PRICES[i])
            self.assertTrue(math.isnan(v), f"[{i}] expected NaN")

        for i in range(14, 252):
            v = mfi.update(_TYPICAL_PRICES[i])
            self.assertFalse(math.isnan(v), f"[{i}] expected non-NaN")

            got = _round_to(v, digits)
            exp = _round_to(_EXPECTED_MFI_VOL1[i], digits)
            self.assertEqual(got, exp, f"[{i}] expected {exp}, got {got}")

    def test_is_primed(self) -> None:
        """Priming requires length+1 samples."""
        mfi = MoneyFlowIndex(MoneyFlowIndexParams(length=5))
        self.assertFalse(mfi.is_primed())

        for i in range(1, 6):
            mfi.update(float(i))
            self.assertFalse(mfi.is_primed(), f"[{i}] expected not primed")

        mfi.update(5.0)
        self.assertTrue(mfi.is_primed())

        mfi.update(6.0)
        self.assertTrue(mfi.is_primed())

    def test_nan_passthrough(self) -> None:
        """NaN inputs produce NaN output."""
        mfi = MoneyFlowIndex(MoneyFlowIndexParams(length=5))
        self.assertTrue(math.isnan(mfi.update(math.nan)))
        self.assertTrue(math.isnan(mfi.update_with_volume(1.0, math.nan)))
        self.assertTrue(math.isnan(mfi.update_with_volume(math.nan, math.nan)))

    def test_metadata(self) -> None:
        """Metadata fields are correct."""
        mfi = MoneyFlowIndex(MoneyFlowIndexParams(length=14))
        md = mfi.metadata()

        self.assertEqual(md.identifier, Identifier.MONEY_FLOW_INDEX)
        self.assertEqual(md.mnemonic, "mfi(14, hlc/3)")
        self.assertEqual(len(md.outputs), 1)
        self.assertEqual(md.outputs[0].kind, int(MoneyFlowIndexOutput.VALUE))

    def test_invalid_params(self) -> None:
        """Invalid length raises ValueError."""
        for length in [0, -8]:
            with self.assertRaises(ValueError, msg=f"length={length}"):
                MoneyFlowIndex(MoneyFlowIndexParams(length=length))

    def test_small_sum(self) -> None:
        """When sum < 1, MFI should be 0."""
        mfi = MoneyFlowIndex(MoneyFlowIndexParams(length=2))

        for _ in range(10):
            mfi.update_with_volume(0.001, 0.5)

        self.assertTrue(mfi.is_primed())

        v = mfi.update_with_volume(0.001, 0.5)
        self.assertEqual(v, 0.0)


if __name__ == '__main__':
    unittest.main()
