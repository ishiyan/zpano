package entities

import (
	"bytes"
	"errors"
	"fmt"
)

// BarComponent describes a component of the Bar type.
type BarComponent int

// BarFunc defines a function to get a component value from the Bar type.
type BarFunc func(b *Bar) float64

// DefaultBarComponent is the default bar component used when no explicit component is specified.
const DefaultBarComponent = BarClosePrice

const (
	// BarOpenPrice is the opening price component.
	BarOpenPrice BarComponent = iota + 1

	// BarHighPrice is the highest price component.
	BarHighPrice

	// BarLowPrice is the lowest price component.
	BarLowPrice

	// BarClosePrice is the closing price component.
	BarClosePrice

	// BarVolume is the volume component.
	BarVolume

	// BarMedianPrice is the median price component, calculated as
	//   (low + high) / 2.
	BarMedianPrice

	// BarTypicalPrice is the typical price component, calculated as
	//   (low + high + close) / 3.
	BarTypicalPrice

	// BarWeightedPrice is the weighted price component, calculated as
	//   (low + high + 2*close) / 4.
	BarWeightedPrice

	// BarAveragePrice is the average price component, calculated as
	//   (low + high + open + close) / 4.
	BarAveragePrice
	barLast
)

const (
	barOpen     = "open"
	barHigh     = "high"
	barLow      = "low"
	barClose    = "close"
	barVolume   = "volume"
	barMedian   = "median"
	barTypical  = "typical"
	barWeighted = "weighted"
	barAverage  = "average"
)

const (
	barMnemonicOpen     = "o"
	barMnemonicHigh     = "h"
	barMnemonicLow      = "l"
	barMnemonicClose    = "c"
	barMnemonicVolume   = "v"
	barMnemonicMedian   = "hl/2"
	barMnemonicTypical  = "hlc/3"
	barMnemonicWeighted = "hlcc/4"
	barMnemonicAverage  = "ohlc/4"
)

var errUnknownBarComponent = errors.New("unknown bar component")

// BarComponentFunc returns an BarFunc function to get a component value from the Bar type.
func BarComponentFunc(c BarComponent) (BarFunc, error) {
	switch c {
	case BarOpenPrice:
		return func(b *Bar) float64 { return b.Open }, nil
	case BarHighPrice:
		return func(b *Bar) float64 { return b.High }, nil
	case BarLowPrice:
		return func(b *Bar) float64 { return b.Low }, nil
	case BarClosePrice:
		return func(b *Bar) float64 { return b.Close }, nil
	case BarVolume:
		return func(b *Bar) float64 { return b.Volume }, nil
	case BarMedianPrice:
		return func(b *Bar) float64 { return b.Median() }, nil
	case BarTypicalPrice:
		return func(b *Bar) float64 { return b.Typical() }, nil
	case BarWeightedPrice:
		return func(b *Bar) float64 { return b.Weighted() }, nil
	case BarAveragePrice:
		return func(b *Bar) float64 { return b.Average() }, nil
	default:
		return nil, fmt.Errorf("%d: %w", int(c), errUnknownBarComponent)
	}
}

// String implements the Stringer interface.
func (s BarComponent) String() string {
	switch s {
	case BarOpenPrice:
		return barOpen
	case BarHighPrice:
		return barHigh
	case BarLowPrice:
		return barLow
	case BarClosePrice:
		return barClose
	case BarVolume:
		return barVolume
	case BarMedianPrice:
		return barMedian
	case BarTypicalPrice:
		return barTypical
	case BarWeightedPrice:
		return barWeighted
	case BarAveragePrice:
		return barAverage
	default:
		return unknown
	}
}

// Mnemonic returns a short mnemonic code for the bar component.
func (s BarComponent) Mnemonic() string {
	switch s {
	case BarOpenPrice:
		return barMnemonicOpen
	case BarHighPrice:
		return barMnemonicHigh
	case BarLowPrice:
		return barMnemonicLow
	case BarClosePrice:
		return barMnemonicClose
	case BarVolume:
		return barMnemonicVolume
	case BarMedianPrice:
		return barMnemonicMedian
	case BarTypicalPrice:
		return barMnemonicTypical
	case BarWeightedPrice:
		return barMnemonicWeighted
	case BarAveragePrice:
		return barMnemonicAverage
	default:
		return unknown
	}
}

// IsKnown determines if this bar component is known.
func (s BarComponent) IsKnown() bool {
	return s >= BarOpenPrice && s < barLast
}

// MarshalJSON implements the Marshaler interface.
func (s BarComponent) MarshalJSON() ([]byte, error) {
	str := s.String()
	if str == unknown {
		return nil, fmt.Errorf(marshalErrFmt, str, errUnknownBarComponent)
	}

	const extra = 2 // Two bytes for quotes.

	b := make([]byte, 0, len(str)+extra)
	b = append(b, dqc)
	b = append(b, str...)
	b = append(b, dqc)

	return b, nil
}

// UnmarshalJSON implements the Unmarshaler interface.
func (s *BarComponent) UnmarshalJSON(data []byte) error {
	d := bytes.Trim(data, dqs)
	str := string(d)

	switch str {
	case barOpen:
		*s = BarOpenPrice
	case barHigh:
		*s = BarHighPrice
	case barLow:
		*s = BarLowPrice
	case barClose:
		*s = BarClosePrice
	case barVolume:
		*s = BarVolume
	case barMedian:
		*s = BarMedianPrice
	case barTypical:
		*s = BarTypicalPrice
	case barWeighted:
		*s = BarWeightedPrice
	case barAverage:
		*s = BarAveragePrice
	default:
		return fmt.Errorf(unmarshalErrFmt, str, errUnknownBarComponent)
	}

	return nil
}
