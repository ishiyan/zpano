//nolint:dupl
package t2exponentialmovingaverage

import (
	"bytes"
	"fmt"
)

// T2ExponentialMovingAverageOutput describes the outputs of the indicator.
type T2ExponentialMovingAverageOutput int

const (
	// The scalar value of the T2 exponential moving average.
	T2ExponentialMovingAverageValue T2ExponentialMovingAverageOutput = iota + 1
	t2ExponentialMovingAverageLast
)

const (
	t2ExponentialMovingAverageValueStr   = "value"
	t2ExponentialMovingAverageUnknownStr = "unknown"
)

// String implements the Stringer interface.
func (o T2ExponentialMovingAverageOutput) String() string {
	switch o {
	case T2ExponentialMovingAverageValue:
		return t2ExponentialMovingAverageValueStr
	default:
		return t2ExponentialMovingAverageUnknownStr
	}
}

// IsKnown determines if this output is known.
func (o T2ExponentialMovingAverageOutput) IsKnown() bool {
	return o >= T2ExponentialMovingAverageValue && o < t2ExponentialMovingAverageLast
}

// MarshalJSON implements the Marshaler interface.
func (o T2ExponentialMovingAverageOutput) MarshalJSON() ([]byte, error) {
	const (
		errFmt = "cannot marshal '%s': unknown t2 exponential moving average output"
		extra  = 2   // Two bytes for quotes.
		dqc    = '"' // Double quote character.
	)

	s := o.String()
	if s == t2ExponentialMovingAverageUnknownStr {
		return nil, fmt.Errorf(errFmt, s)
	}

	b := make([]byte, 0, len(s)+extra)
	b = append(b, dqc)
	b = append(b, s...)
	b = append(b, dqc)

	return b, nil
}

// UnmarshalJSON implements the Unmarshaler interface.
func (o *T2ExponentialMovingAverageOutput) UnmarshalJSON(data []byte) error {
	const (
		errFmt = "cannot unmarshal '%s': unknown t2 exponential moving average output"
		dqs    = "\"" // Double quote string.
	)

	d := bytes.Trim(data, dqs)
	s := string(d)

	switch s {
	case t2ExponentialMovingAverageValueStr:
		*o = T2ExponentialMovingAverageValue
	default:
		return fmt.Errorf(errFmt, s)
	}

	return nil
}
