//nolint:dupl
package exponentialmovingaverage

import (
	"bytes"
	"fmt"
)

// ExponentialMovingAverageOutput describes the outputs of the indicator.
type ExponentialMovingAverageOutput int

const (
	// The scalar value of the the moving average.
	ExponentialMovingAverageValue ExponentialMovingAverageOutput = iota + 1
	exponentialMovingAverageLast
)

const (
	exponentialMovingAverageValue   = "value"
	exponentialMovingAverageUnknown = "unknown"
)

// String implements the Stringer interface.
func (o ExponentialMovingAverageOutput) String() string {
	switch o {
	case ExponentialMovingAverageValue:
		return exponentialMovingAverageValue
	default:
		return exponentialMovingAverageUnknown
	}
}

// IsKnown determines if this output is known.
func (o ExponentialMovingAverageOutput) IsKnown() bool {
	return o >= ExponentialMovingAverageValue && o < exponentialMovingAverageLast
}

// MarshalJSON implements the Marshaler interface.
func (o ExponentialMovingAverageOutput) MarshalJSON() ([]byte, error) {
	const (
		errFmt = "cannot marshal '%s': unknown exponential moving average output"
		extra  = 2   // Two bytes for quotes.
		dqc    = '"' // Double quote character.
	)

	s := o.String()
	if s == exponentialMovingAverageUnknown {
		return nil, fmt.Errorf(errFmt, s)
	}

	b := make([]byte, 0, len(s)+extra)
	b = append(b, dqc)
	b = append(b, s...)
	b = append(b, dqc)

	return b, nil
}

// UnmarshalJSON implements the Unmarshaler interface.
func (o *ExponentialMovingAverageOutput) UnmarshalJSON(data []byte) error {
	const (
		errFmt = "cannot unmarshal '%s': unknown exponential moving average output"
		dqs    = "\"" // Double quote string.
	)

	d := bytes.Trim(data, dqs)
	s := string(d)

	switch s {
	case exponentialMovingAverageValue:
		*o = ExponentialMovingAverageValue
	default:
		return fmt.Errorf(errFmt, s)
	}

	return nil
}
