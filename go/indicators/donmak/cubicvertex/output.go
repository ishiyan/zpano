//nolint:dupl
package cubicvertex

import (
	"bytes"
	"fmt"
)

// Output describes the outputs of the indicator.
type Output int

const (
	// BarsToNearTurn is the number of bars to the more imminent turning point.
	BarsToNearTurn Output = iota + 1

	// BarsToFarTurn is the number of bars to the more distant turning point.
	BarsToFarTurn

	outputLast
)

const (
	nearValueStr = "nearValue"
	farValueStr  = "farValue"
	unknownStr   = "unknown"
)

// String implements the Stringer interface.
func (o Output) String() string {
	switch o {
	case BarsToNearTurn:
		return nearValueStr
	case BarsToFarTurn:
		return farValueStr
	default:
		return unknownStr
	}
}

// IsKnown determines if this output is known.
func (o Output) IsKnown() bool {
	return o >= BarsToNearTurn && o < outputLast
}

// MarshalJSON implements the Marshaler interface.
func (o Output) MarshalJSON() ([]byte, error) {
	const (
		errFmt = "cannot marshal '%s': unknown cubic vertex output"
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
		errFmt = "cannot unmarshal '%s': unknown cubic vertex output"
		dqs    = "\"" // Double quote string.
	)

	d := bytes.Trim(data, dqs)
	s := string(d)

	switch s {
	case nearValueStr:
		*o = BarsToNearTurn
	case farValueStr:
		*o = BarsToFarTurn
	default:
		return fmt.Errorf(errFmt, s)
	}

	return nil
}
