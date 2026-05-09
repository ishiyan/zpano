"""Tests for Welles Wilder's Directional Indicator Minus indicator."""

import math
import unittest

from .directional_indicator_minus import DirectionalIndicatorMinus
from .params import DirectionalIndicatorMinusParams
from .output import DirectionalIndicatorMinusOutput
from ...core.identifier import Identifier

from .test_testdata import (
    _HIGH,
    _LOW,
    _CLOSE,
    _EXPECTED_DIM_14,
)


# fmt: off
# fmt: on


class TestDirectionalIndicatorMinus(unittest.TestCase):
    """Tests for the DirectionalIndicatorMinus indicator."""

    def test_constructor(self) -> None:
        """Valid and invalid constructor parameters."""
        dim = DirectionalIndicatorMinus(DirectionalIndicatorMinusParams(length=14))
        self.assertFalse(dim.is_primed())

        with self.assertRaises(ValueError):
            DirectionalIndicatorMinus(DirectionalIndicatorMinusParams(length=0))

        with self.assertRaises(ValueError):
            DirectionalIndicatorMinus(DirectionalIndicatorMinusParams(length=-8))

    def test_is_primed(self) -> None:
        """Priming with length=14 requires 15 updates."""
        dim = DirectionalIndicatorMinus(DirectionalIndicatorMinusParams(length=14))

        for i in range(14):
            dim.update(_CLOSE[i], _HIGH[i], _LOW[i])
            self.assertFalse(dim.is_primed(),
                             f"should not be primed after {i + 1} updates")

        dim.update(_CLOSE[14], _HIGH[14], _LOW[14])
        self.assertTrue(dim.is_primed(), "should be primed after 15 updates")

    def test_update(self) -> None:
        """Test -DI with length=14 against 252 bars of reference data."""
        dim = DirectionalIndicatorMinus(DirectionalIndicatorMinusParams(length=14))

        for i in range(len(_CLOSE)):
            act = dim.update(_CLOSE[i], _HIGH[i], _LOW[i])
            exp = _EXPECTED_DIM_14[i]

            if exp is None:
                self.assertTrue(math.isnan(act),
                                f"[{i}] expected NaN, got {act}")
            else:
                self.assertFalse(math.isnan(act),
                                 f"[{i}] expected {exp}, got NaN")
                self.assertAlmostEqual(act, exp, delta=1e-8,
                                       msg=f"[{i}] expected {exp}, got {act}")

    def test_nan_passthrough(self) -> None:
        """NaN in any input produces NaN output."""
        dim = DirectionalIndicatorMinus(DirectionalIndicatorMinusParams(length=14))

        self.assertTrue(math.isnan(dim.update(math.nan, 1, 1)),
                        "expected NaN passthrough for NaN close")
        self.assertTrue(math.isnan(dim.update(1, math.nan, 1)),
                        "expected NaN passthrough for NaN high")
        self.assertTrue(math.isnan(dim.update(1, 1, math.nan)),
                        "expected NaN passthrough for NaN low")

        # Also test update_sample
        self.assertTrue(math.isnan(dim.update_sample(math.nan)),
                        "expected NaN passthrough for NaN sample")

    def test_update_bar(self) -> None:
        """UpdateBar produces scalar output with correct time."""
        from py.entities.bar import Bar
        from datetime import datetime

        dim = DirectionalIndicatorMinus(DirectionalIndicatorMinusParams(length=14))
        for i in range(14):
            dim.update(_CLOSE[i], _HIGH[i], _LOW[i])

        tm = datetime(2021, 4, 1)
        b = Bar(time=tm, open=0, high=_HIGH[14], low=_LOW[14],
                close=_CLOSE[14], volume=0)
        result = dim.update_bar(b)
        self.assertEqual(len(result), 1)
        self.assertEqual(result[0].time, tm)

    def test_update_scalar(self) -> None:
        """UpdateScalar produces scalar output with correct time."""
        from py.entities.scalar import Scalar
        from datetime import datetime

        dim = DirectionalIndicatorMinus(DirectionalIndicatorMinusParams(length=14))
        for i in range(14):
            dim.update(_CLOSE[i], _HIGH[i], _LOW[i])

        tm = datetime(2021, 4, 1)
        s = Scalar(time=tm, value=_HIGH[14])
        result = dim.update_scalar(s)
        self.assertEqual(len(result), 1)
        self.assertEqual(result[0].time, tm)

    def test_update_quote(self) -> None:
        """UpdateQuote produces scalar output with correct time."""
        from py.entities.quote import Quote
        from datetime import datetime

        dim = DirectionalIndicatorMinus(DirectionalIndicatorMinusParams(length=14))
        for i in range(14):
            dim.update(_CLOSE[i], _HIGH[i], _LOW[i])

        tm = datetime(2021, 4, 1)
        q = Quote(time=tm, bid_price=_HIGH[14] - 0.5,
                  ask_price=_HIGH[14] + 0.5, bid_size=0, ask_size=0)
        result = dim.update_quote(q)
        self.assertEqual(len(result), 1)
        self.assertEqual(result[0].time, tm)

    def test_update_trade(self) -> None:
        """UpdateTrade produces scalar output with correct time."""
        from py.entities.trade import Trade
        from datetime import datetime

        dim = DirectionalIndicatorMinus(DirectionalIndicatorMinusParams(length=14))
        for i in range(14):
            dim.update(_CLOSE[i], _HIGH[i], _LOW[i])

        tm = datetime(2021, 4, 1)
        r = Trade(time=tm, price=_HIGH[14], volume=0)
        result = dim.update_trade(r)
        self.assertEqual(len(result), 1)
        self.assertEqual(result[0].time, tm)

    def test_metadata(self) -> None:
        """Metadata fields are correct."""
        dim = DirectionalIndicatorMinus(DirectionalIndicatorMinusParams())
        md = dim.metadata()

        self.assertEqual(md.identifier, Identifier.DIRECTIONAL_INDICATOR_MINUS)
        self.assertEqual(md.mnemonic, "-di")
        self.assertEqual(md.description, "Directional Indicator Minus")
        self.assertEqual(len(md.outputs), 4)
        self.assertEqual(md.outputs[0].kind, int(DirectionalIndicatorMinusOutput.VALUE))
        self.assertEqual(md.outputs[0].mnemonic, "-di")
        self.assertEqual(md.outputs[0].description, "Directional Indicator Minus")
        self.assertEqual(md.outputs[1].kind, int(DirectionalIndicatorMinusOutput.DIRECTIONAL_MOVEMENT_MINUS))
        self.assertEqual(md.outputs[2].kind, int(DirectionalIndicatorMinusOutput.AVERAGE_TRUE_RANGE))
        self.assertEqual(md.outputs[3].kind, int(DirectionalIndicatorMinusOutput.TRUE_RANGE))


if __name__ == '__main__':
    unittest.main()
