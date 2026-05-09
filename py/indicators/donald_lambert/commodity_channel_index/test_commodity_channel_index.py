"""Tests for Donald Lambert's Commodity Channel Index indicator."""

import math
import unittest

from .commodity_channel_index import CommodityChannelIndex
from .params import CommodityChannelIndexParams, default_params
from .output import CommodityChannelIndexOutput
from ...core.identifier import Identifier

from .test_testdata import _INPUT


# fmt: off
# Test data from TA-Lib (252 entries), typical price input.
# fmt: on


class TestCommodityChannelIndex(unittest.TestCase):
    """Tests for the CommodityChannelIndex indicator."""

    def test_length_11(self) -> None:
        """Test CCI with length=11 on 252 typical prices."""
        tolerance = 5e-8
        cci = CommodityChannelIndex(CommodityChannelIndexParams(length=11))

        # First 10 values should be NaN.
        for i in range(10):
            v = cci.update(_INPUT[i])
            self.assertTrue(math.isnan(v), f"[{i}] expected NaN, got {v}")

        # Index 10: first value.
        v = cci.update(_INPUT[10])
        self.assertFalse(math.isnan(v), f"[10] expected non-NaN, got NaN")
        self.assertAlmostEqual(v, 87.92686612269590, delta=tolerance,
                               msg=f"[10] expected ~87.9269, got {v}")

        # Index 11.
        v = cci.update(_INPUT[11])
        self.assertAlmostEqual(v, 180.00543014506300, delta=tolerance,
                               msg=f"[11] expected ~180.0054, got {v}")

        # Feed remaining and check last.
        for i in range(12, 251):
            cci.update(_INPUT[i])

        v = cci.update(_INPUT[251])
        self.assertAlmostEqual(v, -169.65514382823800, delta=tolerance,
                               msg=f"[251] expected ~-169.6551, got {v}")

        self.assertTrue(cci.is_primed())

    def test_length_2(self) -> None:
        """Test CCI with length=2."""
        tolerance = 5e-7
        cci = CommodityChannelIndex(CommodityChannelIndexParams(length=2))

        # First value should be NaN.
        v = cci.update(_INPUT[0])
        self.assertTrue(math.isnan(v), f"[0] expected NaN, got {v}")

        # Index 1: first value.
        v = cci.update(_INPUT[1])
        self.assertFalse(math.isnan(v))
        self.assertAlmostEqual(v, 66.66666666666670, delta=tolerance)

        # Feed remaining and check last.
        for i in range(2, 251):
            cci.update(_INPUT[i])

        v = cci.update(_INPUT[251])
        self.assertAlmostEqual(v, -66.66666666666590, delta=tolerance)

    def test_is_primed(self) -> None:
        """Priming requires length samples."""
        cci = CommodityChannelIndex(CommodityChannelIndexParams(length=5))
        self.assertFalse(cci.is_primed())

        for i in range(1, 5):
            cci.update(float(i))
            self.assertFalse(cci.is_primed(), f"[{i}] expected not primed")

        cci.update(5.0)
        self.assertTrue(cci.is_primed())

        cci.update(6.0)
        self.assertTrue(cci.is_primed())

    def test_nan_passthrough(self) -> None:
        """NaN input produces NaN output."""
        cci = CommodityChannelIndex(CommodityChannelIndexParams(length=5))
        v = cci.update(math.nan)
        self.assertTrue(math.isnan(v))

    def test_metadata(self) -> None:
        """Metadata fields are correct."""
        cci = CommodityChannelIndex(CommodityChannelIndexParams(length=20))
        md = cci.metadata()

        self.assertEqual(md.identifier, Identifier.COMMODITY_CHANNEL_INDEX)
        self.assertEqual(md.mnemonic, "cci(20)")
        self.assertEqual(len(md.outputs), 1)
        self.assertEqual(md.outputs[0].kind, int(CommodityChannelIndexOutput.VALUE))

    def test_invalid_params(self) -> None:
        """Invalid length raises ValueError."""
        for length in [1, 0, -8]:
            with self.assertRaises(ValueError, msg=f"length={length}"):
                CommodityChannelIndex(CommodityChannelIndexParams(length=length))

    def test_custom_scaling_factor(self) -> None:
        """Custom inverse scaling factor works."""
        cci = CommodityChannelIndex(CommodityChannelIndexParams(
            length=5, inverse_scaling_factor=0.03))

        for i in range(1, 6):
            cci.update(float(i))

        self.assertTrue(cci.is_primed())


if __name__ == '__main__':
    unittest.main()
