//nolint:dupl
package averagedirectionalmovementindexrating

import (
	"bytes"
	"fmt"
)

// AverageDirectionalMovementIndexRatingOutput describes the outputs of the indicator.
type AverageDirectionalMovementIndexRatingOutput int

const (
	// The scalar value of the average directional movement index rating (ADXR).
	AverageDirectionalMovementIndexRatingValue AverageDirectionalMovementIndexRatingOutput = iota + 1

	// The scalar value of the average directional movement index (ADX).
	AverageDirectionalMovementIndexValue

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

	averageDirectionalMovementIndexRatingLast
)

const (
	averageDirectionalMovementIndexRatingOutputValue                           = "value"
	averageDirectionalMovementIndexRatingOutputAverageDirectionalMovementIndex = "averageDirectionalMovementIndex"
	averageDirectionalMovementIndexRatingOutputDirectionalMovementIndex        = "directionalMovementIndex"
	averageDirectionalMovementIndexRatingOutputDirectionalIndicatorPlus        = "directionalIndicatorPlus"
	averageDirectionalMovementIndexRatingOutputDirectionalIndicatorMinus       = "directionalIndicatorMinus"
	averageDirectionalMovementIndexRatingOutputDirectionalMovementPlus         = "directionalMovementPlus"
	averageDirectionalMovementIndexRatingOutputDirectionalMovementMinus        = "directionalMovementMinus"
	averageDirectionalMovementIndexRatingOutputAverageTrueRange                = "averageTrueRange"
	averageDirectionalMovementIndexRatingOutputTrueRange                       = "trueRange"
	averageDirectionalMovementIndexRatingOutputUnknown                         = "unknown"
)

// String implements the Stringer interface.
func (o AverageDirectionalMovementIndexRatingOutput) String() string {
	switch o {
	case AverageDirectionalMovementIndexRatingValue:
		return averageDirectionalMovementIndexRatingOutputValue
	case AverageDirectionalMovementIndexValue:
		return averageDirectionalMovementIndexRatingOutputAverageDirectionalMovementIndex
	case DirectionalMovementIndexValue:
		return averageDirectionalMovementIndexRatingOutputDirectionalMovementIndex
	case DirectionalIndicatorPlusValue:
		return averageDirectionalMovementIndexRatingOutputDirectionalIndicatorPlus
	case DirectionalIndicatorMinusValue:
		return averageDirectionalMovementIndexRatingOutputDirectionalIndicatorMinus
	case DirectionalMovementPlusValue:
		return averageDirectionalMovementIndexRatingOutputDirectionalMovementPlus
	case DirectionalMovementMinusValue:
		return averageDirectionalMovementIndexRatingOutputDirectionalMovementMinus
	case AverageTrueRangeValue:
		return averageDirectionalMovementIndexRatingOutputAverageTrueRange
	case TrueRangeValue:
		return averageDirectionalMovementIndexRatingOutputTrueRange
	default:
		return averageDirectionalMovementIndexRatingOutputUnknown
	}
}

// IsKnown determines if this output is known.
func (o AverageDirectionalMovementIndexRatingOutput) IsKnown() bool {
	return o >= AverageDirectionalMovementIndexRatingValue && o < averageDirectionalMovementIndexRatingLast
}

// MarshalJSON implements the Marshaler interface.
func (o AverageDirectionalMovementIndexRatingOutput) MarshalJSON() ([]byte, error) {
	const (
		errFmt = "cannot marshal '%s': unknown average directional movement index rating output"
		extra  = 2   // Two bytes for quotes.
		dqc    = '"' // Double quote character.
	)

	s := o.String()
	if s == averageDirectionalMovementIndexRatingOutputUnknown {
		return nil, fmt.Errorf(errFmt, s)
	}

	b := make([]byte, 0, len(s)+extra)
	b = append(b, dqc)
	b = append(b, s...)
	b = append(b, dqc)

	return b, nil
}

// UnmarshalJSON implements the Unmarshaler interface.
func (o *AverageDirectionalMovementIndexRatingOutput) UnmarshalJSON(data []byte) error {
	const (
		errFmt = "cannot unmarshal '%s': unknown average directional movement index rating output"
		dqs    = "\"" // Double quote string.
	)

	d := bytes.Trim(data, dqs)
	s := string(d)

	switch s {
	case averageDirectionalMovementIndexRatingOutputValue:
		*o = AverageDirectionalMovementIndexRatingValue
	case averageDirectionalMovementIndexRatingOutputAverageDirectionalMovementIndex:
		*o = AverageDirectionalMovementIndexValue
	case averageDirectionalMovementIndexRatingOutputDirectionalMovementIndex:
		*o = DirectionalMovementIndexValue
	case averageDirectionalMovementIndexRatingOutputDirectionalIndicatorPlus:
		*o = DirectionalIndicatorPlusValue
	case averageDirectionalMovementIndexRatingOutputDirectionalIndicatorMinus:
		*o = DirectionalIndicatorMinusValue
	case averageDirectionalMovementIndexRatingOutputDirectionalMovementPlus:
		*o = DirectionalMovementPlusValue
	case averageDirectionalMovementIndexRatingOutputDirectionalMovementMinus:
		*o = DirectionalMovementMinusValue
	case averageDirectionalMovementIndexRatingOutputAverageTrueRange:
		*o = AverageTrueRangeValue
	case averageDirectionalMovementIndexRatingOutputTrueRange:
		*o = TrueRangeValue
	default:
		return fmt.Errorf(errFmt, s)
	}

	return nil
}
