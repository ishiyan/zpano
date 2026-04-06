package roundtrips

import (
	"math"
	"testing"
	"time"

	"portf_py/daycounting/conventions"
)

// ---------------------------------------------------------------------------
// Helper to build executions concisely
// ---------------------------------------------------------------------------

func exec(side OrderSide, price, comm, high, low float64, dt time.Time) Execution {
	return Execution{
		Side:                side,
		Price:               price,
		CommissionPerUnit:   comm,
		UnrealizedPriceHigh: high,
		UnrealizedPriceLow:  low,
		DateTime:            dt,
	}
}

// ---------------------------------------------------------------------------
// Shared test roundtrips (6 trades, mix of long/short, winning/losing)
// ---------------------------------------------------------------------------

// RT1: Long winner buy 100 @ $50, sell @ $55
var rt1 = NewRoundtrip(
	exec(Buy, 50.0, 0.01, 56.0, 48.0, time.Date(2024, 1, 1, 9, 30, 0, 0, time.UTC)),
	exec(Sell, 55.0, 0.02, 57.0, 49.0, time.Date(2024, 1, 5, 16, 0, 0, 0, time.UTC)),
	100.0)

// RT2: Short winner sell 200 @ $80, cover @ $72
var rt2 = NewRoundtrip(
	exec(Sell, 80.0, 0.03, 85.0, 72.0, time.Date(2024, 2, 1, 10, 0, 0, 0, time.UTC)),
	exec(Buy, 72.0, 0.02, 83.0, 70.0, time.Date(2024, 2, 10, 15, 30, 0, 0, time.UTC)),
	200.0)

// RT3: Long loser buy 150 @ $60, sell @ $54
var rt3 = NewRoundtrip(
	exec(Buy, 60.0, 0.005, 62.0, 53.0, time.Date(2024, 3, 1, 9, 30, 0, 0, time.UTC)),
	exec(Sell, 54.0, 0.005, 61.0, 52.0, time.Date(2024, 3, 3, 16, 0, 0, 0, time.UTC)),
	150.0)

// RT4: Short loser sell 300 @ $40, cover @ $45
var rt4 = NewRoundtrip(
	exec(Sell, 40.0, 0.01, 42.0, 39.0, time.Date(2024, 4, 1, 10, 0, 0, 0, time.UTC)),
	exec(Buy, 45.0, 0.01, 46.0, 38.0, time.Date(2024, 4, 5, 15, 0, 0, 0, time.UTC)),
	300.0)

// RT5: Long winner buy 50 @ $100, sell @ $110
var rt5 = NewRoundtrip(
	exec(Buy, 100.0, 0.02, 112.0, 98.0, time.Date(2024, 5, 1, 9, 0, 0, 0, time.UTC)),
	exec(Sell, 110.0, 0.02, 115.0, 99.0, time.Date(2024, 5, 15, 16, 0, 0, 0, time.UTC)),
	50.0)

// RT6: Short winner sell 100 @ $90, cover @ $82
var rt6 = NewRoundtrip(
	exec(Sell, 90.0, 0.015, 92.0, 84.0, time.Date(2024, 6, 1, 10, 0, 0, 0, time.UTC)),
	exec(Buy, 82.0, 0.015, 93.0, 80.0, time.Date(2024, 6, 20, 15, 0, 0, 0, time.UTC)),
	100.0)

var allRTs = []Roundtrip{rt1, rt2, rt3, rt4, rt5, rt6}

// ---------------------------------------------------------------------------
// Helper for nil-checks on *float64
// ---------------------------------------------------------------------------

func assertNil(t *testing.T, name string, v *float64) {
	t.Helper()
	if v != nil {
		t.Errorf("%s: expected nil, got %v", name, *v)
	}
}

func assertAlmost(t *testing.T, name string, v *float64, expected float64, eps float64) {
	t.Helper()
	if v == nil {
		t.Errorf("%s: expected %v, got nil", name, expected)
		return
	}
	if !almostEqual(*v, expected, eps) {
		t.Errorf("%s: expected %v, got %v", name, expected, *v)
	}
}

// ---------------------------------------------------------------------------
// Initial state
// ---------------------------------------------------------------------------

func TestRoundtripPerformanceInit(t *testing.T) {
	perf := NewRoundtripPerformance(100000.0, 0.0, 0.0, conventions.RAW)

	t.Run("default_initial_balance", func(t *testing.T) {
		if !almostEqual(perf.InitialBalance, 100000.0, epsilon) {
			t.Errorf("expected 100000.0, got %v", perf.InitialBalance)
		}
	})
	t.Run("default_annual_risk_free_rate", func(t *testing.T) {
		if !almostEqual(perf.AnnualRiskFreeRate, 0.0, epsilon) {
			t.Errorf("expected 0.0, got %v", perf.AnnualRiskFreeRate)
		}
	})
	t.Run("total_count_zero", func(t *testing.T) {
		if perf.TotalCount() != 0 {
			t.Errorf("expected 0, got %v", perf.TotalCount())
		}
	})
	t.Run("roi_mean_none", func(t *testing.T) {
		assertNil(t, "RoiMean", perf.RoiMean())
	})
	t.Run("roi_std_none", func(t *testing.T) {
		assertNil(t, "RoiStd", perf.RoiStd())
	})
	t.Run("roi_tdd_none", func(t *testing.T) {
		assertNil(t, "RoiTdd", perf.RoiTdd())
	})
	t.Run("sharpe_ratio_none", func(t *testing.T) {
		assertNil(t, "SharpeRatio", perf.SharpeRatio())
	})
	t.Run("sortino_ratio_none", func(t *testing.T) {
		assertNil(t, "SortinoRatio", perf.SortinoRatio())
	})
	t.Run("calmar_ratio_none", func(t *testing.T) {
		assertNil(t, "CalmarRatio", perf.CalmarRatio())
	})
	t.Run("empty_roundtrips_list", func(t *testing.T) {
		if len(perf.Roundtrips) != 0 {
			t.Errorf("expected 0, got %v", len(perf.Roundtrips))
		}
	})
	t.Run("total_gross_pnl_zero", func(t *testing.T) {
		if !almostEqual(perf.TotalGrossPnl(), 0.0, epsilon) {
			t.Errorf("expected 0.0, got %v", perf.TotalGrossPnl())
		}
	})
	t.Run("total_net_pnl_zero", func(t *testing.T) {
		if !almostEqual(perf.TotalNetPnl(), 0.0, epsilon) {
			t.Errorf("expected 0.0, got %v", perf.TotalNetPnl())
		}
	})
	t.Run("max_drawdown_zero", func(t *testing.T) {
		if !almostEqual(perf.MaxDrawdown, 0.0, epsilon) {
			t.Errorf("expected 0.0, got %v", perf.MaxDrawdown)
		}
	})
	t.Run("average_net_pnl_zero", func(t *testing.T) {
		if !almostEqual(perf.AverageNetPnl(), 0.0, epsilon) {
			t.Errorf("expected 0.0, got %v", perf.AverageNetPnl())
		}
	})
}

