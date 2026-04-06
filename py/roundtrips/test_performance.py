import unittest
from datetime import datetime

from .execution import Execution, OrderSide
from .side import RoundtripSide
from .roundtrip import Roundtrip
from .performance import RoundtripPerformance
from ..daycounting import DayCountConvention

BUY = OrderSide.BUY
SELL = OrderSide.SELL


# ---------------------------------------------------------------------------
# Helper to build executions concisely
# ---------------------------------------------------------------------------

def _exec(side, price, comm, high, low, dt):
    return Execution(side=side, price=price, commission_per_unit=comm,
                     unrealized_price_high=high, unrealized_price_low=low,
                     dt=dt)


# ---------------------------------------------------------------------------
# Shared test roundtrips (6 trades, mix of long/short, winning/losing)
# ---------------------------------------------------------------------------

# RT1: Long winner  buy 100 @ $50, sell @ $55
_RT1 = Roundtrip(
    _exec(BUY,  50.0, 0.01,  56.0, 48.0, datetime(2024, 1, 1,  9, 30)),
    _exec(SELL, 55.0, 0.02,  57.0, 49.0, datetime(2024, 1, 5,  16, 0)),
    100.0)

# RT2: Short winner  sell 200 @ $80, cover @ $72
_RT2 = Roundtrip(
    _exec(SELL, 80.0, 0.03,  85.0, 72.0, datetime(2024, 2, 1,  10, 0)),
    _exec(BUY,  72.0, 0.02,  83.0, 70.0, datetime(2024, 2, 10, 15, 30)),
    200.0)

# RT3: Long loser  buy 150 @ $60, sell @ $54
_RT3 = Roundtrip(
    _exec(BUY,  60.0, 0.005, 62.0, 53.0, datetime(2024, 3, 1,  9, 30)),
    _exec(SELL, 54.0, 0.005, 61.0, 52.0, datetime(2024, 3, 3,  16, 0)),
    150.0)

# RT4: Short loser  sell 300 @ $40, cover @ $45
_RT4 = Roundtrip(
    _exec(SELL, 40.0, 0.01,  42.0, 39.0, datetime(2024, 4, 1,  10, 0)),
    _exec(BUY,  45.0, 0.01,  46.0, 38.0, datetime(2024, 4, 5,  15, 0)),
    300.0)

# RT5: Long winner  buy 50 @ $100, sell @ $110
_RT5 = Roundtrip(
    _exec(BUY,  100.0, 0.02, 112.0, 98.0, datetime(2024, 5, 1,  9, 0)),
    _exec(SELL, 110.0, 0.02, 115.0, 99.0, datetime(2024, 5, 15, 16, 0)),
    50.0)

# RT6: Short winner  sell 100 @ $90, cover @ $82
_RT6 = Roundtrip(
    _exec(SELL, 90.0, 0.015, 92.0, 84.0, datetime(2024, 6, 1,  10, 0)),
    _exec(BUY,  82.0, 0.015, 93.0, 80.0, datetime(2024, 6, 20, 15, 0)),
    100.0)

_ALL_RTS = [_RT1, _RT2, _RT3, _RT4, _RT5, _RT6]


# ---------------------------------------------------------------------------
# Initial state
# ---------------------------------------------------------------------------

