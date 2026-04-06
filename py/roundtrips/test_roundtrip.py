import unittest
from datetime import datetime, timedelta

from .execution import Execution, OrderSide
from .side import RoundtripSide
from .roundtrip import Roundtrip


# ---------------------------------------------------------------------------
# Concrete test data
# ---------------------------------------------------------------------------

# Long trade: buy 100 shares at $50, sell at $55
# Commission: entry $0.01/unit, exit $0.02/unit
# Unrealized extremes during trade: high $57, low $48
_LONG_ENTRY = Execution(
    side=OrderSide.BUY,
    price=50.0,
    commission_per_unit=0.01,
    unrealized_price_high=56.0,
    unrealized_price_low=48.0,
    dt=datetime(2024, 1, 1, 9, 30, 0),
)
_LONG_EXIT = Execution(
    side=OrderSide.SELL,
    price=55.0,
    commission_per_unit=0.02,
    unrealized_price_high=57.0,
    unrealized_price_low=49.0,
    dt=datetime(2024, 1, 5, 16, 0, 0),
)
_LONG_QTY = 100.0

# Short trade: sell 200 shares at $80, buy-to-cover at $72
# Commission: entry $0.03/unit, exit $0.02/unit
# Unrealized extremes during trade: high $85, low $70
_SHORT_ENTRY = Execution(
    side=OrderSide.SELL,
    price=80.0,
    commission_per_unit=0.03,
    unrealized_price_high=85.0,
    unrealized_price_low=72.0,
    dt=datetime(2024, 2, 1, 10, 0, 0),
)
_SHORT_EXIT = Execution(
    side=OrderSide.BUY,
    price=72.0,
    commission_per_unit=0.02,
    unrealized_price_high=83.0,
    unrealized_price_low=70.0,
    dt=datetime(2024, 2, 10, 15, 30, 0),
)
_SHORT_QTY = 200.0


# ---------------------------------------------------------------------------
# Tests for a LONG round-trip
# ---------------------------------------------------------------------------

class TestRoundtripLong(unittest.TestCase):
    """Tests for a long (buy-to-open) round-trip."""

    def setUp(self):
        self.rt = Roundtrip(_LONG_ENTRY, _LONG_EXIT, _LONG_QTY)

    def test_side(self):
        self.assertEqual(self.rt.side, RoundtripSide.LONG)

    def test_quantity(self):
        self.assertAlmostEqual(self.rt.quantity, 100.0, places=13)

    def test_entry_time(self):
        self.assertEqual(self.rt.entry_time, datetime(2024, 1, 1, 9, 30, 0))

    def test_exit_time(self):
        self.assertEqual(self.rt.exit_time, datetime(2024, 1, 5, 16, 0, 0))

    def test_entry_price(self):
        self.assertAlmostEqual(self.rt.entry_price, 50.0, places=13)

    def test_exit_price(self):
        self.assertAlmostEqual(self.rt.exit_price, 55.0, places=13)

    def test_duration(self):
        expected = datetime(2024, 1, 5, 16, 0, 0) - datetime(2024, 1, 1, 9, 30, 0)
        self.assertEqual(self.rt.duration, expected)

    def test_highest_price(self):
        self.assertAlmostEqual(self.rt.highest_price, 57.0, places=13)

    def test_lowest_price(self):
        self.assertAlmostEqual(self.rt.lowest_price, 48.0, places=13)

    def test_gross_pnl(self):
        # Long: qty * (exit - entry) = 100 * (55 - 50) = 500
        self.assertAlmostEqual(self.rt.gross_pnl, 500.0, places=13)

    def test_commission(self):
        # (0.01 + 0.02) * 100 = 3.0
        self.assertAlmostEqual(self.rt.commission, 3.0, places=13)

    def test_net_pnl(self):
        # 500 - 3 = 497
        self.assertAlmostEqual(self.rt.net_pnl, 497.0, places=13)

    def test_maximum_adverse_price(self):
        # Long: lowest_p = 48
        self.assertAlmostEqual(self.rt.maximum_adverse_price, 48.0, places=13)

    def test_maximum_favorable_price(self):
        # Long: highest_p = 57
        self.assertAlmostEqual(self.rt.maximum_favorable_price, 57.0, places=13)

    def test_maximum_adverse_excursion(self):
        # Long MAE: 100 * (1 - 48/50) = 100 * 0.04 = 4.0
        self.assertAlmostEqual(self.rt.maximum_adverse_excursion, 4.0, places=13)

    def test_maximum_favorable_excursion(self):
        # Long MFE: 100 * (57/55 - 1)
        expected = 100.0 * (57.0 / 55.0 - 1.0)
        self.assertAlmostEqual(self.rt.maximum_favorable_excursion, expected, places=13)

    def test_entry_efficiency(self):
        # Long: 100 * (highest - entry) / delta = 100 * (57 - 50) / 9
        expected = 100.0 * (57.0 - 50.0) / 9.0
        self.assertAlmostEqual(self.rt.entry_efficiency, expected, places=13)

    def test_exit_efficiency(self):
        # Long: 100 * (exit - lowest) / delta = 100 * (55 - 48) / 9
        expected = 100.0 * (55.0 - 48.0) / 9.0
        self.assertAlmostEqual(self.rt.exit_efficiency, expected, places=13)

    def test_total_efficiency(self):
        # Long: 100 * (exit - entry) / delta = 100 * (55 - 50) / 9
        expected = 100.0 * (55.0 - 50.0) / 9.0
        self.assertAlmostEqual(self.rt.total_efficiency, expected, places=13)


