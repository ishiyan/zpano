//nolint:dupl
package directionalindicatorplus

import (
	"bytes"
	"fmt"
)

// DirectionalIndicatorPlusOutput describes the outputs of the indicator.
type DirectionalIndicatorPlusOutput int

const (
	// The scalar value of the directional indicator plus (+DI).
	DirectionalIndicatorPlusValue DirectionalIndicatorPlusOutput = iota + 1

	// The scalar value of the directional movement plus (+DM).
	DirectionalMovementPlusValue

	// The scalar value of the average true range (ATR).
	AverageTrueRangeValue

	// The scalar value of the true range (TR).
	TrueRangeValue

	directionalIndicatorPlusLast
)

const (
	directionalIndicatorPlusOutputValue               = "value"
	directionalIndicatorPlusOutputDirectionalMovement = "directionalMovementPlus"
	directionalIndicatorPlusOutputAverageTrueRange    = "averageTrueRange"
	directionalIndicatorPlusOutputTrueRange           = "trueRange"
	directionalIndicatorPlusOutputUnknown             = "unknown"
)

// String implements the Stringer interface.
func (o DirectionalIndicatorPlusOutput) String() string {
	switch o {
	case DirectionalIndicatorPlusValue:
		return directionalIndicatorPlusOutputValue
	case DirectionalMovementPlusValue:
		return directionalIndicatorPlusOutputDirectionalMovement
	case AverageTrueRangeValue:
		return directionalIndicatorPlusOutputAverageTrueRange
	case TrueRangeValue:
		return directionalIndicatorPlusOutputTrueRange
	default:
		return directionalIndicatorPlusOutputUnknown
	}
}

// IsKnown determines if this output is known.
func (o DirectionalIndicatorPlusOutput) IsKnown() bool {
	return o >= DirectionalIndicatorPlusValue && o < directionalIndicatorPlusLast
}

// MarshalJSON implements the Marshaler interface.
func (o DirectionalIndicatorPlusOutput) MarshalJSON() ([]byte, error) {
	const (
		errFmt = "cannot marshal '%s': unknown directional indicator plus output"
		extra  = 2   // Two bytes for quotes.
		dqc    = '"' // Double quote character.
	)

	s := o.String()
	if s == directionalIndicatorPlusOutputUnknown {
		return nil, fmt.Errorf(errFmt, s)
	}

	b := make([]byte, 0, len(s)+extra)
	b = append(b, dqc)
	b = append(b, s...)
	b = append(b, dqc)

	return b, nil
}

// UnmarshalJSON implements the Unmarshaler interface.
func (o *DirectionalIndicatorPlusOutput) UnmarshalJSON(data []byte) error {
	const (
		errFmt = "cannot unmarshal '%s': unknown directional indicator plus output"
		dqs    = "\"" // Double quote string.
	)

	d := bytes.Trim(data, dqs)
	s := string(d)

	switch s {
	case directionalIndicatorPlusOutputValue:
		*o = DirectionalIndicatorPlusValue
	case directionalIndicatorPlusOutputDirectionalMovement:
		*o = DirectionalMovementPlusValue
	case directionalIndicatorPlusOutputAverageTrueRange:
		*o = AverageTrueRangeValue
	case directionalIndicatorPlusOutputTrueRange:
		*o = TrueRangeValue
	default:
		return fmt.Errorf(errFmt, s)
	}

	return nil
}
