//nolint:dupl
package zerolagerrorcorrectingexponentialmovingaverage

import (
	"bytes"
	"fmt"
)

// ZeroLagErrorCorrectingExponentialMovingAverageOutput describes the outputs of the indicator.
type ZeroLagErrorCorrectingExponentialMovingAverageOutput int

const (
	// The scalar value of the zero-lag error-correcting exponential moving average.
	ZeroLagErrorCorrectingExponentialMovingAverageValue ZeroLagErrorCorrectingExponentialMovingAverageOutput = iota + 1
	zeroLagErrorCorrectingExponentialMovingAverageLast
)

const (
	zeroLagErrorCorrectingExponentialMovingAverageOutputValue   = "value"
	zeroLagErrorCorrectingExponentialMovingAverageOutputUnknown = "unknown"
)

// String implements the Stringer interface.
func (o ZeroLagErrorCorrectingExponentialMovingAverageOutput) String() string {
	switch o {
	case ZeroLagErrorCorrectingExponentialMovingAverageValue:
		return zeroLagErrorCorrectingExponentialMovingAverageOutputValue
	default:
		return zeroLagErrorCorrectingExponentialMovingAverageOutputUnknown
	}
}

// IsKnown determines if this output is known.
func (o ZeroLagErrorCorrectingExponentialMovingAverageOutput) IsKnown() bool {
	return o >= ZeroLagErrorCorrectingExponentialMovingAverageValue && o < zeroLagErrorCorrectingExponentialMovingAverageLast
}

// MarshalJSON implements the Marshaler interface.
func (o ZeroLagErrorCorrectingExponentialMovingAverageOutput) MarshalJSON() ([]byte, error) {
	const (
		errFmt = "cannot marshal '%s': unknown zero-lag error-correcting exponential moving average output"
		extra  = 2   // Two bytes for quotes.
		dqc    = '"' // Double quote character.
	)

	s := o.String()
	if s == zeroLagErrorCorrectingExponentialMovingAverageOutputUnknown {
		return nil, fmt.Errorf(errFmt, s)
	}

	b := make([]byte, 0, len(s)+extra)
	b = append(b, dqc)
	b = append(b, s...)
	b = append(b, dqc)

	return b, nil
}

// UnmarshalJSON implements the Unmarshaler interface.
func (o *ZeroLagErrorCorrectingExponentialMovingAverageOutput) UnmarshalJSON(data []byte) error {
	const (
		errFmt = "cannot unmarshal '%s': unknown zero-lag error-correcting exponential moving average output"
		dqs    = "\"" // Double quote string.
	)

	d := bytes.Trim(data, dqs)
	s := string(d)

	switch s {
	case zeroLagErrorCorrectingExponentialMovingAverageOutputValue:
		*o = ZeroLagErrorCorrectingExponentialMovingAverageValue
	default:
		return fmt.Errorf(errFmt, s)
	}

	return nil
}