class TestRoundtripPerformanceInit(unittest.TestCase):
    """Verify default state immediately after construction."""

    def setUp(self):
        self.perf = RoundtripPerformance()

    def test_default_initial_balance(self):
        self.assertAlmostEqual(self.perf.initial_balance, 100000.0, places=13)

    def test_default_annual_risk_free_rate(self):
        self.assertAlmostEqual(self.perf.annual_risk_free_rate, 0.0, places=13)

    def test_total_count_zero(self):
        self.assertEqual(self.perf.total_count, 0)

    def test_roi_mean_none(self):
        self.assertIsNone(self.perf.roi_mean)

    def test_roi_std_none(self):
        self.assertIsNone(self.perf.roi_std)

    def test_roi_tdd_none(self):
        self.assertIsNone(self.perf.roi_tdd)

    def test_sharpe_ratio_none(self):
        self.assertIsNone(self.perf.sharpe_ratio)

    def test_sortino_ratio_none(self):
        self.assertIsNone(self.perf.sortino_ratio)

    def test_calmar_ratio_none(self):
        self.assertIsNone(self.perf.calmar_ratio)

    def test_empty_roundtrips_list(self):
        self.assertEqual(len(self.perf.roundtrips), 0)

    def test_total_gross_pnl_zero(self):
        self.assertAlmostEqual(self.perf.total_gross_pnl, 0.0, places=13)

    def test_total_net_pnl_zero(self):
        self.assertAlmostEqual(self.perf.total_net_pnl, 0.0, places=13)

    def test_max_drawdown_zero(self):
        self.assertAlmostEqual(self.perf.max_drawdown, 0.0, places=13)

    def test_average_net_pnl_zero(self):
        self.assertAlmostEqual(self.perf.average_net_pnl, 0.0, places=13)


# ---------------------------------------------------------------------------
# Reset
# ---------------------------------------------------------------------------

class TestRoundtripPerformanceReset(unittest.TestCase):
    """After adding roundtrips and calling reset(), state returns to initial."""

    def setUp(self):
        self.perf = RoundtripPerformance(
            initial_balance=100000.0,
            annual_risk_free_rate=0.0,
            day_count_convention=DayCountConvention.RAW)
        self.perf.add_roundtrip(_RT1)
        self.perf.add_roundtrip(_RT3)
        self.perf.reset()

    def test_total_count_zero_after_reset(self):
        self.assertEqual(self.perf.total_count, 0)

    def test_total_net_pnl_zero_after_reset(self):
        self.assertAlmostEqual(self.perf.total_net_pnl, 0.0, places=13)

    def test_roi_mean_none_after_reset(self):
        self.assertIsNone(self.perf.roi_mean)

    def test_roundtrips_list_empty_after_reset(self):
        self.assertEqual(len(self.perf.roundtrips), 0)

    def test_returns_on_investments_empty_after_reset(self):
        self.assertEqual(len(self.perf.returns_on_investments), 0)

    def test_max_drawdown_zero_after_reset(self):
        self.assertAlmostEqual(self.perf.max_drawdown, 0.0, places=13)


# ---------------------------------------------------------------------------
# Single long winner
# ---------------------------------------------------------------------------

