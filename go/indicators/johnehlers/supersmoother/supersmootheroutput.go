//nolint:dupl
package supersmoother

import (
	"bytes"
	"fmt"
)

// SuperSmootherOutput describes the outputs of the indicator.
type SuperSmootherOutput int

const (
	// The scalar value of the super smoother.
	SuperSmootherValue SuperSmootherOutput = iota + 1
	superSmootherLast
)

const (
	superSmootherOutputValue   = "value"
	superSmootherOutputUnknown = "unknown"
)

// String implements the Stringer interface.
func (o SuperSmootherOutput) String() string {
	switch o {
	case SuperSmootherValue:
		return superSmootherOutputValue
	default:
		return superSmootherOutputUnknown
	}
}

// IsKnown determines if this output is known.
func (o SuperSmootherOutput) IsKnown() bool {
	return o >= SuperSmootherValue && o < superSmootherLast
}

// MarshalJSON implements the Marshaler interface.
func (o SuperSmootherOutput) MarshalJSON() ([]byte, error) {
	const (
		errFmt = "cannot marshal '%s': unknown super smoother output"
		extra  = 2   // Two bytes for quotes.
		dqc    = '"' // Double quote character.
	)

	s := o.String()
	if s == superSmootherOutputUnknown {
		return nil, fmt.Errorf(errFmt, s)
	}

	b := make([]byte, 0, len(s)+extra)
	b = append(b, dqc)
	b = append(b, s...)
	b = append(b, dqc)

	return b, nil
}

// UnmarshalJSON implements the Unmarshaler interface.
func (o *SuperSmootherOutput) UnmarshalJSON(data []byte) error {
	const (
		errFmt = "cannot unmarshal '%s': unknown super smoother output"
		dqs    = "\"" // Double quote string.
	)

	d := bytes.Trim(data, dqs)
	s := string(d)

	switch s {
	case superSmootherOutputValue:
		*o = SuperSmootherValue
	default:
		return fmt.Errorf(errFmt, s)
	}

	return nil
}
