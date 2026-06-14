//nolint:dupl
package schafftrendcycle

import (
	"bytes"
	"fmt"
)

// Output describes the outputs of the indicator.
type Output int

const (
	// STC is the Schaff Trend Cycle oscillator value (range [0, 100]).
	STC Output = iota + 1

	// MACD is the gated MACD line (XMAC) value.
	MACD

	// PF is the first smoothed %D stage value.
	PF

	outputLast
)

const (
	sTCValueStr  = "stcValue"
	mACDValueStr = "macdValue"
	pFValueStr   = "pfValue"
	unknownStr   = "unknown"
)

// String implements the Stringer interface.
func (o Output) String() string {
	switch o {
	case STC:
		return sTCValueStr
	case MACD:
		return mACDValueStr
	case PF:
		return pFValueStr
	default:
		return unknownStr
	}
}

// IsKnown determines if this output is known.
func (o Output) IsKnown() bool {
	return o >= STC && o < outputLast
}

// MarshalJSON implements the Marshaler interface.
func (o Output) MarshalJSON() ([]byte, error) {
	const (
		errFmt = "cannot marshal '%s': unknown schaff trend cycle output"
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
		errFmt = "cannot unmarshal '%s': unknown schaff trend cycle output"
		dqs    = "\"" // Double quote string.
	)

	d := bytes.Trim(data, dqs)
	s := string(d)

	switch s {
	case sTCValueStr:
		*o = STC
	case mACDValueStr:
		*o = MACD
	case pFValueStr:
		*o = PF
	default:
		return fmt.Errorf(errFmt, s)
	}

	return nil
}