class TestRoundtripPerformanceSingleLongWinner(unittest.TestCase):
    """Add one long winning roundtrip (RT1) and verify all key properties."""

    def setUp(self):
        self.perf = RoundtripPerformance(
            initial_balance=100000.0,
            annual_risk_free_rate=0.0,
            day_count_convention=DayCountConvention.RAW)
        self.perf.add_roundtrip(_RT1)

    # --- counts ---
    def test_total_count(self):
        self.assertEqual(self.perf.total_count, 1)

    def test_long_count(self):
        self.assertEqual(self.perf.long_count, 1)

    def test_short_count(self):
        self.assertEqual(self.perf.short_count, 0)

    def test_gross_winning_count(self):
        self.assertEqual(self.perf.gross_winning_count, 1)

    def test_gross_loosing_count(self):
        self.assertEqual(self.perf.gross_loosing_count, 0)

    def test_net_winning_count(self):
        self.assertEqual(self.perf.net_winning_count, 1)

    def test_net_loosing_count(self):
        self.assertEqual(self.perf.net_loosing_count, 0)

    # --- PnL ---
    def test_total_gross_pnl(self):
        self.assertAlmostEqual(self.perf.total_gross_pnl, 500.0, places=13)

    def test_total_net_pnl(self):
        self.assertAlmostEqual(self.perf.total_net_pnl, 497.0, places=13)

    def test_total_commission(self):
        self.assertAlmostEqual(self.perf.total_commission, 3.0, places=13)

    # --- ROI ---
    def test_roi_mean(self):
        # roi = 497 / (100 * 50) = 0.0994
        self.assertAlmostEqual(self.perf.roi_mean, 0.0994, places=13)

    def test_roi_std_zero(self):
        # single data point -> std = 0
        self.assertAlmostEqual(self.perf.roi_std, 0.0, places=13)

    def test_roi_tdd_none(self):
        # positive roi, no downside -> None
        self.assertIsNone(self.perf.roi_tdd)

    # --- risk-adjusted ratios ---
    def test_sharpe_ratio_none(self):
        # std = 0 -> None
        self.assertIsNone(self.perf.sharpe_ratio)

    def test_sortino_ratio_none(self):
        # no downside -> tdd is None -> None
        self.assertIsNone(self.perf.sortino_ratio)

    def test_calmar_ratio_none(self):
        # no drawdown (max_drawdown_percent = 0) -> None
        self.assertIsNone(self.perf.calmar_ratio)

    # --- drawdown ---
    def test_max_drawdown_zero(self):
        self.assertAlmostEqual(self.perf.max_drawdown, 0.0, places=13)

    # --- rate of return ---
    def test_rate_of_return(self):
        # 497 / 100000 = 0.00497
        self.assertAlmostEqual(self.perf.rate_of_return, 0.00497, places=13)

    # --- ratios ---
    def test_gross_winning_ratio(self):
        self.assertAlmostEqual(self.perf.gross_winning_ratio, 1.0, places=13)

    def test_net_winning_ratio(self):
        self.assertAlmostEqual(self.perf.net_winning_ratio, 1.0, places=13)

    # --- profit ratio ---
    def test_gross_profit_ratio_none(self):
        # no loosing trades -> denominator 0 -> None
        self.assertIsNone(self.perf.gross_profit_ratio)

    def test_net_profit_ratio_none(self):
        self.assertIsNone(self.perf.net_profit_ratio)

    # --- MAE/MFE/efficiency ---
    def test_average_mae(self):
        self.assertAlmostEqual(self.perf.average_maximum_adverse_excursion,
                               _RT1.maximum_adverse_excursion, places=13)

    def test_average_mfe(self):
        self.assertAlmostEqual(self.perf.average_maximum_favorable_excursion,
                               _RT1.maximum_favorable_excursion, places=13)

    def test_average_entry_efficiency(self):
        self.assertAlmostEqual(self.perf.average_entry_efficiency,
                               _RT1.entry_efficiency, places=13)

    def test_average_exit_efficiency(self):
        self.assertAlmostEqual(self.perf.average_exit_efficiency,
                               _RT1.exit_efficiency, places=13)

    def test_average_total_efficiency(self):
        self.assertAlmostEqual(self.perf.average_total_efficiency,
                               _RT1.total_efficiency, places=13)

    # --- duration ---
    def test_average_duration_seconds(self):
        self.assertAlmostEqual(self.perf.average_duration_seconds,
                               369000.0, places=13)

    # --- consecutive ---
    def test_max_consecutive_gross_winners(self):
        self.assertEqual(self.perf.max_consecutive_gross_winners, 1)

    def test_max_consecutive_gross_loosers(self):
        self.assertEqual(self.perf.max_consecutive_gross_loosers, 0)


# ---------------------------------------------------------------------------
# Single long loser
# ---------------------------------------------------------------------------

class TestRoundtripPerformanceSingleLooser(unittest.TestCase):
    """Add one losing roundtrip (RT3) — verifies drawdown path."""

    def setUp(self):
        self.perf = RoundtripPerformance(
            initial_balance=100000.0,
            annual_risk_free_rate=0.0,
            day_count_convention=DayCountConvention.RAW)
        self.perf.add_roundtrip(_RT3)

    def test_total_net_pnl_negative(self):
        self.assertAlmostEqual(self.perf.total_net_pnl, -901.5, places=13)

    def test_max_drawdown(self):
        self.assertAlmostEqual(self.perf.max_drawdown, 901.5, places=13)

    def test_max_drawdown_percent(self):
        # drawdown / (initial_balance + max_net_pnl)
        # max_net_pnl stays 0 (net_pnl never exceeded 0)
        # 901.5 / (100000 + 0) = 0.009015
        self.assertAlmostEqual(self.perf.max_drawdown_percent,
                               0.009015, places=13)

    def test_calmar_ratio(self):
        # roi_mean / max_drawdown_percent
        # roi_mean = -0.10016666... / 0.009015 = -11.1111...
        self.assertAlmostEqual(self.perf.calmar_ratio,
                               -11.11111111111111, places=10)

    def test_roi_mean_negative(self):
        self.assertAlmostEqual(self.perf.roi_mean,
                               -0.10016666666666667, places=13)

    def test_roi_tdd(self):
        self.assertAlmostEqual(self.perf.roi_tdd,
                               0.10016666666666667, places=13)

    def test_sortino_ratio(self):
        # (roi_mean - 0) / tdd = roi_mean / tdd = -1.0
        self.assertAlmostEqual(self.perf.sortino_ratio, -1.0, places=13)

    def test_gross_loosing_count(self):
        self.assertEqual(self.perf.gross_loosing_count, 1)

    def test_net_loosing_count(self):
        self.assertEqual(self.perf.net_loosing_count, 1)