// ---------------------------------------------------------------------------
// Reset
// ---------------------------------------------------------------------------

func TestRoundtripPerformanceReset(t *testing.T) {
	perf := NewRoundtripPerformance(100000.0, 0.0, 0.0, conventions.RAW)
	perf.AddRoundtrip(rt1)
	perf.AddRoundtrip(rt3)
	perf.Reset()

	t.Run("total_count_zero_after_reset", func(t *testing.T) {
		if perf.TotalCount() != 0 {
			t.Errorf("expected 0, got %v", perf.TotalCount())
		}
	})
	t.Run("total_net_pnl_zero_after_reset", func(t *testing.T) {
		if !almostEqual(perf.TotalNetPnl(), 0.0, epsilon) {
			t.Errorf("expected 0.0, got %v", perf.TotalNetPnl())
		}
	})
	t.Run("roi_mean_none_after_reset", func(t *testing.T) {
		assertNil(t, "RoiMean", perf.RoiMean())
	})
	t.Run("roundtrips_list_empty_after_reset", func(t *testing.T) {
		if len(perf.Roundtrips) != 0 {
			t.Errorf("expected 0, got %v", len(perf.Roundtrips))
		}
	})
	t.Run("returns_on_investments_empty_after_reset", func(t *testing.T) {
		if len(perf.ReturnsOnInvestments) != 0 {
			t.Errorf("expected 0, got %v", len(perf.ReturnsOnInvestments))
		}
	})
	t.Run("max_drawdown_zero_after_reset", func(t *testing.T) {
		if !almostEqual(perf.MaxDrawdown, 0.0, epsilon) {
			t.Errorf("expected 0.0, got %v", perf.MaxDrawdown)
		}
	})
}

// ---------------------------------------------------------------------------
// Single long winner
// ---------------------------------------------------------------------------

