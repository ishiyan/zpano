//nolint:dupl
package absolutepriceoscillator

import (
	"bytes"
	"fmt"
)

// AbsolutePriceOscillatorOutput describes the outputs of the indicator.
type AbsolutePriceOscillatorOutput int

const (
	// The scalar value of the absolute price oscillator.
	AbsolutePriceOscillatorValue AbsolutePriceOscillatorOutput = iota + 1
	absolutePriceOscillatorLast
)

const (
	absolutePriceOscillatorOutputValue   = "value"
	absolutePriceOscillatorOutputUnknown = "unknown"
)

// String implements the Stringer interface.
func (o AbsolutePriceOscillatorOutput) String() string {
	switch o {
	case AbsolutePriceOscillatorValue:
		return absolutePriceOscillatorOutputValue
	default:
		return absolutePriceOscillatorOutputUnknown
	}
}

// IsKnown determines if this output is known.
func (o AbsolutePriceOscillatorOutput) IsKnown() bool {
	return o >= AbsolutePriceOscillatorValue && o < absolutePriceOscillatorLast
}

// MarshalJSON implements the Marshaler interface.
func (o AbsolutePriceOscillatorOutput) MarshalJSON() ([]byte, error) {
	const (
		errFmt = "cannot marshal '%s': unknown absolute price oscillator output"
		extra  = 2   // Two bytes for quotes.
		dqc    = '"' // Double quote character.
	)

	s := o.String()
	if s == absolutePriceOscillatorOutputUnknown {
		return nil, fmt.Errorf(errFmt, s)
	}

	b := make([]byte, 0, len(s)+extra)
	b = append(b, dqc)
	b = append(b, s...)
	b = append(b, dqc)

	return b, nil
}

// UnmarshalJSON implements the Unmarshaler interface.
func (o *AbsolutePriceOscillatorOutput) UnmarshalJSON(data []byte) error {
	const (
		errFmt = "cannot unmarshal '%s': unknown absolute price oscillator output"
		dqs    = "\"" // Double quote string.
	)

	d := bytes.Trim(data, dqs)
	s := string(d)

	switch s {
	case absolutePriceOscillatorOutputValue:
		*o = AbsolutePriceOscillatorValue
	default:
		return fmt.Errorf(errFmt, s)
	}

	return nil
}
