//nolint:dupl
package macdindex

import (
	"bytes"
	"fmt"
)

// Output describes the outputs of the indicator.
type Output int

const (
	// MACDI is the MACD Index line value, in raw price units (unbounded).
	MACDI Output = iota + 1

	// Signal is the signal-line value: the ul-period EMA of the index.
	Signal

	outputLast
)

const (
	mACDIValueStr  = "macdiValue"
	signalValueStr = "signalValue"
	unknownStr     = "unknown"
)

// String implements the Stringer interface.
func (o Output) String() string {
	switch o {
	case MACDI:
		return mACDIValueStr
	case Signal:
		return signalValueStr
	default:
		return unknownStr
	}
}

// IsKnown determines if this output is known.
func (o Output) IsKnown() bool {
	return o >= MACDI && o < outputLast
}

// MarshalJSON implements the Marshaler interface.
func (o Output) MarshalJSON() ([]byte, error) {
	const (
		errFmt = "cannot marshal '%s': unknown macd index output"
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
		errFmt = "cannot unmarshal '%s': unknown macd index output"
		dqs    = "\"" // Double quote string.
	)

	d := bytes.Trim(data, dqs)
	s := string(d)

	switch s {
	case mACDIValueStr:
		*o = MACDI
	case signalValueStr:
		*o = Signal
	default:
		return fmt.Errorf(errFmt, s)
	}

	return nil
}
