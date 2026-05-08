package jurikadaptiverelativetrendstrengthindex

import (
	"bytes"
	"fmt"
)

type Output int

const (
	Value Output = iota + 1
	adaptiveRelativeTrendStrengthIndexLast
)

const (
	adaptiveRelativeTrendStrengthIndexValue   = "value"
	adaptiveRelativeTrendStrengthIndexUnknown = "unknown"
)

func (o Output) String() string {
	switch o {
	case Value:
		return adaptiveRelativeTrendStrengthIndexValue
	default:
		return adaptiveRelativeTrendStrengthIndexUnknown
	}
}

func (o Output) IsKnown() bool {
	return o >= Value && o < adaptiveRelativeTrendStrengthIndexLast
}

func (o Output) MarshalJSON() ([]byte, error) {
	const (
		errFmt = "cannot marshal '%s': unknown jurik adaptive relative trend strength index output"
		extra  = 2
		dqc    = '"'
	)
	s := o.String()
	if s == adaptiveRelativeTrendStrengthIndexUnknown {
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
		errFmt = "cannot unmarshal '%s': unknown jurik adaptive relative trend strength index output"
		dqs    = "\""
	)
	d := bytes.Trim(data, dqs)
	s := string(d)
	switch s {
	case adaptiveRelativeTrendStrengthIndexValue:
		*o = Value
	default:
		return fmt.Errorf(errFmt, s)
	}
	return nil
}
