//nolint:dupl
package standarddeviation

import (
	"bytes"
	"fmt"
)

// StandardDeviationOutput describes the outputs of the indicator.
type StandardDeviationOutput int

const (
	// The scalar value of the the standard deviation.
	StandardDeviationValue StandardDeviationOutput = iota + 1
	standardDeviationLast
)

const (
	standardDeviationValue   = "value"
	standardDeviationUnknown = "unknown"
)

// String implements the Stringer interface.
func (o StandardDeviationOutput) String() string {
	switch o {
	case StandardDeviationValue:
		return standardDeviationValue
	default:
		return standardDeviationUnknown
	}
}

// IsKnown determines if this output is known.
func (o StandardDeviationOutput) IsKnown() bool {
	return o >= StandardDeviationValue && o < standardDeviationLast
}

// MarshalJSON implements the Marshaler interface.
func (o StandardDeviationOutput) MarshalJSON() ([]byte, error) {
	const (
		errFmt = "cannot marshal '%s': unknown standard deviation output"
		extra  = 2   // Two bytes for quotes.
		dqc    = '"' // Double quote character.
	)

	s := o.String()
	if s == standardDeviationUnknown {
		return nil, fmt.Errorf(errFmt, s)
	}

	b := make([]byte, 0, len(s)+extra)
	b = append(b, dqc)
	b = append(b, s...)
	b = append(b, dqc)

	return b, nil
}

// UnmarshalJSON implements the Unmarshaler interface.
func (o *StandardDeviationOutput) UnmarshalJSON(data []byte) error {
	const (
		errFmt = "cannot unmarshal '%s': unknown standard deviation output"
		dqs    = "\"" // Double quote string.
	)

	d := bytes.Trim(data, dqs)
	s := string(d)

	switch s {
	case standardDeviationValue:
		*o = StandardDeviationValue
	default:
		return fmt.Errorf(errFmt, s)
	}

	return nil
}
