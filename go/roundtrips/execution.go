package roundtrips

import "time"

// OrderSide enumerates the sides of an order.
type OrderSide int

const (
	// Buy represents a buy order.
	Buy OrderSide = iota
	// Sell represents a sell order.
	Sell
)

// IsSell returns true if the order side is Sell.
func (s OrderSide) IsSell() bool {
	return s == Sell
}

// Execution represents an order execution (fill).
type Execution struct {
	// Side is the side of the order.
	Side OrderSide
	// Price is the execution price.
	Price float64
	// CommissionPerUnit is the commission per unit of quantity.
	CommissionPerUnit float64
	// UnrealizedPriceHigh is the highest unrealized price during the execution period.
	UnrealizedPriceHigh float64
	// UnrealizedPriceLow is the lowest unrealized price during the execution period.
	UnrealizedPriceLow float64
	// DateTime is the date and time of the execution.
	DateTime time.Time
}
