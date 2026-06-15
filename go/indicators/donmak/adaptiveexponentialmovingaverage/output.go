//nolint:dupl
package adaptiveexponentialmovingaverage

import (
	"bytes"
	"fmt"
)

// Output describes the outputs of the indicator.
type Output int

const (
	// Value is the adaptively smoothed price value.
	Value Output = iota + 1

	// Omega is the instantaneous frequency estimate (may be NaN).
	Omega

	// Alpha is the smoothing factor used for the bar.
	Alpha

	outputLast
)

const (
	valueStr   = "value"
	omegaStr   = "omega"
	alphaStr   = "alpha"
	unknownStr = "unknown"
)

// String implements the Stringer interface.
func (o Output) String() string {
	switch o {
	case Value:
		return valueStr
	case Omega:
		return omegaStr
	case Alpha:
		return alphaStr
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
		errFmt = "cannot marshal '%s': unknown adaptive exponential moving average output"
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
		errFmt = "cannot unmarshal '%s': unknown adaptive exponential moving average output"
		dqs    = "\"" // Double quote string.
	)

	d := bytes.Trim(data, dqs)
	s := string(d)

	switch s {
	case valueStr:
		*o = Value
	case omegaStr:
		*o = Omega
	case alphaStr:
		*o = Alpha
	default:
		return fmt.Errorf(errFmt, s)
	}

	return nil
}
