//nolint:dupl
package relativestrengthindex

import (
	"bytes"
	"fmt"
)

// RelativeStrengthIndexOutput describes the outputs of the indicator.
type RelativeStrengthIndexOutput int

const (
	// The scalar value of the relative strength index.
	RelativeStrengthIndexValue RelativeStrengthIndexOutput = iota + 1
	relativeStrengthIndexLast
)

const (
	relativeStrengthIndexOutputValue   = "value"
	relativeStrengthIndexOutputUnknown = "unknown"
)

// String implements the Stringer interface.
func (o RelativeStrengthIndexOutput) String() string {
	switch o {
	case RelativeStrengthIndexValue:
		return relativeStrengthIndexOutputValue
	default:
		return relativeStrengthIndexOutputUnknown
	}
}

// IsKnown determines if this output is known.
func (o RelativeStrengthIndexOutput) IsKnown() bool {
	return o >= RelativeStrengthIndexValue && o < relativeStrengthIndexLast
}

// MarshalJSON implements the Marshaler interface.
func (o RelativeStrengthIndexOutput) MarshalJSON() ([]byte, error) {
	const (
		errFmt = "cannot marshal '%s': unknown relative strength index output"
		extra  = 2   // Two bytes for quotes.
		dqc    = '"' // Double quote character.
	)

	s := o.String()
	if s == relativeStrengthIndexOutputUnknown {
		return nil, fmt.Errorf(errFmt, s)
	}

	b := make([]byte, 0, len(s)+extra)
	b = append(b, dqc)
	b = append(b, s...)
	b = append(b, dqc)

	return b, nil
}

// UnmarshalJSON implements the Unmarshaler interface.
func (o *RelativeStrengthIndexOutput) UnmarshalJSON(data []byte) error {
	const (
		errFmt = "cannot unmarshal '%s': unknown relative strength index output"
		dqs    = "\"" // Double quote string.
	)

	d := bytes.Trim(data, dqs)
	s := string(d)

	switch s {
	case relativeStrengthIndexOutputValue:
		*o = RelativeStrengthIndexValue
	default:
		return fmt.Errorf(errFmt, s)
	}

	return nil
}
