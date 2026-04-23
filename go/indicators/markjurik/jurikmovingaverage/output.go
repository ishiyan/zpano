package jurikmovingaverage

import (
	"bytes"
	"fmt"
)

// Output describes the outputs of the indicator.
type Output int

const (
	// The scalar value of the moving average.
	MovingAverage Output = iota + 1
	movingAverageLast
)

const (
	movingAverageValue   = "value"
	movingAverageUnknown = "unknown"
)

// String implements the Stringer interface.
func (o Output) String() string {
	switch o {
	case MovingAverage:
		return movingAverageValue
	default:
		return movingAverageUnknown
	}
}

// IsKnown determines if this output is known.
func (o Output) IsKnown() bool {
	return o >= MovingAverage && o < movingAverageLast
}

// MarshalJSON implements the Marshaler interface.
func (o Output) MarshalJSON() ([]byte, error) {
	const (
		errFmt = "cannot marshal '%s': unknown jurik moving average output"
		extra  = 2   // Two bytes for quotes.
		dqc    = '"' // Double quote character.
	)

	s := o.String()
	if s == movingAverageUnknown {
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
		errFmt = "cannot unmarshal '%s': unknown jurik moving average output"
		dqs    = "\"" // Double quote string.
	)

	d := bytes.Trim(data, dqs)
	s := string(d)

	switch s {
	case movingAverageValue:
		*o = MovingAverage
	default:
		return fmt.Errorf(errFmt, s)
	}

	return nil
}