func TestRoundtripPerformanceSingleLongWinner(t *testing.T) {
	perf := NewRoundtripPerformance(100000.0, 0.0, 0.0, conventions.RAW)
	perf.AddRoundtrip(rt1)

	// --- counts ---
	t.Run("total_count", func(t *testing.T) {
		if perf.TotalCount() != 1 {
			t.Errorf("expected 1, got %v", perf.TotalCount())
		}
	})
	t.Run("long_count", func(t *testing.T) {
		if perf.LongCount() != 1 {
			t.Errorf("expected 1, got %v", perf.LongCount())
		}
	})
	t.Run("short_count", func(t *testing.T) {
		if perf.ShortCount() != 0 {
			t.Errorf("expected 0, got %v", perf.ShortCount())
		}
	})
	t.Run("gross_winning_count", func(t *testing.T) {
		if perf.GrossWinningCount() != 1 {
			t.Errorf("expected 1, got %v", perf.GrossWinningCount())
		}
	})
	t.Run("gross_loosing_count", func(t *testing.T) {
		if perf.GrossLoosingCount() != 0 {
			t.Errorf("expected 0, got %v", perf.GrossLoosingCount())
		}
	})
	t.Run("net_winning_count", func(t *testing.T) {
		if perf.NetWinningCount() != 1 {
			t.Errorf("expected 1, got %v", perf.NetWinningCount())
		}
	})
	t.Run("net_loosing_count", func(t *testing.T) {
		if perf.NetLoosingCount() != 0 {
			t.Errorf("expected 0, got %v", perf.NetLoosingCount())
		}
	})

	// --- PnL ---
	t.Run("total_gross_pnl", func(t *testing.T) {
		if !almostEqual(perf.TotalGrossPnl(), 500.0, epsilon) {
			t.Errorf("expected 500.0, got %v", perf.TotalGrossPnl())
		}
	})
	t.Run("total_net_pnl", func(t *testing.T) {
		if !almostEqual(perf.TotalNetPnl(), 497.0, epsilon) {
			t.Errorf("expected 497.0, got %v", perf.TotalNetPnl())
		}
	})
	t.Run("total_commission", func(t *testing.T) {
		if !almostEqual(perf.TotalCommission, 3.0, epsilon) {
			t.Errorf("expected 3.0, got %v", perf.TotalCommission)
		}
	})

	// --- ROI ---
	t.Run("roi_mean", func(t *testing.T) {
		// roi = 497 / (100 * 50) = 0.0994
		assertAlmost(t, "RoiMean", perf.RoiMean(), 0.0994, epsilon)
	})
	t.Run("roi_std_zero", func(t *testing.T) {
		// single data point -> std = 0
		assertAlmost(t, "RoiStd", perf.RoiStd(), 0.0, epsilon)
	})
	t.Run("roi_tdd_none", func(t *testing.T) {
		// positive roi, no downside -> None
		assertNil(t, "RoiTdd", perf.RoiTdd())
	})

	// --- risk-adjusted ratios ---
	t.Run("sharpe_ratio_none", func(t *testing.T) {
		// std = 0 -> None
		assertNil(t, "SharpeRatio", perf.SharpeRatio())
	})
	t.Run("sortino_ratio_none", func(t *testing.T) {
		assertNil(t, "SortinoRatio", perf.SortinoRatio())
	})
	t.Run("calmar_ratio_none", func(t *testing.T) {
		assertNil(t, "CalmarRatio", perf.CalmarRatio())
	})

	// --- drawdown ---
	t.Run("max_drawdown_zero", func(t *testing.T) {
		if !almostEqual(perf.MaxDrawdown, 0.0, epsilon) {
			t.Errorf("expected 0.0, got %v", perf.MaxDrawdown)
		}
	})

	// --- rate of return ---
	t.Run("rate_of_return", func(t *testing.T) {
		// 497 / 100000 = 0.00497
		assertAlmost(t, "RateOfReturn", perf.RateOfReturn(), 0.00497, epsilon)
	})

	// --- ratios ---
	t.Run("gross_winning_ratio", func(t *testing.T) {
		if !almostEqual(perf.GrossWinningRatio(), 1.0, epsilon) {
			t.Errorf("expected 1.0, got %v", perf.GrossWinningRatio())
		}
	})
	t.Run("net_winning_ratio", func(t *testing.T) {
		if !almostEqual(perf.NetWinningRatio(), 1.0, epsilon) {
			t.Errorf("expected 1.0, got %v", perf.NetWinningRatio())
		}
	})

	// --- profit ratio ---
	t.Run("gross_profit_ratio_none", func(t *testing.T) {
		assertNil(t, "GrossProfitRatio", perf.GrossProfitRatio())
	})
	t.Run("net_profit_ratio_none", func(t *testing.T) {
		assertNil(t, "NetProfitRatio", perf.NetProfitRatio())
	})

	// --- MAE/MFE/efficiency ---
	t.Run("average_mae", func(t *testing.T) {
		if !almostEqual(perf.AverageMaximumAdverseExcursion(), rt1.MaximumAdverseExcursion, epsilon) {
			t.Errorf("expected %v, got %v", rt1.MaximumAdverseExcursion, perf.AverageMaximumAdverseExcursion())
		}
	})
	t.Run("average_mfe", func(t *testing.T) {
		if !almostEqual(perf.AverageMaximumFavorableExcursion(), rt1.MaximumFavorableExcursion, epsilon) {
			t.Errorf("expected %v, got %v", rt1.MaximumFavorableExcursion, perf.AverageMaximumFavorableExcursion())
		}
	})
	t.Run("average_entry_efficiency", func(t *testing.T) {
		if !almostEqual(perf.AverageEntryEfficiency(), rt1.EntryEfficiency, epsilon) {
			t.Errorf("expected %v, got %v", rt1.EntryEfficiency, perf.AverageEntryEfficiency())
		}
	})
	t.Run("average_exit_efficiency", func(t *testing.T) {
		if !almostEqual(perf.AverageExitEfficiency(), rt1.ExitEfficiency, epsilon) {
			t.Errorf("expected %v, got %v", rt1.ExitEfficiency, perf.AverageExitEfficiency())
		}
	})
	t.Run("average_total_efficiency", func(t *testing.T) {
		if !almostEqual(perf.AverageTotalEfficiency(), rt1.TotalEfficiency, epsilon) {
			t.Errorf("expected %v, got %v", rt1.TotalEfficiency, perf.AverageTotalEfficiency())
		}
	})

	// --- duration ---
	t.Run("average_duration_seconds", func(t *testing.T) {
		if !almostEqual(perf.AverageDurationSeconds(), 369000.0, epsilon) {
			t.Errorf("expected 369000.0, got %v", perf.AverageDurationSeconds())
		}
	})

	// --- consecutive ---
	t.Run("max_consecutive_gross_winners", func(t *testing.T) {
		if perf.MaxConsecutiveGrossWinners() != 1 {
			t.Errorf("expected 1, got %v", perf.MaxConsecutiveGrossWinners())
		}
	})
	t.Run("max_consecutive_gross_loosers", func(t *testing.T) {
		if perf.MaxConsecutiveGrossLoosers() != 0 {
			t.Errorf("expected 0, got %v", perf.MaxConsecutiveGrossLoosers())
		}
	})
}

// ---------------------------------------------------------------------------
// Single long loser
// ---------------------------------------------------------------------------

func TestRoundtripPerformanceSingleLooser(t *testing.T) {
	perf := NewRoundtripPerformance(100000.0, 0.0, 0.0, conventions.RAW)
	perf.AddRoundtrip(rt3)

	t.Run("total_net_pnl_negative", func(t *testing.T) {
		if !almostEqual(perf.TotalNetPnl(), -901.5, epsilon) {
			t.Errorf("expected -901.5, got %v", perf.TotalNetPnl())
		}
	})
	t.Run("max_drawdown", func(t *testing.T) {
		if !almostEqual(perf.MaxDrawdown, 901.5, epsilon) {
			t.Errorf("expected 901.5, got %v", perf.MaxDrawdown)
		}
	})
	t.Run("max_drawdown_percent", func(t *testing.T) {
		// 901.5 / (100000 + 0) = 0.009015
		if !almostEqual(perf.MaxDrawdownPercent, 0.009015, epsilon) {
			t.Errorf("expected 0.009015, got %v", perf.MaxDrawdownPercent)
		}
	})
	t.Run("calmar_ratio", func(t *testing.T) {
		// roi_mean / max_drawdown_percent = -0.10016666... / 0.009015
		assertAlmost(t, "CalmarRatio", perf.CalmarRatio(), -11.11111111111111, 1e-10)
	})
	t.Run("roi_mean_negative", func(t *testing.T) {
		assertAlmost(t, "RoiMean", perf.RoiMean(), -0.10016666666666667, epsilon)
	})
	t.Run("roi_tdd", func(t *testing.T) {
		assertAlmost(t, "RoiTdd", perf.RoiTdd(), 0.10016666666666667, epsilon)
	})
	t.Run("sortino_ratio", func(t *testing.T) {
		// (roi_mean - 0) / tdd = -1.0
		assertAlmost(t, "SortinoRatio", perf.SortinoRatio(), -1.0, epsilon)
	})
	t.Run("gross_loosing_count", func(t *testing.T) {
		if perf.GrossLoosingCount() != 1 {
			t.Errorf("expected 1, got %v", perf.GrossLoosingCount())
		}
	})
	t.Run("net_loosing_count", func(t *testing.T) {
		if perf.NetLoosingCount() != 1 {
			t.Errorf("expected 1, got %v", perf.NetLoosingCount())
		}
	})
}

