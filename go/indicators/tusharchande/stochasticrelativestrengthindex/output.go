//nolint:dupl
package stochasticrelativestrengthindex

import (
	"bytes"
	"fmt"
)

// Output describes the outputs of the indicator.
type Output int

const (
	// The Fast-K line of the stochastic RSI.
	FastK Output = iota + 1

	// The Fast-D line of the stochastic RSI (smoothed Fast-K).
	FastD

	outputLast
)

const (
	fastKStr   = "fastK"
	fastDStr   = "fastD"
	unknownStr = "unknown"
)

// String implements the Stringer interface.
func (o Output) String() string {
	switch o {
	case FastK:
		return fastKStr
	case FastD:
		return fastDStr
	default:
		return unknownStr
	}
}

// IsKnown determines if this output is known.
func (o Output) IsKnown() bool {
	return o >= FastK && o < outputLast
}

// MarshalJSON implements the Marshaler interface.
func (o Output) MarshalJSON() ([]byte, error) {
	const (
		errFmt = "cannot marshal '%s': unknown stochastic relative strength index output"
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
		errFmt = "cannot unmarshal '%s': unknown stochastic relative strength index output"
		dqs    = "\"" // Double quote string.
	)

	d := bytes.Trim(data, dqs)
	s := string(d)

	switch s {
	case fastKStr:
		*o = FastK
	case fastDStr:
		*o = FastD
	default:
		return fmt.Errorf(errFmt, s)
	}

	return nil
}
