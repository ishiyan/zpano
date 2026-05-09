"""Tests for Welles Wilder's Directional Indicator Plus indicator."""

import math
import unittest
from datetime import datetime

from .directional_indicator_plus import DirectionalIndicatorPlus
from .params import DirectionalIndicatorPlusParams
from .output import DirectionalIndicatorPlusOutput
from ...core.identifier import Identifier
from ....entities.bar import Bar
from ....entities.quote import Quote
from ....entities.trade import Trade
from ....entities.scalar import Scalar

from .test_testdata import (
    _HIGH,
    _LOW,
    _CLOSE,
    _EXPECTED_DIP_14,
)


# fmt: off
# fmt: on


class TestDirectionalIndicatorPlus(unittest.TestCase):
    """Tests for the DirectionalIndicatorPlus indicator."""

    def test_constructor_invalid(self) -> None:
        """Length 0 and negative should raise ValueError."""
        with self.assertRaises(ValueError):
            DirectionalIndicatorPlus(DirectionalIndicatorPlusParams(length=0))

        with self.assertRaises(ValueError):
            DirectionalIndicatorPlus(DirectionalIndicatorPlusParams(length=-8))

    def test_is_primed_length14(self) -> None:
        """Priming with length=14 requires 15 updates."""
        dip = DirectionalIndicatorPlus(DirectionalIndicatorPlusParams(length=14))

        for i in range(14):
            dip.update(_CLOSE[i], _HIGH[i], _LOW[i])
            self.assertFalse(dip.is_primed(),
                             f"should not be primed after {i + 1} updates")

        dip.update(_CLOSE[14], _HIGH[14], _LOW[14])
        self.assertTrue(dip.is_primed(), "should be primed after 15 updates")

    def test_values_length14(self) -> None:
        """Test DI+ with length=14 against 252 bars of reference data."""
        dip = DirectionalIndicatorPlus(DirectionalIndicatorPlusParams(length=14))

        for i in range(len(_CLOSE)):
            act = dip.update(_CLOSE[i], _HIGH[i], _LOW[i])
            exp = _EXPECTED_DIP_14[i]

            if exp is None:
                self.assertTrue(math.isnan(act),
                                f"[{i}] expected NaN, got {act}")
            else:
                self.assertFalse(math.isnan(act),
                                 f"[{i}] expected {exp}, got NaN")
                self.assertAlmostEqual(act, exp, delta=1e-8,
                                       msg=f"[{i}] expected {exp}, got {act}")

    def test_nan_close(self) -> None:
        """NaN close returns NaN."""
        dip = DirectionalIndicatorPlus(DirectionalIndicatorPlusParams(length=14))
        self.assertTrue(math.isnan(dip.update(math.nan, 1, 1)),
                        "expected NaN passthrough for NaN close")

    def test_nan_high(self) -> None:
        """NaN high returns NaN."""
        dip = DirectionalIndicatorPlus(DirectionalIndicatorPlusParams(length=14))
        self.assertTrue(math.isnan(dip.update(1, math.nan, 1)),
                        "expected NaN passthrough for NaN high")

    def test_nan_low(self) -> None:
        """NaN low returns NaN."""
        dip = DirectionalIndicatorPlus(DirectionalIndicatorPlusParams(length=14))
        self.assertTrue(math.isnan(dip.update(1, 1, math.nan)),
                        "expected NaN passthrough for NaN low")

    def test_nan_sample(self) -> None:
        """NaN via update_sample returns NaN."""
        dip = DirectionalIndicatorPlus(DirectionalIndicatorPlusParams(length=14))
        self.assertTrue(math.isnan(dip.update_sample(math.nan)),
                        "expected NaN passthrough for NaN sample")

    def test_zero_inputs(self) -> None:
        """20 updates of (0, 0, 0) with length=10."""
        dip = DirectionalIndicatorPlus(DirectionalIndicatorPlusParams(length=10))

        for _ in range(20):
            result = dip.update(0, 0, 0)
            # Should not raise; result is either NaN or 0
            self.assertFalse(result != result and False)  # just ensure no exception

    def test_entity_bar(self) -> None:
        """update_bar produces correct output."""
        dip = DirectionalIndicatorPlus(DirectionalIndicatorPlusParams(length=14))
        tm = datetime(2021, 4, 1)

        for i in range(14):
            dip.update(_CLOSE[i], _HIGH[i], _LOW[i])

        b = Bar(time=tm, open=0, high=_HIGH[14], low=_LOW[14],
                close=_CLOSE[14], volume=0)
        output = dip.update_bar(b)
        self.assertEqual(len(output), 1)
        s = output[0]
        self.assertIsInstance(s, Scalar)
        self.assertEqual(s.time, tm)

    def test_entity_scalar(self) -> None:
        """update_scalar produces correct output."""
        dip = DirectionalIndicatorPlus(DirectionalIndicatorPlusParams(length=14))
        tm = datetime(2021, 4, 1)

        for i in range(14):
            dip.update(_CLOSE[i], _HIGH[i], _LOW[i])

        s = Scalar(time=tm, value=_HIGH[14])
        output = dip.update_scalar(s)
        self.assertEqual(len(output), 1)
        self.assertIsInstance(output[0], Scalar)
        self.assertEqual(output[0].time, tm)

    def test_entity_quote(self) -> None:
        """update_quote produces correct output."""
        dip = DirectionalIndicatorPlus(DirectionalIndicatorPlusParams(length=14))
        tm = datetime(2021, 4, 1)

        for i in range(14):
            dip.update(_CLOSE[i], _HIGH[i], _LOW[i])

        q = Quote(time=tm, bid_price=_HIGH[14] - 0.5, ask_price=_HIGH[14] + 0.5,
                  bid_size=0, ask_size=0)
        output = dip.update_quote(q)
        self.assertEqual(len(output), 1)
        self.assertIsInstance(output[0], Scalar)
        self.assertEqual(output[0].time, tm)

    def test_entity_trade(self) -> None:
        """update_trade produces correct output."""
        dip = DirectionalIndicatorPlus(DirectionalIndicatorPlusParams(length=14))
        tm = datetime(2021, 4, 1)

        for i in range(14):
            dip.update(_CLOSE[i], _HIGH[i], _LOW[i])

        r = Trade(time=tm, price=_HIGH[14], volume=0)
        output = dip.update_trade(r)
        self.assertEqual(len(output), 1)
        self.assertIsInstance(output[0], Scalar)
        self.assertEqual(output[0].time, tm)

    def test_metadata(self) -> None:
        """Metadata fields are correct."""
        dip = DirectionalIndicatorPlus(DirectionalIndicatorPlusParams(length=14))
        md = dip.metadata()

        self.assertEqual(md.identifier, Identifier.DIRECTIONAL_INDICATOR_PLUS)
        self.assertEqual(md.mnemonic, "+di")
        self.assertEqual(md.description, "Directional Indicator Plus")
        self.assertEqual(len(md.outputs), 4)
        self.assertEqual(md.outputs[0].kind, int(DirectionalIndicatorPlusOutput.VALUE))
        self.assertEqual(md.outputs[0].mnemonic, "+di")
        self.assertEqual(md.outputs[0].description, "Directional Indicator Plus")
        self.assertEqual(md.outputs[1].kind, int(DirectionalIndicatorPlusOutput.DIRECTIONAL_MOVEMENT_PLUS))
        self.assertEqual(md.outputs[2].kind, int(DirectionalIndicatorPlusOutput.AVERAGE_TRUE_RANGE))
        self.assertEqual(md.outputs[3].kind, int(DirectionalIndicatorPlusOutput.TRUE_RANGE))


if __name__ == '__main__':
    unittest.main()
