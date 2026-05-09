"""Tests for Igor Livshin's Balance of Power indicator."""

import math
import unittest
from datetime import datetime

from .balance_of_power import BalanceOfPower
from .params import BalanceOfPowerParams
from .output import BalanceOfPowerOutput
from ...core.identifier import Identifier
from ....entities.bar import Bar
from ....entities.scalar import Scalar

from .test_testdata import (
    _OPEN,
    _HIGH,
    _LOW,
    _CLOSE,
    _EXPECTED,
)


# fmt: off
# fmt: on


def _round_to(v: float, digits: int) -> float:
    p = 10.0 ** digits
    return round(v * p) / p


class TestBalanceOfPower(unittest.TestCase):
    """Tests for the BalanceOfPower indicator."""

    def test_update_ohlc(self) -> None:
        """Test OHLC update with 20 bars of TA-Lib reference data."""
        digits = 9
        bop = BalanceOfPower(BalanceOfPowerParams())

        for i in range(len(_OPEN)):
            act = bop.update_ohlc(_OPEN[i], _HIGH[i], _LOW[i], _CLOSE[i])
            self.assertFalse(math.isnan(act), f"[{i}] expected non-NaN, got NaN")
            self.assertTrue(bop.is_primed(), f"[{i}] expected primed")

            got = _round_to(act, digits)
            exp = _round_to(_EXPECTED[i], digits)
            self.assertEqual(got, exp, f"[{i}] expected {exp}, got {got}")

    def test_is_primed(self) -> None:
        """BOP is always primed."""
        bop = BalanceOfPower(BalanceOfPowerParams())
        self.assertTrue(bop.is_primed(), "expected primed initially")

        bop.update_ohlc(92.5, 93.25, 90.75, 91.5)
        self.assertTrue(bop.is_primed(), "expected still primed after update")

    def test_nan_passthrough(self) -> None:
        """NaN in any OHLC input produces NaN output."""
        bop = BalanceOfPower(BalanceOfPowerParams())

        self.assertTrue(math.isnan(bop.update(math.nan)))
        self.assertTrue(math.isnan(bop.update_ohlc(math.nan, 1.0, 2.0, 3.0)))
        self.assertTrue(math.isnan(bop.update_ohlc(1.0, math.nan, 2.0, 3.0)))
        self.assertTrue(math.isnan(bop.update_ohlc(1.0, 2.0, math.nan, 3.0)))
        self.assertTrue(math.isnan(bop.update_ohlc(1.0, 2.0, 3.0, math.nan)))

    def test_zero_range(self) -> None:
        """When H == L, range < epsilon, result is 0."""
        bop = BalanceOfPower(BalanceOfPowerParams())
        v = bop.update_ohlc(0.001, 0.001, 0.001, 0.001)
        self.assertEqual(v, 0.0)

    def test_scalar_always_zero(self) -> None:
        """Scalar update uses same value for O/H/L/C, so BOP=0."""
        bop = BalanceOfPower(BalanceOfPowerParams())
        self.assertEqual(bop.update(50.0), 0.0)
        self.assertEqual(bop.update(100.0), 0.0)

    def test_metadata(self) -> None:
        """Metadata fields are correct."""
        bop = BalanceOfPower(BalanceOfPowerParams())
        md = bop.metadata()

        self.assertEqual(md.identifier, Identifier.BALANCE_OF_POWER)
        self.assertEqual(md.mnemonic, "bop")
        self.assertEqual(md.description, "Balance of Power")
        self.assertEqual(len(md.outputs), 1)
        self.assertEqual(md.outputs[0].kind, int(BalanceOfPowerOutput.VALUE))

    def test_update_bar(self) -> None:
        """Test update_bar extracts OHLC from Bar."""
        digits = 9
        bop = BalanceOfPower(BalanceOfPowerParams())
        tm = datetime(2021, 4, 1)

        for i in range(len(_OPEN)):
            bar = Bar(time=tm, open=_OPEN[i], high=_HIGH[i], low=_LOW[i],
                      close=_CLOSE[i], volume=0.0)
            out = bop.update_bar(bar)
            v = out[0].value

            got = _round_to(v, digits)
            exp = _round_to(_EXPECTED[i], digits)
            self.assertEqual(got, exp, f"[{i}] expected {exp}, got {got}")


if __name__ == '__main__':
    unittest.main()
