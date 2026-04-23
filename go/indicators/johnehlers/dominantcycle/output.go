package dominantcycle

import (
	"bytes"
	"fmt"
)

// Output describes the outputs of the indicator.
type Output int

const (
	// RawPeriod is the raw instantaneous cycle period produced by the Hilbert transformer estimator.
	RawPeriod Output = iota + 1
	// Period is the dominant cycle period obtained by additional EMA smoothing of the raw period.
	Period
	// Phase is the dominant cycle phase, in degrees.
	Phase
	outputLast
)

const (
	rawPeriodStr = "rawPeriod"
	periodStr    = "period"
	phaseStr     = "phase"
	unknownStr   = "unknown"
)

// String implements the Stringer interface.
func (o Output) String() string {
	switch o {
	case RawPeriod:
		return rawPeriodStr
	case Period:
		return periodStr
	case Phase:
		return phaseStr
	default:
		return unknownStr
	}
}

// IsKnown determines if this output is known.
func (o Output) IsKnown() bool {
	return o >= RawPeriod && o < outputLast
}

// MarshalJSON implements the Marshaler interface.
func (o Output) MarshalJSON() ([]byte, error) {
	const (
		errFmt = "cannot marshal '%s': unknown dominant cycle output"
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
		errFmt = "cannot unmarshal '%s': unknown dominant cycle output"
		dqs    = "\"" // Double quote string.
	)

	d := bytes.Trim(data, dqs)
	s := string(d)

	switch s {
	case rawPeriodStr:
		*o = RawPeriod
	case periodStr:
		*o = Period
	case phaseStr:
		*o = Phase
	default:
		return fmt.Errorf(errFmt, s)
	}

	return nil
}
