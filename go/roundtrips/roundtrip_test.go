package roundtrips

import (
	"math"
	"testing"
	"time"
)

const epsilon = 1e-13

func almostEqual(a, b, eps float64) bool {
	return math.Abs(a-b) < eps
}

// ---------------------------------------------------------------------------
// Concrete test data
// ---------------------------------------------------------------------------

// Long trade: buy 100 shares at $50, sell at $55
var longEntry = Execution{
	Side:                Buy,
	Price:               50.0,
	CommissionPerUnit:   0.01,
	UnrealizedPriceHigh: 56.0,
	UnrealizedPriceLow:  48.0,
	DateTime:            time.Date(2024, 1, 1, 9, 30, 0, 0, time.UTC),
}
var longExit = Execution{
	Side:                Sell,
	Price:               55.0,
	CommissionPerUnit:   0.02,
	UnrealizedPriceHigh: 57.0,
	UnrealizedPriceLow:  49.0,
	DateTime:            time.Date(2024, 1, 5, 16, 0, 0, 0, time.UTC),
}
var longQty = 100.0

// Short trade: sell 200 shares at $80, buy-to-cover at $72
var shortEntry = Execution{
	Side:                Sell,
	Price:               80.0,
	CommissionPerUnit:   0.03,
	UnrealizedPriceHigh: 85.0,
	UnrealizedPriceLow:  72.0,
	DateTime:            time.Date(2024, 2, 1, 10, 0, 0, 0, time.UTC),
}
var shortExit = Execution{
	Side:                Buy,
	Price:               72.0,
	CommissionPerUnit:   0.02,
	UnrealizedPriceHigh: 83.0,
	UnrealizedPriceLow:  70.0,
	DateTime:            time.Date(2024, 2, 10, 15, 30, 0, 0, time.UTC),
}
var shortQty = 200.0

// ---------------------------------------------------------------------------
// Tests for a LONG round-trip
// ---------------------------------------------------------------------------

