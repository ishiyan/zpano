//nolint:dupl
package directionalmovementplus

import (
	"bytes"
	"fmt"
)

// DirectionalMovementPlusOutput describes the outputs of the indicator.
type DirectionalMovementPlusOutput int

const (
	// The scalar value of the directional movement plus.
	DirectionalMovementPlusValue DirectionalMovementPlusOutput = iota + 1
	directionalMovementPlusLast
)

const (
	directionalMovementPlusOutputValue   = "value"
	directionalMovementPlusOutputUnknown = "unknown"
)

// String implements the Stringer interface.
func (o DirectionalMovementPlusOutput) String() string {
	switch o {
	case DirectionalMovementPlusValue:
		return directionalMovementPlusOutputValue
	default:
		return directionalMovementPlusOutputUnknown
	}
}

// IsKnown determines if this output is known.
func (o DirectionalMovementPlusOutput) IsKnown() bool {
	return o >= DirectionalMovementPlusValue && o < directionalMovementPlusLast
}

// MarshalJSON implements the Marshaler interface.
func (o DirectionalMovementPlusOutput) MarshalJSON() ([]byte, error) {
	const (
		errFmt = "cannot marshal '%s': unknown directional movement plus output"
		extra  = 2   // Two bytes for quotes.
		dqc    = '"' // Double quote character.
	)

	s := o.String()
	if s == directionalMovementPlusOutputUnknown {
		return nil, fmt.Errorf(errFmt, s)
	}

	b := make([]byte, 0, len(s)+extra)
	b = append(b, dqc)
	b = append(b, s...)
	b = append(b, dqc)

	return b, nil
}

// UnmarshalJSON implements the Unmarshaler interface.
func (o *DirectionalMovementPlusOutput) UnmarshalJSON(data []byte) error {
	const (
		errFmt = "cannot unmarshal '%s': unknown directional movement plus output"
		dqs    = "\"" // Double quote string.
	)

	d := bytes.Trim(data, dqs)
	s := string(d)

	switch s {
	case directionalMovementPlusOutputValue:
		*o = DirectionalMovementPlusValue
	default:
		return fmt.Errorf(errFmt, s)
	}

	return nil
}
