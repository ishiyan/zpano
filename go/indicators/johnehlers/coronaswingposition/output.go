package coronaswingposition

import (
	"bytes"
	"fmt"
)

// Output describes the outputs of the indicator.
type Output int

const (
	// Value is the Corona swing position heatmap column.
	Value Output = iota + 1
	// SwingPosition is the scalar swing position in [MinParameterValue, MaxParameterValue].
	SwingPosition
	outputLast
)

const (
	valueStr         = "value"
	swingPositionStr = "swingPosition"
	unknownStr       = "unknown"
)

// String implements the Stringer interface.
func (o Output) String() string {
	switch o {
	case Value:
		return valueStr
	case SwingPosition:
		return swingPositionStr
	default:
		return unknownStr
	}
}

// IsKnown determines if this output is known.
func (o Output) IsKnown() bool {
	return o >= Value && o < outputLast
}

// MarshalJSON implements the Marshaler interface.
func (o Output) MarshalJSON() ([]byte, error) {
	const (
		errFmt = "cannot marshal '%s': unknown corona swing position output"
		extra  = 2
		dqc    = '"'
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
		errFmt = "cannot unmarshal '%s': unknown corona swing position output"
		dqs    = "\""
	)

	d := bytes.Trim(data, dqs)
	s := string(d)

	switch s {
	case valueStr:
		*o = Value
	case swingPositionStr:
		*o = SwingPosition
	default:
		return fmt.Errorf(errFmt, s)
	}

	return nil
}
