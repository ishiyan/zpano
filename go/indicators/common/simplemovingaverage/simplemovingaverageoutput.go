//nolint:dupl
package simplemovingaverage

import (
	"bytes"
	"fmt"
)

// SimpleMovingAverageOutput describes the outputs of the indicator.
type SimpleMovingAverageOutput int

const (
	// The scalar value of the the moving average.
	SimpleMovingAverageValue SimpleMovingAverageOutput = iota + 1
	simpleMovingAverageLast
)

const (
	simpleMovingAverageValue   = "value"
	simpleMovingAverageUnknown = "unknown"
)

// String implements the Stringer interface.
func (o SimpleMovingAverageOutput) String() string {
	switch o {
	case SimpleMovingAverageValue:
		return simpleMovingAverageValue
	default:
		return simpleMovingAverageUnknown
	}
}

// IsKnown determines if this output is known.
func (o SimpleMovingAverageOutput) IsKnown() bool {
	return o >= SimpleMovingAverageValue && o < simpleMovingAverageLast
}

// MarshalJSON implements the Marshaler interface.
func (o SimpleMovingAverageOutput) MarshalJSON() ([]byte, error) {
	const (
		errFmt = "cannot marshal '%s': unknown simple moving average output"
		extra  = 2   // Two bytes for quotes.
		dqc    = '"' // Double quote character.
	)

	s := o.String()
	if s == simpleMovingAverageUnknown {
		return nil, fmt.Errorf(errFmt, s)
	}

	b := make([]byte, 0, len(s)+extra)
	b = append(b, dqc)
	b = append(b, s...)
	b = append(b, dqc)

	return b, nil
}

// UnmarshalJSON implements the Unmarshaler interface.
func (o *SimpleMovingAverageOutput) UnmarshalJSON(data []byte) error {
	const (
		errFmt = "cannot unmarshal '%s': unknown simple moving average output"
		dqs    = "\"" // Double quote string.
	)

	d := bytes.Trim(data, dqs)
	s := string(d)

	switch s {
	case simpleMovingAverageValue:
		*o = SimpleMovingAverageValue
	default:
		return fmt.Errorf(errFmt, s)
	}

	return nil
}
