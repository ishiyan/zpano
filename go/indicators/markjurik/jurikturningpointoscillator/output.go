package jurikturningpointoscillator

import (
	"bytes"
	"fmt"
)

type Output int

const (
	Value Output = iota + 1
	turningPointOscillatorLast
)

const (
	turningPointOscillatorValue   = "value"
	turningPointOscillatorUnknown = "unknown"
)

func (o Output) String() string {
	switch o {
	case Value:
		return turningPointOscillatorValue
	default:
		return turningPointOscillatorUnknown
	}
}

func (o Output) IsKnown() bool {
	return o >= Value && o < turningPointOscillatorLast
}

func (o Output) MarshalJSON() ([]byte, error) {
	const (
		errFmt = "cannot marshal '%s': unknown jurik turning point oscillator output"
		extra  = 2
		dqc    = '"'
	)
	s := o.String()
	if s == turningPointOscillatorUnknown {
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
		errFmt = "cannot unmarshal '%s': unknown jurik turning point oscillator output"
		dqs    = "\""
	)
	d := bytes.Trim(data, dqs)
	s := string(d)
	switch s {
	case turningPointOscillatorValue:
		*o = Value
	default:
		return fmt.Errorf(errFmt, s)
	}
	return nil
}
