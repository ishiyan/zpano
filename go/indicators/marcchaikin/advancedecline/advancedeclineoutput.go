//nolint:dupl
package advancedecline

import (
	"bytes"
	"fmt"
)

// AdvanceDeclineOutput describes the outputs of the indicator.
type AdvanceDeclineOutput int

const (
	// The scalar value of the advance-decline line.
	AdvanceDeclineValue AdvanceDeclineOutput = iota + 1
	advanceDeclineLast
)

const (
	advanceDeclineOutputValue   = "value"
	advanceDeclineOutputUnknown = "unknown"
)

// String implements the Stringer interface.
func (o AdvanceDeclineOutput) String() string {
	switch o {
	case AdvanceDeclineValue:
		return advanceDeclineOutputValue
	default:
		return advanceDeclineOutputUnknown
	}
}

// IsKnown determines if this output is known.
func (o AdvanceDeclineOutput) IsKnown() bool {
	return o >= AdvanceDeclineValue && o < advanceDeclineLast
}

// MarshalJSON implements the Marshaler interface.
func (o AdvanceDeclineOutput) MarshalJSON() ([]byte, error) {
	const (
		errFmt = "cannot marshal '%s': unknown advance-decline output"
		extra  = 2   // Two bytes for quotes.
		dqc    = '"' // Double quote character.
	)

	s := o.String()
	if s == advanceDeclineOutputUnknown {
		return nil, fmt.Errorf(errFmt, s)
	}

	b := make([]byte, 0, len(s)+extra)
	b = append(b, dqc)
	b = append(b, s...)
	b = append(b, dqc)

	return b, nil
}

// UnmarshalJSON implements the Unmarshaler interface.
func (o *AdvanceDeclineOutput) UnmarshalJSON(data []byte) error {
	const (
		errFmt = "cannot unmarshal '%s': unknown advance-decline output"
		dqs    = "\"" // Double quote string.
	)

	d := bytes.Trim(data, dqs)
	s := string(d)

	switch s {
	case advanceDeclineOutputValue:
		*o = AdvanceDeclineValue
	default:
		return fmt.Errorf(errFmt, s)
	}

	return nil
}
