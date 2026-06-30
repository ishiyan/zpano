//nolint:dupl
package truestrengthindex

import (
	"bytes"
	"fmt"
)

// Output describes the outputs of the indicator.
type Output int

const (
	// TSI is the True Strength Index oscillator value (range [-100, +100]).
	TSI Output = iota + 1

	// Signal is the signal-line value: the ul-period EMA of the oscillator.
	Signal

	outputLast
)

const (
	tSIValueStr    = "tsiValue"
	signalValueStr = "signalValue"
	unknownStr     = "unknown"
)

// String implements the Stringer interface.
func (o Output) String() string {
	switch o {
	case TSI:
		return tSIValueStr
	case Signal:
		return signalValueStr
	default:
		return unknownStr
	}
}

// IsKnown determines if this output is known.
func (o Output) IsKnown() bool {
	return o >= TSI && o < outputLast
}

// MarshalJSON implements the Marshaler interface.
func (o Output) MarshalJSON() ([]byte, error) {
	const (
		errFmt = "cannot marshal '%s': unknown true strength index output"
		extra  = 2   // Two bytes for quotes.
		dqc    = '"' // Double quote character.
	)

	s := o.String()
	if s == unknownStr {
		return nil, fmt.Errorf(errFmt, s)
	}

	b := make([]byte, 0, len(s)+extra)
	b = append(b, dqc)
	b = append(b, s...)
	b = append(b, dqc)

	return b, nil
}

// UnmarshalJSON implements the Unmarshaler interface.
func (o *Output) UnmarshalJSON(data []byte) error {
	const (
		errFmt = "cannot unmarshal '%s': unknown true strength index output"
		dqs    = "\"" // Double quote string.
	)

	d := bytes.Trim(data, dqs)
	s := string(d)

	switch s {
	case tSIValueStr:
		*o = TSI
	case signalValueStr:
		*o = Signal
	default:
		return fmt.Errorf(errFmt, s)
	}

	return nil
}
