//nolint:dupl
package directionalindicatorminus

import (
	"bytes"
	"fmt"
)

// DirectionalIndicatorMinusOutput describes the outputs of the indicator.
type DirectionalIndicatorMinusOutput int

const (
	// The scalar value of the directional indicator minus (-DI).
	DirectionalIndicatorMinusValue DirectionalIndicatorMinusOutput = iota + 1

	// The scalar value of the directional movement minus (-DM).
	DirectionalMovementMinusValue

	// The scalar value of the average true range (ATR).
	AverageTrueRangeValue

	// The scalar value of the true range (TR).
	TrueRangeValue

	directionalIndicatorMinusLast
)

const (
	directionalIndicatorMinusOutputValue               = "value"
	directionalIndicatorMinusOutputDirectionalMovement = "directionalMovementMinus"
	directionalIndicatorMinusOutputAverageTrueRange    = "averageTrueRange"
	directionalIndicatorMinusOutputTrueRange           = "trueRange"
	directionalIndicatorMinusOutputUnknown             = "unknown"
)

// String implements the Stringer interface.
func (o DirectionalIndicatorMinusOutput) String() string {
	switch o {
	case DirectionalIndicatorMinusValue:
		return directionalIndicatorMinusOutputValue
	case DirectionalMovementMinusValue:
		return directionalIndicatorMinusOutputDirectionalMovement
	case AverageTrueRangeValue:
		return directionalIndicatorMinusOutputAverageTrueRange
	case TrueRangeValue:
		return directionalIndicatorMinusOutputTrueRange
	default:
		return directionalIndicatorMinusOutputUnknown
	}
}

// IsKnown determines if this output is known.
func (o DirectionalIndicatorMinusOutput) IsKnown() bool {
	return o >= DirectionalIndicatorMinusValue && o < directionalIndicatorMinusLast
}

// MarshalJSON implements the Marshaler interface.
func (o DirectionalIndicatorMinusOutput) MarshalJSON() ([]byte, error) {
	const (
		errFmt = "cannot marshal '%s': unknown directional indicator minus output"
		extra  = 2   // Two bytes for quotes.
		dqc    = '"' // Double quote character.
	)

	s := o.String()
	if s == directionalIndicatorMinusOutputUnknown {
		return nil, fmt.Errorf(errFmt, s)
	}

	b := make([]byte, 0, len(s)+extra)
	b = append(b, dqc)
	b = append(b, s...)
	b = append(b, dqc)

	return b, nil
}

// UnmarshalJSON implements the Unmarshaler interface.
func (o *DirectionalIndicatorMinusOutput) UnmarshalJSON(data []byte) error {
	const (
		errFmt = "cannot unmarshal '%s': unknown directional indicator minus output"
		dqs    = "\"" // Double quote string.
	)

	d := bytes.Trim(data, dqs)
	s := string(d)

	switch s {
	case directionalIndicatorMinusOutputValue:
		*o = DirectionalIndicatorMinusValue
	case directionalIndicatorMinusOutputDirectionalMovement:
		*o = DirectionalMovementMinusValue
	case directionalIndicatorMinusOutputAverageTrueRange:
		*o = AverageTrueRangeValue
	case directionalIndicatorMinusOutputTrueRange:
		*o = TrueRangeValue
	default:
		return fmt.Errorf(errFmt, s)
	}

	return nil
}