# ---------------------------------------------------------------------------
# Tests for a SHORT round-trip
# ---------------------------------------------------------------------------

class TestRoundtripShort(unittest.TestCase):
    """Tests for a short (sell-to-open) round-trip."""

    def setUp(self):
        self.rt = Roundtrip(_SHORT_ENTRY, _SHORT_EXIT, _SHORT_QTY)

    def test_side(self):
        self.assertEqual(self.rt.side, RoundtripSide.SHORT)

    def test_quantity(self):
        self.assertAlmostEqual(self.rt.quantity, 200.0, places=13)

    def test_entry_time(self):
        self.assertEqual(self.rt.entry_time, datetime(2024, 2, 1, 10, 0, 0))

    def test_exit_time(self):
        self.assertEqual(self.rt.exit_time, datetime(2024, 2, 10, 15, 30, 0))

    def test_entry_price(self):
        self.assertAlmostEqual(self.rt.entry_price, 80.0, places=13)

    def test_exit_price(self):
        self.assertAlmostEqual(self.rt.exit_price, 72.0, places=13)

    def test_duration(self):
        expected = datetime(2024, 2, 10, 15, 30, 0) - datetime(2024, 2, 1, 10, 0, 0)
        self.assertEqual(self.rt.duration, expected)

    def test_highest_price(self):
        self.assertAlmostEqual(self.rt.highest_price, 85.0, places=13)

    def test_lowest_price(self):
        self.assertAlmostEqual(self.rt.lowest_price, 70.0, places=13)

    def test_gross_pnl(self):
        # Short: qty * (entry - exit) = 200 * (80 - 72) = 1600
        self.assertAlmostEqual(self.rt.gross_pnl, 1600.0, places=13)

    def test_commission(self):
        # (0.03 + 0.02) * 200 = 10.0
        self.assertAlmostEqual(self.rt.commission, 10.0, places=13)

    def test_net_pnl(self):
        # 1600 - 10 = 1590
        self.assertAlmostEqual(self.rt.net_pnl, 1590.0, places=13)

    def test_maximum_adverse_price(self):
        # Short: highest_p = 85
        self.assertAlmostEqual(self.rt.maximum_adverse_price, 85.0, places=13)

    def test_maximum_favorable_price(self):
        # Short: lowest_p = 70
        self.assertAlmostEqual(self.rt.maximum_favorable_price, 70.0, places=13)

    def test_maximum_adverse_excursion(self):
        # Short MAE: 100 * (85/80 - 1) = 100 * 0.0625 = 6.25
        self.assertAlmostEqual(self.rt.maximum_adverse_excursion, 6.25, places=13)

    def test_maximum_favorable_excursion(self):
        # Short MFE: 100 * (1 - 70/72)
        expected = 100.0 * (1.0 - 70.0 / 72.0)
        self.assertAlmostEqual(self.rt.maximum_favorable_excursion, expected, places=13)

    def test_entry_efficiency(self):
        # Short: 100 * (entry - lowest) / delta = 100 * (80 - 70) / 15
        expected = 100.0 * (80.0 - 70.0) / 15.0
        self.assertAlmostEqual(self.rt.entry_efficiency, expected, places=13)

    def test_exit_efficiency(self):
        # Short: 100 * (highest - exit) / delta = 100 * (85 - 72) / 15
        expected = 100.0 * (85.0 - 72.0) / 15.0
        self.assertAlmostEqual(self.rt.exit_efficiency, expected, places=13)

    def test_total_efficiency(self):
        # Short: 100 * (entry - exit) / delta = 100 * (80 - 72) / 15
        expected = 100.0 * (80.0 - 72.0) / 15.0
        self.assertAlmostEqual(self.rt.total_efficiency, expected, places=13)