# ---------------------------------------------------------------------------
# Multiple mixed roundtrips (all 6)
# ---------------------------------------------------------------------------

class TestRoundtripPerformanceMultipleMixed(unittest.TestCase):
    """Add all 6 roundtrips and verify comprehensive properties.

    All expected values were computed by running the actual code.
    """

    def setUp(self):
        self.perf = RoundtripPerformance(
            initial_balance=100000.0,
            annual_risk_free_rate=0.0,
            day_count_convention=DayCountConvention.RAW)
        for rt in _ALL_RTS:
            self.perf.add_roundtrip(rt)

    # ====================== counts ======================

    def test_total_count(self):
        self.assertEqual(self.perf.total_count, 6)

    def test_long_count(self):
        self.assertEqual(self.perf.long_count, 3)

    def test_short_count(self):
        self.assertEqual(self.perf.short_count, 3)

    def test_gross_winning_count(self):
        self.assertEqual(self.perf.gross_winning_count, 4)

    def test_gross_loosing_count(self):
        self.assertEqual(self.perf.gross_loosing_count, 2)

    def test_net_winning_count(self):
        self.assertEqual(self.perf.net_winning_count, 4)

    def test_net_loosing_count(self):
        self.assertEqual(self.perf.net_loosing_count, 2)

    def test_gross_long_winning_count(self):
        self.assertEqual(self.perf.gross_long_winning_count, 2)

    def test_gross_long_loosing_count(self):
        self.assertEqual(self.perf.gross_long_loosing_count, 1)

    def test_net_long_winning_count(self):
        self.assertEqual(self.perf.net_long_winning_count, 2)

    def test_net_long_loosing_count(self):
        self.assertEqual(self.perf.net_long_loosing_count, 1)

    def test_gross_short_winning_count(self):
        self.assertEqual(self.perf.gross_short_winning_count, 2)

    def test_gross_short_loosing_count(self):
        self.assertEqual(self.perf.gross_short_loosing_count, 1)

    def test_net_short_winning_count(self):
        self.assertEqual(self.perf.net_short_winning_count, 2)

    def test_net_short_loosing_count(self):
        self.assertEqual(self.perf.net_short_loosing_count, 1)

    # ====================== PnL totals ======================

    def test_total_gross_pnl(self):
        self.assertAlmostEqual(self.perf.total_gross_pnl, 1000.0, places=13)

    def test_total_net_pnl(self):
        self.assertAlmostEqual(self.perf.total_net_pnl, 974.5, places=13)

    def test_winning_gross_pnl(self):
        self.assertAlmostEqual(self.perf.winning_gross_pnl, 3400.0, places=13)

    def test_loosing_gross_pnl(self):
        self.assertAlmostEqual(self.perf.loosing_gross_pnl, -2400.0, places=13)

    def test_winning_net_pnl(self):
        self.assertAlmostEqual(self.perf.winning_net_pnl, 3382.0, places=13)

    def test_loosing_net_pnl(self):
        self.assertAlmostEqual(self.perf.loosing_net_pnl, -2407.5, places=13)

    def test_winning_gross_long_pnl(self):
        self.assertAlmostEqual(self.perf.winning_gross_long_pnl,
                               1000.0, places=13)

    def test_loosing_gross_long_pnl(self):
        self.assertAlmostEqual(self.perf.loosing_gross_long_pnl,
                               -900.0, places=13)

    def test_winning_gross_short_pnl(self):
        self.assertAlmostEqual(self.perf.winning_gross_short_pnl,
                               2400.0, places=13)

    def test_loosing_gross_short_pnl(self):
        self.assertAlmostEqual(self.perf.loosing_gross_short_pnl,
                               -1500.0, places=13)

    # ====================== commission ======================

    def test_total_commission(self):
        self.assertAlmostEqual(self.perf.total_commission, 25.5, places=13)

    def test_gross_winning_commission(self):
        self.assertAlmostEqual(self.perf.gross_winning_commission,
                               18.0, places=13)

    def test_gross_loosing_commission(self):
        self.assertAlmostEqual(self.perf.gross_loosing_commission,
                               7.5, places=13)

    def test_net_winning_commission(self):
        self.assertAlmostEqual(self.perf.net_winning_commission,
                               18.0, places=13)

    def test_net_loosing_commission(self):
        self.assertAlmostEqual(self.perf.net_loosing_commission,
                               7.5, places=13)

    # ====================== average PnL ======================

    def test_average_gross_pnl(self):
        self.assertAlmostEqual(self.perf.average_gross_pnl,
                               1000.0 / 6.0, places=13)

    def test_average_net_pnl(self):
        self.assertAlmostEqual(self.perf.average_net_pnl,
                               974.5 / 6.0, places=13)

    def test_average_winning_gross_pnl(self):
        self.assertAlmostEqual(self.perf.average_winning_gross_pnl,
                               3400.0 / 4.0, places=13)

    def test_average_loosing_gross_pnl(self):
        self.assertAlmostEqual(self.perf.average_loosing_gross_pnl,
                               -2400.0 / 2.0, places=13)

    def test_average_winning_net_pnl(self):
        self.assertAlmostEqual(self.perf.average_winning_net_pnl,
                               3382.0 / 4.0, places=13)

    def test_average_loosing_net_pnl(self):
        self.assertAlmostEqual(self.perf.average_loosing_net_pnl,
                               -2407.5 / 2.0, places=13)

    def test_average_gross_long_pnl(self):
        # (500 - 900 + 500) / 3 = 100/3
        self.assertAlmostEqual(self.perf.average_gross_long_pnl,
                               100.0 / 3.0, places=13)

    def test_average_gross_short_pnl(self):
        # (1600 - 1500 + 800) / 3 = 900/3 = 300
        self.assertAlmostEqual(self.perf.average_gross_short_pnl,
                               300.0, places=13)

    # ====================== win/loss ratios ======================

    def test_gross_winning_ratio(self):
        self.assertAlmostEqual(self.perf.gross_winning_ratio,
                               4.0 / 6.0, places=13)

    def test_gross_loosing_ratio(self):
        self.assertAlmostEqual(self.perf.gross_loosing_ratio,
                               2.0 / 6.0, places=13)

    def test_net_winning_ratio(self):
        self.assertAlmostEqual(self.perf.net_winning_ratio,
                               4.0 / 6.0, places=13)

    def test_net_loosing_ratio(self):
        self.assertAlmostEqual(self.perf.net_loosing_ratio,
                               2.0 / 6.0, places=13)

    def test_gross_long_winning_ratio(self):
        self.assertAlmostEqual(self.perf.gross_long_winning_ratio,
                               2.0 / 3.0, places=13)

    def test_gross_short_winning_ratio(self):
        self.assertAlmostEqual(self.perf.gross_short_winning_ratio,
                               2.0 / 3.0, places=13)

    # ====================== profit ratios ======================

    def test_gross_profit_ratio(self):
        # |3400 / -2400| = 1.41666...
        self.assertAlmostEqual(self.perf.gross_profit_ratio,
                               1.4166666666666667, places=13)

    def test_net_profit_ratio(self):
        # |3382 / -2407.5| = 1.40477...
        self.assertAlmostEqual(self.perf.net_profit_ratio,
                               1.4047767393561785, places=13)

    def test_gross_profit_long_ratio(self):
        self.assertAlmostEqual(self.perf.gross_profit_long_ratio,
                               1.1111111111111112, places=13)

    def test_gross_profit_short_ratio(self):
        self.assertAlmostEqual(self.perf.gross_profit_short_ratio,
                               1.6, places=13)

    # ====================== profit PnL ratio ======================

    def test_gross_profit_pnl_ratio(self):
        # 3400 / 1000 = 3.4
        self.assertAlmostEqual(self.perf.gross_profit_pnl_ratio,
                               3.4, places=13)

    def test_net_profit_pnl_ratio(self):
        # 3382 / 974.5
        self.assertAlmostEqual(self.perf.net_profit_pnl_ratio,
                               3382.0 / 974.5, places=13)

    # ====================== average win/loss ratio ======================

    def test_average_gross_winning_loosing_ratio(self):
        # avg_winning_gross / avg_loosing_gross = 850 / -1200
        self.assertAlmostEqual(self.perf.average_gross_winning_loosing_ratio,
                               850.0 / -1200.0, places=13)

    def test_average_net_winning_loosing_ratio(self):
        # 845.5 / -1203.75
        self.assertAlmostEqual(self.perf.average_net_winning_loosing_ratio,
                               845.5 / -1203.75, places=13)

    # ====================== ROI statistics ======================

    def test_roi_mean(self):
        self.assertAlmostEqual(self.perf.roi_mean,
                               0.026877314814814812, places=13)

    def test_roi_std(self):
        self.assertAlmostEqual(self.perf.roi_std,
                               0.0991356544050762, places=13)

    def test_roi_tdd(self):
        self.assertAlmostEqual(self.perf.roi_tdd,
                               0.11354208715518468, places=13)

    def test_roiann_mean(self):
        self.assertAlmostEqual(self.perf.roiann_mean,
                               -1.7233887909446202, places=12)

    def test_roiann_std(self):
        self.assertAlmostEqual(self.perf.roiann_std,
                               8.73138705463156, places=12)

    def test_roiann_tdd(self):
        self.assertAlmostEqual(self.perf.roiann_tdd,
                               13.751365296707874, places=12)

    # ====================== risk-adjusted ratios ======================

    def test_sharpe_ratio(self):
        self.assertAlmostEqual(self.perf.sharpe_ratio,
                               0.27111653194916085, places=13)

    def test_sharpe_ratio_annual(self):
        self.assertAlmostEqual(self.perf.sharpe_ratio_annual,
                               -0.1973785814512082, places=12)

    def test_sortino_ratio(self):
        self.assertAlmostEqual(self.perf.sortino_ratio,
                               0.23671675841293985, places=13)

    def test_sortino_ratio_annual(self):
        self.assertAlmostEqual(self.perf.sortino_ratio_annual,
                               -0.1253249225629404, places=12)

    def test_calmar_ratio(self):
        self.assertAlmostEqual(self.perf.calmar_ratio,
                               1.139698624091381, places=12)

    def test_calmar_ratio_annual(self):
        self.assertAlmostEqual(self.perf.calmar_ratio_annual,
                               -73.07812731097131, places=10)

    # ====================== rate of return ======================

    def test_rate_of_return(self):
        self.assertAlmostEqual(self.perf.rate_of_return,
                               0.009745, places=13)

    def test_rate_of_return_annual(self):
        self.assertAlmostEqual(self.perf.rate_of_return_annual,
                               0.020786693247353695, places=12)

    def test_recovery_factor(self):
        self.assertAlmostEqual(self.perf.recovery_factor,
                               0.8814335009522727, places=12)

    # ====================== drawdown ======================

    def test_max_net_pnl(self):
        self.assertAlmostEqual(self.perf.max_net_pnl, 2087.0, places=13)

    def test_max_drawdown(self):
        self.assertAlmostEqual(self.perf.max_drawdown, 2407.5, places=13)

    def test_max_drawdown_percent(self):
        # 2407.5 / (100000 + 2087) = 0.023582826...
        self.assertAlmostEqual(self.perf.max_drawdown_percent,
                               2407.5 / (100000.0 + 2087.0), places=13)

    # ====================== duration ======================

    def test_average_duration_seconds(self):
        self.assertAlmostEqual(self.perf.average_duration_seconds,
                               770100.0, places=13)

    def test_average_long_duration_seconds(self):
        self.assertAlmostEqual(self.perf.average_long_duration_seconds,
                               600000.0, places=13)

    def test_average_short_duration_seconds(self):
        self.assertAlmostEqual(self.perf.average_short_duration_seconds,
                               940200.0, places=13)

    def test_average_gross_winning_duration_seconds(self):
        self.assertAlmostEqual(
            self.perf.average_gross_winning_duration_seconds,
            1015200.0, places=13)

    def test_average_gross_loosing_duration_seconds(self):
        self.assertAlmostEqual(
            self.perf.average_gross_loosing_duration_seconds,
            279900.0, places=13)

    def test_minimum_duration_seconds(self):
        self.assertAlmostEqual(self.perf.minimum_duration_seconds,
                               196200.0, places=13)

    def test_maximum_duration_seconds(self):
        self.assertAlmostEqual(self.perf.maximum_duration_seconds,
                               1659600.0, places=13)

    def test_minimum_long_duration_seconds(self):
        self.assertAlmostEqual(self.perf.minimum_long_duration_seconds,
                               196200.0, places=13)

    def test_maximum_long_duration_seconds(self):
        self.assertAlmostEqual(self.perf.maximum_long_duration_seconds,
                               1234800.0, places=13)

    def test_minimum_short_duration_seconds(self):
        self.assertAlmostEqual(self.perf.minimum_short_duration_seconds,
                               363600.0, places=13)

    def test_maximum_short_duration_seconds(self):
        self.assertAlmostEqual(self.perf.maximum_short_duration_seconds,
                               1659600.0, places=13)

    # ====================== MAE / MFE / efficiency ======================

    def test_average_mae(self):
        expected = sum(r.maximum_adverse_excursion for r in _ALL_RTS) / 6.0
        self.assertAlmostEqual(self.perf.average_maximum_adverse_excursion,
                               expected, places=13)

    def test_average_mfe(self):
        expected = sum(r.maximum_favorable_excursion for r in _ALL_RTS) / 6.0
        self.assertAlmostEqual(self.perf.average_maximum_favorable_excursion,
                               expected, places=13)

    def test_average_entry_efficiency(self):
        expected = sum(r.entry_efficiency for r in _ALL_RTS) / 6.0
        self.assertAlmostEqual(self.perf.average_entry_efficiency,
                               expected, places=13)

    def test_average_exit_efficiency(self):
        expected = sum(r.exit_efficiency for r in _ALL_RTS) / 6.0
        self.assertAlmostEqual(self.perf.average_exit_efficiency,
                               expected, places=13)

    def test_average_total_efficiency(self):
        expected = sum(r.total_efficiency for r in _ALL_RTS) / 6.0
        self.assertAlmostEqual(self.perf.average_total_efficiency,
                               expected, places=13)

    # ====================== consecutive ======================

    def test_max_consecutive_gross_winners(self):
        # RT1(W), RT2(W), RT3(L), RT4(L), RT5(W), RT6(W) -> max streak = 2
        self.assertEqual(self.perf.max_consecutive_gross_winners, 2)

    def test_max_consecutive_gross_loosers(self):
        self.assertEqual(self.perf.max_consecutive_gross_loosers, 2)

    def test_max_consecutive_net_winners(self):
        self.assertEqual(self.perf.max_consecutive_net_winners, 2)

    def test_max_consecutive_net_loosers(self):
        self.assertEqual(self.perf.max_consecutive_net_loosers, 2)

    # ====================== time tracking ======================

    def test_first_time(self):
        self.assertEqual(self.perf.first_time, datetime(2024, 1, 1, 9, 30))

    def test_last_time(self):
        self.assertEqual(self.perf.last_time, datetime(2024, 6, 20, 15, 0))


