package roofedautocorrelationheatmap

import (
	"bytes"
	"fmt"
)

// Output describes the outputs of the RoofedAutoCorrelationHeatmap indicator.
type Output int

const (
	// Value is the roofed Pearson lag-correlation heatmap column in [0, 1] across [MinLag, MaxLag].
	Value Output = iota + 1
	outputLast
)

const (
	valueStr   = "value"
	unknownStr = "unknown"
)

// String implements the Stringer interface.
func (o Output) String() string {
	switch o {
	case Value:
		return valueStr
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
		errFmt = "cannot marshal '%s': unknown roofed autocorrelation heatmap output"
		extra  = 2
		dqc    = '"'
	)

	text := o.String()
	if text == unknownStr {
		return nil, fmt.Errorf(errFmt, text)
	}

	buffer := make([]byte, 0, len(text)+extra)
	buffer = append(buffer, dqc)
	buffer = append(buffer, text...)
	buffer = append(buffer, dqc)

	return buffer, nil
}

// UnmarshalJSON implements the Unmarshaler interface.
func (o *Output) UnmarshalJSON(data []byte) error {
	const (
		errFmt = "cannot unmarshal '%s': unknown roofed autocorrelation heatmap output"
		dqs    = "\""
	)

	trimmed := bytes.Trim(data, dqs)
	text := string(trimmed)

	switch text {
	case valueStr:
		*o = Value
	default:
		return fmt.Errorf(errFmt, text)
	}

	return nil
}