# ---------------------------------------------------------------------------
# Tests for zero-delta edge case (highest == lowest)
# ---------------------------------------------------------------------------

class TestRoundtripZeroDelta(unittest.TestCase):
    """When high == low the price range (delta) is zero; efficiencies default to 0."""

    def setUp(self):
        entry = Execution(
            side=OrderSide.BUY,
            price=100.0,
            commission_per_unit=0.0,
            unrealized_price_high=100.0,
            unrealized_price_low=100.0,
            dt=datetime(2024, 3, 1, 9, 0, 0),
        )
        exit_ = Execution(
            side=OrderSide.SELL,
            price=100.0,
            commission_per_unit=0.0,
            unrealized_price_high=100.0,
            unrealized_price_low=100.0,
            dt=datetime(2024, 3, 1, 10, 0, 0),
        )
        self.rt = Roundtrip(entry, exit_, 50.0)

    def test_entry_efficiency_zero(self):
        self.assertAlmostEqual(self.rt.entry_efficiency, 0.0, places=13)

    def test_exit_efficiency_zero(self):
        self.assertAlmostEqual(self.rt.exit_efficiency, 0.0, places=13)

    def test_total_efficiency_zero(self):
        self.assertAlmostEqual(self.rt.total_efficiency, 0.0, places=13)

    def test_gross_pnl_zero(self):
        self.assertAlmostEqual(self.rt.gross_pnl, 0.0, places=13)

    def test_net_pnl_zero(self):
        self.assertAlmostEqual(self.rt.net_pnl, 0.0, places=13)


# ---------------------------------------------------------------------------
# Immutability tests
# ---------------------------------------------------------------------------

class TestRoundtripImmutability(unittest.TestCase):
    """Roundtrip instances must be frozen after construction."""

    def setUp(self):
        self.rt = Roundtrip(_LONG_ENTRY, _LONG_EXIT, _LONG_QTY)

    def test_cannot_set_existing_attribute(self):
        with self.assertRaises(TypeError):
            self.rt.gross_pnl = 999.0

    def test_cannot_set_new_attribute(self):
        with self.assertRaises(TypeError):
            self.rt.new_attr = 'anything'

    def test_cannot_modify_side(self):
        with self.assertRaises(TypeError):
            self.rt.side = RoundtripSide.SHORT


# ---------------------------------------------------------------------------
# Long losing trade (exit < entry) -- verifies negative PnL path
# ---------------------------------------------------------------------------

