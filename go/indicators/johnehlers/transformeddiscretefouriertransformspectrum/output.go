package transformeddiscretefouriertransformspectrum

import (
	"bytes"
	"fmt"
)

// Output describes the outputs of the TransformedDiscreteFourierTransformSpectrum indicator.
type Output int

const (
	// Value is the transformed discrete Fourier transform spectrum heatmap column in dB.
	Value Output = iota + 1
	// DominantCycle is the 3 dB center-of-gravity dominant cycle period estimate in bars.
	DominantCycle
	outputLast
)

const (
	valueStr         = "value"
	dominantCycleStr = "dominantCycle"
	unknownStr       = "unknown"
)

// String implements the Stringer interface.
func (o Output) String() string {
	switch o {
	case Value:
		return valueStr
	case DominantCycle:
		return dominantCycleStr
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
		errFmt = "cannot marshal '%s': unknown transformed discrete Fourier transform spectrum output"
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
		errFmt = "cannot unmarshal '%s': unknown transformed discrete Fourier transform spectrum output"
		dqs    = "\""
	)

	trimmed := bytes.Trim(data, dqs)
	text := string(trimmed)

	switch text {
	case valueStr:
		*o = Value
	case dominantCycleStr:
		*o = DominantCycle
	default:
		return fmt.Errorf(errFmt, text)
	}

	return nil
}
