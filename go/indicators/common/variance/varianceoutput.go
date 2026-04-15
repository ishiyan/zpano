//nolint:dupl
package variance

import (
	"bytes"
	"fmt"
)

// VarianceOutput describes the outputs of the indicator.
type VarianceOutput int

const (
	// The scalar value of the the variance.
	VarianceValue VarianceOutput = iota + 1
	varianceLast
)

const (
	varianceValue   = "value"
	varianceUnknown = "unknown"
)

// String implements the Stringer interface.
func (o VarianceOutput) String() string {
	switch o {
	case VarianceValue:
		return varianceValue
	default:
		return varianceUnknown
	}
}

// IsKnown determines if this output is known.
func (o VarianceOutput) IsKnown() bool {
	return o >= VarianceValue && o < varianceLast
}

// MarshalJSON implements the Marshaler interface.
func (o VarianceOutput) MarshalJSON() ([]byte, error) {
	const (
		errFmt = "cannot marshal '%s': unknown variance output"
		extra  = 2   // Two bytes for quotes.
		dqc    = '"' // Double quote character.
	)

	s := o.String()
	if s == varianceUnknown {
		return nil, fmt.Errorf(errFmt, s)
	}

	b := make([]byte, 0, len(s)+extra)
	b = append(b, dqc)
	b = append(b, s...)
	b = append(b, dqc)

	return b, nil
}

// UnmarshalJSON implements the Unmarshaler interface.
func (o *VarianceOutput) UnmarshalJSON(data []byte) error {
	const (
		errFmt = "cannot unmarshal '%s': unknown variance output"
		dqs    = "\"" // Double quote string.
	)

	d := bytes.Trim(data, dqs)
	s := string(d)

	switch s {
	case varianceValue:
		*o = VarianceValue
	default:
		return fmt.Errorf(errFmt, s)
	}

	return nil
}
