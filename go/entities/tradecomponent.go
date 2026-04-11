package data

import (
	"bytes"
	"errors"
	"fmt"
)

// TradeComponent describes a component of the Trade type.
type TradeComponent int

// TradeFunc defines a function to get a component value from the Trade type.
type TradeFunc func(t *Trade) float64

const (
	// TradePrice is the price component.
	TradePrice TradeComponent = iota + 1

	// TradeVolume is the volume component.
	TradeVolume
	tradeLast
)

const (
	tradePrice  = "price"
	tradeVolume = "volume"
)

var errUnknownTradeComponent = errors.New("unknown trade component")

// TradeComponentFunc returns a TradeFunc function to get a component value from the Trade type.
func TradeComponentFunc(c TradeComponent) (TradeFunc, error) {
	switch c {
	case TradePrice:
		return func(t *Trade) float64 { return t.Price }, nil
	case TradeVolume:
		return func(t *Trade) float64 { return t.Volume }, nil
	default:
		return nil, fmt.Errorf("%d: %w", int(c), errUnknownTradeComponent)
	}
}

// String implements the Stringer interface.
func (c TradeComponent) String() string {
	switch c {
	case TradePrice:
		return tradePrice
	case TradeVolume:
		return tradeVolume
	default:
		return unknown
	}
}

// IsKnown determines if this trade component is known.
func (c TradeComponent) IsKnown() bool {
	return c >= TradePrice && c < tradeLast
}

// MarshalJSON implements the Marshaler interface.
func (c TradeComponent) MarshalJSON() ([]byte, error) {
	s := c.String()
	if s == unknown {
		return nil, fmt.Errorf(marshalErrFmt, s, errUnknownTradeComponent)
	}

	const extra = 2 // Two bytes for quotes.

	b := make([]byte, 0, len(s)+extra)
	b = append(b, dqc)
	b = append(b, s...)
	b = append(b, dqc)

	return b, nil
}

// UnmarshalJSON implements the Unmarshaler interface.
func (c *TradeComponent) UnmarshalJSON(data []byte) error {
	d := bytes.Trim(data, dqs)
	s := string(d)

	switch s {
	case tradePrice:
		*c = TradePrice
	case tradeVolume:
		*c = TradeVolume
	default:
		return fmt.Errorf(unmarshalErrFmt, s, errUnknownTradeComponent)
	}

	return nil
}
