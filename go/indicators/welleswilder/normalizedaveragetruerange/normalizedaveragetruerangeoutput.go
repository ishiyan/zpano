//nolint:dupl
package normalizedaveragetruerange

import (
	"bytes"
	"fmt"
)

// NormalizedAverageTrueRangeOutput describes the outputs of the indicator.
type NormalizedAverageTrueRangeOutput int

const (
	// The scalar value of the normalized average true range.
	NormalizedAverageTrueRangeValue NormalizedAverageTrueRangeOutput = iota + 1
	normalizedAverageTrueRangeLast
)

const (
	normalizedAverageTrueRangeOutputValue   = "value"
	normalizedAverageTrueRangeOutputUnknown = "unknown"
)

// String implements the Stringer interface.
func (o NormalizedAverageTrueRangeOutput) String() string {
	switch o {
	case NormalizedAverageTrueRangeValue:
		return normalizedAverageTrueRangeOutputValue
	default:
		return normalizedAverageTrueRangeOutputUnknown
	}
}

// IsKnown determines if this output is known.
func (o NormalizedAverageTrueRangeOutput) IsKnown() bool {
	return o >= NormalizedAverageTrueRangeValue && o < normalizedAverageTrueRangeLast
}

// MarshalJSON implements the Marshaler interface.
func (o NormalizedAverageTrueRangeOutput) MarshalJSON() ([]byte, error) {
	const (
		errFmt = "cannot marshal '%s': unknown normalized average true range output"
		extra  = 2   // Two bytes for quotes.
		dqc    = '"' // Double quote character.
	)

	s := o.String()
	if s == normalizedAverageTrueRangeOutputUnknown {
		return nil, fmt.Errorf(errFmt, s)
	}

	b := make([]byte, 0, len(s)+extra)
	b = append(b, dqc)
	b = append(b, s...)
	b = append(b, dqc)

	return b, nil
}

// UnmarshalJSON implements the Unmarshaler interface.
func (o *NormalizedAverageTrueRangeOutput) UnmarshalJSON(data []byte) error {
	const (
		errFmt = "cannot unmarshal '%s': unknown normalized average true range output"
		dqs    = "\"" // Double quote string.
	)

	d := bytes.Trim(data, dqs)
	s := string(d)

	switch s {
	case normalizedAverageTrueRangeOutputValue:
		*o = NormalizedAverageTrueRangeValue
	default:
		return fmt.Errorf(errFmt, s)
	}

	return nil
}
