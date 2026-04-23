package coronasignaltonoiseratio

import (
	"bytes"
	"fmt"
)

// Output describes the outputs of the indicator.
type Output int

const (
	// Value is the Corona signal-to-noise ratio heatmap column.
	Value Output = iota + 1
	// SignalToNoiseRatio is the scalar signal-to-noise ratio value in the range of [MinParameterValue, MaxParameterValue].
	SignalToNoiseRatio
	outputLast
)

const (
	valueStr              = "value"
	signalToNoiseRatioStr = "signalToNoiseRatio"
	unknownStr            = "unknown"
)

// String implements the Stringer interface.
func (o Output) String() string {
	switch o {
	case Value:
		return valueStr
	case SignalToNoiseRatio:
		return signalToNoiseRatioStr
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
		errFmt = "cannot marshal '%s': unknown corona signal to noise ratio output"
		extra  = 2   // Two bytes for quotes.
		dqc    = '"' // Double quote character.
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
		errFmt = "cannot unmarshal '%s': unknown corona signal to noise ratio output"
		dqs    = "\"" // Double quote string.
	)

	d := bytes.Trim(data, dqs)
	s := string(d)

	switch s {
	case valueStr:
		*o = Value
	case signalToNoiseRatioStr:
		*o = SignalToNoiseRatio
	default:
		return fmt.Errorf(errFmt, s)
	}

	return nil
}
