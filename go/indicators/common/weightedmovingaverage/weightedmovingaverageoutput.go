//nolint:dupl
package weightedmovingaverage

import (
	"bytes"
	"fmt"
)

// WeightedMovingAverageOutput describes the outputs of the indicator.
type WeightedMovingAverageOutput int

const (
	// The scalar value of the the moving average.
	WeightedMovingAverageValue WeightedMovingAverageOutput = iota + 1
	weightedMovingAverageLast
)

const (
	weightedMovingAverageValue   = "value"
	weightedMovingAverageUnknown = "unknown"
)

// String implements the Stringer interface.
func (o WeightedMovingAverageOutput) String() string {
	switch o {
	case WeightedMovingAverageValue:
		return weightedMovingAverageValue
	default:
		return weightedMovingAverageUnknown
	}
}

// IsKnown determines if this output is known.
func (o WeightedMovingAverageOutput) IsKnown() bool {
	return o >= WeightedMovingAverageValue && o < weightedMovingAverageLast
}

// MarshalJSON implements the Marshaler interface.
func (o WeightedMovingAverageOutput) MarshalJSON() ([]byte, error) {
	const (
		errFmt = "cannot marshal '%s': unknown weighted moving average output"
		extra  = 2   // Two bytes for quotes.
		dqc    = '"' // Double quote character.
	)

	s := o.String()
	if s == weightedMovingAverageUnknown {
		return nil, fmt.Errorf(errFmt, s)
	}

	b := make([]byte, 0, len(s)+extra)
	b = append(b, dqc)
	b = append(b, s...)
	b = append(b, dqc)

	return b, nil
}

// UnmarshalJSON implements the Unmarshaler interface.
func (o *WeightedMovingAverageOutput) UnmarshalJSON(data []byte) error {
	const (
		errFmt = "cannot unmarshal '%s': unknown weighted moving average output"
		dqs    = "\"" // Double quote string.
	)

	d := bytes.Trim(data, dqs)
	s := string(d)

	switch s {
	case weightedMovingAverageValue:
		*o = WeightedMovingAverageValue
	default:
		return fmt.Errorf(errFmt, s)
	}

	return nil
}