class TestRoundtripLongLooser(unittest.TestCase):
    """Long trade that loses money: exit_price < entry_price."""

    def setUp(self):
        entry = Execution(
            side=OrderSide.BUY,
            price=60.0,
            commission_per_unit=0.005,
            unrealized_price_high=62.0,
            unrealized_price_low=53.0,
            dt=datetime(2024, 4, 1, 9, 30, 0),
        )
        exit_ = Execution(
            side=OrderSide.SELL,
            price=54.0,
            commission_per_unit=0.005,
            unrealized_price_high=61.0,
            unrealized_price_low=52.0,
            dt=datetime(2024, 4, 3, 16, 0, 0),
        )
        self.rt = Roundtrip(entry, exit_, 150.0)

    def test_side(self):
        self.assertEqual(self.rt.side, RoundtripSide.LONG)

    def test_gross_pnl_negative(self):
        # 150 * (54 - 60) = -900
        self.assertAlmostEqual(self.rt.gross_pnl, -900.0, places=13)

    def test_commission(self):
        # (0.005 + 0.005) * 150 = 1.5
        self.assertAlmostEqual(self.rt.commission, 1.5, places=13)

    def test_net_pnl_negative(self):
        # -900 - 1.5 = -901.5
        self.assertAlmostEqual(self.rt.net_pnl, -901.5, places=13)

    def test_highest_price(self):
        self.assertAlmostEqual(self.rt.highest_price, 62.0, places=13)

    def test_lowest_price(self):
        self.assertAlmostEqual(self.rt.lowest_price, 52.0, places=13)

    def test_mae(self):
        # 100 * (1 - 52/60)
        expected = 100.0 * (1.0 - 52.0 / 60.0)
        self.assertAlmostEqual(self.rt.maximum_adverse_excursion, expected, places=13)

    def test_mfe(self):
        # 100 * (62/54 - 1)
        expected = 100.0 * (62.0 / 54.0 - 1.0)
        self.assertAlmostEqual(self.rt.maximum_favorable_excursion, expected, places=13)


# ---------------------------------------------------------------------------
# Short losing trade (exit > entry) -- verifies negative PnL path for shorts
# ---------------------------------------------------------------------------

class TestRoundtripShortLooser(unittest.TestCase):
    """Short trade that loses money: exit_price > entry_price."""

    def setUp(self):
        entry = Execution(
            side=OrderSide.SELL,
            price=40.0,
            commission_per_unit=0.01,
            unrealized_price_high=42.0,
            unrealized_price_low=39.0,
            dt=datetime(2024, 5, 1, 10, 0, 0),
        )
        exit_ = Execution(
            side=OrderSide.BUY,
            price=45.0,
            commission_per_unit=0.01,
            unrealized_price_high=46.0,
            unrealized_price_low=38.0,
            dt=datetime(2024, 5, 5, 15, 0, 0),
        )
        self.rt = Roundtrip(entry, exit_, 300.0)

    def test_side(self):
        self.assertEqual(self.rt.side, RoundtripSide.SHORT)

    def test_gross_pnl_negative(self):
        # 300 * (40 - 45) = -1500
        self.assertAlmostEqual(self.rt.gross_pnl, -1500.0, places=13)

    def test_commission(self):
        # (0.01 + 0.01) * 300 = 6.0
        self.assertAlmostEqual(self.rt.commission, 6.0, places=13)

    def test_net_pnl_negative(self):
        # -1500 - 6 = -1506
        self.assertAlmostEqual(self.rt.net_pnl, -1506.0, places=13)

    def test_maximum_adverse_price(self):
        # Short: highest = 46
        self.assertAlmostEqual(self.rt.maximum_adverse_price, 46.0, places=13)

    def test_maximum_favorable_price(self):
        # Short: lowest = 38
        self.assertAlmostEqual(self.rt.maximum_favorable_price, 38.0, places=13)

    def test_mae(self):
        # Short MAE: 100 * (46/40 - 1) = 100 * 0.15 = 15.0
        self.assertAlmostEqual(self.rt.maximum_adverse_excursion, 15.0, places=13)

    def test_mfe(self):
        # Short MFE: 100 * (1 - 38/45)
        expected = 100.0 * (1.0 - 38.0 / 45.0)
        self.assertAlmostEqual(self.rt.maximum_favorable_excursion, expected, places=13)


if __name__ == '__main__':
    unittest.main()
