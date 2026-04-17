//nolint:dupl
package averagedirectionalmovementindex

import (
	"bytes"
	"fmt"
)

// AverageDirectionalMovementIndexOutput describes the outputs of the indicator.
type AverageDirectionalMovementIndexOutput int

const (
	// The scalar value of the average directional movement index (ADX).
	AverageDirectionalMovementIndexValue AverageDirectionalMovementIndexOutput = iota + 1

	// The scalar value of the directional movement index (DX).
	DirectionalMovementIndexValue

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

	averageDirectionalMovementIndexLast
)

const (
	averageDirectionalMovementIndexOutputValue                     = "value"
	averageDirectionalMovementIndexOutputDirectionalMovementIndex  = "directionalMovementIndex"
	averageDirectionalMovementIndexOutputDirectionalIndicatorPlus  = "directionalIndicatorPlus"
	averageDirectionalMovementIndexOutputDirectionalIndicatorMinus = "directionalIndicatorMinus"
	averageDirectionalMovementIndexOutputDirectionalMovementPlus   = "directionalMovementPlus"
	averageDirectionalMovementIndexOutputDirectionalMovementMinus  = "directionalMovementMinus"
	averageDirectionalMovementIndexOutputAverageTrueRange          = "averageTrueRange"
	averageDirectionalMovementIndexOutputTrueRange                 = "trueRange"
	averageDirectionalMovementIndexOutputUnknown                   = "unknown"
)

// String implements the Stringer interface.
func (o AverageDirectionalMovementIndexOutput) String() string {
	switch o {
	case AverageDirectionalMovementIndexValue:
		return averageDirectionalMovementIndexOutputValue
	case DirectionalMovementIndexValue:
		return averageDirectionalMovementIndexOutputDirectionalMovementIndex
	case DirectionalIndicatorPlusValue:
		return averageDirectionalMovementIndexOutputDirectionalIndicatorPlus
	case DirectionalIndicatorMinusValue:
		return averageDirectionalMovementIndexOutputDirectionalIndicatorMinus
	case DirectionalMovementPlusValue:
		return averageDirectionalMovementIndexOutputDirectionalMovementPlus
	case DirectionalMovementMinusValue:
		return averageDirectionalMovementIndexOutputDirectionalMovementMinus
	case AverageTrueRangeValue:
		return averageDirectionalMovementIndexOutputAverageTrueRange
	case TrueRangeValue:
		return averageDirectionalMovementIndexOutputTrueRange
	default:
		return averageDirectionalMovementIndexOutputUnknown
	}
}

// IsKnown determines if this output is known.
func (o AverageDirectionalMovementIndexOutput) IsKnown() bool {
	return o >= AverageDirectionalMovementIndexValue && o < averageDirectionalMovementIndexLast
}

// MarshalJSON implements the Marshaler interface.
func (o AverageDirectionalMovementIndexOutput) MarshalJSON() ([]byte, error) {
	const (
		errFmt = "cannot marshal '%s': unknown average directional movement index output"
		extra  = 2   // Two bytes for quotes.
		dqc    = '"' // Double quote character.
	)

	s := o.String()
	if s == averageDirectionalMovementIndexOutputUnknown {
		return nil, fmt.Errorf(errFmt, s)
	}

	b := make([]byte, 0, len(s)+extra)
	b = append(b, dqc)
	b = append(b, s...)
	b = append(b, dqc)

	return b, nil
}

// UnmarshalJSON implements the Unmarshaler interface.
func (o *AverageDirectionalMovementIndexOutput) UnmarshalJSON(data []byte) error {
	const (
		errFmt = "cannot unmarshal '%s': unknown average directional movement index output"
		dqs    = "\"" // Double quote string.
	)

	d := bytes.Trim(data, dqs)
	s := string(d)

	switch s {
	case averageDirectionalMovementIndexOutputValue:
		*o = AverageDirectionalMovementIndexValue
	case averageDirectionalMovementIndexOutputDirectionalMovementIndex:
		*o = DirectionalMovementIndexValue
	case averageDirectionalMovementIndexOutputDirectionalIndicatorPlus:
		*o = DirectionalIndicatorPlusValue
	case averageDirectionalMovementIndexOutputDirectionalIndicatorMinus:
		*o = DirectionalIndicatorMinusValue
	case averageDirectionalMovementIndexOutputDirectionalMovementPlus:
		*o = DirectionalMovementPlusValue
	case averageDirectionalMovementIndexOutputDirectionalMovementMinus:
		*o = DirectionalMovementMinusValue
	case averageDirectionalMovementIndexOutputAverageTrueRange:
		*o = AverageTrueRangeValue
	case averageDirectionalMovementIndexOutputTrueRange:
		*o = TrueRangeValue
	default:
		return fmt.Errorf(errFmt, s)
	}

	return nil
}