func TestRoundtripLong(t *testing.T) {
	rt := NewRoundtrip(longEntry, longExit, longQty)

	t.Run("side", func(t *testing.T) {
		if rt.Side != Long {
			t.Errorf("expected Long, got %v", rt.Side)
		}
	})
	t.Run("quantity", func(t *testing.T) {
		if !almostEqual(rt.Quantity, 100.0, epsilon) {
			t.Errorf("expected 100.0, got %v", rt.Quantity)
		}
	})
	t.Run("entry_time", func(t *testing.T) {
		expected := time.Date(2024, 1, 1, 9, 30, 0, 0, time.UTC)
		if !rt.EntryTime.Equal(expected) {
			t.Errorf("expected %v, got %v", expected, rt.EntryTime)
		}
	})
	t.Run("exit_time", func(t *testing.T) {
		expected := time.Date(2024, 1, 5, 16, 0, 0, 0, time.UTC)
		if !rt.ExitTime.Equal(expected) {
			t.Errorf("expected %v, got %v", expected, rt.ExitTime)
		}
	})
	t.Run("entry_price", func(t *testing.T) {
		if !almostEqual(rt.EntryPrice, 50.0, epsilon) {
			t.Errorf("expected 50.0, got %v", rt.EntryPrice)
		}
	})
	t.Run("exit_price", func(t *testing.T) {
		if !almostEqual(rt.ExitPrice, 55.0, epsilon) {
			t.Errorf("expected 55.0, got %v", rt.ExitPrice)
		}
	})
	t.Run("duration", func(t *testing.T) {
		expected := time.Date(2024, 1, 5, 16, 0, 0, 0, time.UTC).Sub(
			time.Date(2024, 1, 1, 9, 30, 0, 0, time.UTC))
		if rt.Duration != expected {
			t.Errorf("expected %v, got %v", expected, rt.Duration)
		}
	})
	t.Run("highest_price", func(t *testing.T) {
		if !almostEqual(rt.HighestPrice, 57.0, epsilon) {
			t.Errorf("expected 57.0, got %v", rt.HighestPrice)
		}
	})
	t.Run("lowest_price", func(t *testing.T) {
		if !almostEqual(rt.LowestPrice, 48.0, epsilon) {
			t.Errorf("expected 48.0, got %v", rt.LowestPrice)
		}
	})
	t.Run("gross_pnl", func(t *testing.T) {
		// Long: qty * (exit - entry) = 100 * (55 - 50) = 500
		if !almostEqual(rt.GrossPnl, 500.0, epsilon) {
			t.Errorf("expected 500.0, got %v", rt.GrossPnl)
		}
	})
	t.Run("commission", func(t *testing.T) {
		// (0.01 + 0.02) * 100 = 3.0
		if !almostEqual(rt.Commission, 3.0, epsilon) {
			t.Errorf("expected 3.0, got %v", rt.Commission)
		}
	})
	t.Run("net_pnl", func(t *testing.T) {
		// 500 - 3 = 497
		if !almostEqual(rt.NetPnl, 497.0, epsilon) {
			t.Errorf("expected 497.0, got %v", rt.NetPnl)
		}
	})
	t.Run("maximum_adverse_price", func(t *testing.T) {
		// Long: lowest_p = 48
		if !almostEqual(rt.MaximumAdversePrice, 48.0, epsilon) {
			t.Errorf("expected 48.0, got %v", rt.MaximumAdversePrice)
		}
	})
	t.Run("maximum_favorable_price", func(t *testing.T) {
		// Long: highest_p = 57
		if !almostEqual(rt.MaximumFavorablePrice, 57.0, epsilon) {
			t.Errorf("expected 57.0, got %v", rt.MaximumFavorablePrice)
		}
	})
	t.Run("maximum_adverse_excursion", func(t *testing.T) {
		// Long MAE: 100 * (1 - 48/50) = 4.0
		if !almostEqual(rt.MaximumAdverseExcursion, 4.0, epsilon) {
			t.Errorf("expected 4.0, got %v", rt.MaximumAdverseExcursion)
		}
	})
	t.Run("maximum_favorable_excursion", func(t *testing.T) {
		// Long MFE: 100 * (57/55 - 1)
		expected := 100.0 * (57.0/55.0 - 1.0)
		if !almostEqual(rt.MaximumFavorableExcursion, expected, epsilon) {
			t.Errorf("expected %v, got %v", expected, rt.MaximumFavorableExcursion)
		}
	})
	t.Run("entry_efficiency", func(t *testing.T) {
		// Long: 100 * (highest - entry) / delta = 100 * (57 - 50) / 9
		expected := 100.0 * (57.0 - 50.0) / 9.0
		if !almostEqual(rt.EntryEfficiency, expected, epsilon) {
			t.Errorf("expected %v, got %v", expected, rt.EntryEfficiency)
		}
	})
	t.Run("exit_efficiency", func(t *testing.T) {
		// Long: 100 * (exit - lowest) / delta = 100 * (55 - 48) / 9
		expected := 100.0 * (55.0 - 48.0) / 9.0
		if !almostEqual(rt.ExitEfficiency, expected, epsilon) {
			t.Errorf("expected %v, got %v", expected, rt.ExitEfficiency)
		}
	})
	t.Run("total_efficiency", func(t *testing.T) {
		// Long: 100 * (exit - entry) / delta = 100 * (55 - 50) / 9
		expected := 100.0 * (55.0 - 50.0) / 9.0
		if !almostEqual(rt.TotalEfficiency, expected, epsilon) {
			t.Errorf("expected %v, got %v", expected, rt.TotalEfficiency)
		}
	})
}

// ---------------------------------------------------------------------------
// Tests for a SHORT round-trip
// ---------------------------------------------------------------------------

