package jurikdirectionalmovementindex

import (
	"bytes"
	"fmt"
)

// Output describes the outputs of the indicator.
type Output int

const (
	// Bipolar is the bipolar directional movement index value: 100*(Plus-Minus)/(Plus+Minus).
	Bipolar Output = iota + 1
	// Plus is the positive directional movement index value.
	Plus
	// Minus is the negative directional movement index value.
	Minus
	outputLast
)

const (
	bipolarStr = "bipolar"
	plusStr    = "plus"
	minusStr   = "minus"
	unknownStr = "unknown"
)

// String implements the Stringer interface.
func (o Output) String() string {
	switch o {
	case Bipolar:
		return bipolarStr
	case Plus:
		return plusStr
	case Minus:
		return minusStr
	default:
		return unknownStr
	}
}

// IsKnown determines if this output is known.
func (o Output) IsKnown() bool {
	return o >= Bipolar && o < outputLast
}

// MarshalJSON implements the Marshaler interface.
func (o Output) MarshalJSON() ([]byte, error) {
	const (
		errFmt = "cannot marshal '%s': unknown jurik directional movement index output"
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
		errFmt = "cannot unmarshal '%s': unknown jurik directional movement index output"
		dqs    = "\"" // Double quote string.
	)

	d := bytes.Trim(data, dqs)
	s := string(d)

	switch s {
	case bipolarStr:
		*o = Bipolar
	case plusStr:
		*o = Plus
	case minusStr:
		*o = Minus
	default:
		return fmt.Errorf(errFmt, s)
	}

	return nil
}
