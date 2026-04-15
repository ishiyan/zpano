//nolint:dupl
package kaufmanadaptivemovingaverage

import (
	"bytes"
	"fmt"
)

// KaufmanAdaptiveMovingAverageOutput describes the outputs of the indicator.
type KaufmanAdaptiveMovingAverageOutput int

const (
	// The scalar value of the moving average.
	KaufmanAdaptiveMovingAverageValue KaufmanAdaptiveMovingAverageOutput = iota + 1
	kaufmanAdaptiveMovingAverageLast
)

const (
	kaufmanAdaptiveMovingAverageValue   = "value"
	kaufmanAdaptiveMovingAverageUnknown = "unknown"
)

// String implements the Stringer interface.
func (o KaufmanAdaptiveMovingAverageOutput) String() string {
	switch o {
	case KaufmanAdaptiveMovingAverageValue:
		return kaufmanAdaptiveMovingAverageValue
	default:
		return kaufmanAdaptiveMovingAverageUnknown
	}
}

// IsKnown determines if this output is known.
func (o KaufmanAdaptiveMovingAverageOutput) IsKnown() bool {
	return o >= KaufmanAdaptiveMovingAverageValue && o < kaufmanAdaptiveMovingAverageLast
}

// MarshalJSON implements the Marshaler interface.
func (o KaufmanAdaptiveMovingAverageOutput) MarshalJSON() ([]byte, error) {
	const (
		errFmt = "cannot marshal '%s': unknown kaufman adaptive moving average output"
		extra  = 2   // Two bytes for quotes.
		dqc    = '"' // Double quote character.
	)

	s := o.String()
	if s == kaufmanAdaptiveMovingAverageUnknown {
		return nil, fmt.Errorf(errFmt, s)
	}

	b := make([]byte, 0, len(s)+extra)
	b = append(b, dqc)
	b = append(b, s...)
	b = append(b, dqc)

	return b, nil
}

// UnmarshalJSON implements the Unmarshaler interface.
func (o *KaufmanAdaptiveMovingAverageOutput) UnmarshalJSON(data []byte) error {
	const (
		errFmt = "cannot unmarshal '%s': unknown kaufman adaptive moving average output"
		dqs    = "\"" // Double quote string.
	)

	d := bytes.Trim(data, dqs)
	s := string(d)

	switch s {
	case kaufmanAdaptiveMovingAverageValue:
		*o = KaufmanAdaptiveMovingAverageValue
	default:
		return fmt.Errorf(errFmt, s)
	}

	return nil
}
