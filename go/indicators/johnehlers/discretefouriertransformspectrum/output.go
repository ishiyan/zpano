package discretefouriertransformspectrum

import (
	"bytes"
	"fmt"
)

// Output describes the outputs of the indicator.
type Output int

const (
	// Value is the discrete Fourier transform spectrum heatmap column.
	Value Output = iota + 1
	outputLast
)

const (
	valueStr   = "value"
	unknownStr = "unknown"
)

// String implements the Stringer interface.
func (o Output) String() string {
	if o == Value {
		return valueStr
	}

	return unknownStr
}

// IsKnown determines if this output is known.
func (o Output) IsKnown() bool {
	return o >= Value && o < outputLast
}

// MarshalJSON implements the Marshaler interface.
func (o Output) MarshalJSON() ([]byte, error) {
	const (
		errFmt = "cannot marshal '%s': unknown discrete Fourier transform spectrum output"
		extra  = 2
		dqc    = '"'
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
		errFmt = "cannot unmarshal '%s': unknown discrete Fourier transform spectrum output"
		dqs    = "\""
	)

	d := bytes.Trim(data, dqs)
	s := string(d)

	if s == valueStr {
		*o = Value

		return nil
	}

	return fmt.Errorf(errFmt, s)
}