func TestRoundtripShort(t *testing.T) {
	rt := NewRoundtrip(shortEntry, shortExit, shortQty)

	t.Run("side", func(t *testing.T) {
		if rt.Side != Short {
			t.Errorf("expected Short, got %v", rt.Side)
		}
	})
	t.Run("quantity", func(t *testing.T) {
		if !almostEqual(rt.Quantity, 200.0, epsilon) {
			t.Errorf("expected 200.0, got %v", rt.Quantity)
		}
	})
	t.Run("entry_time", func(t *testing.T) {
		expected := time.Date(2024, 2, 1, 10, 0, 0, 0, time.UTC)
		if !rt.EntryTime.Equal(expected) {
			t.Errorf("expected %v, got %v", expected, rt.EntryTime)
		}
	})
	t.Run("exit_time", func(t *testing.T) {
		expected := time.Date(2024, 2, 10, 15, 30, 0, 0, time.UTC)
		if !rt.ExitTime.Equal(expected) {
			t.Errorf("expected %v, got %v", expected, rt.ExitTime)
		}
	})
	t.Run("entry_price", func(t *testing.T) {
		if !almostEqual(rt.EntryPrice, 80.0, epsilon) {
			t.Errorf("expected 80.0, got %v", rt.EntryPrice)
		}
	})
	t.Run("exit_price", func(t *testing.T) {
		if !almostEqual(rt.ExitPrice, 72.0, epsilon) {
			t.Errorf("expected 72.0, got %v", rt.ExitPrice)
		}
	})
	t.Run("duration", func(t *testing.T) {
		expected := time.Date(2024, 2, 10, 15, 30, 0, 0, time.UTC).Sub(
			time.Date(2024, 2, 1, 10, 0, 0, 0, time.UTC))
		if rt.Duration != expected {
			t.Errorf("expected %v, got %v", expected, rt.Duration)
		}
	})
	t.Run("highest_price", func(t *testing.T) {
		if !almostEqual(rt.HighestPrice, 85.0, epsilon) {
			t.Errorf("expected 85.0, got %v", rt.HighestPrice)
		}
	})
	t.Run("lowest_price", func(t *testing.T) {
		if !almostEqual(rt.LowestPrice, 70.0, epsilon) {
			t.Errorf("expected 70.0, got %v", rt.LowestPrice)
		}
	})
	t.Run("gross_pnl", func(t *testing.T) {
		// Short: qty * (entry - exit) = 200 * (80 - 72) = 1600
		if !almostEqual(rt.GrossPnl, 1600.0, epsilon) {
			t.Errorf("expected 1600.0, got %v", rt.GrossPnl)
		}
	})
	t.Run("commission", func(t *testing.T) {
		// (0.03 + 0.02) * 200 = 10.0
		if !almostEqual(rt.Commission, 10.0, epsilon) {
			t.Errorf("expected 10.0, got %v", rt.Commission)
		}
	})
	t.Run("net_pnl", func(t *testing.T) {
		// 1600 - 10 = 1590
		if !almostEqual(rt.NetPnl, 1590.0, epsilon) {
			t.Errorf("expected 1590.0, got %v", rt.NetPnl)
		}
	})
	t.Run("maximum_adverse_price", func(t *testing.T) {
		// Short: highest_p = 85
		if !almostEqual(rt.MaximumAdversePrice, 85.0, epsilon) {
			t.Errorf("expected 85.0, got %v", rt.MaximumAdversePrice)
		}
	})
	t.Run("maximum_favorable_price", func(t *testing.T) {
		// Short: lowest_p = 70
		if !almostEqual(rt.MaximumFavorablePrice, 70.0, epsilon) {
			t.Errorf("expected 70.0, got %v", rt.MaximumFavorablePrice)
		}
	})
	t.Run("maximum_adverse_excursion", func(t *testing.T) {
		// Short MAE: 100 * (85/80 - 1) = 6.25
		if !almostEqual(rt.MaximumAdverseExcursion, 6.25, epsilon) {
			t.Errorf("expected 6.25, got %v", rt.MaximumAdverseExcursion)
		}
	})
	t.Run("maximum_favorable_excursion", func(t *testing.T) {
		// Short MFE: 100 * (1 - 70/72)
		expected := 100.0 * (1.0 - 70.0/72.0)
		if !almostEqual(rt.MaximumFavorableExcursion, expected, epsilon) {
			t.Errorf("expected %v, got %v", expected, rt.MaximumFavorableExcursion)
		}
	})
	t.Run("entry_efficiency", func(t *testing.T) {
		// Short: 100 * (entry - lowest) / delta = 100 * (80 - 70) / 15
		expected := 100.0 * (80.0 - 70.0) / 15.0
		if !almostEqual(rt.EntryEfficiency, expected, epsilon) {
			t.Errorf("expected %v, got %v", expected, rt.EntryEfficiency)
		}
	})
	t.Run("exit_efficiency", func(t *testing.T) {
		// Short: 100 * (highest - exit) / delta = 100 * (85 - 72) / 15
		expected := 100.0 * (85.0 - 72.0) / 15.0
		if !almostEqual(rt.ExitEfficiency, expected, epsilon) {
			t.Errorf("expected %v, got %v", expected, rt.ExitEfficiency)
		}
	})
	t.Run("total_efficiency", func(t *testing.T) {
		// Short: 100 * (entry - exit) / delta = 100 * (80 - 72) / 15
		expected := 100.0 * (80.0 - 72.0) / 15.0
		if !almostEqual(rt.TotalEfficiency, expected, epsilon) {
			t.Errorf("expected %v, got %v", expected, rt.TotalEfficiency)
		}
	})
}

