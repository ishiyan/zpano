//nolint:dupl
package directionalindicatorminus

import (
	"bytes"
	"fmt"
)

// Output describes the outputs of the indicator.
type Output int

const (
	// The scalar value of the directional indicator minus (-DI).
	Value Output = iota + 1

	// The scalar value of the directional movement minus (-DM).
	DirectionalMovementMinus

	// The scalar value of the average true range (ATR).
	AverageTrueRange

	// The scalar value of the true range (TR).
	TrueRange

	outputLast
)

const (
	valueStr               = "value"
	directionalMovementStr = "directionalMovementMinus"
	averageTrueRangeStr    = "averageTrueRange"
	trueRangeStr           = "trueRange"
	unknownStr             = "unknown"
)

// String implements the Stringer interface.
func (o Output) String() string {
	switch o {
	case Value:
		return valueStr
	case DirectionalMovementMinus:
		return directionalMovementStr
	case AverageTrueRange:
		return averageTrueRangeStr
	case TrueRange:
		return trueRangeStr
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
		errFmt = "cannot marshal '%s': unknown directional indicator minus output"
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
		errFmt = "cannot unmarshal '%s': unknown directional indicator minus output"
		dqs    = "\"" // Double quote string.
	)

	d := bytes.Trim(data, dqs)
	s := string(d)

	switch s {
	case valueStr:
		*o = Value
	case directionalMovementStr:
		*o = DirectionalMovementMinus
	case averageTrueRangeStr:
		*o = AverageTrueRange
	case trueRangeStr:
		*o = TrueRange
	default:
		return fmt.Errorf(errFmt, s)
	}

	return nil
}
