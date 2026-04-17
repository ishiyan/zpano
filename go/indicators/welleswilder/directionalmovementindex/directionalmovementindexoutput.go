//nolint:dupl
package directionalmovementindex

import (
	"bytes"
	"fmt"
)

// DirectionalMovementIndexOutput describes the outputs of the indicator.
type DirectionalMovementIndexOutput int

const (
	// The scalar value of the directional movement index (DX).
	DirectionalMovementIndexValue DirectionalMovementIndexOutput = iota + 1

	// The scalar value of the directional indicator plus (+DI).
	DirectionalIndicatorPlusValue

	// The scalar value of the directional indicator minus (-DI).
	DirectionalIndicatorMinusValue

	// The scalar value of the directional movement plus (+DM).
	DirectionalMovementPlusValue

	// The scalar value of the directional movement minus (-DM).
	DirectionalMovementMinusValue

	// The scalar value of the average true range (ATR).
	AverageTrueRangeValue

	// The scalar value of the true range (TR).
	TrueRangeValue

	directionalMovementIndexLast
)

const (
	directionalMovementIndexOutputValue                     = "value"
	directionalMovementIndexOutputDirectionalIndicatorPlus  = "directionalIndicatorPlus"
	directionalMovementIndexOutputDirectionalIndicatorMinus = "directionalIndicatorMinus"
	directionalMovementIndexOutputDirectionalMovementPlus   = "directionalMovementPlus"
	directionalMovementIndexOutputDirectionalMovementMinus  = "directionalMovementMinus"
	directionalMovementIndexOutputAverageTrueRange          = "averageTrueRange"
	directionalMovementIndexOutputTrueRange                 = "trueRange"
	directionalMovementIndexOutputUnknown                   = "unknown"
)

// String implements the Stringer interface.
func (o DirectionalMovementIndexOutput) String() string {
	switch o {
	case DirectionalMovementIndexValue:
		return directionalMovementIndexOutputValue
	case DirectionalIndicatorPlusValue:
		return directionalMovementIndexOutputDirectionalIndicatorPlus
	case DirectionalIndicatorMinusValue:
		return directionalMovementIndexOutputDirectionalIndicatorMinus
	case DirectionalMovementPlusValue:
		return directionalMovementIndexOutputDirectionalMovementPlus
	case DirectionalMovementMinusValue:
		return directionalMovementIndexOutputDirectionalMovementMinus
	case AverageTrueRangeValue:
		return directionalMovementIndexOutputAverageTrueRange
	case TrueRangeValue:
		return directionalMovementIndexOutputTrueRange
	default:
		return directionalMovementIndexOutputUnknown
	}
}

// IsKnown determines if this output is known.
func (o DirectionalMovementIndexOutput) IsKnown() bool {
	return o >= DirectionalMovementIndexValue && o < directionalMovementIndexLast
}

// MarshalJSON implements the Marshaler interface.
func (o DirectionalMovementIndexOutput) MarshalJSON() ([]byte, error) {
	const (
		errFmt = "cannot marshal '%s': unknown directional movement index output"
		extra  = 2   // Two bytes for quotes.
		dqc    = '"' // Double quote character.
	)

	s := o.String()
	if s == directionalMovementIndexOutputUnknown {
		return nil, fmt.Errorf(errFmt, s)
	}

	b := make([]byte, 0, len(s)+extra)
	b = append(b, dqc)
	b = append(b, s...)
	b = append(b, dqc)

	return b, nil
}

// UnmarshalJSON implements the Unmarshaler interface.
func (o *DirectionalMovementIndexOutput) UnmarshalJSON(data []byte) error {
	const (
		errFmt = "cannot unmarshal '%s': unknown directional movement index output"
		dqs    = "\"" // Double quote string.
	)

	d := bytes.Trim(data, dqs)
	s := string(d)

	switch s {
	case directionalMovementIndexOutputValue:
		*o = DirectionalMovementIndexValue
	case directionalMovementIndexOutputDirectionalIndicatorPlus:
		*o = DirectionalIndicatorPlusValue
	case directionalMovementIndexOutputDirectionalIndicatorMinus:
		*o = DirectionalIndicatorMinusValue
	case directionalMovementIndexOutputDirectionalMovementPlus:
		*o = DirectionalMovementPlusValue
	case directionalMovementIndexOutputDirectionalMovementMinus:
		*o = DirectionalMovementMinusValue
	case directionalMovementIndexOutputAverageTrueRange:
		*o = AverageTrueRangeValue
	case directionalMovementIndexOutputTrueRange:
		*o = TrueRangeValue
	default:
		return fmt.Errorf(errFmt, s)
	}

	return nil
}
