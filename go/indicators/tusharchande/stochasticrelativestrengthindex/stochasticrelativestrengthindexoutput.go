//nolint:dupl
package stochasticrelativestrengthindex

import (
	"bytes"
	"fmt"
)

// StochasticRelativeStrengthIndexOutput describes the outputs of the indicator.
type StochasticRelativeStrengthIndexOutput int

const (
	// The Fast-K line of the stochastic RSI.
	StochasticRelativeStrengthIndexFastK StochasticRelativeStrengthIndexOutput = iota + 1

	// The Fast-D line of the stochastic RSI (smoothed Fast-K).
	StochasticRelativeStrengthIndexFastD

	stochasticRelativeStrengthIndexLast
)

const (
	stochasticRelativeStrengthIndexOutputFastK   = "fastK"
	stochasticRelativeStrengthIndexOutputFastD   = "fastD"
	stochasticRelativeStrengthIndexOutputUnknown = "unknown"
)

// String implements the Stringer interface.
func (o StochasticRelativeStrengthIndexOutput) String() string {
	switch o {
	case StochasticRelativeStrengthIndexFastK:
		return stochasticRelativeStrengthIndexOutputFastK
	case StochasticRelativeStrengthIndexFastD:
		return stochasticRelativeStrengthIndexOutputFastD
	default:
		return stochasticRelativeStrengthIndexOutputUnknown
	}
}

// IsKnown determines if this output is known.
func (o StochasticRelativeStrengthIndexOutput) IsKnown() bool {
	return o >= StochasticRelativeStrengthIndexFastK && o < stochasticRelativeStrengthIndexLast
}

// MarshalJSON implements the Marshaler interface.
func (o StochasticRelativeStrengthIndexOutput) MarshalJSON() ([]byte, error) {
	const (
		errFmt = "cannot marshal '%s': unknown stochastic relative strength index output"
		extra  = 2   // Two bytes for quotes.
		dqc    = '"' // Double quote character.
	)

	s := o.String()
	if s == stochasticRelativeStrengthIndexOutputUnknown {
		return nil, fmt.Errorf(errFmt, s)
	}

	b := make([]byte, 0, len(s)+extra)
	b = append(b, dqc)
	b = append(b, s...)
	b = append(b, dqc)

	return b, nil
}

// UnmarshalJSON implements the Unmarshaler interface.
func (o *StochasticRelativeStrengthIndexOutput) UnmarshalJSON(data []byte) error {
	const (
		errFmt = "cannot unmarshal '%s': unknown stochastic relative strength index output"
		dqs    = "\"" // Double quote string.
	)

	d := bytes.Trim(data, dqs)
	s := string(d)

	switch s {
	case stochasticRelativeStrengthIndexOutputFastK:
		*o = StochasticRelativeStrengthIndexFastK
	case stochasticRelativeStrengthIndexOutputFastD:
		*o = StochasticRelativeStrengthIndexFastD
	default:
		return fmt.Errorf(errFmt, s)
	}

	return nil
}