// ---------------------------------------------------------------------------
// Tests for zero-delta edge case (highest == lowest)
// ---------------------------------------------------------------------------

func TestRoundtripZeroDelta(t *testing.T) {
	entry := Execution{
		Side:                Buy,
		Price:               100.0,
		CommissionPerUnit:   0.0,
		UnrealizedPriceHigh: 100.0,
		UnrealizedPriceLow:  100.0,
		DateTime:            time.Date(2024, 3, 1, 9, 0, 0, 0, time.UTC),
	}
	exit := Execution{
		Side:                Sell,
		Price:               100.0,
		CommissionPerUnit:   0.0,
		UnrealizedPriceHigh: 100.0,
		UnrealizedPriceLow:  100.0,
		DateTime:            time.Date(2024, 3, 1, 10, 0, 0, 0, time.UTC),
	}
	rt := NewRoundtrip(entry, exit, 50.0)

	t.Run("entry_efficiency_zero", func(t *testing.T) {
		if !almostEqual(rt.EntryEfficiency, 0.0, epsilon) {
			t.Errorf("expected 0.0, got %v", rt.EntryEfficiency)
		}
	})
	t.Run("exit_efficiency_zero", func(t *testing.T) {
		if !almostEqual(rt.ExitEfficiency, 0.0, epsilon) {
			t.Errorf("expected 0.0, got %v", rt.ExitEfficiency)
		}
	})
	t.Run("total_efficiency_zero", func(t *testing.T) {
		if !almostEqual(rt.TotalEfficiency, 0.0, epsilon) {
			t.Errorf("expected 0.0, got %v", rt.TotalEfficiency)
		}
	})
	t.Run("gross_pnl_zero", func(t *testing.T) {
		if !almostEqual(rt.GrossPnl, 0.0, epsilon) {
			t.Errorf("expected 0.0, got %v", rt.GrossPnl)
		}
	})
	t.Run("net_pnl_zero", func(t *testing.T) {
		if !almostEqual(rt.NetPnl, 0.0, epsilon) {
			t.Errorf("expected 0.0, got %v", rt.NetPnl)
		}
	})
}

// ---------------------------------------------------------------------------
// Long losing trade (exit < entry) -- verifies negative PnL path
// ---------------------------------------------------------------------------

func TestRoundtripLongLooser(t *testing.T) {
	entry := Execution{
		Side:                Buy,
		Price:               60.0,
		CommissionPerUnit:   0.005,
		UnrealizedPriceHigh: 62.0,
		UnrealizedPriceLow:  53.0,
		DateTime:            time.Date(2024, 4, 1, 9, 30, 0, 0, time.UTC),
	}
	exit := Execution{
		Side:                Sell,
		Price:               54.0,
		CommissionPerUnit:   0.005,
		UnrealizedPriceHigh: 61.0,
		UnrealizedPriceLow:  52.0,
		DateTime:            time.Date(2024, 4, 3, 16, 0, 0, 0, time.UTC),
	}
	rt := NewRoundtrip(entry, exit, 150.0)

	t.Run("side", func(t *testing.T) {
		if rt.Side != Long {
			t.Errorf("expected Long, got %v", rt.Side)
		}
	})
	t.Run("gross_pnl_negative", func(t *testing.T) {
		// 150 * (54 - 60) = -900
		if !almostEqual(rt.GrossPnl, -900.0, epsilon) {
			t.Errorf("expected -900.0, got %v", rt.GrossPnl)
		}
	})
	t.Run("commission", func(t *testing.T) {
		// (0.005 + 0.005) * 150 = 1.5
		if !almostEqual(rt.Commission, 1.5, epsilon) {
			t.Errorf("expected 1.5, got %v", rt.Commission)
		}
	})
	t.Run("net_pnl_negative", func(t *testing.T) {
		// -900 - 1.5 = -901.5
		if !almostEqual(rt.NetPnl, -901.5, epsilon) {
			t.Errorf("expected -901.5, got %v", rt.NetPnl)
		}
	})
	t.Run("highest_price", func(t *testing.T) {
		if !almostEqual(rt.HighestPrice, 62.0, epsilon) {
			t.Errorf("expected 62.0, got %v", rt.HighestPrice)
		}
	})
	t.Run("lowest_price", func(t *testing.T) {
		if !almostEqual(rt.LowestPrice, 52.0, epsilon) {
			t.Errorf("expected 52.0, got %v", rt.LowestPrice)
		}
	})
	t.Run("mae", func(t *testing.T) {
		// 100 * (1 - 52/60)
		expected := 100.0 * (1.0 - 52.0/60.0)
		if !almostEqual(rt.MaximumAdverseExcursion, expected, epsilon) {
			t.Errorf("expected %v, got %v", expected, rt.MaximumAdverseExcursion)
		}
	})
	t.Run("mfe", func(t *testing.T) {
		// 100 * (62/54 - 1)
		expected := 100.0 * (62.0/54.0 - 1.0)
		if !almostEqual(rt.MaximumFavorableExcursion, expected, epsilon) {
			t.Errorf("expected %v, got %v", expected, rt.MaximumFavorableExcursion)
		}
	})
}

