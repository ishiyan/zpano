package mesaadaptivemovingaverage

import (
	"bytes"
	"fmt"
)

// Output describes the outputs of the indicator.
type Output int

const (
	// Value is the scalar value of the MAMA (Mesa Adaptive Moving Average).
	Value Output = iota + 1
	// Fama is the scalar value of the FAMA (Following Adaptive Moving Average).
	Fama
	// Band is the band output, with MAMA as the upper line and FAMA as the lower line.
	Band
	outputLast
)

const (
	valueStr   = "value"
	famaStr    = "fama"
	bandStr    = "band"
	unknownStr = "unknown"
)

// String implements the Stringer interface.
func (o Output) String() string {
	switch o {
	case Value:
		return valueStr
	case Fama:
		return famaStr
	case Band:
		return bandStr
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
		errFmt = "cannot marshal '%s': unknown mesa adaptive moving average output"
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
		errFmt = "cannot unmarshal '%s': unknown mesa adaptive moving average output"
		dqs    = "\"" // Double quote string.
	)

	d := bytes.Trim(data, dqs)
	s := string(d)

	switch s {
	case valueStr:
		*o = Value
	case famaStr:
		*o = Fama
	case bandStr:
		*o = Band
	default:
		return fmt.Errorf(errFmt, s)
	}

	return nil
}
