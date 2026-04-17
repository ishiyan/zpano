//nolint:dupl
package percentagepriceoscillator

import (
	"bytes"
	"fmt"
)

// PercentagePriceOscillatorOutput describes the outputs of the indicator.
type PercentagePriceOscillatorOutput int

const (
	// The scalar value of the percentage price oscillator.
	PercentagePriceOscillatorValue PercentagePriceOscillatorOutput = iota + 1
	percentagePriceOscillatorLast
)

const (
	percentagePriceOscillatorOutputValue   = "value"
	percentagePriceOscillatorOutputUnknown = "unknown"
)

// String implements the Stringer interface.
func (o PercentagePriceOscillatorOutput) String() string {
	switch o {
	case PercentagePriceOscillatorValue:
		return percentagePriceOscillatorOutputValue
	default:
		return percentagePriceOscillatorOutputUnknown
	}
}

// IsKnown determines if this output is known.
func (o PercentagePriceOscillatorOutput) IsKnown() bool {
	return o >= PercentagePriceOscillatorValue && o < percentagePriceOscillatorLast
}

// MarshalJSON implements the Marshaler interface.
func (o PercentagePriceOscillatorOutput) MarshalJSON() ([]byte, error) {
	const (
		errFmt = "cannot marshal '%s': unknown percentage price oscillator output"
		extra  = 2   // Two bytes for quotes.
		dqc    = '"' // Double quote character.
	)

	s := o.String()
	if s == percentagePriceOscillatorOutputUnknown {
		return nil, fmt.Errorf(errFmt, s)
	}

	b := make([]byte, 0, len(s)+extra)
	b = append(b, dqc)
	b = append(b, s...)
	b = append(b, dqc)

	return b, nil
}

// UnmarshalJSON implements the Unmarshaler interface.
func (o *PercentagePriceOscillatorOutput) UnmarshalJSON(data []byte) error {
	const (
		errFmt = "cannot unmarshal '%s': unknown percentage price oscillator output"
		dqs    = "\"" // Double quote string.
	)

	d := bytes.Trim(data, dqs)
	s := string(d)

	switch s {
	case percentagePriceOscillatorOutputValue:
		*o = PercentagePriceOscillatorValue
	default:
		return fmt.Errorf(errFmt, s)
	}

	return nil
}
