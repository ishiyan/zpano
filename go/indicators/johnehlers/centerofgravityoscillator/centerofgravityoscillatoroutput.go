package centerofgravityoscillator

import (
	"bytes"
	"fmt"
)

// Output describes the outputs of the indicator.
type Output int

const (
	// Value is the scalar value of the center of gravity oscillator.
	Value Output = iota + 1
	// Trigger is the scalar value of the trigger line (previous value of the oscillator).
	Trigger
	outputLast
)

const (
	valueStr   = "value"
	triggerStr = "trigger"
	unknownStr = "unknown"
)

// String implements the Stringer interface.
func (o Output) String() string {
	switch o {
	case Value:
		return valueStr
	case Trigger:
		return triggerStr
	default:
		return unknownStr
	}
}

// IsKnown determines if this output is known.
func (o Output) IsKnown() bool {
	return o >= Value && o < outputLast
}

// MarshalJSON implements the Marshaler interface.
func (o Output) MarshalJSON() ([]byte, error) {
	const (
		errFmt = "cannot marshal '%s': unknown center of gravity oscillator output"
		extra  = 2   // Two bytes for quotes.
		dqc    = '"' // Double quote character.
	)

	s := o.String()
	if s == unknownStr {
		return nil, fmt.Errorf(errFmt, s)
	}

	b := make([]byte, 0, len(s)+extra)
	b = append(b, dqc)
	b = append(b, s...)
	b = append(b, dqc)

	return b, nil
}

// UnmarshalJSON implements the Unmarshaler interface.
func (o *Output) UnmarshalJSON(data []byte) error {
	const (
		errFmt = "cannot unmarshal '%s': unknown center of gravity oscillator output"
		dqs    = "\"" // Double quote string.
	)

	d := bytes.Trim(data, dqs)
	s := string(d)

	switch s {
	case valueStr:
		*o = Value
	case triggerStr:
		*o = Trigger
	default:
		return fmt.Errorf(errFmt, s)
	}

	return nil
}
