package trendcyclemode

import (
	"bytes"
	"fmt"
)

// Output describes the outputs of the indicator.
type Output int

const (
	// Value is the Trend-versus-Cycle Mode value: +1 in trend mode, -1 in cycle mode.
	Value Output = iota + 1
	// IsTrendMode is 1 if the trend mode is declared, 0 otherwise.
	IsTrendMode
	// IsCycleMode is 1 if the cycle mode is declared, 0 otherwise (= 1 - IsTrendMode).
	IsCycleMode
	// InstantaneousTrendLine is the WMA-smoothed instantaneous trend line.
	InstantaneousTrendLine
	// SineWave is the sine wave value, sin(phase·Deg2Rad).
	SineWave
	// SineWaveLead is the sine wave lead value, sin((phase+45)·Deg2Rad).
	SineWaveLead
	// DominantCyclePeriod is the smoothed dominant cycle period.
	DominantCyclePeriod
	// DominantCyclePhase is the dominant cycle phase, in degrees.
	DominantCyclePhase
	outputLast
)

const (
	valueStr                  = "value"
	isTrendModeStr            = "isTrendMode"
	isCycleModeStr            = "isCycleMode"
	instantaneousTrendLineStr = "instantaneousTrendLine"
	sineWaveStr               = "sineWave"
	sineWaveLeadStr           = "sineWaveLead"
	dominantCyclePeriodStr    = "dominantCyclePeriod"
	dominantCyclePhaseStr     = "dominantCyclePhase"
	unknownStr                = "unknown"
)

// String implements the Stringer interface.
func (o Output) String() string {
	switch o {
	case Value:
		return valueStr
	case IsTrendMode:
		return isTrendModeStr
	case IsCycleMode:
		return isCycleModeStr
	case InstantaneousTrendLine:
		return instantaneousTrendLineStr
	case SineWave:
		return sineWaveStr
	case SineWaveLead:
		return sineWaveLeadStr
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
		errFmt = "cannot marshal '%s': unknown trend cycle mode output"
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
		errFmt = "cannot unmarshal '%s': unknown trend cycle mode output"
		dqs    = "\"" // Double quote string.
	)

	d := bytes.Trim(data, dqs)
	s := string(d)

	switch s {
	case valueStr:
		*o = Value
	case isTrendModeStr:
		*o = IsTrendMode
	case isCycleModeStr:
		*o = IsCycleMode
	case instantaneousTrendLineStr:
		*o = InstantaneousTrendLine
	case sineWaveStr:
		*o = SineWave
	case sineWaveLeadStr:
		*o = SineWaveLead
	case dominantCyclePeriodStr:
		*o = DominantCyclePeriod
	case dominantCyclePhaseStr:
		*o = DominantCyclePhase
	default:
		return fmt.Errorf(errFmt, s)
	}

	return nil
}
