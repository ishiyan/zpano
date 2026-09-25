//nolint:dupl
package doublesmoothedstochastic

import (
	"bytes"
	"fmt"
)

// Output describes the outputs of the indicator.
type Output int

const (
	// DSS is the Double Smoothed Stochastic oscillator value (range [0, 100]).
	DSS Output = iota + 1

	// Signal is the signal-line value: the g-period SMA of the oscillator.
	Signal

	outputLast
)

const (
	dssValueStr    = "dssValue"
	signalValueStr = "signalValue"
	unknownStr     = "unknown"
)

// String implements the Stringer interface.
func (o Output) String() string {
	switch o {
	case DSS:
		return dssValueStr
	case Signal:
		return signalValueStr
	default:
		return unknownStr
	}
}

// IsKnown determines if this output is known.
func (o Output) IsKnown() bool {
	return o >= DSS && o < outputLast
}

// MarshalJSON implements the Marshaler interface.
func (o Output) MarshalJSON() ([]byte, error) {
	const (
		errFmt = "cannot marshal '%s': unknown double smoothed stochastic output"
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
		errFmt = "cannot unmarshal '%s': unknown double smoothed stochastic output"
		dqs    = "\"" // Double quote string.
	)

	d := bytes.Trim(data, dqs)
	s := string(d)

	switch s {
	case dssValueStr:
		*o = DSS
	case signalValueStr:
		*o = Signal
	default:
		return fmt.Errorf(errFmt, s)
	}

	return nil
}
