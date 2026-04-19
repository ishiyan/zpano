//nolint:dupl
package advancedeclineoscillator

import (
	"bytes"
	"fmt"
)

// AdvanceDeclineOscillatorOutput describes the outputs of the indicator.
type AdvanceDeclineOscillatorOutput int

const (
	// The scalar value of the Advance-Decline Oscillator.
	AdvanceDeclineOscillatorValue AdvanceDeclineOscillatorOutput = iota + 1
	advanceDeclineOscillatorLast
)

const (
	advanceDeclineOscillatorOutputValue   = "value"
	advanceDeclineOscillatorOutputUnknown = "unknown"
)

// String implements the Stringer interface.
func (o AdvanceDeclineOscillatorOutput) String() string {
	switch o {
	case AdvanceDeclineOscillatorValue:
		return advanceDeclineOscillatorOutputValue
	default:
		return advanceDeclineOscillatorOutputUnknown
	}
}

// IsKnown determines if this output is known.
func (o AdvanceDeclineOscillatorOutput) IsKnown() bool {
	return o >= AdvanceDeclineOscillatorValue && o < advanceDeclineOscillatorLast
}

// MarshalJSON implements the Marshaler interface.
func (o AdvanceDeclineOscillatorOutput) MarshalJSON() ([]byte, error) {
	const (
		errFmt = "cannot marshal '%s': unknown advance-decline oscillator output"
		extra  = 2   // Two bytes for quotes.
		dqc    = '"' // Double quote character.
	)

	s := o.String()
	if s == advanceDeclineOscillatorOutputUnknown {
		return nil, fmt.Errorf(errFmt, s)
	}

	b := make([]byte, 0, len(s)+extra)
	b = append(b, dqc)
	b = append(b, s...)
	b = append(b, dqc)

	return b, nil
}

// UnmarshalJSON implements the Unmarshaler interface.
func (o *AdvanceDeclineOscillatorOutput) UnmarshalJSON(data []byte) error {
	const (
		errFmt = "cannot unmarshal '%s': unknown advance-decline oscillator output"
		dqs    = "\"" // Double quote string.
	)

	d := bytes.Trim(data, dqs)
	s := string(d)

	switch s {
	case advanceDeclineOscillatorOutputValue:
		*o = AdvanceDeclineOscillatorValue
	default:
		return fmt.Errorf(errFmt, s)
	}

	return nil
}
