//nolint:dupl
package truerange

import (
	"bytes"
	"fmt"
)

// TrueRangeOutput describes the outputs of the indicator.
type TrueRangeOutput int

const (
	// The scalar value of the true range.
	TrueRangeValue TrueRangeOutput = iota + 1
	trueRangeLast
)

const (
	trueRangeOutputValue   = "value"
	trueRangeOutputUnknown = "unknown"
)

// String implements the Stringer interface.
func (o TrueRangeOutput) String() string {
	switch o {
	case TrueRangeValue:
		return trueRangeOutputValue
	default:
		return trueRangeOutputUnknown
	}
}

// IsKnown determines if this output is known.
func (o TrueRangeOutput) IsKnown() bool {
	return o >= TrueRangeValue && o < trueRangeLast
}

// MarshalJSON implements the Marshaler interface.
func (o TrueRangeOutput) MarshalJSON() ([]byte, error) {
	const (
		errFmt = "cannot marshal '%s': unknown true range output"
		extra  = 2   // Two bytes for quotes.
		dqc    = '"' // Double quote character.
	)

	s := o.String()
	if s == trueRangeOutputUnknown {
		return nil, fmt.Errorf(errFmt, s)
	}

	b := make([]byte, 0, len(s)+extra)
	b = append(b, dqc)
	b = append(b, s...)
	b = append(b, dqc)

	return b, nil
}

// UnmarshalJSON implements the Unmarshaler interface.
func (o *TrueRangeOutput) UnmarshalJSON(data []byte) error {
	const (
		errFmt = "cannot unmarshal '%s': unknown true range output"
		dqs    = "\"" // Double quote string.
	)

	d := bytes.Trim(data, dqs)
	s := string(d)

	switch s {
	case trueRangeOutputValue:
		*o = TrueRangeValue
	default:
		return fmt.Errorf(errFmt, s)
	}

	return nil
}
