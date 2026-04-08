package roundtrips

import (
	"time"
)

// Roundtrip represents an immutable position round-trip.
type Roundtrip struct {
	// Side is the side of the round-trip.
	Side RoundtripSide
	// Quantity is the total unsigned quantity of the position.
	Quantity float64
	// EntryTime is the date and time the position was opened.
	EntryTime time.Time
	// EntryPrice is the (average) price at which the position was opened.
	EntryPrice float64
	// ExitTime is the date and time the position was closed.
	ExitTime time.Time
	// ExitPrice is the (average) price at which the position was closed.
	ExitPrice float64
	// Duration is the duration of the round-trip.
	Duration time.Duration
	// HighestPrice is the highest price of the instrument during the round-trip.
	HighestPrice float64
	// LowestPrice is the lowest price of the instrument during the round-trip.
	LowestPrice float64
	// Commission is the total commission paid for the round-trip.
	Commission float64
	// GrossPnl is the gross Profit and Loss of the round-trip.
	GrossPnl float64
	// NetPnl is the net Profit and Loss of the round-trip.
	NetPnl float64
	// MaximumAdversePrice is the maximum adverse price during the round-trip.
	MaximumAdversePrice float64
	// MaximumFavorablePrice is the maximum favorable price during the round-trip.
	MaximumFavorablePrice float64
	// MaximumAdverseExcursion is the MAE percentage.
	MaximumAdverseExcursion float64
	// MaximumFavorableExcursion is the MFE percentage.
	MaximumFavorableExcursion float64
	// EntryEfficiency measures how close the entry was to the best possible entry.
	EntryEfficiency float64
	// ExitEfficiency measures how close the exit was to the best possible exit.
	ExitEfficiency float64
	// TotalEfficiency measures the ability to capture max profit potential.
	TotalEfficiency float64
}

// NewRoundtrip creates a new immutable Roundtrip from entry and exit executions
// and a quantity.
func NewRoundtrip(entry, exit Execution, quantity float64) Roundtrip {
	side := Long
	if entry.Side.IsSell() {
		side = Short
	}

	entryP := entry.Price
	exitP := exit.Price

	var pnl float64
	if side == Short {
		pnl = quantity * (entryP - exitP)
	} else {
		pnl = quantity * (exitP - entryP)
	}

	commission := (entry.CommissionPerUnit + exit.CommissionPerUnit) * quantity

	highestP := max(entry.UnrealizedPriceHigh, exit.UnrealizedPriceHigh)
	lowestP := min(entry.UnrealizedPriceLow, exit.UnrealizedPriceLow)
	delta := highestP - lowestP

	entryEfficiency := 0.0
	exitEfficiency := 0.0
	totalEfficiency := 0.0

	var maximumAdversePrice float64
	var maximumFavorablePrice float64
	var maximumAdverseExcursion float64
	var maximumFavorableExcursion float64

	if side == Long {
		maximumAdversePrice = lowestP
		maximumFavorablePrice = highestP
		maximumAdverseExcursion = 100.0 * (1.0 - lowestP/entryP)
		maximumFavorableExcursion = 100.0 * (highestP/exitP - 1.0)
		if delta != 0.0 {
			entryEfficiency = 100.0 * (highestP - entryP) / delta
			exitEfficiency = 100.0 * (exitP - lowestP) / delta
			totalEfficiency = 100.0 * (exitP - entryP) / delta
		}
	} else {
		maximumAdversePrice = highestP
		maximumFavorablePrice = lowestP
		maximumAdverseExcursion = 100.0 * (highestP/entryP - 1.0)
		maximumFavorableExcursion = 100.0 * (1.0 - lowestP/exitP)
		if delta != 0.0 {
			entryEfficiency = 100.0 * (entryP - lowestP) / delta
			exitEfficiency = 100.0 * (highestP - exitP) / delta
			totalEfficiency = 100.0 * (entryP - exitP) / delta
		}
	}

	duration := exit.DateTime.Sub(entry.DateTime)

	return Roundtrip{
		Side:                      side,
		Quantity:                  quantity,
		EntryTime:                 entry.DateTime,
		EntryPrice:                entryP,
		ExitTime:                  exit.DateTime,
		ExitPrice:                 exitP,
		Duration:                  duration,
		HighestPrice:              highestP,
		LowestPrice:               lowestP,
		Commission:                commission,
		GrossPnl:                  pnl,
		NetPnl:                    pnl - commission,
		MaximumAdversePrice:       maximumAdversePrice,
		MaximumFavorablePrice:     maximumFavorablePrice,
		MaximumAdverseExcursion:   maximumAdverseExcursion,
		MaximumFavorableExcursion: maximumFavorableExcursion,
		EntryEfficiency:           entryEfficiency,
		ExitEfficiency:            exitEfficiency,
		TotalEfficiency:           totalEfficiency,
	}
}
