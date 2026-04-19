//nolint:dupl
package parabolicstopandreverse

import (
	"bytes"
	"fmt"
)

// ParabolicStopAndReverseOutput describes the outputs of the indicator.
type ParabolicStopAndReverseOutput int

const (
	// The scalar value of the Parabolic Stop And Reverse.
	// Positive values indicate a long position; negative values indicate a short position.
	ParabolicStopAndReverseValue ParabolicStopAndReverseOutput = iota + 1
	parabolicStopAndReverseLast
)

const (
	parabolicStopAndReverseOutputValue   = "value"
	parabolicStopAndReverseOutputUnknown = "unknown"
)

// String implements the Stringer interface.
func (o ParabolicStopAndReverseOutput) String() string {
	switch o {
	case ParabolicStopAndReverseValue:
		return parabolicStopAndReverseOutputValue
	default:
		return parabolicStopAndReverseOutputUnknown
	}
}

// IsKnown determines if this output is known.
func (o ParabolicStopAndReverseOutput) IsKnown() bool {
	return o >= ParabolicStopAndReverseValue && o < parabolicStopAndReverseLast
}

// MarshalJSON implements the Marshaler interface.
func (o ParabolicStopAndReverseOutput) MarshalJSON() ([]byte, error) {
	const (
		errFmt = "cannot marshal '%s': unknown parabolic stop and reverse output"
		extra  = 2   // Two bytes for quotes.
		dqc    = '"' // Double quote character.
	)

	s := o.String()
	if s == parabolicStopAndReverseOutputUnknown {
		return nil, fmt.Errorf(errFmt, s)
	}

	b := make([]byte, 0, len(s)+extra)
	b = append(b, dqc)
	b = append(b, s...)
	b = append(b, dqc)

	return b, nil
}

// UnmarshalJSON implements the Unmarshaler interface.
func (o *ParabolicStopAndReverseOutput) UnmarshalJSON(data []byte) error {
	const (
		errFmt = "cannot unmarshal '%s': unknown parabolic stop and reverse output"
		dqs    = "\"" // Double quote string.
	)

	d := bytes.Trim(data, dqs)
	s := string(d)

	switch s {
	case parabolicStopAndReverseOutputValue:
		*o = ParabolicStopAndReverseValue
	default:
		return fmt.Errorf(errFmt, s)
	}

	return nil
}
