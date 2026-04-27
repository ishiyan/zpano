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


# fmt: off
_OPEN = [
    92.500, 91.500, 95.155, 93.970, 95.500, 94.500, 95.000, 91.500, 91.815, 91.125,
    93.875, 97.500, 98.815, 92.000, 91.125, 91.875, 93.405, 89.750, 89.345, 92.250,
]

_HIGH = [
    93.250000, 94.940000, 96.375000, 96.190000, 96.000000, 94.720000, 95.000000, 93.720000, 92.470000, 92.750000,
    96.250000, 99.625000, 99.125000, 92.750000, 91.315000, 93.250000, 93.405000, 90.655000, 91.970000, 92.250000,
]

_LOW = [
    90.750000, 91.405000, 94.250000, 93.500000, 92.815000, 93.500000, 92.000000, 89.750000, 89.440000, 90.625000,
    92.750000, 96.315000, 96.030000, 88.815000, 86.750000, 90.940000, 88.905000, 88.780000, 89.250000, 89.750000,
]

_CLOSE = [
    91.500000, 94.815000, 94.375000, 95.095000, 93.780000, 94.625000, 92.530000, 92.750000, 90.315000, 92.470000,
    96.125000, 97.250000, 98.500000, 89.875000, 91.000000, 92.815000, 89.155000, 89.345000, 91.625000, 89.875000,
]

_EXPECTED = [
    -0.400000000000000, 0.937765205091938, -0.367058823529412, 0.418215613382900, -0.540031397174254,
    0.102459016393443, -0.823333333333333, 0.314861460957179, -0.495049504950495, 0.632941176470588,
    0.642857142857143, -0.075528700906344, -0.101777059773828, -0.540025412960610, -0.027382256297919,
    0.406926406926406, -0.944444444444444, -0.216000000000001, 0.838235294117648, -0.950000000000000,
]
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