# ---------------------------------------------------------------------------
# Edge cases
# ---------------------------------------------------------------------------

class TestRoundtripPerformanceEdgeCases(unittest.TestCase):
    """Edge cases: zero balance, empty state, ratio boundaries."""

    def test_zero_initial_balance_rate_of_return_none(self):
        perf = RoundtripPerformance(
            initial_balance=0.0,
            day_count_convention=DayCountConvention.RAW)
        self.assertIsNone(perf.rate_of_return)

    def test_no_roundtrips_average_gross_pnl_zero(self):
        perf = RoundtripPerformance()
        self.assertAlmostEqual(perf.average_gross_pnl, 0.0, places=13)

    def test_no_roundtrips_average_net_pnl_zero(self):
        perf = RoundtripPerformance()
        self.assertAlmostEqual(perf.average_net_pnl, 0.0, places=13)

    def test_no_roundtrips_gross_winning_ratio_zero(self):
        perf = RoundtripPerformance()
        self.assertAlmostEqual(perf.gross_winning_ratio, 0.0, places=13)

    def test_no_roundtrips_average_duration_zero(self):
        perf = RoundtripPerformance()
        self.assertAlmostEqual(perf.average_duration_seconds, 0.0, places=13)

    def test_sharpe_none_single_point(self):
        """Single roundtrip -> std=0 -> sharpe is None."""
        perf = RoundtripPerformance(
            initial_balance=100000.0,
            day_count_convention=DayCountConvention.RAW)
        perf.add_roundtrip(_RT1)
        self.assertIsNone(perf.sharpe_ratio)

    def test_rate_of_return_annual_none_when_zero_duration(self):
        """No roundtrips -> duration annualized = 0 -> None."""
        perf = RoundtripPerformance(
            initial_balance=100000.0,
            day_count_convention=DayCountConvention.RAW)
        self.assertIsNone(perf.rate_of_return_annual)

    def test_recovery_factor_none_no_drawdown(self):
        """Single winning trade -> no drawdown -> recovery_factor None."""
        perf = RoundtripPerformance(
            initial_balance=100000.0,
            day_count_convention=DayCountConvention.RAW)
        perf.add_roundtrip(_RT1)
        self.assertIsNone(perf.recovery_factor)


