package sinewave

import (
	"bytes"
	"fmt"
)

// Output describes the outputs of the indicator.
type Output int

const (
	// Value is the sine wave value, sin(phase·Deg2Rad).
	Value Output = iota + 1
	// Lead is the sine wave lead value, sin((phase+45)·Deg2Rad).
	Lead
	// Band is the band formed by the sine wave (upper) and the lead sine wave (lower).
	Band
	// DominantCyclePeriod is the smoothed dominant cycle period.
	DominantCyclePeriod
	// DominantCyclePhase is the dominant cycle phase, in degrees.
	DominantCyclePhase
	outputLast
)

const (
	valueStr               = "value"
	leadStr                = "lead"
	bandStr                = "band"
	dominantCyclePeriodStr = "dominantCyclePeriod"
	dominantCyclePhaseStr  = "dominantCyclePhase"
	unknownStr             = "unknown"
)

// String implements the Stringer interface.
func (o Output) String() string {
	switch o {
	case Value:
		return valueStr
	case Lead:
		return leadStr
	case Band:
		return bandStr
	case DominantCyclePeriod:
		return dominantCyclePeriodStr
	case DominantCyclePhase:
		return dominantCyclePhaseStr
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
		errFmt = "cannot marshal '%s': unknown sine wave output"
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
		errFmt = "cannot unmarshal '%s': unknown sine wave output"
		dqs    = "\"" // Double quote string.
	)

	d := bytes.Trim(data, dqs)
	s := string(d)

	switch s {
	case valueStr:
		*o = Value
	case leadStr:
		*o = Lead
	case bandStr:
		*o = Band
	case dominantCyclePeriodStr:
		*o = DominantCyclePeriod
	case dominantCyclePhaseStr:
		*o = DominantCyclePhase
	default:
		return fmt.Errorf(errFmt, s)
	}

	return nil
}
