//nolint:dupl
package t3exponentialmovingaverage

import (
	"bytes"
	"fmt"
)

// T3ExponentialMovingAverageOutput describes the outputs of the indicator.
type T3ExponentialMovingAverageOutput int

const (
	// The scalar value of the T3 exponential moving average.
	T3ExponentialMovingAverageValue T3ExponentialMovingAverageOutput = iota + 1
	t3ExponentialMovingAverageLast
)

const (
	t3ExponentialMovingAverageValueStr   = "value"
	t3ExponentialMovingAverageUnknownStr = "unknown"
)

// String implements the Stringer interface.
func (o T3ExponentialMovingAverageOutput) String() string {
	switch o {
	case T3ExponentialMovingAverageValue:
		return t3ExponentialMovingAverageValueStr
	default:
		return t3ExponentialMovingAverageUnknownStr
	}
}

// IsKnown determines if this output is known.
func (o T3ExponentialMovingAverageOutput) IsKnown() bool {
	return o >= T3ExponentialMovingAverageValue && o < t3ExponentialMovingAverageLast
}

// MarshalJSON implements the Marshaler interface.
func (o T3ExponentialMovingAverageOutput) MarshalJSON() ([]byte, error) {
	const (
		errFmt = "cannot marshal '%s': unknown t3 exponential moving average output"
		extra  = 2   // Two bytes for quotes.
		dqc    = '"' // Double quote character.
	)

	s := o.String()
	if s == t3ExponentialMovingAverageUnknownStr {
		return nil, fmt.Errorf(errFmt, s)
	}

	b := make([]byte, 0, len(s)+extra)
	b = append(b, dqc)
	b = append(b, s...)
	b = append(b, dqc)

	return b, nil
}

// UnmarshalJSON implements the Unmarshaler interface.
func (o *T3ExponentialMovingAverageOutput) UnmarshalJSON(data []byte) error {
	const (
		errFmt = "cannot unmarshal '%s': unknown t3 exponential moving average output"
		dqs    = "\"" // Double quote string.
	)

	d := bytes.Trim(data, dqs)
	s := string(d)

	switch s {
	case t3ExponentialMovingAverageValueStr:
		*o = T3ExponentialMovingAverageValue
	default:
		return fmt.Errorf(errFmt, s)
	}

	return nil
}
