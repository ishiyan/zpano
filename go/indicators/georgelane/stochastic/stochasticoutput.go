//nolint:dupl
package stochastic

import (
	"bytes"
	"fmt"
)

// StochasticOutput describes the outputs of the indicator.
type StochasticOutput int

const (
	// The Fast-K line (raw stochastic).
	StochasticFastK StochasticOutput = iota + 1

	// The Slow-K line (smoothed Fast-K, also known as Fast-D).
	StochasticSlowK

	// The Slow-D line (smoothed Slow-K).
	StochasticSlowD

	stochasticLast
)

const (
	stochasticOutputFastK   = "fastK"
	stochasticOutputSlowK   = "slowK"
	stochasticOutputSlowD   = "slowD"
	stochasticOutputUnknown = "unknown"
)

// String implements the Stringer interface.
func (o StochasticOutput) String() string {
	switch o {
	case StochasticFastK:
		return stochasticOutputFastK
	case StochasticSlowK:
		return stochasticOutputSlowK
	case StochasticSlowD:
		return stochasticOutputSlowD
	default:
		return stochasticOutputUnknown
	}
}

// IsKnown determines if this output is known.
func (o StochasticOutput) IsKnown() bool {
	return o >= StochasticFastK && o < stochasticLast
}

// MarshalJSON implements the Marshaler interface.
func (o StochasticOutput) MarshalJSON() ([]byte, error) {
	const (
		errFmt = "cannot marshal '%s': unknown stochastic output"
		extra  = 2   // Two bytes for quotes.
		dqc    = '"' // Double quote character.
	)

	s := o.String()
	if s == stochasticOutputUnknown {
		return nil, fmt.Errorf(errFmt, s)
	}

	b := make([]byte, 0, len(s)+extra)
	b = append(b, dqc)
	b = append(b, s...)
	b = append(b, dqc)

	return b, nil
}

// UnmarshalJSON implements the Unmarshaler interface.
func (o *StochasticOutput) UnmarshalJSON(data []byte) error {
	const (
		errFmt = "cannot unmarshal '%s': unknown stochastic output"
		dqs    = "\"" // Double quote string.
	)

	d := bytes.Trim(data, dqs)
	s := string(d)

	switch s {
	case stochasticOutputFastK:
		*o = StochasticFastK
	case stochasticOutputSlowK:
		*o = StochasticSlowK
	case stochasticOutputSlowD:
		*o = StochasticSlowD
	default:
		return fmt.Errorf(errFmt, s)
	}

	return nil
}
