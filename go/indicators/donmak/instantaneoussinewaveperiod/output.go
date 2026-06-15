//nolint:dupl
package instantaneoussinewaveperiod

import (
	"bytes"
	"fmt"
)

// Output describes the outputs of the indicator.
type Output int

const (
	// Period is the estimated cycle period in bars (may be NaN).
	Period Output = iota + 1

	// Omega is the circular frequency in radians/bar (may be NaN).
	Omega

	// Velocity is the wave velocity (may be NaN).
	Velocity

	// Acceleration is the wave acceleration (may be NaN).
	Acceleration

	// Amplitude is the estimated sine wave amplitude (may be NaN).
	Amplitude

	// Phase is the phase angle in radians (may be NaN).
	Phase

	// DcLevel is the constant level D (may be NaN).
	DcLevel

	outputLast
)

const (
	periodStr       = "period"
	omegaStr        = "omega"
	velocityStr     = "velocity"
	accelerationStr = "acceleration"
	amplitudeStr    = "amplitude"
	phaseStr        = "phase"
	dcLevelStr      = "dcLevel"
	unknownStr      = "unknown"
)

// String implements the Stringer interface.
func (o Output) String() string {
	switch o {
	case Period:
		return periodStr
	case Omega:
		return omegaStr
	case Velocity:
		return velocityStr
	case Acceleration:
		return accelerationStr
	case Amplitude:
		return amplitudeStr
	case Phase:
		return phaseStr
	case DcLevel:
		return dcLevelStr
	default:
		return unknownStr
	}
}

// IsKnown determines if this output is known.
func (o Output) IsKnown() bool {
	return o >= Period && o < outputLast
}

// MarshalJSON implements the Marshaler interface.
func (o Output) MarshalJSON() ([]byte, error) {
	const (
		errFmt = "cannot marshal '%s': unknown instantaneous sine wave period output"
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
		errFmt = "cannot unmarshal '%s': unknown instantaneous sine wave period output"
		dqs    = "\"" // Double quote string.
	)

	d := bytes.Trim(data, dqs)
	s := string(d)

	switch s {
	case periodStr:
		*o = Period
	case omegaStr:
		*o = Omega
	case velocityStr:
		*o = Velocity
	case accelerationStr:
		*o = Acceleration
	case amplitudeStr:
		*o = Amplitude
	case phaseStr:
		*o = Phase
	case dcLevelStr:
		*o = DcLevel
	default:
		return fmt.Errorf(errFmt, s)
	}

	return nil
}
