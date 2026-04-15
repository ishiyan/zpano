package jurikmovingaverage

import (
	"bytes"
	"fmt"
)

// JurikMovingAverageOutput describes the outputs of the indicator.
type JurikMovingAverageOutput int

const (
	// The scalar value of the moving average.
	MovingAverageValue JurikMovingAverageOutput = iota + 1
	movingAverageLast
)

const (
	movingAverageValue   = "value"
	movingAverageUnknown = "unknown"
)

// String implements the Stringer interface.
func (o JurikMovingAverageOutput) String() string {
	switch o {
	case MovingAverageValue:
		return movingAverageValue
	default:
		return movingAverageUnknown
	}
}

// IsKnown determines if this output is known.
func (o JurikMovingAverageOutput) IsKnown() bool {
	return o >= MovingAverageValue && o < movingAverageLast
}

// MarshalJSON implements the Marshaler interface.
func (o JurikMovingAverageOutput) MarshalJSON() ([]byte, error) {
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
func (o *JurikMovingAverageOutput) UnmarshalJSON(data []byte) error {
	const (
		errFmt = "cannot unmarshal '%s': unknown jurik moving average output"
		dqs    = "\"" // Double quote string.
	)

	d := bytes.Trim(data, dqs)
	s := string(d)

	switch s {
	case movingAverageValue:
		*o = MovingAverageValue
	default:
		return fmt.Errorf(errFmt, s)
	}

	return nil
}
