//nolint:dupl
package doubleexponentialmovingaverage

import (
	"bytes"
	"fmt"
)

// DoubleExponentialMovingAverageOutput describes the outputs of the indicator.
type DoubleExponentialMovingAverageOutput int

const (
	// The scalar value of the double exponential moving average.
	DoubleExponentialMovingAverageValue DoubleExponentialMovingAverageOutput = iota + 1
	doubleExponentialMovingAverageLast
)

const (
	doubleExponentialMovingAverageValueStr   = "value"
	doubleExponentialMovingAverageUnknownStr = "unknown"
)

// String implements the Stringer interface.
func (o DoubleExponentialMovingAverageOutput) String() string {
	switch o {
	case DoubleExponentialMovingAverageValue:
		return doubleExponentialMovingAverageValueStr
	default:
		return doubleExponentialMovingAverageUnknownStr
	}
}

// IsKnown determines if this output is known.
func (o DoubleExponentialMovingAverageOutput) IsKnown() bool {
	return o >= DoubleExponentialMovingAverageValue && o < doubleExponentialMovingAverageLast
}

// MarshalJSON implements the Marshaler interface.
func (o DoubleExponentialMovingAverageOutput) MarshalJSON() ([]byte, error) {
	const (
		errFmt = "cannot marshal '%s': unknown double exponential moving average output"
		extra  = 2   // Two bytes for quotes.
		dqc    = '"' // Double quote character.
	)

	s := o.String()
	if s == doubleExponentialMovingAverageUnknownStr {
		return nil, fmt.Errorf(errFmt, s)
	}

	b := make([]byte, 0, len(s)+extra)
	b = append(b, dqc)
	b = append(b, s...)
	b = append(b, dqc)

	return b, nil
}

// UnmarshalJSON implements the Unmarshaler interface.
func (o *DoubleExponentialMovingAverageOutput) UnmarshalJSON(data []byte) error {
	const (
		errFmt = "cannot unmarshal '%s': unknown double exponential moving average output"
		dqs    = "\"" // Double quote string.
	)

	d := bytes.Trim(data, dqs)
	s := string(d)

	switch s {
	case doubleExponentialMovingAverageValueStr:
		*o = DoubleExponentialMovingAverageValue
	default:
		return fmt.Errorf(errFmt, s)
	}

	return nil
}