// ---------------------------------------------------------------------------
// Multiple mixed roundtrips (all 6)
// ---------------------------------------------------------------------------

func TestRoundtripPerformanceMultipleMixed(t *testing.T) {
	perf := NewRoundtripPerformance(100000.0, 0.0, 0.0, conventions.RAW)
	for _, rt := range allRTs {
		perf.AddRoundtrip(rt)
	}

	// ====================== counts ======================

	t.Run("total_count", func(t *testing.T) {
		if perf.TotalCount() != 6 {
			t.Errorf("expected 6, got %v", perf.TotalCount())
		}
	})
	t.Run("long_count", func(t *testing.T) {
		if perf.LongCount() != 3 {
			t.Errorf("expected 3, got %v", perf.LongCount())
		}
	})
	t.Run("short_count", func(t *testing.T) {
		if perf.ShortCount() != 3 {
			t.Errorf("expected 3, got %v", perf.ShortCount())
		}
	})
	t.Run("gross_winning_count", func(t *testing.T) {
		if perf.GrossWinningCount() != 4 {
			t.Errorf("expected 4, got %v", perf.GrossWinningCount())
		}
	})
	t.Run("gross_loosing_count", func(t *testing.T) {
		if perf.GrossLoosingCount() != 2 {
			t.Errorf("expected 2, got %v", perf.GrossLoosingCount())
		}
	})
	t.Run("net_winning_count", func(t *testing.T) {
		if perf.NetWinningCount() != 4 {
			t.Errorf("expected 4, got %v", perf.NetWinningCount())
		}
	})
	t.Run("net_loosing_count", func(t *testing.T) {
		if perf.NetLoosingCount() != 2 {
			t.Errorf("expected 2, got %v", perf.NetLoosingCount())
		}
	})
	t.Run("gross_long_winning_count", func(t *testing.T) {
		if perf.GrossLongWinningCount() != 2 {
			t.Errorf("expected 2, got %v", perf.GrossLongWinningCount())
		}
	})
	t.Run("gross_long_loosing_count", func(t *testing.T) {
		if perf.GrossLongLoosingCount() != 1 {
			t.Errorf("expected 1, got %v", perf.GrossLongLoosingCount())
		}
	})
	t.Run("net_long_winning_count", func(t *testing.T) {
		if perf.NetLongWinningCount() != 2 {
			t.Errorf("expected 2, got %v", perf.NetLongWinningCount())
		}
	})
	t.Run("net_long_loosing_count", func(t *testing.T) {
		if perf.NetLongLoosingCount() != 1 {
			t.Errorf("expected 1, got %v", perf.NetLongLoosingCount())
		}
	})
	t.Run("gross_short_winning_count", func(t *testing.T) {
		if perf.GrossShortWinningCount() != 2 {
			t.Errorf("expected 2, got %v", perf.GrossShortWinningCount())
		}
	})
	t.Run("gross_short_loosing_count", func(t *testing.T) {
		if perf.GrossShortLoosingCount() != 1 {
			t.Errorf("expected 1, got %v", perf.GrossShortLoosingCount())
		}
	})
	t.Run("net_short_winning_count", func(t *testing.T) {
		if perf.NetShortWinningCount() != 2 {
			t.Errorf("expected 2, got %v", perf.NetShortWinningCount())
		}
	})
	t.Run("net_short_loosing_count", func(t *testing.T) {
		if perf.NetShortLoosingCount() != 1 {
			t.Errorf("expected 1, got %v", perf.NetShortLoosingCount())
		}
	})

	// ====================== PnL totals ======================

	t.Run("total_gross_pnl", func(t *testing.T) {
		if !almostEqual(perf.TotalGrossPnl(), 1000.0, epsilon) {
			t.Errorf("expected 1000.0, got %v", perf.TotalGrossPnl())
		}
	})
	t.Run("total_net_pnl", func(t *testing.T) {
		if !almostEqual(perf.TotalNetPnl(), 974.5, epsilon) {
			t.Errorf("expected 974.5, got %v", perf.TotalNetPnl())
		}
	})
	t.Run("winning_gross_pnl", func(t *testing.T) {
		if !almostEqual(perf.WinningGrossPnl(), 3400.0, epsilon) {
			t.Errorf("expected 3400.0, got %v", perf.WinningGrossPnl())
		}
	})
	t.Run("loosing_gross_pnl", func(t *testing.T) {
		if !almostEqual(perf.LoosingGrossPnl(), -2400.0, epsilon) {
			t.Errorf("expected -2400.0, got %v", perf.LoosingGrossPnl())
		}
	})
	t.Run("winning_net_pnl", func(t *testing.T) {
		if !almostEqual(perf.WinningNetPnl(), 3382.0, epsilon) {
			t.Errorf("expected 3382.0, got %v", perf.WinningNetPnl())
		}
	})
	t.Run("loosing_net_pnl", func(t *testing.T) {
		if !almostEqual(perf.LoosingNetPnl(), -2407.5, epsilon) {
			t.Errorf("expected -2407.5, got %v", perf.LoosingNetPnl())
		}
	})
	t.Run("winning_gross_long_pnl", func(t *testing.T) {
		if !almostEqual(perf.WinningGrossLongPnl(), 1000.0, epsilon) {
			t.Errorf("expected 1000.0, got %v", perf.WinningGrossLongPnl())
		}
	})
	t.Run("loosing_gross_long_pnl", func(t *testing.T) {
		if !almostEqual(perf.LoosingGrossLongPnl(), -900.0, epsilon) {
			t.Errorf("expected -900.0, got %v", perf.LoosingGrossLongPnl())
		}
	})
	t.Run("winning_gross_short_pnl", func(t *testing.T) {
		if !almostEqual(perf.WinningGrossShortPnl(), 2400.0, epsilon) {
			t.Errorf("expected 2400.0, got %v", perf.WinningGrossShortPnl())
		}
	})
	t.Run("loosing_gross_short_pnl", func(t *testing.T) {
		if !almostEqual(perf.LoosingGrossShortPnl(), -1500.0, epsilon) {
			t.Errorf("expected -1500.0, got %v", perf.LoosingGrossShortPnl())
		}
	})

	// ====================== commission ======================

	t.Run("total_commission", func(t *testing.T) {
		if !almostEqual(perf.TotalCommission, 25.5, epsilon) {
			t.Errorf("expected 25.5, got %v", perf.TotalCommission)
		}
	})
	t.Run("gross_winning_commission", func(t *testing.T) {
		if !almostEqual(perf.GrossWinningCommission, 18.0, epsilon) {
			t.Errorf("expected 18.0, got %v", perf.GrossWinningCommission)
		}
	})
	t.Run("gross_loosing_commission", func(t *testing.T) {
		if !almostEqual(perf.GrossLoosingCommission, 7.5, epsilon) {
			t.Errorf("expected 7.5, got %v", perf.GrossLoosingCommission)
		}
	})
	t.Run("net_winning_commission", func(t *testing.T) {
		if !almostEqual(perf.NetWinningCommission, 18.0, epsilon) {
			t.Errorf("expected 18.0, got %v", perf.NetWinningCommission)
		}
	})
	t.Run("net_loosing_commission", func(t *testing.T) {
		if !almostEqual(perf.NetLoosingCommission, 7.5, epsilon) {
			t.Errorf("expected 7.5, got %v", perf.NetLoosingCommission)
		}
	})

	// ====================== average PnL ======================

	t.Run("average_gross_pnl", func(t *testing.T) {
		if !almostEqual(perf.AverageGrossPnl(), 1000.0/6.0, epsilon) {
			t.Errorf("expected %v, got %v", 1000.0/6.0, perf.AverageGrossPnl())
		}
	})
	t.Run("average_net_pnl", func(t *testing.T) {
		if !almostEqual(perf.AverageNetPnl(), 974.5/6.0, epsilon) {
			t.Errorf("expected %v, got %v", 974.5/6.0, perf.AverageNetPnl())
		}
	})
	t.Run("average_winning_gross_pnl", func(t *testing.T) {
		if !almostEqual(perf.AverageWinningGrossPnl(), 3400.0/4.0, epsilon) {
			t.Errorf("expected %v, got %v", 3400.0/4.0, perf.AverageWinningGrossPnl())
		}
	})
	t.Run("average_loosing_gross_pnl", func(t *testing.T) {
		if !almostEqual(perf.AverageLoosingGrossPnl(), -2400.0/2.0, epsilon) {
			t.Errorf("expected %v, got %v", -2400.0/2.0, perf.AverageLoosingGrossPnl())
		}
	})
	t.Run("average_winning_net_pnl", func(t *testing.T) {
		if !almostEqual(perf.AverageWinningNetPnl(), 3382.0/4.0, epsilon) {
			t.Errorf("expected %v, got %v", 3382.0/4.0, perf.AverageWinningNetPnl())
		}
	})
	t.Run("average_loosing_net_pnl", func(t *testing.T) {
		if !almostEqual(perf.AverageLoosingNetPnl(), -2407.5/2.0, epsilon) {
			t.Errorf("expected %v, got %v", -2407.5/2.0, perf.AverageLoosingNetPnl())
		}
	})
	t.Run("average_gross_long_pnl", func(t *testing.T) {
		// (500 - 900 + 500) / 3 = 100/3
		if !almostEqual(perf.AverageGrossLongPnl(), 100.0/3.0, epsilon) {
			t.Errorf("expected %v, got %v", 100.0/3.0, perf.AverageGrossLongPnl())
		}
	})
	t.Run("average_gross_short_pnl", func(t *testing.T) {
		// (1600 - 1500 + 800) / 3 = 300
		if !almostEqual(perf.AverageGrossShortPnl(), 300.0, epsilon) {
			t.Errorf("expected 300.0, got %v", perf.AverageGrossShortPnl())
		}
	})

	// ====================== win/loss ratios ======================

	t.Run("gross_winning_ratio", func(t *testing.T) {
		if !almostEqual(perf.GrossWinningRatio(), 4.0/6.0, epsilon) {
			t.Errorf("expected %v, got %v", 4.0/6.0, perf.GrossWinningRatio())
		}
	})
	t.Run("gross_loosing_ratio", func(t *testing.T) {
		if !almostEqual(perf.GrossLoosingRatio(), 2.0/6.0, epsilon) {
			t.Errorf("expected %v, got %v", 2.0/6.0, perf.GrossLoosingRatio())
		}
	})
	t.Run("net_winning_ratio", func(t *testing.T) {
		if !almostEqual(perf.NetWinningRatio(), 4.0/6.0, epsilon) {
			t.Errorf("expected %v, got %v", 4.0/6.0, perf.NetWinningRatio())
		}
	})
	t.Run("net_loosing_ratio", func(t *testing.T) {
		if !almostEqual(perf.NetLoosingRatio(), 2.0/6.0, epsilon) {
			t.Errorf("expected %v, got %v", 2.0/6.0, perf.NetLoosingRatio())
		}
	})
	t.Run("gross_long_winning_ratio", func(t *testing.T) {
		if !almostEqual(perf.GrossLongWinningRatio(), 2.0/3.0, epsilon) {
			t.Errorf("expected %v, got %v", 2.0/3.0, perf.GrossLongWinningRatio())
		}
	})
	t.Run("gross_short_winning_ratio", func(t *testing.T) {
		if !almostEqual(perf.GrossShortWinningRatio(), 2.0/3.0, epsilon) {
			t.Errorf("expected %v, got %v", 2.0/3.0, perf.GrossShortWinningRatio())
		}
	})

	// ====================== profit ratios ======================

	t.Run("gross_profit_ratio", func(t *testing.T) {
		assertAlmost(t, "GrossProfitRatio", perf.GrossProfitRatio(), 1.4166666666666667, epsilon)
	})
	t.Run("net_profit_ratio", func(t *testing.T) {
		assertAlmost(t, "NetProfitRatio", perf.NetProfitRatio(), 1.4047767393561785, epsilon)
	})
	t.Run("gross_profit_long_ratio", func(t *testing.T) {
		assertAlmost(t, "GrossProfitLongRatio", perf.GrossProfitLongRatio(), 1.1111111111111112, epsilon)
	})
	t.Run("gross_profit_short_ratio", func(t *testing.T) {
		assertAlmost(t, "GrossProfitShortRatio", perf.GrossProfitShortRatio(), 1.6, epsilon)
	})

	// ====================== profit PnL ratio ======================

	t.Run("gross_profit_pnl_ratio", func(t *testing.T) {
		if !almostEqual(perf.GrossProfitPnlRatio(), 3.4, epsilon) {
			t.Errorf("expected 3.4, got %v", perf.GrossProfitPnlRatio())
		}
	})
	t.Run("net_profit_pnl_ratio", func(t *testing.T) {
		expected := 3382.0 / 974.5
		if !almostEqual(perf.NetProfitPnlRatio(), expected, epsilon) {
			t.Errorf("expected %v, got %v", expected, perf.NetProfitPnlRatio())
		}
	})

	// ====================== average win/loss ratio ======================

	t.Run("average_gross_winning_loosing_ratio", func(t *testing.T) {
		expected := 850.0 / -1200.0
		if !almostEqual(perf.AverageGrossWinningLoosingRatio(), expected, epsilon) {
			t.Errorf("expected %v, got %v", expected, perf.AverageGrossWinningLoosingRatio())
		}
	})
	t.Run("average_net_winning_loosing_ratio", func(t *testing.T) {
		expected := 845.5 / -1203.75
		if !almostEqual(perf.AverageNetWinningLoosingRatio(), expected, epsilon) {
			t.Errorf("expected %v, got %v", expected, perf.AverageNetWinningLoosingRatio())
		}
	})

	// ====================== ROI statistics ======================

	t.Run("roi_mean", func(t *testing.T) {
		assertAlmost(t, "RoiMean", perf.RoiMean(), 0.026877314814814812, epsilon)
	})
	t.Run("roi_std", func(t *testing.T) {
		assertAlmost(t, "RoiStd", perf.RoiStd(), 0.0991356544050762, epsilon)
	})
	t.Run("roi_tdd", func(t *testing.T) {
		assertAlmost(t, "RoiTdd", perf.RoiTdd(), 0.11354208715518468, epsilon)
	})
	t.Run("roiann_mean", func(t *testing.T) {
		assertAlmost(t, "RoiannMean", perf.RoiannMean(), -1.7233887909446202, 1e-12)
	})
	t.Run("roiann_std", func(t *testing.T) {
		assertAlmost(t, "RoiannStd", perf.RoiannStd(), 8.73138705463156, 1e-12)
	})
	t.Run("roiann_tdd", func(t *testing.T) {
		assertAlmost(t, "RoiannTdd", perf.RoiannTdd(), 13.751365296707874, 1e-12)
	})

	// ====================== risk-adjusted ratios ======================

	t.Run("sharpe_ratio", func(t *testing.T) {
		assertAlmost(t, "SharpeRatio", perf.SharpeRatio(), 0.27111653194916085, epsilon)
	})
	t.Run("sharpe_ratio_annual", func(t *testing.T) {
		assertAlmost(t, "SharpeRatioAnnual", perf.SharpeRatioAnnual(), -0.1973785814512082, 1e-12)
	})
	t.Run("sortino_ratio", func(t *testing.T) {
		assertAlmost(t, "SortinoRatio", perf.SortinoRatio(), 0.23671675841293985, epsilon)
	})
	t.Run("sortino_ratio_annual", func(t *testing.T) {
		assertAlmost(t, "SortinoRatioAnnual", perf.SortinoRatioAnnual(), -0.1253249225629404, 1e-12)
	})
	t.Run("calmar_ratio", func(t *testing.T) {
		assertAlmost(t, "CalmarRatio", perf.CalmarRatio(), 1.139698624091381, 1e-12)
	})
	t.Run("calmar_ratio_annual", func(t *testing.T) {
		assertAlmost(t, "CalmarRatioAnnual", perf.CalmarRatioAnnual(), -73.07812731097131, 1e-10)
	})

	// ====================== rate of return ======================

	t.Run("rate_of_return", func(t *testing.T) {
		assertAlmost(t, "RateOfReturn", perf.RateOfReturn(), 0.009745, epsilon)
	})
	t.Run("rate_of_return_annual", func(t *testing.T) {
		assertAlmost(t, "RateOfReturnAnnual", perf.RateOfReturnAnnual(), 0.020786693247353695, 1e-12)
	})
	t.Run("recovery_factor", func(t *testing.T) {
		assertAlmost(t, "RecoveryFactor", perf.RecoveryFactor(), 0.8814335009522727, 1e-12)
	})

	// ====================== drawdown ======================

	t.Run("max_net_pnl", func(t *testing.T) {
		if !almostEqual(perf.MaxNetPnl, 2087.0, epsilon) {
			t.Errorf("expected 2087.0, got %v", perf.MaxNetPnl)
		}
	})
	t.Run("max_drawdown", func(t *testing.T) {
		if !almostEqual(perf.MaxDrawdown, 2407.5, epsilon) {
			t.Errorf("expected 2407.5, got %v", perf.MaxDrawdown)
		}
	})
	t.Run("max_drawdown_percent", func(t *testing.T) {
		expected := 2407.5 / (100000.0 + 2087.0)
		if !almostEqual(perf.MaxDrawdownPercent, expected, epsilon) {
			t.Errorf("expected %v, got %v", expected, perf.MaxDrawdownPercent)
		}
	})

	// ====================== duration ======================

	t.Run("average_duration_seconds", func(t *testing.T) {
		if !almostEqual(perf.AverageDurationSeconds(), 770100.0, epsilon) {
			t.Errorf("expected 770100.0, got %v", perf.AverageDurationSeconds())
		}
	})
	t.Run("average_long_duration_seconds", func(t *testing.T) {
		if !almostEqual(perf.AverageLongDurationSeconds(), 600000.0, epsilon) {
			t.Errorf("expected 600000.0, got %v", perf.AverageLongDurationSeconds())
		}
	})
	t.Run("average_short_duration_seconds", func(t *testing.T) {
		if !almostEqual(perf.AverageShortDurationSeconds(), 940200.0, epsilon) {
			t.Errorf("expected 940200.0, got %v", perf.AverageShortDurationSeconds())
		}
	})
	t.Run("average_gross_winning_duration_seconds", func(t *testing.T) {
		if !almostEqual(perf.AverageGrossWinningDurationSeconds(), 1015200.0, epsilon) {
			t.Errorf("expected 1015200.0, got %v", perf.AverageGrossWinningDurationSeconds())
		}
	})
	t.Run("average_gross_loosing_duration_seconds", func(t *testing.T) {
		if !almostEqual(perf.AverageGrossLoosingDurationSeconds(), 279900.0, epsilon) {
			t.Errorf("expected 279900.0, got %v", perf.AverageGrossLoosingDurationSeconds())
		}
	})
	t.Run("minimum_duration_seconds", func(t *testing.T) {
		if !almostEqual(perf.MinimumDurationSeconds(), 196200.0, epsilon) {
			t.Errorf("expected 196200.0, got %v", perf.MinimumDurationSeconds())
		}
	})
	t.Run("maximum_duration_seconds", func(t *testing.T) {
		if !almostEqual(perf.MaximumDurationSeconds(), 1659600.0, epsilon) {
			t.Errorf("expected 1659600.0, got %v", perf.MaximumDurationSeconds())
		}
	})
	t.Run("minimum_long_duration_seconds", func(t *testing.T) {
		if !almostEqual(perf.MinimumLongDurationSeconds(), 196200.0, epsilon) {
			t.Errorf("expected 196200.0, got %v", perf.MinimumLongDurationSeconds())
		}
	})
	t.Run("maximum_long_duration_seconds", func(t *testing.T) {
		if !almostEqual(perf.MaximumLongDurationSeconds(), 1234800.0, epsilon) {
			t.Errorf("expected 1234800.0, got %v", perf.MaximumLongDurationSeconds())
		}
	})
	t.Run("minimum_short_duration_seconds", func(t *testing.T) {
		if !almostEqual(perf.MinimumShortDurationSeconds(), 363600.0, epsilon) {
			t.Errorf("expected 363600.0, got %v", perf.MinimumShortDurationSeconds())
		}
	})
	t.Run("maximum_short_duration_seconds", func(t *testing.T) {
		if !almostEqual(perf.MaximumShortDurationSeconds(), 1659600.0, epsilon) {
			t.Errorf("expected 1659600.0, got %v", perf.MaximumShortDurationSeconds())
		}
	})

	// ====================== MAE / MFE / efficiency ======================

	t.Run("average_mae", func(t *testing.T) {
		sum := 0.0
		for _, r := range allRTs {
			sum += r.MaximumAdverseExcursion
		}
		expected := sum / 6.0
		if !almostEqual(perf.AverageMaximumAdverseExcursion(), expected, epsilon) {
			t.Errorf("expected %v, got %v", expected, perf.AverageMaximumAdverseExcursion())
		}
	})
	t.Run("average_mfe", func(t *testing.T) {
		sum := 0.0
		for _, r := range allRTs {
			sum += r.MaximumFavorableExcursion
		}
		expected := sum / 6.0
		if !almostEqual(perf.AverageMaximumFavorableExcursion(), expected, epsilon) {
			t.Errorf("expected %v, got %v", expected, perf.AverageMaximumFavorableExcursion())
		}
	})
	t.Run("average_entry_efficiency", func(t *testing.T) {
		sum := 0.0
		for _, r := range allRTs {
			sum += r.EntryEfficiency
		}
		expected := sum / 6.0
		if !almostEqual(perf.AverageEntryEfficiency(), expected, epsilon) {
			t.Errorf("expected %v, got %v", expected, perf.AverageEntryEfficiency())
		}
	})
	t.Run("average_exit_efficiency", func(t *testing.T) {
		sum := 0.0
		for _, r := range allRTs {
			sum += r.ExitEfficiency
		}
		expected := sum / 6.0
		if !almostEqual(perf.AverageExitEfficiency(), expected, epsilon) {
			t.Errorf("expected %v, got %v", expected, perf.AverageExitEfficiency())
		}
	})
	t.Run("average_total_efficiency", func(t *testing.T) {
		sum := 0.0
		for _, r := range allRTs {
			sum += r.TotalEfficiency
		}
		expected := sum / 6.0
		if !almostEqual(perf.AverageTotalEfficiency(), expected, epsilon) {
			t.Errorf("expected %v, got %v", expected, perf.AverageTotalEfficiency())
		}
	})

	// ====================== consecutive ======================

	t.Run("max_consecutive_gross_winners", func(t *testing.T) {
		// RT1(W), RT2(W), RT3(L), RT4(L), RT5(W), RT6(W) -> max streak = 2
		if perf.MaxConsecutiveGrossWinners() != 2 {
			t.Errorf("expected 2, got %v", perf.MaxConsecutiveGrossWinners())
		}
	})
	t.Run("max_consecutive_gross_loosers", func(t *testing.T) {
		if perf.MaxConsecutiveGrossLoosers() != 2 {
			t.Errorf("expected 2, got %v", perf.MaxConsecutiveGrossLoosers())
		}
	})
	t.Run("max_consecutive_net_winners", func(t *testing.T) {
		if perf.MaxConsecutiveNetWinners() != 2 {
			t.Errorf("expected 2, got %v", perf.MaxConsecutiveNetWinners())
		}
	})
	t.Run("max_consecutive_net_loosers", func(t *testing.T) {
		if perf.MaxConsecutiveNetLoosers() != 2 {
			t.Errorf("expected 2, got %v", perf.MaxConsecutiveNetLoosers())
		}
	})

	// ====================== time tracking ======================

	t.Run("first_time", func(t *testing.T) {
		expected := time.Date(2024, 1, 1, 9, 30, 0, 0, time.UTC)
		if perf.FirstTime == nil || !perf.FirstTime.Equal(expected) {
			t.Errorf("expected %v, got %v", expected, perf.FirstTime)
		}
	})
	t.Run("last_time", func(t *testing.T) {
		expected := time.Date(2024, 6, 20, 15, 0, 0, 0, time.UTC)
		if perf.LastTime == nil || !perf.LastTime.Equal(expected) {
			t.Errorf("expected %v, got %v", expected, perf.LastTime)
		}
	})
}

