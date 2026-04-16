//nolint:dupl
package roofingfilter

import (
	"bytes"
	"fmt"
)

// RoofingFilterOutput describes the outputs of the indicator.
type RoofingFilterOutput int

const (
	// The scalar value of the roofing filter.
	RoofingFilterValue RoofingFilterOutput = iota + 1
	roofingFilterLast
)

const (
	roofingFilterOutputValue   = "value"
	roofingFilterOutputUnknown = "unknown"
)

// String implements the Stringer interface.
func (o RoofingFilterOutput) String() string {
	switch o {
	case RoofingFilterValue:
		return roofingFilterOutputValue
	default:
		return roofingFilterOutputUnknown
	}
}

// IsKnown determines if this output is known.
func (o RoofingFilterOutput) IsKnown() bool {
	return o >= RoofingFilterValue && o < roofingFilterLast
}

// MarshalJSON implements the Marshaler interface.
func (o RoofingFilterOutput) MarshalJSON() ([]byte, error) {
	const (
		errFmt = "cannot marshal '%s': unknown roofing filter output"
		extra  = 2   // Two bytes for quotes.
		dqc    = '"' // Double quote character.
	)

	s := o.String()
	if s == roofingFilterOutputUnknown {
		return nil, fmt.Errorf(errFmt, s)
	}

	b := make([]byte, 0, len(s)+extra)
	b = append(b, dqc)
	b = append(b, s...)
	b = append(b, dqc)

	return b, nil
}

// UnmarshalJSON implements the Unmarshaler interface.
func (o *RoofingFilterOutput) UnmarshalJSON(data []byte) error {
	const (
		errFmt = "cannot unmarshal '%s': unknown roofing filter output"
		dqs    = "\"" // Double quote string.
	)

	d := bytes.Trim(data, dqs)
	s := string(d)

	switch s {
	case roofingFilterOutputValue:
		*o = RoofingFilterValue
	default:
		return fmt.Errorf(errFmt, s)
	}

	return nil
}