// ---------------------------------------------------------------------------
// Short losing trade (exit > entry) -- verifies negative PnL path for shorts
// ---------------------------------------------------------------------------

func TestRoundtripShortLooser(t *testing.T) {
	entry := Execution{
		Side:                Sell,
		Price:               40.0,
		CommissionPerUnit:   0.01,
		UnrealizedPriceHigh: 42.0,
		UnrealizedPriceLow:  39.0,
		DateTime:            time.Date(2024, 5, 1, 10, 0, 0, 0, time.UTC),
	}
	exit := Execution{
		Side:                Buy,
		Price:               45.0,
		CommissionPerUnit:   0.01,
		UnrealizedPriceHigh: 46.0,
		UnrealizedPriceLow:  38.0,
		DateTime:            time.Date(2024, 5, 5, 15, 0, 0, 0, time.UTC),
	}
	rt := NewRoundtrip(entry, exit, 300.0)

	t.Run("side", func(t *testing.T) {
		if rt.Side != Short {
			t.Errorf("expected Short, got %v", rt.Side)
		}
	})
	t.Run("gross_pnl_negative", func(t *testing.T) {
		// 300 * (40 - 45) = -1500
		if !almostEqual(rt.GrossPnl, -1500.0, epsilon) {
			t.Errorf("expected -1500.0, got %v", rt.GrossPnl)
		}
	})
	t.Run("commission", func(t *testing.T) {
		// (0.01 + 0.01) * 300 = 6.0
		if !almostEqual(rt.Commission, 6.0, epsilon) {
			t.Errorf("expected 6.0, got %v", rt.Commission)
		}
	})
	t.Run("net_pnl_negative", func(t *testing.T) {
		// -1500 - 6 = -1506
		if !almostEqual(rt.NetPnl, -1506.0, epsilon) {
			t.Errorf("expected -1506.0, got %v", rt.NetPnl)
		}
	})
	t.Run("maximum_adverse_price", func(t *testing.T) {
		// Short: highest = 46
		if !almostEqual(rt.MaximumAdversePrice, 46.0, epsilon) {
			t.Errorf("expected 46.0, got %v", rt.MaximumAdversePrice)
		}
	})
	t.Run("maximum_favorable_price", func(t *testing.T) {
		// Short: lowest = 38
		if !almostEqual(rt.MaximumFavorablePrice, 38.0, epsilon) {
			t.Errorf("expected 38.0, got %v", rt.MaximumFavorablePrice)
		}
	})
	t.Run("mae", func(t *testing.T) {
		// Short MAE: 100 * (46/40 - 1) = 15.0
		if !almostEqual(rt.MaximumAdverseExcursion, 15.0, epsilon) {
			t.Errorf("expected 15.0, got %v", rt.MaximumAdverseExcursion)
		}
	})
	t.Run("mfe", func(t *testing.T) {
		// Short MFE: 100 * (1 - 38/45)
		expected := 100.0 * (1.0 - 38.0/45.0)
		if !almostEqual(rt.MaximumFavorableExcursion, expected, epsilon) {
			t.Errorf("expected %v, got %v", expected, rt.MaximumFavorableExcursion)
		}
	})
}