// ---------------------------------------------------------------------------
// Edge cases
// ---------------------------------------------------------------------------

func TestRoundtripPerformanceEdgeCases(t *testing.T) {
	t.Run("zero_initial_balance_rate_of_return_none", func(t *testing.T) {
		perf := NewRoundtripPerformance(0.0, 0.0, 0.0, conventions.RAW)
		assertNil(t, "RateOfReturn", perf.RateOfReturn())
	})
	t.Run("no_roundtrips_average_gross_pnl_zero", func(t *testing.T) {
		perf := NewRoundtripPerformance(100000.0, 0.0, 0.0, conventions.RAW)
		if !almostEqual(perf.AverageGrossPnl(), 0.0, epsilon) {
			t.Errorf("expected 0.0, got %v", perf.AverageGrossPnl())
		}
	})
	t.Run("no_roundtrips_average_net_pnl_zero", func(t *testing.T) {
		perf := NewRoundtripPerformance(100000.0, 0.0, 0.0, conventions.RAW)
		if !almostEqual(perf.AverageNetPnl(), 0.0, epsilon) {
			t.Errorf("expected 0.0, got %v", perf.AverageNetPnl())
		}
	})
	t.Run("no_roundtrips_gross_winning_ratio_zero", func(t *testing.T) {
		perf := NewRoundtripPerformance(100000.0, 0.0, 0.0, conventions.RAW)
		if !almostEqual(perf.GrossWinningRatio(), 0.0, epsilon) {
			t.Errorf("expected 0.0, got %v", perf.GrossWinningRatio())
		}
	})
	t.Run("no_roundtrips_average_duration_zero", func(t *testing.T) {
		perf := NewRoundtripPerformance(100000.0, 0.0, 0.0, conventions.RAW)
		if !almostEqual(perf.AverageDurationSeconds(), 0.0, epsilon) {
			t.Errorf("expected 0.0, got %v", perf.AverageDurationSeconds())
		}
	})
	t.Run("sharpe_none_single_point", func(t *testing.T) {
		perf := NewRoundtripPerformance(100000.0, 0.0, 0.0, conventions.RAW)
		perf.AddRoundtrip(rt1)
		assertNil(t, "SharpeRatio", perf.SharpeRatio())
	})
	t.Run("rate_of_return_annual_none_when_zero_duration", func(t *testing.T) {
		perf := NewRoundtripPerformance(100000.0, 0.0, 0.0, conventions.RAW)
		assertNil(t, "RateOfReturnAnnual", perf.RateOfReturnAnnual())
	})
	t.Run("recovery_factor_none_no_drawdown", func(t *testing.T) {
		perf := NewRoundtripPerformance(100000.0, 0.0, 0.0, conventions.RAW)
		perf.AddRoundtrip(rt1)
		assertNil(t, "RecoveryFactor", perf.RecoveryFactor())
	})
}

