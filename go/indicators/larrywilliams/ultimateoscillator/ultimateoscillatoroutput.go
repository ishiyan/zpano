//nolint:dupl
package ultimateoscillator

import (
	"bytes"
	"fmt"
)

// UltimateOscillatorOutput describes the outputs of the indicator.
type UltimateOscillatorOutput int

const (
	// The scalar value of the ultimate oscillator.
	UltimateOscillatorValue UltimateOscillatorOutput = iota + 1
	ultimateOscillatorLast
)

const (
	ultimateOscillatorOutputValue   = "value"
	ultimateOscillatorOutputUnknown = "unknown"
)

// String implements the Stringer interface.
func (o UltimateOscillatorOutput) String() string {
	switch o {
	case UltimateOscillatorValue:
		return ultimateOscillatorOutputValue
	default:
		return ultimateOscillatorOutputUnknown
	}
}

// IsKnown determines if this output is known.
func (o UltimateOscillatorOutput) IsKnown() bool {
	return o >= UltimateOscillatorValue && o < ultimateOscillatorLast
}

// MarshalJSON implements the Marshaler interface.
func (o UltimateOscillatorOutput) MarshalJSON() ([]byte, error) {
	const (
		errFmt = "cannot marshal '%s': unknown ultimate oscillator output"
		extra  = 2   // Two bytes for quotes.
		dqc    = '"' // Double quote character.
	)

	s := o.String()
	if s == ultimateOscillatorOutputUnknown {
		return nil, fmt.Errorf(errFmt, s)
	}

	b := make([]byte, 0, len(s)+extra)
	b = append(b, dqc)
	b = append(b, s...)
	b = append(b, dqc)

	return b, nil
}

// UnmarshalJSON implements the Unmarshaler interface.
func (o *UltimateOscillatorOutput) UnmarshalJSON(data []byte) error {
	const (
		errFmt = "cannot unmarshal '%s': unknown ultimate oscillator output"
		dqs    = "\"" // Double quote string.
	)

	d := bytes.Trim(data, dqs)
	s := string(d)

	switch s {
	case ultimateOscillatorOutputValue:
		*o = UltimateOscillatorValue
	default:
		return fmt.Errorf(errFmt, s)
	}

	return nil
}
