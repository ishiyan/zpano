//nolint:dupl
package williamspercentr

import (
	"bytes"
	"fmt"
)

// WilliamsPercentROutput describes the outputs of the indicator.
type WilliamsPercentROutput int

const (
	// The scalar value of the Williams %R.
	WilliamsPercentRValue WilliamsPercentROutput = iota + 1
	williamsPercentRLast
)

const (
	williamsPercentROutputValue   = "value"
	williamsPercentROutputUnknown = "unknown"
)

// String implements the Stringer interface.
func (o WilliamsPercentROutput) String() string {
	switch o {
	case WilliamsPercentRValue:
		return williamsPercentROutputValue
	default:
		return williamsPercentROutputUnknown
	}
}

// IsKnown determines if this output is known.
func (o WilliamsPercentROutput) IsKnown() bool {
	return o >= WilliamsPercentRValue && o < williamsPercentRLast
}

// MarshalJSON implements the Marshaler interface.
func (o WilliamsPercentROutput) MarshalJSON() ([]byte, error) {
	const (
		errFmt = "cannot marshal '%s': unknown williams percent r output"
		extra  = 2   // Two bytes for quotes.
		dqc    = '"' // Double quote character.
	)

	s := o.String()
	if s == williamsPercentROutputUnknown {
		return nil, fmt.Errorf(errFmt, s)
	}

	b := make([]byte, 0, len(s)+extra)
	b = append(b, dqc)
	b = append(b, s...)
	b = append(b, dqc)

	return b, nil
}

// UnmarshalJSON implements the Unmarshaler interface.
func (o *WilliamsPercentROutput) UnmarshalJSON(data []byte) error {
	const (
		errFmt = "cannot unmarshal '%s': unknown williams percent r output"
		dqs    = "\"" // Double quote string.
	)

	d := bytes.Trim(data, dqs)
	s := string(d)

	switch s {
	case williamsPercentROutputValue:
		*o = WilliamsPercentRValue
	default:
		return fmt.Errorf(errFmt, s)
	}

	return nil
}