// ---------------------------------------------------------------------------
// Incremental update — verify ROI list grows correctly
// ---------------------------------------------------------------------------

func TestRoundtripPerformanceIncremental(t *testing.T) {
	t.Run("roi_list_length", func(t *testing.T) {
		perf := NewRoundtripPerformance(100000.0, 0.0, 0.0, conventions.RAW)
		for i, rt := range allRTs {
			perf.AddRoundtrip(rt)
			if len(perf.ReturnsOnInvestments) != i+1 {
				t.Errorf("after %d roundtrips: expected length %d, got %d", i+1, i+1, len(perf.ReturnsOnInvestments))
			}
		}
	})
	t.Run("roi_values", func(t *testing.T) {
		expectedROIs := []float64{
			0.0994,               // 497 / (100*50)
			0.099375,             // 1590 / (200*80)
			-0.10016666666666667, // -901.5 / (150*60)
			-0.1255,              // -1506 / (300*40)
			0.0996,               // 498 / (50*100)
			0.08855555555555556,  // 797 / (100*90)
		}
		perf := NewRoundtripPerformance(100000.0, 0.0, 0.0, conventions.RAW)
		for _, rt := range allRTs {
			perf.AddRoundtrip(rt)
		}
		for i, expected := range expectedROIs {
			if !almostEqual(perf.ReturnsOnInvestments[i], expected, epsilon) {
				t.Errorf("ROI[%d]: expected %v, got %v", i, expected, perf.ReturnsOnInvestments[i])
			}
		}
	})
	t.Run("sortino_downside_count", func(t *testing.T) {
		perf := NewRoundtripPerformance(100000.0, 0.0, 0.0, conventions.RAW)
		for _, rt := range allRTs {
			perf.AddRoundtrip(rt)
		}
		// RT3 and RT4 have negative ROI -> 2 downside entries
		if len(perf.SortinoDownsideReturns) != 2 {
			t.Errorf("expected 2 downside entries, got %d", len(perf.SortinoDownsideReturns))
		}
	})
}

// Ensure unused import doesn't cause issues
var _ = math.Abs
