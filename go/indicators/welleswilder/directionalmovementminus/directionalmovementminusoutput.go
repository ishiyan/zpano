//nolint:dupl
package directionalmovementminus

import (
	"bytes"
	"fmt"
)

// DirectionalMovementMinusOutput describes the outputs of the indicator.
type DirectionalMovementMinusOutput int

const (
	// The scalar value of the directional movement minus.
	DirectionalMovementMinusValue DirectionalMovementMinusOutput = iota + 1
	directionalMovementMinusLast
)

const (
	directionalMovementMinusOutputValue   = "value"
	directionalMovementMinusOutputUnknown = "unknown"
)

// String implements the Stringer interface.
func (o DirectionalMovementMinusOutput) String() string {
	switch o {
	case DirectionalMovementMinusValue:
		return directionalMovementMinusOutputValue
	default:
		return directionalMovementMinusOutputUnknown
	}
}

// IsKnown determines if this output is known.
func (o DirectionalMovementMinusOutput) IsKnown() bool {
	return o >= DirectionalMovementMinusValue && o < directionalMovementMinusLast
}

// MarshalJSON implements the Marshaler interface.
func (o DirectionalMovementMinusOutput) MarshalJSON() ([]byte, error) {
	const (
		errFmt = "cannot marshal '%s': unknown directional movement minus output"
		extra  = 2   // Two bytes for quotes.
		dqc    = '"' // Double quote character.
	)

	s := o.String()
	if s == directionalMovementMinusOutputUnknown {
		return nil, fmt.Errorf(errFmt, s)
	}

	b := make([]byte, 0, len(s)+extra)
	b = append(b, dqc)
	b = append(b, s...)
	b = append(b, dqc)

	return b, nil
}

// UnmarshalJSON implements the Unmarshaler interface.
func (o *DirectionalMovementMinusOutput) UnmarshalJSON(data []byte) error {
	const (
		errFmt = "cannot unmarshal '%s': unknown directional movement minus output"
		dqs    = "\"" // Double quote string.
	)

	d := bytes.Trim(data, dqs)
	s := string(d)

	switch s {
	case directionalMovementMinusOutputValue:
		*o = DirectionalMovementMinusValue
	default:
		return fmt.Errorf(errFmt, s)
	}

	return nil
}
