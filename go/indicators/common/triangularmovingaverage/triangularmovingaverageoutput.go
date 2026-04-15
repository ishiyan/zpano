//nolint:dupl
package triangularmovingaverage

import (
	"bytes"
	"fmt"
)

// TriangularMovingAverageOutput describes the outputs of the indicator.
type TriangularMovingAverageOutput int

const (
	// The scalar value of the the moving average.
	TriangularMovingAverageValue TriangularMovingAverageOutput = iota + 1
	triangularMovingAverageLast
)

const (
	triangularMovingAverageValue   = "value"
	triangularMovingAverageUnknown = "unknown"
)

// String implements the Stringer interface.
func (o TriangularMovingAverageOutput) String() string {
	switch o {
	case TriangularMovingAverageValue:
		return triangularMovingAverageValue
	default:
		return triangularMovingAverageUnknown
	}
}

// IsKnown determines if this output is known.
func (o TriangularMovingAverageOutput) IsKnown() bool {
	return o >= TriangularMovingAverageValue && o < triangularMovingAverageLast
}

// MarshalJSON implements the Marshaler interface.
func (o TriangularMovingAverageOutput) MarshalJSON() ([]byte, error) {
	const (
		errFmt = "cannot marshal '%s': unknown triangular moving average output"
		extra  = 2   // Two bytes for quotes.
		dqc    = '"' // Double quote character.
	)

	s := o.String()
	if s == triangularMovingAverageUnknown {
		return nil, fmt.Errorf(errFmt, s)
	}

	b := make([]byte, 0, len(s)+extra)
	b = append(b, dqc)
	b = append(b, s...)
	b = append(b, dqc)

	return b, nil
}

// UnmarshalJSON implements the Unmarshaler interface.
func (o *TriangularMovingAverageOutput) UnmarshalJSON(data []byte) error {
	const (
		errFmt = "cannot unmarshal '%s': unknown triangular moving average output"
		dqs    = "\"" // Double quote string.
	)

	d := bytes.Trim(data, dqs)
	s := string(d)

	switch s {
	case triangularMovingAverageValue:
		*o = TriangularMovingAverageValue
	default:
		return fmt.Errorf(errFmt, s)
	}

	return nil
}
