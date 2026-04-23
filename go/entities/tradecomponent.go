package entities

import (
	"bytes"
	"errors"
	"fmt"
)

// TradeComponent describes a component of the Trade type.
type TradeComponent int

// TradeFunc defines a function to get a component value from the Trade type.
type TradeFunc func(t *Trade) float64

// DefaultTradeComponent is the default trade component used when no explicit component is specified.
const DefaultTradeComponent = TradePrice

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

const (
	tradeMnemonicPrice  = "p"
	tradeMnemonicVolume = "v"
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
func (s TradeComponent) String() string {
	switch s {
	case TradePrice:
		return tradePrice
	case TradeVolume:
		return tradeVolume
	default:
		return unknown
	}
}

// Mnemonic returns a short mnemonic code for the trade component.
func (s TradeComponent) Mnemonic() string {
	switch s {
	case TradePrice:
		return tradeMnemonicPrice
	case TradeVolume:
		return tradeMnemonicVolume
	default:
		return unknown
	}
}

// IsKnown determines if this trade component is known.
func (s TradeComponent) IsKnown() bool {
	return s >= TradePrice && s < tradeLast
}

// MarshalJSON implements the Marshaler interface.
func (s TradeComponent) MarshalJSON() ([]byte, error) {
	str := s.String()
	if str == unknown {
		return nil, fmt.Errorf(marshalErrFmt, str, errUnknownTradeComponent)
	}

	const extra = 2 // Two bytes for quotes.

	b := make([]byte, 0, len(str)+extra)
	b = append(b, dqc)
	b = append(b, str...)
	b = append(b, dqc)

	return b, nil
}

// UnmarshalJSON implements the Unmarshaler interface.
func (s *TradeComponent) UnmarshalJSON(data []byte) error {
	d := bytes.Trim(data, dqs)
	str := string(d)

	switch str {
	case tradePrice:
		*s = TradePrice
	case tradeVolume:
		*s = TradeVolume
	default:
		return fmt.Errorf(unmarshalErrFmt, str, errUnknownTradeComponent)
	}

	return nil
}
