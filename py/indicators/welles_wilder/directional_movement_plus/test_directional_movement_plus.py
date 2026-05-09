"""Tests for Welles Wilder's Directional Movement Plus indicator."""

import math
import unittest
from datetime import datetime

from .directional_movement_plus import DirectionalMovementPlus
from .params import DirectionalMovementPlusParams
from .output import DirectionalMovementPlusOutput
from ...core.identifier import Identifier
from ....entities.bar import Bar
from ....entities.quote import Quote
from ....entities.trade import Trade
from ....entities.scalar import Scalar

from .test_testdata import (
    _HIGH,
    _LOW,
    _EXPECTED_DMP1,
    _EXPECTED_DMP14,
)


# fmt: off
# fmt: on


class TestDirectionalMovementPlus(unittest.TestCase):
    """Tests for the DirectionalMovementPlus indicator."""

    def test_constructor_invalid(self) -> None:
        """Length 0 and negative should raise ValueError."""
        with self.assertRaises(ValueError):
            DirectionalMovementPlus(DirectionalMovementPlusParams(length=0))

        with self.assertRaises(ValueError):
            DirectionalMovementPlus(DirectionalMovementPlusParams(length=-8))

    def test_is_primed_length1(self) -> None:
        """Length=1 requires 2 updates to prime."""
        dmp = DirectionalMovementPlus(DirectionalMovementPlusParams(length=1))
        self.assertFalse(dmp.is_primed())

        dmp.update(_HIGH[0], _LOW[0])
        self.assertFalse(dmp.is_primed(), "[0] should not be primed yet")

        dmp.update(_HIGH[1], _LOW[1])
        self.assertTrue(dmp.is_primed(), "[1] should be primed")

    def test_is_primed_length14(self) -> None:
        """Length=14 requires 15 updates to prime."""
        dmp = DirectionalMovementPlus(DirectionalMovementPlusParams(length=14))

        for i in range(14):
            dmp.update(_HIGH[i], _LOW[i])
            self.assertFalse(dmp.is_primed(), f"[{i}] should not be primed yet")

        dmp.update(_HIGH[14], _LOW[14])
        self.assertTrue(dmp.is_primed(), "[14] should be primed")

    def test_values_length1(self) -> None:
        """Feed all 252 high/low pairs, check DMP1 output."""
        dmp = DirectionalMovementPlus(DirectionalMovementPlusParams(length=1))

        for i in range(len(_HIGH)):
            act = dmp.update(_HIGH[i], _LOW[i])
            exp = _EXPECTED_DMP1[i]

            if exp is None:
                self.assertTrue(math.isnan(act), f"[{i}] expected NaN, got {act}")
            else:
                self.assertFalse(math.isnan(act), f"[{i}] expected {exp}, got NaN")
                self.assertAlmostEqual(act, exp, delta=1e-8,
                                       msg=f"[{i}] expected {exp}, got {act}")

    def test_values_length14(self) -> None:
        """Feed all 252 high/low pairs, check DMP14 output."""
        dmp = DirectionalMovementPlus(DirectionalMovementPlusParams(length=14))

        for i in range(len(_HIGH)):
            act = dmp.update(_HIGH[i], _LOW[i])
            exp = _EXPECTED_DMP14[i]

            if exp is None:
                self.assertTrue(math.isnan(act), f"[{i}] expected NaN, got {act}")
            else:
                self.assertFalse(math.isnan(act), f"[{i}] expected {exp}, got NaN")
                self.assertAlmostEqual(act, exp, delta=1e-8,
                                       msg=f"[{i}] expected {exp}, got {act}")

    def test_nan_high(self) -> None:
        """NaN high produces NaN output."""
        dmp = DirectionalMovementPlus(DirectionalMovementPlusParams(length=14))
        self.assertTrue(math.isnan(dmp.update(math.nan, 1)),
                        "expected NaN passthrough for NaN high")

    def test_nan_low(self) -> None:
        """NaN low produces NaN output."""
        dmp = DirectionalMovementPlus(DirectionalMovementPlusParams(length=14))
        self.assertTrue(math.isnan(dmp.update(1, math.nan)),
                        "expected NaN passthrough for NaN low")

    def test_nan_both(self) -> None:
        """NaN high and low produces NaN output."""
        dmp = DirectionalMovementPlus(DirectionalMovementPlusParams(length=14))
        self.assertTrue(math.isnan(dmp.update(math.nan, math.nan)),
                        "expected NaN passthrough for NaN high and low")

    def test_nan_sample(self) -> None:
        """NaN sample produces NaN output."""
        dmp = DirectionalMovementPlus(DirectionalMovementPlusParams(length=14))
        self.assertTrue(math.isnan(dmp.update_sample(math.nan)),
                        "expected NaN passthrough for NaN sample")

    def test_high_low_swap(self) -> None:
        """When high < low, they should be swapped internally."""
        dmp1 = DirectionalMovementPlus(DirectionalMovementPlusParams(length=1))
        dmp2 = DirectionalMovementPlus(DirectionalMovementPlusParams(length=1))

        # Prime both.
        dmp1.update(10, 5)
        dmp2.update(5, 10)  # Swapped.

        # Update both with same effective values.
        v1 = dmp1.update(12, 6)
        v2 = dmp2.update(6, 12)  # Swapped.

        self.assertEqual(v1, v2, f"high/low swap should produce same result: {v1} vs {v2}")

    def test_zero_inputs(self) -> None:
        """20 updates of (0,0) with length=10."""
        dmp = DirectionalMovementPlus(DirectionalMovementPlusParams(length=10))

        for _ in range(20):
            dmp.update_sample(0)

        self.assertTrue(dmp.is_primed(), "should be primed after 20 updates with length 10")

    def test_entity_bar(self) -> None:
        """update_bar with Bar."""
        dmp = DirectionalMovementPlus(DirectionalMovementPlusParams(length=14))
        for i in range(14):
            dmp.update(_HIGH[i], _LOW[i])

        tm = datetime(2021, 4, 1)
        b = Bar(time=tm, open=0, high=_HIGH[14], low=_LOW[14], close=0, volume=0)
        result = dmp.update_bar(b)

        self.assertEqual(len(result), 1)
        self.assertIsInstance(result[0], Scalar)
        self.assertEqual(result[0].time, tm)

    def test_entity_scalar(self) -> None:
        """update_scalar with Scalar."""
        dmp = DirectionalMovementPlus(DirectionalMovementPlusParams(length=14))
        for i in range(14):
            dmp.update(_HIGH[i], _LOW[i])

        tm = datetime(2021, 4, 1)
        s = Scalar(time=tm, value=_HIGH[14])
        result = dmp.update_scalar(s)

        self.assertEqual(len(result), 1)
        self.assertIsInstance(result[0], Scalar)
        self.assertEqual(result[0].time, tm)

    def test_entity_quote(self) -> None:
        """update_quote with Quote."""
        dmp = DirectionalMovementPlus(DirectionalMovementPlusParams(length=14))
        for i in range(14):
            dmp.update(_HIGH[i], _LOW[i])

        tm = datetime(2021, 4, 1)
        q = Quote(time=tm, bid_price=_HIGH[14] - 0.5, ask_price=_HIGH[14] + 0.5,
                  bid_size=0, ask_size=0)
        result = dmp.update_quote(q)

        self.assertEqual(len(result), 1)
        self.assertIsInstance(result[0], Scalar)
        self.assertEqual(result[0].time, tm)

    def test_entity_trade(self) -> None:
        """update_trade with Trade."""
        dmp = DirectionalMovementPlus(DirectionalMovementPlusParams(length=14))
        for i in range(14):
            dmp.update(_HIGH[i], _LOW[i])

        tm = datetime(2021, 4, 1)
        t = Trade(time=tm, price=_HIGH[14], volume=0)
        result = dmp.update_trade(t)

        self.assertEqual(len(result), 1)
        self.assertIsInstance(result[0], Scalar)
        self.assertEqual(result[0].time, tm)

    def test_metadata(self) -> None:
        """Metadata fields are correct."""
        dmp = DirectionalMovementPlus(DirectionalMovementPlusParams(length=14))
        md = dmp.metadata()

        self.assertEqual(md.identifier, Identifier.DIRECTIONAL_MOVEMENT_PLUS)
        self.assertEqual(md.mnemonic, "+dm")
        self.assertEqual(md.description, "Directional Movement Plus")
        self.assertEqual(len(md.outputs), 1)
        self.assertEqual(md.outputs[0].kind, int(DirectionalMovementPlusOutput.VALUE))
        self.assertEqual(md.outputs[0].mnemonic, "+dm")
        self.assertEqual(md.outputs[0].description, "Directional Movement Plus")


if __name__ == '__main__':
    unittest.main()