# ---------------------------------------------------------------------------
# Incremental update — verify ROI list grows correctly
# ---------------------------------------------------------------------------

class TestRoundtripPerformanceIncremental(unittest.TestCase):
    """Verify that returns_on_investments list is built incrementally."""

    def test_roi_list_length(self):
        perf = RoundtripPerformance(
            initial_balance=100000.0,
            day_count_convention=DayCountConvention.RAW)
        for i, rt in enumerate(_ALL_RTS, 1):
            perf.add_roundtrip(rt)
            self.assertEqual(len(perf.returns_on_investments), i)

    def test_roi_values(self):
        """Each ROI = net_pnl / (quantity * entry_price)."""
        expected_rois = [
            0.0994,                 # 497 / (100*50)
            0.099375,               # 1590 / (200*80)
            -0.10016666666666667,   # -901.5 / (150*60)
            -0.1255,                # -1506 / (300*40)
            0.0996,                 # 498 / (50*100)
            0.08855555555555556,    # 797 / (100*90)
        ]
        perf = RoundtripPerformance(
            initial_balance=100000.0,
            day_count_convention=DayCountConvention.RAW)
        for i, rt in enumerate(_ALL_RTS):
            perf.add_roundtrip(rt)
        for i in range(len(_ALL_RTS)):
            self.assertAlmostEqual(perf.returns_on_investments[i],
                                   expected_rois[i], places=13)

    def test_sortino_downside_count(self):
        """Only negative (roi - rfr) values go into sortino_downside_returns."""
        perf = RoundtripPerformance(
            initial_balance=100000.0,
            annual_risk_free_rate=0.0,
            day_count_convention=DayCountConvention.RAW)
        for rt in _ALL_RTS:
            perf.add_roundtrip(rt)
        # RT3 and RT4 have negative ROI -> 2 downside entries
        self.assertEqual(len(perf.sortino_downside_returns), 2)


if __name__ == '__main__':
    unittest.main()
