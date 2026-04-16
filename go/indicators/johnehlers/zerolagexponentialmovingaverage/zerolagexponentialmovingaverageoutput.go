//nolint:dupl
package zerolagexponentialmovingaverage

import (
	"bytes"
	"fmt"
)

// ZeroLagExponentialMovingAverageOutput describes the outputs of the indicator.
type ZeroLagExponentialMovingAverageOutput int

const (
	// The scalar value of the zero-lag exponential moving average.
	ZeroLagExponentialMovingAverageValue ZeroLagExponentialMovingAverageOutput = iota + 1
	zeroLagExponentialMovingAverageLast
)

const (
	zeroLagExponentialMovingAverageOutputValue   = "value"
	zeroLagExponentialMovingAverageOutputUnknown = "unknown"
)

// String implements the Stringer interface.
func (o ZeroLagExponentialMovingAverageOutput) String() string {
	switch o {
	case ZeroLagExponentialMovingAverageValue:
		return zeroLagExponentialMovingAverageOutputValue
	default:
		return zeroLagExponentialMovingAverageOutputUnknown
	}
}

// IsKnown determines if this output is known.
func (o ZeroLagExponentialMovingAverageOutput) IsKnown() bool {
	return o >= ZeroLagExponentialMovingAverageValue && o < zeroLagExponentialMovingAverageLast
}

// MarshalJSON implements the Marshaler interface.
func (o ZeroLagExponentialMovingAverageOutput) MarshalJSON() ([]byte, error) {
	const (
		errFmt = "cannot marshal '%s': unknown zero-lag exponential moving average output"
		extra  = 2   // Two bytes for quotes.
		dqc    = '"' // Double quote character.
	)

	s := o.String()
	if s == zeroLagExponentialMovingAverageOutputUnknown {
		return nil, fmt.Errorf(errFmt, s)
	}

	b := make([]byte, 0, len(s)+extra)
	b = append(b, dqc)
	b = append(b, s...)
	b = append(b, dqc)

	return b, nil
}

// UnmarshalJSON implements the Unmarshaler interface.
func (o *ZeroLagExponentialMovingAverageOutput) UnmarshalJSON(data []byte) error {
	const (
		errFmt = "cannot unmarshal '%s': unknown zero-lag exponential moving average output"
		dqs    = "\"" // Double quote string.
	)

	d := bytes.Trim(data, dqs)
	s := string(d)

	switch s {
	case zeroLagExponentialMovingAverageOutputValue:
		*o = ZeroLagExponentialMovingAverageValue
	default:
		return fmt.Errorf(errFmt, s)
	}

	return nil
}
