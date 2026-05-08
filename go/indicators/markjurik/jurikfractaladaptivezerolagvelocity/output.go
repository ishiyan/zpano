package jurikfractaladaptivezerolagvelocity

import (
	"bytes"
	"fmt"
)

type Output int

const (
	Value Output = iota + 1
	fractalAdaptiveZeroLagVelocityLast
)

const (
	fractalAdaptiveZeroLagVelocityValue   = "value"
	fractalAdaptiveZeroLagVelocityUnknown = "unknown"
)

func (o Output) String() string {
	switch o {
	case Value:
		return fractalAdaptiveZeroLagVelocityValue
	default:
		return fractalAdaptiveZeroLagVelocityUnknown
	}
}

func (o Output) IsKnown() bool {
	return o >= Value && o < fractalAdaptiveZeroLagVelocityLast
}

func (o Output) MarshalJSON() ([]byte, error) {
	const (
		errFmt = "cannot marshal '%s': unknown jurik fractal adaptive zero lag velocity output"
		extra  = 2
		dqc    = '"'
	)
	s := o.String()
	if s == fractalAdaptiveZeroLagVelocityUnknown {
		return nil, fmt.Errorf(errFmt, s)
	}
	b := make([]byte, 0, len(s)+extra)
	b = append(b, dqc)
	b = append(b, s...)
	b = append(b, dqc)
	return b, nil
}

func (o *Output) UnmarshalJSON(data []byte) error {
	const (
		errFmt = "cannot unmarshal '%s': unknown jurik fractal adaptive zero lag velocity output"
		dqs    = "\""
	)
	d := bytes.Trim(data, dqs)
	s := string(d)
	switch s {
	case fractalAdaptiveZeroLagVelocityValue:
		*o = Value
	default:
		return fmt.Errorf(errFmt, s)
	}
	return nil
}
